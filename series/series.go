package series

import (
	"fmt"
	"math"
	"sort"
)

// Series represents a time series instance.
type Series struct {
	Points  []Point          `json:"points"`
	Summary map[string]Value `json:"summary"`
}

// Scale applies a factor on a series of points.
func (s *Series) Scale(factor Value) {
	for i := range s.Points {
		if !s.Points[i].Value.IsNaN() {
			s.Points[i].Value *= factor
		}
	}
}

// Summarize calculates the min/max/average/last and percentile values for a time series.
func (s *Series) Summarize(percentiles []float64) {
	var (
		min, max, total, current Value
		nValidPoints             int64
	)

	min = Value(math.NaN())
	max = Value(math.NaN())
	current = Value(math.NaN())

	for i := range s.Points {
		if !s.Points[i].Value.IsNaN() {
			current = s.Points[i].Value
			if current < min || min.IsNaN() {
				min = s.Points[i].Value
			}
			if current > max || max.IsNaN() {
				max = current
			}

			total += current
			nValidPoints++
		}
	}

	if s.Summary == nil {
		s.Summary = make(map[string]Value)
	}

	s.Summary["min"] = min
	s.Summary["max"] = max
	s.Summary["avg"] = total / Value(nValidPoints)
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

	for i := range s.Points {
		if !s.Points[i].Value.IsNaN() {
			set = append(set, float64(s.Points[i].Value))
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

// ZeroNulls replaces all null point values by zero.
func (s *Series) ZeroNulls() {
	for i := range s.Points {
		if s.Points[i].Value.IsNaN() {
			s.Points[i].Value = Value(0)
		}
	}
}
