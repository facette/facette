package types

import (
	"fmt"
	"strings"
	"time"
)

// PlotQuery represents a connector plot query.
type PlotQuery struct {
	Group       *GroupQuery
	StartTime   time.Time
	EndTime     time.Time
	Step        time.Duration
	Percentiles []float64
}

// GroupQuery represents a plot group query.
type GroupQuery struct {
	Name   string
	Type   int
	Series []*SerieQuery
	Scale  float64
}

// SerieQuery represents a serie entry in a GroupQuery.
type SerieQuery struct {
	Name   string
	Metric *MetricQuery
	Scale  float64
}

// MetricQuery represents a metric entry in a SerieQuery.
type MetricQuery struct {
	Name   string
	Origin string
	Source string
}

// PlotResult represents a plot request result.
type PlotResult struct {
	Plots []PlotValue
	Info  map[string]PlotValue
}

func (plotQuery *PlotQuery) String() string {
	return fmt.Sprintf(
		"PlotQuery{StartTime:%s EndTime:%s Step:%s Percentiles:[%s] Group:%s}",
		plotQuery.StartTime.String(),
		plotQuery.EndTime.String(),
		plotQuery.Step.String(),
		func(percentiles []float64) string {
			percentilesStrings := make([]string, len(percentiles))

			for i, percentile := range percentiles {
				percentilesStrings[i] = fmt.Sprintf("%s", percentile)
			}

			return strings.Join(percentilesStrings, ", ")
		}(plotQuery.Percentiles),
		plotQuery.Group,
	)
}

func (groupQuery *GroupQuery) String() string {
	return fmt.Sprintf(
		"GroupQuery{Name:\"%s\" Type:%d Scale:%g Series:[%s]}",
		groupQuery.Name,
		groupQuery.Type,
		groupQuery.Scale,
		func(series []*SerieQuery) string {
			seriesStrings := make([]string, len(series))

			for i, serie := range series {
				seriesStrings[i] = fmt.Sprintf("%s", serie)
			}

			return strings.Join(seriesStrings, ", ")
		}(groupQuery.Series),
	)
}

func (serieQuery *SerieQuery) String() string {
	return fmt.Sprintf(
		"SerieQuery{Name:\"%s\" Scale:%g Metric:%s}",
		serieQuery.Name,
		serieQuery.Scale,
		serieQuery.Metric,
	)
}

func (metricQuery *MetricQuery) String() string {
	return fmt.Sprintf(
		"MetricQuery{Name:\"%s\" Source:\"%s\" Origin:\"%s\"}",
		metricQuery.Name,
		metricQuery.Source,
		metricQuery.Origin,
	)
}
