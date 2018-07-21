// +build !disable_connector_graphite

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

	"facette.io/facette/catalog"
	"facette.io/facette/series"
	"facette.io/httputil"
	"facette.io/logger"
	"facette.io/maputil"
)

const (
	graphiteURLMetrics = "/metrics/index.json"
	graphiteURLRender  = "/render"
)

type graphitePoint struct {
	Target     string
	Datapoints [][2]float64
}

// graphiteConnector implements the connector handler for a Graphite instance.
type graphiteConnector struct {
	name          string
	url           string
	timeout       int
	allowInsecure bool
	pattern       *regexp.Regexp
	client        *http.Client
	series        map[string]map[string]string
}

func init() {
	connectors["graphite"] = func(name string, settings *maputil.Map, log *logger.Logger) (Connector, error) {
		var err error

		c := &graphiteConnector{
			name:   name,
			series: make(map[string]map[string]string),
		}

		// Load provider configuration
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

		pattern, err := settings.GetString("pattern", "")
		if err != nil {
			return nil, err
		} else if pattern == "" {
			return nil, ErrMissingConnectorSetting("pattern")
		}

		// Check remote instance URL
		if _, err := url.Parse(c.url); err != nil {
			return nil, fmt.Errorf("unable to parse URL: %s", err)
		}

		// Check and compile regexp pattern
		if c.pattern, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		}

		// Create new HTTP client
		c.client = httputil.NewClient(time.Duration(c.timeout)*time.Second, true, c.allowInsecure)

		return c, nil
	}
}

// Name returns the name of the current connector.
func (c *graphiteConnector) Name() string {
	return c.name
}

// Refresh triggers the connector data refresh.
func (c *graphiteConnector) Refresh(output chan<- *catalog.Record) error {
	var series []string

	req, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+graphiteURLMetrics, nil)
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	req.Header.Add("User-Agent", "facette/"+version)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	// Parse back-end response
	if err = graphiteCheckBackendResponse(resp); err != nil {
		return fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	if err = json.Unmarshal(data, &series); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for _, s := range series {
		var sourceName, metricName string

		// FIXME: we should return the matchPattern() error to the caller via the eventChan
		seriesMatch, _ := matchPattern(c.pattern, s)

		sourceName, metricName = seriesMatch[0], seriesMatch[1]

		if _, ok := c.series[sourceName]; !ok {
			c.series[sourceName] = make(map[string]string)
		}

		c.series[sourceName][metricName] = s

		output <- &catalog.Record{
			Origin:    c.name,
			Source:    sourceName,
			Metric:    metricName,
			Connector: c,
		}
	}

	return nil
}

// Points retrieves the time series data according to the query parameters and a time interval.
func (c *graphiteConnector) Points(q *series.Query) ([]series.Series, error) {
	var (
		points  []graphitePoint
		results []series.Series
	)

	if len(q.Series) == 0 {
		return nil, fmt.Errorf("requested series list is empty")
	}

	queryURL, err := graphiteBuildQueryURL(q, c.series)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to build query URL: %s", c.name, err)
	}

	// Request data from back-end
	r, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+graphiteURLRender+"?"+queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to set up HTTP request: %s", c.name, err)
	}

	r.Header.Add("User-Agent", "Facette")
	r.Header.Add("X-Requested-With", "GraphiteConnector")

	rsp, err := c.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to perform HTTP request: %s", c.name, err)
	}
	defer rsp.Body.Close()

	// Parse back-end response
	if err = graphiteCheckBackendResponse(rsp); err != nil {
		return nil, fmt.Errorf("graphite[%s]: invalid HTTP backend response: %s", c.name, err)
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to read HTTP response body: %s", c.name, err)
	}

	if err = json.Unmarshal(data, &points); err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to unmarshal JSON data: %s", c.name, err)
	}

	// Extract results from response
	if results, err = graphiteExtractResult(points); err != nil {
		return nil, fmt.Errorf("graphite[%s]: unable to extract point values from back-end response: %s", c.name, err)
	}

	return results, nil
}

func graphiteCheckBackendResponse(resp *http.Response) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", resp.StatusCode)
	}

	if ct, err := httputil.GetContentType(resp); err != nil {
		return err
	} else if ct != "application/json" {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", resp.Header["Content-Type"])
	}

	return nil
}

func graphiteBuildQueryURL(q *series.Query, graphiteSeries map[string]map[string]string) (string, error) {
	now := time.Now()

	fromTime := 0

	queryURL := "format=json"

	for _, series := range q.Series {
		queryURL += fmt.Sprintf("&target=%s", url.QueryEscape(graphiteSeries[series.Source][series.Metric]))
	}

	if q.StartTime.Before(now) {
		fromTime = int(now.Sub(q.StartTime).Seconds())
	}

	queryURL += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify `until' parameter if EndTime is still in the past
	if q.EndTime.Before(now) {
		queryURL += fmt.Sprintf("&until=-%ds", int(time.Now().Sub(q.EndTime).Seconds()))
	}

	return queryURL, nil
}

func graphiteExtractResult(points []graphitePoint) ([]series.Series, error) {
	var results []series.Series

	for _, p := range points {
		s := series.Series{}
		for _, d := range p.Datapoints {
			s.Points = append(s.Points, series.Point{
				Time:  time.Unix(int64(d[1]), 0),
				Value: series.Value(d[0]),
			})
		}

		results = append(results, s)
	}

	return results, nil
}
