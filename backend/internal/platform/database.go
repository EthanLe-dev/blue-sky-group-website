package platform

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	c, e := pgxpool.ParseConfig(url)
	if e != nil {
		return nil, fmt.Errorf("invalid database configuration")
	}
	c.MaxConns = 10
	c.MinConns = 0
	p, e := pgxpool.NewWithConfig(ctx, c)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, connectionError(e)
	}
	return p, nil
}

// Return useful diagnostics without including a connection URL or credentials.
func connectionError(err error) error {
	var databaseError *pgconn.PgError
	if errors.As(err, &databaseError) && databaseError.Code == "28P01" {
		return errors.New("database authentication failed: check .env against credentials stored in the existing PostgreSQL volume; POSTGRES_PASSWORD does not reset an initialized database")
	}
	return errors.New("database unavailable: check PostgreSQL readiness and network connectivity")
}
func Migrate(ctx context.Context, p *pgxpool.Pool) error {
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(735291)"); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations(version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); e != nil {
		return e
	}
	entries, _ := migrations.ReadDir("migrations")
	for _, f := range entries {
		if !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		var exists bool
		if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", f.Name()).Scan(&exists); e != nil {
			return e
		}
		if exists {
			continue
		}
		b, e := migrations.ReadFile("migrations/" + f.Name())
		if e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, string(b)); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", f.Name()); e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
