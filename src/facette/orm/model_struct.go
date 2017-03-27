package orm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"

	"github.com/jinzhu/inflection"
)

type tabler interface {
	TableName() string
}

type stackItem struct {
	typ   reflect.Type
	index int
}

func parseStruct(rt reflect.Type, value interface{}, db *DB) error {
	var cur stackItem

	if rt.Kind() != reflect.Struct {
		return ErrInvalidStruct
	}

	// Get model name
	name := ""
	if t, ok := value.(tabler); ok {
		name = t.TableName()
	} else {
		name = FormatName(inflection.Plural(rt.Name()))
	}

	// Create new model
	m := &model{
		name:   name,
		fields: []*field{},
		db:     db,
	}

	// Parse value for fields
	stack := []stackItem{{rt, 0}}
	for len(stack) > 0 {
		cur, stack = stack[0], stack[1:]

		if cur.typ.Kind() == reflect.Ptr {
			cur.typ = cur.typ.Elem()
		}

		n := cur.typ.NumField()
		for i := 0; i < n; i++ {
			cf := cur.typ.Field(i)

			ct := cf.Type
			for ct.Kind() == reflect.Ptr {
				ct = ct.Elem()
			}

			// Check for embedded struct
			if cf.Anonymous {
				stack = append(stack, stackItem{ct, i})
				continue
			}

			// Map struct tag settings
			f := &field{
				fieldName:  cf.Name,
				properties: []string{},
				nullable:   true,
				typ:        ct,
				model:      m,
			}

			m.mapSettings(f, cf.Tag.Get("orm"))

			// Skip if field must be ignored
			if f.ignore {
				continue
			}

			// Apply field defaults
			if f.name == "" {
				f.name = FormatName(cf.Name)
			}

			if f.foreignKey != "" {
				if f.foreignOnDelete == "" {
					f.foreignOnDelete = "CASCADE"
				}
				if f.foreignOnUpdate == "" {
					f.foreignOnUpdate = "CASCADE"
				}
			}

			// Check if value implements scanner
			if _, ok := reflect.New(ct).Interface().(sql.Scanner); ok {
				f.scanner = true
			}

			m.fields = append(m.fields[:cur.index], append([]*field{f}, m.fields[cur.index:]...)...)
			cur.index++
		}
	}

	// Handle fallback types and associations
	for _, f := range m.fields {
		if f.scanner {
			if valuer, ok := reflect.New(f.typ).Interface().(driver.Valuer); ok {
				v, _ := valuer.Value()
				sqlType, err := db.driver.typeOf(reflect.ValueOf(v), f.autoIncrement)
				if err != nil {
					return fmt.Errorf("failed to guess type for %q scanner", f.name)
				}

				f.sqlType = sqlType
			}

			continue
		}

		if sqlType, err := db.driver.typeOf(reflect.Zero(f.typ), f.autoIncrement); err != nil {
			// Check for relations
			switch f.typ.Kind() {
			case reflect.Struct:
				var modelSub *model

				if f.typ == rt {
					modelSub = m
				} else {
					if _, ok := db.models[f.typ]; !ok {
						if err := parseStruct(f.typ, reflect.New(f.typ).Interface(), db); err != nil {
							return err
						}
					}

					modelSub = db.models[f.typ]
				}

				if f.foreignField = modelSub.fieldByName(f.foreignKey); f.foreignField != nil {
					f.sqlType, _ = db.driver.typeOf(reflect.Zero(f.foreignField.typ), false)
					continue
				}

			case reflect.Slice:
				f.typ = f.typ.Elem()
				for f.typ.Kind() == reflect.Ptr {
					f.typ = f.typ.Elem()
				}

				if f.typ.Kind() == reflect.Struct {
					var modelSub *model

					if f.typ == rt {
						modelSub = m
					} else {
						if _, ok := db.models[f.typ]; !ok {
							if err := parseStruct(f.typ, reflect.New(f.typ).Interface(), db); err != nil {
								return err
							}
						}

						modelSub = db.models[f.typ]
					}

					if foreignField := modelSub.fieldByName(f.foreignKey); foreignField != nil {
						f.hasMany = true

						modelAssoc := &model{
							name:   m.name + "_" + modelSub.name + "_assoc",
							fields: make([]*field, 2),
							db:     db,
						}

						primaryField := m.primaryField()
						sqlType, _ := db.driver.typeOf(reflect.Zero(primaryField.typ), false)

						modelAssoc.fields[0] = &field{
							name:         m.name + "_" + primaryField.name,
							fieldName:    f.fieldName,
							sqlType:      sqlType,
							properties:   []string{"NOT NULL"},
							primaryKey:   true,
							foreignKey:   primaryField.fieldName,
							foreignField: primaryField,
							model:        modelAssoc,
							typ:          primaryField.typ,
						}

						modelAssoc.fields[1] = &field{
							name:         modelSub.name + "_" + foreignField.name,
							fieldName:    foreignField.fieldName,
							sqlType:      foreignField.sqlType,
							properties:   []string{"NOT NULL"},
							primaryKey:   true,
							foreignKey:   foreignField.fieldName,
							foreignField: foreignField,
							model:        modelAssoc,
							typ:          foreignField.typ,
						}

						f.foreignField = modelAssoc.fields[0]

						continue
					}
				}
			}

			return fmt.Errorf("%s: unable to handle %q column %q value", err, f.name, f.typ)
		} else if f.sqlType == "" {
			f.sqlType = sqlType
		}
	}

	// Register model
	db.models[rt] = m

	return nil
}
