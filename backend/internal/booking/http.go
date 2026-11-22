package booking

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
	"time"
)

func (s *Server) Register(v, admin *gin.RouterGroup, guest, origin, lookupLimit, checkoutLimit, read, write gin.HandlerFunc) {
	v.POST("/booking-quotes", origin, guest, checkoutLimit, s.quoteHTTP(false))
	v.POST("/bookings", origin, guest, checkoutLimit, s.createHTTP(false))
	v.GET("/bookings/:id", guest, s.readHTTP(false))
	v.POST("/booking-lookups", origin, lookupLimit, s.lookup)
	v.GET("/booking-lookups/current", s.lookupCurrent)
	admin.POST("/booking-quotes", write, s.quoteHTTP(true))
	admin.POST("/bookings", write, s.createHTTP(true))
	admin.GET("/bookings", read, s.list)
	admin.GET("/bookings/:id", read, s.readHTTP(true))
	admin.PATCH("/bookings/:id", write, s.patch)
	admin.POST("/bookings/:id/notes", write, s.note)
	admin.POST("/bookings/:id/cancel", write, s.transition("cancelled"))
	admin.POST("/bookings/:id/complete", write, s.transition("completed"))
}
func respondError(c *gin.Context, e error) {
	code := e.Error()
	var pgError *pgconn.PgError
	if errors.As(e, &pgError) && pgError.Code == "55P03" {
		code = "REQUEST_IN_PROGRESS"
	}
	status := 409
	switch code {
	case "NOT_FOUND", "BOOKING_NOT_FOUND":
		status = 404
	case "CAPTCHA_FAILED", "FORBIDDEN":
		status = 403
	case "VALIDATION_ERROR", "PASSENGER_MISMATCH", "IDEMPOTENCY_KEY_REQUIRED":
		status = 422
	case "HOLD_EXPIRED", "PAYMENT_EXPIRED", "PAYMENT_IN_PROGRESS", "INVALID_STATE", "REFUND_UNAVAILABLE", "SOLD_OUT", "QUOTE_EXPIRED", "PRICE_CHANGED", "QUOTE_ALREADY_USED", "SALES_CLOSED", "PAYMENT_OPTION_UNAVAILABLE", "IDEMPOTENCY_CONFLICT", "REQUEST_IN_PROGRESS", "VERSION_CONFLICT", "INVALID_TRANSITION":
	default:
		status = 503
		code = "SERVICE_UNAVAILABLE"
	}
	httpx.Error(c, status, code)
}
func (s *Server) quoteHTTP(staff bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, valid := s.Principal(c, staff)
		if !valid {
			return
		}
		var b QuoteInput
		if !httpx.Body(c, &b) {
			return
		}
		q, e := s.Quote(c.Request.Context(), p, b)
		if e != nil {
			respondError(c, e)
			return
		}
		httpx.OK(c, 201, q)
	}
}
func (s *Server) createHTTP(staff bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, valid := s.Principal(c, staff)
		if !valid {
			return
		}
		var b CreateInput
		if !httpx.Body(c, &b) {
			return
		}
		out, e := s.Create(c.Request.Context(), p, c.GetHeader("Idempotency-Key"), b)
		if e != nil {
			respondError(c, e)
			return
		}
		httpx.OK(c, 201, out)
	}
}

const baseJSON = `(to_jsonb(b)-'guest_session_id'-'created_by_staff_id'-'customer_id'-'customer_verified_at'-'customer_verified_by_staff_id')||jsonb_build_object('total_amount',b.total_amount::text,'deposit_amount',b.deposit_amount::text,'paid_amount',b.paid_amount::text,'refunded_amount',b.refunded_amount::text,'net_received',(b.paid_amount-b.refunded_amount)::text,'balance_amount',(b.total_amount-b.paid_amount)::text,'server_time',clock_timestamp())`

