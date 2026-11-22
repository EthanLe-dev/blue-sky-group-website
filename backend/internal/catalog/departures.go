package catalog

import (
	"context"
	"encoding/json"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strconv"
	"strings"
	"time"
)

type Money int64

func (m Money) MarshalJSON() ([]byte, error) { return json.Marshal(strconv.FormatInt(int64(m), 10)) }
func (m *Money) UnmarshalJSON(b []byte) error {
	var v string
	if e := json.Unmarshal(b, &v); e != nil {
		return e
	}
	n, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		return e
	}
	if n < 0 || n > 1000000000000 {
		return strconv.ErrRange
	}
	*m = Money(n)
	return nil
}

type Price struct {
	Type   string `json:"passenger_type"`
	Amount Money  `json:"unit_price"`
	Seats  int    `json:"seats_per_person"`
	Min    int    `json:"min_age"`
	Max    *int   `json:"max_age_exclusive"`
}
type Surcharge struct {
	Code     string `json:"code"`
	Label    string `json:"label"`
	Amount   Money  `json:"amount"`
	Unit     string `json:"unit"`
	Required bool   `json:"required"`
}
type DepartureWrite struct {
	Start      time.Time   `json:"starts_at"`
	End        time.Time   `json:"ends_at"`
	Close      time.Time   `json:"sales_close_at"`
	Capacity   int         `json:"capacity"`
	Deposit    int         `json:"deposit_bps"`
	Due        time.Time   `json:"balance_due_at"`
	Prices     []Price     `json:"prices"`
	Surcharges []Surcharge `json:"surcharges"`
	Version    int         `json:"expected_price_version"`
}

func validDeparture(b DepartureWrite) bool {
	if b.Start.IsZero() || !b.End.After(b.Start) || b.Close.After(b.Start) || b.Capacity < 1 || b.Capacity > 10000 || b.Deposit < 1 || b.Deposit > 10000 || !b.Due.Before(b.Start) || len(b.Prices) != 3 || len(b.Surcharges) > 30 {
		return false
	}
	types := map[string]bool{}
	for _, p := range b.Prices {
		if types[p.Type] || p.Type != "adult" && p.Type != "child" && p.Type != "infant" || p.Min < 0 || p.Seats < 0 || p.Seats > 1 || p.Amount < 0 || p.Amount > 1000000000000 || p.Max != nil && *p.Max <= p.Min {
			return false
		}
		types[p.Type] = true
		if p.Type == "adult" && (p.Seats != 1 || p.Amount == 0) {
			return false
		}
	}
	for i, a := range b.Prices {
		for j, z := range b.Prices {
			if i >= j {
				continue
			}
			if (a.Max == nil || z.Min < *a.Max) && (z.Max == nil || a.Min < *z.Max) {
				return false
			}
		}
	}
	end := 0
	covered := false
	for count := 0; count < 3; count++ {
		found := false
		for _, p := range b.Prices {
			if p.Min == end {
				found = true
				if p.Max == nil {
					if count != 2 || p.Type != "adult" {
						return false
					}
					covered = true
					break
				}
				end = *p.Max
				break
			}
		}
		if !found {
			return false
		}
	}
	if !covered {
		return false
	}
	codes := map[string]bool{}
	for _, s := range b.Surcharges {
		if s.Code == "" || len(s.Code) > 40 || codes[s.Code] || s.Label == "" || s.Amount < 0 || s.Amount > 1000000000000 || s.Unit != "per_booking" && s.Unit != "per_person" {
			return false
		}
		codes[s.Code] = true
	}
	return true
}

const departureJSON = `to_jsonb(d)||jsonb_build_object('prices',COALESCE((SELECT jsonb_agg(to_jsonb(p)||jsonb_build_object('unit_price',p.unit_price::text) ORDER BY min_age) FROM departure_prices p WHERE p.departure_id=d.id),'[]'),'surcharges',COALESCE((SELECT jsonb_agg(to_jsonb(s)||jsonb_build_object('amount',s.amount::text) ORDER BY code) FROM departure_surcharges s WHERE s.departure_id=d.id),'[]'),'remaining_seats',d.capacity-d.held_seats-d.confirmed_seats,'availability',CASE WHEN d.status!='open' OR d.sales_close_at<=clock_timestamp() THEN 'closed' WHEN d.capacity=d.held_seats+d.confirmed_seats THEN 'sold_out' WHEN d.capacity-d.held_seats-d.confirmed_seats<=5 THEN 'low' ELSE 'available' END,'server_time',clock_timestamp())`

