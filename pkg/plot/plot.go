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

// Plot represents a graph plot.
type Plot struct {
	Time  time.Time
	Value Value
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
	Series  []*QuerySerie
	Options map[string]interface{}
}

func (queryGroup *QueryGroup) String() string {
	return fmt.Sprintf(
		"QueryGroup{Type:%d Scale:%g Series:[%s] Options:%v}",
		queryGroup.Type,
		func(series []*QuerySerie) string {
			seriesStrings := make([]string, len(series))

			for i, serie := range series {
				seriesStrings[i] = fmt.Sprintf("%s", serie)
			}

			return strings.Join(seriesStrings, ", ")
		}(queryGroup.Series),
		queryGroup.Options,
	)
}

// QuerySerie represents a serie entry in a QueryGroup.
type QuerySerie struct {
	Metric  *QueryMetric
	Options map[string]interface{}
}

func (querySerie *QuerySerie) String() string {
	return fmt.Sprintf(
		"QuerySerie{Metric:%s Options:%v}",
		querySerie.Metric,
		querySerie.Options,
	)
}

// QueryMetric represents a metric entry in a QuerySerie.
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

// Result represents the result of a plot request.
type Result struct {
	Name    string
	Plots   []Value
	Summary map[string]Value
}

// Downsample applies a sampling function on Result plots, reducing the number of points.
func (result *Result) Downsample(sample int) {
	if sample >= len(result.Plots) {
		return
	}

	plots := result.Plots[:]
	result.Plots = make([]Value, 0)

	pad := len(plots) / sample

	refinePad := 0
	if len(plots)%sample > 0 {
		refinePad = len(plots) / (len(plots) % sample)
	}

	padCount := 0

	bucket := 0.0
	bucketCount := 0.0

	for i := 0; i < len(plots); i++ {
		// Refine sampling by appending one more plot at regular interval (pad + 1)
		if refinePad == 0 || (i+1)%refinePad != 0 {
			padCount++
		}

		if !plots[i].IsNaN() {
			bucket += float64(plots[i])
			bucketCount++
		}

		if padCount == pad {
			padCount = 0

			if bucketCount == 0 {
				result.Plots = append(result.Plots, Value(math.NaN()))
				continue
			}

			result.Plots = append(result.Plots, Value(bucket/bucketCount))

			bucket = 0
			bucketCount = 0
		}
	}

	if bucketCount > 0 {
		result.Plots = append(result.Plots, Value(bucket/bucketCount))
	}
}

// Summarize calculates the min/max/average/last and percentile values of a Result plots, and stores the results
// into the Summary map.
func (result Result) Summarize(percentiles []float64) {
	var (
		min, max, total Value
		nValidPlots     int64
		nPlots          = len(result.Plots)
	)

	if nPlots > 0 {
		min = result.Plots[0]
		result.Summary["last"] = result.Plots[nPlots-1]
	}

	for i := range result.Plots {
		if !result.Plots[i].IsNaN() && result.Plots[i] < min || min.IsNaN() {
			min = result.Plots[i]
		}

		if result.Plots[i] > max {
			max = result.Plots[i]
		}

		if !result.Plots[i].IsNaN() {
			total += result.Plots[i]
			nValidPlots++
		}
	}

	result.Summary["min"] = min
	result.Summary["max"] = max
	result.Summary["avg"] = total / Value(nValidPlots)

	if len(percentiles) > 0 {
		result.Percentiles(percentiles)
	}
}

// Percentiles calculates the percentile values of a Result plots
func (result Result) Percentiles(percentiles []float64) {
	var set []float64

	for i := range result.Plots {
		if !result.Plots[i].IsNaN() {
			set = append(set, float64(result.Plots[i]))
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
			result.Summary[percentileString] = Value(set[0])
			continue
		} else if rank-1.0 >= float64(setSize) {
			result.Summary[percentileString] = Value(set[setSize-1])
			continue
		}

		result.Summary[percentileString] = Value(set[rankInt-1] + rankFrac*(set[rankInt]-set[rankInt-1]))
	}
}
