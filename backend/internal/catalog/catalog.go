package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	DB     *pgxpool.Pool
	UserID func(*gin.Context) string
	Expire func(context.Context, string) error
	Media  *MediaStore
}

func (s *Server) Register(v, admin *gin.RouterGroup, product, content, media, private gin.HandlerFunc) {
	v.GET("/tours", s.listTours)
	v.GET("/tours/suggestions", s.suggestions)
	v.GET("/tours/:slug", s.tour)
	v.GET("/tours/:slug/departures", s.departures)
	v.GET("/tours/:slug/reviews", s.reviews)
	v.GET("/home", s.home)
	v.GET("/gallery", s.gallery)
	admin.GET("/tours", product, s.adminTours)
	admin.POST("/tours", product, s.createTour)
	admin.GET("/tours/:id", product, s.adminTour)
	admin.PATCH("/tours/:id", product, s.patchTour)
	admin.DELETE("/tours/:id", product, s.archiveTour)
	admin.POST("/tours/:id/archive", product, s.archiveTour)
	admin.POST("/tours/:id/publish", product, s.publishTour)
	admin.POST("/tours/:id/clone", product, s.cloneTour)
	admin.GET("/tours/:id/departures", product, s.adminDepartures)
	admin.POST("/tours/:id/departures", product, s.createDeparture)
	admin.POST("/tours/:id/departures/bulk", product, s.bulkDepartures)
	admin.GET("/departures/:id", product, s.getDeparture)
	admin.PATCH("/departures/:id", product, s.patchDeparture)
	admin.POST("/departures/:id/sales-status", product, s.salesStatus)
	admin.POST("/tours/:id/poster-verification", product, s.verifyPoster)
	s.registerContent(v, admin, content, product)
	s.registerMedia(v, admin, media, private)
}

const mediaObject = `jsonb_build_object('id',m.id,'kind',m.kind,'url','/api/v1/media/'||m.id::text,'alt_text',m.alt_text,'caption',m.caption,'width',m.width,'height',m.height)`
const tourJSON = `(to_jsonb(t)-'cover_media_id'-'poster_media_id'-'poster_verified_by') || jsonb_build_object(
 'currency','VND','cover',(SELECT ` + mediaObject + ` FROM media_assets m WHERE m.id=t.cover_media_id AND m.visibility='public' AND m.status='ready'),
 'poster',(SELECT ` + mediaObject + ` FROM media_assets m WHERE m.id=t.poster_media_id AND t.poster_verified_at IS NOT NULL AND m.visibility='public' AND m.status='ready'),
 'price_from',best.price::text,'next_departure_at',best.starts_at,
 'rating',(SELECT round(avg(r.rating),1) FROM reviews r WHERE r.tour_id=t.id AND r.status='approved'), 'review_count',(SELECT count(*) FROM reviews r WHERE r.tour_id=t.id AND r.status='approved'))`
const tourDetail = ` || jsonb_build_object('days',COALESCE((SELECT jsonb_agg(to_jsonb(d) ORDER BY day_number) FROM tour_days d WHERE d.tour_id=t.id),'[]'::jsonb),'stops',COALESCE((SELECT jsonb_agg(to_jsonb(d) ORDER BY day_number,sort_order) FROM tour_stops d WHERE d.tour_id=t.id),'[]'::jsonb),'gallery',COALESCE((SELECT jsonb_agg(` + mediaObject + ` ORDER BY tm.sort_order) FROM tour_media tm JOIN media_assets m ON m.id=tm.media_id WHERE tm.tour_id=t.id AND m.visibility='public' AND m.status='ready'),'[]'::jsonb),'destinations',COALESCE((SELECT jsonb_agg(to_jsonb(d)) FROM tour_destinations td JOIN destinations d ON d.id=td.destination_id WHERE td.tour_id=t.id AND d.published),'[]'::jsonb),'categories',COALESCE((SELECT jsonb_agg(to_jsonb(d)) FROM tour_category_links tc JOIN tour_categories d ON d.id=tc.category_id WHERE tc.tour_id=t.id),'[]'::jsonb),'faqs',COALESCE((SELECT jsonb_agg(to_jsonb(f) ORDER BY sort_order) FROM faqs f WHERE f.tour_id=t.id AND f.published),'[]'::jsonb))`
const bestDeparture = ` LEFT JOIN LATERAL(SELECT p.unit_price AS price,d.starts_at FROM departures d JOIN departure_prices p ON p.departure_id=d.id AND p.passenger_type='adult' WHERE d.tour_id=t.id AND d.status='open' AND d.sales_close_at>clock_timestamp() AND d.capacity>d.held_seats+d.confirmed_seats ORDER BY p.unit_price,d.starts_at LIMIT 1)best ON true `

