package plot

import (
	"fmt"
)

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
