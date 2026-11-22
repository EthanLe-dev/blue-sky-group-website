package platform

import (
	"strings"
	"testing"
)

func TestDemoRejectsLocalStorageAndInsecureDatabase(t *testing.T) {
	for key, value := range map[string]string{
		"APP_ENV": "demo", "COOKIE_SECURE": "true", "TURNSTILE_MODE": "real",
		"TURNSTILE_SECRET_KEY": "test-key", "TURNSTILE_ALLOWED_HOSTNAMES": "demo.example,admin.example",
		"CORS_ALLOWED_ORIGINS": "https://demo.example,https://admin.example",
		"JWT_SIGNING_SECRET":   strings.Repeat("j", 32), "CSRF_SECRET": strings.Repeat("c", 32),
		"DATABASE_URL":           "postgresql://user:pass@pool.example/demo?sslmode=require",
		"DATABASE_MIGRATION_URL": "postgresql://user:pass@direct.example/demo?sslmode=require",
		"STORAGE_PROVIDER":       "r2", "R2_ENDPOINT": "https://r2.example", "R2_ACCESS_KEY_ID": "example",
		"R2_SECRET_ACCESS_KEY": "example", "R2_PUBLIC_BUCKET": "public", "R2_PRIVATE_BUCKET": "private",
		"RENDER_GIT_COMMIT": strings.Repeat("a", 40),
	} {
		t.Setenv(key, value)
	}
	cfg, err := LoadConfig()
	if err != nil || cfg.Revision != strings.Repeat("a", 40) {
		t.Fatalf("valid demo config: %+v", err)
	}
	for _, tc := range []struct{ key, value string }{
		{"STORAGE_PROVIDER", "local"}, {"DATABASE_MIGRATION_URL", "postgresql://localhost/demo?sslmode=disable"},
		{"R2_PRIVATE_BUCKET", "public"}, {"CORS_ALLOWED_ORIGINS", "http://demo.example"},
		{"TURNSTILE_MODE", "mock"}, {"TURNSTILE_ALLOWED_HOSTNAMES", ""},
	} {
		t.Run(tc.key, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := LoadConfig(); err == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}
}
