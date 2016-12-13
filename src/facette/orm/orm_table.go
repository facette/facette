package orm

import (
	"fmt"
	"strings"
)

func (db *DB) createTable(model *model) error {
	// Generate columns definitions
	defs := []string{}
	pkeys := []string{}

	for _, field := range model.fields {
		if field.hasMany {
			continue
		}

		defs = append(defs, field.columnDef())

		if field.primaryKey {
			pkeys = append(pkeys, db.driver.QuoteName(field.name))
		}
	}

	if len(pkeys) > 0 {
		defs = append(defs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pkeys, ", ")))
	}

	// Create new table
	_, err := db.Raw(fmt.Sprintf(
		"CREATE TABLE %s (%s)",
		db.driver.QuoteName(model.name),
		strings.Join(defs, ", "),
	)).Result()

	return err
}

func (db *DB) createColumn(model *model, field *field) error {
	// Create new table column
	_, err := db.Raw(fmt.Sprintf(
		"ALTER TABLE %s ADD COLUMN %s",
		db.driver.QuoteName(model.name),
		field.columnDef(),
	)).Result()

	return err
}

func (db *DB) dropTable() *DB {
	table := db.query.from
	if db.model != nil {
		table = db.model.name
	}

	_, err := db.Raw("DROP TABLE " + db.driver.QuoteName(table)).Result()
	return db.setError(err)
}
