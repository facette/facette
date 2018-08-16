package maputil

import (
	"fmt"
	"reflect"
)

// Map represents an instance of keys mapping.
//
// All the function returning values associated with keys will return a fallback value if the key is not present in the
// map or if the value can't be converted to the requested type.
type Map map[string]interface{}

// Clone returns a clone of the keys mapping instance.
func (m Map) Clone() Map {
	clone := Map{}
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// Set sets the mapping key k to the value v.
func (m Map) Set(k string, v interface{}) {
	m[k] = v
}

// Has returns true if key is present in the mapping, false otherwise.
func (m Map) Has(key string) bool {
	_, ok := m[key]
	return ok
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

// GetInterface returns the interface value associated with a key.
func (m Map) GetInterface(key string, fallback interface{}) (interface{}, error) {
	val, ok := m[key]
	if !ok {
		val = fallback
	}

	return val, nil
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
	var out []string

	val, err := m.getKey(reflect.Slice, key, fallback)

	rv := reflect.ValueOf(val)
	if !rv.IsValid() {
		return nil, err
	}

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
