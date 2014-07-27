package plot

import (
	"fmt"
)

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
