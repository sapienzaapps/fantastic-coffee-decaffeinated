package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Migrate applies all pending SQL migrations in the migrations file system to
// db. It is shared by the database implementations (see the sqlite and postgres
// sub-packages), each of which embeds its own migrations directory and passes it
// here. It is safe to call at every startup: already-applied migrations are
// recorded in the schema_migrations table and skipped.
//
// Migrations are the *.sql files at the root of migrations, applied in
// lexicographic order of their file name, so the numeric prefix (0001_, 0002_,
// ...) defines the execution order. Each file is executed inside a single
// transaction together with the row that records it as applied, so a failed
// migration leaves the database unchanged.
//
// recordSQL is the statement that records a version as applied, written with
// the placeholder style of the database, for example
// "INSERT INTO schema_migrations (version) VALUES (?)" for SQLite or
// "INSERT INTO schema_migrations (version) VALUES ($1)" for PostgreSQL.
func Migrate(ctx context.Context, db *sql.DB, migrations fs.FS, recordSQL string) error {
	// TIMESTAMP and CURRENT_TIMESTAMP are understood by both SQLite and PostgreSQL.
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT NOT NULL PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}

	files, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(path.Base(file), ".sql")
		if applied[version] {
			continue
		}

		script, err := fs.ReadFile(migrations, file)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", version, err)
		}
		if err := applyMigration(ctx, db, recordSQL, version, string(script)); err != nil {
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
func applyMigration(ctx context.Context, db *sql.DB, recordSQL, version, script string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction for migration %s: %w", version, err)
	}
	// Rollback is a no-op once the transaction has been committed.
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("executing migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, recordSQL, version); err != nil {
		return fmt.Errorf("recording migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration %s: %w", version, err)
	}
	return nil
}
