package backend

import (
	"reflect"
	"testing"
	"time"

	"github.com/facette/sliceutil"
	"github.com/facette/sqlstorage"
)

func testInterfaceToSlice(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := []interface{}{}
	for i, n := 0, rv.Len(); i < n; i++ {
		result = append(result, rv.Index(i).Interface())
	}

	return result
}

func testNewItem(v interface{}) interface{} {
	return reflect.New(reflect.TypeOf(v).Elem()).Interface()
}

func testNewItemSlice(v interface{}) interface{} {
	return reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(v)), 0, 0).Type()).Interface()
}

func testItemCreate(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	item := testItems[0]
	if err := b.Storage().Save(item); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}
}

func testItemCreateInvalid(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	item := testNewItem(refItem)
	reflect.Indirect(reflect.ValueOf(item)).FieldByName("Name").SetString("invalid!")
	if err := b.Storage().Save(item); err != ErrInvalidName {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidName, err)
		t.Fail()
	}

	item = testNewItem(refItem)
	reflect.Indirect(reflect.ValueOf(item)).FieldByName("ID").SetString("invalid!")
	if err := b.Storage().Save(item); err != ErrInvalidID {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidID, err)
		t.Fail()
	}
}

func testItemGet(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	item := testNewItem(refItem)
	rv := reflect.Indirect(reflect.ValueOf(item))
	if err := b.Storage().Get("name", "item1", item); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if id := rv.FieldByName("ID").String(); id == "" {
		t.Logf("\nExpected (non-empty .ID)\nbut got  %#v", id)
		t.Fail()
	} else if created, ok := rv.FieldByName("Created").Interface().(time.Time); !ok || created.IsZero() {
		t.Logf("\nExpected (non-zero .Created)\nbut got  %s", created)
		t.Fail()
	}

	if !reflect.DeepEqual(item, testItems[0]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[0], item)
		t.Fail()
	}
}

func testItemGetUnknown(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	item := testNewItem(refItem)
	if err := b.Storage().Get("name", "unknown1", item); err != sqlstorage.ErrItemNotFound {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}
}

func testItemUpdate(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	desc := "A great item description"

	dv := reflect.Indirect(reflect.ValueOf(testItems[0])).FieldByName("Description")
	dv.Set(reflect.ValueOf(&desc))

	if err := b.Storage().Save(testItems[0]); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}

	item := testNewItem(refItem)
	rv := reflect.Indirect(reflect.ValueOf(item))
	if err := b.Storage().Get("name", "item1", item); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if id := rv.FieldByName("ID").String(); id == "" {
		t.Logf("\nExpected (non-empty .ID)\nbut got  %#v", id)
		t.Fail()
	} else if created, ok := rv.FieldByName("Created").Interface().(time.Time); !ok || created.IsZero() {
		t.Logf("\nExpected (non-zero .Created)\nbut got  %s", created)
		t.Fail()
	} else if modified, ok := rv.FieldByName("Modified").Interface().(time.Time); !ok || modified.IsZero() {
		t.Logf("\nExpected (non-zero .Modified)\nbut got  %s", modified)
		t.Fail()
	}

	if !reflect.DeepEqual(item, testItems[0]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[0], item)
		t.Fail()
	}

	desc = ""
	dv.Set(reflect.ValueOf(&desc))

	if err := b.Storage().Save(testItems[0]); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}

	item = testNewItem(refItem)
	rv = reflect.Indirect(reflect.ValueOf(item))
	if err := b.Storage().Get("name", "item1", item); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if !reflect.DeepEqual(item, testItems[0]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[0], item)
		t.Fail()
	} else if v, ok := dv.Interface().(*string); !ok || v != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", v)
		t.Fail()
	}
}

func testItemCount(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	if count, err := b.Storage().Count(testNewItem(refItem)); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != 3 {
		t.Logf("\nExpected %d\nbut got  %d", 3, count)
		t.Fail()
	}
}

func testItemList(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	for _, item := range testItems {
		if err := b.Storage().Save(item); err != nil {
			t.Logf("\nExpected <nil>\nbut got  %#v", err)
			t.Fail()
		}
	}

	items := testNewItemSlice(refItem)
	if count, err := b.Storage().List(items, nil, []string{"name"}, 0, 0); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != len(testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", len(testItems), count)
		t.Fail()
	} else if a := testInterfaceToSlice(items); !reflect.DeepEqual(a, testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems, a)
		t.Fail()
	}

	items = testNewItemSlice(refItem)
	if count, err := b.Storage().List(items, map[string]interface{}{"name": "item1"},
		[]string{"name"}, 0, 0); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != 1 {
		t.Logf("\nExpected %#v\nbut got  %#v", 1, count)
		t.Fail()
	} else if a := testInterfaceToSlice(items); !reflect.DeepEqual(a, testItems[:1]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[:0], a)
		t.Fail()
	}

	items = testNewItemSlice(refItem)
	if count, err := b.Storage().List(items, nil, []string{"name"}, 0, 2); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != len(testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", len(testItems), count)
		t.Fail()
	} else if a := testInterfaceToSlice(items); !reflect.DeepEqual(a, testItems[:2]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[:2], a)
		t.Fail()
	}

	items = testNewItemSlice(refItem)
	if count, err := b.Storage().List(items, nil, []string{"name"}, 1, 1); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != len(testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", len(testItems), count)
		t.Fail()
	} else if a := testInterfaceToSlice(items); !reflect.DeepEqual(a, testItems[1:2]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[1:2], a)
		t.Fail()
	}

	sliceutil.Reverse(testItems)

	items = testNewItemSlice(refItem)
	if count, err := b.Storage().List(items, nil, []string{"-name"}, 0, 0); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != len(testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", len(testItems), count)
		t.Fail()
	} else if a := testInterfaceToSlice(items); !reflect.DeepEqual(a, testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems, a)
		t.Fail()
	}

	items = testNewItemSlice(refItem)
	if count, err := b.Storage().List(items, nil, []string{"-name"}, 2, 10); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if count != len(testItems) {
		t.Logf("\nExpected %#v\nbut got  %#v", len(testItems), count)
		t.Fail()
	} else if a := testInterfaceToSlice(items); !reflect.DeepEqual(a, testItems[2:]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testItems[2:], a)
		t.Fail()
	}
}

func testItemDelete(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	if err := b.Storage().Delete(testItems[0]); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}

	item := testNewItem(refItem)
	if err := b.Storage().Get("name", "item1", item); err != sqlstorage.ErrItemNotFound {
		t.Logf("\nExpected %#v\nbut got  %#v", sqlstorage.ErrItemNotFound, err)
		t.Fail()
	}
}

func testItemDeleteAll(b *Backend, refItem interface{}, testItems []interface{}, t *testing.T) {
	if err := b.Storage().Delete(testNewItem(refItem)); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}
}
