// +build graphite

package connector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	graphiteDefaultTimeout int    = 10
	graphiteURLMetrics     string = "/metrics/index.json"
	graphiteURLRender      string = "/render"
)

type graphitePlot struct {
	Target     string
	Datapoints [][2]float64
}

// GraphiteConnector represents the main structure of the Graphite connector.
type GraphiteConnector struct {
	name        string
	url         string
	timeout     int
	insecureTLS bool
	re          *regexp.Regexp
	series      map[string]map[string]string
}

func init() {
	Connectors["graphite"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var (
			pattern string
			err     error
		)

		c := &GraphiteConnector{
			name:   name,
			series: make(map[string]map[string]string),
		}

		if c.url, err = config.GetString(settings, "url", true); err != nil {
			return nil, err
		}

		if c.timeout, err = config.GetInt(settings, "timeout", false); err != nil {
			return nil, err
		}
		if c.timeout <= 0 {
			c.timeout = graphiteDefaultTimeout
		}

		if c.insecureTLS, err = config.GetBool(settings, "allow_insecure_tls", false); err != nil {
			return nil, err
		}

		if pattern, err = config.GetString(settings, "pattern", true); err != nil {
			return nil, err
		}
		if c.re, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		}

		return c, nil
	}
}

// GetName returns the name of the current connector.
func (c *GraphiteConnector) GetName() string {
	return c.name
}

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (c *GraphiteConnector) GetPlots(query *plot.Query) ([]*plot.Series, error) {
	var (
		plots   []graphitePlot
		results []*plot.Series
	)

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("graphite[%s]: requested series list is empty", c.name)
	}

	// Build query URL
	queryURL, err := graphiteBuildQueryURL(query, c.series)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to build query URL: %s", c.name, err)
	}

	// Request data from backend
	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	r, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+graphiteURLRender+"?"+queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "GraphiteConnector")

	rsp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	// Parse backend response
	if err = graphiteCheckBackendResponse(rsp); err != nil {
		return nil, fmt.Errorf("graphite[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to read HTTP response body: %s", c.name, err)
	}

	if err = json.Unmarshal(data, &plots); err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	// Extract results from response
	if results, err = graphiteExtractResult(plots); err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to extract plot values from backend response: %s", c.name, err)
	}

	return results, nil
}

// Refresh triggers a full connector data update.
func (c *GraphiteConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	var series []string

	// Request metrics from backend
	client := utils.NewHTTPClient(c.timeout, c.insecureTLS)

	r, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+graphiteURLMetrics, nil)
	if err != nil {
		return fmt.Errorf("graphite[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "GraphiteConnector")

	rsp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("graphite[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	// Parse backend response
	if err = graphiteCheckBackendResponse(rsp); err != nil {
		return fmt.Errorf("graphite[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("graphite[%s]: unable to read HTTP response body: %s", c.name, err)
	}

	if err = json.Unmarshal(data, &series); err != nil {
		return fmt.Errorf("graphite[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	for _, s := range series {
		var sourceName, metricName string

		// Get pattern matches
		m, err := matchSeriesPattern(c.re, s)
		if err != nil {
			logger.Log(
				logger.LevelInfo,
				"connector",
				"graphite[%s]: file `%s' does not match pattern, ignoring",
				c.name,
				s,
			)
			continue
		}

		sourceName, metricName = m[0], m[1]

		if _, ok := c.series[sourceName]; !ok {
			c.series[sourceName] = make(map[string]string)
		}

		c.series[sourceName][metricName] = s

		outputChan <- &catalog.Record{
			Origin:    originName,
			Source:    sourceName,
			Metric:    metricName,
			Connector: c,
		}
	}

	return nil
}

func graphiteCheckBackendResponse(r *http.Response) error {
	if r.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", r.StatusCode)
	}

	if utils.HTTPGetContentType(r) != "application/json" {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", r.Header["Content-Type"])
	}

	return nil
}

func graphiteBuildQueryURL(query *plot.Query, graphiteSeries map[string]map[string]string) (string, error) {
	now := time.Now()

	fromTime := 0

	queryURL := "format=json"

	for _, series := range query.Series {
		queryURL += fmt.Sprintf(
			"&target=alias(%s, \"%s\")",
			url.QueryEscape(graphiteSeries[series.Source][series.Metric]),
			series.Name,
		)
	}

	if query.StartTime.Before(now) {
		fromTime = int(now.Sub(query.StartTime).Seconds())
	}

	queryURL += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify `until' parameter if EndTime is still in the past
	if query.EndTime.Before(now) {
		queryURL += fmt.Sprintf("&until=-%ds", int(time.Now().Sub(query.EndTime).Seconds()))
	}

	return queryURL, nil
}

func graphiteExtractResult(plots []graphitePlot) ([]*plot.Series, error) {
	var results []*plot.Series

	for _, p := range plots {
		series := &plot.Series{
			Name: p.Target,
		}

		for _, d := range p.Datapoints {
			series.Plots = append(series.Plots, plot.Plot{
				Time:  time.Unix(int64(d[1]), 0),
				Value: plot.Value(d[0]),
			})
		}

		results = append(results, series)
	}

	return results, nil
}
