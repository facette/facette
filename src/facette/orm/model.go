package orm

import (
	"fmt"
	"reflect"
	"strings"
)

type model struct {
	name   string
	fields []*field
	db     *DB
}

func (m *model) primaryField() *field {
	for _, field := range m.fields {
		if field.primaryKey {
			return field
		}
	}
	return nil
}

func (m *model) fieldByName(name string) *field {
	if idx := m.fieldIndex(name); idx != -1 {
		return m.fields[idx]
	}

	return nil
}

func (m *model) fieldIndex(name string) int {
	for idx, field := range m.fields {
		if field.fieldName == name {
			return idx
		}
	}

	return -1
}

func (m *model) mapSettings(field *field, tag string) {
	// Parse settings form struct tag
	settings := map[string]string{}
	for _, entry := range strings.Split(tag, ";") {
		parts := strings.SplitN(strings.Trim(entry, ""), ":", 2)
		if len(parts) == 1 {
			settings[parts[0]] = ""
		} else {
			settings[parts[0]] = parts[1]
		}
	}

	// Map settings to field
	for key, value := range settings {
		switch key {
		case "-":
			field.ignore = true

		case "column":
			field.name = value

		case "type":
			field.sqlType = value

		case "not_null", "unique":
			if key == "not_null" {
				field.nullable = false
			}

			field.properties = append(field.properties, strings.ToUpper(strings.Replace(key, "_", " ", -1)))

		case "primary_key":
			field.primaryKey = true

		case "foreign_key":
			field.foreignKey = value

		case "index":
			field.indexes = []string{}
			for _, part := range strings.Split(value, ",") {
				field.indexes = append(field.indexes, strings.Trim(part, " "))
			}

		case "auto_increment":
			field.autoIncrement = true

		case "default":
			if value == "now()" {
				value = m.db.driver.NowCall()
			} else if value == "true" || value == "false" {
				value = m.db.driver.BooleanValue(value)
			}

			field.properties = append(field.properties, "DEFAULT "+value)
		}
	}
}

func newModel(value interface{}, db *DB) (*model, error) {
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Ptr {
		if rv.Elem().IsValid() {
			rv = rv.Elem()
		} else {
			rv = reflect.Indirect(reflect.New(rv.Type().Elem()))
		}
	}

	// Use element type if value is a slice
	if rv.Kind() == reflect.Slice {
		if rv.Len() == 0 {
			rv = reflect.Indirect(reflect.New(rv.Type().Elem()))
		} else {
			rv = rv.Index(0)
		}

		return newModel(rv.Interface(), db)
	}

	// Check for known models to avoid parsing
	rt := rv.Type()

	if _, ok := db.models[rt]; !ok {
		if err := parseStruct(rt, value, db); err != nil {
			return nil, err
		}
	}

	// Clone the model and apply current values
	modelSrc := db.models[rt]

	model := &model{
		name:   modelSrc.name,
		fields: make([]*field, len(modelSrc.fields)),
		db:     modelSrc.db,
	}

	for idx := range model.fields {
		model.fields[idx] = &field{}
		*model.fields[idx] = *modelSrc.fields[idx]

		model.fields[idx].value = rv.FieldByName(model.fields[idx].fieldName)
		model.fields[idx].model = model
	}

	return model, nil
}

type field struct {
	name          string
	fieldName     string
	sqlType       string
	ignore        bool
	scanner       bool
	properties    []string
	primaryKey    bool
	foreignKey    string
	foreignField  *field
	indexes       []string
	nullable      bool
	autoIncrement bool
	hasMany       bool
	value         reflect.Value
	typ           reflect.Type
	model         *model
}

func (f *field) columnDef() string {
	// Generate column definition
	def := f.model.db.driver.QuoteName(f.name) + " " + f.sqlType

	props := f.properties
	if f.foreignField != nil {
		props = append(props, fmt.Sprintf(
			"REFERENCES %s (%s) ON UPDATE CASCADE ON DELETE CASCADE",
			f.model.db.driver.QuoteName(f.foreignField.model.name),
			f.model.db.driver.QuoteName(f.foreignField.name),
		))
	}

	if len(props) > 0 {
		def += " " + strings.Join(props, " ")
	}

	return def
}
