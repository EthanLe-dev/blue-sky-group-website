package platform

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Env, Address, DatabaseURL, JWTSecret, CSRFSecret, TurnstileSecret string
	Origins, Hostnames                                                []string
	MockPayment, MockNotification, MockCaptcha, Secure                bool
}

func LoadConfig() (Config, error) {
	c := Config{Env: os.Getenv("APP_ENV"), Address: os.Getenv("HOST") + ":" + os.Getenv("PORT"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SIGNING_SECRET"), CSRFSecret: os.Getenv("CSRF_SECRET"), TurnstileSecret: os.Getenv("TURNSTILE_SECRET_KEY"), Origins: strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","), Hostnames: strings.Split(os.Getenv("TURNSTILE_ALLOWED_HOSTNAMES"), ","), MockCaptcha: os.Getenv("TURNSTILE_MODE") == "mock", Secure: os.Getenv("COOKIE_SECURE") == "true"}
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
