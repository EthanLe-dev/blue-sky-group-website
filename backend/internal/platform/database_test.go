package platform

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestConnectionErrorExplainsAuthenticationWithoutLeakingCredentials(t *testing.T) {
	err := fmt.Errorf("connection failed: %w", &pgconn.PgError{
		Code: "28P01", Message: "sensitive password or connection details",
	})
	message := connectionError(err).Error()
	if !strings.Contains(message, "initialized database") || strings.Contains(message, "sensitive") {
		t.Fatal("authentication diagnostic is missing or leaks underlying details")
	}
	message = connectionError(errors.New("postgres://user:secret@host/db")).Error()
	if strings.Contains(message, "secret") || strings.Contains(message, "postgres://") {
		t.Fatal("network diagnostic leaks the connection URL")
	}
}
