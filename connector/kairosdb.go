// +build !disable_connector_kairosdb

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
	"facette.io/facette/set"
	"facette.io/facette/version"
	"facette.io/httputil"
	"facette.io/logger"
	"facette.io/maputil"
)

const (
	kairosDBURLDatapointsQuery = "/api/v1/datapoints/query"
	kairosDBURLMetricNames     = "/api/v1/metricnames"
)

var (
	kairosDBDefaultSourceTags = []string{
		"host",
		"server",
		"device",
	}

	kairosDBDefaultAggregators = []string{
		"avg",
		"max",
		"min",
	}
)

func init() {
	connectors["kairosdb"] = func(name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
		var err error

		c := &kairosDBConnector{
			name:    name,
			metrics: make(map[string]map[string]*kairosDBMetric),
		}

		// Get connector handler settings
		c.url, err = settings.GetString("url", "")
		if err != nil {
			return nil, err
		} else if c.url == "" {
			return nil, ErrMissingConnectorSetting("url")
		}
		c.url = normalizeURL(c.url)

		c.aggregators, err = settings.GetStringSlice("aggregators", kairosDBDefaultAggregators)
		if err != nil {
			return nil, err
		}

		c.sourceTags, err = settings.GetStringSlice("source_tags", kairosDBDefaultSourceTags)
		if err != nil {
			return nil, err
		}

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

type kairosDBConnector struct {
	name          string
	url           string
	aggregators   []string
	sourceTags    []string
	timeout       int
	allowInsecure bool
	client        *http.Client
	metrics       map[string]map[string]*kairosDBMetric
}

func (c *kairosDBConnector) Name() string {
	return c.name
}

func (c *kairosDBConnector) Refresh(output chan<- *catalog.Record) error {
	// Prepare source tags set (used for tags filtering)
	tags := set.New()
	for _, t := range c.sourceTags {
		tags.Add(t)
	}

	req, err := http.NewRequest("GET", c.url+kairosDBURLMetricNames, nil)
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}
	req.Header.Add("User-Agent", "facette/"+version.Version)

	// Retrieve metrics list
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	mr := kairosDBMetricResponse{}
	if err = httputil.BindJSON(resp, &mr); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	// Retrieve metrics associated tags
	q := kairosDBQuery{Metrics: []kairosDBQueryMetric{}}
	for _, metric := range mr.Results {
		q.Metrics = append(q.Metrics, kairosDBQueryMetric{Name: metric})
	}

	body, err := json.Marshal(q)
	if err != nil {
		return fmt.Errorf("unable to marshal tags request: %s", err)
	}

	req, err = http.NewRequest("POST", c.url+kairosDBURLDatapointsQuery+"/tags", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}
	req.Header.Add("User-Agent", "facette/"+version.Version)
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	r := kairosDBResponse{}
	if err := httputil.BindJSON(resp, &r); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for _, q := range r.Queries {
		for _, r := range q.Results {
			for key, values := range r.Tags {
				if !tags.Has(key) {
					continue
				}

				for _, aggr := range c.aggregators {
					metric := r.Name + "/" + aggr

					for _, value := range values {
						if _, ok := c.metrics[value]; !ok {
							c.metrics[value] = make(map[string]*kairosDBMetric)
						}

						c.metrics[value][metric] = &kairosDBMetric{
							Name:       r.Name,
							Aggregator: aggr,
							Tag:        [2]string{key, value},
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

func (c *kairosDBConnector) Points(query *series.Query) ([]series.Series, error) {
	step := query.EndTime.Sub(query.StartTime) / time.Duration(query.Sample) // FIXME: use q.Step()
	sampling := step.Nanoseconds() / 1000000

	q := kairosDBQuery{
		StartAbsolute: query.StartTime.Unix() * 1000,
		EndAbsolute:   query.EndTime.Unix() * 1000,
		Metrics:       []kairosDBQueryMetric{},
	}

	for _, s := range query.Series {
		if _, ok := c.metrics[s.Source]; !ok {
			return nil, ErrUnknownSource
		} else if _, ok := c.metrics[s.Source][s.Metric]; !ok {
			return nil, ErrUnknownMetric
		}

		q.Metrics = append(q.Metrics, kairosDBQueryMetric{
			Name: c.metrics[s.Source][s.Metric].Name,
			Tags: map[string]interface{}{
				c.metrics[s.Source][s.Metric].Tag[0]: []string{c.metrics[s.Source][s.Metric].Tag[1]},
			},
			Aggregators: []kairosDBQueryAggregator{{
				Name: c.metrics[s.Source][s.Metric].Aggregator,
				Sampling: kairosDBQuerySampling{
					Value: sampling,
					Unit:  "milliseconds",
				}},
			},
		})
	}

	body, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal tags request: %s", err)
	}

	req, err := http.NewRequest("POST", c.url+kairosDBURLDatapointsQuery, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to set up HTTP request: %s", err)
	}
	req.Header.Add("User-Agent", "facette/"+version.Version)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	pr := kairosDBResponse{}
	if err := httputil.BindJSON(resp, &pr); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	result := []series.Series{}
	for _, q := range pr.Queries {
		s := series.Series{}
		for _, value := range q.Results[0].Values {
			s.Points = append(s.Points, series.Point{
				Time:  time.Unix(int64(value[0]/1000), 0),
				Value: series.Value(value[1]),
			})
		}

		result = append(result, s)
	}

	return result, nil
}

type kairosDBMetric struct {
	Name       string
	Aggregator string
	Tag        [2]string
}

type kairosDBQuery struct {
	StartAbsolute int64                 `json:"start_absolute"`
	EndAbsolute   int64                 `json:"end_absolute,omitempty"`
	Metrics       []kairosDBQueryMetric `json:"metrics"`
}

type kairosDBQueryMetric struct {
	Name        string                    `json:"name"`
	Tags        map[string]interface{}    `json:"tags,omitempty"`
	Aggregators []kairosDBQueryAggregator `json:"aggregators,omitempty"`
}

type kairosDBQueryAggregator struct {
	Name     string                `json:"name"`
	Sampling kairosDBQuerySampling `json:"sampling"`
}

type kairosDBQuerySampling struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit"`
}

type kairosDBResponse struct {
	Queries []kairosDBResponseQuery `json:"queries"`
}

type kairosDBResponseQuery struct {
	Results []kairosDBResponseResult `json:"results"`
}

type kairosDBResponseResult struct {
	Name   string              `json:"name"`
	Tags   map[string][]string `json:"tags"`
	Values [][2]float64        `json:"values"`
}

type kairosDBMetricResponse struct {
	Results []string `json:"results"`
}
