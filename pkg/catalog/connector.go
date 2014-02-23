package catalog

import (
	"time"

	"github.com/facette/facette/pkg/types"
)

const (
	// OperGroupTypeNone represents a null operation group mode.
	OperGroupTypeNone = iota
	// OperGroupTypeAvg represents a AVG operation group mode.
	OperGroupTypeAvg
	// OperGroupTypeSum represents a SUM operation group mode.
	OperGroupTypeSum
)

var (
	// ConnectorHandlers represents the list of available connector handlers.
	ConnectorHandlers = make(map[string]func(*Origin, map[string]string) error)
)

// SerieQuery represents a serie entry in a GroupQuery.
type SerieQuery struct {
	Name   string
	Metric *Metric
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

// ConnectorHandler represents the main interface of connector handlers.
type ConnectorHandler interface {
	GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
		percentiles []float64) (map[string]*PlotResult, error)
	GetValue(query *GroupQuery, refTime time.Time,
		percentiles []float64) (map[string]map[string]types.PlotValue, error)
	Update() error
}
