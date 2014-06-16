package types

import "time"

// PlotQuery represents a connector plot query.
type PlotQuery struct {
	Group       *GroupQuery
	StartTime   time.Time
	EndTime     time.Time
	Step        time.Duration
	Percentiles []float64
}

// MetricQuery represents a metric entry in a SerieQuery.
type MetricQuery struct {
	Name   string
	Origin string
	Source string
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
	Plots []PlotValue
	Info  map[string]PlotValue
}
