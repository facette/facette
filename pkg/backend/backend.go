package backend

import (
	"time"

	"github.com/facette/facette/pkg/common"
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
	// BackendHandlers represents the list of available backend handlers.
	BackendHandlers = make(map[string]func(*Origin, map[string]string) error)
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
	Plots []common.PlotValue
	Info  map[string]common.PlotValue
}

// BackendHandler represents the main interface of backend handlers.
type BackendHandler interface {
	GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
		percentiles []float64) (map[string]*PlotResult, error)
	GetValue(query *GroupQuery, refTime time.Time,
		percentiles []float64) (map[string]map[string]common.PlotValue, error)
	Update() error
}
