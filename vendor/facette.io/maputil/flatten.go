package maputil

import (
	"fmt"
	"reflect"
)

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
