// Package connector implements the connectors handling third-party data sources.
package connector

import (
	"time"

	"github.com/facette/facette/pkg/types"
)

// Connector represents the main interface of a connector handler.
type Connector interface {
	GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
		percentiles []float64) (map[string]*PlotResult, error)
	Refresh(chan error)
}

// MetricQuery represents a metric entry in a SerieQuery.
type MetricQuery struct {
	Name       string
	SourceName string
}

// SerieQuery represents a serie entry in a GroupQuery.
type SerieQuery struct {
	Name   string
	Metric *MetricQuery
	Scale  float64
}

// GroupQuery represents a plot group query.
type GroupQuery struct {
	Name   string
	Type   int
	Series []*SerieQuery
	Scale  float64
}

// PlotResult represents a plot request result.
type PlotResult struct {
	Plots []types.PlotValue
	Info  map[string]types.PlotValue
}

const (
	_ = iota
	// OperGroupTypeNone represents a null operation group mode.
	OperGroupTypeNone
	// OperGroupTypeAvg represents a AVG operation group mode.
	OperGroupTypeAvg
	// OperGroupTypeSum represents a SUM operation group mode.
	OperGroupTypeSum
)

var (
	// Connectors represents the list of all available connector handlers.
	Connectors = make(map[string]func(*chan [2]string, map[string]string) (interface{}, error))
)
