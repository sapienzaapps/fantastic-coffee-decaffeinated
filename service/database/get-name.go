package database

// GetName is an example that shows you how to query data
func (db *appdbimpl) GetName() (string, error) {
	var name string
	err := db.c.QueryRow("SELECT name FROM example_table WHERE id=1").Scan(&name)
	return name, err
}
