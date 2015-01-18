// +build facette

package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/plot"
	"github.com/facette/facette/pkg/utils"
)

const (
	facetteURLCatalog            string  = "/api/v1/catalog/"
	facetteURLLibraryGraphsPlots string  = "/api/v1/library/graphs/plots"
	facetteDefaultTimeout        float64 = 10
)

type facettePlotRequest struct {
	Time        time.Time     `json:"time"`
	Range       string        `json:"range"`
	Sample      int           `json:"sample"`
	Percentiles []float64     `json:"percentiles"`
	Graph       library.Graph `json:"graph"`
}

type facettePlotResponse struct {
	ID          string           `json:"id"`
	Start       string           `json:"start"`
	End         string           `json:"end"`
	Step        float64          `json:"step"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        int              `json:"type"`
	StackMode   int              `json:"stack_mode"`
	Series      []*facetteSeries `json:"series"`
	Modified    time.Time        `json:"modified"`
}

type facetteSeries struct {
	Name    string                 `json:"name"`
	StackID int                    `json:"stack_id"`
	Plots   []plot.Plot            `json:"plots"`
	Summary map[string]plot.Value  `json:"summary"`
	Options map[string]interface{} `json:"options"`
}

// FacetteConnector represents the main structure of the Facette connector.
type FacetteConnector struct {
	name     string
	upstream string
	timeout  float64
}

func init() {
	Connectors["facette"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var err error

		connector := &FacetteConnector{name: name}

		if connector.upstream, err = config.GetString(settings, "upstream", true); err != nil {
			return nil, err
		}

		if connector.timeout, err = config.GetFloat(settings, "timeout", false); err != nil {
			return nil, err
		}

		if connector.timeout <= 0 {
			connector.timeout = facetteDefaultTimeout
		}

		return connector, nil
	}
}

// GetName returns the name of the current connector.
func (connector *FacetteConnector) GetName() string {
	return connector.name
}

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (connector *FacetteConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	var resultSeries []plot.Series

	// Convert plotQuery into plotRequest-like to forward query to upstream Facette API
	plotRequest := facettePlotRequest{
		Time:   query.StartTime,
		Range:  utils.DurationToRange(query.EndTime.Sub(query.StartTime)),
		Sample: query.Sample,
		Graph: library.Graph{
			Item: library.Item{
				Name: "facette",
			},
			Groups: []*library.OperGroup{
				&library.OperGroup{
					Name: "group0",
					Series: func(series []plot.QuerySeries) []*library.Series {
						requestSeries := make([]*library.Series, len(series))

						for index, entry := range series {
							requestSeries[index] = &library.Series{
								Name:   fmt.Sprintf("series%d", index),
								Origin: entry.Origin,
								Source: entry.Source,
								Metric: entry.Metric,
							}
						}

						return requestSeries
					}(query.Series),
				},
			},
		},
	}

	requestBody, err := json.Marshal(plotRequest)
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to marshal plot request: %s", connector.name, err)
	}

	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			// Enable dual IPv4/IPv6 stack connectivity:
			DualStack: true,
			// Enforce HTTP connection timeout:
			Timeout: time.Duration(connector.timeout) * time.Second,
		}).Dial,
	}

	httpClient := http.Client{Transport: httpTransport}

	request, err := http.NewRequest(
		"POST",
		strings.TrimSuffix(connector.upstream, "/")+facetteURLLibraryGraphsPlots,
		bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "FacetteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to perform HTTP request: %s", connector.name, err)
	}

	if err := facetteCheckConnectorResponse(response); err != nil {
		return nil, fmt.Errorf("facette[%s]: invalid upstream HTTP response: %s", connector.name, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to read HTTP response body: %s", connector.name, err)
	}
	defer response.Body.Close()

	plotResponse := facettePlotResponse{}

	if err := json.Unmarshal(data, &plotResponse); err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to unmarshal upstream response: %s", connector.name, err)
	}

	for _, series := range plotResponse.Series {
		resultSeries = append(resultSeries, plot.Series{
			Plots:   series.Plots,
			Summary: series.Summary,
		})
	}

	return resultSeries, nil
}

// Refresh triggers a full connector data update.
func (connector *FacetteConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			// Enable dual IPv4/IPv6 stack connectivity:
			DualStack: true,
			// Enforce HTTP connection timeout:
			Timeout: time.Duration(connector.timeout) * time.Second,
		}).Dial,
	}

	httpClient := http.Client{Transport: httpTransport}

	request, err := http.NewRequest("GET", strings.TrimSuffix(connector.upstream, "/")+facetteURLCatalog, nil)
	if err != nil {
		return fmt.Errorf("facette[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "FacetteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("facette[%s]: unable to perform HTTP request: %s", connector.name, err)
	}
	defer response.Body.Close()

	if err = facetteCheckConnectorResponse(response); err != nil {
		return fmt.Errorf("facette[%s]: invalid HTTP backend response: %s", connector.name, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("facette[%s]: unable to read HTTP response body: %s", connector.name, err)
	}

	upstreamCatalog := make(map[string]map[string][]string)
	if err = json.Unmarshal(data, &upstreamCatalog); err != nil {
		return fmt.Errorf("facette[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	// Parse the upstream catalog entries and append them to our local catalog
	for upstreamOriginName, upstreamOrigin := range upstreamCatalog {
		for sourceName, metrics := range upstreamOrigin {
			for _, metric := range metrics {
				outputChan <- &catalog.Record{
					Origin:    upstreamOriginName,
					Source:    sourceName,
					Metric:    metric,
					Connector: connector,
				}
			}
		}
	}

	return nil
}

func facetteCheckConnectorResponse(response *http.Response) error {
	if response.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", response.StatusCode)
	}

	if !strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", response.Header["Content-Type"])
	}

	return nil
}
