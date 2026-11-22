package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"regexp"
	"strings"
	"time"
)

type Day struct {
	Number        int    `json:"day_number"`
	Title         string `json:"title"`
	Body          string `json:"body"`
	Meals         string `json:"meals"`
	Accommodation string `json:"accommodation"`
}
type Stop struct {
	Day         int     `json:"day_number"`
	Order       int     `json:"sort_order"`
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Description string  `json:"description"`
}
type TourWrite struct {
	Slug           string          `json:"slug"`
	Code           string          `json:"code"`
	Title          string          `json:"title"`
	Summary        string          `json:"summary"`
	Description    string          `json:"description"`
	Region         string          `json:"region"`
	Audience       string          `json:"audience"`
	Days           int             `json:"duration_days"`
	Nights         int             `json:"duration_nights"`
	Origin         string          `json:"origin_name"`
	Transport      string          `json:"transport"`
	Difficulty     string          `json:"difficulty"`
	Hotel          string          `json:"hotel_summary"`
	Included       string          `json:"included"`
	Excluded       string          `json:"excluded"`
	Policy         string          `json:"cancellation_policy"`
	Preparation    string          `json:"preparation_guide"`
	AgePolicy      json.RawMessage `json:"age_policy"`
	Featured       bool            `json:"featured"`
	Cover          *string         `json:"cover_media_id"`
	Poster         *string         `json:"poster_media_id"`
	SEOTitle       string          `json:"seo_title"`
	SEODescription string          `json:"seo_description"`
	IsDemo         bool            `json:"is_demo"`
	Destinations   []string        `json:"destination_ids"`
	Categories     []string        `json:"category_ids"`
	Media          []string        `json:"media_ids"`
	Itinerary      []Day           `json:"days"`
	Stops          []Stop          `json:"stops"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validTour(b TourWrite) bool {
	if !slugPattern.MatchString(b.Slug) || len(b.Slug) > 180 || len(b.Code) < 1 || len(b.Code) > 32 || len(strings.TrimSpace(b.Title)) < 2 || len(b.Title) > 200 || b.Days < 1 || b.Days > 365 || b.Nights < 0 || b.Nights > 365 || b.Origin == "" || b.Transport == "" || len(b.Description) > 100000 || len(b.Summary) > 5000 {
		return false
	}
	if b.Region != "domestic" && b.Region != "international" || b.Audience != "general" && b.Audience != "school" && b.Audience != "company" || b.Difficulty != "easy" && b.Difficulty != "moderate" && b.Difficulty != "hard" {
		return false
	}
	for _, v := range []*string{b.Cover, b.Poster} {
		if v != nil && uuid.Validate(*v) != nil {
			return false
		}
	}
	for _, list := range [][]string{b.Destinations, b.Categories, b.Media} {
		if len(list) > 100 {
			return false
		}
		seen := map[string]bool{}
		for _, id := range list {
			if uuid.Validate(id) != nil || seen[id] {
				return false
			}
			seen[id] = true
		}
	}
	seen := map[int]bool{}
	for _, d := range b.Itinerary {
		if d.Number < 1 || d.Number > b.Days || d.Title == "" || len(d.Body) > 100000 || seen[d.Number] {
			return false
		}
		seen[d.Number] = true
	}
	for _, p := range b.Stops {
		if p.Day < 1 || p.Day > b.Days || p.Latitude < -90 || p.Latitude > 90 || p.Longitude < -180 || p.Longitude > 180 || p.Name == "" {
			return false
		}
	}
	return true
}
func (s *Server) auditTx(ctx context.Context, tx pgx.Tx, c *gin.Context, action, kind, id string) error {
	_, e := tx.Exec(ctx, "INSERT INTO audit_logs(actor_staff_id,action,entity_type,entity_id,request_id) VALUES($1,$2,$3,$4,$5)", s.UserID(c), action, kind, id, c.GetString("request_id"))
	return e
}
func (s *Server) adminTours(c *gin.Context) {
	p, n, valid := httpx.Page(c)
	if !valid {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	status := c.Query("status")
	q := c.Query("q")
	rows, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT to_jsonb(t) FROM tours t WHERE ($1='' OR status=$1) AND title ILIKE $2 ORDER BY updated_at DESC,id LIMIT $3 OFFSET $4", status, "%"+q+"%", n, (p-1)*n)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	var total int
	s.DB.QueryRow(c.Request.Context(), "SELECT count(*) FROM tours WHERE ($1='' OR status=$1) AND title ILIKE $2", status, "%"+q+"%").Scan(&total)
	httpx.List(c, rows, p, n, total)
}
func (s *Server) loadTour(ctx context.Context, id string) (TourWrite, error) {
	var b TourWrite
	var raw []byte
	e := s.DB.QueryRow(ctx, `SELECT to_jsonb(t)||jsonb_build_object('destination_ids',COALESCE((SELECT jsonb_agg(destination_id) FROM tour_destinations WHERE tour_id=t.id),'[]'),'category_ids',COALESCE((SELECT jsonb_agg(category_id) FROM tour_category_links WHERE tour_id=t.id),'[]'),'media_ids',COALESCE((SELECT jsonb_agg(media_id ORDER BY sort_order) FROM tour_media WHERE tour_id=t.id),'[]'),'days',COALESCE((SELECT jsonb_agg(to_jsonb(d) ORDER BY day_number) FROM tour_days d WHERE tour_id=t.id),'[]'),'stops',COALESCE((SELECT jsonb_agg(to_jsonb(d) ORDER BY day_number,sort_order) FROM tour_stops d WHERE tour_id=t.id),'[]')) FROM tours t WHERE t.id=$1`, id).Scan(&raw)
	if e == nil {
		e = json.Unmarshal(raw, &b)
	}
	return b, e
}
func (s *Server) adminTour(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	b, e := s.loadTour(c.Request.Context(), c.Param("id"))
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	var status string
	var revision int
	s.DB.QueryRow(c.Request.Context(), "SELECT status,poster_revision FROM tours WHERE id=$1", c.Param("id")).Scan(&status, &revision)
	raw, _ := json.Marshal(b)
	var out map[string]any
	json.Unmarshal(raw, &out)
	out["id"] = c.Param("id")
	out["status"] = status
	out["poster_revision"] = revision
	httpx.OK(c, 200, out)
}
func (s *Server) createTour(c *gin.Context) {
	b := TourWrite{Region: "domestic", Audience: "general", Difficulty: "easy", Days: 2, Nights: 1, IsDemo: true, AgePolicy: json.RawMessage(`{}`)}
	if !httpx.Body(c, &b) {
		return
	}
	if b.Policy == "" {
		var defaults struct {
			Policy string `json:"cancellation_policy"`
		}
		var raw []byte
		s.DB.QueryRow(c.Request.Context(), "SELECT value FROM site_settings WHERE key='booking_defaults'").Scan(&raw)
		json.Unmarshal(raw, &defaults)
		b.Policy = defaults.Policy
	}
	s.saveTour(c, "", b)
}
func (s *Server) patchTour(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	existing, e := s.loadTour(c.Request.Context(), c.Param("id"))
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	var patch map[string]json.RawMessage
	if !httpx.Body(c, &patch) {
		return
	}
	base, _ := json.Marshal(existing)
	var merged map[string]json.RawMessage
	json.Unmarshal(base, &merged)
	for k, v := range patch {
		if _, exists := merged[k]; !exists {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
		merged[k] = v
	}
	raw, _ := json.Marshal(merged)
	var b TourWrite
	if json.Unmarshal(raw, &b) != nil {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	s.saveTour(c, c.Param("id"), b)
}
func (s *Server) saveTour(c *gin.Context, id string, b TourWrite) {
	b.Description = richText.Sanitize(b.Description)
	b.Included = richText.Sanitize(b.Included)
	b.Excluded = richText.Sanitize(b.Excluded)
	b.Policy = richText.Sanitize(b.Policy)
	for i := range b.Itinerary {
		b.Itinerary[i].Body = richText.Sanitize(b.Itinerary[i].Body)
	}
	if !validTour(b) {
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
	created := id == ""
	if created {
		id = uuid.NewString()
	} else {
		var state string
		if tx.QueryRow(ctx, "SELECT status FROM tours WHERE id=$1 FOR UPDATE", id).Scan(&state) != nil {
			httpx.Error(c, 404, "NOT_FOUND")
			return
		}
	}
	for _, mid := range append(append([]string{}, b.Media...), optionalIDs(b.Cover, b.Poster)...) {
		var valid bool
		if tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1 AND visibility='public' AND status='ready' AND kind IN ('image','video'))", mid).Scan(&valid) != nil || !valid {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	for _, mid := range optionalIDs(b.Cover, b.Poster) {
		var good bool
		if tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1 AND kind='image' AND status='ready' AND visibility='public')", mid).Scan(&good) != nil || !good {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	fields := []string{"slug", "code", "title", "summary", "description", "region", "audience", "duration_days", "duration_nights", "origin_name", "transport", "difficulty", "hotel_summary", "included", "excluded", "cancellation_policy", "preparation_guide", "age_policy", "featured", "cover_media_id", "poster_media_id", "seo_title", "seo_description", "is_demo"}
	vals := []any{b.Slug, b.Code, b.Title, b.Summary, b.Description, b.Region, b.Audience, b.Days, b.Nights, b.Origin, b.Transport, b.Difficulty, b.Hotel, b.Included, b.Excluded, b.Policy, b.Preparation, b.AgePolicy, b.Featured, b.Cover, b.Poster, b.SEOTitle, b.SEODescription, b.IsDemo}
	args := append([]any{id}, vals...)
	params := []string{}
	sets := []string{}
	for i, f := range fields {
		ph := fmt.Sprintf("$%d", i+2)
		params = append(params, ph)
		sets = append(sets, f+"="+ph)
	}
	if created {
		_, e = tx.Exec(ctx, "INSERT INTO tours(id,"+strings.Join(fields, ",")+") VALUES($1,"+strings.Join(params, ",")+")", args...)
	} else {
		_, e = tx.Exec(ctx, "UPDATE tours SET "+strings.Join(sets, ",")+",poster_verified_at=NULL,poster_verified_by=NULL,poster_revision=poster_revision+1,updated_at=now() WHERE id=$1", args...)
	}
	if e != nil {
		httpx.Error(c, 409, "CONFLICT")
		return
	}
	for _, table := range []string{"tour_destinations", "tour_category_links", "tour_days", "tour_stops", "tour_media"} {
		if _, e = tx.Exec(ctx, "DELETE FROM "+table+" WHERE tour_id=$1", id); e != nil {
			httpx.Error(c, 500, "INTERNAL_ERROR")
			return
		}
	}
	for _, v := range b.Destinations {
		_, e = tx.Exec(ctx, "INSERT INTO tour_destinations(tour_id,destination_id) VALUES($1,$2)", id, v)
		if e != nil {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	for _, v := range b.Categories {
		_, e = tx.Exec(ctx, "INSERT INTO tour_category_links(tour_id,category_id) VALUES($1,$2)", id, v)
		if e != nil {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	for _, d := range b.Itinerary {
		_, e = tx.Exec(ctx, "INSERT INTO tour_days(tour_id,day_number,title,body,meals,accommodation) VALUES($1,$2,$3,$4,$5,$6)", id, d.Number, d.Title, d.Body, d.Meals, d.Accommodation)
		if e != nil {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	for _, p := range b.Stops {
		_, e = tx.Exec(ctx, "INSERT INTO tour_stops(tour_id,day_number,sort_order,name,latitude,longitude,description) VALUES($1,$2,$3,$4,$5,$6,$7)", id, p.Day, p.Order, p.Name, p.Latitude, p.Longitude, p.Description)
		if e != nil {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	for order, v := range b.Media {
		_, e = tx.Exec(ctx, "INSERT INTO tour_media(tour_id,media_id,sort_order) VALUES($1,$2,$3)", id, v, order)
		if e != nil {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	if e = s.auditTx(ctx, tx, c, "tour.save", "tours", id); e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	if tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	c.Params = append(c.Params, gin.Param{Key: "id", Value: id})
	if created {
		httpx.OK(c, 201, gin.H{"id": id, "slug": b.Slug, "status": "draft"})
		return
	}
	s.adminTour(c)
}
func optionalIDs(values ...*string) []string {
	out := []string{}
	for _, v := range values {
		if v != nil {
			out = append(out, *v)
		}
	}
	return out
}
func (s *Server) archiveTour(c *gin.Context) { s.tourState(c, "archived") }
func (s *Server) publishTour(c *gin.Context) { s.tourState(c, "published") }
func (s *Server) tourState(c *gin.Context, state string) {
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
	var exists bool
	e = tx.QueryRow(ctx, "SELECT true FROM tours WHERE id=$1 FOR UPDATE", c.Param("id")).Scan(&exists)
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if state == "published" {
		var valid bool
		e = tx.QueryRow(ctx, `SELECT length(summary)>0 AND length(included)>0 AND length(cancellation_policy)>0 AND (SELECT count(*) FROM tour_days WHERE tour_id=t.id)=duration_days AND EXISTS(SELECT 1 FROM departures d WHERE d.tour_id=t.id AND (SELECT count(*) FROM departure_prices p WHERE p.departure_id=d.id)=3) AND EXISTS(SELECT 1 FROM media_assets m WHERE m.id=t.cover_media_id AND m.visibility='public' AND m.status='ready') FROM tours t WHERE id=$1`, c.Param("id")).Scan(&valid)
		if e != nil || !valid {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	_, e = tx.Exec(ctx, "UPDATE tours SET status=$2,published_at=CASE WHEN $2='published' THEN clock_timestamp() ELSE published_at END,updated_at=now() WHERE id=$1", c.Param("id"), state)
	if e != nil || s.auditTx(ctx, tx, c, "tour."+state, "tours", c.Param("id")) != nil || tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.OK(c, 200, gin.H{"id": c.Param("id"), "status": state})
}
func (s *Server) cloneTour(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var input struct {
		Slug  string `json:"slug"`
		Code  string `json:"code"`
		Title string `json:"title"`
	}
	if !httpx.Body(c, &input) {
		return
	}
	b, e := s.loadTour(c.Request.Context(), c.Param("id"))
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	b.Slug = input.Slug
	b.Code = input.Code
	b.Title = input.Title
	c.Params = nil
	s.saveTour(c, "", b)
}
func (s *Server) verifyPoster(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var b struct {
		Version   int  `json:"expected_poster_revision"`
		Confirmed bool `json:"confirmed"`
	}
	if !httpx.Body(c, &b) || !b.Confirmed {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	tag, e := s.DB.Exec(c.Request.Context(), "UPDATE tours SET poster_verified_at=now(),poster_verified_by=$3 WHERE id=$1 AND poster_revision=$2 AND EXISTS(SELECT 1 FROM media_assets m WHERE m.id=tours.poster_media_id AND m.kind='image' AND m.visibility='public' AND m.status='ready')", c.Param("id"), b.Version, s.UserID(c))
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	if tag.RowsAffected() != 1 {
		httpx.Error(c, 409, "VERSION_CONFLICT")
		return
	}
	httpx.OK(c, 200, gin.H{"verified": true, "verified_at": time.Now().UTC()})
}