func (s *Server) expireTour(ctx context.Context, tour string) error {
	if s.Expire == nil {
		return nil
	}
	rows, e := s.DB.Query(ctx, "SELECT id FROM departures WHERE tour_id=$1 ORDER BY id", tour)
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
	rows.Close()
	for _, id := range ids {
		if e = s.Expire(ctx, id); e != nil {
			return e
		}
	}
	return nil
}
func (s *Server) listTours(c *gin.Context) {
	allowed := map[string]bool{}
	for _, k := range strings.Split("q,destination,category,region,audience,depart_from,depart_to,budget_min,budget_max,duration_min,duration_max,travellers,transport,difficulty,min_rating,sort,page,page_size", ",") {
		allowed[k] = true
	}
	for k := range c.Request.URL.Query() {
		if !allowed[k] {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}
	for k, choices := range map[string]string{"region": ",domestic,international,", "audience": ",general,school,company,", "difficulty": ",easy,moderate,hard,"} {
		if v := c.Query(k); v != "" && !strings.Contains(choices, ","+v+",") {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
	}

	p, n, valid := httpx.Page(c)
	if !valid {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	args := []any{}
	where := []string{"t.status='published'"}
	dep := []string{"d.tour_id=t.id", "d.status='open'", "d.sales_close_at>clock_timestamp()", "d.capacity>d.held_seats+d.confirmed_seats"}
	add := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	filtered := false
	q := c.Query("q")
	if len(q) > 100 {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	if q != "" {
		param := add("%" + q + "%")
		where = append(where, "(t.title ILIKE "+param+" OR EXISTS(SELECT 1 FROM tour_destinations td JOIN destinations des ON des.id=td.destination_id WHERE td.tour_id=t.id AND des.name ILIKE "+param+"))")
	}
	for _, k := range []string{"region", "audience", "transport", "difficulty"} {
		if v := c.Query(k); v != "" {
			where = append(where, "t."+k+"="+add(v))
		}
	}
	for _, x := range []struct{ key, table, join, column string }{{"destination", "tour_destinations", "destinations", "destination_id"}, {"category", "tour_category_links", "tour_categories", "category_id"}} {
		if v := c.Query(x.key); v != "" {
			where = append(where, "EXISTS(SELECT 1 FROM "+x.table+" l JOIN "+x.join+" j ON j.id=l."+x.column+" WHERE l.tour_id=t.id AND j.slug="+add(v)+")")
		}
	}
	for _, k := range []string{"depart_from", "depart_to"} {
		if v := c.Query(k); v != "" {
			if _, e := time.Parse("2006-01-02", v); e != nil {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
			op := ">="
			if k == "depart_to" {
				op = "<="
			}
			dep = append(dep, "(d.starts_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date"+op+add(v)+"::date")
			filtered = true
		}
	}
	if c.Query("depart_from") != "" && c.Query("depart_to") != "" && c.Query("depart_from") > c.Query("depart_to") {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	for _, k := range []string{"budget_min", "budget_max", "duration_min", "duration_max", "travellers"} {
		if v := c.Query(k); v != "" {
			num, e := strconv.ParseInt(v, 10, 64)
			if e != nil || num < 0 || num > 1000000000000 || (strings.HasPrefix(k, "duration") && (num < 1 || num > 365)) || (k == "travellers" && (num < 1 || num > 20)) {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
			switch k {
			case "budget_min":
				dep = append(dep, "p.unit_price>="+add(num))
				filtered = true
			case "budget_max":
				dep = append(dep, "p.unit_price<="+add(num))
				filtered = true
			case "duration_min":
				where = append(where, "t.duration_days>="+add(num))
			case "duration_max":
				where = append(where, "t.duration_days<="+add(num))
			case "travellers":
				dep = append(dep, "d.capacity-d.held_seats-d.confirmed_seats>="+add(num))
				filtered = true
			}
		}
	}
	for _, prefix := range []string{"budget", "duration"} {
		lo, hi := c.Query(prefix+"_min"), c.Query(prefix+"_max")
		if lo != "" && hi != "" {
			l, _ := strconv.ParseInt(lo, 10, 64)
			h, _ := strconv.ParseInt(hi, 10, 64)
			if l > h {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
	}
	if v := c.Query("min_rating"); v != "" {
		r, e := strconv.ParseFloat(v, 64)
		if e != nil || math.IsNaN(r) || r < 1 || r > 5 {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
		where = append(where, "COALESCE((SELECT avg(r.rating) FROM reviews r WHERE r.tour_id=t.id AND r.status='approved'),0)>="+add(r))
	}
	sorts := map[string]string{"": "t.featured DESC,t.published_at DESC,t.id", "recommended": "t.featured DESC,t.published_at DESC,t.id", "price_asc": "best.price ASC NULLS LAST,t.id", "price_desc": "best.price DESC NULLS LAST,t.id", "duration_asc": "t.duration_days,t.id", "departure_asc": "best.starts_at ASC NULLS LAST,t.id", "rating_desc": "(SELECT avg(r.rating) FROM reviews r WHERE r.tour_id=t.id AND r.status='approved') DESC NULLS LAST,t.id"}
	sort, found := sorts[c.Query("sort")]
	if !found {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	if s.Expire != nil {
		ids, e := s.DB.Query(c.Request.Context(), "SELECT id FROM departures WHERE held_seats>0 ORDER BY id LIMIT 500")
		if e == nil {
			list := []string{}
			for ids.Next() {
				var id string
				ids.Scan(&id)
				list = append(list, id)
			}
			ids.Close()
			for _, id := range list {
				if s.Expire(c.Request.Context(), id) != nil {
					httpx.Error(c, 503, "INTERNAL_ERROR")
					return
				}
			}
		}
	}
	join := " LEFT JOIN LATERAL(SELECT p.unit_price price,d.starts_at FROM departures d JOIN departure_prices p ON p.departure_id=d.id AND p.passenger_type='adult' WHERE " + strings.Join(dep, " AND ") + " ORDER BY p.unit_price,d.starts_at LIMIT 1)best ON true "
	if filtered {
		where = append(where, "best.price IS NOT NULL")
	}
	from := " FROM tours t " + join + " WHERE " + strings.Join(where, " AND ")
	var total int
	if e := s.DB.QueryRow(c.Request.Context(), "SELECT count(*)"+from, args...).Scan(&total); e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	countArgs := append(append([]any{}, args...), n, (p-1)*n)
	rows, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT "+tourJSON+from+" ORDER BY "+sort+fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2), countArgs...)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.List(c, rows, p, n, total)
}
func (s *Server) tour(c *gin.Context) {
	var id string
	if s.DB.QueryRow(c.Request.Context(), "SELECT id FROM tours WHERE slug=$1 AND status='published'", c.Param("slug")).Scan(&id) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if s.expireTour(c.Request.Context(), id) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	var raw json.RawMessage
	e := s.DB.QueryRow(c.Request.Context(), "SELECT "+tourJSON+tourDetail+" FROM tours t "+bestDeparture+" WHERE t.id=$1", id).Scan(&raw)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.OK(c, 200, raw)
}
func (s *Server) suggestions(c *gin.Context) {
	q := c.Query("q")
	if len([]rune(q)) < 2 || len(q) > 100 {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	limit := 5
	if v := c.Query("limit"); v != "" {
		i, e := strconv.Atoi(v)
		if e != nil || i < 1 || i > 10 {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
		limit = i
	}
	t, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT jsonb_build_object('id',id,'slug',slug,'title',title) FROM tours WHERE status='published' AND title ILIKE $1 ORDER BY title LIMIT $2", "%"+q+"%", limit)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	d, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT jsonb_build_object('slug',slug,'name',name) FROM destinations WHERE published AND name ILIKE $1 ORDER BY name LIMIT $2", "%"+q+"%", limit)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.OK(c, 200, gin.H{"tours": t, "destinations": d})
}
func validID(c *gin.Context, key string) bool {
	if uuid.Validate(c.Param(key)) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return false
	}
	return true
}
func safeLink(value string) bool {
	u, e := url.Parse(value)
	return e == nil && (strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") || (u.Scheme == "https" && u.Host != ""))
}
