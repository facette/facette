// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package catalog provides the metrics catalog system.
package catalog

import (
	"sort"
	"strings"
	"sync"

	"facette.io/facette/pkg/labels"
	"facette.io/facette/pkg/set"
)

// Catalog is a metrics catalog.
type Catalog struct {
	sections map[string]*Section
	l        sync.RWMutex
}

// New creates a new metrics catalog instance.
func New() *Catalog {
	return &Catalog{
		sections: make(map[string]*Section),
	}
}

// Link links a section to the metrics catalog.
func (c *Catalog) Link(name string, section *Section) {
	c.l.Lock()
	defer c.l.Unlock()

	c.sections[name] = section
}

// Labels returns all labels names matching the given labels matcher from the
// catalog.
func (c *Catalog) Labels(matcher labels.Matcher, filter string) []string {
	ls := set.New()

	for _, metric := range c.Metrics(matcher) {
		for _, label := range metric.Labels {
			if filter == "" || strings.Contains(label.Name, filter) {
				ls.Add(label.Name)
			}
		}
	}

	result := set.StringSlice(ls)
	sort.Strings(result)

	return result
}

// Metrics returns all metrics matching the given labels matcher from the
// catalog.
func (c *Catalog) Metrics(matcher labels.Matcher) []Metric {
	metrics := []Metric{}
	for _, section := range c.sections {
		metrics = append(metrics, section.Query(matcher)...)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].String() < metrics[j].String()
	})

	return metrics
}

// Unlink unlinks a section from the metrics catalog.
func (c *Catalog) Unlink(name string) {
	c.l.Lock()
	defer c.l.Unlock()

	delete(c.sections, name)
}

// Values returns all metrics values matching the given label name and labels
// matcher.
func (c *Catalog) Values(label string, matcher labels.Matcher, filter string) []string {
	values := set.New()

	for _, metric := range c.Metrics(matcher) {
		for _, l := range metric.Labels {
			if l.Name == label && (filter == "" || strings.Contains(l.Value, filter)) {
				values.Add(l.Value)
			}
		}
	}

	result := set.StringSlice(values)
	sort.Strings(result)

	return result
}
