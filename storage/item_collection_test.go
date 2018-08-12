package storage

import (
	"testing"

	"facette.io/maputil"
	"github.com/stretchr/testify/assert"
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

func testCollectionCreate(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemCreate(s, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionCreateInvalid(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemCreateInvalid(s, &Collection{}, testInterfaceToSlice(testCollections), t)
	alias := "invalid!"
	assert.Equal(t, ErrInvalidAlias, s.SQL().Save(&Collection{Item: Item{Name: "name"}, Alias: &alias}))
}

func testCollectionGet(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemGet(s, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionGetUnknown(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemGetUnknown(s, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionUpdate(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemUpdate(s, &Collection{}, testInterfaceToSlice(testCollections), t)

	val := ""
	testCollections[0].Alias = &val
	testCollections[0].LinkID = &val
	testCollections[0].ParentID = &val

	assert.Nil(t, s.SQL().Save(testCollections[0]))

	collection := &Collection{}
	assert.Nil(t, s.SQL().Get("name", "item1", collection, true))
	assert.Nil(t, collection.Alias)
	assert.Nil(t, collection.LinkID)
	assert.Nil(t, collection.ParentID)
	assert.Equal(t, testCollections[0], collection)
}

func testCollectionCount(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemCount(s, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionList(s *Storage, testCollections []*Collection, testGraphs []*Graph, t *testing.T) {
	for _, graph := range testGraphs {
		assert.Nil(t, s.SQL().Save(graph))
	}
	testCollections[1].Entries = append(testCollections[1].Entries, &CollectionEntry{GraphID: testGraphs[1].ID})
	testItemList(s, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionDelete(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemDelete(s, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionDeleteAll(s *Storage, testCollections []*Collection, t *testing.T) {
	testItemDeleteAll(s, &Collection{}, testInterfaceToSlice(testCollections), t)
	assert.Nil(t, s.SQL().Delete(&Graph{}))
}

func testCollectionResolve(s *Storage, testCollections []*Collection, t *testing.T) {
	assert.Equal(t, ErrUnresolvableItem, testCollections[2].Resolve(nil))
	testCollections[1].SetStorage(s)
	testCollections[2].SetStorage(s)
	assert.Nil(t, testCollections[2].Resolve(nil))
	assert.Equal(t, testCollections[1], testCollections[2].Link)
}

func testCollectionExpand(s *Storage, testCollections []*Collection, t *testing.T) {
	collection := testCollections[2].Clone()
	assert.Nil(t, collection.Expand(maputil.Map{"source": "other1"}))
	assert.Equal(t, "other1", collection.Options["title"])

	collection = testCollections[2]
	assert.Nil(t, collection.Expand(nil))
	assert.Equal(t, testCollections[2].Attributes["source"], collection.Entries[0].Options["title"])
	assert.Equal(t, testCollections[2].Attributes["source"], collection.Options["title"])
}

func testCollectionTree(s *Storage, testCollections []*Collection, t *testing.T) {
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

	tree, err := s.NewCollectionTree("")
	assert.Nil(t, err)
	assert.Len(t, *tree, len(*expected))
	assert.Equal(t, expected, tree)
}
