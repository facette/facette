// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package catalog

import (
	"sort"

	"facette.io/facette/pkg/labels"
)

// Section is a metrics catalog section.
type Section struct {
	connector interface{}
	metrics   []Metric
}

// NewSection creates a new metrics catalog section instance.
func NewSection(connector interface{}) *Section {
	return &Section{
		connector: connector,
		metrics:   make([]Metric, 0),
	}
}

// Insert inserts a new metric into the metrics catalog section.
func (s *Section) Insert(metric Metric) error {
	sort.Sort(metric.Labels)

	err := metric.Labels.Validate()
	if err != nil {
		return err
	}

	metric.section = s
	s.metrics = append(s.metrics, metric)

	return nil
}

// Query returns all metrics matching the given labels matcher from the catalog
// section.
func (s *Section) Query(matcher labels.Matcher) []Metric {
	if len(matcher) == 0 {
		return append([]Metric(nil), s.metrics...)
	}

	result := []Metric{}

	for _, metric := range s.metrics {
		if metric.Labels.Match(matcher) {
			result = append(result, metric)
		}
	}

	return result
}
