package catalog

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/facette/facette/pkg/types"
)

const (
	graphiteMetricsURL string = "/metrics/index.json"
	graphiteRenderURL  string = "/render"
)

// GraphiteConnectorHandler represents the main structure of the Graphite backend.
type GraphiteConnectorHandler struct {
	URL                  string
	AllowBadCertificates bool
	origin               *Origin
}

type graphitePlot struct {
	Target     string
	Datapoints [][2]float64
}

// GetPlots calculates and returns plot data based on a time interval.
func (handler *GraphiteConnectorHandler) GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64) (map[string]*PlotResult, error) {

	var (
		data          []byte
		err           error
		graphitePlots []graphitePlot
		httpTransport http.RoundTripper
		queryURL      string
		res           *http.Response
		serieName     string
	)

	if handler.AllowBadCertificates {
		httpTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		httpTransport = &http.Transport{}
	}

	httpClient := http.Client{Transport: httpTransport}

	pr := make(map[string]*PlotResult)

	if serieName, queryURL, err = graphiteBuildQueryURL(query, startTime, endTime); err != nil {
		return nil, fmt.Errorf("unable to build Graphite query URL: %s", err)
	}

	graphiteQueryURL := fmt.Sprintf("%s%s", strings.TrimSuffix(handler.URL, "/"), queryURL)

	if res, err = httpClient.Get(graphiteQueryURL); err != nil {
		return nil, err
	}

	if err = graphiteCheckConnectorResponse(res); err != nil {
		return nil, fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	if data, err = ioutil.ReadAll(res.Body); err != nil {
		return nil, fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	if err = json.Unmarshal(data, &graphitePlots); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	if pr[serieName], err = graphiteExtractPlotResult(graphitePlots); err != nil {
		return nil, fmt.Errorf("unable to extract plot values from backend response: %s", err)
	}

	return pr, nil
}

// GetValue calculates and returns plot data at a specific reference time.
func (handler *GraphiteConnectorHandler) GetValue(query *GroupQuery, refTime time.Time,
	percentiles []float64) (map[string]map[string]types.PlotValue, error) {

	return nil, nil
}

// Update triggers a full backend data update.
func (handler *GraphiteConnectorHandler) Update() error {
	var (
		data           []byte
		err            error
		facetteMetric  string
		facetteSource  string
		httpClient     http.Client
		httpTransport  http.RoundTripper
		metrics        []string
		res            *http.Response
		sourceSepIndex int
	)

	if handler.AllowBadCertificates {
		httpTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		httpTransport = &http.Transport{}
	}

	httpClient = http.Client{Transport: httpTransport}

	if res, err = httpClient.Get(strings.TrimSuffix(handler.URL, "/") + graphiteMetricsURL); err != nil {
		return err
	}

	if err = graphiteCheckConnectorResponse(res); err != nil {
		return fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	if data, err = ioutil.ReadAll(res.Body); err != nil {
		return fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	if err = json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for _, metric := range metrics {
		sourceSepIndex = strings.Index(metric, ".")

		if sourceSepIndex == -1 {
			facetteSource = handler.origin.Name
			facetteMetric = metric
		} else {
			facetteSource = metric[0:sourceSepIndex]
			facetteMetric = metric[sourceSepIndex+1:]
		}

		handler.origin.inputChan <- [2]string{facetteSource, facetteMetric}
	}

	// Close channel once updated
	close(handler.origin.inputChan)

	return nil
}

func init() {
	ConnectorHandlers["graphite"] = NewGraphiteConnectorHandler
}

func graphiteCheckConnectorResponse(res *http.Response) error {
	if res.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("got HTTP content type \"%s\", expected \"application/json\"",
			res.Header["Content-Type"])
	}

	return nil
}

func graphiteBuildQueryURL(query *GroupQuery, startTime, endTime time.Time) (string, string, error) {
	var (
		serieName string
		target    string
		targets   []string
	)

	now := time.Now()

	fromTime := 0

	graphiteQueryURL := fmt.Sprintf("%s?format=json", graphiteRenderURL)

	if query.Type == OperGroupTypeNone {
		serieName = query.Series[0].Name
		target = fmt.Sprintf("%s.%s", query.Series[0].Metric.source.OriginalName, query.Series[0].Metric.OriginalName)
	} else {
		serieName = query.Name
		targets = make([]string, 0)

		for _, s := range query.Series {
			targets = append(targets, fmt.Sprintf("%s.%s", s.Metric.source.OriginalName, s.Metric.OriginalName))
		}

		target = fmt.Sprintf("group(%s)", strings.Join(targets, ","))

		switch query.Type {
		case OperGroupTypeAvg:
			target = fmt.Sprintf("averageSeries(%s)", target)
		case OperGroupTypeSum:
			target = fmt.Sprintf("sumSeries(%s)", target)
		}
	}

	target = fmt.Sprintf("legendValue(%s, 'min', 'max', 'avg', 'last')", target)

	graphiteQueryURL += fmt.Sprintf("&target=%s", target)

	if startTime.Before(now) {
		fromTime = int(now.Sub(startTime).Seconds())
	}

	graphiteQueryURL += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify "until" parameter if endTime is still in the past
	if endTime.Before(now) {
		untilTime := int(time.Now().Sub(endTime).Seconds())
		graphiteQueryURL += fmt.Sprintf("&until=-%ds", untilTime)
	}

	return serieName, graphiteQueryURL, nil
}

func graphiteExtractPlotResult(graphitePlots []graphitePlot) (*PlotResult, error) {
	var min, max, avg, last float64

	pr := &PlotResult{Info: make(map[string]types.PlotValue)}

	// Return an empty plotResult if Graphite API didn't return any datapoint matching the query
	if len(graphitePlots) == 0 || len(graphitePlots[0].Datapoints) == 0 {
		return pr, nil
	}

	for _, plotPoint := range graphitePlots[0].Datapoints {
		pr.Plots = append(pr.Plots, types.PlotValue(plotPoint[0]))
	}

	// Scan the target legend for plot min/max/avg/last info
	if idx := strings.Index(graphitePlots[0].Target, "(min"); idx > 0 {
		fmt.Sscanf(graphitePlots[0].Target[idx:],
			"(min: %f) (max: %f) (avg: %f) (last: %f)",
			&min,
			&max,
			&avg,
			&last)
	}

	pr.Info["min"] = types.PlotValue(min)
	pr.Info["max"] = types.PlotValue(max)
	pr.Info["avg"] = types.PlotValue(avg)
	pr.Info["last"] = types.PlotValue(last)

	return pr, nil
}

// NewGraphiteConnectorHandler creates a new instance of ConnectorHandler.
func NewGraphiteConnectorHandler(origin *Origin, config map[string]string) error {
	var (
		graphiteConnector *GraphiteConnectorHandler
	)

	if _, present := config["url"]; !present {
		return fmt.Errorf("missing `url' mandatory backend definition")
	}

	graphiteConnector = &GraphiteConnectorHandler{
		URL:                  config["url"],
		AllowBadCertificates: false,
		origin:               origin,
	}

	if config["allow_bad_certificates"] == "yes" {
		graphiteConnector.AllowBadCertificates = true
	}

	origin.Connector = graphiteConnector

	return nil
}
