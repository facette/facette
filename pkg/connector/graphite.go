package connector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/types"
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
	URL         string
	insecureTLS bool
	timeout     float64
}

func init() {
	Connectors["graphite"] = func(settings map[string]interface{}) (Connector, error) {
		var err error

		connector := &GraphiteConnector{
			insecureTLS: false,
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

		if connector.timeout <= 0 {
			connector.timeout = graphiteDefaultTimeout
		}

		return connector, nil
	}
}

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (connector *GraphiteConnector) GetPlots(query *types.PlotQuery) (map[string]*types.PlotResult, error) {
	var result map[string]*types.PlotResult

	if len(query.Group.Series) == 0 {
		return nil, fmt.Errorf("group has no series")
	} else if query.Group.Type != OperGroupTypeNone && len(query.Group.Series) == 1 {
		query.Group.Type = OperGroupTypeNone
	}

	queryURL, err := graphiteBuildQueryURL(query.Group, query.StartTime, query.EndTime)
	if err != nil {
		return nil, fmt.Errorf("unable to build Graphite query URL: %s", err.Error())
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

	request, err := http.NewRequest("GET", strings.TrimSuffix(connector.URL, "/")+queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "GraphiteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %s", err)
	}

	if err = graphiteCheckBackendResponse(response); err != nil {
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

	if result, err = graphiteExtractPlotResult(graphitePlots); err != nil {
		return nil, fmt.Errorf("unable to extract plot values from backend response: %s", err)
	}

	return result, nil
}

// Refresh triggers a full connector data update.
func (connector *GraphiteConnector) Refresh(originName string, outputChan chan *catalog.CatalogRecord) error {
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
		return fmt.Errorf("unable to set up HTTP request: %s", err)
	}

	request.Header.Add("User-Agent", "Facette")
	request.Header.Add("X-Requested-With", "GraphiteConnector")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %s", err)
	}

	if err = graphiteCheckBackendResponse(response); err != nil {
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
			sourceName = "unknown"
			metricName = metric
		} else {
			sourceName = metric[0:index]
			metricName = metric[index+1:]
		}

		outputChan <- &catalog.CatalogRecord{
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

func graphiteBuildQueryURL(query *types.GroupQuery, startTime, endTime time.Time) (string, error) {
	var (
		serieName string
		target    string
	)

	now := time.Now()

	fromTime := 0

	queryURL := fmt.Sprintf("%s?format=json", graphiteURLRender)

	if query.Type == OperGroupTypeNone {
		for _, serie := range query.Series {
			target := fmt.Sprintf("%s.%s", serie.Metric.Source, serie.Metric.Name)

			if serie.Scale != 0 {
				target = fmt.Sprintf("scale(%s, %f)", target, serie.Scale)
			}

			queryURL += fmt.Sprintf(
				"&target=legendValue(alias(%s, '%s'), 'min', 'max', 'avg', 'last')",
				target,
				serie.Name,
			)
		}

	} else {
		serieName = query.Name
		targets := make([]string, 0)

		for _, serie := range query.Series {
			targets = append(targets, fmt.Sprintf("%s.%s", serie.Metric.Source, serie.Metric.Name))
		}

		target = fmt.Sprintf("group(%s)", strings.Join(targets, ","))

		if query.Series[0].Scale != 0 {
			target = fmt.Sprintf("scale(%s, %f)", target, query.Series[0].Scale)
		}

		switch query.Type {
		case OperGroupTypeAvg:
			target = fmt.Sprintf("averageSeries(%s)", target)
		case OperGroupTypeSum:
			target = fmt.Sprintf("sumSeries(%s)", target)
		}

		target = fmt.Sprintf(
			"legendValue(alias(%s, '%s'), 'min', 'max', 'avg', 'last')",
			target,
			serieName,
		)

		queryURL += fmt.Sprintf("&target=%s", target)
	}

	if startTime.Before(now) {
		fromTime = int(now.Sub(startTime).Seconds())
	}

	queryURL += fmt.Sprintf("&from=-%ds", fromTime)

	// Only specify `until' parameter if endTime is still in the past
	if endTime.Before(now) {
		untilTime := int(time.Now().Sub(endTime).Seconds())
		queryURL += fmt.Sprintf("&until=-%ds", untilTime)
	}

	return queryURL, nil
}

func graphiteExtractPlotResult(plots []graphitePlot) (map[string]*types.PlotResult, error) {
	var (
		serieName           string
		min, max, avg, last float64
	)

	results := make(map[string]*types.PlotResult)

	for _, plot := range plots {
		result := &types.PlotResult{Info: make(map[string]types.PlotValue)}

		for _, plotPoint := range plot.Datapoints {
			result.Plots = append(result.Plots, types.PlotValue(plotPoint[0]))
		}

		// Scan the target legend for serie name and plot min/max/avg/last info
		if index := strings.Index(plots[0].Target, "(min"); index > 0 {
			fmt.Sscanf(plot.Target[0:index], "%s ", &serieName)
			fmt.Sscanf(plot.Target[index:], "(min: %f) (max: %f) (avg: %f) (last: %f)", &min, &max, &avg, &last)
		}

		result.Info["min"] = types.PlotValue(min)
		result.Info["max"] = types.PlotValue(max)
		result.Info["avg"] = types.PlotValue(avg)
		result.Info["last"] = types.PlotValue(last)

		results[serieName] = result
	}

	return results, nil
}
