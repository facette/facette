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
	aggregator MetricsQueryAggregation
}

type MetricsQueryAggregation struct {
	Name     string            `json:"name"`
	Sampling map[string]string `json:"sampling"`
}

type MetricsQueryEntry struct {
	Name        string                    `json:"name"`
	Tags        map[string][]string       `json:"tags"`
	Aggregators []MetricsQueryAggregation `json:"aggregators"`
}

type MetricsQueryResponse struct {
	Sample_size int64                `json:"sample_size"`
	Results     []MetricsQueryResult `json:"results"`
}

type MetricsQueryResult struct {
	Name     string              `json:"name"`
	Group_by []map[string]string `json:"group_by"`
	Tags     map[string][]string `json:"tags"`
	Values   [][2]float64        `json:"values"`
}

// KairosdbConnector represents the main structure of the Kairosdb connector.
type KairosdbConnector struct {
	name        string
	URL         string
	insecureTLS bool
	timeout     float64
	srcTags     []string       // TODO: make configurable
	sourceRe    *regexp.Regexp // REVIEW: see below
	metricRe    *regexp.Regexp // REVIEW: see below
	series      map[string]map[string]kairosdbSeriesEntry
}

func init() {
	Connectors["kairosdb"] = func(name string, settings map[string]interface{}) (Connector, error) {
		// var sp, mp string // REVIEW: see below
		var err error

		connector := &KairosdbConnector{
			name:        name,
			insecureTLS: true,
			series:      make(map[string]map[string]kairosdbSeriesEntry),
			srcTags:     []string{"host", "name"}, // TODO: make configurable
		}

		if connector.URL, err = config.GetString(settings, "url", true); err != nil {
			return nil, err
		}

		if connector.insecureTLS, err = config.GetBool(settings, "allow_insecure_tls", false); err != nil {
			return nil, err
		}

		if connector.timeout, err = config.GetFloat(settings, "timeout", false); err != nil {
			return nil, err
		}
		/*
		 * REVIEW: Seems, that this feature isn't mandatory.
		 *         Usage of connectors filter capability allows to skip/modify source and metric names
		 *
		 * Add e.g.
		 *	"source_pattern": "^(?P<source>[^\\./]+)[^/]*$",
		 *	"metric_pattern": "^cpu.idle.summation|entropy.entropy.value$"
		 * to connector in provider.json and uncomment, if you think its usefully...

		   		if sp, err = config.GetString(settings, "source_pattern", false); err != nil {
		   			return nil, err
		   		}
		   		if mp, err = config.GetString(settings, "metric_pattern", false); err != nil {
		   			return nil, err
		   		}

		   		// Check and compile regexp pattern
		   		if connector.sourceRe, err = compileRePattern(sp); err != nil {
		   			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		   		}
		   		if connector.metricRe, err = compileRePattern(mp); err != nil {
		   			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		   		}
		*/

		// Enforce minimal timeout value bound
		if connector.timeout <= 0 {
			connector.timeout = kairosdbDefaultTimeout
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
			// Enable dual IPv4/IPv6 stack connectivity:
			DualStack: true,
			// Enforce HTTP connection timeout:
			Timeout: time.Duration(connector.timeout) * time.Second,
		}).Dial,
	}
	if connector.insecureTLS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	httpClient := http.Client{Transport: httpTransport}

	request, err := http.NewRequest("POST", strings.TrimSuffix(connector.URL, "/")+kairosdbURLQueryMetric, bytes.NewBuffer(JSONquery))
	if err != nil {
		return nil, fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "KairosdbConnector")
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

	// return nil, fmt.Errorf("kairosdb[%s]: not implemented, yet.", connector.name)

	return resultSeries, nil
}