func (s *Server) departures(c *gin.Context) {
	var tour string
	if s.DB.QueryRow(c.Request.Context(), "SELECT id FROM tours WHERE slug=$1 AND status='published'", c.Param("slug")).Scan(&tour) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	s.departureList(c, tour, false)
}
func (s *Server) adminDepartures(c *gin.Context) {
	if validID(c, "id") {
		s.departureList(c, c.Param("id"), true)
	}
}
func (s *Server) departureList(c *gin.Context, tour string, admin bool) {
	p, n, valid := httpx.Page(c)
	if !valid {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	if s.expireTour(c.Request.Context(), tour) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	from, to := c.Query("from"), c.Query("to")
	for _, v := range []string{from, to} {
		if v != "" {
			if _, e := time.Parse("2006-01-02", v); e != nil {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
	}
	if from != "" && to != "" && from > to {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	where := ` WHERE d.tour_id=$1 AND ($2 OR d.status IN ('open','closed')) AND ($3='' OR (d.starts_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date>=NULLIF($3,'')::date) AND ($4='' OR (d.starts_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date<=NULLIF($4,'')::date)`
	args := []any{tour, admin, from, to, n, (p - 1) * n}
	out, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT "+departureJSON+" FROM departures d"+where+" ORDER BY starts_at,id LIMIT $5 OFFSET $6", args...)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	var total int
	s.DB.QueryRow(c.Request.Context(), "SELECT count(*) FROM departures d"+where, args[:4]...).Scan(&total)
	httpx.List(c, out, p, n, total)
}
func (s *Server) getDeparture(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var raw json.RawMessage
	if s.DB.QueryRow(c.Request.Context(), "SELECT "+departureJSON+" FROM departures d WHERE id=$1", c.Param("id")).Scan(&raw) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	httpx.OK(c, 200, raw)
}
func (s *Server) defaultDeparture(ctx context.Context, b *DepartureWrite) error {
	var raw []byte
	if e := s.DB.QueryRow(ctx, "SELECT value FROM site_settings WHERE key='booking_defaults'").Scan(&raw); e != nil {
		return e
	}
	var defaults struct {
		Deposit int `json:"deposit_bps"`
		Days    int `json:"balance_due_days"`
		Ages    []struct {
			Type  string `json:"passenger_type"`
			Min   int    `json:"min_age"`
			Max   *int   `json:"max_age_exclusive"`
			Seats int    `json:"seats_per_person"`
		} `json:"age_bands"`
	}
	if e := json.Unmarshal(raw, &defaults); e != nil {
		return e
	}
	if b.Deposit == 0 {
		b.Deposit = defaults.Deposit
	}
	if b.Due.IsZero() {
		b.Due = b.Start.AddDate(0, 0, -defaults.Days)
	}
	if b.Close.IsZero() {
		b.Close = b.Start.Add(-24 * time.Hour)
	}
	if b.Surcharges == nil {
		b.Surcharges = []Surcharge{}
	}
	if len(b.Prices) == 3 {
		for i, p := range b.Prices {
			if p.Type != "adult" && p.Type != "child" && p.Type != "infant" {
				continue
			}
			if p.Min == 0 && p.Max == nil && p.Seats == 0 {
				for _, band := range defaults.Ages {
					if band.Type == p.Type {
						b.Prices[i].Min = band.Min
						b.Prices[i].Max = band.Max
						b.Prices[i].Seats = band.Seats
					}
				}
			}
		}
	}
	return nil
}
func (s *Server) createDeparture(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var b DepartureWrite
	if !httpx.Body(c, &b) {
		return
	}
	if s.defaultDeparture(c.Request.Context(), &b) != nil || !validDeparture(b) {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	s.saveDepartures(c, []DepartureWrite{b})
}
func (s *Server) bulkDepartures(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var b struct {
		Departures []DepartureWrite `json:"departures"`
	}
	if !httpx.Body(c, &b) {
		return
	}
	if len(b.Departures) < 1 || len(b.Departures) > 100 {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	for i := range b.Departures {
		if s.defaultDeparture(c.Request.Context(), &b.Departures[i]) != nil || !validDeparture(b.Departures[i]) {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	s.saveDepartures(c, b.Departures)
}
func writePriceRows(ctx context.Context, tx pgx.Tx, id string, b DepartureWrite) error {
	for _, table := range []string{"departure_prices", "departure_surcharges"} {
		if _, e := tx.Exec(ctx, "DELETE FROM "+table+" WHERE departure_id=$1", id); e != nil {
			return e
		}
	}
	for _, p := range b.Prices {
		if _, e := tx.Exec(ctx, "INSERT INTO departure_prices(departure_id,passenger_type,unit_price,seats_per_person,min_age,max_age_exclusive) VALUES($1,$2,$3,$4,$5,$6)", id, p.Type, int64(p.Amount), p.Seats, p.Min, p.Max); e != nil {
			return e
		}
	}
	for _, p := range b.Surcharges {
		if _, e := tx.Exec(ctx, "INSERT INTO departure_surcharges(departure_id,code,label,amount,unit,required) VALUES($1,$2,$3,$4,$5,$6)", id, p.Code, p.Label, int64(p.Amount), p.Unit, p.Required); e != nil {
			return e
		}
	}
	return nil
}
func (s *Server) saveDepartures(c *gin.Context, items []DepartureWrite) {
	ctx := c.Request.Context()
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	defer tx.Rollback(ctx)
	ids := []string{}
	for _, b := range items {
		id := uuid.NewString()
		_, e = tx.Exec(ctx, "INSERT INTO departures(id,tour_id,starts_at,ends_at,sales_close_at,capacity,deposit_bps,balance_due_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", id, c.Param("id"), b.Start, b.End, b.Close, b.Capacity, b.Deposit, b.Due)
		if e != nil || writePriceRows(ctx, tx, id, b) != nil {
			httpx.Error(c, 409, "CONFLICT")
			return
		}
		if s.auditTx(ctx, tx, c, "departure.create", "departures", id) != nil {
			httpx.Error(c, 500, "INTERNAL_ERROR")
			return
		}
		ids = append(ids, id)
	}
	_, e = tx.Exec(ctx, "UPDATE tours SET poster_verified_at=NULL,poster_verified_by=NULL,poster_revision=poster_revision+1 WHERE id=$1", c.Param("id"))
	if e != nil || tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	if len(ids) == 1 {
		httpx.OK(c, 201, gin.H{"id": ids[0], "status": "draft"})
	} else {
		httpx.OK(c, 201, gin.H{"ids": ids})
	}
}
func (s *Server) patchDeparture(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	ctx := c.Request.Context()
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	defer tx.Rollback(ctx)
	var raw []byte
	var tour string
	var held, confirmed, version int
	var start, end time.Time
	e = tx.QueryRow(ctx, "SELECT tour_id,held_seats,confirmed_seats,price_version,starts_at,ends_at FROM departures WHERE id=$1 FOR UPDATE", c.Param("id")).Scan(&tour, &held, &confirmed, &version, &start, &end)
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if s.ExpireLocked != nil {
		if e = s.ExpireLocked(ctx, tx, c.Param("id")); e != nil {
			httpx.Error(c, 500, "INTERNAL_ERROR")
			return
		}
		if e = tx.QueryRow(ctx, "SELECT held_seats,confirmed_seats FROM departures WHERE id=$1", c.Param("id")).Scan(&held, &confirmed); e != nil {
			httpx.Error(c, 500, "INTERNAL_ERROR")
			return
		}
	}
	if tx.QueryRow(ctx, "SELECT "+departureJSON+" FROM departures d WHERE id=$1", c.Param("id")).Scan(&raw) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	var b DepartureWrite
	json.Unmarshal(raw, &b)
	base, _ := json.Marshal(b)
	var fields map[string]json.RawMessage
	json.Unmarshal(base, &fields)
	var patch map[string]json.RawMessage
	if !httpx.Body(c, &patch) {
		return
	}
	for k, v := range patch {
		if _, ok := fields[k]; !ok {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
		fields[k] = v
	}
	raw, _ = json.Marshal(fields)
	if json.Unmarshal(raw, &b) != nil || !validDeparture(b) {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	if b.Version != version {
		httpx.Error(c, 409, "VERSION_CONFLICT")
		return
	}
	if b.Capacity < held+confirmed || ((!b.Start.Equal(start) || !b.End.Equal(end)) && held+confirmed > 0) {
		httpx.Error(c, 409, "CONFLICT")
		return
	}
	_, e = tx.Exec(ctx, "UPDATE departures SET starts_at=$2,ends_at=$3,sales_close_at=$4,capacity=$5,deposit_bps=$6,balance_due_at=$7,price_version=price_version+1,updated_at=now() WHERE id=$1", c.Param("id"), b.Start, b.End, b.Close, b.Capacity, b.Deposit, b.Due)
	if e != nil || writePriceRows(ctx, tx, c.Param("id"), b) != nil {
		httpx.Error(c, 409, "CONFLICT")
		return
	}
	_, e = tx.Exec(ctx, "UPDATE tours SET poster_verified_at=NULL,poster_verified_by=NULL,poster_revision=poster_revision+1 WHERE id=$1", tour)
	if e != nil || s.auditTx(ctx, tx, c, "departure.update", "departures", c.Param("id")) != nil || tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	s.getDeparture(c)
}
func (s *Server) salesStatus(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var b struct {
		Status string `json:"status"`
	}
	if !httpx.Body(c, &b) || b.Status != "open" && b.Status != "closed" {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	ctx := c.Request.Context()
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	defer tx.Rollback(ctx)
	var tour, state string
	var complete bool
	e = tx.QueryRow(ctx, "SELECT tour_id,status,(sales_close_at>clock_timestamp() AND (SELECT count(*) FROM departure_prices p WHERE p.departure_id=d.id)=3) FROM departures d WHERE id=$1 FOR UPDATE", c.Param("id")).Scan(&tour, &state, &complete)
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if strings.Contains("cancelled,completed", state) || (b.Status == "open" && !complete) {
		httpx.Error(c, 409, "INVALID_STATE")
		return
	}
	_, e = tx.Exec(ctx, "UPDATE departures SET status=$2,updated_at=now() WHERE id=$1", c.Param("id"), b.Status)
	if e != nil || s.auditTx(ctx, tx, c, "departure.sales_status", "departures", c.Param("id")) != nil || tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	s.getDeparture(c)
}
