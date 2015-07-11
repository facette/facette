package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/plot"
	"github.com/facette/facette/pkg/server"
	"github.com/facette/facette/pkg/utils"
	"github.com/ziutek/rrd"
)

var (
	serverConfig *config.Config
)

func Test_CatalogOriginList(test *testing.T) {
	var result []string

	base := []string{
		"test1",
		"test2",
	}

	// Test GET on source list
	result = make([]string, 0)

	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/origins/", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on source list (offset and limit)
	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/origins/?limit=1", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[:1], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[:1], result)
		test.Fail()
	}

	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/origins/?offset=1&limit=1",
		serverConfig.BindAddr), nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[1:2], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[1:2], result)
		test.Fail()
	}
}

func Test_CatalogOriginGet(test *testing.T) {
	base := &server.SourceResponse{Name: "source1", Origins: []string{"test1", "test2"}}
	result := &server.SourceResponse{}

	// Test GET on source1 item
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/source1", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on source2 item (with filter settings)
	base = &server.SourceResponse{Name: "source2", Origins: []string{"test1"}}
	result = &server.SourceResponse{}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/source2", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on unknown item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/unknown", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_CatalogSourceList(test *testing.T) {
	var result []string

	base := []string{
		"source1",
		"source2",
	}

	// Test GET on source list
	result = make([]string, 0)

	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/", serverConfig.BindAddr), nil,
		&result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on source list (offset and limit)
	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/?limit=1", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[:1], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[:1], result)
		test.Fail()
	}

	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/?offset=1&limit=1",
		serverConfig.BindAddr), nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[1:2], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[1:2], result)
		test.Fail()
	}
}

func Test_CatalogSourceGet(test *testing.T) {
	base := &server.SourceResponse{Name: "source1", Origins: []string{"test1", "test2"}}
	result := &server.SourceResponse{}

	// Test GET on source1 item
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/source1", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on source2 item
	base = &server.SourceResponse{Name: "source2", Origins: []string{"test1"}}
	result = &server.SourceResponse{}

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/source2", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on unknown item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/sources/unknown", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_CatalogMetricList(test *testing.T) {
	var result []string

	base := []string{
		"database1.test.average",
		"database1/test/average",
		"database2.test.average",
		"database2/test/average",
		"database3/test/average",
	}

	// Test GET on metrics list
	result = make([]string, 0)

	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/metrics/", serverConfig.BindAddr), nil,
		&result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on metrics list (offset and limit)
	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/metrics/?limit=2", serverConfig.BindAddr),
		nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[:2], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[:2], result)
		test.Fail()
	}

	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/metrics/?offset=2&limit=2",
		serverConfig.BindAddr), nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[2:4], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[2:4], result)
		test.Fail()
	}

	// Test GET on metrics list (source-specific)
	result = make([]string, 0)

	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/metrics/?source=source1",
		serverConfig.BindAddr), nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base[:4], result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base[:4], result)
		test.Fail()
	}
}

func Test_CatalogMetricGet(test *testing.T) {
	base := &server.MetricResponse{Name: "database2/test/average", Sources: []string{"source1", "source2"},
		Origins: []string{"test1"}}

	result := &server.MetricResponse{}

	// Test GET on metric item
	response := execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/metrics/database2/test/average",
		serverConfig.BindAddr), nil, &result)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(base, result) {
		test.Logf("\nExpected %#v\nbut got  %#v", base, result)
		test.Fail()
	}

	// Test GET on unknown metric item
	response = execTestRequest(test, "GET", fmt.Sprintf("http://%s/api/v1/catalog/metrics/unknown/test",
		serverConfig.BindAddr), nil, &result)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}
}