// Refresh triggers a full connector data update.
func (connector *KairosdbConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	type MetricsQuery struct {
		Start_relative map[string]string   `json:"start_relative"`
		Metrics        []map[string]string `json:"metrics"`
	}

	var JSONmetrics map[string][]string
	var JSONquery map[string][]map[string][]struct {
		Name string              `json:"name"`
		Tags map[string][]string `json:"tags"`
	}

	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			// Enable dual IPv4/IPv6 stack connectivity:
			DualStack: true,
			// Enforce HTTP connection timeout:
			Timeout: time.Duration(connector.timeout) * time.Second,
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
	request.Header.Add("X-Requested-With", "KairosdbConnector")

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

	// REVIEW: whats a good interval to populate metrics from?
	// TODO: make configurable
	query := MetricsQuery{Start_relative: map[string]string{"value": "3600", "unit": "seconds"}}
	for _, metric := range JSONmetrics["results"] {
		var m = map[string]string{"name": metric}
		query.Metrics = append(query.Metrics, m)
	}
	jsonData, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to marshal JSON data: %s", connector.name, err)
	}

	request, err = http.NewRequest("POST", strings.TrimSuffix(connector.URL, "/")+kairosdbURLMetricsTags, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "KairosdbConnector")
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
	if err = json.Unmarshal(data, &JSONquery); err != nil {
		return fmt.Errorf("kairosdb[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	// REVIEW: If we really remove pattern capability above, the code becomes more simply here
	for _, q := range JSONquery["queries"] {
		for _, r := range q["results"] {
			metricName := r.Name
			m, err := matchRePattern("metric", connector.metricRe, metricName)
			if err != nil {
				logger.Log(logger.LevelInfo, "connector", "kairosdb[%s]: metric `%s' does not match pattern, ignoring", connector.name, metricName)
				continue
			}

			sources := make(map[string]string)
			tag := ""
			for _, t := range connector.srcTags {
				_, ok := r.Tags[t]
				if !ok {
					continue
				}
				tag = t
				for _, sourceName := range r.Tags[t] {
					s, err := matchRePattern("source", connector.sourceRe, sourceName)
					if err != nil {
						logger.Log(logger.LevelInfo, "connector", "kairosdb[%s]: source `%s' does not match pattern, ignoring", connector.name, sourceName)
						continue
					}
					if len(s) > 0 {
						sources[sourceName] = s
					} else {
						sources[sourceName] = sourceName
					}
				}
				break
			}

			if len(sources) > 0 {
				mk := metricName
				if m != metricName && len(m) > 0 {
					mk = m
				}
				for k, v := range sources {
					sk := k
					if k != v {
						sk = v
					}
					if _, ok := connector.series[sk]; !ok {
						connector.series[sk] = make(map[string]kairosdbSeriesEntry)
					}
					// TODO: make aggregator configurable via metrics pattern
					connector.series[sk][mk] = kairosdbSeriesEntry{
						metric: metricName,
						tag:    tag,
						source: k,
						aggregator: MetricsQueryAggregation{
							Name: "max",
							Sampling: map[string]string{
								"value": "5",
								"unit":  "minutes"}}}

					outputChan <- &catalog.Record{
						Origin:    originName,
						Source:    sk,
						Metric:    mk,
						Connector: connector,
					}
				}
			}
		}
	}
	return nil
}

func kairosdbBuildJSONQuery(query *plot.Query, kairosdbSeries map[string]map[string]kairosdbSeriesEntry) ([]byte, error) {
	type PlotsQuery struct {
		Start_absolute int64               `json:"start_absolute"`
		End_absolute   int64               `json:"end_absolute"`
		Metrics        []MetricsQueryEntry `json:"metrics"`
	}

	q := PlotsQuery{Start_absolute: query.StartTime.Unix() * 1000,
			End_absolute: query.EndTime.Unix() * 1000}

	for _, series := range query.Series {
		entry := kairosdbSeries[series.Source][series.Metric]
		var m = MetricsQueryEntry{Name: entry.metric,
			Tags:        map[string][]string{entry.tag: []string{entry.source}},
			Aggregators: []MetricsQueryAggregation{entry.aggregator}}
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

		// REVIEW: is there a better approach to retrieve target?
		var target string = ""
		for _, series := range query.Series {

			entry := kairosdbSeries[series.Source][series.Metric]

			// REVIEW (kairosdb API): more than one result possible?
			m := kairosdbPlot.Results[0].Name

			if _, ok := kairosdbPlot.Results[0].Tags[entry.tag]; !ok {
				continue
			}

			// REVIEW (kairosdb API): more than one result possible?
			s := kairosdbPlot.Results[0].Tags[entry.tag][0]

			if s == series.Source && m == series.Metric {
				if target == "" {
					target = series.Name
				} else {
					return nil, fmt.Errorf("ambiguity during plot target retrieval")
				}
				// break
			}
		}
		if target == "" {
			return nil, fmt.Errorf("no plot target found")
		}
		series := plot.Series{
			Name:    target,
			Summary: make(map[string]plot.Value),
		}

		// REVIEW (kairosdb API): more than one result possible?
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

func compileRePattern(pattern string) (*regexp.Regexp, error) {
	var (
		re  *regexp.Regexp
		err error
	)

	if re, err = regexp.Compile(pattern); err != nil {
		return nil, err
	}
	groups := make(map[string]bool)
	for _, key := range re.SubexpNames() {
		if key == "" {
			continue
		} else if key == "source" || key == "metric" {
			groups[key] = true
		} else {
			return nil, fmt.Errorf("invalid pattern keyword `%s'", key)
		}
	}
	if groups["source"] && groups["metric"] {
		return nil, fmt.Errorf("only one pattern keyword `source' or `metric' allowed")
	}
	return re, nil
}

func matchRePattern(keyword string, re *regexp.Regexp, s string) (string, error) {
	if re == nil {
		return "", nil
	}

	submatch := re.FindStringSubmatch(s)

	if submatch == nil {
		return "", fmt.Errorf("`%s' does not match pattern", s)
	}

	// TODO: (named) subexp matching needs a better less ambiguous implementation
	if re.NumSubexp() > 0 {
		if re.SubexpNames()[1] == keyword {
			return submatch[1], nil
		}
		return "", fmt.Errorf("no named subexp or not `%s' or not the first one", keyword)
	}
	return "", nil
}
