package server

import (
	"encoding/json"
	"facette/backend"
	"facette/common"
	"facette/library"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	flagConfig   string
	serverConfig *common.Config
)

func Test_originList(test *testing.T) {
	var (
		base     []string
		response *http.Response
		result   []string
	)

	base = []string{"test"}

	// Test GET on source list
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/origins", serverConfig.BindAddr, URLCatalogPath),
		nil, false, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}
}

func Test_originShow(test *testing.T) {
	var (
		base     *sourceShowResponse
		response *http.Response
		result   *sourceShowResponse
	)

	base = &sourceShowResponse{Name: "source1", Origins: []string{"test"}}
	result = &sourceShowResponse{}

	// Test GET on source item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/sources/source1", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)
	result.Updated = ""

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/sources/unknown1", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_sourceList(test *testing.T) {
	var (
		base     []string
		response *http.Response
		result   []string
	)

	base = []string{"source1", "source2"}

	// Test GET on source list
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/sources", serverConfig.BindAddr, URLCatalogPath),
		nil, false, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}
}

func Test_sourceShow(test *testing.T) {
	var (
		base     *sourceShowResponse
		response *http.Response
		result   *sourceShowResponse
	)

	base = &sourceShowResponse{Name: "source1", Origins: []string{"test"}}
	result = &sourceShowResponse{}

	// Test GET on source item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/sources/source1", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)
	result.Updated = ""

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/sources/unknown1", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_metricList(test *testing.T) {
	var (
		base     []string
		response *http.Response
		result   []string
	)

	// Test #1 GET on metrics list
	base = []string{"database1/test", "database2/test", "database3/test"}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/metrics", serverConfig.BindAddr, URLCatalogPath),
		nil, false, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test #2 GET on metrics list
	base = []string{"database1/test", "database2/test"}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/metrics?source=source1", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}
}

func Test_metricShow(test *testing.T) {
	var (
		base     *metricShowResponse
		response *http.Response
		result   *metricShowResponse
	)

	base = &metricShowResponse{Name: "database2/test", Sources: []string{"source1", "source2"},
		Origins: []string{"test"}}
	result = &metricShowResponse{}

	// Test GET on metric item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/metrics/database2/test", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)
	result.Updated = ""

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on unknown metric item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s%s/metrics/unknown1/test", serverConfig.BindAddr,
		URLCatalogPath), nil, false, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_sourceGroupHandle(test *testing.T) {
	var (
		expandBase expandRequest
		expandData expandRequest
		group      *library.Group
	)

	// Define a sample source group
	group = &library.Group{Item: library.Item{Name: "group1", Description: "A great group description."}}
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "glob:source*", Origin: "test"})

	expandData = expandRequest{[3]string{"test", "group:group1-updated", "database1/test"}}
	expandBase = append(expandBase, [3]string{"test", "source1", "database1/test"})
	expandBase = append(expandBase, [3]string{"test", "source2", "database1/test"})

	execGroupHandle(test, "sourcegroups", group, expandData, expandBase)
}

func Test_metricGroupHandle(test *testing.T) {
	var (
		expandBase expandRequest
		expandData expandRequest
		group      *library.Group
	)

	// Define a sample metric group
	group = &library.Group{Item: library.Item{Name: "group1", Description: "A great group description."}}
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "database1/test", Origin: "test"})
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "regexp:database[23]/test", Origin: "test"})

	expandData = expandRequest{[3]string{"test", "source1", "group:group1-updated"}}
	expandBase = append(expandBase, [3]string{"test", "source1", "database1/test"})
	expandBase = append(expandBase, [3]string{"test", "source1", "database2/test"})
	expandBase = append(expandBase, [3]string{"test", "source1", "database3/test"})

	execGroupHandle(test, "metricgroups", group, expandData, expandBase)
}

