package booking

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func createPayment(ctx context.Context, tx pgx.Tx, bookingID, purpose string, staffID any) (json.RawMessage, error) {
	var status, option string
	var total, deposit, paid int64
	var holdUntil, now, starts time.Time
	err := tx.QueryRow(ctx, `SELECT b.status,b.payment_option,b.total_amount,b.deposit_amount,b.paid_amount,
		b.expires_at,clock_timestamp(),d.starts_at FROM bookings b JOIN departures d ON d.id=b.departure_id WHERE b.id=$1`,
		bookingID).Scan(&status, &option, &total, &deposit, &paid, &holdUntil, &now, &starts)
	if err != nil {
		return nil, err
	}
	if status == "expired" || (status == "pending" && !now.Before(holdUntil)) {
		return nil, problem("HOLD_EXPIRED")
	}
	amount, expires := total, now.Add(15*time.Minute)
	switch {
	case status == "pending" && purpose == option:
		if purpose == "deposit" {
			amount = deposit
		}
		expires = minTime(expires, holdUntil)
	case status == "confirmed" && purpose == "balance" && paid < total && now.Before(starts):
		amount = total - paid
		expires = minTime(expires, starts)
	default:
		return nil, problem("INVALID_STATE")
	}
	_, err = tx.Exec(ctx, "UPDATE payments SET status='cancelled' WHERE booking_id=$1 AND status='pending' AND expires_at<=$2", bookingID, now)
	if err != nil {
		return nil, err
	}
	var out json.RawMessage
	err = tx.QueryRow(ctx, "INSERT INTO payments AS p(booking_id,purpose,amount,expires_at,initiated_by_staff_id) VALUES($1,$2,$3,$4,$5) ON CONFLICT(booking_id) WHERE status='pending' DO NOTHING RETURNING "+paymentJSON,
		bookingID, purpose, amount, expires, staffID).Scan(&out)
	if errors.Is(err, pgx.ErrNoRows) {
		err = problem("PAYMENT_IN_PROGRESS")
	}
	return out, err
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func confirmPayment(ctx context.Context, tx pgx.Tx, bookingID, departure, id string, staffID any, reason, requestID string) (json.RawMessage, error) {
	var bookingStatus, option, paymentStatus, purpose string
	var total, deposit, paid, amount int64
	var seats int
	var expires, holdUntil, starts, now time.Time
	err := tx.QueryRow(ctx, `SELECT b.status,b.payment_option,b.total_amount,b.deposit_amount,b.paid_amount,b.seat_count,
		b.expires_at,p.status,p.purpose,p.amount,p.expires_at,d.starts_at,clock_timestamp()
		FROM bookings b JOIN payments p ON p.booking_id=b.id JOIN departures d ON d.id=b.departure_id
		WHERE b.id=$1 AND p.id=$2`, bookingID, id).
		Scan(&bookingStatus, &option, &total, &deposit, &paid, &seats, &holdUntil, &paymentStatus, &purpose, &amount, &expires, &starts, &now)
	if err != nil {
		return nil, err
	}
	if paymentStatus != "succeeded" {
		if bookingStatus == "expired" || (bookingStatus == "pending" && !now.Before(holdUntil)) {
			return nil, problem("HOLD_EXPIRED")
		}
		if !now.Before(expires) {
			return nil, problem("PAYMENT_EXPIRED")
		}
		event := "balance_paid"
		switch {
		case bookingStatus == "pending" && purpose == option && paid == 0 &&
			((purpose == "full" && amount == total) || (purpose == "deposit" && amount == deposit)):
			event = "booking_confirmed"
			tag, err := tx.Exec(ctx, `UPDATE seat_holds SET status='converted',updated_at=$2
				WHERE booking_id=$1 AND status='active' AND expires_at>$2`, bookingID, now)
			if err != nil {
				return nil, err
			}
			if tag.RowsAffected() != 1 {
				return nil, problem("HOLD_EXPIRED")
			}
			if _, err = tx.Exec(ctx, `UPDATE departures SET held_seats=held_seats-$2,confirmed_seats=confirmed_seats+$2
				WHERE id=$1`, departure, seats); err != nil {
				return nil, err
			}
		case bookingStatus == "confirmed" && purpose == "balance" && amount == total-paid && now.Before(starts):
		default:
			return nil, problem("INVALID_STATE")
		}
		if paymentStatus != "pending" {
			return nil, problem("INVALID_STATE")
		}
		batch := &pgx.Batch{}
		batch.Queue("UPDATE payments SET status='succeeded',paid_at=$2 WHERE id=$1", id, now)
		batch.Queue(`UPDATE bookings SET status='confirmed',confirmed_at=COALESCE(confirmed_at,$3),paid_amount=$2,
			payment_status=CASE WHEN refunded_amount>0 THEN 'partially_refunded' WHEN $2=total_amount THEN 'paid' ELSE 'partially_paid' END,
			version=version+1,updated_at=$3 WHERE id=$1`, bookingID, paid+amount, now)
		batch.Queue("INSERT INTO payment_events(payment_id,event_key,event_type) VALUES($1,'confirm:'||($1::uuid)::text,'mock_confirmed')", id)
		batch.Queue("INSERT INTO booking_events(booking_id,event_type,actor_staff_id) VALUES($1,$2,$3)", bookingID, event, staffID)
		batch.Queue("INSERT INTO outbox_events(event_key,booking_id,event_type) VALUES($1,$2,$3)", "payment:"+id, bookingID, event)
		if staffID != nil {
			batch.Queue("INSERT INTO audit_logs(actor_staff_id,action,entity_type,entity_id,request_id,reason) VALUES($1,'payment.mock_confirm','payment',$2,$3,$4)", staffID, id, requestID, reason)
		}
		if err = tx.SendBatch(ctx, batch).Close(); err != nil {
			return nil, err
		}
	}
	var payment json.RawMessage
	if err = tx.QueryRow(ctx, "SELECT "+paymentJSON+" FROM payments p WHERE id=$1", id).Scan(&payment); err != nil {
		return nil, err
	}
	booking, err := bookingJSON(ctx, tx, bookingID, true)
	if err != nil {
		return nil, err
	}
	return json.Marshal(gin.H{"payment": payment, "booking": booking})
}

func refundPayment(ctx context.Context, tx pgx.Tx, bookingID, id, staffID, reason string, amount int64, requestID string) (json.RawMessage, error) {
	var status string
	var available int64
	err := tx.QueryRow(ctx, `SELECT p.status,p.amount-COALESCE((SELECT sum(r.amount) FROM refunds r WHERE r.payment_id=p.id),0)
		FROM payments p WHERE p.id=$1 FOR UPDATE`, id).Scan(&status, &available)
	if err != nil {
		return nil, err
	}
	if status != "succeeded" || amount > available {
		return nil, problem("REFUND_UNAVAILABLE")
	}
	var out json.RawMessage
	var refundID string
	err = tx.QueryRow(ctx, "INSERT INTO refunds AS r(payment_id,amount,reason,requested_by_staff_id) VALUES($1,$2,$3,$4) RETURNING r.id,"+refundJSON,
		id, amount, reason, staffID).Scan(&refundID, &out)
	if err != nil {
		return nil, err
	}
	batch := &pgx.Batch{}
	batch.Queue(`UPDATE bookings SET refunded_amount=refunded_amount+$2,
		payment_status=CASE WHEN refunded_amount+$2=paid_amount THEN 'refunded' ELSE 'partially_refunded' END,
		version=version+1,updated_at=clock_timestamp() WHERE id=$1`, bookingID, amount)
	batch.Queue("INSERT INTO audit_logs(actor_staff_id,action,entity_type,entity_id,request_id,reason) VALUES($1,'payment.mock_refund','refund',$2,$3,$4)", staffID, refundID, requestID, reason)
	batch.Queue("INSERT INTO booking_events(booking_id,event_type,actor_staff_id) VALUES($1,'refunded',$2)", bookingID, staffID)
	return out, tx.SendBatch(ctx, batch).Close()
}
