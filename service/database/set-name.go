package database

// SetName is an example that shows you how to execute insert/update. It stores
// the name in the example table, replacing the previous value if present.
func (db *appdbimpl) SetName(name string) error {
	_, err := db.c.Exec(`
		INSERT INTO example_table (id, name) VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name`, name)
	return err
}
