/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

The database schema is defined by the SQL migrations embedded in this package
(see the migrations/ directory and the Migrate function). Apply them before
using an AppDatabase instance:

	// Apply pending schema migrations first.
	if err := database.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	appdb, err := database.New(db)

The concrete implementation returned by New must satisfy the AppDatabase
interface. The rest of the program depends only on that interface, so a
different data layer (for example, sqlc or an ORM) can be plugged in by
providing another implementation of AppDatabase. See the "Database migrations"
section of the README for details.
*/
package database

import (
	"database/sql"
	"errors"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
//
// New does not create or update the schema: call Migrate before New so that the
// tables the methods rely on actually exist.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
