package types

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// PlotQuery represents a plot query to a connector.
type PlotQuery struct {
	Group     *PlotQueryGroup
	StartTime time.Time
	EndTime   time.Time
	Step      time.Duration
}

func (plotQuery *PlotQuery) String() string {
	return fmt.Sprintf(
		"PlotQuery{StartTime:%s EndTime:%s Step:%s Group:%s}",
		plotQuery.StartTime.String(),
		plotQuery.EndTime.String(),
		plotQuery.Step.String(),
		plotQuery.Group,
	)
}

// PlotQueryGroup represents a plot query operation group.
type PlotQueryGroup struct {
	Type    int
	Series  []*PlotQuerySerie
	Options map[string]interface{}
}

func (queryGroup *PlotQueryGroup) String() string {
	return fmt.Sprintf(
		"PlotQueryGroup{Type:%d Scale:%g Series:[%s] Options:%v}",
		queryGroup.Type,
		func(series []*PlotQuerySerie) string {
			seriesStrings := make([]string, len(series))

			for i, serie := range series {
				seriesStrings[i] = fmt.Sprintf("%s", serie)
			}

			return strings.Join(seriesStrings, ", ")
		}(queryGroup.Series),
		queryGroup.Options,
	)
}

// PlotQuerySerie represents a serie entry in a PlotQueryGroup.
type PlotQuerySerie struct {
	Metric  *PlotQueryMetric
	Options map[string]interface{}
}

func (querySerie *PlotQuerySerie) String() string {
	return fmt.Sprintf(
		"PlotQuerySerie{Metric:%s Options:%v}",
		querySerie.Metric,
		querySerie.Options,
	)
}

// PlotQueryMetric represents a metric entry in a PlotQuerySerie.
type PlotQueryMetric struct {
	Name   string
	Origin string
	Source string
}

func (queryMetric *PlotQueryMetric) String() string {
	return fmt.Sprintf(
		"PlotQueryMetric{Name:\"%s\" Source:\"%s\" Origin:\"%s\"}",
		queryMetric.Name,
		queryMetric.Source,
		queryMetric.Origin,
	)
}

// PlotResult represents the result of a plot request.
type PlotResult struct {
	Name  string
	Plots []PlotValue
	Info  map[string]PlotValue
}

// Summarize calculates the min/max/average/last and percentile values of a PlotResult plots, and stores the results
// into the Info map.
func (plotResult PlotResult) Summarize(percentiles []float64) {
	var (
		min, max, total PlotValue
		nValidPlots     int64
		nPlots          = len(plotResult.Plots)
	)

	if nPlots > 0 {
		min = plotResult.Plots[0]
		plotResult.Info["last"] = plotResult.Plots[nPlots-1]
	}

	for i := range plotResult.Plots {
		if !plotResult.Plots[i].IsNaN() && plotResult.Plots[i] < min || min.IsNaN() {
			min = plotResult.Plots[i]
		}

		if plotResult.Plots[i] > max {
			max = plotResult.Plots[i]
		}

		if !plotResult.Plots[i].IsNaN() {
			total += plotResult.Plots[i]
			nValidPlots++
		}
	}

	plotResult.Info["min"] = min
	plotResult.Info["max"] = max
	plotResult.Info["avg"] = total / PlotValue(nValidPlots)

	if len(percentiles) > 0 {
		plotResult.Percentiles(percentiles)
	}
}

// Percentiles calculates the percentile values of a PlotResult plots
func (plotResult PlotResult) Percentiles(percentiles []float64) {
	set := make([]float64, 0)
	for i := range plotResult.Plots {
		if !plotResult.Plots[i].IsNaN() {
			set = append(set, float64(plotResult.Plots[i]))
		}
	}

	if len(percentiles) == 0 {
		return
	}

	setSize := len(set)
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

// PlotValue represents a graph plot value.
type PlotValue float64

// MarshalJSON handles JSON marshalling of the PlotValue type.
func (value PlotValue) MarshalJSON() ([]byte, error) {
	// Handle NaN and near-zero values marshalling
	if math.IsNaN(float64(value)) {
		return json.Marshal(nil)
	} else if math.Exp(float64(value)) == 1 {
		return json.Marshal(0)
	}

	return json.Marshal(float64(value))
}

// IsNaN reports whether the PlotValue is an IEEE 754 “not-a-number” value.
func (value PlotValue) IsNaN() bool {
	return math.IsNaN(float64(value))
}
