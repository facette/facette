package backend

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

// GraphiteBackendHandler represents the main structure of the Graphite backend.
type GraphiteBackendHandler struct {
	URL                  string
	AllowBadCertificates bool
	origin               *Origin
}

type graphitePlot struct {
	Target     string
	Datapoints [][2]float64
}

// GetPlots calculates and returns plot data based on a time interval.
func (handler *GraphiteBackendHandler) GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64) (map[string]*PlotResult, error) {
	var (
		data             []byte
		err              error
		graphitePlots    []graphitePlot
		graphiteQueryURL string
		httpClient       http.Client
		httpTransport    http.RoundTripper
		pr               map[string]*PlotResult
		queryURL         string
		target           string
		res              *http.Response
	)

	if handler.AllowBadCertificates {
		httpTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		httpTransport = &http.Transport{}
	}

	httpClient = http.Client{Transport: httpTransport}

	pr = make(map[string]*PlotResult)

	for _, s := range query.Series {
		if s.Metric == nil {
			continue
		}

		target = fmt.Sprintf("%s.%s", s.Metric.source.OriginalName, s.Metric.OriginalName)

		if queryURL, err = graphiteBuildQueryURL(target, startTime, endTime); err != nil {
			return nil, fmt.Errorf("unable to build Graphite query URL: %s", err)
		}

		graphiteQueryURL = fmt.Sprintf("%s%s", strings.TrimSuffix(handler.URL, "/"), queryURL)

		if res, err = httpClient.Get(graphiteQueryURL); err != nil {
			return nil, err
		}

		if err = graphiteCheckBackendResponse(res); err != nil {
			return nil, fmt.Errorf("invalid HTTP backend response: %s", err)
		}

		if data, err = ioutil.ReadAll(res.Body); err != nil {
			return nil, fmt.Errorf("unable to read HTTP response body: %s", err)
		}

		if err = json.Unmarshal(data, &graphitePlots); err != nil {
			return nil, fmt.Errorf("unable to unmarshal JSON data: %s", err)
		}

		if pr[s.Name], err = graphiteExtractPlotResult(graphitePlots); err != nil {
			return nil, fmt.Errorf("unable to extract plot values from backend response: %s", err)
		}
	}

	return pr, nil
}

// GetValue calculates and returns plot data at a specific reference time.
func (handler *GraphiteBackendHandler) GetValue(query *GroupQuery, refTime time.Time,
	percentiles []float64) (map[string]map[string]types.PlotValue, error) {

	return nil, nil
}

// Update triggers a full backend data update.
func (handler *GraphiteBackendHandler) Update() error {
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

	if err = graphiteCheckBackendResponse(res); err != nil {
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

	return nil
}

func init() {
	BackendHandlers["graphite"] = NewGraphiteBackendHandler
}

func graphiteCheckBackendResponse(res *http.Response) error {
	if res.StatusCode != 200 {
		return fmt.Errorf("got HTTP status code %d, expected 200", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("got HTTP content type \"%s\", expected \"application/json\"",
			res.Header["Content-Type"])
	}

	return nil
}

func graphiteBuildQueryURL(target string, startTime, endTime time.Time) (string, error) {
	var (
		fromTime         int
		untilTime        int
		graphiteQueryURL string
	)

	fromTime = int(time.Now().Sub(startTime).Seconds())
	untilTime = int(time.Now().Sub(endTime).Seconds())

	graphiteQueryURL = fmt.Sprintf("%s?format=json&target=%s&from=-%ds&until=-%ds",
		graphiteRenderURL,
		target,
		fromTime,
		untilTime)

	return graphiteQueryURL, nil
}

func graphiteExtractPlotResult(graphitePlots []graphitePlot) (*PlotResult, error) {
	var (
		pr *PlotResult
	)

	pr = &PlotResult{}

	for _, plotPoint := range graphitePlots[0].Datapoints {
		pr.Plots = append(pr.Plots, types.PlotValue(plotPoint[0]))
	}

	return pr, nil
}

// NewGraphiteBackendHandler creates a new instance of BackendHandler.
func NewGraphiteBackendHandler(origin *Origin, config map[string]string) error {
	var (
		graphiteBackend *GraphiteBackendHandler
	)

	if _, present := config["url"]; !present {
		return fmt.Errorf("missing `url' mandatory backend definition")
	}

	graphiteBackend = &GraphiteBackendHandler{
		URL:                  config["url"],
		AllowBadCertificates: false,
		origin:               origin,
	}

	if config["allow_bad_certificates"] == "yes" {
		graphiteBackend.AllowBadCertificates = true
	}

	origin.Backend = graphiteBackend

	return nil
}
