package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/server"
)

var (
	serverConfig *config.Config
)

func Test_originList(test *testing.T) {
	base := []string{"test"}
	result := make([]string, 0)

	// Test GET on source list
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/origins/", serverConfig.BindAddr),
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
	base := &server.SourceResponse{Name: "source1", Origins: []string{"test"}}
	result := &server.SourceResponse{}

	// Test GET on source item
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/sources/source1", serverConfig.BindAddr),
		nil, false, &result)
	result.Updated = ""

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/sources/unknown1", serverConfig.BindAddr),
		nil, false, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_sourceList(test *testing.T) {
	base := []string{"source1", "source2"}
	result := make([]string, 0)

	// Test GET on source list
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/sources/", serverConfig.BindAddr), nil,
		false, &result)

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
	base := &server.SourceResponse{Name: "source1", Origins: []string{"test"}}
	result := &server.SourceResponse{}

	// Test GET on source item
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/sources/source1", serverConfig.BindAddr),
		nil, false, &result)
	result.Updated = ""

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/sources/unknown1", serverConfig.BindAddr),
		nil, false, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_metricList(test *testing.T) {
	// Test #1 GET on metrics list
	base := []string{"database1/test", "database2/test", "database3/test"}
	result := make([]string, 0)

	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/metrics/", serverConfig.BindAddr), nil,
		false, &result)

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

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/metrics/?source=source1",
		serverConfig.BindAddr), nil, false, &result)

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
	base := &server.MetricResponse{Name: "database2/test", Sources: []string{"source1", "source2"},
		Origins: []string{"test"}}
	result := &server.MetricResponse{}

	// Test GET on metric item
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/metrics/database2/test",
		serverConfig.BindAddr), nil, false, &result)
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
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/catalog/metrics/unknown1/test",
		serverConfig.BindAddr), nil, false, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_sourceGroupHandle(test *testing.T) {
	// Define a sample source group
	group := &library.Group{Item: library.Item{Name: "group1", Description: "A great group description."}}
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "glob:source*", Origin: "test"})

	expandData := server.ExpandRequest{[3]string{"test", "group:group1-updated", "database1/test"}}

	expandBase := server.ExpandRequest{}
	expandBase = append(expandBase, [3]string{"test", "source1", "database1/test"})
	expandBase = append(expandBase, [3]string{"test", "source2", "database1/test"})

	execGroupHandle(test, "sourcegroups", group, expandData, expandBase)
}

func Test_metricGroupHandle(test *testing.T) {
	// Define a sample metric group
	group := &library.Group{Item: library.Item{Name: "group1", Description: "A great group description."}}
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "database1/test", Origin: "test"})
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "regexp:database[23]/test", Origin: "test"})

	expandData := server.ExpandRequest{[3]string{"test", "source1", "group:group1-updated"}}

	expandBase := server.ExpandRequest{}
	expandBase = append(expandBase, [3]string{"test", "source1", "database1/test"})
	expandBase = append(expandBase, [3]string{"test", "source1", "database2/test"})
	expandBase = append(expandBase, [3]string{"test", "source1", "database3/test"})

	execGroupHandle(test, "metricgroups", group, expandData, expandBase)
}

