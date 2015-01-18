// +build graphite

package connector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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
	graphiteURLMetrics     string  = "/metrics/index.json"
	graphiteURLRender      string  = "/render"
	graphiteDefaultTimeout float64 = 10
)

type graphitePlot struct {
	Target     string
	Datapoints [][2]float64
}

// GraphiteConnector represents the main structure of the Graphite connector.
type GraphiteConnector struct {
	name        string
	URL         string
	insecureTLS bool
	timeout     float64
	re          *regexp.Regexp
	series      map[string]map[string]string
}

func init() {
	Connectors["graphite"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var (
			pattern string
			err     error
		)

		connector := &GraphiteConnector{
			name:        name,
			insecureTLS: false,
			series:      make(map[string]map[string]string),
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

		if pattern, err = config.GetString(settings, "pattern", true); err != nil {
			return nil, err
		}

		// Check and compile regexp pattern
		if connector.re, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		}

		// Enforce minimal timeout value bound
		if connector.timeout <= 0 {
			connector.timeout = graphiteDefaultTimeout
		}

		return connector, nil
	}
}

// GetName returns the name of the current connector.
func (connector *GraphiteConnector) GetName() string {
	return connector.name
}

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (connector *GraphiteConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	var (
		graphitePlots []graphitePlot
		resultSeries  []plot.Series
	)

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("graphite[%s]: requested series list is empty", connector.name)
	}

	URLQuery, err := graphiteBuildURLQuery(query, connector.series)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to build query URL: %s", connector.name, err)
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

	request, err := http.NewRequest(
		"GET",
		strings.TrimSuffix(connector.URL, "/")+graphiteURLRender+"?"+URLQuery,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "GraphiteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to perform HTTP request: %s", connector.name, err)
	}
	defer response.Body.Close()

	if err = graphiteCheckBackendResponse(response); err != nil {
		return nil, fmt.Errorf("graphite[%s]: invalid HTTP backend response: %s", connector.name, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to read HTTP response body: %s", connector.name, err)
	}

	if err = json.Unmarshal(data, &graphitePlots); err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	if resultSeries, err = graphiteExtractResult(graphitePlots); err != nil {
		return nil, fmt.Errorf(
			"graphite[%s]: unable to extract plot values from backend response: %s",
			connector.name,
			err,
		)
	}

	return resultSeries, nil
}

// Refresh triggers a full connector data update.
func (connector *GraphiteConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	var seriesList []string

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

	request, err := http.NewRequest("GET", strings.TrimSuffix(connector.URL, "/")+graphiteURLMetrics, nil)
	if err != nil {
		return fmt.Errorf("graphite[%s]: unable to set up HTTP request: %s", connector.name, err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "GraphiteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("graphite[%s]: unable to perform HTTP request: %s", connector.name, err)
	}
	defer response.Body.Close()

	if err = graphiteCheckBackendResponse(response); err != nil {
		return fmt.Errorf("graphite[%s]: invalid HTTP backend response: %s", connector.name, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("graphite[%s]: unable to read HTTP response body: %s", connector.name, err)
	}

	if err = json.Unmarshal(data, &seriesList); err != nil {
		return fmt.Errorf("graphite[%s]: unable to unmarshal JSON data: %s", connector.name, err)
	}

	for _, series := range seriesList {
		var sourceName, metricName string

		seriesMatch, err := matchSeriesPattern(connector.re, series)
		if err != nil {
			logger.Log(
				logger.LevelInfo,
				"connector",
				"graphite[%s]: file `%s' does not match pattern, ignoring",
				connector.name,
				series,
			)
			continue
		}

		sourceName, metricName = seriesMatch[0], seriesMatch[1]

		if _, ok := connector.series[sourceName]; !ok {
			connector.series[sourceName] = make(map[string]string)
		}

		connector.series[sourceName][metricName] = series

		outputChan <- &catalog.Record{
			Origin:    originName,
			Source:    sourceName,
			Metric:    metricName,
			Connector: connector,
		}
	}

	return nil
}

func graphiteCheckBackendResponse(response *http.Response) error {
	if response.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", response.StatusCode)
	}

	if utils.HTTPGetContentType(response) != "application/json" {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", response.Header["Content-Type"])
	}

	return nil
}

func graphiteBuildURLQuery(query *plot.Query, graphiteSeries map[string]map[string]string) (string, error) {
	now := time.Now()

	fromTime := 0

	URLQuery := "format=json"

	for _, series := range query.Series {
		URLQuery += fmt.Sprintf(
			"&target=alias(%s, \"%s\")",
			url.QueryEscape(graphiteSeries[series.Source][series.Metric]),
			series.Name,
		)
	}

	if query.StartTime.Before(now) {
		fromTime = int(now.Sub(query.StartTime).Seconds())
	}

	URLQuery += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify `until' parameter if EndTime is still in the past
	if query.EndTime.Before(now) {
		untilTime := int(time.Now().Sub(query.EndTime).Seconds())
		URLQuery += fmt.Sprintf("&until=-%ds", untilTime)
	}

	return URLQuery, nil
}

func graphiteExtractResult(graphitePlots []graphitePlot) ([]plot.Series, error) {
	var resultSeries []plot.Series

	for _, graphitePlot := range graphitePlots {
		series := plot.Series{
			Name:    graphitePlot.Target,
			Summary: make(map[string]plot.Value),
		}

		for _, plotPoint := range graphitePlot.Datapoints {
			series.Plots = append(
				series.Plots,
				plot.Plot{Value: plot.Value(plotPoint[0]), Time: time.Unix(int64(plotPoint[1]), 0)},
			)
		}

		resultSeries = append(resultSeries, series)
	}

	return resultSeries, nil
}
