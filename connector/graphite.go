// +build !disable_connector_graphite

package connector

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"facette.io/facette/catalog"
	"facette.io/facette/series"
	"facette.io/facette/version"
	"facette.io/httputil"
	"facette.io/logger"
	"facette.io/maputil"
)

const (
	graphiteURLMetrics = "/metrics/index.json"
	graphiteURLRender  = "/render"
)

func init() {
	connectors["graphite"] = func(name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
		var (
			pattern string
			err     error
		)

		c := &graphiteConnector{
			name:   name,
			series: make(map[string]map[string]string),
			logger: logger,
		}

		// Load provider configuration
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

		pattern, err = settings.GetString("pattern", "")
		if err != nil {
			return nil, err
		} else if pattern == "" {
			return nil, ErrMissingConnectorSetting("pattern")
		}

		// Check remote instance URL
		_, err = url.Parse(c.url)
		if err != nil {
			return nil, fmt.Errorf("unable to parse URL: %s", err)
		}

		// Check and compile regexp pattern
		c.pattern, err = compilePattern(pattern)
		if err != nil {
			return nil, err
		}

		c.client = httputil.NewClient(time.Duration(c.timeout)*time.Second, true, c.allowInsecure)

		return c, nil
	}
}

type graphiteConnector struct {
	name          string
	url           string
	timeout       int
	allowInsecure bool
	pattern       *regexp.Regexp
	client        *http.Client
	series        map[string]map[string]string
	logger        *logger.Logger
}

func (c *graphiteConnector) Name() string {
	return c.name
}

func (c *graphiteConnector) Points(query *series.Query) ([]series.Series, error) {
	var (
		points  []graphitePoint
		results []series.Series
	)

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("requested series list is empty")
	}

	queryURL, err := graphiteBuildQueryURL(query, c.series)
	if err != nil {
		return nil, fmt.Errorf("unable to build query URL: %s", err)
	}

	// Request data from back-end
	req, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+graphiteURLRender+"?"+queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to set up HTTP request: %s", err)
	}
	req.Header.Add("User-Agent", "facette/"+version.Version)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	err = httputil.BindJSON(resp, &points)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	// Extract results from response
	results, err = graphiteExtractResult(points)
	if err != nil {
		return nil, fmt.Errorf("unable to extract point values from back-end response: %s", err)
	}

	return results, nil
}

func (c *graphiteConnector) Refresh(output chan<- *catalog.Record) error {
	var series []string

	req, err := http.NewRequest("GET", strings.TrimSuffix(c.url, "/")+graphiteURLMetrics, nil)
	if err != nil {
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}
	req.Header.Add("User-Agent", "facette/"+version.Version)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}
	defer resp.Body.Close()

	err = httputil.BindJSON(resp, &series)
	if err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for _, s := range series {
		var sourceName, metricName string

		seriesMatch, err := matchPattern(c.pattern, s)
		if err != nil {
			c.logger.Error("%s", err)
			continue
		}

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

func graphiteBuildQueryURL(query *series.Query, graphiteSeries map[string]map[string]string) (string, error) {
	now := time.Now()
	fromTime := 0

	queryURL := "format=json"

	for _, series := range query.Series {
		queryURL += fmt.Sprintf("&target=%s", url.QueryEscape(graphiteSeries[series.Source][series.Metric]))
	}

	if query.StartTime.Before(now) {
		fromTime = int(now.Sub(query.StartTime).Seconds())
	}
	queryURL += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify `until' parameter if EndTime is still in the past
	if query.EndTime.Before(now) {
		queryURL += fmt.Sprintf("&until=-%ds", int(time.Since(query.EndTime).Seconds()))
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

type graphitePoint struct {
	Target     string
	Datapoints [][2]float64
}
