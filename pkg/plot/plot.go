// Package plot provides plot-related types and methods.
package plot

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
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

func (plot *Plot) String() string {
	return fmt.Sprintf("Plot{Time:%s Value:%g}", plot.Time, plot.Value)
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
	Requestor string
	StartTime time.Time
	EndTime   time.Time
	Sample    int
	Series    []QuerySeries
}

func (query *Query) String() string {
	return fmt.Sprintf(
		"Query{StartTime:%s EndTime:%s Sample:%d Series:%s}",
		query.StartTime,
		query.EndTime,
		query.Sample,
		query.Series,
	)
}

// QuerySeries represents a series entry in a Query.
type QuerySeries struct {
	Name   string
	Origin string
	Source string
	Metric string
}

func (metric *QuerySeries) String() string {
	return fmt.Sprintf(
		"QuerySeries{Name:%q Source:%q Origin:%q}",
		metric.Name,
		metric.Source,
		metric.Origin,
	)
}

// Series represents a series of plots.
type Series struct {
	Name    string
	Plots   []Plot
	Step    int
	Summary map[string]Value
}

// Scale applies a factor on a series of plots.
func (series *Series) Scale(factor Value) {
	for i := range series.Plots {
		if !series.Plots[i].Value.IsNaN() {
			series.Plots[i].Value *= factor
		}
	}
}

// Summarize calculates the min/max/average/last and percentile values of a series of plots, and stores the results
// into the Summary map.
func (series *Series) Summarize(percentiles []float64) {
	var (
		min, max, total, current Value
		nValidPlots              int64
	)

	min = Value(math.NaN())
	max = Value(math.NaN())

	for i := range series.Plots {
		if !series.Plots[i].Value.IsNaN() {
			current = series.Plots[i].Value
			if current < min || min.IsNaN() {
				min = series.Plots[i].Value
			}
			if current > max || max.IsNaN() {
				max = current
			}

			total += current
			nValidPlots++
		}
	}

	if series.Summary == nil {
		series.Summary = make(map[string]Value)
	}

	series.Summary["min"] = min
	series.Summary["max"] = max
	series.Summary["avg"] = total / Value(nValidPlots)
	series.Summary["last"] = current

	if len(percentiles) > 0 {
		series.Percentiles(percentiles)
	}
}

// Percentiles calculates the percentile values of a series of plots.
func (series *Series) Percentiles(percentiles []float64) {
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
