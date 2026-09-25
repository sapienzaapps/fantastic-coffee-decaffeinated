package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// migrationsFS embeds every .sql file in the migrations directory into the
// executable. The files are the single source of truth for the database schema.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies all pending SQL migrations to db. It is safe to call at every
// startup: already-applied migrations are recorded in the schema_migrations
// table and skipped.
//
// Migrations are applied in lexicographic order of their file name, so the
// numeric prefix (0001_, 0002_, ...) defines the execution order. Each file is
// executed inside a single transaction together with the row that records it as
// applied, so a failed migration leaves the database unchanged.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT NOT NULL PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
)`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}

	files, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("listing embedded migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(path.Base(file), ".sql")
		if applied[version] {
			continue
		}

		script, err := migrationsFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", version, err)
		}
		if err := applyMigration(ctx, db, version, string(script)); err != nil {
			return err
		}
	}

	return nil
}

// appliedVersions returns the set of migration versions already recorded in the
// database.
func appliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

// applyMigration executes a single migration script and records it as applied,
// atomically.
func applyMigration(ctx context.Context, db *sql.DB, version, script string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction for migration %s: %w", version, err)
	}
	// Rollback is a no-op once the transaction has been committed.
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("executing migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return fmt.Errorf("recording migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration %s: %w", version, err)
	}
	return nil
}
