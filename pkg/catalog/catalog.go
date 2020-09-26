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

	sort.Slice(section.metrics, func(i, j int) bool {
		return section.metrics[i].String() < section.metrics[j].String()
	})

	c.sections[name] = section
}

// Labels returns all labels names matching the given labels matcher from the
// catalog.
func (c *Catalog) Labels(opts *ListOptions) []string {
	s := set.New()

	for _, metric := range c.Metrics(opts) {
		for _, label := range metric.Labels {
			if opts.Filter == "" || strings.Contains(label.Name, opts.Filter) {
				s.Add(label.Name)
			}
		}
	}

	labels := set.StringSlice(s)
	sort.Slice(labels, func(i, j int) bool {
		return strings.ToLower(labels[i]) < strings.ToLower(labels[j])
	})

	return labels
}

// Metrics returns all metrics matching the given labels matcher from the
// catalog.
func (c *Catalog) Metrics(opts *ListOptions) []Metric {
	metrics := []Metric{}
	for _, section := range c.sections {
		metrics = append(metrics, section.Query(opts.Matcher)...)
	}

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
func (c *Catalog) Values(label string, opts *ListOptions) []string {
	s := set.New()

	for _, metric := range c.Metrics(opts) {
		for _, l := range metric.Labels {
			if l.Name == label && (opts.Filter == "" || strings.Contains(l.Value, opts.Filter)) {
				s.Add(l.Value)
			}
		}
	}

	values := set.StringSlice(s)
	sort.Strings(values)

	return values
}

// ListOptions are catalog listing options.
type ListOptions struct {
	Filter  string
	Matcher labels.Matcher
}
