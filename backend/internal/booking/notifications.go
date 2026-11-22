package booking

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Server) RunNotifications(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		for range 10 {
			found, err := s.ProcessNotification(ctx)
			if err != nil && ctx.Err() == nil {
				slog.Warn("mock notification worker failed")
			}
			if !found || err != nil {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) ProcessNotification(ctx context.Context) (bool, error) {
	// The attempt number fences a worker that lost its lease after claiming.
	if _, err := s.DB.Exec(ctx, `UPDATE outbox_events SET status='failed',lease_until=NULL,last_error='LEASE_EXPIRED'
		WHERE status='processing' AND attempts=5 AND lease_until<=clock_timestamp()`); err != nil {
		return false, err
	}
	var id string
	var attempt int
	err := s.DB.QueryRow(ctx, `WITH candidate AS (
		SELECT id FROM outbox_events WHERE attempts<5 AND
		((status='pending' AND next_attempt_at<=clock_timestamp()) OR
		 (status='processing' AND lease_until<=clock_timestamp()))
		ORDER BY next_attempt_at,created_at LIMIT 1 FOR UPDATE SKIP LOCKED
	)
	UPDATE outbox_events e SET status='processing',attempts=e.attempts+1,lease_until=clock_timestamp()+interval '30 seconds'
	FROM candidate c WHERE e.id=c.id RETURNING e.id,e.attempts`).Scan(&id, &attempt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err = s.renderNotification(ctx, id, attempt); err != nil {
		_, retryErr := s.DB.Exec(ctx, `UPDATE outbox_events SET status=CASE WHEN attempts=5 THEN 'failed' ELSE 'pending' END,
			lease_until=NULL,next_attempt_at=clock_timestamp()+make_interval(secs=>power(2,attempts)::int),
			last_error='PREVIEW_FAILED' WHERE id=$1 AND attempts=$2 AND status='processing'`, id, attempt)
		return true, errors.Join(err, retryErr)
	}
	return true, nil
}

func (s *Server) renderNotification(ctx context.Context, id string, attempt int) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var bookingID, code, event, key string
	err = tx.QueryRow(ctx, `SELECT e.booking_id,b.code,e.event_type,e.event_key FROM outbox_events e
		JOIN bookings b ON b.id=e.booking_id WHERE e.id=$1 AND e.attempts=$2 AND e.status='processing'
		AND e.lease_until>clock_timestamp() FOR UPDATE OF e`, id, attempt).Scan(&bookingID, &code, &event, &key)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	subject := "Xác nhận đặt tour"
	if event == "balance_paid" {
		subject = "Xác nhận thanh toán phần còn lại"
	}
	if _, err = tx.Exec(ctx, `INSERT INTO notification_logs(booking_id,channel,template_code,subject,body,dedupe_key)
		SELECT $1,x.channel,$2,CASE WHEN x.channel='email' THEN $3::text END,$4,$5||':'||x.channel
		FROM (VALUES ('email'),('sms')) AS x(channel) ON CONFLICT(dedupe_key) DO NOTHING`,
		bookingID, event, subject, subject+" cho đơn "+code+". Đây là thông báo giả lập của Blue Sky Group.", key); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "UPDATE outbox_events SET status='done',lease_until=NULL,last_error=NULL WHERE id=$1", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Server) RegisterNotifications(v, admin *gin.RouterGroup, guest, read gin.HandlerFunc) {
	v.GET("/bookings/:id/notifications", guest, s.notifications(false))
	admin.GET("/notifications", read, s.notifications(true))
}

func (s *Server) notifications(staff bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := s.Principal(c, staff)
		if !ok {
			return
		}
		ctx := c.Request.Context()
		id := c.Param("id")
		channel, status := "", ""
		page, size := 1, 100
		if staff {
			id, channel, status = c.Query("booking_id"), c.Query("channel"), c.Query("status")
			var valid bool
			page, size, valid = httpx.Page(c)
			if !valid {
				respondError(c, problem("VALIDATION_ERROR"))
				return
			}
		} else {
			var err error
			id, err = s.paymentBooking(ctx, principal, id, false)
			if err != nil {
				respondError(c, err)
				return
			}
		}
		if (id != "" && uuid.Validate(id) != nil) || (channel != "" && channel != "email" && channel != "sms") ||
			(status != "" && status != "simulated" && status != "failed" && status != "pending" && status != "processing") {
			respondError(c, problem("VALIDATION_ERROR"))
			return
		}
		query := `SELECT to_jsonb(n)-'dedupe_key' FROM notification_logs n
			WHERE ($1='' OR booking_id=NULLIF($1,'')::uuid) AND ($2='' OR channel=$2) AND ($3='' OR status=$3)
			ORDER BY created_at DESC,id LIMIT $4 OFFSET $5`
		logs, err := httpx.JSONRows(ctx, s.DB, query, id, channel, status, size, (page-1)*size)
		if err != nil {
			respondError(c, err)
			return
		}
		if !staff {
			httpx.OK(c, 200, logs)
			return
		}
		jobs, err := httpx.JSONRows(ctx, s.DB, `SELECT jsonb_build_object('id',id,'booking_id',booking_id,
			'template_code',event_type,'status',status,'attempts',attempts,'error_code',last_error,'created_at',created_at)
			FROM outbox_events WHERE ($1='' OR booking_id=NULLIF($1,'')::uuid)
			AND ($2='' OR status=$2) AND status<>'done' ORDER BY created_at DESC,id LIMIT $3 OFFSET $4`, id, status, size, (page-1)*size)
		if err != nil {
			respondError(c, err)
			return
		}
		httpx.OK(c, 200, gin.H{"logs": logs, "jobs": jobs})
	}
}
