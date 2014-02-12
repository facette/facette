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
)

// GraphiteBackendHandler represents the main structure of the Graphite backend.
type GraphiteBackendHandler struct {
	URL                  string
	AllowBadCertificates bool
	origin               *Origin
}

// GetPlots calculates and returns plot data based on a time interval.
func (handler *GraphiteBackendHandler) GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64) (map[string]*PlotResult, error) {

	return nil, nil
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

	if res.StatusCode != 200 {
		return fmt.Errorf("invalid HTTP status code (%d), expecting 200", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("invalid HTTP response content type (%s), expecting \"application/json\"",
			res.Header["Content-Type"])
	}

	if data, err = ioutil.ReadAll(res.Body); err != nil {
		return fmt.Errorf("unable to read HTTP response body: %s", err)
	}

	if err = json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unable to unmarshall JSON data: %s", err)
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
