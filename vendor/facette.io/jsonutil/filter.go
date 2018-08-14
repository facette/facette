package jsonutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Filter filters an interface given its type and JSON field paths.
func Filter(v interface{}, fields []string) interface{} {
	rv := reflect.ValueOf(v)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Map:
		return FilterMap(v, fields)

	case reflect.Slice:
		return FilterSlice(v, fields)

	case reflect.Struct:
		return FilterStruct(v, fields)
	}

	return nil
}

// FilterMap filters a map given key paths.
func FilterMap(v interface{}, fields []string) map[string]interface{} {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Map {
		return nil
	}

	result := map[string]interface{}{}

	for _, k := range rv.MapKeys() {
		name := fmt.Sprintf("%v", k.Interface())
		if filterMatch(name, fields) {
			iv := rv.MapIndex(k)
			if iv.Elem().Kind() == reflect.Map {
				result[name] = FilterMap(iv.Interface(), filterFields(name, fields))
			} else {
				result[name] = iv.Interface()
			}
		}
	}

	return result
}

// FilterSlice filters a slice of struct given JSON field paths.
func FilterSlice(v interface{}, fields []string) []map[string]interface{} {
	rv := reflect.ValueOf(v)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := []map[string]interface{}{}

	n := rv.Len()
	for i := 0; i < n; i++ {
		result = append(result, FilterStruct(rv.Index(i).Interface(), fields))
	}

	return result
}

// FilterStruct filters a struct given a list of JSON field paths.
// The return value is a map containing the struct fields matched on the `json` field tag.
//
// It is possible to filter nested structures using a dot as level separator, e.g. "item.sub_item".
//
// Fields are ignored when:
// 	* not tagged with `json`
// 	* tagged with `json:"-"`
// 	* tagged with `omitempty` and set to their default value
//
// Example:
// 	type Service struct {
//		ID        string    `json:"id"`
//		Name      string    `json:"name"`
//		Hostgroup Hostgroup `json:"hostgroup"`
//	}
//
//	type Hostgroup struct {
//		ID    string `json:"id"`
//		Name  string `json:"name"`
//		Hosts []Host `json:"hosts"`
//	}
//
//	type Host struct {
//		ID   string `json:"id"`
//		Name string `json:"name"`
//		Addr string `json:"addr"`
//		Port int    `json:"port"`
//	}
//
//	func main() {
//		service := Service{
//			ID:   "CCDE8419-74BB-4A8C-8AE4-8E2A17B3C3DD",
//			Name: "service1",
//			Hostgroup: Hostgroup{
//				ID:   "48F2F97F-EB85-418B-A05D-63C0A3914AA4",
//				Name: "hostgroup1",
//				Hosts: []Host{
//					Host{
//						ID:   "E0011483-5E4A-4A2F-96AA-DADA236E3BD6",
//						Name: "host1",
//						Addr: "1.2.3.4",
//						Port: 9999,
//					},
//					Host{
//						ID:   "C9AFE8C0-1017-4025-B9A3-CBB9901E1F44",
//						Name: "host2",
//						Addr: "5.6.7.8",
//						Port: 9999,
//					},
//					Host{
//						ID:   "E2F99B14-8839-4321-90AF-3877486A68D9",
//						Name: "host3",
//						Addr: "9.10.11.12",
//						Port: 9999,
//					},
//				},
//			},
//		}
//
//		j, err := json.MarshalIndent(jsonutil.FilterStruct(service, []string{
//			"name",
//			"hostgroup.name",
//			"hostgroup.hosts.name",
//		}), "", "  ")
//
//		fmt.Printf("%s\n", j)
//
//	}
//
// ...will output the following JSON data:
//
//	{
//	"hostgroup": {
//		"hosts": [
//		{
//			"name": "host1"
//		},
//		{
//			"name": "host2"
//		},
//		{
//			"name": "host3"
//		}
//		],
//		"name": "hostgroup1"
//	},
//	"name": "service1"
//	}
//
func FilterStruct(v interface{}, fields []string) map[string]interface{} {
	var current reflect.Value

	rv := reflect.ValueOf(v)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]interface{})

	stack := []reflect.Value{rv}

	for len(stack) > 0 {
		current, stack = stack[0], stack[1:]

		n := current.NumField()
		for i := 0; i < n; i++ {
			ft := current.Type().Field(i)
			f := current.Field(i)

			// Handle nested structures
			if ft.Anonymous {
				stack = append(stack, f)
				continue
			}

			// Get field tag and check if it needs to be skipped
			tag := ft.Tag.Get("json")
			if tag == "-" || filterSkip(tag, f) || unicode.IsLower(rune(ft.Name[0])) {
				continue
			}

			// Get field name
			fname := filterBaseField(strings.SplitN(tag, ",", 2)[0])

			if _, ok := f.Interface().(json.Marshaler); !ok && f.Kind() == reflect.Struct {
				// Handle sub struct filtering
				if !filterMatch(fname, fields) {
					continue
				} else if smap := FilterStruct(
					f.Interface(),
					filterFields(fname, fields),
				); len(smap) > 0 {
					filterSetEntry(result, fname, smap)
				}
			} else if f.Kind() == reflect.Slice {
				// Handle slice filtering
				slice := []map[string]interface{}{}

				if f.Type().Elem().Kind() == reflect.Struct {
					n := f.Len()
					for i := 0; i < n; i++ {
						if !filterMatch(fname, fields) {
							continue
						} else if smap := FilterStruct(
							f.Index(i).Interface(),
							filterFields(fname, fields),
						); len(smap) > 0 {
							slice = append(slice, smap)
						}
					}

					if len(slice) > 0 {
						filterSetEntry(result, fname, slice)
					}
				} else if filterMatch(fname, fields) {
					filterSetEntry(result, fname, f.Interface())
				}
			} else if f.Kind() == reflect.Map {
				subFields := filterFields(fname, fields)
				if len(subFields) == 0 && len(fields) > 0 && !filterMatch(fname, fields) {
					continue
				}

				if smap := FilterMap(f.Interface(), subFields); len(smap) > 0 {
					filterSetEntry(result, fname, smap)
				}
			} else if !filterMatch(fname, fields) {
				// Skip unwanted fields
				continue
			} else {
				// Set item value
				filterSetEntry(result, fname, f.Interface())
			}
		}
	}

	return result
}

func filterMatch(name string, fields []string) bool {
	if len(fields) == 0 {
		return true
	}

	for _, s := range fields {
		if name == filterBaseField(s) {
			return true
		}
	}

	return false
}

func filterFields(prefix string, fields []string) []string {
	result := []string{}
	for _, s := range fields {
		if strings.HasPrefix(s, prefix+".") {
			result = append(result, strings.TrimPrefix(s, prefix+"."))
		}
	}

	return result
}

func filterBaseField(name string) string {
	return strings.SplitN(strings.SplitN(name, ",", 2)[0], ".", 2)[0]
}

func filterSkip(tag string, v reflect.Value) bool {
	parts := strings.Split(tag, ",")
	if len(parts) > 1 {
		for _, part := range parts {
			if part == "omitempty" && reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface()) {
				return true
			}
		}
	}

	return false
}

func filterSetEntry(result map[string]interface{}, key string, value interface{}) {
	if _, ok := result[key]; !ok {
		result[key] = value
	}
}
