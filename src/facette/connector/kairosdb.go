// +build !disable_connector_kairosdb

package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"facette/catalog"
	"facette/plot"

	"github.com/facette/httputil"
	"github.com/facette/logger"
	"github.com/facette/maputil"
	"github.com/fatih/set"
)

const (
	kairosdbURLMetricNames = "/api/v1/metricnames"
	kairosdbURLMetricTags  = "/api/v1/datapoints/query/tags"
	kairosdbURLMetric      = "/api/v1/datapoints/query"
)

var (
	kairosdbDefaultSourceTags = []string{
		"host",
		"server",
		"device",
	}

	kairosdbDefaultAggregators = []string{
		"avg",
		"max",
		"min",
	}
)

type kairosdbQuerySampling struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit"`
}

type kairosdbQueryAggregator struct {
	Name     string                `json:"name"`
	Sampling kairosdbQuerySampling `json:"sampling"`
}

type kairosdbQueryMetric struct {
	Name        string                    `json:"name"`
	Tags        map[string]interface{}    `json:"tags,omitempty"`
	Aggregators []kairosdbQueryAggregator `json:"aggregators,omitempty"`
}

type kairosdbMetricResponse struct {
	Results []string `json:"results"`
}

type kairosdbResponseResult struct {
	Name   string              `json:"name"`
	Tags   map[string][]string `json:"tags"`
	Values [][2]float64        `json:"values"`
}

type kairosdbResponseEntry struct {
	Results []kairosdbResponseResult `json:"results"`
}

type kairosdbResponse struct {
	Queries []kairosdbResponseEntry `json:"queries"`
}

type kairosdbMetric struct {
	metric     string
	aggregator string
	tag        [2]string
}

type kairosdbQuery struct {
	StartAbsolute int64                 `json:"start_absolute"`
	EndAbsolute   int64                 `json:"end_absolute,omitempty"`
	Metrics       []kairosdbQueryMetric `json:"metrics"`
}

// kairosdbConnector implements the connector handler for another KairosDB instance.
type kairosdbConnector struct {
	name          string
	url           string
	aggregators   []string
	sourceTags    []string
	timeout       int
	allowInsecure bool
	client        *http.Client
	metrics       map[string]map[string]*kairosdbMetric
}

func init() {
	connectors["kairosdb"] = func(name string, settings *maputil.Map, log *logger.Logger) (Connector, error) {
		var err error

		c := &kairosdbConnector{
			name:    name,
			metrics: make(map[string]map[string]*kairosdbMetric),
		}

		// Get connector handler settings
		if c.url, err = settings.GetString("url", ""); err != nil {
			return nil, err
		} else if c.url == "" {
			return nil, ErrMissingConnectorSetting("url")
		}
		normalizeURL(&c.url)

		if c.aggregators, err = settings.GetStringSlice("aggregators", kairosdbDefaultAggregators); err != nil {
			return nil, err
		}

		if c.sourceTags, err = settings.GetStringSlice("source_tags", kairosdbDefaultSourceTags); err != nil {
			return nil, err
		}

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
func (c *kairosdbConnector) Name() string {
	return c.name
}

// Refresh triggers the connector data refresh.
func (c *kairosdbConnector) Refresh(output chan<- *catalog.Record) error {
	// Prepare source tags set (used for tags filtering)
	tags := set.New()
	for _, t := range c.sourceTags {
		tags.Add(t)
	}

	req, err := http.NewRequest("GET", c.url+kairosdbURLMetricNames, nil)
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	// Retrieve metrics list
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	mr := kairosdbMetricResponse{}
	if err := httputil.BindJSON(resp, &mr); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	// Retrieve metrics associated tags
	tq := kairosdbQuery{Metrics: []kairosdbQueryMetric{}}
	for _, metric := range mr.Results {
		tq.Metrics = append(tq.Metrics, kairosdbQueryMetric{Name: metric})
	}

	body, err := json.Marshal(tq)
	if err != nil {
		return fmt.Errorf("unable to marshal tags request: %s", err)
	}

	req, err = http.NewRequest("POST", c.url+kairosdbURLMetricTags, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	req.Header.Add("User-Agent", "facette/"+version)
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	tr := kairosdbResponse{}
	if err := httputil.BindJSON(resp, &tr); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for _, q := range tr.Queries {
		for _, r := range q.Results {
			for key, values := range r.Tags {
				if !tags.Has(key) {
					continue
				}

				for _, aggr := range c.aggregators {
					metric := r.Name + "/" + aggr

					for _, value := range values {
						if _, ok := c.metrics[value]; !ok {
							c.metrics[value] = make(map[string]*kairosdbMetric)
						}

						c.metrics[value][metric] = &kairosdbMetric{
							metric:     r.Name,
							aggregator: aggr,
							tag:        [2]string{key, value},
						}

						output <- &catalog.Record{
							Origin:    c.name,
							Source:    value,
							Metric:    metric,
							Connector: c,
						}
					}
				}
			}
		}
	}

	return nil
}

// Plots retrieves the time series data according to the query parameters and a time interval.
func (c *kairosdbConnector) Plots(q *plot.Query) ([]plot.Series, error) {
	step := q.EndTime.Sub(q.StartTime) / time.Duration(q.Sample)
	sampling := step.Nanoseconds() / 1000000

	pq := kairosdbQuery{
		StartAbsolute: q.StartTime.Unix() * 1000,
		EndAbsolute:   q.EndTime.Unix() * 1000,
		Metrics:       []kairosdbQueryMetric{},
	}

	for _, s := range q.Series {
		if _, ok := c.metrics[s.Source]; !ok {
			return nil, ErrUnknownSource
		} else if _, ok := c.metrics[s.Source][s.Metric]; !ok {
			return nil, ErrUnknownMetric
		}

		pq.Metrics = append(pq.Metrics, kairosdbQueryMetric{
			Name: c.metrics[s.Source][s.Metric].metric,
			Tags: map[string]interface{}{
				c.metrics[s.Source][s.Metric].tag[0]: []string{c.metrics[s.Source][s.Metric].tag[1]},
			},
			Aggregators: []kairosdbQueryAggregator{{
				Name: c.metrics[s.Source][s.Metric].aggregator,
				Sampling: kairosdbQuerySampling{
					Value: sampling,
					Unit:  "milliseconds",
				}},
			},
		})
	}

	body, err := json.Marshal(pq)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal tags request: %s", err)
	}

	req, err := http.NewRequest("POST", c.url+kairosdbURLMetric, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	req.Header.Add("User-Agent", "facette/"+version)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	pr := kairosdbResponse{}
	if err := httputil.BindJSON(resp, &pr); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	result := []plot.Series{}
	for _, q := range pr.Queries {
		s := plot.Series{}
		for _, value := range q.Results[0].Values {
			s.Plots = append(s.Plots, plot.Plot{
				Time:  time.Unix(int64(value[0]/1000), 0),
				Value: plot.Value(value[1]),
			})
		}

		result = append(result, s)
	}

	return result, nil
}
