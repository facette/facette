package types

import (
	"fmt"
	"sort"
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
	Type   int
	Series []*SerieQuery
	Scale  float64
}

// SerieQuery represents a serie entry in a GroupQuery.
type SerieQuery struct {
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
	Name  string
	Plots []PlotValue
	Info  map[string]PlotValue
}

// Percentiles calculates the percentile values of a PlotResult plots
func (plotResult PlotResult) Percentiles(percentiles []float64) {
	set := make([]float64, len(plotResult.Plots))
	for i, _ := range plotResult.Plots {
		set[i] = float64(plotResult.Plots[i])
	}

	if len(percentiles) == 0 {
		return
	}

	setSize := len(plotResult.Plots)
	if setSize == 0 {
		return
	}

	sort.Float64s(set)

	for _, percentile := range percentiles {
		percentileString := fmt.Sprintf("%gth", percentile)

		rank := (percentile / 100) * float64(setSize+1)
		rankInt := int(rank)
		rankFrac := rank - float64(rankInt)

		if rank <= 0.0 {
			plotResult.Info[percentileString] = PlotValue(set[0])
			continue
		} else if rank-1.0 >= float64(setSize) {
			plotResult.Info[percentileString] = PlotValue(set[setSize-1])
			continue
		}

		plotResult.Info[percentileString] = PlotValue(set[rankInt-1] + rankFrac*(set[rankInt]-set[rankInt-1]))
	}
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
		"GroupQuery{Type:%d Scale:%g Series:[%s]}",
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
		"SerieQuery{Scale:%g Metric:%s}",
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