func Test_LibraryScaleHandle(test *testing.T) {
	var (
		listBase     server.ItemListResponse
		listResult   server.ItemListResponse
		valuesResult []server.ScaleValueResponse
	)

	baseURL := fmt.Sprintf("http://%s/api/v1/library/scales/", serverConfig.BindAddr)

	// Define a sample scale
	scaleBase := &library.Scale{Item: library.Item{Name: "scale0", Description: "A great scale description."},
		Value: 0.125}

	// Test GET on scales list
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown scale item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into scale
	data, _ := json.Marshal(scaleBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	scaleBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test GET on scale item
	scaleResult := &library.Scale{}

	response = execTestRequest(test, "GET", baseURL+scaleBase.ID, nil, &scaleResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(scaleBase, scaleResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", scaleBase, scaleResult)
		test.Fail()
	}

	// Test GET on scales list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:          scaleBase.ID,
		Name:        scaleBase.Name,
		Description: scaleBase.Description,
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on scales values
	valuesBase := []server.ScaleValueResponse{
		server.ScaleValueResponse{Name: scaleBase.Name, Value: scaleBase.Value},
	}

	response = execTestRequest(test, "GET", baseURL+"/values", nil, &valuesResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(valuesBase, valuesResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", valuesBase, valuesResult)
		test.Fail()
	}

	// Test PUT on scale item
	scaleBase.Name = "scale0-updated"

	data, _ = json.Marshal(scaleBase)

	response = execTestRequest(test, "PUT", baseURL+scaleBase.ID, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test GET on scale item
	scaleResult = &library.Scale{}

	response = execTestRequest(test, "GET", baseURL+scaleBase.ID, nil, &scaleResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(scaleBase, scaleResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", scaleBase, scaleResult)
		test.Fail()
	}

	// Test DELETE on scale item
	response = execTestRequest(test, "DELETE", baseURL+scaleBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+scaleBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test GET on scales list (offset and limit)
	listBase = server.ItemListResponse{}

	for i := 0; i < 3; i++ {
		scaleTemp := &library.Scale{}
		utils.Clone(scaleBase, scaleTemp)

		scaleTemp.ID = ""
		scaleTemp.Name = fmt.Sprintf("scale0-%d", i)

		data, _ = json.Marshal(scaleTemp)

		response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

		if response.StatusCode != http.StatusCreated {
			test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
			test.Fail()
		}

		location := response.Header.Get("Location")

		if location == "" {
			test.Logf("\nExpected `Location' header")
			test.Fail()
		}

		scaleTemp.ID = location[strings.LastIndex(location, "/")+1:]

		listBase = append(listBase, &server.ItemResponse{
			ID:          scaleTemp.ID,
			Name:        scaleTemp.Name,
			Description: scaleTemp.Description,
		})
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?limit=1", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[:1], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[:1], listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?offset=1&limit=2", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[1:3], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[1:3], listResult)
		test.Fail()
	}
}

func Test_LibraryUnitHandle(test *testing.T) {
	var (
		listBase     server.ItemListResponse
		listResult   server.ItemListResponse
		labelsResult []server.UnitValueResponse
	)

	baseURL := fmt.Sprintf("http://%s/api/v1/library/units/", serverConfig.BindAddr)

	// Define a sample unit
	unitBase := &library.Unit{Item: library.Item{Name: "unit0", Description: "A great unit description."},
		Label: "B"}

	// Test GET on units list
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown unit item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into unit
	data, _ := json.Marshal(unitBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	unitBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test GET on unit item
	unitResult := &library.Unit{}

	response = execTestRequest(test, "GET", baseURL+unitBase.ID, nil, &unitResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(unitBase, unitResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", unitBase, unitResult)
		test.Fail()
	}

	// Test GET on units list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:          unitBase.ID,
		Name:        unitBase.Name,
		Description: unitBase.Description,
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on units labels
	labelsBase := []server.UnitValueResponse{
		server.UnitValueResponse{Name: unitBase.Name, Label: unitBase.Label},
	}

	response = execTestRequest(test, "GET", baseURL+"/labels", nil, &labelsResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(labelsBase, labelsResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", labelsBase, labelsResult)
		test.Fail()
	}

	// Test PUT on unit item
	unitBase.Name = "unit0-updated"

	data, _ = json.Marshal(unitBase)

	response = execTestRequest(test, "PUT", baseURL+unitBase.ID, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test GET on unit item
	unitResult = &library.Unit{}

	response = execTestRequest(test, "GET", baseURL+unitBase.ID, nil, &unitResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(unitBase, unitResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", unitBase, unitResult)
		test.Fail()
	}

	// Test DELETE on unit item
	response = execTestRequest(test, "DELETE", baseURL+unitBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+unitBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test GET on units list (offset and limit)
	listBase = server.ItemListResponse{}

	for i := 0; i < 3; i++ {
		unitTemp := &library.Unit{}
		utils.Clone(unitBase, unitTemp)

		unitTemp.ID = ""
		unitTemp.Name = fmt.Sprintf("unit0-%d", i)

		data, _ = json.Marshal(unitTemp)

		response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

		if response.StatusCode != http.StatusCreated {
			test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
			test.Fail()
		}

		location := response.Header.Get("Location")

		if location == "" {
			test.Logf("\nExpected `Location' header")
			test.Fail()
		}

		unitTemp.ID = location[strings.LastIndex(location, "/")+1:]

		listBase = append(listBase, &server.ItemResponse{
			ID:          unitTemp.ID,
			Name:        unitTemp.Name,
			Description: unitTemp.Description,
		})
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?limit=1", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[:1], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[:1], listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?offset=1&limit=2", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[1:3], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[1:3], listResult)
		test.Fail()
	}
}

func Test_LibrarySourceGroupHandle(test *testing.T) {
	// Define a sample source group
	group := &library.Group{Item: library.Item{Name: "group0", Description: "A great group description."}}
	group.Entries = append(group.Entries, &library.GroupEntry{Pattern: "glob:source*", Origin: "test1"})

	expandData := server.ExpandRequest{[3]string{"test1", "group:group0-updated", "database1/test/average"}}

	expandBase := server.ExpandRequest{}
	expandBase = append(expandBase, [3]string{"test1", "source1", "database1/test/average"})

	execGroupHandle(test, "sourcegroups", group, expandData, expandBase)
}

func Test_LibraryMetricGroupHandle(test *testing.T) {
	// Define a sample metric group
	group := &library.Group{Item: library.Item{Name: "group0", Description: "A great group description."}}

	group.Entries = append(group.Entries, &library.GroupEntry{
		Pattern: "database1/test/average",
		Origin:  "test1",
	})

	group.Entries = append(group.Entries, &library.GroupEntry{
		Pattern: "regexp:database[23]/test/average",
		Origin:  "test1",
	})

	expandData := server.ExpandRequest{[3]string{"test1", "source1", "group:group0-updated"}}

	expandBase := server.ExpandRequest{}
	expandBase = append(expandBase, [3]string{"test1", "source1", "database1/test/average"})
	expandBase = append(expandBase, [3]string{"test1", "source1", "database2/test/average"})

	execGroupHandle(test, "metricgroups", group, expandData, expandBase)
}

func Test_LibraryGraphHandle(test *testing.T) {
	var (
		listBase   server.ItemListResponse
		listResult server.ItemListResponse
	)

	baseURL := fmt.Sprintf("http://%s/api/v1/library/graphs/", serverConfig.BindAddr)

	// Define a sample graph
	graphBase := &library.Graph{Item: library.Item{Name: "graph0", Description: "A great graph description."},
		StackMode: library.StackModeNormal}

	group := &library.OperGroup{Name: "group0", Type: plot.OperTypeAverage}
	group.Series = append(group.Series, &library.Series{Name: "series0", Origin: "test", Source: "source1",
		Metric: "database1/test"})
	group.Series = append(group.Series, &library.Series{Name: "series1", Origin: "test", Source: "source2",
		Metric: "group:group0"})

	graphBase.Groups = append(graphBase.Groups, group)

	group = &library.OperGroup{Name: "series2", Type: plot.OperTypeNone}
	group.Series = append(group.Series, &library.Series{Name: "series2", Origin: "test", Source: "group:group0",
		Metric: "database2/test"})

	graphBase.Groups = append(graphBase.Groups, group)

	// Test GET on graphs list
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown graph item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into graph
	data, _ := json.Marshal(graphBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	graphBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test GET on graph item
	graphResult := &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test GET on graphs list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:          graphBase.ID,
		Name:        graphBase.Name,
		Description: graphBase.Description,
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on graph item
	graphBase.Name = "graph0-updated"

	data, _ = json.Marshal(graphBase)

	response = execTestRequest(test, "PUT", baseURL+graphBase.ID, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test GET on graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test DELETE on graph item
	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test GET on graphs list (offset and limit)
	listBase = server.ItemListResponse{}

	for i := 0; i < 3; i++ {
		graphTemp := &library.Graph{}
		utils.Clone(graphBase, graphTemp)

		graphTemp.ID = ""
		graphTemp.Name = fmt.Sprintf("graph0-%d", i)

		data, _ = json.Marshal(graphTemp)

		response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

		if response.StatusCode != http.StatusCreated {
			test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
			test.Fail()
		}

		location := response.Header.Get("Location")

		if location == "" {
			test.Logf("\nExpected `Location' header")
			test.Fail()
		}

		graphTemp.ID = location[strings.LastIndex(location, "/")+1:]

		listBase = append(listBase, &server.ItemResponse{
			ID:          graphTemp.ID,
			Name:        graphTemp.Name,
			Description: graphTemp.Description,
		})
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?limit=1", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[:1], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[:1], listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?offset=1&limit=2", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[1:3], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[1:3], listResult)
		test.Fail()
	}

	// Delete created items
	for _, listItem := range listBase {
		response = execTestRequest(test, "DELETE", baseURL+listItem.ID, nil, nil)
	}
}

func Test_LibraryGraphTemplateHandle(test *testing.T) {
	var (
		listBase   server.ItemListResponse
		listResult server.ItemListResponse
	)

	baseURL := fmt.Sprintf("http://%s/api/v1/library/graphs/", serverConfig.BindAddr)

	// Define a sample graph
	graphBase := &library.Graph{
		Item:      library.Item{Name: "graphtmpl0", Description: "A great graph description for {{ .source0 }}."},
		Title:     "graphtmpl-{{ .source0 }}",
		StackMode: library.StackModeNormal,
		Template:  true,
	}

	group := &library.OperGroup{Name: "group0", Type: plot.OperTypeAverage}
	group.Series = append(group.Series, &library.Series{Name: "series0", Origin: "test", Source: "{{ .source0 }}",
		Metric: "{{ .metric0 }}"})

	graphBase.Groups = append(graphBase.Groups, group)

	// Test GET on graphs list
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL+"?type=template", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test POST with extra link or attributes (should fail with 400 status code)
	graphBase.Link = "00000000-0000-0000-0000-000000000000"

	data, _ := json.Marshal(graphBase)
	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusBadRequest {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusBadRequest, response.StatusCode)
		test.Fail()
	}

	graphBase.Link = ""
	graphBase.Attributes = make(map[string]interface{})
	graphBase.Attributes["attr1"] = "value1"

	data, _ = json.Marshal(graphBase)
	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusBadRequest {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusBadRequest, response.StatusCode)
		test.Fail()
	}

	// Test POST into graph of a template
	graphBase.Attributes = nil

	data, _ = json.Marshal(graphBase)
	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	graphBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test GET on graph item
	graphResult := &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test GET on graphs list without template flag (should fail with 400 status code)
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?type=raw", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on graphs list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:          graphBase.ID,
		Name:        graphBase.Name,
		Description: graphBase.Description,
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?type=template", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test POST into graph of a template instance
	tmplBase := &library.Graph{
		Item: library.Item{Name: "graph0"},
		Link: graphBase.ID,
		Attributes: map[string]interface{}{
			"source0": "source1",
			"metric0": "database1/test",
		},
	}

	data, _ = json.Marshal(tmplBase)
	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	tmplBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test POST into graph of a template instance with extra data (should fail with 400 status code)
	tmplBase.Template = true

	data, _ = json.Marshal(tmplBase)
	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusBadRequest {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusBadRequest, response.StatusCode)
		test.Fail()
	}

	tmplBase.Template = false

	tmplBase.Description = "A test description."
	tmplBase.Title = "A test title."
	tmplBase.Type = 1
	tmplBase.StackMode = 1
	tmplBase.UnitType = 1
	tmplBase.UnitLegend = "a"
	tmplBase.Groups = make([]*library.OperGroup, 0)

	data, _ = json.Marshal(tmplBase)
	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusBadRequest {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusBadRequest, response.StatusCode)
		test.Fail()
	}

	tmplBase.Description = ""
	tmplBase.Title = ""
	tmplBase.Type = 0
	tmplBase.StackMode = 0
	tmplBase.UnitType = 0
	tmplBase.UnitLegend = ""
	tmplBase.Groups = nil

	// Test GET on graph template instance item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test GET on graphs list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:   tmplBase.ID,
		Name: tmplBase.Name,
		Description: strings.Replace(graphBase.Description, "{{ .source0 }}",
			tmplBase.Attributes["source0"].(string), -1),
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?type=raw", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[0], listResult[0])
		test.Fail()
	}

	// Test DELETE on graph template instance item
	response = execTestRequest(test, "DELETE", baseURL+tmplBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+tmplBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test PUT on graph item
	graphBase.Name = "graph0-updated"

	data, _ = json.Marshal(graphBase)

	response = execTestRequest(test, "PUT", baseURL+graphBase.ID, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test GET on graph item
	graphResult = &library.Graph{}

	response = execTestRequest(test, "GET", baseURL+graphBase.ID, nil, &graphResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(graphBase, graphResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", graphBase, graphResult)
		test.Fail()
	}

	// Test DELETE on graph item
	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+graphBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test GET on graphs list (offset and limit)
	listBase = server.ItemListResponse{}

	for i := 0; i < 3; i++ {
		graphTemp := &library.Graph{}
		utils.Clone(graphBase, graphTemp)

		graphTemp.ID = ""
		graphTemp.Name = fmt.Sprintf("graph0-%d", i)

		data, _ = json.Marshal(graphTemp)

		response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

		if response.StatusCode != http.StatusCreated {
			test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
			test.Fail()
		}

		location := response.Header.Get("Location")

		if location == "" {
			test.Logf("\nExpected `Location' header")
			test.Fail()
		}

		graphTemp.ID = location[strings.LastIndex(location, "/")+1:]

		listBase = append(listBase, &server.ItemResponse{
			ID:          graphTemp.ID,
			Name:        graphTemp.Name,
			Description: graphTemp.Description,
		})
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?type=template", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?type=template&limit=1", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[:1], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[:1], listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?type=template&offset=1&limit=2", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[1:3], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[1:3], listResult)
		test.Fail()
	}

	// Delete created items
	for _, listItem := range listBase {
		response = execTestRequest(test, "DELETE", baseURL+listItem.ID, nil, nil)
	}
}

func Test_LibraryCollectionHandle(test *testing.T) {
	var (
		listBase       server.ItemListResponse
		listResult     server.ItemListResponse
		collectionBase struct {
			*library.Collection
			Parent string `json:"parent"`
		}
	)

	baseURL := fmt.Sprintf("http://%s/api/v1/library/collections/", serverConfig.BindAddr)

	// Define a sample collection
	collectionBase.Collection = &library.Collection{Item: library.Item{Name: "collection0",
		Description: "A great collection description."}}

	collectionBase.Entries = append(collectionBase.Entries,
		&library.CollectionEntry{ID: "00000000-0000-0000-0000-000000000000",
			Options: map[string]interface{}{"range": "-1h"}})
	collectionBase.Entries = append(collectionBase.Entries,
		&library.CollectionEntry{ID: "00000000-0000-0000-0000-000000000000",
			Options: map[string]interface{}{"range": "-1d"}})
	collectionBase.Entries = append(collectionBase.Entries,
		&library.CollectionEntry{ID: "00000000-0000-0000-0000-000000000000",
			Options: map[string]interface{}{"range": "-1w"}})

	// Test GET on collections list
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown collection item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into collection
	data, _ := json.Marshal(collectionBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	collectionBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test GET on collection item
	collectionResult := &library.Collection{}

	response = execTestRequest(test, "GET", baseURL+collectionBase.ID, nil, &collectionResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(collectionBase.Collection, collectionResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", collectionBase.Collection, collectionResult)
		test.Fail()
	}

	// Test GET on collections list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:          collectionBase.ID,
		Name:        collectionBase.Name,
		Description: collectionBase.Description,
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on collection item
	collectionBase.Name = "collection0-updated"

	data, _ = json.Marshal(collectionBase.Collection)

	response = execTestRequest(test, "PUT", baseURL+collectionBase.ID, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test GET on collection item
	collectionResult = &library.Collection{}

	response = execTestRequest(test, "GET", baseURL+collectionBase.ID, nil, &collectionResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(collectionBase.Collection, collectionResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", collectionBase, collectionResult)
		test.Fail()
	}

	// Test DELETE on collection item
	response = execTestRequest(test, "DELETE", baseURL+collectionBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+collectionBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test GET on collections list (offset and limit)
	listBase = server.ItemListResponse{}

	for i := 0; i < 3; i++ {
		collectionTemp := &library.Collection{}
		utils.Clone(collectionBase, collectionTemp)

		collectionTemp.ID = ""
		collectionTemp.Name = fmt.Sprintf("collection0-%d", i)

		data, _ = json.Marshal(collectionTemp)

		response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

		if response.StatusCode != http.StatusCreated {
			test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
			test.Fail()
		}

		location := response.Header.Get("Location")

		if location == "" {
			test.Logf("\nExpected `Location' header")
			test.Fail()
		}

		collectionTemp.ID = location[strings.LastIndex(location, "/")+1:]

		listBase = append(listBase, &server.ItemResponse{
			ID:          collectionTemp.ID,
			Name:        collectionTemp.Name,
			Description: collectionTemp.Description,
		})
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?limit=1", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[:1], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[:1], listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?offset=1&limit=2", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[1:3], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[1:3], listResult)
		test.Fail()
	}
}

func execGroupHandle(test *testing.T, urlPrefix string, groupBase *library.Group, expandData,
	expandBase server.ExpandRequest) {

	var (
		listBase     server.ItemListResponse
		listResult   server.ItemListResponse
		expandResult []server.ExpandRequest
	)

	baseURL := fmt.Sprintf("http://%s/api/v1/library/%s/", serverConfig.BindAddr, urlPrefix)

	// Test GET on groups list
	listBase = server.ItemListResponse{}
	listResult = server.ItemListResponse{}

	response := execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test GET on a unknown group item
	response = execTestRequest(test, "GET", baseURL+"/00000000-0000-0000-0000-000000000000", nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test POST into group
	data, _ := json.Marshal(groupBase)

	response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusCreated {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
		test.Fail()
	}

	if response.Header.Get("Location") == "" {
		test.Logf("\nExpected `Location' header")
		test.Fail()
	}

	groupBase.ID = response.Header.Get("Location")[strings.LastIndex(response.Header.Get("Location"), "/")+1:]

	// Test GET on group item
	groupResult := &library.Group{}

	response = execTestRequest(test, "GET", baseURL+groupBase.ID, nil, &groupResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	if !reflect.DeepEqual(groupBase, groupResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", groupBase, groupResult)
		test.Fail()
	}

	// Test GET on groups list
	listBase = server.ItemListResponse{&server.ItemResponse{
		ID:          groupBase.ID,
		Name:        groupBase.Name,
		Description: groupBase.Description,
	}}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	// Test PUT on group item
	groupBase.Name = "group0-updated"

	data, _ = json.Marshal(groupBase)

	response = execTestRequest(test, "PUT", baseURL+groupBase.ID, strings.NewReader(string(data)), nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	// Test GET on group item
	groupResult = &library.Group{}

	response = execTestRequest(test, "GET", baseURL+groupBase.ID, nil, &groupResult)

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

	response = execTestRequest(test, "POST", fmt.Sprintf("http://%s/api/v1/library/expand", serverConfig.BindAddr),
		strings.NewReader(string(data)), &expandResult)

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
	response = execTestRequest(test, "DELETE", baseURL+groupBase.ID, nil, nil)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	response = execTestRequest(test, "DELETE", baseURL+groupBase.ID, nil, nil)

	if response.StatusCode != http.StatusNotFound {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusNotFound, response.StatusCode)
		test.Fail()
	}

	// Test GET on groups list (offset and limit)
	listBase = server.ItemListResponse{}

	for i := 0; i < 3; i++ {
		groupTemp := &library.Group{}
		utils.Clone(groupBase, groupTemp)

		groupTemp.ID = ""
		groupTemp.Name = fmt.Sprintf("group0-%d", i)

		data, _ = json.Marshal(groupTemp)

		response = execTestRequest(test, "POST", baseURL, strings.NewReader(string(data)), nil)

		if response.StatusCode != http.StatusCreated {
			test.Logf("\nExpected %d\nbut got  %d", http.StatusCreated, response.StatusCode)
			test.Fail()
		}

		location := response.Header.Get("Location")

		if location == "" {
			test.Logf("\nExpected `Location' header")
			test.Fail()
		}

		groupTemp.ID = location[strings.LastIndex(location, "/")+1:]

		listBase = append(listBase, &server.ItemResponse{
			ID:          groupTemp.ID,
			Name:        groupTemp.Name,
			Description: groupTemp.Description,
		})
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL, nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase, listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase, listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?limit=1", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[:1], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[:1], listResult)
		test.Fail()
	}

	listResult = server.ItemListResponse{}

	response = execTestRequest(test, "GET", baseURL+"?offset=1&limit=2", nil, &listResult)

	if response.StatusCode != http.StatusOK {
		test.Logf("\nExpected %d\nbut got  %d", http.StatusOK, response.StatusCode)
		test.Fail()
	}

	for _, listItem := range listResult {
		listItem.Modified = ""
	}

	if !reflect.DeepEqual(listBase[1:3], listResult) {
		test.Logf("\nExpected %#v\nbut got  %#v", listBase[1:3], listResult)
		test.Fail()
	}
}

func execTestRequest(test *testing.T, method, url string, data io.Reader, result interface{}) *http.Response {
	request, err := http.NewRequest(method, url, data)
	if err != nil {
		test.Fatal(err.Error())
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

	// Create temporary sample RRD files
	rrdDir := path.Join("tests", "rrd")

	os.MkdirAll(path.Join(rrdDir, "source1"), 0755)
	os.MkdirAll(path.Join(rrdDir, "source2"), 0755)

	creator := rrd.NewCreator(path.Join(rrdDir, "source1", "database1.rrd"), time.Now(), 1)
	creator.RRA("AVERAGE", 0.5, 1, 100)
	creator.DS("test", "COUNTER", 2, 0, 100)

	err = creator.Create(true)
	if err != nil {
		log.Fatalln(err)
	}

	os.Link(path.Join(rrdDir, "source1", "database1.rrd"), path.Join(rrdDir, "source1", "database2.rrd"))
	os.Link(path.Join(rrdDir, "source1", "database1.rrd"), path.Join(rrdDir, "source2", "database2.rrd"))
	os.Link(path.Join(rrdDir, "source1", "database1.rrd"), path.Join(rrdDir, "source2", "database3.rrd"))

	// Refresh server to take newly created RRD files into account
	data, err := ioutil.ReadFile(serverConfig.PidFile)
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := strconv.Atoi(strings.Trim(string(data), "\n"))
	if err != nil {
		log.Fatalln(err)
	}

	syscall.Kill(pid, syscall.SIGUSR1)

	// Wait few seconds for the refresh to be completed
	time.Sleep(3 * time.Second)
}
