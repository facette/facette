package plot

import (
	"fmt"
	"math"
	"sort"
)

// Series represents a time series instance.
type Series struct {
	Plots   []Plot           `json:"plots"`
	Step    int              `json:"step"`
	Summary map[string]Value `json:"summary"`
}

// Scale applies a factor on a series of plots.
func (s *Series) Scale(factor Value) {
	for i := range s.Plots {
		if !s.Plots[i].Value.IsNaN() {
			s.Plots[i].Value *= factor
		}
	}
}

// Summarize calculates the min/max/average/last and percentile values for a time series.
func (s *Series) Summarize(percentiles []float64) {
	var (
		min, max, total, current Value
		nValidPlots              int64
	)

	min = Value(math.NaN())
	max = Value(math.NaN())

	for i := range s.Plots {
		if !s.Plots[i].Value.IsNaN() {
			current = s.Plots[i].Value
			if current < min || min.IsNaN() {
				min = s.Plots[i].Value
			}
			if current > max || max.IsNaN() {
				max = current
			}

			total += current
			nValidPlots++
		}
	}

	if s.Summary == nil {
		s.Summary = make(map[string]Value)
	}

	s.Summary["min"] = min
	s.Summary["max"] = max
	s.Summary["avg"] = total / Value(nValidPlots)
	s.Summary["last"] = current

	if len(percentiles) > 0 {
		s.Percentiles(percentiles)
	}
}

// Percentiles calculates the percentile values for a time series.
func (s *Series) Percentiles(values []float64) {
	var set []float64

	// Stop if no percentile value provided
	if len(values) == 0 {
		return
	}

	for i := range s.Plots {
		if !s.Plots[i].Value.IsNaN() {
			set = append(set, float64(s.Plots[i].Value))
		}
	}

	count := len(set)
	if count == 0 {
		return
	}

	sort.Float64s(set)

	// Calculate percentiles
	for _, pct := range values {
		label := fmt.Sprintf("%gth", pct)

		rank := (pct / 100) * float64(count+1)
		rankInt := int(rank)
		rankFrac := rank - float64(rankInt)

		if rank <= 0.0 {
			s.Summary[label] = Value(set[0])
			continue
		} else if rank-1.0 >= float64(count) {
			s.Summary[label] = Value(set[count-1])
			continue
		}

		s.Summary[label] = Value(set[rankInt-1] + rankFrac*(set[rankInt]-set[rankInt-1]))
	}
}
