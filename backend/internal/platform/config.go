package platform

import (
	"errors"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	Env, Address, DatabaseURL, JWTSecret, CSRFSecret, TurnstileSecret string
	Revision                                                          string
	Origins, Hostnames                                                []string
	MockPayment, MockNotification, MockCaptcha, Secure                bool
}

func LoadConfig() (Config, error) {
	c := Config{Env: os.Getenv("APP_ENV"), Address: os.Getenv("HOST") + ":" + os.Getenv("PORT"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SIGNING_SECRET"), CSRFSecret: os.Getenv("CSRF_SECRET"), TurnstileSecret: os.Getenv("TURNSTILE_SECRET_KEY"), Origins: strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","), Hostnames: strings.Split(os.Getenv("TURNSTILE_ALLOWED_HOSTNAMES"), ","), MockCaptcha: os.Getenv("TURNSTILE_MODE") == "mock", Secure: os.Getenv("COOKIE_SECURE") == "true"}
	if revision := os.Getenv("RENDER_GIT_COMMIT"); commitSHA.MatchString(revision) {
		c.Revision = revision
	}
	c.MockPayment = os.Getenv("PAYMENT_PROVIDER") == "mock"
	c.MockNotification = os.Getenv("EMAIL_PROVIDER") == "mock" && os.Getenv("SMS_PROVIDER") == "mock"
	if c.Env != "local" && c.Env != "test" && c.Env != "demo" && c.Env != "production" {
		return c, errors.New("APP_ENV invalid")
	}
	if c.Address == ":" {
		c.Address = "0.0.0.0:8080"
	}
	if c.DatabaseURL == "" || len(c.JWTSecret) < 32 || len(c.CSRFSecret) < 32 {
		return c, errors.New("database and independent secrets of at least 32 characters required")
	}
	if c.JWTSecret == c.CSRFSecret {
		return c, errors.New("secrets must be independent")
	}
	if c.Env != "local" && c.Env != "test" && (!c.Secure || c.MockCaptcha || c.TurnstileSecret == "") {
		return c, errors.New("public environment requires secure cookies and real captcha")
	}
	if c.Env == "demo" {
		if err := validateCloudDemo(c); err != nil {
			return c, err
		}
	}
	if c.Env == "production" {
		for _, k := range []string{"PAYMENT_PROVIDER", "EMAIL_PROVIDER", "SMS_PROVIDER"} {
			if os.Getenv(k) == "" || os.Getenv(k) == "mock" {
				return c, errors.New("production rejects unset/mock providers")
			}
		}
	}
	if len(c.Origins) == 0 || c.Origins[0] == "" {
		return c, errors.New("explicit origins required")
	}
	for _, o := range c.Origins {
		if o == "*" {
			return c, errors.New("wildcard origin forbidden")
		}
	}
	return c, nil
}

var commitSHA = regexp.MustCompile(`^[a-f0-9]{40}$`)

func httpsOrigin(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" &&
		u.User == nil && u.Path == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == ""
}

func validateCloudDemo(c Config) error {
	for _, raw := range []string{c.DatabaseURL, os.Getenv("DATABASE_MIGRATION_URL")} {
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "postgresql" && u.Scheme != "postgres") || u.Hostname() == "" {
			return errors.New("demo requires pooled and direct PostgreSQL URLs")
		}
		switch u.Query().Get("sslmode") {
		case "require", "verify-ca", "verify-full":
		default:
			return errors.New("demo database URLs must require TLS")
		}
	}
	if strings.TrimSpace(c.Hostnames[0]) == "" {
		return errors.New("demo needs an explicit Turnstile hostname")
	}
	for _, origin := range c.Origins {
		if !httpsOrigin(origin) {
			return errors.New("demo needs explicit HTTPS origins")
		}
	}
	if os.Getenv("STORAGE_PROVIDER") != "r2" || !httpsOrigin(os.Getenv("R2_ENDPOINT")) {
		return errors.New("demo requires an HTTPS R2 endpoint")
	}
	for _, key := range []string{"R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_PUBLIC_BUCKET", "R2_PRIVATE_BUCKET"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return errors.New("demo missing " + key)
		}
	}
	if os.Getenv("R2_PUBLIC_BUCKET") == os.Getenv("R2_PRIVATE_BUCKET") {
		return errors.New("demo public and private buckets must differ")
	}
	return nil
}
