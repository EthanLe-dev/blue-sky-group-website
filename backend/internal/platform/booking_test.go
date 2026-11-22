package platform

import (
	"context"
	"encoding/json"
	"fmt"
	demoassets "github.com/EthanLe-dev/blue-sky-group-website/backend/demo-assets"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func bookingFixture(t *testing.T, capacity int) (*App, booking.Principal, string) {
	t.Helper()
	a := testApp(t)
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOCAL_STORAGE_PATH", t.TempDir())
	if e := catalog.Seed(context.Background(), a.DB, catalog.NewMediaStore(a.Config.CSRFSecret), demoassets.Files); e != nil {
		t.Fatal(e)
	}
	var staff, dep string
	a.DB.QueryRow(context.Background(), "SELECT id FROM staff_users LIMIT 1").Scan(&staff)
	a.DB.QueryRow(context.Background(), "SELECT id FROM departures ORDER BY starts_at LIMIT 1").Scan(&dep)
	if _, e := a.DB.Exec(context.Background(), "UPDATE departures SET capacity=$2 WHERE id=$1", dep, capacity); e != nil {
		t.Fatal(e)
	}
	return a, booking.Principal{ID: staff, Staff: true}, dep
}
func quoteFor(t *testing.T, a *App, p booking.Principal, dep string, counts booking.Counts) booking.Quote {
	t.Helper()
	q, e := a.BookingServer().Quote(context.Background(), p, booking.QuoteInput{Departure: dep, Counts: counts, Codes: []string{}})
	if e != nil {
		t.Fatal(e)
	}
	return q
}
func bookingInput(q booking.Quote) booking.CreateInput {
	return booking.CreateInput{Quote: q.ID, Contact: booking.Contact{Name: "Khách Demo", Email: "demo@example.com", Phone: "0901234567"}, Passengers: []booking.Passenger{{Type: "adult", Name: "Khách Demo", Birth: "1990-01-01"}}, Option: "deposit", Source: "phone", Consent: booking.Consent{Accepted: true, Version: "booking-v1"}}
}
func TestBookingConcurrentWebManualAndIdempotency(t *testing.T) {
	a, staff, dep := bookingFixture(t, 3)
	ctx := context.Background()
	var guest string
	a.DB.QueryRow(ctx, "INSERT INTO guest_sessions(token_hash,expires_at) VALUES($1,now()+interval '1 day') RETURNING id", hash("guest-test")).Scan(&guest)
	principals := []booking.Principal{staff, {ID: guest}}
	quotes := []booking.Quote{}
	for i := 0; i < 12; i++ {
		quotes = append(quotes, quoteFor(t, a, principals[i%2], dep, booking.Counts{Adult: 1}))
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i, q := range quotes {
		wg.Add(1)
		go func(i int, q booking.Quote) {
			defer wg.Done()
			p := principals[i%2]
			b := bookingInput(q)
			if !p.Staff {
				b.Source = "web"
				b.Captcha = "local-test-token"
			}
			_, e := a.BookingServer().Create(ctx, p, uuid.NewString(), b)
			if e == nil {
				successes.Add(1)
			} else if e.Error() != "SOLD_OUT" {
				t.Errorf("unexpected create: %v", e)
			}
		}(i, q)
	}
	wg.Wait()
	if successes.Load() != 3 {
		t.Fatalf("created %d for capacity 3", successes.Load())
	}
	var held, actual int
	a.DB.QueryRow(ctx, "SELECT held_seats FROM departures WHERE id=$1", dep).Scan(&held)
	a.DB.QueryRow(ctx, "SELECT sum(seat_count) FROM seat_holds WHERE departure_id=$1 AND status='active'", dep).Scan(&actual)
	if held != 3 || actual != 3 {
		t.Fatalf("counter held=%d sum=%d", held, actual)
	}
	// Replays share one committed response even when callers race.
	a.DB.Exec(ctx, "UPDATE departures SET capacity=4 WHERE id=$1", dep)
	q := quoteFor(t, a, staff, dep, booking.Counts{Adult: 1})
	b := bookingInput(q)
	key := uuid.NewString()
	out, e := a.BookingServer().Create(ctx, staff, key, b)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			again, e := a.BookingServer().Create(ctx, staff, key, b)
			if e != nil {
				t.Error(e)
				return
			}
			var first, second map[string]any
			json.Unmarshal(out, &first)
			json.Unmarshal(again, &second)
			if first["id"] != second["id"] {
				t.Error("idempotency duplicated")
			}
		}()
	}
	wg.Wait()
	b.Contact.Name = "Khách khác"
	if _, e = a.BookingServer().Create(ctx, staff, key, b); e == nil || e.Error() != "IDEMPOTENCY_CONFLICT" {
		t.Fatal("payload conflict missing", e)
	}
	var total int
	a.DB.QueryRow(ctx, "SELECT count(*) FROM bookings WHERE departure_id=$1", dep).Scan(&total)
	if total != 4 {
		t.Fatalf("bookings=%d", total)
	}
}
func TestBookingExpirySnapshotsAndPriceChange(t *testing.T) {
	a, p, dep := bookingFixture(t, 2)
	ctx := context.Background()
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	a.DB.Exec(ctx, "UPDATE departures SET price_version=price_version+1 WHERE id=$1", dep)
	if _, e := a.BookingServer().Create(ctx, p, uuid.NewString(), bookingInput(q)); e == nil || e.Error() != "PRICE_CHANGED" {
		t.Fatal("accepted stale version", e)
	}
	q = quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	b := bookingInput(q)
	b.Passengers[0].Birth = time.Now().Format("2006-01-02")
	if _, e := a.BookingServer().Create(ctx, p, uuid.NewString(), b); e == nil || e.Error() != "PASSENGER_MISMATCH" {
		t.Fatal("wrong age accepted", e)
	}
	b = bookingInput(q)
	out, e := a.BookingServer().Create(ctx, p, uuid.NewString(), b)
	if e != nil {
		t.Fatal(e)
	}
	var data map[string]any
	json.Unmarshal(out, &data)
	id := data["id"].(string)
	a.DB.Exec(ctx, `UPDATE site_settings SET value=jsonb_set(value,'{hold_minutes}','30'),version=version+1 WHERE key='booking_defaults'`)
	var hold int
	a.DB.QueryRow(ctx, "SELECT (settings_snapshot->>'hold_minutes')::int FROM bookings WHERE id=$1", id).Scan(&hold)
	if hold != q.Hold {
		t.Fatal("snapshot changed")
	}
	a.DB.Exec(ctx, "UPDATE seat_holds SET expires_at=clock_timestamp()-interval '1 second' WHERE booking_id=$1", id)
	if e = a.BookingServer().Sweep(ctx); e != nil {
		t.Fatal(e)
	}
	if e = a.BookingServer().Sweep(ctx); e != nil {
		t.Fatal(e)
	}
	var state string
	var held int
	a.DB.QueryRow(ctx, "SELECT status FROM bookings WHERE id=$1", id).Scan(&state)
	a.DB.QueryRow(ctx, "SELECT held_seats FROM departures WHERE id=$1", dep).Scan(&held)
	if state != "expired" || held != 0 {
		t.Fatalf("expiry %s held%d", state, held)
	}
	var events int
	a.DB.QueryRow(ctx, "SELECT count(*) FROM booking_events WHERE booking_id=$1 AND event_type='expired'", id).Scan(&events)
	if events != 1 {
		t.Fatal("duplicate expiry event")
	}
}
func TestGuestLookupPrivacyAndCreatorIsolation(t *testing.T) {
	a, _, dep := bookingFixture(t, 3)
	ctx := context.Background()
	raw := "creator-secret"
	var gid string
	a.DB.QueryRow(ctx, "INSERT INTO guest_sessions(token_hash,expires_at) VALUES($1,now()+interval '1 day') RETURNING id", hash(raw)).Scan(&gid)
	p := booking.Principal{ID: gid}
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	b := bookingInput(q)
	b.Source = "web"
	b.Captcha = "local-test-token"
	out, e := a.BookingServer().Create(ctx, p, uuid.NewString(), b)
	if e != nil {
		t.Fatal(e)
	}
	var d map[string]any
	json.Unmarshal(out, &d)
	id, code := d["id"].(string), d["code"].(string)
	w := request(a, "GET", "/api/v1/bookings/"+id, "", "", "", &http.Cookie{Name: "guest_session", Value: raw})
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	w = request(a, "POST", "/api/v1/guest-sessions", "{}", "", "", nil)
	other := w.Result().Cookies()[0]
	w = request(a, "GET", "/api/v1/bookings/"+id, "", "", "", other)
	if w.Code != 404 {
		t.Fatal("creator isolation", w.Code)
	}
	body := fmt.Sprintf(`{"code":%q,"contact":"0901234567","turnstile_token":"local-test-token"}`, code)
	w = request(a, "POST", "/api/v1/booking-lookups", body, "", "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	lookup := data(t, w)
	summary, _ := json.Marshal(lookup["booking"])
	for _, secret := range []string{"Khách Demo", "demo@example.com", "+84901234567", "1990-01-01", "passengers", "special_requests"} {
		if strings.Contains(string(summary), secret) {
			t.Fatalf("lookup exposes %s", secret)
		}
	}
	token := lookup["lookup_token"].(string)
	w = request(a, "GET", "/api/v1/booking-lookups/current", "", token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	w = request(a, "GET", "/api/v1/bookings/"+id, "", token, "", nil)
	if w.Code != 401 {
		t.Fatal("lookup token grants creator")
	}
	a.DB.Exec(ctx, "UPDATE booking_lookup_tokens SET expires_at=now()-interval '1 second'")
	w = request(a, "GET", "/api/v1/booking-lookups/current", "", token, "", nil)
	if w.Code != 401 {
		t.Fatal("expired lookup token")
	}
	w = request(a, "POST", "/api/v1/bookings", `{}`, "", "", &http.Cookie{Name: "guest_session", Value: raw})
	if w.Code != 403 {
		t.Fatal("CSRF missing", w.Code)
	}
}
func TestAdminCancelVersionAndAtomicRollback(t *testing.T) {
	a, p, dep := bookingFixture(t, 2)
	ctx := context.Background()
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	b := bookingInput(q)
	key := uuid.NewString()
	a.DB.Exec(ctx, `CREATE FUNCTION reject_hold() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected failure'; END $$; CREATE TRIGGER reject_hold BEFORE INSERT ON seat_holds FOR EACH ROW EXECUTE FUNCTION reject_hold()`)
	if _, e := a.BookingServer().Create(ctx, p, key, b); e == nil {
		t.Fatal("failure not injected")
	}
	var count, held int
	a.DB.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count)
	a.DB.QueryRow(ctx, "SELECT held_seats FROM departures WHERE id=$1", dep).Scan(&held)
	if count != 0 || held != 0 {
		t.Fatal("partial transaction")
	}
	a.DB.Exec(ctx, "DROP TRIGGER reject_hold ON seat_holds;DROP FUNCTION reject_hold()")
	out, e := a.BookingServer().Create(ctx, p, key, b)
	if e != nil {
		t.Fatal(e)
	}
	var d map[string]any
	json.Unmarshal(out, &d)
	id := d["id"].(string)
	_, sid, _ := a.createSession(ctx, p.ID)
	token, _ := a.access(p.ID, sid)
	cancel := func(version int, key string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/v1/admin/bookings/"+id+"/cancel", strings.NewReader(fmt.Sprintf(`{"expected_version":%d,"reason":"Khách yêu cầu huỷ demo"}`, version)))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		a.Router().ServeHTTP(w, r)
		return w
	}
	if w := cancel(99, uuid.NewString()); w.Code != 409 {
		t.Fatal("stale version", w.Body)
	}
	cancelKey := uuid.NewString()
	if w := cancel(1, cancelKey); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := cancel(1, cancelKey); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := cancel(1, uuid.NewString()); w.Code != 200 {
		t.Fatal(w.Body)
	}
	a.DB.QueryRow(ctx, "SELECT held_seats FROM departures WHERE id=$1", dep).Scan(&held)
	if held != 0 {
		t.Fatal("double release")
	}
}

func TestQuoteRoundingSurchargesAndExpiry(t *testing.T) {
	a, p, dep := bookingFixture(t, 5)
	ctx := context.Background()
	a.DB.Exec(ctx, "UPDATE departure_prices SET unit_price=101 WHERE departure_id=$1 AND passenger_type='adult'", dep)
	a.DB.Exec(ctx, "UPDATE departures SET deposit_bps=3333 WHERE id=$1", dep)
	a.DB.Exec(ctx, "INSERT INTO departure_surcharges(departure_id,code,label,amount,unit,required) VALUES($1,'fee','Phụ thu bắt buộc',2,'per_person',true),($1,'room','Phòng riêng',7,'per_booking',false)", dep)
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	if q.Snapshot.Total != 103 || q.Snapshot.Deposit != 35 || q.Snapshot.Balance != 68 {
		t.Fatalf("rounding: %+v", q.Snapshot)
	}
	q2, e := a.BookingServer().Quote(ctx, p, booking.QuoteInput{Departure: dep, Counts: booking.Counts{Adult: 1}, Codes: []string{"room"}})
	if e != nil || q2.Snapshot.Total != 110 {
		t.Fatal("optional surcharge", e, q2)
	}
	if _, e = a.BookingServer().Quote(ctx, p, booking.QuoteInput{Departure: dep, Counts: booking.Counts{Adult: 1}, Codes: []string{"unknown"}}); e == nil {
		t.Fatal("unknown surcharge")
	}
	a.DB.Exec(ctx, "UPDATE pricing_quotes SET expires_at=clock_timestamp()-interval '1 second' WHERE id=$1", q.ID)
	if _, e = a.BookingServer().Create(ctx, p, uuid.NewString(), bookingInput(q)); e == nil || e.Error() != "QUOTE_EXPIRED" {
		t.Fatal("expired quote", e)
	}
	a.DB.Exec(ctx, "UPDATE departures SET balance_due_at=clock_timestamp()-interval '1 day' WHERE id=$1", dep)
	q = quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	if len(q.Options) != 1 || q.Options[0] != "full" {
		t.Fatal("late deposit offered")
	}
}
func TestCompletePreservesHistoricalInventory(t *testing.T) {
	a, p, dep := bookingFixture(t, 2)
	ctx := context.Background()
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	out, e := a.BookingServer().Create(ctx, p, uuid.NewString(), bookingInput(q))
	if e != nil {
		t.Fatal(e)
	}
	var d map[string]any
	json.Unmarshal(out, &d)
	id := d["id"].(string)
	a.DB.Exec(ctx, "UPDATE bookings SET status='confirmed' WHERE id=$1", id)
	a.DB.Exec(ctx, "UPDATE seat_holds SET status='converted' WHERE booking_id=$1", id)
	a.DB.Exec(ctx, "UPDATE departures SET held_seats=0,confirmed_seats=1 WHERE id=$1", dep)
	_, sid, _ := a.createSession(ctx, p.ID)
	token, _ := a.access(p.ID, sid)
	w := request(a, "POST", "/api/v1/admin/bookings/"+id+"/complete", `{"expected_version":1}`, token, "", nil)
	if w.Code != 409 {
		t.Fatal("completed future trip", w.Body)
	}
	a.DB.Exec(ctx, "UPDATE departures SET starts_at=now()-interval '3 days',ends_at=now()-interval '2 days',sales_close_at=now()-interval '4 days',balance_due_at=now()-interval '4 days' WHERE id=$1", dep)
	w = request(a, "POST", "/api/v1/admin/bookings/"+id+"/complete", `{"expected_version":1}`, token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	var confirmed, paid int
	a.DB.QueryRow(ctx, "SELECT d.confirmed_seats,b.paid_amount FROM departures d JOIN bookings b ON b.departure_id=d.id WHERE b.id=$1", id).Scan(&confirmed, &paid)
	if confirmed != 1 || paid != 0 {
		t.Fatal("completion mutated seats or cash")
	}
}

func TestQuoteClockAfterWaitingForDepartureLock(t *testing.T) {
	a, p, dep := bookingFixture(t, 2)
	ctx := context.Background()
	q := quoteFor(t, a, p, dep, booking.Counts{Adult: 1})
	a.DB.Exec(ctx, "UPDATE pricing_quotes SET expires_at=clock_timestamp()+interval '100 milliseconds' WHERE id=$1", q.ID)
	tx, e := a.DB.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer tx.Rollback(ctx)
	var locked string
	tx.QueryRow(ctx, "SELECT id FROM departures WHERE id=$1 FOR UPDATE", dep).Scan(&locked)
	result := make(chan error, 1)
	go func() { _, err := a.BookingServer().Create(ctx, p, uuid.NewString(), bookingInput(q)); result <- err }()
	time.Sleep(250 * time.Millisecond)
	tx.Commit(ctx)
	if e := <-result; e == nil || e.Error() != "QUOTE_EXPIRED" {
		t.Fatal("used transaction-start time after lock", e)
	}
}

func TestBookingRoleMatrix(t *testing.T) {
	a, p, dep := bookingFixture(t, 3)
	ctx := context.Background()
	_, sid, _ := a.createSession(ctx, p.ID)
	token, _ := a.access(p.ID, sid)
	for _, tc := range []struct {
		role                 string
		read, write, catalog int
	}{{"admin", 200, 201, 200}, {"product_manager", 403, 403, 200}, {"sales", 200, 201, 403}, {"accountant", 200, 403, 403}, {"executive", 403, 403, 403}} {
		if _, e := a.DB.Exec(ctx, "DELETE FROM staff_roles WHERE staff_id=$1", p.ID); e != nil {
			t.Fatal(e)
		}
		if _, e := a.DB.Exec(ctx, "INSERT INTO staff_roles(staff_id,role_id) SELECT $1,id FROM roles WHERE code=$2", p.ID, tc.role); e != nil {
			t.Fatal(e)
		}
		w := request(a, "GET", "/api/v1/admin/bookings", "", token, "", nil)
		if w.Code != tc.read {
			t.Fatalf("%s read %d", tc.role, w.Code)
		}
		w = request(a, "POST", "/api/v1/admin/booking-quotes", fmt.Sprintf(`{"departure_id":%q,"counts":{"adult":1,"child":0,"infant":0},"surcharge_codes":[]}`, dep), token, "", nil)
		if w.Code != tc.write {
			t.Fatalf("%s quote %d %s", tc.role, w.Code, w.Body)
		}
		w = request(a, "GET", "/api/v1/admin/tours", "", token, "", nil)
		if w.Code != tc.catalog {
			t.Fatalf("%s catalog %d", tc.role, w.Code)
		}
	}
}
