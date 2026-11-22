package catalog

import (
	"encoding/json"
	"fmt"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"strings"
	"time"
)

var richText = bluemonday.UGCPolicy()

type resource struct {
	path, table, fields, required, filter string
	product                               bool
}

var resources = []resource{
	{"destinations", "destinations", "slug,name,description,country_code,latitude,longitude,cover_media_id,seo_title,seo_description,published", "slug,name", "published", true},
	{"tour-categories", "tour_categories", "slug,name,description,sort_order", "slug,name", "", true},
	{"posts", "posts", "slug,title,excerpt,body,category_id,cover_media_id,status,seo_title,seo_description", "slug,title", "status='published'", false},
	{"pages", "pages", "slug,title,body,status,seo_title,seo_description", "slug,title", "status='published'", false},
	{"post-categories", "post_categories", "slug,name", "slug,name", "", false}, {"tags", "tags", "slug,name", "slug,name", "", false},
	{"banners", "banners", "title,media_id,target_url,placement,sort_order,starts_at,ends_at,active", "title,media_id,placement", "active AND (starts_at IS NULL OR starts_at<=now()) AND (ends_at IS NULL OR ends_at>now())", false},
	{"faqs", "faqs", "topic,tour_id,question,answer,sort_order,published", "topic,question,answer", "published AND (tour_id IS NULL OR EXISTS(SELECT 1 FROM tours WHERE id=tour_id AND status='published'))", false},
	{"reviews", "reviews", "tour_id,group_type,display_name,organization_name,rating,body,media_id,source,is_demo", "display_name,rating,body,source", "status='approved' AND (tour_id IS NULL OR EXISTS(SELECT 1 FROM tours WHERE id=tour_id AND status='published'))", false},
}

