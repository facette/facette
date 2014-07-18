package types

import (
	"testing"
)

var plotResult = PlotResult{
	Plots: []PlotValue{
		61.0, 69.0, 98.0, 56.0, 43.0,
		68.0, 87.0, 95.0, 69.0, 79.0,
		99.0, 54.0, 88.0, 99.0, 77.0,
		85.0, 62.0, 71.0, 78.0, 72.0,
		89.0, 70.0, 96.0, 93.0, 66.0,
	},
	Info: make(map[string]PlotValue),
}

func Test_PlotResult_Summarize(test *testing.T) {
	var (
		minExpectedValue, maxExpectedValue, avgExpectedValue, lastExpectedValue PlotValue
		pct20thExpectedValue, pct50thExpectedValue, pct90thExpectedValue        PlotValue
	)

	minExpectedValue = 43.0
	maxExpectedValue = 99.0
	avgExpectedValue = 76.96
	lastExpectedValue = 66.0
	pct20thExpectedValue = 62.8
	pct50thExpectedValue = 77.0
	pct90thExpectedValue = 98.4

	plotResult.Summarize([]float64{20.0, 50.0, 90.0})

	if plotResult.Info["min"] != minExpectedValue {
		test.Logf("\nExpected min=%g\nbut got %g", minExpectedValue, plotResult.Info["min"])
		test.Fail()
	}

	if plotResult.Info["max"] != maxExpectedValue {
		test.Logf("\nExpected max=%g\nbut got %g", maxExpectedValue, plotResult.Info["max"])
		test.Fail()
	}

	if plotResult.Info["avg"] != avgExpectedValue {
		test.Logf("\nExpected avg=%g\nbut got %g", avgExpectedValue, plotResult.Info["avg"])
		test.Fail()
	}

	if plotResult.Info["last"] != lastExpectedValue {
		test.Logf("\nExpected last=%g\nbut got %g", lastExpectedValue, plotResult.Info["last"])
		test.Fail()
	}

	if plotResult.Info["20th"] != pct20thExpectedValue {
		test.Logf("\nExpected 20th=%g\nbut got %g", pct20thExpectedValue, plotResult.Info["20th"])
		test.Fail()
	}

	if plotResult.Info["50th"] != pct50thExpectedValue {
		test.Logf("\nExpected 50th=%g\nbut got %g", pct50thExpectedValue, plotResult.Info["50th"])
		test.Fail()
	}

	if plotResult.Info["90th"] != pct90thExpectedValue {
		test.Logf("\nExpected 90th=%g\nbut got %g", pct90thExpectedValue, plotResult.Info["90th"])
		test.Fail()
	}
}
