package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/google/uuid"
)

func TestMockPaymentLedgerAndNotifications(t *testing.T) {
	a, staff, departure := bookingFixture(t, 4)
	a.Config.MockPayment, a.Config.MockNotification = true, true
	ctx := context.Background()
	_, sid, err := a.createSession(ctx, staff.ID)
	if err != nil {
		t.Fatal(err)
	}
	token, err := a.access(staff.ID, sid)
	if err != nil {
		t.Fatal(err)
	}
	router := a.Router()
	call := func(method, path, body, key, bearer string, guest *http.Cookie, csrf string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		r.Header.Set("Origin", "http://localhost:5173")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", key)
		r.Header.Set("X-CSRF-Token", csrf)
		if bearer != "" {
			r.Header.Set("Authorization", "Bearer "+bearer)
		}
		if guest != nil {
			r.AddCookie(guest)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	admin := func(method, path, body, key string) *httptest.ResponseRecorder {
		return call(method, "/admin"+path, body, key, token, nil, "")
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := a.DB.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	must := func(w *httptest.ResponseRecorder, status int) map[string]any {
		t.Helper()
		if w.Code != status {
			t.Fatalf("status=%d want=%d: %s", w.Code, status, w.Body)
		}
		return data(t, w)
	}
	newBooking := func(p booking.Principal, purpose string) string {
		t.Helper()
		input := bookingInput(quoteFor(t, a, p, departure, booking.Counts{Adult: 1}))
		input.Option = purpose
		if !p.Staff {
			input.Source, input.Captcha = "web", "local-test-token"
		}
		raw, err := a.BookingServer().Create(ctx, p, uuid.NewString(), input)
		if err != nil {
			t.Fatal(err)
		}
		var record struct{ ID string }
		if err = json.Unmarshal(raw, &record); err != nil {
			t.Fatal(err)
		}
		return record.ID
	}
	guestSession := call("POST", "/guest-sessions", "{}", "", "", nil, "")
	guest := guestSession.Result().Cookies()[0]
	csrf := must(guestSession, 201)["csrf_token"].(string)
	var guestID string
	if err = a.DB.QueryRow(ctx, "SELECT id FROM guest_sessions WHERE token_hash=$1", hash(guest.Value)).Scan(&guestID); err != nil {
		t.Fatal(err)
	}
	id := newBooking(booking.Principal{ID: guestID}, "deposit")
	payments := "/bookings/" + id + "/payments"
	if w := call("POST", payments, `{"purpose":"deposit"}`, uuid.NewString(), "", guest, ""); w.Code != 403 {
		t.Fatal("CSRF", w.Body)
	}
	intentKey := uuid.NewString()
	intent := must(call("POST", payments, `{"purpose":"deposit"}`, intentKey, "", guest, csrf), 201)
	paymentID := intent["id"].(string)
	if got := must(call("POST", payments, `{"purpose":"deposit"}`, intentKey, "", guest, csrf), 201)["id"]; got != paymentID {
		t.Fatal("intent replay", got)
	}
	if w := call("POST", payments, `{"purpose":"deposit"}`, uuid.NewString(), "", guest, csrf); w.Code != 409 {
		t.Fatal("second pending intent", w.Body)
	}
	other := call("POST", "/guest-sessions", "{}", "", "", nil, "").Result().Cookies()[0]
	if w := call("GET", payments, "", "", "", other, ""); w.Code != 404 {
		t.Fatal("creator isolation", w.Body)
	}
	confirm := "/payments/" + paymentID + "/mock-confirm"
	requests := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		go func() { requests <- call("POST", confirm, "{}", uuid.NewString(), "", guest, csrf) }()
	}
	for range 2 {
		b := must(<-requests, 200)["booking"].(map[string]any)
		if b["status"] != "confirmed" || b["payment_status"] != "partially_paid" || b["paid_amount"] != intent["amount"] {
			t.Fatal("deposit", b)
		}
	}

	// A worker dies after claiming. The next worker reclaims the expired lease.
	exec("UPDATE outbox_events SET status='processing',attempts=1,lease_until=clock_timestamp()-interval '1 second'")
	if found, err := a.BookingServer().ProcessNotification(ctx); !found || err != nil {
		t.Fatal("reclaim", found, err)
	}
	if found, err := a.BookingServer().ProcessNotification(ctx); found || err != nil {
		t.Fatal("done event replay", found, err)
	}
	preview := call("GET", "/bookings/"+id+"/notifications", "", "", "", guest, "")
	if preview.Code != 200 || strings.Count(preview.Body.String(), `"simulated"`) != 2 {
		t.Fatal(preview.Body)
	}
	if w := call("GET", "/bookings/"+id+"/notifications", "", "", "", other, ""); w.Code != 404 {
		t.Fatal("notification isolation", w.Body)
	}

	balance := must(call("POST", payments, `{"purpose":"balance"}`, uuid.NewString(), "", guest, csrf), 201)
	balanceID := balance["id"].(string)
	b := must(call("POST", "/payments/"+balanceID+"/mock-confirm", "{}", uuid.NewString(), "", guest, csrf), 200)["booking"].(map[string]any)
	if b["payment_status"] != "paid" || b["balance_amount"] != "0" {
		t.Fatal("balance", b)
	}
	exec(`CREATE FUNCTION fail_preview() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'forced preview write failure'; END $$`)
	exec("CREATE TRIGGER fail_preview BEFORE INSERT ON notification_logs FOR EACH ROW EXECUTE FUNCTION fail_preview()")
	for range 5 {
		if found, err := a.BookingServer().ProcessNotification(ctx); !found || err == nil {
			t.Fatal("failed attempt", found, err)
		}
		exec("UPDATE outbox_events SET next_attempt_at=clock_timestamp() WHERE status='pending'")
	}
	exec("DROP TRIGGER fail_preview ON notification_logs")
	if found, err := a.BookingServer().ProcessNotification(ctx); found || err != nil {
		t.Fatal("attempt cap", found, err)
	}
	log := must(admin("GET", "/notifications?booking_id="+id, "", ""), 200)
	if len(log["logs"].([]any)) != 2 || len(log["jobs"].([]any)) != 1 || log["jobs"].([]any)[0].(map[string]any)["status"] != "failed" {
		t.Fatal("failed job visibility", log)
	}

	refundPath := "/payments/" + paymentID + "/refunds"
	body := fmt.Sprintf(`{"amount":%q,"reason":"Kiểm thử hoàn tiền"}`, intent["amount"])
	exec("DELETE FROM staff_roles WHERE staff_id=$1", staff.ID)
	exec("INSERT INTO staff_roles SELECT $1,id FROM roles WHERE code='sales'", staff.ID)
	if w := admin("POST", refundPath, body, uuid.NewString()); w.Code != 403 {
		t.Fatal("sales refund", w.Body)
	}
	exec("DELETE FROM staff_roles WHERE staff_id=$1", staff.ID)
	exec("INSERT INTO staff_roles SELECT $1,id FROM roles WHERE code='accountant'", staff.ID)
	for range 2 {
		go func() { requests <- admin("POST", refundPath, body, uuid.NewString()) }()
	}
	first, second := <-requests, <-requests
	if first.Code+second.Code != 201+409 || first.Code == second.Code {
		t.Fatalf("parallel refund: %d %s / %d %s", first.Code, first.Body, second.Code, second.Body)
	}
	ledger := must(admin("GET", payments, "", ""), 200)
	if ledger["reconciled"] != true || len(ledger["refunds"].([]any)) != 1 {
		t.Fatal(ledger)
	}
	var held, confirmed, events int
	if err = a.DB.QueryRow(ctx, "SELECT held_seats,confirmed_seats,(SELECT count(*) FROM payment_events) FROM departures WHERE id=$1", departure).Scan(&held, &confirmed, &events); err != nil || held != 0 || confirmed != 1 || events != 2 {
		t.Fatalf("inventory/events %d/%d/%d: %v", held, confirmed, events, err)
	}

	expiring := newBooking(staff, "full")
	fullID := must(admin("POST", "/bookings/"+expiring+"/payments", `{"purpose":"full"}`, uuid.NewString()), 201)["id"].(string)
	exec("UPDATE seat_holds SET expires_at=clock_timestamp()-interval '1 second' WHERE booking_id=$1", expiring)
	expired := admin("POST", "/payments/"+fullID+"/mock-confirm", `{"reason":"Kiểm thử hết hạn"}`, uuid.NewString())
	if expired.Code != 409 || !strings.Contains(expired.Body.String(), "HOLD_EXPIRED") {
		t.Fatal("late confirm", expired.Body)
	}
	// Confirm and cancel can enter concurrently, but both lock the departure first.
	exec("DELETE FROM staff_roles WHERE staff_id=$1", staff.ID)
	exec("INSERT INTO staff_roles(staff_id,role_id) SELECT $1,id FROM roles WHERE code='admin'", staff.ID)
	racing := newBooking(staff, "full")
	racingID := must(admin("POST", "/bookings/"+racing+"/payments", `{"purpose":"full"}`, uuid.NewString()), 201)["id"].(string)
	confirmResult := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		confirmResult <- admin("POST", "/payments/"+racingID+"/mock-confirm", `{"reason":"Kiểm thử cạnh tranh"}`, uuid.NewString())
	}()
	cancelled := admin("POST", "/bookings/"+racing+"/cancel", `{"expected_version":1,"reason":"Kiểm thử cạnh tranh"}`, uuid.NewString())
	confirmedResult := <-confirmResult
	if !((cancelled.Code == 200 && confirmedResult.Code == 409) ||
		(cancelled.Code == 409 && confirmedResult.Code == 200)) {
		t.Fatalf("confirm/cancel: %d %s / %d %s", confirmedResult.Code, confirmedResult.Body, cancelled.Code, cancelled.Body)
	}
	var state string
	var paidAmount int64
	if err := a.DB.QueryRow(ctx, "SELECT status,paid_amount FROM bookings WHERE id=$1", racing).Scan(&state, &paidAmount); err != nil {
		t.Fatal(err)
	}
	if cancelled.Code == 200 && (state != "cancelled" || paidAmount != 0) ||
		confirmedResult.Code == 200 && (state != "confirmed" || paidAmount <= 0) {
		t.Fatalf("confirm/cancel left state %s and paid %d", state, paidAmount)
	}

	a.Config.Env = "production"
	if w := request(a, "POST", "/api/v1"+confirm, "{}", "", csrf, guest); w.Code != 404 {
		t.Fatal("production mock route mounted", w.Body)
	}
}
