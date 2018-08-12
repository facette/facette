// +build !disable_connector_facette

package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"facette.io/facette/catalog"
	"facette.io/facette/series"
	"facette.io/facette/storage"
	"facette.io/facette/version"
	"facette.io/httputil"
	"facette.io/logger"
	"facette.io/maputil"
)

const (
	facetteURLCatalog = "/api/v1/catalog/"
	facetteURLPoints  = "/api/v1/series/points"
)

// facetteConnector implements the connector handler for another Facette instance.
type facetteConnector struct {
	name          string
	url           string
	timeout       int
	allowInsecure bool
	client        *http.Client
}

func init() {
	connectors["facette"] = func(name string, settings *maputil.Map, log *logger.Logger) (Connector, error) {
		var err error

		c := &facetteConnector{name: name}

		// Get connector handler settings
		if c.url, err = settings.GetString("url", ""); err != nil {
			return nil, err
		} else if c.url == "" {
			return nil, ErrMissingConnectorSetting("url")
		}
		normalizeURL(&c.url)

		if c.timeout, err = settings.GetInt("timeout", connectorDefaultTimeout); err != nil {
			return nil, err
		}

		if c.allowInsecure, err = settings.GetBool("allow_insecure_tls", false); err != nil {
			return nil, err
		}

		// Check remote instance URL
		if _, err := url.Parse(c.url); err != nil {
			return nil, fmt.Errorf("unable to parse URL: %s", err)
		}

		// Create new HTTP client
		c.client = httputil.NewClient(time.Duration(c.timeout)*time.Second, true, c.allowInsecure)

		return c, nil
	}
}

// Name returns the name of the current connector.
func (c *facetteConnector) Name() string {
	return c.name
}

// Refresh triggers the connector data refresh.
func (c *facetteConnector) Refresh(output chan<- *catalog.Record) error {
	// Create new HTTP request
	req, err := http.NewRequest("GET", c.url+facetteURLCatalog, nil)
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	req.Header.Add("User-Agent", "facette/"+version.Version)

	// Retrieve data from upstream catalog
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	result := make(map[string]map[string][]string)
	if err := httputil.BindJSON(resp, &result); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for origin, sources := range result {
		for source, metrics := range sources {
			for _, metric := range metrics {
				output <- &catalog.Record{
					Origin:    origin,
					Source:    source,
					Metric:    metric,
					Connector: c,
				}
			}
		}
	}

	return nil
}

// Points retrieves the time series data according to the query parameters and a time interval.
func (c *facetteConnector) Points(q *series.Query) ([]series.Series, error) {
	// Convert query into a Facette point request
	body, err := json.Marshal(series.Request{
		StartTime: q.StartTime,
		EndTime:   q.EndTime,
		Sample:    q.Sample,
		Graph: &storage.Graph{
			Item: storage.Item{
				Name: "facette",
			},
			Groups: storage.SeriesGroups{
				{
					Series: func(series []series.QuerySeries) []*storage.Series {
						out := make([]*storage.Series, len(series))
						for i, s := range series {
							out[i] = &storage.Series{
								Name:   fmt.Sprintf("series%d", i),
								Origin: s.Origin,
								Source: s.Source,
								Metric: s.Metric,
							}
						}

						return out
					}(q.Series),
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal points request: %s", err)
	}

	// Create new HTTP request
	req, err := http.NewRequest("POST", c.url+facetteURLPoints, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "facette/"+version.Version)

	// Retrieve upstream data points
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	// Fill result with data received from request
	data := series.Response{}
	if err := httputil.BindJSON(resp, &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	result := []series.Series{}
	for _, s := range data.Series {
		result = append(result, series.Series{
			Points:  s.Points,
			Summary: s.Summary,
		})
	}

	return result, nil
}
