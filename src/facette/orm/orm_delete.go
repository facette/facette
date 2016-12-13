package orm

import (
	"fmt"
	"reflect"
	"strings"
)

func (db *DB) delete() *DB {
	var err error

	pkeys := []string{}
	args := []interface{}{}

	// Get a non-pointer reflect value
	rv := reflect.ValueOf(db.value)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for _, field := range db.model.fields {
		// Skip non primary key fields
		if !field.primaryKey {
			continue
		}

		f := rv.FieldByName(field.fieldName)
		for f.Kind() == reflect.Ptr {
			f = f.Elem()
		}

		args = append(args, db.driver.adaptValue(f).Interface())
		pkeys = append(pkeys, fmt.Sprintf("%s = %s", db.driver.QuoteName(field.name), db.driver.bindVar(len(args))))
	}

	if len(pkeys) == 0 {
		return db.setError(ErrMissingPrimaryKey)
	}

	db.result, err = db.Raw(fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		db.driver.QuoteName(db.model.name),
		strings.Join(pkeys, " AND "),
	), args...).Result()

	return db.setError(err)
}
