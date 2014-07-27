// Package plot provides plot-related types and methods.
package plot

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	_ = iota
	// ConsolidateAverage represents an average consolidation type.
	ConsolidateAverage
	// ConsolidateMax represents a maximal value consolidation type.
	ConsolidateMax
	// ConsolidateMin represents a minimal value consolidation type.
	ConsolidateMin
	// ConsolidateSum represents a sum consolidation type.
	ConsolidateSum
)

// Plot represents a graph plot.
type Plot struct {
	Time  time.Time `json:"time"`
	Value Value     `json:"value"`
}

// MarshalJSON handles JSON marshalling of the Plot type.
func (plot Plot) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]interface{}{int(plot.Time.Unix()), plot.Value})
}

// UnmarshalJSON handles JSON marshalling of the Plot type.
func (plot *Plot) UnmarshalJSON(data []byte) error {
	var input [2]float64

	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	plot.Time = time.Unix(int64(input[0]), 0)
	plot.Value = Value(input[1])

	return nil
}

// Value represents a graph plot value.
type Value float64

// MarshalJSON handles JSON marshalling of the Value type.
func (value Value) MarshalJSON() ([]byte, error) {
	// Handle NaN and near-zero values marshalling
	if math.IsNaN(float64(value)) {
		return json.Marshal(nil)
	} else if math.Exp(float64(value)) == 1 {
		return json.Marshal(0)
	}

	return json.Marshal(float64(value))
}

// IsNaN reports whether the Value is an IEEE 754 “not-a-number” value.
func (value Value) IsNaN() bool {
	return math.IsNaN(float64(value))
}

// Query represents a plot query to a connector.
type Query struct {
	Group     *QueryGroup
	StartTime time.Time
	EndTime   time.Time
	Sample    int
}

func (plotQuery *Query) String() string {
	return fmt.Sprintf(
		"Query{StartTime:%s EndTime:%s Sample:%d Group:%s}",
		plotQuery.StartTime.String(),
		plotQuery.EndTime.String(),
		plotQuery.Sample,
		plotQuery.Group,
	)
}

// QueryGroup represents a plot query operation group.
type QueryGroup struct {
	Type    int
	Series  []*QuerySeries
	Options map[string]interface{}
}

func (queryGroup *QueryGroup) String() string {
	return fmt.Sprintf(
		"QueryGroup{Type:%d Scale:%g Series:[%s] Options:%v}",
		queryGroup.Type,
		func(series []*QuerySeries) string {
			seriesStrings := make([]string, len(series))
			for i, entry := range series {
				seriesStrings[i] = fmt.Sprintf("%s", entry)
			}

			return strings.Join(seriesStrings, ", ")
		}(queryGroup.Series),
		queryGroup.Options,
	)
}

// QuerySeries represents a series entry in a QueryGroup.
type QuerySeries struct {
	Metric  *QueryMetric
	Options map[string]interface{}
}

func (QuerySeries *QuerySeries) String() string {
	return fmt.Sprintf(
		"QuerySeries{Metric:%s Options:%v}",
		QuerySeries.Metric,
		QuerySeries.Options,
	)
}

// QueryMetric represents a metric entry in a QuerySeries.
type QueryMetric struct {
	Name   string
	Origin string
	Source string
}

func (queryMetric *QueryMetric) String() string {
	return fmt.Sprintf(
		"QueryMetric{Name:\"%s\" Source:\"%s\" Origin:\"%s\"}",
		queryMetric.Name,
		queryMetric.Source,
		queryMetric.Origin,
	)
}

// Series represents a series of plots.
type Series struct {
	Name    string
	Plots   []Plot
	Step    int
	Summary map[string]Value
}

// Consolidate consolidates a series of plots given a certain number of values per point.
func (series *Series) Consolidate(pad, consolidationType int) {
	if pad < 2 {
		return
	}

	plotsCount := len(series.Plots)

	plots := series.Plots[:]
	series.Plots = make([]Plot, 0)

	bucket := []float64{}

	for i := 0; i < plotsCount; i++ {
		if !plots[i].Value.IsNaN() {
			bucket = append(bucket, float64(plots[i].Value))
		}

		if (i+1)%pad == 0 {
			if len(bucket) == 0 {
				series.Plots = append(
					series.Plots,
					Plot{Value: Value(math.NaN()), Time: plots[i].Time},
				)

				continue
			}

			series.Plots = append(
				series.Plots,
				Plot{Value: consolidateBucket(bucket, consolidationType), Time: plots[i].Time},
			)

			bucket = make([]float64, 0)
		}
	}

	if len(bucket) > 0 {
		series.Plots = append(
			series.Plots,
			Plot{Value: consolidateBucket(bucket, consolidationType), Time: plots[plotsCount-1].Time},
		)
	}
}

// Downsample applies a sampling function on a series of plots, reducing the number of points.
func (series *Series) Downsample(sample, consolidationType int) {
	series.Consolidate(len(series.Plots)/sample, consolidationType)
}

// Summarize calculates the min/max/average/last and percentile values of a series of plots, and stores the results
// into the Summary map.
func (series Series) Summarize(percentiles []float64) {
	var (
		min, max, total Value
		nValidPlots     int64
		nPlots          = len(series.Plots)
	)

	if nPlots > 0 {
		min = series.Plots[0].Value
		series.Summary["last"] = series.Plots[nPlots-1].Value
	}

	for i := range series.Plots {
		if !series.Plots[i].Value.IsNaN() && series.Plots[i].Value < min || min.IsNaN() {
			min = series.Plots[i].Value
		}

		if series.Plots[i].Value > max {
			max = series.Plots[i].Value
		}

		if !series.Plots[i].Value.IsNaN() {
			total += series.Plots[i].Value
			nValidPlots++
		}
	}

	series.Summary["min"] = min
	series.Summary["max"] = max
	series.Summary["avg"] = total / Value(nValidPlots)

	if len(percentiles) > 0 {
		series.Percentiles(percentiles)
	}
}

// Percentiles calculates the percentile values of a series of plots.
func (series Series) Percentiles(percentiles []float64) {
	var set []float64

	for i := range series.Plots {
		if !series.Plots[i].Value.IsNaN() {
			set = append(set, float64(series.Plots[i].Value))
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
			series.Summary[percentileString] = Value(set[0])
			continue
		} else if rank-1.0 >= float64(setSize) {
			series.Summary[percentileString] = Value(set[setSize-1])
			continue
		}

		series.Summary[percentileString] = Value(set[rankInt-1] + rankFrac*(set[rankInt]-set[rankInt-1]))
	}
}

func consolidateBucket(bucket []float64, consolidationType int) Value {
	switch consolidationType {
	case ConsolidateAverage, ConsolidateSum:
		sum := 0.0
		for _, entry := range bucket {
			sum += entry
		}

		if consolidationType == ConsolidateAverage {
			return Value(sum / float64(len(bucket)))
		} else {
			return Value(sum)
		}
	case ConsolidateMax:
		max := math.NaN()
		for _, entry := range bucket {
			if !math.IsNaN(entry) && entry > max || math.IsNaN(max) {
				max = entry
			}
		}

		return Value(max)
	case ConsolidateMin:
		min := math.NaN()
		for _, entry := range bucket {
			if !math.IsNaN(entry) && entry < min || math.IsNaN(min) {
				min = entry
			}
		}

		return Value(min)
	}

	return Value(math.NaN())
}
