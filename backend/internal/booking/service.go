package booking

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

// ExpireLocked requires the caller to hold the departure row lock in this transaction.
func ExpireLocked(ctx context.Context, tx pgx.Tx, departure string) error {
	rows, e := tx.Query(ctx, "SELECT b.id,b.seat_count FROM bookings b JOIN seat_holds h ON h.booking_id=b.id WHERE b.departure_id=$1 AND b.status='pending' AND h.status='active' AND h.expires_at<=clock_timestamp() ORDER BY b.id FOR UPDATE OF b", departure)
	if e != nil {
		return e
	}
	type expired struct {
		id    string
		seats int
	}
	list := []expired{}
	for rows.Next() {
		var v expired
		if e = rows.Scan(&v.id, &v.seats); e != nil {
			rows.Close()
			return e
		}
		list = append(list, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, v := range list {
		if _, e = tx.Exec(ctx, "UPDATE bookings SET status='expired',version=version+1,updated_at=clock_timestamp() WHERE id=$1", v.id); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "UPDATE seat_holds SET status='expired',updated_at=clock_timestamp() WHERE booking_id=$1 AND status='active'", v.id); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "UPDATE departures SET held_seats=held_seats-$2 WHERE id=$1", departure, v.seats); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "INSERT INTO booking_events(booking_id,event_type) VALUES($1,'expired')", v.id); e != nil {
			return e
		}
	}
	return nil
}
func (s *Server) Expire(ctx context.Context, departure string) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var id string
	if e = tx.QueryRow(ctx, "SELECT id FROM departures WHERE id=$1 FOR UPDATE", departure).Scan(&id); e != nil {
		return e
	}
	if e = ExpireLocked(ctx, tx, departure); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s *Server) Sweep(ctx context.Context) error {
	rows, e := s.DB.Query(ctx, "SELECT DISTINCT departure_id::text FROM seat_holds WHERE status='active' AND expires_at<=clock_timestamp() ORDER BY departure_id::text LIMIT 500")
	if e != nil {
		return e
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return e
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, id := range ids {
		if e = s.Expire(ctx, id); e != nil {
			return e
		}
	}
	return nil
}
func (s *Server) RunWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	_ = s.Sweep(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.Sweep(ctx)
		}
	}
}
func (s *Server) Quote(ctx context.Context, p Principal, b QuoteInput) (Quote, error) {
	q := Quote{}
	if uuid.Validate(b.Departure) != nil || b.Counts.Adult < 1 || b.Counts.Child < 0 || b.Counts.Infant < 0 || b.Counts.Adult > 20 || b.Counts.Child > 20 || b.Counts.Infant > 20 || b.Counts.Total() > 20 || len(b.Codes) > 30 {
		return q, problem("VALIDATION_ERROR")
	}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return q, e
	}
	defer tx.Rollback(ctx)
	var status, policy string
	var close time.Time
	var bps, capacity int
	var tour json.RawMessage
	if e = tx.QueryRow(ctx, "SELECT d.status,d.sales_close_at,d.price_version,d.deposit_bps,d.starts_at,d.balance_due_at,d.capacity-d.held_seats-d.confirmed_seats,jsonb_build_object('id',t.id,'slug',t.slug,'title',t.title,'starts_at',d.starts_at,'ends_at',d.ends_at),t.cancellation_policy FROM departures d JOIN tours t ON t.id=d.tour_id WHERE d.id=$1 AND t.status='published' FOR UPDATE OF d", b.Departure).Scan(&status, &close, &q.Version, &bps, &q.Snapshot.Start, &q.Due, &capacity, &tour, &policy); e != nil {
		return q, ErrNotFound
	}
	if e = ExpireLocked(ctx, tx, b.Departure); e != nil {
		return q, e
	}
	if e = tx.QueryRow(ctx, "SELECT clock_timestamp()").Scan(&q.Now); e != nil {
		return q, e
	}
	if status != "open" || !q.Now.Before(close) {
		return q, problem("SALES_CLOSED")
	}
	var defaultsRaw []byte
	var defaults Defaults
	if e = tx.QueryRow(ctx, "SELECT version,value FROM site_settings WHERE key='booking_defaults'").Scan(&q.SettingsVersion, &defaultsRaw); e != nil {
		return q, e
	}
	if e = json.Unmarshal(defaultsRaw, &defaults); e != nil {
		return q, e
	}
	if b.Counts.Total() > defaults.Max {
		return q, problem("VALIDATION_ERROR")
	}
	rows, e := tx.Query(ctx, "SELECT passenger_type,unit_price,seats_per_person,min_age,max_age_exclusive FROM departure_prices WHERE departure_id=$1 ORDER BY min_age", b.Departure)
	if e != nil {
		return q, e
	}
	q.Snapshot.Prices = []catalog.Price{}
	for rows.Next() {
		var price catalog.Price
		if e = rows.Scan(&price.Type, &price.Amount, &price.Seats, &price.Min, &price.Max); e != nil {
			rows.Close()
			return q, e
		}
		q.Snapshot.Prices = append(q.Snapshot.Prices, price)
		count := b.Counts.Get(price.Type)
		q.Snapshot.Seats += count * price.Seats
		if count > 0 {
			line := Line{price.Type, price.Type, count, price.Amount, price.Amount * catalog.Money(count)}
			q.Snapshot.Lines = append(q.Snapshot.Lines, line)
			q.Snapshot.Total += line.Total
		}
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return q, e
	}
	if len(q.Snapshot.Prices) != 3 || q.Snapshot.Total <= 0 || q.Snapshot.Seats < 1 {
		return q, problem("VALIDATION_ERROR")
	}
	selected := map[string]bool{}
	for _, code := range b.Codes {
		if selected[code] {
			return q, problem("VALIDATION_ERROR")
		}
		selected[code] = true
	}
	rows, e = tx.Query(ctx, "SELECT code,label,amount,unit,required FROM departure_surcharges WHERE departure_id=$1 ORDER BY code", b.Departure)
	if e != nil {
		return q, e
	}
	for rows.Next() {
		var code, label, unit string
		var amount catalog.Money
		var required bool
		if e = rows.Scan(&code, &label, &amount, &unit, &required); e != nil {
			rows.Close()
			return q, e
		}
		take := required || selected[code]
		delete(selected, code)
		if take {
			count := 1
			if unit == "per_person" {
				count = b.Counts.Total()
			}
			line := Line{code, label, count, amount, amount * catalog.Money(count)}
			q.Snapshot.Lines = append(q.Snapshot.Lines, line)
			q.Snapshot.Total += line.Total
		}
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return q, e
	}
	if len(selected) > 0 || q.Snapshot.Total > 1000000000000 {
		return q, problem("VALIDATION_ERROR")
	}
	q.Snapshot.Deposit = (q.Snapshot.Total*catalog.Money(bps) + 9999) / 10000
	q.Snapshot.Balance = q.Snapshot.Total - q.Snapshot.Deposit
	q.Snapshot.Policy = policy
	q.Snapshot.Tour = tour
	q.Snapshot.Due = q.Due
	q.Snapshot.Options = []string{"full"}
	if q.Now.Before(q.Due) && q.Snapshot.Deposit < q.Snapshot.Total {
		q.Snapshot.Options = append(q.Snapshot.Options, "deposit")
	}
	q.ID = uuid.NewString()
	q.Departure = b.Departure
	q.Counts = b.Counts
	q.Seats = q.Snapshot.Seats
	q.Options = q.Snapshot.Options
	q.Hold = defaults.Hold
	q.Max = defaults.Max
	q.Expires = q.Now.Add(5 * time.Minute)
	guest, staff := p.Values()
	counts, _ := json.Marshal(b.Counts)
	codes, _ := json.Marshal(b.Codes)
	snapshot, _ := json.Marshal(q.Snapshot)
	settings, _ := json.Marshal(defaults)
	_, e = tx.Exec(ctx, "INSERT INTO pricing_quotes(id,departure_id,guest_session_id,created_by_staff_id,price_version,counts,selected_surcharge_codes,pricing_snapshot,total_amount,deposit_amount,settings_version,settings_snapshot,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)", q.ID, b.Departure, guest, staff, q.Version, counts, codes, snapshot, int64(q.Snapshot.Total), int64(q.Snapshot.Deposit), q.SettingsVersion, settings, q.Expires)
	if e != nil {
		return q, e
	}
	return q, tx.Commit(ctx)
}
func loadQuote(ctx context.Context, tx pgx.Tx, p Principal, id string) (Quote, error) {
	q := Quote{}
	var counts, snapshot, defaultsRaw []byte
	guest, staff := p.Values()
	e := tx.QueryRow(ctx, "SELECT id,departure_id,price_version,counts,pricing_snapshot,settings_version,settings_snapshot,expires_at FROM pricing_quotes WHERE id=$1 AND (guest_session_id=$2 OR created_by_staff_id=$3)", id, guest, staff).Scan(&q.ID, &q.Departure, &q.Version, &counts, &snapshot, &q.SettingsVersion, &defaultsRaw, &q.Expires)
	if e != nil {
		return q, ErrNotFound
	}
	if e = json.Unmarshal(counts, &q.Counts); e != nil {
		return q, e
	}
	if e = json.Unmarshal(snapshot, &q.Snapshot); e != nil {
		return q, e
	}
	var d Defaults
	if e = json.Unmarshal(defaultsRaw, &d); e != nil {
		return q, e
	}
	q.Hold = d.Hold
	q.Max = d.Max
	q.Seats = q.Snapshot.Seats
	q.Due = q.Snapshot.Due
	q.Options = q.Snapshot.Options
	return q, nil
}
func digest(v any) string {
	raw, _ := json.Marshal(v)
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}
func Claim(ctx context.Context, tx pgx.Tx, p Principal, route, key string, payload any) (json.RawMessage, error) {
	if uuid.Validate(key) != nil {
		return nil, problem("IDEMPOTENCY_KEY_REQUIRED")
	}
	h := digest(payload)
	tag, e := tx.Exec(ctx, "INSERT INTO idempotency_keys(principal,route,key,request_hash) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING", p.Key(), route, key, h)
	if e != nil {
		return nil, e
	}
	if tag.RowsAffected() == 1 {
		return nil, nil
	}
	var old string
	var response json.RawMessage
	e = tx.QueryRow(ctx, "SELECT request_hash,response FROM idempotency_keys WHERE principal=$1 AND route=$2 AND key=$3", p.Key(), route, key).Scan(&old, &response)
	if e != nil {
		return nil, e
	}
	if old != h {
		return nil, problem("IDEMPOTENCY_CONFLICT")
	}
	if len(response) == 0 {
		return nil, problem("REQUEST_IN_PROGRESS")
	}
	return response, nil
}
func Remember(ctx context.Context, tx pgx.Tx, p Principal, route, key string, response json.RawMessage) error {
	_, e := tx.Exec(ctx, "UPDATE idempotency_keys SET response=$4 WHERE principal=$1 AND route=$2 AND key=$3", p.Key(), route, key, response)
	return e
}
func (s *Server) Create(ctx context.Context, p Principal, key string, b CreateInput) (json.RawMessage, error) {
	if uuid.Validate(b.Quote) != nil || !NormalizeContact(&b.Contact) || len(b.Requests) > 5000 || !b.Consent.Accepted || b.Consent.Version != "booking-v1" || b.Option != "full" && b.Option != "deposit" || !p.Staff && b.Source != "" && b.Source != "web" || b.Customer != nil && uuid.Validate(*b.Customer) != nil {
		return nil, problem("VALIDATION_ERROR")
	}
	if !p.Staff && b.Customer != nil {
		return nil, problem("VALIDATION_ERROR")
	}
	if p.Staff {
		if b.Source != "phone" && b.Source != "counter" {
			return nil, problem("VALIDATION_ERROR")
		}
	} else {
		b.Source = "web"
	}
	captcha := b.Captcha
	b.Captcha = ""
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SET LOCAL lock_timeout='5s'"); e != nil {
		return nil, e
	}
	cached, e := Claim(ctx, tx, p, "create-booking", key, b)
	if e != nil {
		return nil, e
	}
	if cached != nil {
		return cached, tx.Commit(ctx)
	}
	if !p.Staff && (s.Captcha == nil || !s.Captcha(ctx, captcha, "checkout")) {
		return nil, problem("CAPTCHA_FAILED")
	}
	q, e := loadQuote(ctx, tx, p, b.Quote)
	if e != nil {
		return nil, e
	}
	var status string
	var version, remaining int
	var close, now time.Time
	if e = tx.QueryRow(ctx, "SELECT status,price_version,sales_close_at FROM departures WHERE id=$1 FOR UPDATE", q.Departure).Scan(&status, &version, &close); e != nil {
		return nil, e
	}
	if e = ExpireLocked(ctx, tx, q.Departure); e != nil {
		return nil, e
	}
	if e = tx.QueryRow(ctx, "SELECT clock_timestamp(),capacity-held_seats-confirmed_seats FROM departures WHERE id=$1", q.Departure).Scan(&now, &remaining); e != nil {
		return nil, e
	}
	if !now.Before(q.Expires) {
		return nil, problem("QUOTE_EXPIRED")
	}
	if status != "open" || !now.Before(close) {
		return nil, problem("SALES_CLOSED")
	}
	if version != q.Version {
		return nil, problem("PRICE_CHANGED")
	}
	if remaining < q.Seats {
		return nil, problem("SOLD_OUT")
	}
	if !validatePassengers(b.Passengers, q) {
		return nil, problem("PASSENGER_MISMATCH")
	}
	if b.Option == "deposit" && (!now.Before(q.Due) || len(q.Options) < 2) {
		return nil, problem("PAYMENT_OPTION_UNAVAILABLE")
	}
	var used bool
	if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM bookings WHERE quote_id=$1)", q.ID).Scan(&used); e != nil {
		return nil, e
	}
	if used {
		return nil, problem("QUOTE_ALREADY_USED")
	}
	id := uuid.NewString()
	code := "BS" + sortCode(id)
	guest, staff := p.Values()
	contact, _ := json.Marshal(b.Contact)
	counts, _ := json.Marshal(q.Counts)
	snapshot, _ := json.Marshal(q.Snapshot)
	policy, _ := json.Marshal(map[string]any{"cancellation_policy": q.Snapshot.Policy, "policy_version": "booking-v1"})
	settings, _ := json.Marshal(Defaults{q.Hold, q.Max})
	consent, _ := json.Marshal(b.Consent)
	expires := now.Add(time.Duration(q.Hold) * time.Minute)
	_, e = tx.Exec(ctx, "INSERT INTO bookings(id,code,quote_id,departure_id,guest_session_id,created_by_staff_id,customer_id,source,contact,counts,seat_count,payment_option,total_amount,deposit_amount,price_snapshot,policy_snapshot,settings_snapshot,tour_snapshot,special_requests,consent,expires_at,balance_due_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)", id, code, q.ID, q.Departure, guest, staff, b.Customer, b.Source, contact, counts, q.Seats, b.Option, int64(q.Snapshot.Total), int64(q.Snapshot.Deposit), snapshot, policy, settings, q.Snapshot.Tour, b.Requests, consent, expires, q.Due)
	if e != nil {
		return nil, e
	}
	for _, line := range q.Snapshot.Lines {
		if _, e = tx.Exec(ctx, "INSERT INTO booking_items(booking_id,code,label,quantity,unit_amount,total_amount) VALUES($1,$2,$3,$4,$5,$6)", id, line.Code, line.Label, line.Quantity, int64(line.Unit), int64(line.Total)); e != nil {
			return nil, e
		}
	}
	if e = writePassengers(ctx, tx, id, b.Passengers); e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO seat_holds(booking_id,departure_id,seat_count,expires_at) VALUES($1,$2,$3,$4)", id, q.Departure, q.Seats, expires); e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, "UPDATE departures SET held_seats=held_seats+$2 WHERE id=$1", q.Departure, q.Seats); e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO booking_events(booking_id,event_type,actor_staff_id) VALUES($1,'created',$2)", id, staff); e != nil {
		return nil, e
	}
	out, e := bookingJSON(ctx, tx, id, true)
	if e != nil {
		return nil, e
	}
	if e = Remember(ctx, tx, p, "create-booking", key, out); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func sortCode(id string) string { return stringsUpper(id[:8] + id[9:13]) }
func stringsUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}
func writePassengers(ctx context.Context, tx pgx.Tx, id string, list []Passenger) error {
	if _, e := tx.Exec(ctx, "DELETE FROM booking_passengers WHERE booking_id=$1", id); e != nil {
		return e
	}
	for _, p := range list {
		if _, e := tx.Exec(ctx, "INSERT INTO booking_passengers(booking_id,passenger_type,full_name,date_of_birth) VALUES($1,$2,$3,$4)", id, p.Type, p.Name, p.Birth); e != nil {
			return e
		}
	}
	return nil
}