func bookingJSON(ctx context.Context, tx pgx.Tx, id string, detail bool) (json.RawMessage, error) {
	sql := "SELECT " + baseJSON
	if detail {
		sql += `||jsonb_build_object('passengers',COALESCE((SELECT jsonb_agg(jsonb_build_object('passenger_type',p.passenger_type,'full_name',p.full_name,'date_of_birth',p.date_of_birth) ORDER BY p.id) FROM booking_passengers p WHERE p.booking_id=b.id),'[]'),'items',COALESCE((SELECT jsonb_agg(to_jsonb(i)||jsonb_build_object('unit_amount',i.unit_amount::text,'total_amount',i.total_amount::text)) FROM booking_items i WHERE i.booking_id=b.id),'[]'))`
	}
	sql += " FROM bookings b WHERE b.id=$1"
	var out json.RawMessage
	e := tx.QueryRow(ctx, sql, id).Scan(&out)
	return out, e
}
func (s *Server) readHTTP(staff bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if uuid.Validate(id) != nil {
			respondError(c, ErrNotFound)
			return
		}
		p, valid := s.Principal(c, staff)
		if !valid {
			return
		}
		ctx := c.Request.Context()
		var dep string
		var guest *string
		if s.DB.QueryRow(ctx, "SELECT departure_id,guest_session_id FROM bookings WHERE id=$1", id).Scan(&dep, &guest) != nil || !staff && (guest == nil || *guest != p.ID) {
			respondError(c, ErrNotFound)
			return
		}
		if e := s.Expire(ctx, dep); e != nil {
			respondError(c, e)
			return
		}
		tx, e := s.DB.Begin(ctx)
		if e != nil {
			respondError(c, e)
			return
		}
		defer tx.Rollback(ctx)
		out, e := bookingJSON(ctx, tx, id, true)
		if e != nil {
			respondError(c, e)
			return
		}
		if staff {
			var data map[string]any
			json.Unmarshal(out, &data)
			events, e := httpx.JSONRows(ctx, s.DB, "SELECT to_jsonb(e) FROM booking_events e WHERE booking_id=$1 ORDER BY created_at", id)
			if e != nil {
				respondError(c, e)
				return
			}
			notes, e := httpx.JSONRows(ctx, s.DB, "SELECT to_jsonb(n) FROM booking_notes n WHERE booking_id=$1 ORDER BY created_at", id)
			if e != nil {
				respondError(c, e)
				return
			}
			data["events"] = events
			data["notes"] = notes
			httpx.OK(c, 200, data)
			return
		}
		httpx.OK(c, 200, out)
	}
}
func (s *Server) list(c *gin.Context) {
	p, n, valid := httpx.Page(c)
	if !valid {
		respondError(c, problem("VALIDATION_ERROR"))
		return
	}
	where := []string{"true"}
	args := []any{}
	allowed := map[string]bool{"page": true, "page_size": true, "q": true}
	for _, k := range []string{"status", "payment_status", "tour_id", "departure_id", "source", "from", "to"} {
		allowed[k] = true
	}
	for k := range c.Request.URL.Query() {
		if !allowed[k] {
			respondError(c, problem("VALIDATION_ERROR"))
			return
		}
	}
	for _, k := range []string{"status", "payment_status", "departure_id", "source"} {
		if v := c.Query(k); v != "" {
			if k == "departure_id" && uuid.Validate(v) != nil {
				respondError(c, problem("VALIDATION_ERROR"))
				return
			}
			args = append(args, v)
			where = append(where, fmt.Sprintf("b.%s=$%d", k, len(args)))
		}
	}
	if v := c.Query("tour_id"); v != "" {
		if uuid.Validate(v) != nil {
			respondError(c, problem("VALIDATION_ERROR"))
			return
		}
		args = append(args, v)
		where = append(where, fmt.Sprintf("b.departure_id IN(SELECT id FROM departures WHERE tour_id=$%d)", len(args)))
	}
	for _, k := range []string{"from", "to"} {
		if v := c.Query(k); v != "" {
			if _, e := time.Parse("2006-01-02", v); e != nil {
				respondError(c, problem("VALIDATION_ERROR"))
				return
			}
			args = append(args, v)
			op := ">="
			if k == "to" {
				op = "<="
			}
			where = append(where, fmt.Sprintf("(b.created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date %s $%d::date", op, len(args)))
		}
	}
	if c.Query("from") > c.Query("to") && c.Query("to") != "" {
		respondError(c, problem("VALIDATION_ERROR"))
		return
	}
	if q := c.Query("q"); q != "" {
		if len(q) > 100 {
			respondError(c, problem("VALIDATION_ERROR"))
			return
		}
		args = append(args, "%"+q+"%")
		where = append(where, fmt.Sprintf("(b.code ILIKE $%d OR b.contact->>'name' ILIKE $%d)", len(args), len(args)))
	}
	if e := s.Sweep(c.Request.Context()); e != nil {
		respondError(c, e)
		return
	}
	clause := " WHERE " + strings.Join(where, " AND ")
	var total int
	if e := s.DB.QueryRow(c.Request.Context(), "SELECT count(*) FROM bookings b"+clause, args...).Scan(&total); e != nil {
		respondError(c, e)
		return
	}
	args = append(args, n, (p-1)*n)
	out, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT "+baseJSON+"-'special_requests'-'consent'-'price_snapshot'-'policy_snapshot'-'settings_snapshot' FROM bookings b"+clause+fmt.Sprintf(" ORDER BY b.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if e != nil {
		respondError(c, e)
		return
	}
	httpx.List(c, out, p, n, total)
}
func (s *Server) lockBooking(ctx context.Context, tx pgx.Tx, id string) (string, error) {
	if uuid.Validate(id) != nil {
		return "", ErrNotFound
	}
	var dep string
	if tx.QueryRow(ctx, "SELECT departure_id FROM bookings WHERE id=$1", id).Scan(&dep) != nil {
		return "", ErrNotFound
	}
	var locked string
	if e := tx.QueryRow(ctx, "SELECT id FROM departures WHERE id=$1 FOR UPDATE", dep).Scan(&locked); e != nil {
		return "", e
	}
	if e := ExpireLocked(ctx, tx, dep); e != nil {
		return "", e
	}
	if e := tx.QueryRow(ctx, "SELECT id FROM bookings WHERE id=$1 FOR UPDATE", id).Scan(&locked); e != nil {
		return "", e
	}
	return dep, nil
}
func (s *Server) transition(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, valid := s.Principal(c, true)
		if !valid {
			return
		}
		var b struct {
			Version int    `json:"expected_version"`
			Reason  string `json:"reason"`
		}
		if !httpx.Body(c, &b) {
			return
		}
		if target == "cancelled" && (len(strings.TrimSpace(b.Reason)) < 2 || len(b.Reason) > 1000) {
			respondError(c, problem("VALIDATION_ERROR"))
			return
		}
		ctx := c.Request.Context()
		tx, e := s.DB.Begin(ctx)
		if e != nil {
			respondError(c, e)
			return
		}
		defer tx.Rollback(ctx)
		key := c.GetHeader("Idempotency-Key")
		route := target + ":" + c.Param("id")
		if target == "cancelled" {
			cached, err := Claim(ctx, tx, p, route, key, b)
			if err != nil {
				respondError(c, err)
				return
			}
			if cached != nil {
				httpx.OK(c, 200, cached)
				return
			}
		}
		dep, e := s.lockBooking(ctx, tx, c.Param("id"))
		if e != nil {
			respondError(c, e)
			return
		}
		var status string
		var version, seats int
		var paid, refunded int64
		var ended bool
		if e = tx.QueryRow(ctx, "SELECT b.status,b.version,b.seat_count,b.paid_amount,b.refunded_amount,d.ends_at<=clock_timestamp() FROM bookings b JOIN departures d ON d.id=b.departure_id WHERE b.id=$1", c.Param("id")).Scan(&status, &version, &seats, &paid, &refunded, &ended); e != nil {
			respondError(c, e)
			return
		}
		if status != target {
			if version != b.Version {
				respondError(c, problem("VERSION_CONFLICT"))
				return
			}
			if target == "cancelled" {
				if status != "pending" && status != "confirmed" {
					respondError(c, problem("INVALID_TRANSITION"))
					return
				}
				if status == "pending" {
					_, e = tx.Exec(ctx, "UPDATE seat_holds SET status='released',updated_at=clock_timestamp() WHERE booking_id=$1 AND status='active'", c.Param("id"))
					if e == nil {
						_, e = tx.Exec(ctx, "UPDATE departures SET held_seats=held_seats-$2 WHERE id=$1", dep, seats)
					}
				} else {
					_, e = tx.Exec(ctx, "UPDATE departures SET confirmed_seats=confirmed_seats-$2 WHERE id=$1", dep, seats)
				}
			} else if status != "confirmed" || !ended {
				respondError(c, problem("INVALID_TRANSITION"))
				return
			}
			if e != nil {
				respondError(c, e)
				return
			}
			_, e = tx.Exec(ctx, "UPDATE bookings SET status=$2,version=version+1,updated_at=clock_timestamp(),cancel_reason=CASE WHEN $2='cancelled' THEN $3 ELSE cancel_reason END,cancelled_at=CASE WHEN $2='cancelled' THEN clock_timestamp() ELSE cancelled_at END WHERE id=$1", c.Param("id"), target, b.Reason)
			if e == nil {
				_, e = tx.Exec(ctx, "INSERT INTO booking_events(booking_id,event_type,actor_staff_id,payload) VALUES($1,$2,$3,jsonb_build_object('reason',$4::text))", c.Param("id"), target, p.ID, b.Reason)
			}
			if e == nil {
				_, e = tx.Exec(ctx, "INSERT INTO audit_logs(actor_staff_id,action,entity_type,entity_id,request_id,reason) VALUES($1,$2,'booking',$3,$4,$5)", p.ID, "booking."+target, c.Param("id"), c.GetString("request_id"), b.Reason)
			}
			if e != nil {
				respondError(c, e)
				return
			}
		}
		out, e := bookingJSON(ctx, tx, c.Param("id"), true)
		if e != nil {
			respondError(c, e)
			return
		}
		if target == "cancelled" {
			out, _ = json.Marshal(gin.H{"booking": out, "refund_required_amount": fmt.Sprint(paid - refunded)})
			if e = Remember(ctx, tx, p, route, key, out); e != nil {
				respondError(c, e)
				return
			}
		}
		if e = tx.Commit(ctx); e != nil {
			respondError(c, e)
			return
		}
		httpx.OK(c, 200, out)
	}
}
func (s *Server) patch(c *gin.Context) {
	p, valid := s.Principal(c, true)
	if !valid {
		return
	}
	var b struct {
		Version          int          `json:"expected_version"`
		Contact          *Contact     `json:"contact"`
		Passengers       *[]Passenger `json:"passengers"`
		Requests         *string      `json:"special_requests"`
		Customer         *string      `json:"customer_id"`
		CustomerVerified bool         `json:"customer_link_verified"`
	}
	if !httpx.Body(c, &b) {
		return
	}
	if b.Contact != nil && !NormalizeContact(b.Contact) || b.Requests != nil && len(*b.Requests) > 5000 ||
		b.CustomerVerified != (b.Customer != nil) || b.Customer != nil && uuid.Validate(*b.Customer) != nil {
		respondError(c, problem("VALIDATION_ERROR"))
		return
	}
	ctx := c.Request.Context()
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		respondError(c, e)
		return
	}
	defer tx.Rollback(ctx)
	if _, e = s.lockBooking(ctx, tx, c.Param("id")); e != nil {
		respondError(c, e)
		return
	}
	var version int
	var status, quoteID string
	if e = tx.QueryRow(ctx, "SELECT version,status,quote_id FROM bookings WHERE id=$1", c.Param("id")).Scan(&version, &status, &quoteID); e != nil {
		respondError(c, e)
		return
	}
	if version != b.Version {
		respondError(c, problem("VERSION_CONFLICT"))
		return
	}
	linkOnly := b.Customer != nil && b.Contact == nil && b.Passengers == nil && b.Requests == nil
	if status != "pending" && status != "confirmed" && !linkOnly {
		respondError(c, problem("INVALID_TRANSITION"))
		return
	}
	if b.Customer != nil {
		if e = authorizeCustomerLink(ctx, tx, p, *b.Customer); e == nil {
			e = recordCustomerLink(ctx, tx, c.Param("id"), *b.Customer, p.ID)
		}
		if e != nil {
			respondError(c, e)
			return
		}
	}
	if b.Passengers != nil {
		var snapshot, counts []byte
		if e = tx.QueryRow(ctx, "SELECT price_snapshot,counts FROM bookings WHERE id=$1", c.Param("id")).Scan(&snapshot, &counts); e != nil {
			respondError(c, e)
			return
		}
		q := Quote{}
		json.Unmarshal(snapshot, &q.Snapshot)
		json.Unmarshal(counts, &q.Counts)
		if !validatePassengers(*b.Passengers, q) {
			respondError(c, problem("PASSENGER_MISMATCH"))
			return
		}
		if e = writePassengers(ctx, tx, c.Param("id"), *b.Passengers); e != nil {
			respondError(c, e)
			return
		}
	}
	var contact any
	if b.Contact != nil {
		contact, _ = json.Marshal(b.Contact)
	}
	_, e = tx.Exec(ctx, "UPDATE bookings SET contact=COALESCE($2,contact),special_requests=COALESCE($3,special_requests),version=version+1,updated_at=clock_timestamp() WHERE id=$1", c.Param("id"), contact, b.Requests)
	if e == nil {
		_, e = tx.Exec(ctx, "INSERT INTO booking_events(booking_id,event_type,actor_staff_id) VALUES($1,'updated',$2)", c.Param("id"), p.ID)
	}
	if e != nil {
		respondError(c, e)
		return
	}
	out, e := bookingJSON(ctx, tx, c.Param("id"), true)
	if e != nil {
		respondError(c, e)
		return
	}
	if e = tx.Commit(ctx); e != nil {
		respondError(c, e)
		return
	}
	httpx.OK(c, 200, out)
}
func (s *Server) note(c *gin.Context) {
	p, valid := s.Principal(c, true)
	if !valid {
		return
	}
	var b struct {
		Body string `json:"body"`
	}
	if !httpx.Body(c, &b) {
		return
	}
	if len(strings.TrimSpace(b.Body)) < 1 || len(b.Body) > 5000 || uuid.Validate(c.Param("id")) != nil {
		respondError(c, problem("VALIDATION_ERROR"))
		return
	}
	_, e := s.DB.Exec(c.Request.Context(), "INSERT INTO booking_notes(booking_id,staff_id,body) VALUES($1,$2,$3)", c.Param("id"), p.ID, b.Body)
	if e != nil {
		respondError(c, ErrNotFound)
		return
	}
	httpx.OK(c, 201, gin.H{"saved": true})
}
func tokenHash(t string) string { h := sha256.Sum256([]byte(t)); return hex.EncodeToString(h[:]) }
func (s *Server) summary(ctx context.Context, id string) (gin.H, error) {
	var code, status, payment, option, total, deposit, paid string
	var contact Contact
	var raw []byte
	var tour json.RawMessage
	var expires, now time.Time
	e := s.DB.QueryRow(ctx, "SELECT code,status,payment_status,payment_option,total_amount::text,deposit_amount::text,paid_amount::text,contact,tour_snapshot,expires_at,clock_timestamp() FROM bookings WHERE id=$1", id).Scan(&code, &status, &payment, &option, &total, &deposit, &paid, &raw, &tour, &expires, &now)
	if e != nil {
		return nil, ErrNotFound
	}
	json.Unmarshal(raw, &contact)
	email := "***"
	if at := strings.Index(contact.Email, "@"); at > 0 {
		email = contact.Email[:1] + "***" + contact.Email[at:]
	}
	phone := "***"
	if len(contact.Phone) > 4 {
		phone = "***" + contact.Phone[len(contact.Phone)-4:]
	}
	return gin.H{"code": code, "status": status, "payment_status": payment, "payment_option": option, "total_amount": total, "deposit_amount": deposit, "paid_amount": paid, "contact_masked": gin.H{"email": email, "phone": phone}, "tour": tour, "expires_at": expires, "server_time": now, "is_demo": true}, nil
}
func (s *Server) lookup(c *gin.Context) {
	var b struct {
		Code    string `json:"code"`
		Contact string `json:"contact"`
		Captcha string `json:"turnstile_token"`
	}
	if !httpx.Body(c, &b) {
		return
	}
	if len(b.Code) > 24 || len(b.Contact) > 254 {
		respondError(c, problem("VALIDATION_ERROR"))
		return
	}
	if s.Captcha == nil || !s.Captcha(c.Request.Context(), b.Captcha, "lookup") {
		respondError(c, problem("CAPTCHA_FAILED"))
		return
	}
	contact := strings.ToLower(strings.TrimSpace(b.Contact))
	if !strings.Contains(contact, "@") {
		contact = NormalizePhone(contact)
	}
	var id, dep string
	ctx := c.Request.Context()
	if s.DB.QueryRow(ctx, "SELECT id,departure_id FROM bookings WHERE code=$1 AND (contact->>'email'=$2 OR contact->>'phone'=$2)", strings.ToUpper(strings.TrimSpace(b.Code)), contact).Scan(&id, &dep) != nil {
		respondError(c, problem("BOOKING_NOT_FOUND"))
		return
	}
	if e := s.Expire(ctx, dep); e != nil {
		respondError(c, e)
		return
	}
	raw := make([]byte, 32)
	if _, e := rand.Read(raw); e != nil {
		respondError(c, e)
		return
	}
	token := hex.EncodeToString(raw)
	var expires time.Time
	if e := s.DB.QueryRow(ctx, "INSERT INTO booking_lookup_tokens(token_hash,booking_id,expires_at) VALUES($1,$2,clock_timestamp()+interval '15 minutes') RETURNING expires_at", tokenHash(token), id).Scan(&expires); e != nil {
		respondError(c, e)
		return
	}
	summary, e := s.summary(ctx, id)
	if e != nil {
		respondError(c, e)
		return
	}
	httpx.OK(c, 200, gin.H{"booking": summary, "lookup_token": token, "expires_at": expires})
}
func (s *Server) lookupCurrent(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	var id, dep string
	ctx := c.Request.Context()
	e := s.DB.QueryRow(ctx, "SELECT b.id,b.departure_id FROM booking_lookup_tokens l JOIN bookings b ON b.id=l.booking_id WHERE l.token_hash=$1 AND l.expires_at>clock_timestamp()", tokenHash(token)).Scan(&id, &dep)
	if e != nil {
		httpx.Error(c, 401, "UNAUTHORIZED")
		return
	}
	if e = s.Expire(ctx, dep); e != nil {
		respondError(c, e)
		return
	}
	out, e := s.summary(ctx, id)
	if e != nil {
		respondError(c, e)
		return
	}
	httpx.OK(c, 200, out)
}
