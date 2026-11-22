package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const vietnamZone = "Asia/Ho_Chi_Minh"

type reportPeriod struct {
	From, To                        string
	Tour, Source, Bucket, GroupType string
}

func parseReport(c *gin.Context) (reportPeriod, bool) {
	p := reportPeriod{
		From: c.Query("from"), To: c.Query("to"), Tour: c.Query("tour_id"),
		Source: c.Query("source"), Bucket: c.DefaultQuery("group_by", "day"),
		GroupType: c.Query("group_type"),
	}
	from, startErr := time.Parse(time.DateOnly, p.From)
	to, endErr := time.Parse(time.DateOnly, p.To)
	valid := startErr == nil && endErr == nil && !to.Before(from) && to.Sub(from) < 366*24*time.Hour &&
		(p.Tour == "" || uuid.Validate(p.Tour) == nil) &&
		(p.Source == "" || p.Source == "web" || p.Source == "phone" || p.Source == "counter") &&
		(p.Bucket == "day" || p.Bucket == "month") &&
		(p.GroupType == "" || p.GroupType == "school" || p.GroupType == "company")
	if !valid {
		httpx.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	}
	return p, valid
}

const reportBounds = `WITH bounds AS (
	SELECT ($1::text::date::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh') AS lo,
	(($2::text::date+1)::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh') AS hi
)`

type revenueRow struct {
	Period   string `json:"period"`
	Gross    string `json:"gross_booking_amount"`
	Received string `json:"cash_received"`
	Refunded string `json:"refund_amount"`
	Net      string `json:"net_received"`
	Count    int64  `json:"booking_count"`
}
type revenueReport struct {
	revenueRow
	Currency string       `json:"currency"`
	Timezone string       `json:"timezone"`
	Demo     bool         `json:"is_demo"`
	Series   []revenueRow `json:"series"`
}

func (a *App) revenue(ctx context.Context, p reportPeriod) (revenueReport, error) {
	// Each money movement contributes once on its own date. Passenger and quote tables do not enter the query.
	query := reportBounds + `, eligible AS (
		SELECT b.id,b.created_at,b.status,b.total_amount
		FROM bookings b JOIN departures d ON d.id=b.departure_id
		WHERE ($3='' OR d.tour_id=NULLIF($3,'')::uuid) AND ($4='' OR b.source=$4)
	), movements AS (
		SELECT b.created_at AS stamp,b.total_amount AS gross,0::bigint AS received,0::bigint AS refunded,1 AS bookings
		FROM eligible b,bounds WHERE b.status IN ('confirmed','completed') AND b.created_at>=lo AND b.created_at<hi
		UNION ALL
		SELECT p.paid_at,0,p.amount,0,0 FROM payments p JOIN eligible b ON b.id=p.booking_id,bounds
		WHERE p.status='succeeded' AND p.is_mock AND p.paid_at>=lo AND p.paid_at<hi
		UNION ALL
		SELECT r.completed_at,0,0,r.amount,0 FROM refunds r JOIN payments p ON p.id=r.payment_id
		JOIN eligible b ON b.id=p.booking_id,bounds
		WHERE r.status='succeeded' AND r.is_mock AND r.completed_at>=lo AND r.completed_at<hi
	), grouped AS (
		SELECT date_trunc($5,stamp AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS period,gross,received,refunded,bookings
		FROM movements
	)
	SELECT COALESCE(period::text,''),coalesce(sum(gross),0)::text,coalesce(sum(received),0)::text,
	coalesce(sum(refunded),0)::text,(coalesce(sum(received),0)-coalesce(sum(refunded),0))::text,
	coalesce(sum(bookings),0) FROM grouped GROUP BY GROUPING SETS ((period),()) ORDER BY period NULLS FIRST`
	out := revenueReport{Currency: "VND", Timezone: vietnamZone, Demo: true, Series: []revenueRow{}}
	rows, err := a.DB.Query(ctx, query, p.From, p.To, p.Tour, p.Source, p.Bucket)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var row revenueRow
		if err = rows.Scan(&row.Period, &row.Gross, &row.Received, &row.Refunded, &row.Net, &row.Count); err != nil {
			return out, err
		}
		if row.Period == "" {
			out.revenueRow = row
		} else {
			out.Series = append(out.Series, row)
		}
	}
	return out, rows.Err()
}

func reportReply(c *gin.Context, result any, err error) {
	if err != nil {
		httpx.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE")
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (a *App) revenueReport(c *gin.Context) {
	p, ok := parseReport(c)
	if !ok {
		return
	}
	result, err := a.revenue(c.Request.Context(), p)
	reportReply(c, result, err)
}

func (a *App) occupancyReport(c *gin.Context) {
	p, ok := parseReport(c)
	if !ok {
		return
	}
	result, err := httpx.JSONRows(c.Request.Context(), a.DB, reportBounds+`
	SELECT jsonb_build_object('departure_id',d.id,'tour_title',t.title,'starts_at',d.starts_at,
	'capacity',d.capacity,'held_seats',h.seats,'confirmed_seats',d.confirmed_seats,
	'occupancy_percent',coalesce(round(100.0*d.confirmed_seats/nullif(d.capacity,0),2),0))
	FROM departures d JOIN tours t ON t.id=d.tour_id CROSS JOIN bounds
	CROSS JOIN LATERAL(SELECT coalesce(sum(seat_count),0) seats FROM seat_holds
		WHERE departure_id=d.id AND status='active' AND expires_at>statement_timestamp()) h
	WHERE d.starts_at>=lo AND d.starts_at<hi AND ($3='' OR t.id=NULLIF($3,'')::uuid)
	ORDER BY d.starts_at,d.id`, p.From, p.To, p.Tour)
	reportReply(c, result, err)
}

func (a *App) quoteReport(c *gin.Context) {
	p, ok := parseReport(c)
	if !ok {
		return
	}
	var result json.RawMessage
	err := a.DB.QueryRow(c.Request.Context(), reportBounds+`, cohort AS (
	SELECT status,count(*) AS n FROM group_requests,bounds WHERE created_at>=lo AND created_at<hi
	AND ($3='' OR group_type=$3) GROUP BY status
	), totals AS (
	SELECT coalesce(sum(n),0) AS requests,coalesce(sum(n) FILTER (WHERE status='won'),0) AS won,
	jsonb_build_object('new',0,'preparing',0,'quoted',0,'won',0,'lost',0)||coalesce(jsonb_object_agg(status,n),'{}') AS statuses
	FROM cohort
	)
	SELECT jsonb_build_object('request_count',requests,'by_status',statuses,'won_count',won,
	'won_rate_percent',coalesce(round(100*won/nullif(requests,0),2),0)) FROM totals`, p.From, p.To, p.GroupType).Scan(&result)
	reportReply(c, result, err)
}

func (a *App) reportTours(c *gin.Context) {
	result, err := httpx.JSONRows(c.Request.Context(), a.DB,
		"SELECT jsonb_build_object('id',id,'title',title) FROM tours ORDER BY title,id")
	reportReply(c, result, err)
}
