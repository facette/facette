// +build kairosdb

package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/plot"
	"github.com/facette/facette/pkg/utils"
)

const (
	kairosdbDefaultTimeout int    = 10
	kairosdbURLVersion     string = "/api/v1/version"
	kairosdbURLMetricNames string = "/api/v1/metricnames"
	kairosdbURLMetricTags  string = "/api/v1/datapoints/query/tags"
	kairosdbURLQueryMetric string = "/api/v1/datapoints/query"
)

type kairosdbSeriesEntry struct {
	metric     string
	tag        string
	source     string
	aggregator interface{}
}

type metricAggregator struct {
	pattern string
	re      *regexp.Regexp
	hook    interface{}
}

type metricQueryEntry struct {
	Name        string              `json:"name"`
	Tags        map[string][]string `json:"tags"`
	Aggregators []interface{}       `json:"aggregators,omitempty"`
}

type metricQueryResponse struct {
	SampleSize int64               `json:"sample_size"`
	Results    []metricQueryResult `json:"results"`
}

type metricQueryResult struct {
	Name    string              `json:"name"`
	GroupBy []map[string]string `json:"group_by"`
	Tags    map[string][]string `json:"tags"`
	Values  [][2]float64        `json:"values"`
}

type plotsQuery struct {
	StartAbsolute int64              `json:"start_absolute"`
	EndAbsolute   int64              `json:"end_absolute"`
	Metrics       []metricQueryEntry `json:"metrics"`
}

// KairosdbConnector represents the main structure of the Kairosdb connector.
type KairosdbConnector struct {
	name              string
	url               string
	timeout           int
	insecureTLS       bool
	sourceTags        []string
	startAbsolute     int
	startRelative     interface{}
	endAbsolute       int
	endRelative       interface{}
	defaultAggregator interface{}
	aggregators       []metricAggregator
	series            map[string]map[string]kairosdbSeriesEntry
}

func init() {
	Connectors["kairosdb"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var (
			aggregators interface{}
			err         error
		)

		c := &KairosdbConnector{
			name:    name,
			timeout: kairosdbDefaultTimeout,
			series:  make(map[string]map[string]kairosdbSeriesEntry),
		}

		if c.url, err = config.GetString(settings, "url", true); err != nil {
			return nil, err
		}

		if c.timeout, err = config.GetInt(settings, "timeout", false); err != nil {
			return nil, err
		}
		if c.timeout <= 0 {
			c.timeout = kairosdbDefaultTimeout
		}

		if c.insecureTLS, err = config.GetBool(settings, "allow_insecure_tls", false); err != nil {
			return nil, err
		}

		if c.sourceTags, err = config.GetStringSlice(settings, "source_tags", false); err != nil {
			return nil, err
		}

		if c.startAbsolute, err = config.GetInt(settings, "start_absolute", false); err != nil {
			return nil, err
		}
		if c.endAbsolute, err = config.GetInt(settings, "end_absolute", false); err != nil {
			return nil, err
		}

		if c.startRelative, err = config.GetJsonObj(settings, "start_relative", false); err != nil {
			return nil, err
		}
		if c.endRelative, err = config.GetJsonObj(settings, "end_relative", false); err != nil {
			return nil, err
		}

		if c.defaultAggregator, err = config.GetJsonObj(settings, "default_aggregator", false); err != nil {
			return nil, err
		}

		if aggregators, err = config.GetJsonArray(settings, "aggregators", false); err != nil {
			return nil, err
		}

		c.aggregators = compileAggregatorPatterns(aggregators, c.name)

		if c.startAbsolute > 0 && c.startRelative != nil {
			return nil, fmt.Errorf("kairosdb[%s]: start_absolute/start_relative are mutually exclusive", c.name)
		}
		if c.endAbsolute > 0 && c.endRelative != nil {
			return nil, fmt.Errorf("kairosdb[%s]: end_absolute/end_relative are mutually exclusive", c.name)
		}

		// Enforce startRelative defaults
		if c.startAbsolute <= 0 && c.startRelative == nil {
			c.startRelative = map[string]interface{}{"value": 3, "unit": "months"}
		}

		// Enforce sourceTags defaults
		if c.sourceTags == nil {
			c.sourceTags = []string{"host", "server", "device"}
		}

		version, version_array, err := kairosdbGetVersion(c)
		if err != nil {
			return nil, fmt.Errorf("kairosdb[%s]: unable to get KairosDB version: %s", c.name, err)
		}

		if version_array[0] < 1 {
			if version_array[1] <= 9 {
				if version_array[1] < 9 || (version_array[1] == 9 && version_array[2] < 4) {
					return nil, fmt.Errorf(
						"kairosdb[%s]: only KairosDB version 0.9.4 and greater are supported (%s detected)",
						c.name,
						version,
					)
				}
			}
		}

		return c, nil
	}
}

