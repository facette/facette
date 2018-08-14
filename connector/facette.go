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

func init() {
	connectors["facette"] = func(name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
		var err error

		c := &facetteConnector{
			name: name,
		}

		// Get connector handler settings
		c.url, err = settings.GetString("url", "")
		if err != nil {
			return nil, err
		} else if c.url == "" {
			return nil, ErrMissingConnectorSetting("url")
		}
		c.url = normalizeURL(c.url)

		c.timeout, err = settings.GetInt("timeout", defaultTimeout)
		if err != nil {
			return nil, err
		}

		c.allowInsecure, err = settings.GetBool("allow_insecure_tls", false)
		if err != nil {
			return nil, err
		}

		// Check remote instance URL
		_, err = url.Parse(c.url)
		if err != nil {
			return nil, fmt.Errorf("unable to parse URL: %s", err)
		}

		c.client = httputil.NewClient(time.Duration(c.timeout)*time.Second, true, c.allowInsecure)

		return c, nil
	}
}

type facetteConnector struct {
	name          string
	url           string
	timeout       int
	allowInsecure bool
	client        *http.Client
}

func (c *facetteConnector) Name() string {
	return c.name
}

func (c *facetteConnector) Points(query *series.Query) ([]series.Series, error) {
	// Convert query into a Facette point request
	body, err := json.Marshal(series.Request{
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		Sample:    query.Sample,
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
					}(query.Series),
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
	req.Header.Add("User-Agent", "facette/"+version.Version)
	req.Header.Add("Content-Type", "application/json")

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
