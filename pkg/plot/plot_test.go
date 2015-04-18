package plot

import (
	"fmt"
	"math"
	"testing"
	"time"
)

type sampleTest struct {
	Sample int
	Plots  []Plot
}

var (
	plotSeries, plotSeriesNeg Series
	startTime, endTime        time.Time
)

func init() {
	plotSeries = Series{
		Plots: []Plot{
			{Value: Value(math.NaN())}, {Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43},
			{Value: 68}, {Value: Value(math.NaN())}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79},
			{Value: 99}, {Value: 54}, {Value: 88}, {Value: Value(math.NaN())}, {Value: 99}, {Value: 77},
			{Value: 85}, {Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72},
			{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}, {Value: Value(math.NaN())},
		},
		Summary: make(map[string]Value),
	}

	plotSeriesNeg = Series{
		Plots: []Plot{
			{Value: Value(math.NaN())}, {Value: -61}, {Value: -69}, {Value: -98}, {Value: -56}, {Value: -43},
			{Value: -68}, {Value: Value(math.NaN())}, {Value: -87}, {Value: -95}, {Value: -69}, {Value: -79},
			{Value: -99}, {Value: -54}, {Value: -88}, {Value: Value(math.NaN())}, {Value: -99}, {Value: -77},
			{Value: -85}, {Value: Value(math.NaN())}, {Value: -62}, {Value: -71}, {Value: -78}, {Value: -72},
			{Value: -89}, {Value: -70}, {Value: -96}, {Value: -93}, {Value: -66}, {Value: Value(math.NaN())},
		},
		Summary: make(map[string]Value),
	}

	startTime = time.Now()
	endTime = startTime.Add(time.Duration(len(plotSeries.Plots)) * time.Second)

	for plotIndex := range plotSeries.Plots {
		plotSeries.Plots[plotIndex].Time = startTime.Add(time.Duration(plotIndex) * time.Second)
		plotSeriesNeg.Plots[plotIndex].Time = startTime.Add(time.Duration(plotIndex) * time.Second)
	}
}

func Test_SeriesScale(test *testing.T) {
	var (
		testSeries = Series{
			Plots: []Plot{{Value: 0.61}, {Value: 0.69}, {Value: 0.98}, {Value: Value(math.NaN())}, {Value: 0.43}},
		}

		expectedSeries = Series{
			Plots: []Plot{{Value: 61}, {Value: 69}, {Value: 98}, {Value: Value(math.NaN())}, {Value: 43}},
		}
	)

	testSeries.Scale(Value(100))
	if err := compareSeries(expectedSeries, testSeries); err != nil {
		test.Logf(fmt.Sprintf("%s", err))
		test.Fail()
		return
	}
}

