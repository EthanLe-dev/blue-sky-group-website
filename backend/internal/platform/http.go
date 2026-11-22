package platform

import (
	"bytes"
	"context"
	"crypto/hmac"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:embed openapi.json
var openAPIDocument []byte

type App struct {
	Config Config
	DB     *pgxpool.Pool
}

func New(c Config, p *pgxpool.Pool) *App { return &App{c, p} }
func ok(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data, "error": nil, "message": nil})
}
func fail(c *gin.Context, status int, code string) {
	c.AbortWithStatusJSON(status, gin.H{"success": false, "data": nil, "error": gin.H{"code": code, "fields": gin.H{}}, "message": "Không thể thực hiện yêu cầu. Vui lòng kiểm tra và thử lại."})
}
func body(c *gin.Context, v any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		fail(c, 422, "VALIDATION_ERROR")
		return false
	}
	if d.Decode(&struct{}{}) != io.EOF {
		fail(c, 422, "VALIDATION_ERROR")
		return false
	}
	return true
}
func (a *App) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) { fail(c, 500, "INTERNAL_ERROR") }), func(c *gin.Context) {
		start := time.Now()
		id := uuid.NewString()
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Header("Cache-Control", "no-store")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
		slog.Info("http", "request_id", id, "method", c.Request.Method, "status", c.Writer.Status(), "duration_ms", time.Since(start).Milliseconds())
	})
	r.Use(a.cors)
	r.GET("/openapi.json", func(c *gin.Context) { c.Data(200, "application/json", openAPIDocument) })
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		var ready bool
		e := a.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version='000005_payment_notification.sql')").Scan(&ready)
		if e != nil || !ready {
			c.JSON(503, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	v := r.Group("/api/v1")
	v.GET("/site-settings", a.publicSettings)
	v.GET("/booking-defaults", a.publicDefaults)
	v.POST("/guest-sessions", a.limit("guest", 30, 600), a.origin, a.guest)
	auth := v.Group("/admin/auth")
	auth.GET("/session-csrf", a.origin, func(c *gin.Context) {
		t, e := c.Cookie("staff_refresh")
		if e != nil {
			fail(c, 401, "AUTH_REQUIRED")
			return
		}
		ok(c, 200, gin.H{"csrf_token": a.csrf(t)})
	})
	auth.POST("/login", a.limit("login", 5, 60), a.origin, a.login)
	auth.POST("/refresh", a.origin, a.cookieCSRF("staff_refresh"), a.refresh)
	auth.POST("/logout", a.origin, a.cookieCSRF("staff_refresh"), a.logout)
	auth.GET("/me", a.bearer, func(c *gin.Context) { ok(c, 200, c.MustGet("user")) })
	auth.POST("/forgot-password", a.limit("forgot", 5, 900), a.origin, a.forgot)
	auth.POST("/reset-password", a.limit("reset", 5, 900), a.origin, a.reset)
	admin := v.Group("/admin", a.bearer)
	admin.GET("/site-settings/:key", a.permission("settings.write"), a.getSetting)
	admin.PUT("/site-settings/:key", a.permission("settings.write"), a.putSetting)
	admin.GET("/audit-logs", a.permission("audit.read"), a.audit)
	admin.GET("/auth/mock-reset-deliveries", a.permission("staff.write"), a.resetPreviews)
	bookingServer := a.BookingServer()
	bookingServer.Register(v, admin, a.guestAccess, a.origin, a.limit("lookup", 5, 900), a.limit("checkout", 10, 600), a.permission("booking.read"), a.permission("booking.write"))
	if a.Config.Env != "production" && a.Config.MockPayment {
		bookingServer.RegisterPayments(v, admin, a.guestAccess, a.origin, a.limit("payment", 30, 60), a.permission("booking.read"), a.permission("payment.write"))
	}
	if a.Config.Env != "production" && a.Config.MockNotification {
		bookingServer.RegisterNotifications(v, admin, a.guestAccess, a.permission("notification.read"))
	}
	catalogServer := &catalog.Server{ExpireLocked: booking.ExpireLocked, Expire: bookingServer.Expire, DB: a.DB, UserID: func(c *gin.Context) string { return c.MustGet("user").(User).ID }, Media: catalog.NewMediaStore(a.Config.CSRFSecret)}
	catalogServer.Register(v, admin, a.permission("catalog.write"), a.permission("content.write"), a.permission("media.write"), a.permission("media.private.write"))
	r.NoRoute(func(c *gin.Context) { fail(c, 404, "NOT_FOUND") })
	return r
}
func (a *App) cors(c *gin.Context) {
	o := c.GetHeader("Origin")
	for _, allowed := range a.Config.Origins {
		if o == allowed {
			c.Header("Access-Control-Allow-Origin", o)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-CSRF-Token,Idempotency-Key")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		}
	}
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}
func (a *App) origin(c *gin.Context) {
	if c.Request.Method == http.MethodGet && c.GetHeader("Origin") == "" && c.GetHeader("Sec-Fetch-Site") == "same-origin" {
		ref, e := url.Parse(c.GetHeader("Referer"))
		if e == nil {
			for _, o := range a.Config.Origins {
				if ref.Scheme+"://"+ref.Host == o {
					c.Next()
					return
				}
			}
		}
	}

	for _, o := range a.Config.Origins {
		if o == c.GetHeader("Origin") {
			c.Next()
			return
		}
	}
	fail(c, 403, "FORBIDDEN")
}
func (a *App) cookieCSRF(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, e := c.Cookie(name)
		if e != nil || !hmac.Equal([]byte(a.csrf(token)), []byte(c.GetHeader("X-CSRF-Token"))) {
			fail(c, 403, "CSRF_INVALID")
			return
		}
		c.Next()
	}
}
func (a *App) cookie(c *gin.Context, name, value string, age int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, age, "/", "", a.Config.Secure, true)
}
func (a *App) limit(scope string, max, seconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal := ""
		if scope == "checkout" {
			principal = c.GetString("guest_id")
		}
		if scope == "lookup" {
			raw, e := io.ReadAll(io.LimitReader(c.Request.Body, 256*1024+1))
			if e != nil || len(raw) > 256*1024 {
				fail(c, 413, "PAYLOAD_TOO_LARGE")
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			var input struct {
				Code string `json:"code"`
			}
			if json.Unmarshal(raw, &input) != nil || len(input.Code) > 24 {
				fail(c, 422, "VALIDATION_ERROR")
				return
			}
			principal = strings.ToUpper(strings.TrimSpace(input.Code))
		}

		key := hash(scope + ":" + c.ClientIP() + ":" + principal)
		var count int
		e := a.DB.QueryRow(c.Request.Context(), `INSERT INTO rate_limit_buckets(key_hash,window_start,count,expires_at) VALUES($1,now(),1,now()+$2*interval '1 second') ON CONFLICT(key_hash) DO UPDATE SET count=CASE WHEN rate_limit_buckets.expires_at<=now() THEN 1 ELSE rate_limit_buckets.count+1 END,expires_at=CASE WHEN rate_limit_buckets.expires_at<=now() THEN now()+$2*interval '1 second' ELSE rate_limit_buckets.expires_at END RETURNING count`, key, seconds).Scan(&count)
		if e != nil {
			fail(c, 503, "SERVICE_UNAVAILABLE")
			return
		}
		if count > max {
			c.Header("Retry-After", fmt.Sprint(seconds))
			fail(c, 429, "RATE_LIMITED")
			return
		}
		c.Next()
	}
}
func (a *App) captcha(ctx context.Context, token, action string) bool {
	if a.Config.MockCaptcha {
		return (a.Config.Env == "local" || a.Config.Env == "test") && token == "local-test-token"
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	form := url.Values{"secret": {a.Config.TurnstileSecret}, "response": {token}}
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://challenges.cloudflare.com/turnstile/v0/siteverify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return false
	}
	defer resp.Body.Close()
	var out struct {
		Success  bool   `json:"success"`
		Action   string `json:"action"`
		Hostname string `json:"hostname"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&out) != nil || !out.Success || out.Action != action {
		return false
	}
	for _, h := range a.Config.Hostnames {
		if h == out.Hostname {
			return true
		}
	}
	return false
}
func (a *App) bearer(c *gin.Context) {
	raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	t, e := jwt.Parse(raw, func(t *jwt.Token) (any, error) { return []byte(a.Config.JWTSecret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("blue-sky-api"), jwt.WithAudience("blue-sky-admin"), jwt.WithExpirationRequired())
	if e != nil || !t.Valid {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	claims := t.Claims.(jwt.MapClaims)
	id, _ := claims["sub"].(string)
	sid, _ := claims["sid"].(string)
	if uuid.Validate(id) != nil || uuid.Validate(sid) != nil {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	var valid bool
	e = a.DB.QueryRow(c.Request.Context(), "SELECT EXISTS(SELECT 1 FROM staff_sessions WHERE id=$1 AND staff_id=$2 AND revoked_at IS NULL AND expires_at>now())", sid, id).Scan(&valid)
	if e != nil || !valid {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	u, e := a.user(c.Request.Context(), id)
	if e != nil {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	c.Set("user", u)
	c.Next()
}
func (a *App) permission(p string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.MustGet("user").(User)
		for _, v := range u.Permissions {
			if v == p {
				c.Next()
				return
			}
		}
		fail(c, 403, "FORBIDDEN")
	}
}
func (a *App) login(c *gin.Context) {
	var b struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Captcha  string `json:"turnstile_token"`
	}
	if !body(c, &b) {
		return
	}
	if !a.captcha(c.Request.Context(), b.Captcha, "login") {
		fail(c, 403, "CAPTCHA_INVALID")
		return
	}
	if len(b.Password) > 72 || len(b.Email) > 254 {
		fail(c, 422, "VALIDATION_ERROR")
		return
	}
	var id, password string
	e := a.DB.QueryRow(c.Request.Context(), "SELECT id,password_hash FROM staff_users WHERE lower(email)=$1 AND status='active'", strings.ToLower(strings.TrimSpace(b.Email))).Scan(&id, &password)
	if e != nil {
		password = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}
	valid := bcrypt.CompareHashAndPassword([]byte(password), []byte(b.Password)) == nil
	if e != nil || !valid {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	token, sid, e := a.createSession(c.Request.Context(), id)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	a.issue(c, id, sid, token)
}
func (a *App) issue(c *gin.Context, id, sid, token string) {
	access, e := a.access(id, sid)
	if e != nil {
		fail(c, 500, "INTERNAL_ERROR")
		return
	}
	u, e := a.user(c.Request.Context(), id)
	if e != nil {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	a.cookie(c, "staff_refresh", token, 604800)
	ok(c, 200, gin.H{"access_token": access, "token_type": "Bearer", "expires_in": 900, "csrf_token": a.csrf(token), "user": u})
}
func (a *App) refresh(c *gin.Context) {
	token, _ := c.Cookie("staff_refresh")
	ctx := c.Request.Context()
	tx, e := a.DB.Begin(ctx)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer tx.Rollback(ctx)
	var id, user, family string
	var revoked bool
	e = tx.QueryRow(ctx, "SELECT id,staff_id,family_id,revoked_at IS NOT NULL FROM staff_sessions WHERE token_hash=$1 AND expires_at>now() FOR UPDATE", hash(token)).Scan(&id, &user, &family, &revoked)
	if e != nil {
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	if revoked {
		if _, e = tx.Exec(ctx, "UPDATE staff_sessions SET revoked_at=now() WHERE family_id=$1", family); e == nil {
			e = tx.Commit(ctx)
		}
		fail(c, 401, "AUTH_REQUIRED")
		return
	}
	next, nid := secret(), uuid.NewString()
	if _, e = tx.Exec(ctx, "INSERT INTO staff_sessions(id,staff_id,token_hash,family_id,expires_at) VALUES($1,$2,$3,$4,now()+interval '7 days')", nid, user, hash(next), family); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	if _, e = tx.Exec(ctx, "UPDATE staff_sessions SET revoked_at=now(),replaced_by=$2 WHERE id=$1", id, nid); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	if e = tx.Commit(ctx); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	a.issue(c, user, nid, next)
}
func (a *App) logout(c *gin.Context) {
	t, _ := c.Cookie("staff_refresh")
	_, e := a.DB.Exec(c.Request.Context(), "UPDATE staff_sessions SET revoked_at=now() WHERE family_id IN(SELECT family_id FROM staff_sessions WHERE token_hash=$1)", hash(t))
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	a.cookie(c, "staff_refresh", "", -1)
	ok(c, 200, gin.H{"logged_out": true})
}
func (a *App) guest(c *gin.Context) {
	token, _ := c.Cookie("guest_session")
	var expires time.Time
	if token != "" {
		e := a.DB.QueryRow(c.Request.Context(), "SELECT expires_at FROM guest_sessions WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at>now()", hash(token)).Scan(&expires)
		if e == nil {
			ok(c, 200, gin.H{"csrf_token": a.csrf(token), "expires_at": expires})
			return
		}
	}
	token = secret()
	e := a.DB.QueryRow(c.Request.Context(), "INSERT INTO guest_sessions(token_hash,expires_at) VALUES($1,now()+interval '24 hours') RETURNING expires_at", hash(token)).Scan(&expires)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	a.cookie(c, "guest_session", token, 86400)
	ok(c, 201, gin.H{"csrf_token": a.csrf(token), "expires_at": expires})
}
func (a *App) forgot(c *gin.Context) {
	var b struct {
		Email   string `json:"email"`
		Captcha string `json:"turnstile_token"`
	}
	if !body(c, &b) {
		return
	}
	if !a.captcha(c.Request.Context(), b.Captcha, "forgot-password") {
		fail(c, 403, "CAPTCHA_INVALID")
		return
	}
	if a.Config.Env == "production" {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	ctx := c.Request.Context()
	tx, e := a.DB.Begin(ctx)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer tx.Rollback(ctx)
	var id string
	e = tx.QueryRow(ctx, "SELECT id FROM staff_users WHERE lower(email)=$1 AND status='active'", strings.ToLower(strings.TrimSpace(b.Email))).Scan(&id)
	if e == nil {
		token := secret()
		encrypted, err := a.encrypt(token)
		if err != nil {
			fail(c, 500, "INTERNAL_ERROR")
			return
		}
		if _, e = tx.Exec(ctx, "INSERT INTO password_reset_tokens(staff_id,token_hash,expires_at) VALUES($1,$2,now()+interval '15 minutes')", id, hash(token)); e != nil {
			fail(c, 503, "SERVICE_UNAVAILABLE")
			return
		}
		if _, e = tx.Exec(ctx, "INSERT INTO mock_password_deliveries(staff_id,encrypted_token,expires_at) VALUES($1,$2,now()+interval '15 minutes')", id, encrypted); e != nil {
			fail(c, 503, "SERVICE_UNAVAILABLE")
			return
		}
	}
	if e = tx.Commit(ctx); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	ok(c, 202, gin.H{"accepted": true})
}
func (a *App) reset(c *gin.Context) {
	var b struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !body(c, &b) {
		return
	}
	if len(b.Password) < 12 || len(b.Password) > 72 {
		fail(c, 422, "VALIDATION_ERROR")
		return
	}
	pw, e := bcrypt.GenerateFromPassword([]byte(b.Password), bcrypt.DefaultCost)
	if e != nil {
		fail(c, 500, "INTERNAL_ERROR")
		return
	}
	ctx := c.Request.Context()
	tx, e := a.DB.Begin(ctx)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer tx.Rollback(ctx)
	var uid string
	e = tx.QueryRow(ctx, "UPDATE password_reset_tokens SET used_at=now() WHERE token_hash=$1 AND used_at IS NULL AND expires_at>now() RETURNING staff_id", hash(b.Token)).Scan(&uid)
	if e != nil {
		fail(c, 422, "VALIDATION_ERROR")
		return
	}
	for _, q := range []string{"UPDATE staff_users SET password_hash=$2,updated_at=now() WHERE id=$1", "UPDATE staff_sessions SET revoked_at=now() WHERE staff_id=$1", "UPDATE password_reset_tokens SET used_at=now() WHERE staff_id=$1", "DELETE FROM mock_password_deliveries WHERE staff_id=$1"} {
		if strings.Contains(q, "password_hash") {
			_, e = tx.Exec(ctx, q, uid, string(pw))
		} else {
			_, e = tx.Exec(ctx, q, uid)
		}
		if e != nil {
			fail(c, 503, "SERVICE_UNAVAILABLE")
			return
		}
	}
	if e = tx.Commit(ctx); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	ok(c, 200, gin.H{"reset": true})
}
func (a *App) resetPreviews(c *gin.Context) {
	if a.Config.Env == "production" {
		fail(c, 404, "NOT_FOUND")
		return
	}
	rows, e := a.DB.Query(c.Request.Context(), "SELECT id,staff_id,encrypted_token,expires_at FROM mock_password_deliveries WHERE expires_at>now() ORDER BY created_at DESC LIMIT 100")
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, staff, enc string
		var expires time.Time
		if rows.Scan(&id, &staff, &enc, &expires) != nil {
			fail(c, 500, "INTERNAL_ERROR")
			return
		}
		token, e := a.decrypt(enc)
		if e != nil {
			fail(c, 500, "INTERNAL_ERROR")
			return
		}
		out = append(out, gin.H{"id": id, "staff_id": staff, "token": token, "expires_at": expires, "is_mock": true})
	}
	u := c.MustGet("user").(User)
	_, e = a.DB.Exec(c.Request.Context(), "INSERT INTO audit_logs(actor_staff_id,action,entity_type,request_id) VALUES($1,'mock_reset_preview','password_reset',$2)", u.ID, c.GetString("request_id"))
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	ok(c, 200, out)
}
func (a *App) audit(c *gin.Context) {
	rows, e := a.DB.Query(c.Request.Context(), "SELECT id,action,entity_type,created_at FROM audit_logs ORDER BY created_at DESC LIMIT 100")
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, action, entity string
		var at time.Time
		if rows.Scan(&id, &action, &entity, &at) != nil {
			fail(c, 500, "INTERNAL_ERROR")
			return
		}
		out = append(out, gin.H{"id": id, "action": action, "entity_type": entity, "created_at": at})
	}
	ok(c, 200, out)
}

func (a *App) guestAccess(c *gin.Context) {
	token, _ := c.Cookie("guest_session")
	var id string
	if token == "" || a.DB.QueryRow(c.Request.Context(), "SELECT id FROM guest_sessions WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at>clock_timestamp()", hash(token)).Scan(&id) != nil {
		fail(c, 401, "UNAUTHORIZED")
		return
	}
	if c.Request.Method != "GET" && !hmac.Equal([]byte(c.GetHeader("X-CSRF-Token")), []byte(a.csrf(token))) {
		fail(c, 403, "CSRF_FAILED")
		return
	}
	c.Set("guest_id", id)
	c.Next()
}
func (a *App) BookingServer() *booking.Server {
	return &booking.Server{DB: a.DB, Captcha: a.captcha, Principal: func(c *gin.Context, staff bool) (booking.Principal, bool) {
		if staff {
			u, ok := c.Get("user")
			if !ok {
				fail(c, 401, "UNAUTHORIZED")
				return booking.Principal{}, false
			}
			return booking.Principal{ID: u.(User).ID, Staff: true}, true
		}
		id := c.GetString("guest_id")
		if id == "" {
			fail(c, 401, "UNAUTHORIZED")
			return booking.Principal{}, false
		}
		return booking.Principal{ID: id}, true
	}}
}