// GetName returns the name of the current connector.
func (c *KairosdbConnector) GetName() string {
	return c.name
}

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (c *KairosdbConnector) GetPlots(query *plot.Query) ([]*plot.Series, error) {
	var (
		jsonResponse map[string][]metricQueryResponse
		results      []*plot.Series
	)

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("kairosdb[%s]: requested series list is empty", c.name)
	}

	jsonQuery, err := kairosdbBuildJSONQuery(query, c.series)
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to build or marshal JSON query: %s", c.name, err)
	}

	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: API Call to %s: %s", c.name,
		strings.TrimSuffix(c.url, "/")+kairosdbURLQueryMetric,
		string(jsonQuery))

	r, err := http.NewRequest("POST", strings.TrimSuffix(c.url, "/")+kairosdbURLQueryMetric, bytes.NewBuffer(jsonQuery))
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "KairosDBConnector")
	r.Header.Set("Content-Type", "application/json")

	rsp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	if err = kairosdbCheckBackendResponse(rsp); err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to read HTTP response body: %s", c.name, err)
	}

	if err = json.Unmarshal(data, &jsonResponse); err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	if results, err = kairosdbExtractPlots(query, c.series, jsonResponse["queries"]); err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to extract plot values from backend response: %s", c.name, err)
	}

	return results, nil
}

// Refresh triggers a full connector data update.
func (c *KairosdbConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	var (
		jsonMetrics map[string][]string
		jsonQuery   map[string][]map[string][]struct {
			Name string              `json:"name"`
			Tags map[string][]string `json:"tags"`
		}
	)

	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	r, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+kairosdbURLMetricNames, nil)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "KairosDBConnector")

	rsp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	if err = kairosdbCheckBackendResponse(rsp); err != nil {
		return fmt.Errorf("kairosdb[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to read HTTP response body: %s", c.name, err)
	}
	if err = json.Unmarshal(data, &jsonMetrics); err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	metrics := make([]map[string]string, 0)
	for _, m := range jsonMetrics["results"] {
		metrics = append(metrics, map[string]string{"name": m})
	}

	query := map[string]interface{}{"metrics": metrics}

	if c.startAbsolute > 0 {
		query["start_absolute"] = c.startAbsolute
	} else {
		query["start_relative"] = c.startRelative
	}

	if c.endAbsolute > 0 {
		query["end_absolute"] = c.endAbsolute
	}

	if c.endRelative != nil {
		query["end_relative"] = c.endRelative
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to marshal JSON data: %s", c.name, err)
	}

	logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: API Call to %s: %s", c.name,
		strings.TrimSuffix(c.url, "/")+kairosdbURLMetricTags, string(jsonData))

	r, err = http.NewRequest("POST", strings.TrimSuffix(c.url, "/")+kairosdbURLMetricTags, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "KairosDBConnector")
	r.Header.Set("Content-Type", "application/json")

	rsp, err = client.Do(r)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	if err = kairosdbCheckBackendResponse(rsp); err != nil {
		return fmt.Errorf("kairosdb[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to read HTTP response body: %s", c.name, err)
	}

	logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: API Response from %s: %s", c.name,
		strings.TrimSuffix(c.url, "/")+kairosdbURLMetricTags, string(data))

	if err = json.Unmarshal(data, &jsonQuery); err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	for _, q := range jsonQuery["queries"] {
		for _, r := range q["results"] {
			metricName := r.Name

			aggregator := matchAggregatorPattern(c.aggregators, metricName)
			if aggregator == nil {
				aggregator = c.defaultAggregator
			}

			for _, t := range c.sourceTags {
				if _, ok := r.Tags[t]; !ok {
					continue
				}

				for _, sourceName := range r.Tags[t] {
					if _, ok := c.series[sourceName]; !ok {
						c.series[sourceName] = make(map[string]kairosdbSeriesEntry)
					}

					c.series[sourceName][metricName] = kairosdbSeriesEntry{
						tag:        t,
						source:     sourceName,
						metric:     metricName,
						aggregator: aggregator,
					}

					outputChan <- &catalog.Record{
						Origin:    originName,
						Source:    sourceName,
						Metric:    metricName,
						Connector: c,
					}
				}

				if aggregator != nil {
					a, _ := json.Marshal(aggregator)
					logger.Log(logger.LevelInfo, "connector", "kairosdb[%s]: `%s' applied to `%s'", c.name, string(a),
						metricName)
				}

				break
			}
		}
	}

	return nil
}