func Test_graphHandle(test *testing.T) {
	baseURL := fmt.Sprintf("http://%s/library/graphs/", serverConfig.BindAddr)

	// Define a sample graph
	stack := &library.Stack{Name: "stack0"}

	group := &library.OperGroup{Name: "group0", Type: connector.OperGroupTypeAvg}
	group.Series = append(group.Series, &library.Serie{Name: "serie0", Origin: "test", Source: "source1",
		Metric: "database1/test"})
	group.Series = append(group.Series, &library.Serie{Name: "serie1", Origin: "test", Source: "source2",
		Metric: "group:group1"})

	stack.Groups = append(stack.Groups, group)

	group = &library.OperGroup{Name: "serie2", Type: connector.OperGroupTypeNone}
	group.Series = append(group.Series, &library.Serie{Name: "serie2", Origin: "test", Source: "group:group1",
		Metric: "database2/test"})

	stack.Groups = append(stack.Groups, group)

	graphBase := &library.Graph{Item: library.Item{Name: "graph1", Description: "A great graph description."},
		StackMode: library.StackModeNormal}
	graphBase.Stacks = append(graphBase.Stacks, stack)

	// Test #1 GET on graphs list
	listBase := &server.ItemListResponse{}
	listResult := &server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, false, &listResult)

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
	data, _ := json.Marshal(graphBase)

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
	graphResult := &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test #2 GET on graphs list
	listBase = &server.ItemListResponse{&server.ItemResponse{
		ID:          graphBase.ID,
		Name:        graphBase.Name,
		Description: graphBase.Description,
	}}

	listResult = &server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range *listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on graph item
	graphBase.Name = "graph1-updated"

	data, _ = json.Marshal(graphBase)

	response = execTestRequest(test, "PUT", baseURL+graphBase.ID, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "PUT", baseURL+graphBase.ID, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test #2 GET on graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test DELETE on graph item
	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, true, nil)

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

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, false, &graphResult)

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

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, false, &graphResult)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_collectionHandle(test *testing.T) {
	var collectionBase struct {
		*library.Collection
		Parent string `json:"parent"`
	}

	baseURL := fmt.Sprintf("http://%s/library/collections/", serverConfig.BindAddr)

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
	listBase := &server.ItemListResponse{}
	listResult := &server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, false, &listResult)

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
	data, _ := json.Marshal(collectionBase)

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
	collectionResult := &library.Collection{}

	response = execTestRequest(test, "GET", baseURL+collectionBase.ID, nil, false, &collectionResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(collectionBase.Collection, collectionResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", collectionBase.Collection, collectionResult)
		test.Fail()
	}

	// Test #2 GET on collections list
	listBase = &server.ItemListResponse{&server.ItemResponse{
		ID:          collectionBase.ID,
		Name:        collectionBase.Name,
		Description: collectionBase.Description,
	}}

	listResult = &server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range *listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on collection item
	collectionBase.Name = "collection1-updated"

	data, _ = json.Marshal(collectionBase.Collection)

	response = execTestRequest(test, "PUT", baseURL+collectionBase.ID, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "PUT", baseURL+collectionBase.ID, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test #2 GET on collection item
	collectionResult = &library.Collection{}

	response = execTestRequest(test, "GET", baseURL+collectionBase.ID, nil, false, &collectionResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(collectionBase.Collection, collectionResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", collectionBase, collectionResult)
		test.Fail()
	}

	// Test DELETE on collection item
	response = execTestRequest(test, "DELETE", baseURL+collectionBase.ID, nil, false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+collectionBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+collectionBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func execGroupHandle(test *testing.T, urlPrefix string, groupBase *library.Group, expandData,
	expandBase server.ExpandRequest) {

	baseURL := fmt.Sprintf("http://%s/library/%s/", serverConfig.BindAddr, urlPrefix)

	// Test #1 GET on groups list
	listBase := &server.ItemListResponse{}
	listResult := &server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, false, &listResult)

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
	data, _ := json.Marshal(groupBase)

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
	groupResult := &library.Group{}

	response = execTestRequest(test, "GET", baseURL+groupBase.ID, nil, false, &groupResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(groupBase, groupResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", groupBase, groupResult)
		test.Fail()
	}

	// Test #2 GET on groups list
	listBase = &server.ItemListResponse{&server.ItemResponse{
		ID:          groupBase.ID,
		Name:        groupBase.Name,
		Description: groupBase.Description,
	}}

	listResult = &server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, false, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range *listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on group item
	groupBase.Name = "group1-updated"

	data, _ = json.Marshal(groupBase)

	response = execTestRequest(test, "PUT", baseURL+groupBase.ID, strings.NewReader(string(data)), false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "PUT", baseURL+groupBase.ID, strings.NewReader(string(data)), true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test #2 GET on group item
	groupResult = &library.Group{}

	response = execTestRequest(test, "GET", baseURL+groupBase.ID, nil, false, &groupResult)

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

	expandResult := make([]server.ExpandRequest, 0)

	response = execTestRequest(test, "POST", fmt.Sprintf("http://%s/library/expand", serverConfig.BindAddr),
		strings.NewReader(string(data)), false, &expandResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if len(expandResult) == 0 {
		test.Logf("\nExpected %#v\nbut got  %#v", expandBase, expandResult)
		test.Fail()
	} else if !reflect.DeepEqual(expandBase, expandResult[0]) {
		test.Logf("\nExpected %#v\nbut got  %#v", expandBase, expandResult[0])
		test.Fail()
	}

	// Test DELETE on group item
	response = execTestRequest(test, "DELETE", baseURL+groupBase.ID, nil, false, nil)

	if response.StatusCode != http.StatusUnauthorized {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusUnauthorized, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+groupBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+groupBase.ID, nil, true, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func execTestRequest(test *testing.T, method, url string, data io.Reader, auth bool,
	result interface{}) *http.Response {

	request, err := http.NewRequest(method, url, data)
	if err != nil {
		test.Fatal(err.Error())
	}

	if auth {
		// Add authentication (login: unittest, password: unittest)
		request.Header.Add("Authorization", "Basic dW5pdHRlc3Q6dW5pdHRlc3Q=")
	}

	if data != nil {
		request.Header.Add("Content-Type", "application/json")
	}

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		test.Fatal(err.Error())
	}

	defer response.Body.Close()

	if result != nil {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			test.Fatal(err.Error())
		}

		json.Unmarshal(body, result)
	}

	return response
}

func init() {
	// Load server configuration
	serverConfig = &config.Config{}
	if err := serverConfig.Load(flagConfig); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}
