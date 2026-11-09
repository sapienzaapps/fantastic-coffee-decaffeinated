/*
Package sqlite implements database.AppDatabase on top of SQLite, using a local database file
(driver: github.com/mattn/go-sqlite3).

Queries in this package are written in SQLite's SQL dialect, with "?" placeholders. The schema is defined by the
migrations in the migrations/ directory, which are embedded in the executable and applied by New.
*/
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	_ "github.com/mattn/go-sqlite3" // registers the "sqlite3" driver
)

// migrationsFS embeds every .sql file in the migrations directory into the
// executable. The files are the single source of truth for the database schema.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// sqliteDB is the SQLite implementation of database.AppDatabase. Its methods
// (GetName, SetName, ...) are defined in the other files of this package.
type sqliteDB struct {
	c *sql.DB
}

// New opens the SQLite database stored in `filename` (created if it does not exist), applies pending migrations and
// returns it as a database.AppDatabase. Call Close on the returned value when done.
func New(ctx context.Context, filename string) (database.AppDatabase, error) {
	if filename == "" {
		return nil, errors.New("the SQLite database filename is required")
	}

	c, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, fmt.Errorf("opening SQLite: %w", err)
	}
	// SQLite allows a single writer at a time: one connection avoids
	// "database is locked" errors under concurrent requests.
	c.SetMaxOpenConns(1)

	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := database.Migrate(ctx, c, migrations,
		`INSERT INTO schema_migrations (version) VALUES (?)`); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	return &sqliteDB{c: c}, nil
}

func (db *sqliteDB) Ping() error {
	return db.c.Ping()
}

func (db *sqliteDB) Close() error {
	return db.c.Close()
}
