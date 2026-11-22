package booking

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const paymentJSON = "to_jsonb(p)-'initiated_by_staff_id'||jsonb_build_object('amount',p.amount::text)"
const refundJSON = "to_jsonb(r)-'requested_by_staff_id'||jsonb_build_object('amount',r.amount::text)"

func (s *Server) RegisterPayments(v, admin *gin.RouterGroup, guest, origin, limit, read, write gin.HandlerFunc) {
	v.GET("/bookings/:id/payments", guest, s.paymentList(false))
	v.POST("/bookings/:id/payments", origin, guest, limit, s.paymentAction(false, "intent", 201))
	v.POST("/payments/:id/mock-confirm", origin, guest, limit, s.paymentAction(false, "confirm", 200))
	admin.GET("/bookings/:id/payments", read, s.paymentList(true))
	admin.POST("/bookings/:id/payments", write, s.paymentAction(true, "intent", 201))
	admin.POST("/payments/:id/mock-confirm", write, s.paymentAction(true, "confirm", 200))
	admin.POST("/payments/:id/refunds", write, s.paymentAction(true, "refund", 201))
}

func (s *Server) paymentBooking(ctx context.Context, principal Principal, id string, byPayment bool) (string, error) {
	if uuid.Validate(id) != nil {
		return "", ErrNotFound
	}
	query := "SELECT b.id FROM bookings b WHERE b.id=$1"
	if byPayment {
		query = "SELECT b.id FROM payments p JOIN bookings b ON b.id=p.booking_id WHERE p.id=$1"
	}
	args := []any{id}
	if !principal.Staff {
		query += " AND b.guest_session_id=$2"
		args = append(args, principal.ID)
	}
	var bookingID string
	err := s.DB.QueryRow(ctx, query, args...).Scan(&bookingID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return bookingID, err
}

func (s *Server) paymentList(staff bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := s.Principal(c, staff)
		if !ok {
			return
		}
		ctx := c.Request.Context()
		id, err := s.paymentBooking(ctx, principal, c.Param("id"), false)
		if err != nil {
			respondError(c, err)
			return
		}
		var out json.RawMessage
		query := "SELECT jsonb_build_object('server_time',clock_timestamp(),'payments',COALESCE((SELECT jsonb_agg(" + paymentJSON +
			" ORDER BY p.created_at) FROM payments p WHERE p.booking_id=b.id),'[]')"
		if staff {
			query += ",'refunds',COALESCE((SELECT jsonb_agg(" + refundJSON +
				" ORDER BY r.created_at) FROM refunds r JOIN payments p ON p.id=r.payment_id WHERE p.booking_id=b.id),'[]')," +
				"'reconciled',b.paid_amount=COALESCE((SELECT sum(p.amount) FROM payments p WHERE p.booking_id=b.id AND p.status='succeeded'),0) AND " +
				"b.refunded_amount=COALESCE((SELECT sum(r.amount) FROM refunds r JOIN payments p ON p.id=r.payment_id WHERE p.booking_id=b.id),0)"
		}
		err = s.DB.QueryRow(ctx, query+") FROM bookings b WHERE b.id=$1", id).Scan(&out)
		if err != nil {
			respondError(c, err)
			return
		}
		httpx.OK(c, 200, out)
	}
}

func (s *Server) paymentAction(staff bool, action string, status int) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := s.Principal(c, staff)
		if !ok {
			return
		}
		var purpose, reason string
		var amount int64
		switch action {
		case "intent":
			var input struct {
				Purpose string `json:"purpose"`
			}
			if !httpx.Body(c, &input) {
				return
			}
			purpose = input.Purpose
			if purpose != "full" && purpose != "deposit" && purpose != "balance" {
				respondError(c, problem("VALIDATION_ERROR"))
				return
			}
		case "confirm":
			if !staff {
				var input struct{}
				if !httpx.Body(c, &input) {
					return
				}
				break
			}
			var input struct {
				Reason string `json:"reason"`
			}
			if !httpx.Body(c, &input) {
				return
			}
			reason = strings.TrimSpace(input.Reason)
		case "refund":
			var input struct {
				Amount string `json:"amount"`
				Reason string `json:"reason"`
			}
			if !httpx.Body(c, &input) {
				return
			}
			var err error
			amount, err = strconv.ParseInt(input.Amount, 10, 64)
			if err != nil || amount <= 0 || input.Amount != strconv.FormatInt(amount, 10) {
				respondError(c, problem("VALIDATION_ERROR"))
				return
			}
			reason = strings.TrimSpace(input.Reason)
		}
		if staff && action != "intent" && (len([]rune(reason)) < 2 || len([]rune(reason)) > 1000) {
			respondError(c, problem("VALIDATION_ERROR"))
			return
		}
		ctx := c.Request.Context()
		id := c.Param("id")
		bookingID, err := s.paymentBooking(ctx, principal, id, action != "intent")
		if err != nil {
			respondError(c, err)
			return
		}
		tx, err := s.DB.Begin(ctx)
		if err != nil {
			respondError(c, err)
			return
		}
		defer tx.Rollback(ctx)
		route, key := "payment."+action+":"+id, c.GetHeader("Idempotency-Key")
		out, err := claim(ctx, tx, principal, route, key, []any{purpose, amount, reason})
		if err != nil {
			respondError(c, err)
			return
		}
		if out != nil {
			httpx.OK(c, status, out)
			return
		}
		departure, err := s.lockBooking(ctx, tx, bookingID)
		if err == nil {
			_, staffID := principal.Values()
			switch action {
			case "intent":
				out, err = createPayment(ctx, tx, bookingID, purpose, staffID)
			case "confirm":
				out, err = confirmPayment(ctx, tx, bookingID, departure, id, staffID, reason, c.GetString("request_id"))
			case "refund":
				out, err = refundPayment(ctx, tx, bookingID, id, principal.ID, reason, amount, c.GetString("request_id"))
			}
		}
		if err == nil {
			err = remember(ctx, tx, principal, route, key, out)
		}
		if err == nil {
			err = tx.Commit(ctx)
		}
		if err != nil {
			respondError(c, err)
			return
		}
		httpx.OK(c, status, out)
	}
}