func (s *Server) registerContent(v, admin *gin.RouterGroup, content, product gin.HandlerFunc) {
	for _, cfg := range resources {
		r := cfg
		guard := content
		if r.product {
			guard = product
		}
		if r.path != "banners" && r.path != "reviews" {
			v.GET("/"+r.path, func(c *gin.Context) { s.contentList(c, r, false) })
			if r.path == "posts" || r.path == "pages" || r.path == "destinations" {
				v.GET("/"+r.path+"/:slug", func(c *gin.Context) { s.contentDetail(c, r, false) })
			}
		}
		admin.GET("/"+r.path, guard, func(c *gin.Context) { s.contentList(c, r, true) })
		admin.POST("/"+r.path, guard, func(c *gin.Context) { s.contentSave(c, r, false) })
		admin.GET("/"+r.path+"/:id", guard, func(c *gin.Context) { s.contentDetail(c, r, true) })
		admin.PATCH("/"+r.path+"/:id", guard, func(c *gin.Context) { s.contentSave(c, r, true) })
		admin.DELETE("/"+r.path+"/:id", guard, func(c *gin.Context) { s.contentDelete(c, r) })
	}
	v.GET("/testimonials", func(c *gin.Context) { s.contentList(c, resources[8], false) })
	admin.POST("/reviews/:id/moderation", content, s.moderateReview)
}
func (s *Server) contentWhere(c *gin.Context, r resource, admin bool) (string, []any) {
	clauses := []string{"true"}
	args := []any{}
	if !admin && r.filter != "" {
		clauses = append(clauses, r.filter)
	}
	if q := c.Query("q"); q != "" {
		args = append(args, "%"+q+"%")
		clauses = append(clauses, fmt.Sprintf("to_jsonb(x)::text ILIKE $%d", len(args)))
	}
	for _, field := range []string{"topic", "tour_id", "group_type", "status", "collection", "kind"} {
		if val := c.Query(field); val != "" && strings.Contains(","+r.fields+",", ","+field+",") {
			if strings.HasSuffix(field, "_id") && uuid.Validate(val) != nil {
				continue
			}
			args = append(args, val)
			clauses = append(clauses, fmt.Sprintf("x.%s=$%d", field, len(args)))
		}
	}
	if r.table == "posts" {
		if val := c.Query("category"); val != "" {
			args = append(args, val)
			clauses = append(clauses, fmt.Sprintf("EXISTS(SELECT 1 FROM post_categories cat WHERE cat.id=x.category_id AND cat.slug=$%d)", len(args)))
		}
		if val := c.Query("tag"); val != "" {
			args = append(args, val)
			clauses = append(clauses, fmt.Sprintf("EXISTS(SELECT 1 FROM post_tags pt JOIN tags t ON t.id=pt.tag_id WHERE pt.post_id=x.id AND t.slug=$%d)", len(args)))
		}
	}
	return strings.Join(clauses, " AND "), args
}
func (s *Server) contentList(c *gin.Context, r resource, admin bool) {
	p, n, valid := httpx.Page(c)
	if !valid || len(c.Query("q")) > 100 {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	where, args := s.contentWhere(c, r, admin)
	var total int
	if e := s.DB.QueryRow(c.Request.Context(), "SELECT count(*) FROM "+r.table+" x WHERE "+where, args...).Scan(&total); e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	args = append(args, n, (p-1)*n)
	value := "to_jsonb(x)"
	if !admin {
		value += "-'author_id'-'reviewed_by'-'reviewed_at'"
	}
	if r.table == "posts" {
		value += `||jsonb_build_object('tags',COALESCE((SELECT jsonb_agg(to_jsonb(t)) FROM post_tags pt JOIN tags t ON t.id=pt.tag_id WHERE pt.post_id=x.id),'[]'))`
	}
	out, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT "+value+" FROM "+r.table+" x WHERE "+where+fmt.Sprintf(" ORDER BY created_at DESC,id LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.List(c, out, p, n, total)
}
func (s *Server) contentDetail(c *gin.Context, r resource, admin bool) {
	where := "slug=$1"
	key := c.Param("slug")
	if admin {
		if !validID(c, "id") {
			return
		}
		where = "id=$1"
		key = c.Param("id")
	} else if r.filter != "" {
		where += " AND (" + r.filter + ")"
	}
	var raw json.RawMessage
	value := "to_jsonb(x)"
	if !admin {
		value += "-'author_id'-'reviewed_by'-'reviewed_at'"
	}
	if r.table == "posts" {
		value += `||jsonb_build_object('tag_ids',COALESCE((SELECT jsonb_agg(tag_id) FROM post_tags WHERE post_id=x.id),'[]'),'tags',COALESCE((SELECT jsonb_agg(to_jsonb(t)) FROM post_tags pt JOIN tags t ON t.id=pt.tag_id WHERE pt.post_id=x.id),'[]'))`
	}
	if r.table == "posts" && !admin {
		value += `||jsonb_build_object('category',(SELECT to_jsonb(cat) FROM post_categories cat WHERE cat.id=x.category_id),'related_posts',COALESCE((SELECT jsonb_agg(to_jsonb(rel)) FROM (SELECT p.id,p.slug,p.title,p.excerpt,p.cover_media_id FROM posts p WHERE p.status='published' AND p.id<>x.id AND (p.category_id=x.category_id OR EXISTS(SELECT 1 FROM post_tags a JOIN post_tags b ON a.tag_id=b.tag_id WHERE a.post_id=x.id AND b.post_id=p.id)) ORDER BY p.published_at DESC LIMIT 3) rel),'[]'))`
	}
	if s.DB.QueryRow(c.Request.Context(), "SELECT "+value+" FROM "+r.table+" x WHERE "+where, key).Scan(&raw) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	httpx.OK(c, 200, raw)
}
func (s *Server) contentSave(c *gin.Context, r resource, patch bool) {
	if patch && !validID(c, "id") {
		return
	}
	var body map[string]json.RawMessage
	if !httpx.Body(c, &body) {
		return
	}
	allowed := strings.Split(r.fields, ",")
	valid := map[string]bool{}
	for _, f := range allowed {
		valid[f] = true
	}
	var tags []string
	if r.table == "posts" {
		if raw, ok := body["tag_ids"]; ok {
			if json.Unmarshal(raw, &tags) != nil || len(tags) > 100 {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
			for _, id := range tags {
				if uuid.Validate(id) != nil {
					httpx.Error(c, 422, "VALIDATION_ERROR")
					return
				}
			}
			delete(body, "tag_ids")
		}
	}
	if !patch {
		for _, f := range strings.Split(r.required, ",") {
			if _, ok := body[f]; !ok {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
	}
	fields := []string{}
	values := []any{}
	for field, raw := range body {
		if !valid[field] {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
		var v any
		if json.Unmarshal(raw, &v) != nil {
			httpx.Error(c, 422, "VALIDATION_ERROR")
			return
		}
		if text, ok := v.(string); ok {
			if strings.Contains(","+r.required+",", ","+field+",") && strings.TrimSpace(text) == "" || len(text) > 100000 || field == "slug" && (!slugPattern.MatchString(text) || len(text) > 180) || strings.HasSuffix(field, "_id") && uuid.Validate(text) != nil {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
			if field == "target_url" && !safeLink(text) {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
			if strings.HasSuffix(field, "_at") {
				if _, e := time.Parse(time.RFC3339, text); e != nil {
					httpx.Error(c, 422, "VALIDATION_ERROR")
					return
				}
			}
			if field == "body" || field == "description" || field == "answer" {
				v = richText.Sanitize(text)
			}
			if field == "status" && text != "draft" && text != "published" && text != "archived" {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
		if v != nil && (field == "cover_media_id" || field == "media_id") {
			var ready bool
			if s.DB.QueryRow(c.Request.Context(), "SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1 AND kind='image' AND visibility='public' AND status='ready')", v).Scan(&ready) != nil || !ready {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
		if num, ok := v.(float64); ok {
			if field == "latitude" || field == "longitude" {
			} else if num == float64(int64(num)) {
				v = int64(num)
			} else {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
		if v != nil {
			if strings.Contains(",sort_order,rating,latitude,longitude,", ","+field+",") {
				if _, a := v.(int64); !a {
					if _, b := v.(float64); !b {
						httpx.Error(c, 422, "VALIDATION_ERROR")
						return
					}
				}
			} else if strings.Contains(",published,active,is_demo,", ","+field+",") {
				if _, ok := v.(bool); !ok {
					httpx.Error(c, 422, "VALIDATION_ERROR")
					return
				}
			} else if _, ok := v.(string); !ok {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
		fields = append(fields, field)
		values = append(values, v)
	}
	if len(fields) == 0 && tags == nil {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	id := c.Param("id")
	if !patch {
		id = uuid.NewString()
		fields = append(fields, "id")
		values = append(values, id)
		if r.table == "posts" {
			fields = append(fields, "author_id")
			values = append(values, s.UserID(c))
		}
	}
	if r.table == "posts" || r.table == "pages" {
		if raw, ok := body["status"]; ok && string(raw) == `"published"` {
			fields = append(fields, "published_at")
			values = append(values, time.Now().UTC())
		}
	}
	ctx := c.Request.Context()
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	defer tx.Rollback(ctx)
	params, sets := []string{}, []string{}
	for i, f := range fields {
		ph := fmt.Sprintf("$%d", i+1)
		params = append(params, ph)
		sets = append(sets, f+"="+ph)
	}
	var affected int64
	if patch {
		values = append(values, id)
		tag, err := tx.Exec(ctx, "UPDATE "+r.table+" SET "+strings.Join(append(sets, "updated_at=now()"), ",")+" WHERE id="+fmt.Sprintf("$%d", len(values)), values...)
		e = err
		affected = tag.RowsAffected()
	} else {
		tag, err := tx.Exec(ctx, "INSERT INTO "+r.table+"("+strings.Join(fields, ",")+") VALUES("+strings.Join(params, ",")+")", values...)
		e = err
		affected = tag.RowsAffected()
	}
	if e != nil {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	if affected == 0 {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if r.table == "posts" && tags != nil {
		if _, e = tx.Exec(ctx, "DELETE FROM post_tags WHERE post_id=$1", id); e != nil {
			httpx.Error(c, 500, "INTERNAL_ERROR")
			return
		}
		for _, t := range tags {
			if _, e = tx.Exec(ctx, "INSERT INTO post_tags(post_id,tag_id) VALUES($1,$2) ON CONFLICT DO NOTHING", id, t); e != nil {
				httpx.Error(c, 422, "VALIDATION_ERROR")
				return
			}
		}
	}
	if s.auditTx(ctx, tx, c, r.table+".save", r.table, id) != nil || tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	status := 200
	if !patch {
		status = 201
	}
	httpx.OK(c, status, gin.H{"id": id})
}
func (s *Server) contentDelete(c *gin.Context, r resource) {
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
	q := "DELETE FROM " + r.table + " WHERE id=$1"
	archive := r.table == "posts" || r.table == "pages"
	if archive {
		q = "UPDATE " + r.table + " SET status='archived',updated_at=now() WHERE id=$1"
	}
	tag, e := tx.Exec(ctx, q, c.Param("id"))
	if e != nil {
		httpx.Error(c, 409, "CONFLICT")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if s.auditTx(ctx, tx, c, r.table+".delete", r.table, c.Param("id")) != nil || tx.Commit(ctx) != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.OK(c, 200, gin.H{"id": c.Param("id"), "deleted": !archive, "archived": archive})
}
func (s *Server) moderateReview(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var b struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if !httpx.Body(c, &b) || b.Status != "approved" && b.Status != "rejected" {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	tag, e := s.DB.Exec(c.Request.Context(), "UPDATE reviews SET status=$2,reviewed_by=$3,reviewed_at=now(),updated_at=now() WHERE id=$1", c.Param("id"), b.Status, s.UserID(c))
	if e != nil || tag.RowsAffected() != 1 {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	httpx.OK(c, 200, gin.H{"id": c.Param("id"), "status": b.Status})
}
func (s *Server) reviews(c *gin.Context) {
	var id string
	if s.DB.QueryRow(c.Request.Context(), "SELECT id FROM tours WHERE slug=$1 AND status='published'", c.Param("slug")).Scan(&id) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	c.Request.URL.RawQuery = "tour_id=" + id + "&page=" + c.DefaultQuery("page", "1") + "&page_size=" + c.DefaultQuery("page_size", "20")
	s.contentList(c, resources[8], false)
}
func (s *Server) home(c *gin.Context) {
	ctx := c.Request.Context()
	t, e := httpx.JSONRows(ctx, s.DB, "SELECT "+tourJSON+" FROM tours t "+bestDeparture+" WHERE status='published' ORDER BY featured DESC,published_at DESC LIMIT 6")
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	d, _ := httpx.JSONRows(ctx, s.DB, "SELECT to_jsonb(d) FROM destinations d WHERE published ORDER BY name LIMIT 6")
	b, _ := httpx.JSONRows(ctx, s.DB, "SELECT to_jsonb(b) FROM banners b WHERE active AND (starts_at IS NULL OR starts_at<=now()) AND (ends_at IS NULL OR ends_at>now()) ORDER BY sort_order LIMIT 6")
	posts, _ := httpx.JSONRows(ctx, s.DB, "SELECT to_jsonb(p)-'body'-'author_id' FROM posts p WHERE status='published' ORDER BY published_at DESC LIMIT 3")
	reviews, _ := httpx.JSONRows(ctx, s.DB, "SELECT to_jsonb(r)-'reviewed_by' FROM reviews r WHERE status='approved' ORDER BY created_at DESC LIMIT 4")
	httpx.OK(c, 200, gin.H{"featured_tours": t, "destinations": d, "banners": b, "latest_posts": posts, "testimonials": reviews, "service_commitments": []string{}})
}
func (s *Server) gallery(c *gin.Context) {
	p, n, valid := httpx.Page(c)
	if !valid {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	kind, collection := c.Query("kind"), c.Query("collection")
	out, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT "+mediaObject+" FROM media_assets m WHERE visibility='public' AND status='ready' AND ($1='' OR kind=$1) AND ($2='' OR collection=$2) ORDER BY created_at DESC LIMIT $3 OFFSET $4", kind, collection, n, (p-1)*n)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	var total int
	s.DB.QueryRow(c.Request.Context(), "SELECT count(*) FROM media_assets WHERE visibility='public' AND status='ready' AND ($1='' OR kind=$1) AND ($2='' OR collection=$2)", kind, collection).Scan(&total)
	httpx.List(c, out, p, n, total)
}
