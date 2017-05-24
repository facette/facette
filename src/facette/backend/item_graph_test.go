package backend

import (
	"reflect"
	"testing"

	"github.com/facette/maputil"
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

func testGraphCreate(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemCreate(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphCreateInvalid(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemCreateInvalid(b, &Graph{}, testInterfaceToSlice(testGraphs), t)

	alias := "invalid!"
	if err := b.Storage().Save(&Graph{Item: Item{Name: "name"}, Alias: &alias}); err != ErrInvalidAlias {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidAlias, err)
		t.Fail()
	}
}

func testGraphGet(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemGet(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphGetUnknown(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemGetUnknown(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphUpdate(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemUpdate(b, &Graph{}, testInterfaceToSlice(testGraphs), t)

	val := ""
	testGraphs[0].Alias = &val
	testGraphs[0].LinkID = &val

	if err := b.Storage().Save(testGraphs[0]); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	}

	graph := &Graph{}
	if err := b.Storage().Get("name", "item1", graph); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if !reflect.DeepEqual(graph, testGraphs[0]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testGraphs[0], graph)
		t.Fail()
	} else if testGraphs[0].Alias != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", testGraphs[0].Alias)
		t.Fail()
	} else if testGraphs[0].LinkID != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", testGraphs[0].LinkID)
		t.Fail()
	}
}

func testGraphCount(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemCount(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphList(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemList(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphDelete(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemDelete(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphDeleteAll(b *Backend, testGraphs []*Graph, t *testing.T) {
	testItemDeleteAll(b, &Graph{}, testInterfaceToSlice(testGraphs), t)
}

func testGraphResolve(b *Backend, testGraphs []*Graph, t *testing.T) {
	if err := testGraphs[2].Resolve(); err != ErrUnresolvableItem {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrUnresolvableItem, err)
		t.Fail()
	}

	testGraphs[1].SetBackend(b)
	testGraphs[2].SetBackend(b)

	if err := testGraphs[2].Resolve(); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if !reflect.DeepEqual(testGraphs[2].Link, testGraphs[1]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testGraphs[1], testGraphs[2].Link)
		t.Fail()
	}
}

func testGraphExpand(b *Backend, testGraphs []*Graph, t *testing.T) {
	graph := testGraphs[2].Clone()
	if err := graph.Expand(nil); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if graph.Groups[0].Series[0].Source != graph.Attributes["source"] {
		t.Logf("\nExpected %#v\nbut got  %#v", graph.Attributes["source"], graph.Groups[0].Series[0].Source)
		t.Fail()
	} else if graph.Options["title"] != testGraphs[2].Attributes["source"] {
		t.Logf("\nExpected %#v\nbut got  %#v", testGraphs[2].Attributes["source"], graph.Options["title"])
		t.Fail()
	}

	graph = testGraphs[2].Clone()
	if err := graph.Expand(maputil.Map{"source": "other1"}); err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if graph.Groups[0].Series[0].Source != "other1" {
		t.Logf("\nExpected %#v\nbut got  %#v", "other1", graph.Groups[0].Series[0].Source)
		t.Fail()
	} else if graph.Options["title"] != "other1" {
		t.Logf("\nExpected %#v\nbut got  %#v", "other1", graph.Options["title"])
		t.Fail()
	}
}
