package platform

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestConfigRejectsMissingSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SIGNING_SECRET", "")
	if _, e := LoadConfig(); e == nil {
		t.Fatal("accepted missing secret")
	}
}
func testApp(t *testing.T) *App {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		if os.Getenv("CI")=="true"{t.Fatal("CI requires TEST_DATABASE_URL")};t.Skip("TEST_DATABASE_URL unset: integration not run")
	}
	parsed, err := pgxpool.ParseConfig(url)
	if err != nil || !strings.HasSuffix(parsed.ConnConfig.Database, "_test") {
		t.Fatal("integration requires a database ending in _test")
	}
	ctx := context.Background()
	base, e := pgxpool.New(ctx, url)
	if e != nil {
		t.Fatal(e)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, e = base.Exec(ctx, "CREATE SCHEMA "+schema); e != nil {
		t.Fatal(e)
	}
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	p, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { p.Close(); base.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); base.Close() })
	if e = Migrate(ctx, p); e != nil {
		t.Fatal(e)
	}
	if e = Migrate(ctx, p); e != nil {
		t.Fatal(e)
	}
	if e = Bootstrap(ctx, p, "admin@example.com", "test-password-123"); e != nil {
		t.Fatal(e)
	}
	if e = Bootstrap(ctx, p, "admin@example.com", "test-password-123"); e != nil {
		t.Fatal(e)
	}
	return New(Config{Env: "test", JWTSecret: strings.Repeat("a", 32), CSRFSecret: strings.Repeat("b", 32), MockCaptcha: true, Origins: []string{"http://localhost:5173"}}, p)
}
func request(a *App, method, path, body, token, csrf string, cookie *http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Origin", "http://localhost:5173")
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, r)
	return w
}
func data(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out struct {
		Data map[string]any `json:"data"`
	}
	if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	return out.Data
}
func TestAuthRotationGuestAndReset(t *testing.T) {
	a := testApp(t)
	login := func(password string) *httptest.ResponseRecorder {
		return request(a, "POST", "/api/v1/admin/auth/login", `{"email":"admin@example.com","password":"`+password+`","turnstile_token":"local-test-token"}`, "", "", nil)
	}
	if w := login("wrong-password"); w.Code != 401 {
		t.Fatalf("wrong password: %d %s", w.Code, w.Body)
	}
	w := login("test-password-123")
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	d := data(t, w)
	token := d["access_token"].(string)
	csrf := d["csrf_token"].(string)
	cookie := w.Result().Cookies()[0]
	if w = request(a, "GET", "/api/v1/admin/auth/me", "", token, "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w = request(a, "POST", "/api/v1/admin/auth/refresh", `{}`, "", "bad", cookie); w.Code != 403 {
		t.Fatal("csrf accepted")
	}
	w = request(a, "POST", "/api/v1/admin/auth/refresh", `{}`, "", csrf, cookie)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	newToken := data(t, w)["access_token"].(string)
	if w = request(a, "POST", "/api/v1/admin/auth/refresh", `{}`, "", csrf, cookie); w.Code != 401 {
		t.Fatal("reuse accepted")
	}
	if w = request(a, "GET", "/api/v1/admin/auth/me", "", newToken, "", nil); w.Code != 401 {
		t.Fatal("family not revoked")
	}
	w = request(a, "POST", "/api/v1/guest-sessions", `{}`, "", "", nil)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	g := w.Result().Cookies()[0]
	w = request(a, "POST", "/api/v1/guest-sessions", `{}`, "", "", g)
	if w.Code != 200 {
		t.Fatal("guest not reused")
	}
	w = request(a, "POST", "/api/v1/admin/auth/forgot-password", `{"email":"admin@example.com","turnstile_token":"local-test-token"}`, "", "", nil)
	if w.Code != 202 {
		t.Fatal(w.Body)
	}
	var enc string
	if e := a.DB.QueryRow(context.Background(), "SELECT encrypted_token FROM mock_password_deliveries LIMIT 1").Scan(&enc); e != nil {
		t.Fatal(e)
	}
	resetToken, e := a.decrypt(enc)
	if e != nil {
		t.Fatal(e)
	}
	w = request(a, "POST", "/api/v1/admin/auth/reset-password", `{"token":"`+resetToken+`","password":"new-password-123"}`, "", "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w = request(a, "POST", "/api/v1/admin/auth/reset-password", `{"token":"`+resetToken+`","password":"new-password-456"}`, "", "", nil); w.Code != 422 {
		t.Fatal("reset reused")
	}
}
func TestRBACLimiterAndReadiness(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	var id string
	a.DB.QueryRow(ctx, "SELECT id FROM staff_users LIMIT 1").Scan(&id)
	a.DB.Exec(ctx, "DELETE FROM staff_roles WHERE staff_id=$1", id)
	a.DB.Exec(ctx, "INSERT INTO staff_roles SELECT $1,id,now() FROM roles WHERE code='executive'", id)
	_, sid, e := a.createSession(ctx, id)
	if e != nil {
		t.Fatal(e)
	}
	token, _ := a.access(id, sid)
	if w := request(a, "GET", "/api/v1/admin/audit-logs", "", token, "", nil); w.Code != 403 {
		t.Fatal("executive read audit")
	}
	if w := request(a, "GET", "/readyz", "", "", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	for i := 0; i < 31; i++ {
		w := request(a, "POST", "/api/v1/guest-sessions", `{}`, "", "", nil)
		if i == 30 && w.Code != 429 {
			t.Fatal("rate limit not persisted")
		}
	}
	restarted := New(a.Config, a.DB)
	if w := request(restarted, "POST", "/api/v1/guest-sessions", `{}`, "", "", nil); w.Code != 429 {
		t.Fatal("rate limit reset after restart")
	}
}

func TestSettingsVersionValidationAndAudit(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	var id string
	if e := a.DB.QueryRow(ctx, "SELECT id FROM staff_users LIMIT 1").Scan(&id); e != nil {
		t.Fatal(e)
	}
	_, sid, e := a.createSession(ctx, id)
	if e != nil {
		t.Fatal(e)
	}
	token, _ := a.access(id, sid)
	w := request(a, "GET", "/api/v1/admin/site-settings/response_sla", "", token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	input := `{"value":{"business_days":2,"message":"Demo SLA"},"expected_version":1}`
	w = request(a, "PUT", "/api/v1/admin/site-settings/response_sla", input, token, "", nil)
	if w.Code != 200 {
		t.Fatal(w.Body)
	}
	w = request(a, "PUT", "/api/v1/admin/site-settings/response_sla", input, token, "", nil)
	if w.Code != 409 {
		t.Fatal("stale version accepted")
	}
	w = request(a, "PUT", "/api/v1/admin/site-settings/response_sla", `{"value":{"business_days":0,"message":"bad"},"expected_version":2}`, token, "", nil)
	if w.Code != 422 {
		t.Fatal("invalid setting accepted")
	}
	var count int
	a.DB.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE action='settings.update'").Scan(&count)
	if count != 1 {
		t.Fatal("audit not atomic")
	}
	if validateSetting("booking_defaults", []byte(`{"hold_minutes":99}`)) {
		t.Fatal("invalid hold")
	}
	w = request(a, "GET", "/api/v1/site-settings", "", "", "", nil)
	if w.Code != 200 || strings.Contains(w.Body.String(), "updated_by") {
		t.Fatal("public settings leak")
	}
}
func TestUnknownFieldsOriginExpiredAndProduction(t *testing.T) {
	a := testApp(t)
	w := request(a, "POST", "/api/v1/admin/auth/login", `{"email":"a","extra":true}`, "", "", nil)
	if w.Code != 422 {
		t.Fatal("unknown fields")
	}
	r := httptest.NewRequest("POST", "/api/v1/guest-sessions", strings.NewReader("{}"))
	r.Header.Set("Origin", "https://attacker.example")
	rec := httptest.NewRecorder()
	a.Router().ServeHTTP(rec, r)
	if rec.Code != 403 {
		t.Fatal("origin")
	}
	ctx := context.Background()
	var id string
	a.DB.QueryRow(ctx, "SELECT id FROM staff_users LIMIT 1").Scan(&id)
	_, sid, _ := a.createSession(ctx, id)
	a.DB.Exec(ctx, "UPDATE staff_sessions SET expires_at=now()-interval '1 second' WHERE id=$1", sid)
	token, _ := a.access(id, sid)
	if w = request(a, "GET", "/api/v1/admin/auth/me", "", token, "", nil); w.Code != 401 {
		t.Fatal("expired session")
	}
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("JWT_SIGNING_SECRET", strings.Repeat("a", 32))
	t.Setenv("CSRF_SECRET", strings.Repeat("b", 32))
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("TURNSTILE_SECRET_KEY", "key")
	t.Setenv("TURNSTILE_MODE", "real")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://demo.example")
	t.Setenv("PAYMENT_PROVIDER", "mock")
	if _, e := LoadConfig(); e == nil {
		t.Fatal("production mock accepted")
	}
}

func TestMigrationRollbackAndForeignKeys(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	var count int
	a.DB.QueryRow(ctx, "SELECT count(*) FROM roles").Scan(&count)
	if count != 5 {
		t.Fatal("role seed not idempotent")
	}
	if _, e := a.DB.Exec(ctx, "INSERT INTO staff_roles(staff_id,role_id) VALUES(gen_random_uuid(),gen_random_uuid())"); e == nil {
		t.Fatal("missing FK")
	}
	tx, e := a.DB.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	tx.Exec(ctx, "CREATE TABLE rollback_probe(id int)")
	tx.Rollback(ctx)
	var exists bool
	a.DB.QueryRow(ctx, "SELECT to_regclass('rollback_probe') IS NOT NULL").Scan(&exists)
	if exists {
		t.Fatal("DDL rollback failed")
	}
}
