package postgres

import (
	"database/sql"
	"errors"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
)

// GetName is an example that shows you how to query data.
func (db *postgresDB) GetName() (string, error) {
	var name string
	err := db.c.QueryRow("SELECT name FROM example_table WHERE id=1").Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", database.ErrNameNotFound
	}
	if err != nil {
		return "", err
	}
	return name, nil
}