func Test_graphHandle(test *testing.T) {
	var (
		baseURL     string
		graphBase   *library.Graph
		graphResult *library.Graph
		data        []byte
		group       *library.OperGroup
		listBase    *libraryListResponse
		listResult  *libraryListResponse
		response    *http.Response
		stack       *library.Stack
	)

	baseURL = fmt.Sprintf("http://%s%s/graphs", serverConfig.BindAddr, URLLibraryPath)

	// Define a sample graph
	stack = &library.Stack{Name: "stack0"}

	group = &library.OperGroup{Name: "group0", Type: backend.OperGroupTypeAvg}
	group.Series = append(group.Series, &library.Serie{Name: "serie0", Origin: "test", Source: "source1",
		Metric: "database1/test"})
	group.Series = append(group.Series, &library.Serie{Name: "serie1", Origin: "test", Source: "source2",
		Metric: "group:group1"})

	stack.Groups = append(stack.Groups, group)

	group = &library.OperGroup{Name: "serie2", Type: backend.OperGroupTypeNone}
	group.Series = append(group.Series, &library.Serie{Name: "serie2", Origin: "test", Source: "group:group1",
		Metric: "database2/test"})

	stack.Groups = append(stack.Groups, group)

	graphBase = &library.Graph{Item: library.Item{Name: "graph1", Description: "A great graph description."},
		StackMode: library.StackModeNormal}
	graphBase.Stacks = append(graphBase.Stacks, stack)

	// Test #1 GET on graphs list
	listBase = &libraryListResponse{}
	listResult = &libraryListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult.Items)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown graph item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, false, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into graph
	data, _ = json.Marshal(graphBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	graphBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test #1 GET on graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+"/"+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test #2 GET on graphs list
	listBase = &libraryListResponse{}
	listBase.Items = append(listBase.Items, &libraryItemResponse{ID: graphBase.ID, Name: graphBase.Name,
		Description: graphBase.Description})

	listResult = &libraryListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult.Items)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult.Items {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on graph item
	graphBase.Name = "graph1-updated"

	data, _ = json.Marshal(graphBase)

	response = execTestRequest(test, "PUT", baseURL+"/"+graphBase.ID, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "PUT", baseURL+"/"+graphBase.ID, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test #2 GET on graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+"/"+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test DELETE on graph item
	response = execTestRequest(test, "DELETE", baseURL+"/"+graphBase.ID, nil, false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+"/"+graphBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+"/"+graphBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test volatile POST into graph
	graphBase.ID = ""
	data, _ = json.Marshal(graphBase)

	response = execTestRequest(test, "POST", baseURL+"?volatile=1", strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "POST", baseURL+"?volatile=1", strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	graphBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test #1 GET on volatile graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+"/"+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test #2 GET on volatile graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+"/"+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_collectionHandle(test *testing.T) {
	var (
		baseURL        string
		collectionBase struct {
			*library.Collection
			Parent string `json:"parent"`
		}
		collectionResult *library.Collection
		data             []byte
		listBase         *libraryListResponse
		listResult       *libraryListResponse
		response         *http.Response
	)

	baseURL = fmt.Sprintf("http://%s%s/collections", serverConfig.BindAddr, URLLibraryPath)

	// Define a sample collection
	collectionBase.Collection = &library.Collection{Item: library.Item{Name: "collection0",
		Description: "A great collection description."}}

	collectionBase.Entries = append(collectionBase.Entries,
		&library.CollectionEntry{ID: "00000000-0000-0000-0000-000000000000",
			Options: map[string]string{"range": "-1h"}})
	collectionBase.Entries = append(collectionBase.Entries,
		&library.CollectionEntry{ID: "00000000-0000-0000-0000-000000000000",
			Options: map[string]string{"range": "-1d"}})
	collectionBase.Entries = append(collectionBase.Entries,
		&library.CollectionEntry{ID: "00000000-0000-0000-0000-000000000000",
			Options: map[string]string{"range": "-1w"}})

	// Test #1 GET on collections list
	listBase = &libraryListResponse{}
	listResult = &libraryListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult.Items)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown collection item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, false, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into collection
	data, _ = json.Marshal(collectionBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	collectionBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test #1 GET on collection item
	collectionResult = &library.Collection{}

	response = execTestRequest(test, "GET", baseURL+"/"+collectionBase.ID, nil, false, &collectionResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(collectionBase.Collection, collectionResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", collectionBase.Collection, collectionResult)
		test.Fail()
	}

	// Test #2 GET on collections list
	listBase = &libraryListResponse{}
	listBase.Items = append(listBase.Items, &libraryItemResponse{ID: collectionBase.ID, Name: collectionBase.Name,
		Description: collectionBase.Description})

	listResult = &libraryListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult.Items)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult.Items {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on collection item
	collectionBase.Name = "collection1-updated"

	data, _ = json.Marshal(collectionBase.Collection)

	response = execTestRequest(test, "PUT", baseURL+"/"+collectionBase.ID, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "PUT", baseURL+"/"+collectionBase.ID, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test #2 GET on collection item
	collectionResult = &library.Collection{}

	response = execTestRequest(test, "GET", baseURL+"/"+collectionBase.ID, nil, false, &collectionResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(collectionBase.Collection, collectionResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", collectionBase, collectionResult)
		test.Fail()
	}

	// Test DELETE on collection item
	response = execTestRequest(test, "DELETE", baseURL+"/"+collectionBase.ID, nil, false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+"/"+collectionBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+"/"+collectionBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func execGroupHandle(test *testing.T, urlPrefix string, groupBase *library.Group, expandData,
	expandBase expandRequest) {
	var (
		baseURL      string
		data         []byte
		expandResult []expandRequest
		groupResult  *library.Group
		listBase     *libraryListResponse
		listResult   *libraryListResponse
		response     *http.Response
	)

	baseURL = fmt.Sprintf("http://%s%s/%s", serverConfig.BindAddr, URLLibraryPath, urlPrefix)

	// Test #1 GET on groups list
	listBase = &libraryListResponse{}
	listResult = &libraryListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult.Items)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown group item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, false, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into group
	data, _ = json.Marshal(groupBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	groupBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test #1 GET on group item
	groupResult = &library.Group{}

	response = execTestRequest(test, "GET", baseURL+"/"+groupBase.ID, nil, false, &groupResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(groupBase, groupResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", groupBase, groupResult)
		test.Fail()
	}

	// Test #2 GET on groups list
	listBase = &libraryListResponse{}
	listBase.Items = append(listBase.Items, &libraryItemResponse{ID: groupBase.ID, Name: groupBase.Name,
		Description: groupBase.Description})

	listResult = &libraryListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult.Items)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult.Items {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on group item
	groupBase.Name = "group1-updated"

	data, _ = json.Marshal(groupBase)

	response = execTestRequest(test, "PUT", baseURL+"/"+groupBase.ID, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "PUT", baseURL+"/"+groupBase.ID, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test #2 GET on group item
	groupResult = &library.Group{}

	response = execTestRequest(test, "GET", baseURL+"/"+groupBase.ID, nil, false, &groupResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(groupBase, groupResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", groupBase, groupResult)
		test.Fail()
	}

	// Test group expansion
	data, _ = json.Marshal(expandData)

	response = execTestRequest(test, "POST", fmt.Sprintf("http://%s%s/expand", serverConfig.BindAddr, URLLibraryPath),
		strings.NewReader(string(data)), false, &expandResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(expandBase, expandResult[0]) {
		test.Logf("\nExpected %#v\nbut got  %#v", expandBase, expandResult[0])
		test.Fail()
	}

	// Test DELETE on group item
	response = execTestRequest(test, "DELETE", baseURL+"/"+groupBase.ID, nil, false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+"/"+groupBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+"/"+groupBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func execTestRequest(test *testing.T, method, url string, data io.Reader, auth bool,
	result interface{}) *http.Response {
	var (
		body     []byte
		client   *http.Client
		err      error
		request  *http.Request
		response *http.Response
	)

	if request, err = http.NewRequest(method, url, data); err != nil {
		test.Fatal(err.Error())
	}

	if auth {
		// Add authentication (login: unittest, password: unittest)
		request.Header.Add("Authorization", "Basic dW5pdHRlc3Q6dW5pdHRlc3Q=")
	}

	if data != nil {
		request.Header.Add("Content-Type", "application/json")
	}

	client = &http.Client{}

	if response, err = client.Do(request); err != nil {
		test.Fatal(err.Error())
	}

	defer response.Body.Close()

	if result != nil {
		if body, err = ioutil.ReadAll(response.Body); err != nil {
			test.Fatal(err.Error())
		}

		json.Unmarshal(body, result)
	}

	return response
}

func init() {
	var (
		err error
	)

	flag.StringVar(&flagConfig, "c", common.DefaultConfigFile, "configuration file path")
	flag.Parse()

	if flagConfig == "" {
		fmt.Printf("Error: configuration file path is mandatory\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load server configuration
	serverConfig = &common.Config{}
	if err = serverConfig.Load(flagConfig); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}