func kairosdbBuildJSONQuery(query *plot.Query,
	kairosdbSeries map[string]map[string]kairosdbSeriesEntry) ([]byte, error) {
	q := plotsQuery{StartAbsolute: query.StartTime.Unix() * 1000,
		EndAbsolute: query.EndTime.Unix() * 1000}

	for _, series := range query.Series {
		entry := kairosdbSeries[series.Source][series.Metric]
		m := metricQueryEntry{
			Name: entry.metric,
			Tags: map[string][]string{entry.tag: []string{entry.source}},
		}

		// catch `json:"aggregators,omitempty"` to avoid "aggregators": [null]
		if entry.aggregator != nil {
			m.Aggregators = []interface{}{entry.aggregator}
		} else {
			m.Aggregators = nil
		}

		q.Metrics = append(q.Metrics, m)
	}

	jsonData, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func kairosdbExtractPlots(query *plot.Query, kairosdbSeries map[string]map[string]kairosdbSeriesEntry,
	kairosdbPlots []metricQueryResponse) ([]*plot.Series, error) {

	var results []*plot.Series

	for _, kairosdbPlot := range kairosdbPlots {
		target := ""
		for _, series := range query.Series {

			entry := kairosdbSeries[series.Source][series.Metric]

			m := kairosdbPlot.Results[0].Name

			if _, ok := kairosdbPlot.Results[0].Tags[entry.tag]; !ok {
				continue
			}

			s := kairosdbPlot.Results[0].Tags[entry.tag][0]

			if s == series.Source && m == series.Metric {
				if target == "" {
					target = series.Name
				} else {
					return nil, fmt.Errorf("ambiguity during plot target retrieval")
				}
			}
		}

		if target == "" {
			return nil, fmt.Errorf("no plot target found")
		}

		series := &plot.Series{
			Name: target,
		}

		for _, plotPoint := range kairosdbPlot.Results[0].Values {
			series.Plots = append(series.Plots, plot.Plot{
				Time:  time.Unix(int64(plotPoint[0]/1000), 0),
				Value: plot.Value(plotPoint[1]),
			})
		}

		results = append(results, series)
	}

	return results, nil
}

func kairosdbCheckBackendResponse(r *http.Response) error {
	if r.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", r.StatusCode)
	}

	if utils.HTTPGetContentType(r) != "application/json" {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", r.Header["Content-Type"])
	}

	return nil
}

func kairosdbGetVersion(c *KairosdbConnector) (string, [3]int, error) {
	var (
		array       [3]int
		jsonVersion map[string]string
	)

	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	r, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+kairosdbURLVersion, nil)
	if err != nil {
		return "", array, fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "KairosDBConnector")

	rsp, err := client.Do(r)
	if err != nil {
		return "", array, fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer rsp.Body.Close()

	if err = kairosdbCheckBackendResponse(rsp); err != nil {
		return "", array, fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", array, fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	if err := json.Unmarshal(data, &jsonVersion); err != nil {
		return "", array, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	if v, ok := jsonVersion["version"]; ok {
		re, _ := regexp.Compile(`^KairosDB (\d+)\.(\d+)\.(\d+)`) // {"version": "KairosDB 0.9.4-6.20140730155353"}

		submatch := re.FindStringSubmatch(v)
		if submatch == nil || len(submatch) != 4 {
			return "", array, fmt.Errorf("can't match KairosDB version")
		}

		array[0], _ = strconv.Atoi(submatch[1])
		array[1], _ = strconv.Atoi(submatch[2])
		array[2], _ = strconv.Atoi(submatch[3])

		return v, array, nil
	}

	return "", array, fmt.Errorf("can't fetch KairosDB version")
}

func matchAggregatorPattern(aggregators []metricAggregator, metric string) interface{} {
	if aggregators == nil {
		return nil
	}

	for _, a := range aggregators {
		if a.re.MatchString(metric) {
			return a.hook
		}
	}

	return nil
}

func compileAggregatorPatterns(aggregators interface{}, connector string) []metricAggregator {
	var (
		re  *regexp.Regexp
		err error
	)

	if aggregators == nil {
		return nil
	}

	list := aggregators.([]interface{})
	out := make([]metricAggregator, 0)

	for _, a := range list {
		aggregator := a.(map[string]interface{})

		if re, err = regexp.Compile(aggregator["metric"].(string)); err != nil {
			logger.Log(logger.LevelWarning, "connector", "kairosdb[%s]: can't compile `%s', skipping", connector,
				aggregator["metric"].(string))
			continue
		}

		out = append(out, metricAggregator{
			pattern: aggregator["metric"].(string),
			re:      re,
			hook:    aggregator["aggregator"],
		})
	}

	return out
}
