package plot

import (
	"fmt"
	"math"
	"time"
)

const (
	_ = iota
	// ConsolidateAverage represents an average consolidation type.
	ConsolidateAverage
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
	_ = iota
	// OperTypeNone represents a null operation group mode.
	OperTypeNone
	// OperTypeAverage represents a AVG operation group mode.
	OperTypeAverage
	// OperTypeSum represents a SUM operation group mode.
	OperTypeSum
)

type plotBucket struct {
	startTime time.Time
	plots     []Plot
}

// Consolidate consolidates plots buckets based on consolidation function.
func (bucket plotBucket) Consolidate(consolidationType int) Plot {
	consolidatedPlot := Plot{
		Value: Value(math.NaN()),
		Time:  bucket.startTime,
	}

	bucketCount := len(bucket.plots)
	if bucketCount == 0 {
		return consolidatedPlot
	}

	switch consolidationType {
	case ConsolidateAverage:
		sum := 0.0
		sumCount := 0
		for _, plot := range bucket.plots {
			if plot.Value.IsNaN() {
				continue
			}

			sum += float64(plot.Value)
			sumCount++
		}

		if sumCount > 0 {
			consolidatedPlot.Value = Value(sum / float64(sumCount))
		}

		if bucketCount == 1 {
			consolidatedPlot.Time = bucket.plots[0].Time
		} else {
			// Interpolate median time
			consolidatedPlot.Time = bucket.plots[0].Time.Add(bucket.plots[bucketCount-1].Time.
				Sub(bucket.plots[0].Time) / 2)
		}

	case ConsolidateSum:
		sum := 0.0
		sumCount := 0
		for _, plot := range bucket.plots {
			if plot.Value.IsNaN() {
				continue
			}

			sum += float64(plot.Value)
			sumCount++
		}

		if sumCount > 0 {
			consolidatedPlot.Value = Value(sum)
		}

		consolidatedPlot.Time = bucket.plots[bucketCount-1].Time

	case ConsolidateLast:
		consolidatedPlot = bucket.plots[bucketCount-1]

	case ConsolidateMax:
		for _, plot := range bucket.plots {
			if !plot.Value.IsNaN() && plot.Value > consolidatedPlot.Value || consolidatedPlot.Value.IsNaN() {
				consolidatedPlot = plot
			}
		}

	case ConsolidateMin:
		for _, plot := range bucket.plots {
			if !plot.Value.IsNaN() && plot.Value < consolidatedPlot.Value || consolidatedPlot.Value.IsNaN() {
				consolidatedPlot = plot
			}
		}
	}

	return consolidatedPlot
}

