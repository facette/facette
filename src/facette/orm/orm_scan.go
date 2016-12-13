package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/facette/sliceutil"
)

// Scan maps a given row into a struct.
func (db *DB) Scan(rows *sql.Rows, value interface{}) *DB {
	var (
		model  *model
		rvOrig reflect.Value
		err    error
	)

	// Set model given the value if not already known
	if db.model != nil {
		model = db.model
	} else {
		model, err = newModel(value, db)
		if err != nil {
			db.setError(err)
			return db
		}
	}

	// Get columns list
	columns, err := rows.Columns()
	if err != nil {
		return db.setError(err)
	}

	// Scan rows for values
	values := []interface{}{}
	for _, field := range model.fields {
		if !field.hasMany && sliceutil.Has(columns, field.name) {
			values = append(values, new(interface{}))
		}
	}

	valuesMap := map[int]interface{}{}
	for idx, field := range model.fields {
		if field.hasMany {
			continue
		}

		if valueIdx := sliceutil.IndexOf(columns, field.name); valueIdx != -1 {
			valuesMap[idx] = values[valueIdx]
		}
	}

	if err = rows.Scan(values...); err != nil {
		return db.setError(err)
	}

	// Parse returned values
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice {
		rvOrig = rv

		rv = reflect.Indirect(reflect.New(rv.Type().Elem()))
		for rv.Kind() == reflect.Ptr {
			if rv.Elem().IsValid() {
				rv = rv.Elem()
			} else {
				rv = reflect.Indirect(reflect.New(rv.Type().Elem()))
			}
		}
	} else if rv.Kind() != reflect.Struct {
		return db.setError(ErrInvalidScanValue)
	}

	for idx, field := range model.fields {
		var value reflect.Value

		if field.foreignField != nil && field.hasMany {
			value = reflect.ValueOf(valuesMap[model.fieldIndex(field.foreignField.foreignKey)]).Elem()
		} else if len(columns) > 0 && !sliceutil.Has(columns, field.name) {
			// Skip fields not present in columns
			continue
		} else {
			value = reflect.ValueOf(valuesMap[idx]).Elem()
		}

		if value.Kind() == reflect.Interface {
			value = value.Elem()
		}

		if !value.IsValid() {
			continue
		}

		f := rv.FieldByName(field.fieldName)
		if field.foreignField != nil {
			var fv reflect.Value

			if f.Kind() == reflect.Ptr && !f.Elem().IsValid() {
				fv = reflect.Indirect(reflect.New(f.Type().Elem()))
			} else {
				fv = reflect.Indirect(reflect.New(f.Type()))
			}

			// Convert value to foreign field type
			conv := reflect.Indirect(reflect.New(field.foreignField.typ))
			db.convert(conv, value)

			if field.hasMany {
				// Fetch associated entries identifiers
				rows, err := db.clone().
					Select(field.foreignField.model.fields[1].name).
					From(field.foreignField.model.name).
					Where(fmt.Sprintf(
						"%s = %s",
						db.driver.QuoteName(field.foreignField.name),
						db.driver.bindVar(1),
					), conv.Interface()).
					Rows()
				if err != nil {
					return db.setError(err)
				}
				defer rows.Close()

				binds := []string{}
				args := []interface{}{}
				for rows.Next() {
					id := new(interface{})
					rows.Scan(id)

					// Convert scanned value
					value := reflect.Indirect(reflect.New(field.foreignField.model.fields[1].typ))
					db.convert(value, reflect.ValueOf(*id))

					binds = append(binds, db.driver.bindVar(len(binds)+1))
					args = append(args, value.Interface())
				}

				// Get associated entries
				if len(args) > 0 {
					err = db.clone().
						Where(fmt.Sprintf(
							"%s IN (%s)",
							db.driver.QuoteName(field.foreignField.model.fields[1].foreignField.name),
							strings.Join(binds, ", "),
						), args...).
						Find(fv.Addr().Interface()).
						Error()
					if err != nil {
						return db.setError(err)
					}
				}
			} else {
				err := db.clone().
					Where(fmt.Sprintf(
						"%s = %s",
						db.driver.QuoteName(field.foreignField.name),
						db.driver.bindVar(1),
					), conv.Interface()).
					Find(fv.Addr().Interface()).
					Error()
				if err != nil {
					return db.setError(err)
				}

				// Set reference field
				ref := rv.FieldByName(field.fieldName + field.foreignKey)
				if ref.Kind() == reflect.Ptr {
					ref.Set(conv.Addr())
				} else {
					ref.Set(conv)
				}
			}

			if f.Kind() == reflect.Ptr {
				f.Set(fv.Addr())
			} else {
				f.Set(fv)
			}
		} else if err := db.convert(f, value); err != nil {
			db.setError(fmt.Errorf("unable to convert %q column from %q to %q: %s",
				field.name, value.Type(), f.Type(), err))
			break
		}
	}

	// If kind was slice, append item to it
	if rvOrig.IsValid() {
		if rvOrig.Type().Elem().Kind() == reflect.Ptr {
			rvOrig.Set(reflect.Append(rvOrig, rv.Addr()))
		} else {
			rvOrig.Set(reflect.Append(rvOrig, rv))
		}
	}

	return db
}

func (db *DB) convert(field, value reflect.Value) error {
	if err := db.driver.scanValue(field, value); err == nil {
		return nil
	} else if ft := field.Type(); ft.Kind() == reflect.Ptr && value.Type().ConvertibleTo(ft.Elem()) {
		v := reflect.New(ft.Elem())
		v.Elem().Set(value.Convert(ft.Elem()))
		field.Set(v)
		return nil
	} else if value.Type().ConvertibleTo(field.Type()) {
		field.Set(value.Convert(field.Type()))
		return nil
	} else if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
		if err := scanner.Scan(value.Interface()); err != nil {
			return err
		}

		return nil
	}

	return ErrNotConvertible
}
