package backend

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

func testCollectionCreate(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemCreate(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionCreateInvalid(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemCreateInvalid(b, &Collection{}, testInterfaceToSlice(testCollections), t)
	alias := "invalid!"
	assert.Equal(t, ErrInvalidAlias, b.Storage().Save(&Collection{Item: Item{Name: "name"}, Alias: &alias}))
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

	assert.Nil(t, b.Storage().Save(testCollections[0]))

	collection := &Collection{}
	assert.Nil(t, b.Storage().Get("name", "item1", collection, true))
	assert.Nil(t, collection.Alias)
	assert.Nil(t, collection.LinkID)
	assert.Nil(t, collection.ParentID)
	assert.Equal(t, testCollections[0], collection)
}

func testCollectionCount(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemCount(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionList(b *Backend, testCollections []*Collection, testGraphs []*Graph, t *testing.T) {
	for _, graph := range testGraphs {
		assert.Nil(t, b.Storage().Save(graph))
	}
	testCollections[1].Entries = append(testCollections[1].Entries, &CollectionEntry{GraphID: testGraphs[1].ID})
	testItemList(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionDelete(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemDelete(b, &Collection{}, testInterfaceToSlice(testCollections), t)
}

func testCollectionDeleteAll(b *Backend, testCollections []*Collection, t *testing.T) {
	testItemDeleteAll(b, &Collection{}, testInterfaceToSlice(testCollections), t)
	assert.Nil(t, b.Storage().Delete(&Graph{}))
}

func testCollectionResolve(b *Backend, testCollections []*Collection, t *testing.T) {
	assert.Equal(t, ErrUnresolvableItem, testCollections[2].Resolve(nil))
	testCollections[1].SetBackend(b)
	testCollections[2].SetBackend(b)
	assert.Nil(t, testCollections[2].Resolve(nil))
	assert.Equal(t, testCollections[1], testCollections[2].Link)
}

func testCollectionExpand(b *Backend, testCollections []*Collection, t *testing.T) {
	collection := testCollections[2].Clone()
	assert.Nil(t, collection.Expand(maputil.Map{"source": "other1"}))
	assert.Equal(t, "other1", collection.Options["title"])

	collection = testCollections[2]
	assert.Nil(t, collection.Expand(nil))
	assert.Equal(t, testCollections[2].Attributes["source"], collection.Entries[0].Options["title"])
	assert.Equal(t, testCollections[2].Attributes["source"], collection.Options["title"])
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
	assert.Nil(t, err)
	assert.Len(t, *tree, len(*expected))
	assert.Equal(t, expected, tree)
}
