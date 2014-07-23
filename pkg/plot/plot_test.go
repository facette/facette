package plot

import (
	"math"
	"testing"

	"github.com/facette/facette/pkg/utils"
)

var plotSeries = Series{
	Plots: []Plot{
		{Value: Value(math.NaN())}, {Value: 61.0}, {Value: 69.0}, {Value: 98.0}, {Value: 56.0}, {Value: 43.0},
		{Value: 68.0}, {Value: Value(math.NaN())}, {Value: 87.0}, {Value: 95.0}, {Value: 69.0}, {Value: 79.0},
		{Value: 99.0}, {Value: 54.0}, {Value: 88.0}, {Value: Value(math.NaN())}, {Value: 99.0}, {Value: 77.0},
		{Value: 85.0}, {Value: Value(math.NaN())}, {Value: 62.0}, {Value: 71.0}, {Value: 78.0}, {Value: 72.0},
		{Value: 89.0}, {Value: 70.0}, {Value: 96.0}, {Value: 93.0}, {Value: 66.0}, {Value: Value(math.NaN())},
	},
	Summary: make(map[string]Value),
}

func Test_Series_Downsample(test *testing.T) {
	type sampleTest struct {
		Sample int
		Series []Plot
	}

	equalFunc := func(a, b []Plot) bool {
		if len(a) != len(b) {
			return false
		}

		for i := range a {
			if a[i].Value.IsNaN() && !b[i].Value.IsNaN() || !a[i].Value.IsNaN() && a[i].Value != b[i].Value {
				return false
			}
		}

		return true
	}

	for _, entry := range []sampleTest{
		sampleTest{5, []Plot{{Value: 65.4}, {Value: 79.6}, {Value: 83.4}, {Value: 73.6}, {Value: 82.8}}},
		sampleTest{15, []Plot{
			{Value: 61}, {Value: 83.5}, {Value: 49.5}, {Value: 68}, {Value: 91},
			{Value: 74}, {Value: 76.5}, {Value: 88}, {Value: 88}, {Value: 85},
			{Value: 66.5}, {Value: 75}, {Value: 79.5}, {Value: 94.5}, {Value: 66}},
		},
		sampleTest{30, plotSeries.Plots},
		sampleTest{60, plotSeries.Plots},
	} {
		series := Series{}
		utils.Clone(&plotSeries, &series)

		series.Downsample(entry.Sample)

		if !equalFunc(entry.Series, series.Plots) {
			test.Logf("\nExpected %#v\nbut got  %#v", entry.Series, series.Plots)
			test.Fail()
		}
	}
}

func Test_Series_Summarize(test *testing.T) {
	var (
		minExpectedValue, maxExpectedValue, avgExpectedValue, lastExpectedValue Value
		pct20thExpectedValue, pct50thExpectedValue, pct90thExpectedValue        Value
	)

	minExpectedValue = 43.0
	maxExpectedValue = 99.0
	avgExpectedValue = 76.96
	lastExpectedValue = Value(math.NaN())
	pct20thExpectedValue = 62.8
	pct50thExpectedValue = 77.0
	pct90thExpectedValue = 98.4

	plotSeries.Summarize([]float64{20.0, 50.0, 90.0})

	if plotSeries.Summary["min"] != minExpectedValue {
		test.Logf("\nExpected min=%g\nbut got %g", minExpectedValue, plotSeries.Summary["min"])
		test.Fail()
	}

	if plotSeries.Summary["max"] != maxExpectedValue {
		test.Logf("\nExpected max=%g\nbut got %g", maxExpectedValue, plotSeries.Summary["max"])
		test.Fail()
	}

	if plotSeries.Summary["avg"] != avgExpectedValue {
		test.Logf("\nExpected avg=%g\nbut got %g", avgExpectedValue, plotSeries.Summary["avg"])
		test.Fail()
	}

	if !plotSeries.Summary["last"].IsNaN() {
		test.Logf("\nExpected last=%g\nbut got %g", lastExpectedValue, plotSeries.Summary["last"])
		test.Fail()
	}

	if plotSeries.Summary["20th"] != pct20thExpectedValue {
		test.Logf("\nExpected 20th=%g\nbut got %g", pct20thExpectedValue, plotSeries.Summary["20th"])
		test.Fail()
	}

	if plotSeries.Summary["50th"] != pct50thExpectedValue {
		test.Logf("\nExpected 50th=%g\nbut got %g", pct50thExpectedValue, plotSeries.Summary["50th"])
		test.Fail()
	}

	if plotSeries.Summary["90th"] != pct90thExpectedValue {
		test.Logf("\nExpected 90th=%g\nbut got %g", pct90thExpectedValue, plotSeries.Summary["90th"])
		test.Fail()
	}
}
