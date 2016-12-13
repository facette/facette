package orm

import (
	"fmt"
	"strings"
)

func (db *DB) migrate() *DB {
	// Stop if an error has already been triggered
	if db.Error() != nil {
		return db
	}

	// Stop if model already migrated
	if _, ok := db.migrated[db.model.name]; ok {
		return db
	}

	// Migrate one-to-one association models
	for _, field := range db.model.fields {
		if field.foreignField != nil && field.foreignField.model.name != db.model.name {
			if err := db.From(field.value.Interface()).migrate().Error(); err != nil {
				return db.setError(err)
			}
		}
	}

	// Stop migration if an error occured
	if db.Error() != nil {
		return db
	}

	if !db.HasTable(db.model.name) {
		// Create new table from model
		if err := db.createTable(db.model); err != nil {
			return db.setError(err)
		}
	} else {
		// Loop through fields to create missing columns
		for _, field := range db.model.fields {
			if !field.hasMany && !db.HasColumn(db.model.name, field.name) {
				if err := db.createColumn(db.model, field); err != nil {
					return db.setError(err)
				}
			}
		}
	}

	// Create indexes and one-to-many association tables
	indexes := map[string][]string{}

	for _, field := range db.model.fields {
		for _, name := range field.indexes {
			if name == "" {
				name = field.name
			}

			if _, ok := indexes[name]; !ok {
				indexes[name] = []string{}
			}

			indexes[name] = append(indexes[name], field.name)
		}

		if field.foreignField != nil && field.hasMany && !db.HasTable(field.foreignField.model.name) {
			if err := db.createTable(field.foreignField.model); err != nil {
				return db.setError(err)
			}
		}
	}

	// Create indexes
	for name, columns := range indexes {
		indexName := db.model.name + "_" + name

		if !db.driver.hasIndex(db.model.name, indexName) {
			_, err := db.Raw(fmt.Sprintf(
				"CREATE INDEX %s ON %s (%s)",
				indexName,
				db.driver.QuoteName(db.model.name),
				strings.Join(columns, ", "),
			)).Result()

			if err != nil {
				return db.setError(err)
			}
		}
	}

	// Set model migrated
	db.migrated[db.model.name] = struct{}{}

	return db
}
