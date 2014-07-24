package plot

import (
	"fmt"
	"math"
	"testing"

	"github.com/facette/facette/pkg/utils"
)

var plotSeries = Series{
	Plots: []Plot{
		{Value: Value(math.NaN())}, {Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43},
		{Value: 68}, {Value: Value(math.NaN())}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79},
		{Value: 99}, {Value: 54}, {Value: 88}, {Value: Value(math.NaN())}, {Value: 99}, {Value: 77},
		{Value: 85}, {Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72},
		{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}, {Value: Value(math.NaN())},
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
		sampleTest{5, []Plot{
			{Value: 65.4}, {Value: 79.6}, {Value: 83.4}, {Value: 73.6}, {Value: 82.8},
		}},
		sampleTest{15, []Plot{
			{Value: 61}, {Value: 83.5}, {Value: 49.5}, {Value: 68}, {Value: 91},
			{Value: 74}, {Value: 76.5}, {Value: 88}, {Value: 88}, {Value: 85},
			{Value: 66.5}, {Value: 75}, {Value: 79.5}, {Value: 94.5}, {Value: 66},
		}},
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

	minExpectedValue = 43
	maxExpectedValue = 99
	avgExpectedValue = 76.96
	lastExpectedValue = Value(math.NaN())
	pct20thExpectedValue = 62.8
	pct50thExpectedValue = 77
	pct90thExpectedValue = 98.4

	plotSeries.Summarize([]float64{20, 50, 90})

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

func Test_SumSeries(test *testing.T) {
	var (
		// Valid series
		testFull = []Series{
			{Plots: []Plot{{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}}},
			{Plots: []Plot{{Value: 68}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79}}},
			{Plots: []Plot{{Value: 99}, {Value: 54}, {Value: 88}, {Value: 99}, {Value: 77}}},
			{Plots: []Plot{{Value: 85}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}}},
			{Plots: []Plot{{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}}},
		}

		expectedFull = Series{
			Plots: []Plot{{Value: 402}, {Value: 342}, {Value: 448}, {Value: 395}, {Value: 337}},
		}

		// Valid series featuring NaN plot values
		testNaN = []Series{
			{Plots: []Plot{
				{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}},
			},
			{Plots: []Plot{
				{Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}},
			},
			{Plots: []Plot{
				{Value: 89}, {Value: 70}, {Value: Value(math.NaN())}, {Value: 93}, {Value: 66}},
			},
		}

		expectedNaN = Series{
			Plots: []Plot{{Value: 150}, {Value: 201}, {Value: 169}, {Value: 227}, {Value: 181}},
		}

		// Valid series: not normalized
		testNotNormalized = []Series{
			Series{Plots: []Plot{{Value: 85}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}}},
			Series{Plots: []Plot{{Value: 70}, {Value: 96}, {Value: 93}}},
			Series{Plots: []Plot{{Value: 55}, {Value: 48}, {Value: 39}, {Value: 53}}},
		}

		expectedNotNormalized = Series{
			Plots: []Plot{{Value: 210}, {Value: 206}, {Value: 203}, {Value: 131}, {Value: 72}},
		}
	)

	sumFull, err := SumSeries(testFull)
	if err != nil {
		test.Logf("SumSeries(testFull) returned an error: %s", err)
		test.Fail()
	}

	if err = compareSeries(expectedFull, sumFull); err != nil {
		test.Logf(fmt.Sprintf("SumSeries(testFull): %s", err))
		test.Fail()
		return
	}

	sumNaN, err := SumSeries(testNaN)
	if err != nil {
		test.Logf("SumSeries(testNaN) returned an error: %s", err)
		test.Fail()
	}

	if err = compareSeries(expectedNaN, sumNaN); err != nil {
		test.Logf(fmt.Sprintf("SumSeries(testNaN): %s", err))
		test.Fail()
		return
	}

	sumNotNormalized, err := SumSeries(testNotNormalized)
	if err != nil {
		test.Logf("SumSeries(testNotNormalized) returned an error: %s", err)
		test.Fail()
	}

	if err = compareSeries(expectedNotNormalized, sumNotNormalized); err != nil {
		test.Logf(fmt.Sprintf("SumSeries(testNotNormalized): %s", err))
		test.Fail()
		return
	}
}

func compareSeries(expected, actual Series) error {
	for i := range expected.Plots {
		if expected.Plots[i] != actual.Plots[i] {
			return fmt.Errorf("\nExpected %v\nbut got %v", expected.Plots, actual.Plots)
		}
	}

	return nil
}