// Normalize aligns series steps to the less precise one.
func Normalize(seriesList []Series, startTime, endTime time.Time, sample int, consolidationType int) ([]Series, error) {
	if sample == 0 {
		return nil, fmt.Errorf("sample must be greater than zero")
	}

	seriesCount := len(seriesList)
	if seriesCount == 0 {
		return nil, fmt.Errorf("no series provided")
	}

	consolidatedSeries := make([]Series, seriesCount)

	buckets := make([][]plotBucket, seriesCount)

	// Override sample to max series length if smaller than requested
	maxLength := 0
	for _, series := range seriesList {
		seriesLength := len(series.Plots)
		if seriesLength > maxLength {
			maxLength = seriesLength
		}
	}

	if maxLength > 0 && maxLength < sample {
		sample = maxLength
	}

	// Calculate the common step for all series based on time specs and requested sampling
	step := endTime.Sub(startTime) / time.Duration(sample)

	// Store each series' plots into the proper time step plot buckets,
	// then consolidate each series plots buckets according to consolidation function
	for seriesIndex, series := range seriesList {
		buckets[seriesIndex] = make([]plotBucket, sample)

		// Initialize time step plot buckets
		for stepIndex := 0; stepIndex < sample; stepIndex++ {
			buckets[seriesIndex][stepIndex] = plotBucket{
				startTime: startTime.Add(time.Duration(stepIndex) * step),
				plots:     make([]Plot, 0),
			}
		}

		// Dispatch series plots in the right time plot bucket
		for _, plot := range series.Plots {
			// Discard series plots out of time specs range
			if plot.Time.Before(startTime) || plot.Time.After(endTime) {
				continue
			}

			plotIndex := int64(float64(plot.Time.UnixNano()-startTime.UnixNano())/float64(step.Nanoseconds())+1) - 1
			if plotIndex >= int64(sample) {
				continue
			}

			buckets[seriesIndex][plotIndex].plots = append(buckets[seriesIndex][plotIndex].plots, plot)
		}

		consolidatedSeries[seriesIndex] = Series{
			Name:    seriesList[seriesIndex].Name,
			Plots:   make([]Plot, sample),
			Summary: make(map[string]Value),
		}

		seriesLength := len(series.Plots)

		plotRatio := sample / seriesLength
		plotCount := 0
		plotLast := Value(math.NaN())
		plotStep := endTime.Sub(startTime) / time.Duration(sample)

		// Consolidate each series' plot buckets
		for bucketIndex := range buckets[seriesIndex] {
			consolidatedSeries[seriesIndex].Plots[bucketIndex] = buckets[seriesIndex][bucketIndex].
				Consolidate(consolidationType)

			if seriesCount == 1 {
				continue
			}

			plot := &consolidatedSeries[seriesIndex].Plots[bucketIndex]

			// Align times on consolidated series lists
			plot.Time = buckets[seriesIndex][bucketIndex].startTime.Add(plotStep)

			if plotRatio <= 1 {
				continue
			}

			// Interpolate missing plots values
			if !plot.Value.IsNaN() {
				if plotCount <= plotRatio && !plotLast.IsNaN() {
					plotChunk := (plot.Value - plotLast) / Value(plotCount+1)

					for plotIndex := bucketIndex - plotCount; plotIndex < bucketIndex; plotIndex++ {
						consolidatedSeries[seriesIndex].Plots[plotIndex].Value = plotLast +
							Value(plotCount-(bucketIndex-plotIndex)+1)*plotChunk
					}
				}

				plotLast = plot.Value
				plotCount = 0
			} else {
				plotCount++
			}
		}
	}

	return consolidatedSeries, nil
}

// AverageSeries returns a new series averaging each series' datapoints.
func AverageSeries(seriesList []Series) (Series, error) {
	return operSeries(seriesList, OperTypeAverage)
}

// SumSeries add series plots together and return the sum at each datapoint.
func SumSeries(seriesList []Series) (Series, error) {
	return operSeries(seriesList, OperTypeSum)
}

func operSeries(seriesList []Series, operType int) (Series, error) {
	nSeries := len(seriesList)

	if nSeries == 0 {
		return Series{}, fmt.Errorf("no series provided")
	}

	plotsCount := len(seriesList[0].Plots)

	operSeries := Series{
		Plots:   make([]Plot, plotsCount),
		Summary: make(map[string]Value),
	}

	for plotIndex := 0; plotIndex < plotsCount; plotIndex++ {
		operSeries.Plots[plotIndex].Time = seriesList[0].Plots[plotIndex].Time

		sumCount := 0

		for _, series := range seriesList {
			if series.Plots[plotIndex].Value.IsNaN() {
				continue
			}

			operSeries.Plots[plotIndex].Value += series.Plots[plotIndex].Value
			sumCount++
		}

		if sumCount == 0 {
			operSeries.Plots[plotIndex].Value = Value(math.NaN())
		} else if operType == OperTypeAverage {
			operSeries.Plots[plotIndex].Value /= Value(sumCount)
		}
	}

	return operSeries, nil
}

func gcd(a, b int) int {
	if a <= 0 || b <= 0 {
		return 0
	}

	c := a % b
	if c == 0 {
		return b
	}

	return gcd(b, c)
}

func lcm(a, b int) int {
	return a * b / gcd(a, b)
}
