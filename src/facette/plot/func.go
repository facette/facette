package plot

import (
	"math"
	"time"
)

const (
	_ = iota
	// ConsolidateAverage represents an average consolidation type.
	ConsolidateAverage
	// ConsolidateFirst represents a first value consolidation type.
	ConsolidateFirst
	// ConsolidateLast represents a last value consolidation type.
	ConsolidateLast
	// ConsolidateMax represents a maximal value consolidation type.
	ConsolidateMax
	// ConsolidateMin represents a minimal value consolidation type.
	ConsolidateMin
	// ConsolidateSum represents a sum consolidation type.
	ConsolidateSum
)

const (
	// OperatorNone represents a null operation type.
	OperatorNone = iota
	// OperatorAverage represents an average operation type.
	OperatorAverage
	// OperatorSum represents a sum operation type.
	OperatorSum
)

type bucket struct {
	startTime time.Time
	plots     []Plot
}

// Consolidate consolidates plots buckets based on consolidation function.
func (b bucket) Consolidate(consolidation int) Plot {
	plot := Plot{
		Value: Value(math.NaN()),
		Time:  b.startTime,
	}

	length := len(b.plots)
	if length == 0 {
		return plot
	}

	switch consolidation {
	case ConsolidateAverage:
		sum := 0.0
		sumCount := 0
		for _, p := range b.plots {
			if p.Value.IsNaN() {
				continue
			}

			sum += float64(p.Value)
			sumCount++
		}

		if sumCount > 0 {
			plot.Value = Value(sum / float64(sumCount))
		}

		if length == 1 {
			plot.Time = b.plots[0].Time
		} else {
			// Interpolate median time
			plot.Time = b.plots[0].Time.Add(b.plots[length-1].Time.Sub(b.plots[0].Time) / 2)
		}

	case ConsolidateSum:
		sum := 0.0
		sumCount := 0
		for _, p := range b.plots {
			if p.Value.IsNaN() {
				continue
			}

			sum += float64(p.Value)
			sumCount++
		}

		if sumCount > 0 {
			plot.Value = Value(sum)
		}

		plot.Time = b.plots[length-1].Time

	case ConsolidateFirst:
		plot = b.plots[0]

	case ConsolidateLast:
		plot = b.plots[length-1]

	case ConsolidateMax:
		for _, p := range b.plots {
			if !p.Value.IsNaN() && p.Value > plot.Value || plot.Value.IsNaN() {
				plot = p
			}
		}

	case ConsolidateMin:
		for _, p := range b.plots {
			if !p.Value.IsNaN() && p.Value < plot.Value || plot.Value.IsNaN() {
				plot = p
			}
		}
	}

	return plot
}

// Normalize aligns multiple plot series on a common time step, consolidates plots samples if necessary.
func Normalize(series []Series, startTime, endTime time.Time, sample int, consolidation int,
	interpolate bool) ([]Series, error) {

	if sample <= 0 {
		return nil, ErrInvalidSample
	}

	length := len(series)
	if length == 0 {
		return nil, ErrEmptySeries
	}

	result := make([]Series, length)
	buckets := make([][]bucket, length)

	// Calculate the common step for all series based on time range and requested sampling
	step := endTime.Sub(startTime) / time.Duration(sample)

	// Dispatch plots into proper time step buckets and then apply consolidation function
	for i, s := range series {
		if s.Plots == nil {
			continue
		}

		buckets[i] = make([]bucket, sample)

		// Initialize time steps
		for j := 0; j < sample; j++ {
			buckets[i][j] = bucket{
				startTime: startTime.Add(time.Duration(j) * step),
				plots:     make([]Plot, 0),
			}
		}

		for _, p := range s.Plots {
			// Discard series plots out of time specs range
			if p.Time.Before(startTime) || p.Time.After(endTime) {
				continue
			}

			// Stop if index goes beyond the requested sample
			idx := int64(float64(p.Time.UnixNano()-startTime.UnixNano())/float64(step.Nanoseconds())+1) - 1
			if idx >= int64(sample) {
				continue
			}

			buckets[i][idx].plots = append(buckets[i][idx].plots, p)
		}

		result[i] = Series{
			Plots:   make([]Plot, sample),
			Summary: make(map[string]Value),
		}

		// Consolidate plot buckets
		lastKnown := -1

		for j := range buckets[i] {
			result[i].Plots[j] = buckets[i][j].Consolidate(consolidation)

			if interpolate {
				// Keep reference of last and next known plots
				if lastKnown != -1 {
					result[i].Plots[j].prev = &result[i].Plots[lastKnown]
				}

				if !result[i].Plots[j].Value.IsNaN() {
					if lastKnown != -1 {
						for k := lastKnown; k < j; k++ {
							result[i].Plots[k].next = &result[i].Plots[j]
						}
					}

					lastKnown = j
				}
			}

			// Align consolidated plots timestamps among normalized series lists
			result[i].Plots[j].Time = buckets[i][j].startTime.Add(time.Duration(step.Seconds() * float64(j))).
				Round(time.Second)
		}

		// Interpolate missing points
		if !interpolate {
			continue
		}

		for j, plot := range result[i].Plots {
			if !plot.Value.IsNaN() || plot.prev == nil || plot.next == nil {
				continue
			}

			a := float64(plot.next.Value-plot.prev.Value) / float64(plot.next.Time.UnixNano()-plot.prev.Time.UnixNano())
			b := float64(plot.prev.Value) - a*float64(plot.Time.UnixNano())

			result[i].Plots[j].Value = Value(a*float64(plot.next.Time.UnixNano()) + b)
		}
	}

	return result, nil
}

// Average returns a new series averaging each datapoints.
func Average(series []Series) (Series, error) {
	return applyOperator(series, OperatorAverage)
}

// Sum returns a new series summing each datapoints.
func Sum(series []Series) (Series, error) {
	return applyOperator(series, OperatorSum)
}

func applyOperator(series []Series, operator int) (Series, error) {
	length := len(series)
	if length == 0 {
		return Series{}, ErrEmptySeries
	}

	count := len(series[0].Plots)

	result := Series{
		Plots:   make([]Plot, count),
		Summary: make(map[string]Value),
	}

	for i := 0; i < count; i++ {
		sumCount := 0

		result.Plots[i].Time = series[0].Plots[i].Time

		for _, s := range series {
			if len(s.Plots) != count {
				return Series{}, ErrUnnormalizedSeries
			} else if s.Plots[i].Value.IsNaN() {
				continue
			}

			result.Plots[i].Value += s.Plots[i].Value
			sumCount++
		}

		if sumCount == 0 {
			result.Plots[i].Value = Value(math.NaN())
		} else if operator == OperatorAverage {
			result.Plots[i].Value /= Value(sumCount)
		}
	}

	return result, nil
}
