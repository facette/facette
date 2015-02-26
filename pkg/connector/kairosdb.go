// +build kairosdb

package connector

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
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
	kairosdbURLVersion     string  = "/api/v1/version"
	kairosdbURLMetricNames string  = "/api/v1/metricnames"
	kairosdbURLMetricsTags string  = "/api/v1/datapoints/query/tags"
	kairosdbURLQueryMetric string  = "/api/v1/datapoints/query"
	kairosdbDefaultTimeout float64 = 10
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

type MetricsQueryEntry struct {
	Name        string              `json:"name"`
	Tags        map[string][]string `json:"tags"`
	Aggregators []interface{}       `json:"aggregators,omitempty"`
}

type MetricsQueryResponse struct {
	SampleSize int64                `json:"sample_size"`
	Results    []MetricsQueryResult `json:"results"`
}

type MetricsQueryResult struct {
	Name    string              `json:"name"`
	GroupBy []map[string]string `json:"group_by"`
	Tags    map[string][]string `json:"tags"`
	Values  [][2]float64        `json:"values"`
}

// KairosdbConnector represents the main structure of the Kairosdb connector.
type KairosdbConnector struct {
	name              string
	URL               string
	insecureTLS       bool
	timeout           float64
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
		var err error
		var aggregators interface{}
		connector := &KairosdbConnector{
			name:              name,
			insecureTLS:       true,
			timeout:           kairosdbDefaultTimeout,
			sourceTags:        nil,
			startAbsolute:     0, // Note: Must be > 0 because of config.GetInt() behavior
			startRelative:     nil,
			endAbsolute:       0, // Note: Must be > 0 because of config.GetInt() behavior
			endRelative:       nil,
			defaultAggregator: nil,
			aggregators:       nil,
			series:            make(map[string]map[string]kairosdbSeriesEntry),
		}

		if connector.URL, err = config.GetString(settings, "url", true); err != nil {
			return nil, err
		}

		if connector.sourceTags, err = config.GetStringSlice(settings, "source_tags", false); err != nil {
			return nil, err
		}

		if connector.startAbsolute, err = config.GetInt(settings, "start_absolute", false); err != nil {
			return nil, err
		}
		if connector.endAbsolute, err = config.GetInt(settings, "end_absolute", false); err != nil {
			return nil, err
		}

		if connector.startRelative, err = config.GetJsonObj(settings, "start_relative", false); err != nil {
			return nil, err
		}
		if connector.endRelative, err = config.GetJsonObj(settings, "end_relative", false); err != nil {
			return nil, err
		}

		if connector.defaultAggregator, err = config.GetJsonObj(settings, "default_aggregator", false); err != nil {
			return nil, err
		}

		if aggregators, err = config.GetJsonArray(settings, "aggregators", false); err != nil {
			return nil, err
		}
		connector.aggregators = compileAggregatorPatterns(aggregators, connector.name)

		if connector.insecureTLS, err = config.GetBool(settings, "allow_insecure_tls", false); err != nil {
			return nil, err
		}

		if connector.timeout, err = config.GetFloat(settings, "timeout", false); err != nil {
			return nil, err
		}

		if connector.startAbsolute > 0 && connector.startRelative != nil {
			return nil, fmt.Errorf("kairosdb[%s]: start_absolute/start_relative are mutually exclusive", connector.name)
		}
		if connector.endAbsolute > 0 && connector.endRelative != nil {
			return nil, fmt.Errorf("kairosdb[%s]: end_absolute/end_relative are mutually exclusive", connector.name)
		}

		// Enforce startRelative default
		if connector.startAbsolute <= 0 && connector.startRelative == nil {
			connector.startRelative = map[string]interface{}{"value": 3, "unit": "months"}
		}
		// Enforce sourceTags default
		if connector.sourceTags == nil {
			connector.sourceTags = []string{"host", "server", "device"}
		}
		// Enforce minimal timeout value bound
		if connector.timeout <= 0 {
			connector.timeout = kairosdbDefaultTimeout
		}

		version, version_array, err := kairosdbGetVersion(connector)
		if err != nil {
			return nil, fmt.Errorf("kairosdb[%s]: unable to get KairosDB version: %s", connector.name, err)
		}

		logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: `%s' found", connector.name, version)

		if version_array[0] != 0 && version_array[1] != 9 && (version_array[2] != 4 || version_array[2] != 5) {
			return nil, fmt.Errorf("kairosdb[%s]: KairosDB versions 0.9.4/5 supported only", connector.name)
		}

		return connector, nil
	}
}

// GetName returns the name of the current connector.
func (connector *KairosdbConnector) GetName() string {
	return connector.name
}

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (connector *KairosdbConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	var resultSeries []plot.Series

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("kairosdb[%s]: requested series list is empty", connector.name)
	}

	JSONquery, err := kairosdbBuildJSONQuery(query, connector.series)
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to build or marshal JSON query: %s", connector.name, err)
	}

	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			DualStack: true,
			Timeout:   time.Duration(connector.timeout) * time.Second,
		}).Dial,
	}
	if connector.insecureTLS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	httpClient := http.Client{Transport: httpTransport}

	logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: API Call to %s: %s", connector.name,
		strings.TrimSuffix(connector.URL, "/")+kairosdbURLQueryMetric,
		string(JSONquery))

	request, err := http.NewRequest("POST", strings.TrimSuffix(connector.URL, "/")+kairosdbURLQueryMetric, bytes.NewBuffer(JSONquery))
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "KairosDBConnector")
	request.Header.Set("Content-Type", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to perform HTTP request: %s", connector.name, err)
	}
	defer response.Body.Close()

	if err = kairosdbCheckBackendResponse(response); err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: invalid HTTP backend response: %s", connector.name, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to read HTTP response body: %s", connector.name, err)
	}

	var JSONresponse map[string][]MetricsQueryResponse
	if err = json.Unmarshal(data, &JSONresponse); err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	if resultSeries, err = kairosdbExtractPlots(query, connector.series, JSONresponse["queries"]); err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to extract plot values from backend response: %s", connector.name, err)
	}

	return resultSeries, nil
}

// Refresh triggers a full connector data update.
func (connector *KairosdbConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	var JSONmetrics map[string][]string
	var JSONquery map[string][]map[string][]struct {
		Name string              `json:"name"`
		Tags map[string][]string `json:"tags"`
	}

	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			DualStack: true,
			Timeout:   time.Duration(connector.timeout) * time.Second,
		}).Dial,
	}

	if connector.insecureTLS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	httpClient := http.Client{Transport: httpTransport}

	request, err := http.NewRequest("GET", strings.TrimSuffix(connector.URL, "/")+kairosdbURLMetricNames, nil)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "KairosDBConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to perform HTTP request: %s", connector.name, err)
	}
	defer response.Body.Close()

	if err = kairosdbCheckBackendResponse(response); err != nil {
		return fmt.Errorf("kairosdb[%s]: invalid HTTP backend response: %s", connector.name, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to read HTTP response body: %s", connector.name, err)
	}
	if err = json.Unmarshal(data, &JSONmetrics); err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	metrics := make([]map[string]string, 0)
	for _, m := range JSONmetrics["results"] {
		metrics = append(metrics, map[string]string{"name": m})
	}
	query := map[string]interface{}{"metrics": metrics}
	if connector.startAbsolute > 0 {
		query["start_absolute"] = connector.startAbsolute
	} else {
		query["start_relative"] = connector.startRelative
	}
	if connector.endAbsolute > 0 {
		query["end_absolute"] = connector.endAbsolute
	}
	if connector.endRelative != nil {
		query["end_relative"] = connector.endRelative
	}
	jsonData, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to marshal JSON data: %s", connector.name, err)
	}

	logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: API Call to %s: %s", connector.name,
		strings.TrimSuffix(connector.URL, "/")+kairosdbURLMetricsTags,
		string(jsonData))

	request, err = http.NewRequest("POST", strings.TrimSuffix(connector.URL, "/")+kairosdbURLMetricsTags, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "KairosDBConnector")
	request.Header.Set("Content-Type", "application/json")

	response, err = httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to perform HTTP request: %s", connector.name, err)
	}
	defer response.Body.Close()

	if err = kairosdbCheckBackendResponse(response); err != nil {
		return fmt.Errorf("kairosdb[%s]: invalid HTTP backend response: %s", connector.name, err)
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to read HTTP response body: %s", connector.name, err)
	}

	logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: API Response from %s: %s", connector.name,
		strings.TrimSuffix(connector.URL, "/")+kairosdbURLMetricsTags,
		string(data))

	if err = json.Unmarshal(data, &JSONquery); err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	for _, q := range JSONquery["queries"] {
		for _, r := range q["results"] {
			metricName := r.Name
			aggregator := matchAggregatorPattern(connector.aggregators, metricName)
			if aggregator == nil {
				aggregator = connector.defaultAggregator
			}

			for _, t := range connector.sourceTags {
				if _, ok := r.Tags[t]; !ok {
					continue
				}
				var sc uint64
				for _, sourceName := range r.Tags[t] {
					if _, ok := connector.series[sourceName]; !ok {
						connector.series[sourceName] = make(map[string]kairosdbSeriesEntry)
					}
					connector.series[sourceName][metricName] = kairosdbSeriesEntry{
						tag:        t,
						source:     sourceName,
						metric:     metricName,
						aggregator: aggregator,
					}
					outputChan <- &catalog.Record{
						Origin:    originName,
						Source:    sourceName,
						Metric:    metricName,
						Connector: connector,
					}
					sc++
				}
				logger.Log(logger.LevelDebug, "connector", "kairosdb[%s]: %d sources for `%s'", connector.name, sc, metricName)
				if aggregator != nil {
					a, _ := json.Marshal(aggregator)
					logger.Log(logger.LevelInfo, "connector", "kairosdb[%s]: `%s' applied to `%s'",
						connector.name, string(a), metricName)
				}
				break
			}
		}
	}
	return nil
}

