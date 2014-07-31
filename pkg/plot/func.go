package plot

import (
	"fmt"

	"github.com/facette/facette/pkg/utils"
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

// NormalizeSeries aligns series steps to the less precise one.
func NormalizeSeries(series []Series, consolidationType int) ([]Series, error) {
	var step int

	seriesCount := len(series)

	if seriesCount == 0 {
		return nil, fmt.Errorf("no series provided")
	}

	outputSeries := make([]Series, seriesCount)

	// Get least common multiple
	step = series[0].Step
	if seriesCount > 1 {
		for i := 1; i < seriesCount; i++ {
			step = lcm(step, series[i].Step)
		}
	}

	for i, serie := range series {
		outputSeries[i] = Series{}
		utils.Clone(&serie, &outputSeries[i])

		outputSeries[i].Consolidate(step/outputSeries[i].Step, consolidationType)
		outputSeries[i].Step = step
	}

	return outputSeries, nil
}

// AvgSeries returns a new series averaging each series' datapoints.
func AvgSeries(seriesList []Series) (Series, error) {
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

	avgSeries := Series{
		Plots:   make([]Plot, maxPlots),
		Summary: make(map[string]Value),
	}

	for plotIndex := 0; plotIndex < maxPlots; plotIndex++ {
		var validPlots Value

		for _, series := range seriesList {
			// Skip shorter series
			if plotIndex >= len(series.Plots) {
				continue
			}

			if !series.Plots[plotIndex].Value.IsNaN() {
				avgSeries.Plots[plotIndex].Value += series.Plots[plotIndex].Value
				avgSeries.Plots[plotIndex].Time = series.Plots[plotIndex].Time
				validPlots++
			}

		}

		if validPlots > 0 {
			avgSeries.Plots[plotIndex].Value /= validPlots
		}
	}

	return avgSeries, nil
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

func gcd(a, b int) int {
	c := a % b
	if c == 0 {
		return b
	}

	return gcd(b, c)
}

func lcm(a, b int) int {
	return a * b / gcd(a, b)
}
