package mapper

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// Map represents an instance of keys mappings.
//
// All the function returning values associated with keys will return a fallback value if the key is not present in the
// map or if the value can't be converted to the requested type.
type Map map[string]interface{}

// Set sets the mapping key k to the value v.
func (m Map) Set(k string, v interface{}) {
	m[k] = v
}

// Has returns true if key is present in the mapping, false otherwise.
func (m Map) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Value marshals the keys mapping for compatibility with SQL drivers.
func (m Map) Value() (driver.Value, error) {
	data, err := json.Marshal(m)
	return data, err
}

// Scan unmarshals the keys mapping retrieved from SQL drivers.
func (m *Map) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), m)
}

// GetBool returns the boolean value associated with a key.
func (m Map) GetBool(key string, fallback bool) (bool, error) {
	val, err := m.getKey(reflect.Bool, key, fallback)
	if err != nil {
		return fallback, err
	}

	return val.(bool), err
}

// GetFloat returns the floating-point number value associated with a key.
func (m Map) GetFloat(key string, fallback float64) (float64, error) {
	val, err := m.getKey(reflect.Float64, key, fallback)
	if err != nil {
		return fallback, err
	}

	return val.(float64), err
}

// GetInt64 returns the 64-bit integer value associated with a key.
func (m Map) GetInt64(key string, fallback int64) (int64, error) {
	val, err := m.getKey(reflect.Int64, key, fallback)
	if err != nil {
		return fallback, err
	}

	return val.(int64), err
}

// GetInt returns the integer value associated with a key.
func (m Map) GetInt(key string, fallback int) (int, error) {
	val, err := m.getKey(reflect.Int, key, fallback)
	if err != nil {
		return fallback, err
	}

	return val.(int), err
}

// GetMap returns the map value associated with a key.
func (m Map) GetMap(key string, fallback Map) (Map, error) {
	val, err := m.getKey(reflect.Map, key, fallback)
	if err != nil {
		return fallback, err
	}

	result := make(Map)
	switch val.(type) {
	case Map:
		for k, v := range val.(Map) {
			result[k] = v
		}

	default:
		for k, v := range val.(map[string]interface{}) {
			result[fmt.Sprintf("%v", k)] = v
		}
	}

	return result, nil
}

// GetString returns the string value associated with a key.
func (m Map) GetString(key, fallback string) (string, error) {
	val, err := m.getKey(reflect.String, key, fallback)
	if err != nil {
		return fallback, err
	}

	return val.(string), err
}

// GetStringSlice returns the string value associated with a key.
func (m Map) GetStringSlice(key string, fallback []string) ([]string, error) {
	out := []string{}

	val, err := m.getKey(reflect.Slice, key, fallback)

	rv := reflect.ValueOf(val)
	count := rv.Len()
	for i := 0; i < count; i++ {
		out = append(out, fmt.Sprintf("%v", rv.Index(i).Interface()))
	}

	return out, err
}

// Merge merges the source map entries into the receiver map. If replace parameter is true, existing entries in the
// receiver map will be replaced with entries from the source map.
func (m *Map) Merge(source Map, replace bool) {
	if *m == nil {
		*m = make(Map)
	}

	for sourceKey, sourceValue := range source {
		if _, ok := map[string]interface{}(*m)[sourceKey]; !ok || replace {
			map[string]interface{}(*m)[sourceKey] = sourceValue
		}
	}
}

func (m Map) getKey(k reflect.Kind, key string, fallback interface{}) (interface{}, error) {
	if m == nil {
		return fallback, nil
	}

	val, ok := m[key]
	if !ok {
		return fallback, nil
	}

	vk := reflect.ValueOf(val).Kind()
	if vk != k {
		return nil, ErrInvalidType
	}

	return val, nil
}

// Flatten returns a "flattened" version of the map.
// * all nested elements are moved at the top level, with sub-level keys prefixed with the original key name using the glue string
// * any value whose type is different than maps or slice is returned as string representation (fmt "%v" verb)
// * slice elements are flattened, with their numeric index value as suffix
// * any pointer value is dereferenced
func (m Map) Flatten(glue string) map[string]string {
	type stackItem struct {
		prefix string
		value  interface{}
	}

	var cur stackItem

	if glue == "" {
		glue = "."
	}

	result := map[string]string{}

	stack := []stackItem{{"", m}}

	for len(stack) > 0 {
		cur, stack = stack[0], stack[1:]

		rv := reflect.ValueOf(cur.value)
		for rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Map {
			result[cur.prefix] = fmt.Sprintf("%v", rv.Interface())
			continue
		}

		for _, k := range rv.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if cur.prefix != "" {
				name = cur.prefix + glue + name
			}

			iv := rv.MapIndex(k)
			for iv.Kind() == reflect.Ptr || iv.Kind() == reflect.Interface {
				iv = iv.Elem()
			}

			if iv.Kind() == reflect.Map {
				stack = append(stack, stackItem{name, iv.Interface()})
			} else if iv.Kind() == reflect.Slice {
				n := iv.Len()
				for i := 0; i < n; i++ {
					stack = append(stack, stackItem{fmt.Sprintf("%s%s%d", name, glue, i), iv.Index(i).Interface()})
				}
			} else {
				result[name] = fmt.Sprintf("%v", iv.Interface())
			}
		}
	}

	return result
}
