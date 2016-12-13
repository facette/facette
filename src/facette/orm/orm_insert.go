package orm

import (
	"fmt"
	"reflect"
	"strings"
)

// Insert inserts a record into the database.
func (db *DB) Insert(value interface{}) *DB {
	dbClone := db.From(value)
	if dbClone.Error() != nil {
		return dbClone
	}

	dbClone.insert()
	return dbClone
}

func (db *DB) insert() *DB {
	var err error

	columns := []string{}
	binds := []string{}
	args := []interface{}{}

	// Get a non-pointer reflect value
	rv := reflect.ValueOf(db.value)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for _, field := range db.model.fields {
		f := rv.FieldByName(field.fieldName)
		for f.Kind() == reflect.Ptr {
			f = f.Elem()
		}

		if field.foreignField != nil {
			// Skip one-to-many association
			if f.Kind() == reflect.Slice {
				continue
			}

			// Handle one-to-one association
			f = rv.FieldByName(field.fieldName + field.foreignKey)
		}

		// Skip invalid or default values
		if !f.IsValid() || IsDefault(f) {
			continue
		}

		columns = append(columns, db.driver.QuoteName(field.name))
		args = append(args, db.driver.adaptValue(f).Interface())
		binds = append(binds, db.driver.bindVar(len(args)))
	}

	// Insert new value into database
	q := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		db.driver.QuoteName(db.model.name),
		strings.Join(columns, ", "),
		strings.Join(binds, ", "),
	)

	primaryField := db.model.primaryField()

	if !primaryField.autoIncrement {
		if _, err := db.Raw(q, args...).Result(); err != nil {
			return db.setError(err)
		}

		goto skipReturnedID
	}

	// Check for returned primary field
	if suffix := db.driver.returningSuffix(primaryField.name); suffix != "" {
		if err := db.Raw(q+suffix, args...).Row().
			Scan(rv.FieldByName(primaryField.fieldName).Addr().Interface()); err != nil {

			db.setError(err)
		}
	} else {
		if result, err := db.Raw(q, args...).Result(); err != nil {
			return db.setError(err)
		} else if result != nil {
			id, _ := result.LastInsertId()

			f := rv.FieldByName(primaryField.fieldName)
			f.Set(reflect.ValueOf(id).Convert(f.Type()))
		}
	}

skipReturnedID:
	// Handle one-to-many associations
	for _, field := range db.model.fields {
		f := rv.FieldByName(field.fieldName)
		for f.Kind() == reflect.Ptr {
			f = f.Elem()
		}

		if field.foreignField != nil && f.Kind() == reflect.Slice {
			for i := 0; i < f.Len(); i++ {
				rv := f.Index(i)
				for rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}

				subID := rv.FieldByName(field.foreignField.model.fields[1].fieldName).Interface()

				// Save association
				db.result, err = db.Raw(fmt.Sprintf(
					"INSERT INTO %s VALUES (%s, %s)",
					db.driver.QuoteName(field.foreignField.model.name),
					db.driver.bindVar(1),
					db.driver.bindVar(2),
				), primaryField.value.Interface(), subID).Result()
				if err != nil {
					db.setError(err)
				}
			}
		}
	}

	return db
}
