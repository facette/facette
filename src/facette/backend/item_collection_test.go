package backend

import (
	"reflect"
	"testing"

	"github.com/facette/maputil"
)

func testCollectionNew() []*Collection {
	parent := &Collection{
		Item: Item{
			Name: "item1",
		},
		Entries: []*CollectionEntry{},
	}

	tmpl := &Collection{
		Item: Item{
			Name: "item2",
		},
		Entries: []*CollectionEntry{},
		Options: map[string]interface{}{
			"title": "{{ .source }}",
		},
		Template: true,
	}

	return []*Collection{
		parent,

		tmpl,

		&Collection{
			Item: Item{
				Name: "item3",
			},
			Entries: []*CollectionEntry{},
			LinkID:  &tmpl.ID,
			Attributes: map[string]interface{}{
				"source": "source1",
			},
			ParentID: &parent.ID,
		},
	}
}

func testCollectionCreate(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemCreate(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionCreateInvalid(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemCreateInvalid(b, &Collection{}, testInterfaceToSlice(testCollections), t)

	alias := "invalid!"
	if err := b.Storage().Save(&Collection{Item: Item{Name: "name"}, Alias: &alias}); err != ErrInvalidAlias {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidAlias, err)
		t.Fail()
	}
}

func testCollectionGet(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemGet(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionGetUnknown(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemGetUnknown(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionUpdate(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemUpdate(b, &Collection{}, testInterfaceToSlice(testCollections), t)

	val := ""
	testCollections[0].Alias = &val
	testCollections[0].LinkID = &val
	testCollections[0].ParentID = &val

	if err := b.Storage().Save(testCollections[0]); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}

	collection := &Collection{}
	if err := b.Storage().Get("name", "item1", collection, true); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if !reflect.DeepEqual(collection, testCollections[0]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testCollections[0], collection)
		t.Fail()
	} else if testCollections[0].Alias != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", testCollections[0].Alias)
		t.Fail()
	} else if testCollections[0].LinkID != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", testCollections[0].LinkID)
		t.Fail()
	}
}

func testCollectionCount(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemCount(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionList(b *Backend, testCollections []*Collection, testGraphs []*Graph, t *testing.T) {
	for _, graph := range testGraphs {
		if err := b.Storage().Save(graph); err != nil {
			t.Logf("\nExpected <nil>\nbut got  %#v", err)
			t.Fail()
		}
	}

	testCollections[1].Entries = append(testCollections[1].Entries, &CollectionEntry{GraphID: testGraphs[1].ID})

	testItemList(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionDelete(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemDelete(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionDeleteAll(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemDeleteAll(b, &Collection{}, testInterfaceToSlice(testCollections), t)

	if err := b.Storage().Delete(&Graph{}); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}
}

func testCollectionResolve(b *Backend, testCollections []*Collection, t *testing.T) {
	if err := testCollections[2].Resolve(nil); err != ErrUnresolvableItem {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrUnresolvableItem, err)
		t.Fail()
	}

	testCollections[1].SetBackend(b)
	testCollections[2].SetBackend(b)

	if err := testCollections[2].Resolve(nil); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if !reflect.DeepEqual(testCollections[2].Link, testCollections[1]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testCollections[1], testCollections[2].Link)
		t.Fail()
	}
}

func testCollectionExpand(b *Backend, testCollections []*Collection, t *testing.T) {
	collection := testCollections[2].Clone()
	if err := collection.Expand(maputil.Map{"source": "other1"}); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if collection.Options["title"] != "other1" {
		t.Logf("\nExpected %#v\nbut got  %#v", "other1", collection.Options["title"])
		t.Fail()
	}

	collection = testCollections[2]
	if err := collection.Expand(nil); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if collection.Entries[0].Options["title"] != testCollections[2].Attributes["source"] {
		t.Logf("\nExpected %#v\nbut got  %#v", testCollections[2].Attributes["source"],
			collection.Entries[0].Options["title"])
		t.Fail()
	} else if collection.Options["title"] != testCollections[2].Attributes["source"] {
		t.Logf("\nExpected %#v\nbut got  %#v", testCollections[2].Attributes["source"], collection.Options["title"])
		t.Fail()
	}

}

func testCollectionTree(b *Backend, testCollections []*Collection, t *testing.T) {
	expected := &CollectionTree{
		{
			ID:    testCollections[0].ID,
			Label: testCollections[0].Name,
			Children: &CollectionTree{
				{
					ID:       testCollections[2].ID,
					Label:    testCollections[2].Attributes["source"].(string),
					Parent:   testCollections[0].ID,
					Children: &CollectionTree{},
				},
			},
		},
	}

	tree, err := b.NewCollectionTree("")
	if err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if len(*tree) != len(*expected) {
		t.Logf("\nExpected %#v\nbut got  %#v", len(*expected), len(*tree))
		t.Fail()
	} else if !reflect.DeepEqual(tree, expected) {
		t.Logf("\nExpected %#v\nbut got  %#v", expected, tree)
		t.Fail()
	}
}
