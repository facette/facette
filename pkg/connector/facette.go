package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/types"
)

const (
	facetteURLCatalog            string = "/api/v1/catalog/"
	facetteURLLibraryGraphsPlots string = "/api/v1/library/graphs/plots"
)

// FacetteConnector represents the main structure of the Facette connector.
type FacetteConnector struct {
	upstream string
}

func init() {
	Connectors["facette"] = func(settings map[string]interface{}) (Connector, error) {
		var err error

		connector := &FacetteConnector{}

		if connector.upstream, err = config.GetString(settings, "upstream", true); err != nil {
			return nil, err
		}

		return connector, nil
	}
}

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (connector *FacetteConnector) GetPlots(query *types.PlotQuery) (map[string]*types.PlotResult, error) {
	result := make(map[string]*types.PlotResult)

	// Convert plotQuery into plotRequest to forward query to upstream Facette API
	plotRequest := struct {
		Time  time.Time     `json:"time"`
		Graph library.Graph `json:"graph"`
	}{
		Time: query.StartTime,
		Graph: library.Graph{
			Item: library.Item{
				Name: "facette",
			},
			Groups: []*library.OperGroup{
				&library.OperGroup{
					Type: query.Group.Type,
					Series: func(series []*types.SerieQuery) []*library.Serie {
						requestSeries := make([]*library.Serie, len(series))

						for i, serie := range series {
							requestSeries[i] = &library.Serie{
								Name:   serie.Name,
								Origin: serie.Metric.Origin,
								Source: serie.Metric.Source,
								Metric: serie.Metric.Name,
								Scale:  serie.Scale,
							}
						}

						return requestSeries
					}(query.Group.Series),
				},
			},
		},
	}

	requestBody, err := json.Marshal(plotRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal plot request: %s", err)
	}

	log.Printf("DEBUG: facetteConnector: >>> %s", requestBody)

	httpClient := http.Client{}

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
		return nil, err
	}

	if err := facetteCheckConnectorResponse(response); err != nil {
		return nil, fmt.Errorf("invalid upstream HTTP response: %s", err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	log.Printf("DEBUG: facetteConnector: <<< %s", data)

	// result["blah"] = &types.PlotResult{
	// 	Plots: []types.PlotValue{42.0},
	// }

	return result, nil
}

// Refresh triggers a full connector data update.
func (connector *FacetteConnector) Refresh(originName string, outputChan chan *catalog.CatalogRecord) error {
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			// Enable dual IPv4/IPv6 stack connectivity:
			DualStack: true,
		}).Dial,
	}

	httpClient := http.Client{Transport: httpTransport}

	response, err := httpClient.Get(strings.TrimSuffix(connector.upstream, "/") + facetteURLCatalog)
	if err != nil {
		return err
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
