package connector

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
	graphiteURLMetrics string = "/metrics/index.json"
	graphiteURLRender  string = "/render"
)

type graphitePlot struct {
	Target     string
	Datapoints [][2]float64
}

// GraphiteConnector represents the main structure of the Graphite connector.
type GraphiteConnector struct {
	URL         string
	InsecureTLS bool
	inputChan   *chan [2]string
}

// GetPlots calculates and returns plots data based on a time interval.
func (handler *GraphiteConnector) GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64) (map[string]*PlotResult, error) {

	result := make(map[string]*PlotResult)

	httpTransport := &http.Transport{}
	if handler.InsecureTLS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	httpClient := http.Client{Transport: httpTransport}

	serieName, queryURL, err := graphiteBuildQueryURL(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("unable to build Graphite query URL: %s", err.Error())
	}

	response, err := httpClient.Get(strings.TrimSuffix(handler.URL, "/") + queryURL)
	if err != nil {
		return nil, err
	}

	if err = graphiteCheckConnectorResponse(response); err != nil {
		return nil, fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	graphitePlots := make([]graphitePlot, 0)
	if err = json.Unmarshal(data, &graphitePlots); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	if result[serieName], err = graphiteExtractPlotResult(graphitePlots); err != nil {
		return nil, fmt.Errorf("unable to extract plot values from backend response: %s", err)
	}

	return result, nil
}

// Refresh triggers a full connector data update.
func (handler *GraphiteConnector) Refresh() error {
	httpTransport := &http.Transport{}
	if handler.InsecureTLS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	httpClient := http.Client{Transport: httpTransport}

	response, err := httpClient.Get(strings.TrimSuffix(handler.URL, "/") + graphiteURLMetrics)
	if err != nil {
		return err
	}

	if err = graphiteCheckConnectorResponse(response); err != nil {
		return fmt.Errorf("invalid HTTP backend response: %s", err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	metrics := make([]string, 0)
	if err = json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %s", err)
	}

	for _, metric := range metrics {
		var sourceName, metricName string

		index := strings.Index(metric, ".")

		if index == -1 {
			// TODO: fix?
			sourceName = "<unknown>"
			metricName = metric
		} else {
			sourceName = metric[0:index]
			metricName = metric[index+1:]
		}

		*handler.inputChan <- [2]string{sourceName, metricName}
	}

	// Close channel once updated
	close(*handler.inputChan)

	return nil
}

func graphiteCheckConnectorResponse(response *http.Response) error {
	if response.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", response.StatusCode)
	}

	if response.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("got HTTP content type `%s', expected `application/json'", response.Header["Content-Type"])
	}

	return nil
}

func graphiteBuildQueryURL(query *GroupQuery, startTime, endTime time.Time) (string, string, error) {
	var (
		serieName string
		target    string
	)

	now := time.Now()

	fromTime := 0

	queryURL := fmt.Sprintf("%s?format=json", graphiteURLRender)

	if query.Type == OperGroupTypeNone {
		serieName = query.Series[0].Name
		target = fmt.Sprintf("%s.%s", query.Series[0].Metric.SourceName, query.Series[0].Metric.Name)
	} else {
		serieName = query.Name
		targets := make([]string, 0)

		for _, s := range query.Series {
			targets = append(targets, fmt.Sprintf("%s.%s", s.Metric.SourceName, s.Metric.Name))
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

	queryURL += fmt.Sprintf("&target=%s", target)

	if startTime.Before(now) {
		fromTime = int(now.Sub(startTime).Seconds())
	}

	queryURL += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify `until' parameter if endTime is still in the past
	if endTime.Before(now) {
		untilTime := int(time.Now().Sub(endTime).Seconds())
		queryURL += fmt.Sprintf("&until=-%ds", untilTime)
	}

	return serieName, queryURL, nil
}

func graphiteExtractPlotResult(plots []graphitePlot) (*PlotResult, error) {
	var min, max, avg, last float64

	result := &PlotResult{Info: make(map[string]types.PlotValue)}

	// Return an empty plotResult if Graphite API didn't return any datapoint matching the query
	if len(plots) == 0 || len(plots[0].Datapoints) == 0 {
		return result, nil
	}

	for _, plotPoint := range plots[0].Datapoints {
		result.Plots = append(result.Plots, types.PlotValue(plotPoint[0]))
	}

	// Scan the target legend for plot min/max/avg/last info
	if index := strings.Index(plots[0].Target, "(min"); index > 0 {
		fmt.Sscanf(plots[0].Target[index:], "(min: %f) (max: %f) (avg: %f) (last: %f)", &min, &max, &avg, &last)
	}

	result.Info["min"] = types.PlotValue(min)
	result.Info["max"] = types.PlotValue(max)
	result.Info["avg"] = types.PlotValue(avg)
	result.Info["last"] = types.PlotValue(last)

	return result, nil
}

func init() {
	Connectors["graphite"] = func(inputChan *chan [2]string, config map[string]string) (interface{}, error) {
		if _, ok := config["url"]; !ok {
			return nil, fmt.Errorf("missing `url' mandatory connector setting")
		}

		connector := &GraphiteConnector{
			URL:       config["url"],
			inputChan: inputChan,
		}

		if config["allow_insecure_tls"] == "yes" {
			connector.InsecureTLS = true
		}

		return connector, nil
	}
}
