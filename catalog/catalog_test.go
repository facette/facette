package catalog

import (
	"reflect"
	"sort"
	"testing"

	"facette.io/sliceutil"
)

var (
	testCatalogs catalogList
	testRecords  []*Record
)

func init() {
	testCatalogs = catalogList{
		NewCatalog("catalog1"),
		NewCatalog("catalog2"),
	}

	testRecords = []*Record{
		&Record{Origin: "origin1", Source: "source1", Metric: "metric1"},
		&Record{Origin: "origin1", Source: "source1", Metric: "metric2"},
		&Record{Origin: "origin2", Source: "source2", Metric: "metric3"},
	}

	testCatalogs[1].Insert(testRecords[2])
}

func Test_Catalog_Name(t *testing.T) {
	if result := testCatalogs[0].Name(); result != "catalog1" {
		t.Logf("\nExpected %#v\nbut got  %#v", "catalog1", result)
		t.Fail()
	}
}

func Test_Catalog_SetPriority(t *testing.T) {
	testCatalogs[0].SetPriority(10)

	if testCatalogs[0].priority != 10 {
		t.Logf("\nExpected %#v\nbut got  %#v", 10, testCatalogs[0].priority)
		t.Fail()
	}
}

func Test_Catalog_Insert(t *testing.T) {
	for _, r := range testRecords {
		testCatalogs[0].Insert(r)
	}
}

func Test_Catalog_Origin(t *testing.T) {
	origin, err := testCatalogs[0].Origin(testRecords[0].Origin)
	if err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if c := origin.Catalog(); !reflect.DeepEqual(c, testCatalogs[0]) {
		t.Logf("\nExpected %#v\nbut got  %#v", testCatalogs[0], c)
		t.Fail()
	}

	expectedSources := []string{}
	for _, r := range testRecords {
		if r.Origin == testRecords[0].Origin && !sliceutil.Has(expectedSources, r.Source) {
			expectedSources = append(expectedSources, r.Source)
		}
	}
	sort.Strings(expectedSources)

	sources := []string{}
	for _, s := range origin.Sources() {
		sources = append(sources, s.Name)
	}

	if !reflect.DeepEqual(sources, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, sources)
		t.Fail()
	}
}

func Test_Catalog_Origin_Sources(t *testing.T) {
	expectedSources := []string{"source1"}

	origin, _ := testCatalogs[0].Origin(testRecords[0].Origin)
	sources := []string{}
	for _, o := range origin.Sources() {
		sources = append(sources, o.Name)
	}

	if !reflect.DeepEqual(sources, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, sources)
		t.Fail()
	}
}

func Test_Catalog_Origins(t *testing.T) {
	expectedOrigins := []string{}
	for _, r := range testRecords {
		if !sliceutil.Has(expectedOrigins, r.Origin) {
			expectedOrigins = append(expectedOrigins, r.Origin)
		}
	}
	sort.Strings(expectedOrigins)

	origins := []string{}
	for _, o := range testCatalogs[0].Origins() {
		origins = append(origins, o.Name)
	}

	if !reflect.DeepEqual(origins, expectedOrigins) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedOrigins, origins)
		t.Fail()
	}
}

func Test_Catalog_Source(t *testing.T) {
	expectedOrigin, _ := testCatalogs[0].Origin(testRecords[1].Origin)

	source, err := testCatalogs[0].Source(testRecords[1].Origin, testRecords[1].Source)
	if err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if o := source.Origin(); !reflect.DeepEqual(o, expectedOrigin) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedOrigin, o)
		t.Fail()
	}

	expectedMetrics := []string{}
	for _, r := range testRecords {
		if r.Origin == testRecords[1].Origin && r.Source == testRecords[1].Source &&
			!sliceutil.Has(expectedMetrics, r.Metric) {
			expectedMetrics = append(expectedMetrics, r.Metric)
		}
	}
	sort.Strings(expectedMetrics)

	metrics := []string{}
	for _, m := range source.Metrics() {
		metrics = append(metrics, m.Name)
	}
	sort.Strings(metrics)

	if !reflect.DeepEqual(metrics, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, metrics)
		t.Fail()
	}
}

func Test_Catalog_Source_Metrics(t *testing.T) {
	expectedMetrics := []string{"metric1", "metric2"}

	source, _ := testCatalogs[0].Source(testRecords[0].Origin, testRecords[0].Source)
	metrics := []string{}
	for _, o := range source.Metrics() {
		metrics = append(metrics, o.Name)
	}

	if !reflect.DeepEqual(metrics, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, metrics)
		t.Fail()
	}
}

func Test_Catalog_Metric(t *testing.T) {
	expectedSource, _ := testCatalogs[0].Source(testRecords[2].Origin, testRecords[2].Source)

	metric, err := testCatalogs[0].Metric(testRecords[2].Origin, testRecords[2].Source, testRecords[2].Metric)
	if err != nil {
		t.Logf("\nExpected <nil>\nbut got  %#v", err)
		t.Fail()
	} else if s := metric.Source(); !reflect.DeepEqual(s, expectedSource) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSource, s)
		t.Fail()
	}
}
