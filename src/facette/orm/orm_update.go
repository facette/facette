package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/facette/sliceutil"
)

// Update updates a record from the database.
func (db *DB) Update(value interface{}) *DB {
	dbClone := db.From(value)
	if dbClone.Error() != nil {
		return dbClone
	}

	dbClone.update()
	return dbClone
}

func (db *DB) update() *DB {
	var err error

	sets := []string{}
	args := []interface{}{}
	conditions := []string{}
	conditionsArgs := map[string]interface{}{}

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

		// Skip invalid values
		if !f.IsValid() {
			continue
		}

		// Save id and skip primary key field
		if field.primaryKey {
			conditionsArgs[field.name] = f.Interface()
			continue
		}

		// Handle nullable columns having default values
		if field.nullable && IsDefault(f) {
			args = append(args, nil)
		} else {
			args = append(args, db.driver.adaptValue(f).Interface())
		}

		sets = append(sets, fmt.Sprintf("%s = %s", db.driver.QuoteName(field.name), db.driver.bindVar(len(args))))
	}

	// Append condition arguments to query ones
	for fieldName, arg := range conditionsArgs {
		args = append(args, arg)

		conditions = append(conditions, fmt.Sprintf(
			"%s = %s",
			db.driver.QuoteName(fieldName),
			db.driver.bindVar(len(args)),
		))
	}

	db.result, err = db.Raw(fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		db.driver.QuoteName(db.model.name),
		strings.Join(sets, ", "),
		strings.Join(conditions, " AND "),
	), args...).Result()

	// Handle one-to-many associations
	primaryField := db.model.primaryField()

	for _, field := range db.model.fields {
		f := rv.FieldByName(field.fieldName)
		for f.Kind() == reflect.Ptr {
			f = f.Elem()
		}

		if field.foreignField != nil && f.Kind() == reflect.Slice {
			// Fetch associated entries identifiers
			rows, err := db.
				Select(field.foreignField.model.fields[1].name).
				From(field.foreignField.model.name).
				Where(fmt.Sprintf("%s = %s", field.foreignField.name, db.driver.bindVar(1)),
					primaryField.value.Interface()).
				Rows()
			if err != nil {
				return db.setError(err)
			}
			defer rows.Close()

			current := reflect.Indirect(reflect.New(reflect.SliceOf(primaryField.typ)))

			existing := reflect.Indirect(reflect.New(reflect.SliceOf(primaryField.typ)))
			for rows.Next() {
				id := new(interface{})
				rows.Scan(id)

				// Convert identifier to the proper field type
				value := reflect.Indirect(reflect.New(primaryField.typ))
				db.convert(value, reflect.ValueOf(*id))

				existing = reflect.Append(existing, value)
			}

			for i := 0; i < f.Len(); i++ {
				rv := f.Index(i)
				for rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}

				// Save associated item
				if err := db.Save(rv.Addr().Interface()).Error(); err != nil {
					return db.setError(err)
				}

				subID := rv.FieldByName(field.foreignField.model.fields[1].fieldName).
					Convert(primaryField.typ).Interface()

				if sliceutil.Has(existing.Interface(), subID) {
					current = reflect.Append(current, reflect.ValueOf(subID))
					continue
				}

				// Save association
				_, err := db.Raw(fmt.Sprintf(
					"INSERT INTO %s VALUES (%s, %s)",
					db.driver.QuoteName(field.foreignField.model.name),
					db.driver.bindVar(1),
					db.driver.bindVar(2),
				), primaryField.value.Interface(), subID).Result()
				if err != nil {
					db.setError(err)
				}
			}

			// Remove no longer associated references
			for i := 0; i < existing.Len(); i++ {
				id := existing.Index(i).Interface()
				if sliceutil.Has(current.Interface(), id) {
					continue
				}

				_, err := db.Raw(fmt.Sprintf(
					"DELETE FROM %s WHERE %s = %s",
					db.driver.QuoteName(field.foreignField.model.name),
					db.driver.QuoteName(field.foreignField.model.fields[1].name),
					db.driver.bindVar(1),
				), id).Result()
				if err != nil {
					db.setError(err)
				}
			}
		}
	}

	return db.setError(err)
}