func Test_SeriesSummarize(test *testing.T) {
	var (
		minExpectedValue, maxExpectedValue, avgExpectedValue, lastExpectedValue             Value
		pct20thExpectedValue, pct50thExpectedValue, pct90thExpectedValue                    Value
		minExpectedNegValue, maxExpectedNegValue, avgExpectedNegValue, lastExpectedNegValue Value
		pct20thExpectedNegValue, pct50thExpectedNegValue, pct90thExpectedNegValue           Value
	)

	minExpectedValue = 43
	minExpectedNegValue = -99
	maxExpectedValue = 99
	maxExpectedNegValue = -43
	avgExpectedValue = 76.96
	avgExpectedNegValue = -76.96
	lastExpectedValue = 66
	lastExpectedNegValue = -66
	pct20thExpectedValue = 62.8
	pct20thExpectedNegValue = -94.6
	pct50thExpectedValue = 77
	pct50thExpectedNegValue = -77
	pct90thExpectedValue = 98.4
	pct90thExpectedNegValue = -55.199999999999996

	plotSeries.Summarize([]float64{20, 50, 90})
	plotSeriesNeg.Summarize([]float64{20, 50, 90})

	if plotSeries.Summary["min"] != minExpectedValue {
		test.Logf("\nExpected min=%g\nbut got %g", minExpectedValue, plotSeries.Summary["min"])
		test.Fail()
		return
	}

	if plotSeries.Summary["max"] != maxExpectedValue {
		test.Logf("\nExpected max=%g\nbut got %g", maxExpectedValue, plotSeries.Summary["max"])
		test.Fail()
		return
	}

	if plotSeries.Summary["avg"] != avgExpectedValue {
		test.Logf("\nExpected avg=%g\nbut got %g", avgExpectedValue, plotSeries.Summary["avg"])
		test.Fail()
		return
	}

	if plotSeries.Summary["last"] != lastExpectedValue {
		test.Logf("\nExpected last=%g\nbut got %g", lastExpectedValue, plotSeries.Summary["last"])
		test.Fail()
		return
	}

	if plotSeries.Summary["20th"] != pct20thExpectedValue {
		test.Logf("\nExpected 20th=%g\nbut got %g", pct20thExpectedValue, plotSeries.Summary["20th"])
		test.Fail()
		return
	}

	if plotSeries.Summary["50th"] != pct50thExpectedValue {
		test.Logf("\nExpected 50th=%g\nbut got %g", pct50thExpectedValue, plotSeries.Summary["50th"])
		test.Fail()
		return
	}

	if plotSeries.Summary["90th"] != pct90thExpectedValue {
		test.Logf("\nExpected 90th=%g\nbut got %g", pct90thExpectedValue, plotSeries.Summary["90th"])
		test.Fail()
		return
	}

	// Summaries for negative only plotSeries

	if plotSeriesNeg.Summary["min"] != minExpectedNegValue {
		test.Logf("\nExpected min=%g\nbut got %g", minExpectedNegValue, plotSeriesNeg.Summary["min"])
		test.Fail()
		return
	}

	if plotSeriesNeg.Summary["max"] != maxExpectedNegValue {
		test.Logf("\nExpected max=%g\nbut got %g", maxExpectedNegValue, plotSeriesNeg.Summary["max"])
		test.Fail()
		return
	}

	if plotSeriesNeg.Summary["avg"] != avgExpectedNegValue {
		test.Logf("\nExpected avg=%g\nbut got %g", avgExpectedNegValue, plotSeriesNeg.Summary["avg"])
		test.Fail()
		return
	}

	if plotSeriesNeg.Summary["last"] != lastExpectedNegValue {
		test.Logf("\nExpected last=%g\nbut got %g", lastExpectedNegValue, plotSeriesNeg.Summary["last"])
		test.Fail()
		return
	}

	if plotSeriesNeg.Summary["20th"] != pct20thExpectedNegValue {
		test.Logf("\nExpected 20th=%g\nbut got %g", pct20thExpectedNegValue, plotSeriesNeg.Summary["20th"])
		test.Fail()
		return
	}

	if plotSeriesNeg.Summary["50th"] != pct50thExpectedNegValue {
		test.Logf("\nExpected 50th=%g\nbut got %g", pct50thExpectedNegValue, plotSeriesNeg.Summary["50th"])
		test.Fail()
		return
	}

	if plotSeriesNeg.Summary["90th"] != pct90thExpectedNegValue {
		test.Logf("\nExpected 90th=%g\nbut got %g", pct90thExpectedNegValue, plotSeriesNeg.Summary["90th"])
		test.Fail()
		return
	}

}

func compareSeries(expected, actual Series) error {
	for i := range expected.Plots {
		if expected.Plots[i].Value.IsNaN() {
			if expected.Plots[i].Value.IsNaN() && !actual.Plots[i].Value.IsNaN() {
				return fmt.Errorf("\nExpected %v\nbut got %v", expected.Plots, actual.Plots)
			}
		} else if expected.Plots[i] != actual.Plots[i] {
			return fmt.Errorf("\nExpected %v\nbut got %v", expected.Plots, actual.Plots)
		}
	}

	return nil
}
