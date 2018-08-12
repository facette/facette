package storage

import (
	"testing"

	"facette.io/maputil"
	"github.com/stretchr/testify/assert"
)

func testGraphNew() []*Graph {
	tmpl := &Graph{
		Item: Item{
			Name: "item2",
		},
		Groups: []*SeriesGroup{
			{
				Series: []*Series{
					{
						Name:   "series1",
						Origin: "origin1",
						Source: "{{ .source }}",
						Metric: "metric1",
						Options: maputil.Map{
							"key1": "abc",
						},
					},
				},
				Options: maputil.Map{
					"key1": 123.456,
				},
			},
		},
		Options: map[string]interface{}{
			"title": "{{ .source }}",
		},
		Template: true,
	}

	return []*Graph{
		&Graph{
			Item: Item{
				Name: "item1",
			},
			Groups: []*SeriesGroup{
				{
					Series: []*Series{
						{Name: "series1", Origin: "origin1", Source: "source1", Metric: "metric1"},
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A great graph title",
			},
		},

		tmpl,

		&Graph{
			Item: Item{
				Name: "item3",
			},
			LinkID: &tmpl.ID,
			Attributes: map[string]interface{}{
				"source": "source1",
			},
		},
	}
}

func testGraphCreate(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemCreate(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphCreateInvalid(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemCreateInvalid(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
	alias := "invalid!"
	assert.Equal(t, ErrInvalidAlias, s.SQL().Save(&Graph{Item: Item{Name: "name"}, Alias: &alias}))
}

func testGraphGet(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemGet(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphGetUnknown(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemGetUnknown(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphUpdate(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemUpdate(s, &Graph{}, testInterfaceToSlice(testGraphs), t)

	val := ""
	testGraphs[0].Alias = &val
	testGraphs[0].LinkID = &val

	assert.Nil(t, s.SQL().Save(testGraphs[0]))

	graph := &Graph{}
	assert.Nil(t, s.SQL().Get("name", "item1", graph, true))
	assert.Nil(t, graph.Alias)
	assert.Nil(t, graph.LinkID)
	assert.Equal(t, testGraphs[0], graph)
}

func testGraphCount(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemCount(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphList(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemList(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphDelete(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemDelete(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphDeleteAll(s *Storage, testGraphs []*Graph, t *testing.T) {
	testItemDeleteAll(s, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphResolve(s *Storage, testGraphs []*Graph, t *testing.T) {
	assert.Equal(t, ErrUnresolvableItem, testGraphs[2].Resolve())
	testGraphs[1].SetStorage(s)
	testGraphs[2].SetStorage(s)
	assert.Nil(t, testGraphs[2].Resolve())
	assert.Equal(t, testGraphs[1], testGraphs[2].Link)
}

func testGraphExpand(s *Storage, testGraphs []*Graph, t *testing.T) {
	graph := testGraphs[2].Clone()
	assert.Nil(t, graph.Expand(nil))
	assert.Equal(t, graph.Attributes["source"], graph.Groups[0].Series[0].Source)

	graph = testGraphs[2].Clone()
	assert.Nil(t, graph.Expand(maputil.Map{"source": "other1"}))
	assert.Equal(t, "other1", graph.Groups[0].Series[0].Source)
	assert.Equal(t, "other1", graph.Options["title"])
}
