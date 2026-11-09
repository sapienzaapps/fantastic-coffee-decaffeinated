/*
Package postgres implements database.AppDatabase on top of PostgreSQL (driver: github.com/jackc/pgx/v5).

Queries in this package are written in PostgreSQL's SQL dialect, with numbered placeholders ($1, $2, ...). The schema
is defined by the migrations in the migrations/ directory, which are embedded in the executable and applied by New.
*/
package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
)

// migrationsFS embeds every .sql file in the migrations directory into the
// executable. The files are the single source of truth for the database schema.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// postgresDB is the PostgreSQL implementation of database.AppDatabase. Its
// methods (GetName, SetName, ...) are defined in the other files of this package.
type postgresDB struct {
	c *sql.DB
}

// New connects to the PostgreSQL server described by `dsn`, applies pending migrations and returns it as a
// database.AppDatabase. Call Close on the returned value when done.
//
// `dsn` is a connection string such as "postgres://user:password@localhost:5432/dbname?sslmode=disable".
func New(ctx context.Context, dsn string) (database.AppDatabase, error) {
	if dsn == "" {
		return nil, errors.New("the PostgreSQL connection string (DSN) is required")
	}

	c, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening PostgreSQL: %w", err)
	}

	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := database.Migrate(ctx, c, migrations,
		`INSERT INTO schema_migrations (version) VALUES ($1)`); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	return &postgresDB{c: c}, nil
}

func (db *postgresDB) Ping() error {
	return db.c.Ping()
}

func (db *postgresDB) Close() error {
	return db.c.Close()
}
