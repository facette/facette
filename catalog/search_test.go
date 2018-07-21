package catalog

import (
	"reflect"
	"testing"
)

var (
	testSearcher *Searcher
)

func init() {
	testSearcher = NewSearcher()
}

func Test_Search_Register(t *testing.T) {
	expectedCatalogs := catalogList{testCatalogs[0], testCatalogs[1]}

	for _, c := range testCatalogs {
		testSearcher.Register(c)
	}

	if !reflect.DeepEqual(testSearcher.catalogs, expectedCatalogs) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedCatalogs, testSearcher.catalogs)
		t.Fail()
	}
}

func Test_Search_ApplyPriorities(t *testing.T) {
	expectedCatalogs := catalogList{testCatalogs[1], testCatalogs[0]}

	testSearcher.ApplyPriorities()

	if !reflect.DeepEqual(testSearcher.catalogs, expectedCatalogs) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedCatalogs, testSearcher.catalogs)
		t.Fail()
	}
}

func Test_Search_Origins(t *testing.T) {
	expectedOrigins := make([]*Origin, 3)
	expectedOrigins[0], _ = testCatalogs[1].Origin("origin2")
	expectedOrigins[1], _ = testCatalogs[0].Origin("origin1")
	expectedOrigins[2], _ = testCatalogs[0].Origin("origin2")

	if result := testSearcher.Origins("", -1); !reflect.DeepEqual(result, expectedOrigins) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedOrigins, result)
		t.Fail()
	}
}

func Test_Search_Origins_Limit(t *testing.T) {
	expectedOrigins := make([]*Origin, 1)
	expectedOrigins[0], _ = testCatalogs[1].Origin("origin2")

	if result := testSearcher.Origins("", 1); !reflect.DeepEqual(result, expectedOrigins) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedOrigins, result)
		t.Fail()
	}
}

func Test_Search_Origins_Name(t *testing.T) {
	expectedOrigins := make([]*Origin, 2)
	expectedOrigins[0], _ = testCatalogs[1].Origin("origin2")
	expectedOrigins[1], _ = testCatalogs[0].Origin("origin2")

	if result := testSearcher.Origins("origin2", -1); !reflect.DeepEqual(result, expectedOrigins) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedOrigins, result)
		t.Fail()
	}
}

func Test_Search_Origins_Name_Limit(t *testing.T) {
	expectedOrigins := make([]*Origin, 1)
	expectedOrigins[0], _ = testCatalogs[1].Origin("origin2")

	if result := testSearcher.Origins("origin2", 1); !reflect.DeepEqual(result, expectedOrigins) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedOrigins, result)
		t.Fail()
	}
}

func Test_Search_Sources(t *testing.T) {
	expectedSources := make([]*Source, 3)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")
	expectedSources[1], _ = testCatalogs[0].Source("origin1", "source1")
	expectedSources[2], _ = testCatalogs[0].Source("origin2", "source2")

	if result := testSearcher.Sources("", "", -1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_Limit(t *testing.T) {
	expectedSources := make([]*Source, 1)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")

	if result := testSearcher.Sources("", "", 1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_Origin(t *testing.T) {
	expectedSources := make([]*Source, 2)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")
	expectedSources[1], _ = testCatalogs[0].Source("origin2", "source2")

	if result := testSearcher.Sources("origin2", "", -1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_Origin_Limit(t *testing.T) {
	expectedSources := make([]*Source, 1)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")

	if result := testSearcher.Sources("origin2", "", 1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_OriginName(t *testing.T) {
	expectedSources := make([]*Source, 2)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")
	expectedSources[1], _ = testCatalogs[0].Source("origin2", "source2")

	if result := testSearcher.Sources("origin2", "source2", -1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_OriginName_Limit(t *testing.T) {
	expectedSources := make([]*Source, 1)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")

	if result := testSearcher.Sources("origin2", "source2", 1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_Name(t *testing.T) {
	expectedSources := make([]*Source, 2)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")
	expectedSources[1], _ = testCatalogs[0].Source("origin2", "source2")

	if result := testSearcher.Sources("", "source2", -1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Sources_Name_Limit(t *testing.T) {
	expectedSources := make([]*Source, 1)
	expectedSources[0], _ = testCatalogs[1].Source("origin2", "source2")

	if result := testSearcher.Sources("", "source2", 1); !reflect.DeepEqual(result, expectedSources) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedSources, result)
		t.Fail()
	}
}

func Test_Search_Metrics(t *testing.T) {
	expectedMetrics := make([]*Metric, 4)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expectedMetrics[2], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")
	expectedMetrics[3], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("", "", "", -1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("", "", "", 1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Origin(t *testing.T) {
	expectedMetrics := make([]*Metric, 2)
	expectedMetrics[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")

	if result := testSearcher.Metrics("origin1", "", "", -1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Origin_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")

	if result := testSearcher.Metrics("origin1", "", "", 1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_OriginSource(t *testing.T) {
	expectedMetrics := make([]*Metric, 2)
	expectedMetrics[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")

	if result := testSearcher.Metrics("origin1", "source1", "", -1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_OriginSource_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")

	if result := testSearcher.Metrics("origin1", "source1", "", 1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_OriginSourceName(t *testing.T) {
	expectedMetrics := make([]*Metric, 2)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("origin2", "source2", "metric3", -1); !reflect.DeepEqual(result,
		expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_OriginSourceName_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("origin2", "source2", "metric3", 1); !reflect.DeepEqual(result,
		expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Source(t *testing.T) {
	expectedMetrics := make([]*Metric, 2)
	expectedMetrics[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")

	if result := testSearcher.Metrics("", "source1", "", -1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Source_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")

	if result := testSearcher.Metrics("", "source1", "", 1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_SourceName(t *testing.T) {
	expectedMetrics := make([]*Metric, 2)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("", "source2", "metric3", -1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_SourceName_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("", "source2", "metric3", 1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Name(t *testing.T) {
	expectedMetrics := make([]*Metric, 2)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expectedMetrics[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("", "", "metric3", -1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Metrics_Name_Limit(t *testing.T) {
	expectedMetrics := make([]*Metric, 1)
	expectedMetrics[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")

	if result := testSearcher.Metrics("", "", "metric3", 1); !reflect.DeepEqual(result, expectedMetrics) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedMetrics, result)
		t.Fail()
	}
}

func Test_Search_Unregister(t *testing.T) {
	expectedCatalogs := catalogList{}

	for _, c := range testCatalogs {
		testSearcher.Unregister(c)
	}

	if !reflect.DeepEqual(testSearcher.catalogs, expectedCatalogs) {
		t.Logf("\nExpected %#v\nbut got  %#v", expectedCatalogs, testSearcher.catalogs)
		t.Fail()
	}
}