func kairosdbBuildJSONQuery(query *plot.Query, kairosdbSeries map[string]map[string]kairosdbSeriesEntry) ([]byte, error) {
	type PlotsQuery struct {
		StartAbsolute int64               `json:"start_absolute"`
		EndAbsolute   int64               `json:"end_absolute"`
		Metrics       []MetricsQueryEntry `json:"metrics"`
	}

	q := PlotsQuery{StartAbsolute: query.StartTime.Unix() * 1000,
		EndAbsolute: query.EndTime.Unix() * 1000}

	for _, series := range query.Series {
		entry := kairosdbSeries[series.Source][series.Metric]
		m := MetricsQueryEntry{
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

func kairosdbExtractPlots(query *plot.Query, kairosdbSeries map[string]map[string]kairosdbSeriesEntry, kairosdbPlots []MetricsQueryResponse) ([]plot.Series, error) {
	var resultSeries []plot.Series

	for _, kairosdbPlot := range kairosdbPlots {

		// Is there a better approach to retrieve target?
		var target string = ""
		for _, series := range query.Series {

			entry := kairosdbSeries[series.Source][series.Metric]

			// (KairosDB API): more than one result possible?
			m := kairosdbPlot.Results[0].Name

			if _, ok := kairosdbPlot.Results[0].Tags[entry.tag]; !ok {
				continue
			}

			// (KairosDB API): more than one result possible?
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
		series := plot.Series{
			Name:    target,
			Summary: make(map[string]plot.Value),
		}

		// (KairosDB API): more than one result possible?
		for _, plotPoint := range kairosdbPlot.Results[0].Values {
			series.Plots = append(
				series.Plots,
				plot.Plot{Value: plot.Value(plotPoint[1]), Time: time.Unix(int64(plotPoint[0]/1000), 0)},
			)
		}
		resultSeries = append(resultSeries, series)
	}
	return resultSeries, nil
}

func kairosdbCheckBackendResponse(response *http.Response) error {
	if response.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", response.StatusCode)
	}

	if utils.HTTPGetContentType(response) != "application/json" {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", response.Header["Content-Type"])
	}

	return nil
}

func kairosdbGetVersion(connector *KairosdbConnector) (string, [3]int, error) {
	var array [3]int
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			DualStack: true,
			Timeout:   time.Duration(connector.timeout) * time.Second,
		}).Dial,
	}
	if connector.insecureTLS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	httpClient := http.Client{Transport: httpTransport}

	request, err := http.NewRequest("GET", strings.TrimSuffix(connector.URL, "/")+kairosdbURLVersion, nil)
	if err != nil {
		return "", array, fmt.Errorf("unable to set up HTTP request: %s", err)
	}
	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "KairosDBConnector")
	response, err := httpClient.Do(request)
	if err != nil {
		return "", array, fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer response.Body.Close()
	if err = kairosdbCheckBackendResponse(response); err != nil {
		return "", array, fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", array, fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	var versionJSON map[string]string

	if err := json.Unmarshal(data, &versionJSON); err != nil {
		return "", array, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	if v, ok := versionJSON["version"]; ok {
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
	var re *regexp.Regexp
	var err error

	if aggregators == nil {
		return nil
	}

	list := aggregators.([]interface{})
	res := make([]metricAggregator, 0)

	for _, a := range list {
		aggregator := a.(map[string]interface{})
		if re, err = regexp.Compile(aggregator["metric"].(string)); err != nil {
			logger.Log(logger.LevelWarning, "connector", "kairosdb[%s]: can't compile `%s', skipping",
				connector, aggregator["metric"].(string))
			continue
		}
		res = append(res, metricAggregator{pattern: aggregator["metric"].(string), re: re, hook: aggregator["aggregator"]})
	}
	return res
}
