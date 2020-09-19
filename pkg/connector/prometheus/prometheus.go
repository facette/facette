// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package prometheus provides a Prometheus time series connector.
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	prometheus "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	promparser "github.com/prometheus/prometheus/promql/parser"

	"facette.io/facette/pkg/catalog"
	"facette.io/facette/pkg/connector"
	httpclient "facette.io/facette/pkg/http/client"
	"facette.io/facette/pkg/labels"
	"facette.io/facette/pkg/series"
)

const defaultFilter = `{job=~".+"}`

// Connector is a Prometheus time series connector.
type Connector struct {
	name     string
	settings Settings
	api      prometheusv1.API
}

// New creates a new Prometheus time series connector.
func New(name string, settings json.RawMessage) (connector.Connector, error) {
	c := &Connector{
		name: name,
		settings: Settings{
			Filter: defaultFilter,
		},
	}

	err := json.Unmarshal(settings, &c.settings)
	if err != nil {
		return nil, err
	}

	client, err := prometheus.NewClient(prometheus.Config{
		Address:      c.settings.URL,
		RoundTripper: httpclient.NewRoundTripper(c.settings.SkipVerify),
	})
	if err != nil {
		return nil, err
	}

	c.api = prometheusv1.NewAPI(client)

	return c, nil
}

// Metrics retrieves metrics from the upstream service.
func (c *Connector) Metrics(ctx context.Context, ch chan<- catalog.Metric, errCh chan<- error) {
	defer func() {
		close(ch)
		close(errCh)
	}()

	value, _, err := c.api.Query(ctx, c.settings.Filter, time.Now().UTC())
	if err != nil {
		errCh <- err
		return
	}

	vector, ok := value.(prommodel.Vector)
	if !ok {
		errCh <- fmt.Errorf("unsupported response type: %T", value)
		return
	}

	for _, sample := range vector {
		ls := labels.New(labels.Label{Name: labels.Provider, Value: c.name})

		for name, value := range sample.Metric {
			n, v := string(name), string(value)

			if n == prommodel.MetricNameLabel {
				ls.Append(labels.Label{Name: labels.Name, Value: v})
			} else if !strings.HasPrefix(n, prommodel.ReservedLabelPrefix) {
				ls.Append(labels.Label{Name: n, Value: v})
			}
		}

		ch <- catalog.Metric{
			Labels: ls,
			Attributes: catalog.Attributes{
				"metric": sample.Metric.String(),
			},
		}
	}
}

// Query retrieves Data points from the upstream service.
func (c *Connector) Query(ctx context.Context, q *connector.Query) (connector.Result, error) {
	result := connector.Result{}

	for _, metric := range q.Metrics {
		value, _, err := c.api.QueryRange(
			ctx,
			metric.Attributes["metric"].(string),
			prometheusv1.Range{
				Start: q.From,
				End:   q.To,
				Step:  q.Step,
			},
		)
		if err != nil {
			return nil, err
		}

		for _, stream := range value.(prommodel.Matrix) {
			sample := connector.Sample{Metric: metric}
			for _, value := range stream.Values {
				sample.Points = append(sample.Points, series.Point{
					Time:  value.Timestamp.Time(),
					Value: series.Value(value.Value),
				})
			}

			result = append(result, sample)
		}
	}

	return result, nil
}

// Test tests for validity of the time series connector.
func (c *Connector) Test(ctx context.Context) error {
	now := time.Now().UTC()

	_, _, err := c.api.LabelNames(ctx, now, now.Add(-1*time.Hour))
	if err != nil {
		return httpclient.Error(err)
	}

	if c.settings.Filter != "" {
		_, err = promparser.ParseExpr(c.settings.Filter)
		if err != nil {
			return fmt.Errorf("invalid filter: %s", c.settings.Filter)
		}
	}

	return nil
}

func init() {
	connector.Register("prometheus", New)
}
