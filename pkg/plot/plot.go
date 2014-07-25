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

// Series represents a series of plots.
type Series struct {
	Name    string
	Plots   []Plot
	Summary map[string]Value
}

// Downsample applies a sampling function on a series of plots, reducing the number of points.
func (series *Series) Downsample(sample int) {
	var nPlots = len(series.Plots)

	if sample >= len(series.Plots) {
		return
	}

	plots := series.Plots[:]
	series.Plots = make([]Plot, 0)

	pad := nPlots / sample

	refinePad := 0
	if nPlots%sample > 0 {
		refinePad = nPlots / (nPlots % sample)
	}

	padCount := 0

	bucket := 0.0
	bucketCount := 0.0

	for i := 0; i < nPlots; i++ {
		// Refine sampling by appending one more plot at regular interval (pad + 1)
		if refinePad == 0 || (i+1)%refinePad != 0 {
			padCount++
		}

		if !plots[i].Value.IsNaN() {
			bucket += float64(plots[i].Value)
			bucketCount++
		}

		if padCount == pad {
			padCount = 0

			if bucketCount == 0 {
				series.Plots = append(series.Plots, Plot{Value: Value(math.NaN()), Time: plots[i].Time})
				continue
			}

			series.Plots = append(series.Plots, Plot{Value: Value(bucket / bucketCount), Time: plots[i].Time})

			bucket = 0
			bucketCount = 0
		}
	}

	if bucketCount > 0 {
		series.Plots = append(series.Plots, Plot{Value: Value(bucket / bucketCount), Time: plots[nPlots-1].Time})
	}
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

// SumSeries add series plots together and return the sum at each datapoint.
func SumSeries(seriesList []Series) (Series, error) {
	nSeries := len(seriesList)

	if nSeries == 0 {
		return Series{}, fmt.Errorf("no series provided")
	}

	// Find out the longest series of the list
	maxPlots := len(seriesList[0].Plots)
	for i := range seriesList {
		if len(seriesList[i].Plots) > maxPlots {
			maxPlots = len(seriesList[i].Plots)
		}
	}

	sumSeries := Series{
		Plots:   make([]Plot, maxPlots),
		Summary: make(map[string]Value),
	}

	for plotIndex := 0; plotIndex < maxPlots; plotIndex++ {
		for _, series := range seriesList {
			// Skip shorter series
			if plotIndex >= len(series.Plots) {
				continue
			}

			if !series.Plots[plotIndex].Value.IsNaN() {
				sumSeries.Plots[plotIndex].Value += series.Plots[plotIndex].Value
				sumSeries.Plots[plotIndex].Time = series.Plots[plotIndex].Time
			}
		}
	}

	return sumSeries, nil
}
