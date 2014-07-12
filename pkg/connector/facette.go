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
	"github.com/facette/facette/pkg/types"
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
	ID          string          `json:"id"`
	Start       string          `json:"start"`
	End         string          `json:"end"`
	Step        float64         `json:"step"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        int             `json:"type"`
	StackMode   int             `json:"stack_mode"`
	Series      []*facetteSerie `json:"series"`
	Modified    time.Time       `json:"modified"`
}

type facetteSerie struct {
	Name    string                     `json:"name"`
	StackID int                        `json:"stack_id"`
	Plots   []types.PlotValue          `json:"plots"`
	Info    map[string]types.PlotValue `json:"info"`
	Options map[string]interface{}     `json:"options"`
}

// FacetteConnector represents the main structure of the Facette connector.
type FacetteConnector struct {
	upstream string
	timeout  float64
}

func init() {
	Connectors["facette"] = func(settings map[string]interface{}) (Connector, error) {
		var err error

		connector := &FacetteConnector{}

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

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (connector *FacetteConnector) GetPlots(query *types.PlotQuery) ([]*types.PlotResult, error) {
	// Convert plotQuery into plotRequest-like to forward query to upstream Facette API
	plotRequest := facettePlotRequest{
		Time:  query.StartTime,
		Range: utils.DurationToRange(query.EndTime.Sub(query.StartTime)),
		Graph: library.Graph{
			Item: library.Item{
				Name: "facette",
			},
			Groups: []*library.OperGroup{
				&library.OperGroup{
					Name: "group0",
					Type: query.Group.Type,
					Series: func(series []*types.SerieQuery) []*library.Serie {
						requestSeries := make([]*library.Serie, len(series))

						for index, serie := range series {
							requestSeries[index] = &library.Serie{
								Name:   fmt.Sprintf("serie%d", index),
								Origin: serie.Metric.Origin,
								Source: serie.Metric.Source,
								Metric: serie.Metric.Name,
								Scale:  serie.Scale,
							}
						}

						return requestSeries
					}(query.Group.Series),
					Scale: query.Group.Scale,
				},
			},
		},
	}

	if query.Step != 0 {
		plotRequest.Sample = int((query.EndTime.Sub(query.StartTime) / query.Step).Seconds())
	}

	requestBody, err := json.Marshal(plotRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal plot request: %s", err)
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
		return nil, fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "FacetteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %s", err)
	}

	if err := facetteCheckConnectorResponse(response); err != nil {
		return nil, fmt.Errorf("invalid upstream HTTP response: %s", err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	plotResponse := facettePlotResponse{}

	if err := json.Unmarshal(data, &plotResponse); err != nil {
		return nil, fmt.Errorf("unable to unmarshal upstream response: %s", err)
	}

	result := make([]*types.PlotResult, 0)

	for _, serie := range plotResponse.Series {
		result = append(result, &types.PlotResult{
			Plots: serie.Plots,
			Info:  serie.Info,
		})
	}

	return result, nil
}

// Refresh triggers a full connector data update.
func (connector *FacetteConnector) Refresh(originName string, outputChan chan *catalog.CatalogRecord) error {
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
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "FacetteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}

	if err = facetteCheckConnectorResponse(response); err != nil {
		return fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	upstreamCatalog := make(map[string]map[string][]string)
	if err = json.Unmarshal(data, &upstreamCatalog); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	// Parse the upstream catalog entries and append them to our local catalog
	for upstreamOriginName, upstreamOrigin := range upstreamCatalog {
		for sourceName, metrics := range upstreamOrigin {
			for _, metric := range metrics {
				outputChan <- &catalog.CatalogRecord{
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
