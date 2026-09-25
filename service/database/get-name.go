package database

import (
	"database/sql"
	"errors"
)

// ErrNameNotFound is returned by GetName when no name has been stored yet. The
// API layer translates it into an HTTP 404 response.
var ErrNameNotFound = errors.New("name not found")

// GetName is an example that shows you how to query data.
func (db *appdbimpl) GetName() (string, error) {
	var name string
	err := db.c.QueryRow("SELECT name FROM example_table WHERE id=1").Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNameNotFound
	}
	if err != nil {
		return "", err
	}
	return name, nil
}
