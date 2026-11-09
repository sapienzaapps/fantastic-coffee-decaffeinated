/*
Package database defines the contract between the app and its data layer. All data (de)serialization (save/load) from
a persistent database is hidden behind the AppDatabase interface: the rest of the program (for example the API
handlers) depends only on that interface, never on a concrete database.

Each supported database has its own implementation in a sub-package, with its own SQL and its own migrations:

  - service/database/sqlite: SQLite (a local file), the default
  - service/database/postgres: PostgreSQL

Both constructors open the connection, apply pending migrations and return an AppDatabase:

	db, err := sqlite.New(ctx, "/tmp/decaf.db")
	// or: db, err := postgres.New(ctx, "postgres://user:password@localhost:5432/decaf?sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

To support another database or data layer (for example sqlc or an ORM), add another package that implements
AppDatabase. See the "Database migrations" section of the README for details on migrations.
*/
package database

import "errors"

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error

	Ping() error
	Close() error
}

// ErrNameNotFound is returned by GetName when no name has been stored yet. The
// API layer translates it into an HTTP 404 response.
var ErrNameNotFound = errors.New("name not found")
