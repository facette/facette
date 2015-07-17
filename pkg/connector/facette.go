// +build facette

package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	facetteDefaultTimeout int    = 10
	facetteURLCatalog     string = "/api/v1/catalog/"
	facetteURLPlots       string = "/api/v1/plots"
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
	name          string
	upstream      string
	timeout       int
	insecureTLS   bool
	serverID      string
	httpTransport *http.Transport
}

func init() {
	Connectors["facette"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var err error

		c := &FacetteConnector{name: name}

		if c.upstream, err = config.GetString(settings, "upstream", true); err != nil {
			return nil, err
		}

		if c.timeout, err = config.GetInt(settings, "timeout", false); err != nil {
			return nil, err
		}
		if c.timeout <= 0 {
			c.timeout = facetteDefaultTimeout
		}

		if c.insecureTLS, err = config.GetBool(settings, "allow_insecure_tls", false); err != nil {
			return nil, err
		}

		if c.serverID, err = config.GetString(settings, "_id", true); err != nil {
			return nil, err
		}

		return c, nil
	}
}

// GetName returns the name of the current connector.
func (c *FacetteConnector) GetName() string {
	return c.name
}

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (c *FacetteConnector) GetPlots(query *plot.Query) ([]*plot.Series, error) {
	var results []*plot.Series

	// Convert plotQuery into plotRequest-like to forward query to upstream Facette API
	plotReq := facettePlotRequest{
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
						out := make([]*library.Series, len(series))

						for i, s := range series {
							out[i] = &library.Series{
								Name:   fmt.Sprintf("series%d", i),
								Origin: s.Origin,
								Source: s.Source,
								Metric: s.Metric,
							}
						}

						return out
					}(query.Series),
				},
			},
		},
	}

	body, err := json.Marshal(plotReq)
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to marshal plot request: %s", c.name, err)
	}

	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	r, err := http.NewRequest("POST", strings.TrimSuffix(c.upstream, "/")+facetteURLPlots, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "FacetteConnector")

	if query.Requestor != "" {
		r.Header.Add("X-Facette-Requestor", query.Requestor)
	} else {
		r.Header.Add("X-Facette-Requestor", c.serverID)
	}

	rsp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to perform HTTP request: %s", c.name, err)
	}

	if err := facetteCheckConnectorResponse(rsp); err != nil {
		return nil, fmt.Errorf("facette[%s]: invalid upstream HTTP response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to read HTTP response body: %s", c.name, err)
	}
	defer rsp.Body.Close()

	plotRsp := facettePlotResponse{}

	if err := json.Unmarshal(data, &plotRsp); err != nil {
		return nil, fmt.Errorf("facette[%s]: unable to unmarshal upstream response: %s", c.name, err)
	}

	for _, s := range plotRsp.Series {
		results = append(results, &plot.Series{
			Plots:   s.Plots,
			Summary: s.Summary,
		})
	}

	return results, nil
}

// Refresh triggers a full connector data update.
func (c *FacetteConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	r, err := http.NewRequest("GET", strings.TrimSuffix(c.upstream, "/")+facetteURLCatalog, nil)
	if err != nil {
		return fmt.Errorf("facette[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "FacetteConnector")

	rsp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("facette[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	if err = facetteCheckConnectorResponse(rsp); err != nil {
		return fmt.Errorf("facette[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("facette[%s]: unable to read HTTP response body: %s", c.name, err)
	}

	upstreamCatalog := make(map[string]map[string][]string)
	if err = json.Unmarshal(data, &upstreamCatalog); err != nil {
		return fmt.Errorf("facette[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	// Parse the upstream catalog entries and append them to our local catalog
	for upstreamOriginName, upstreamOrigin := range upstreamCatalog {
		for sourceName, metrics := range upstreamOrigin {
			for _, metric := range metrics {
				outputChan <- &catalog.Record{
					Origin:    upstreamOriginName,
					Source:    sourceName,
					Metric:    metric,
					Connector: c,
				}
			}
		}
	}

	return nil
}

func facetteCheckConnectorResponse(r *http.Response) error {
	if r.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", r.StatusCode)
	}

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", r.Header["Content-Type"])
	}

	return nil
}
