package plot

import (
	"math"
	"testing"

	"github.com/facette/facette/pkg/utils"
)

var plotResult = Result{
	Plots: []Value{
		Value(math.NaN()), 61.0, 69.0, 98.0, 56.0, 43.0,
		68.0, Value(math.NaN()), 87.0, 95.0, 69.0, 79.0,
		99.0, 54.0, 88.0, Value(math.NaN()), 99.0, 77.0,
		85.0, Value(math.NaN()), 62.0, 71.0, 78.0, 72.0,
		89.0, 70.0, 96.0, 93.0, 66.0, Value(math.NaN()),
	},
	Summary: make(map[string]Value),
}

func Test_Result_Downsample(test *testing.T) {
	type sampleTest struct {
		Sample int
		Result []Value
	}

	equalFunc := func(a, b []Value) bool {
		if len(a) != len(b) {
			return false
		}

		for i := range a {
			if a[i].IsNaN() && !b[i].IsNaN() || !a[i].IsNaN() && a[i] != b[i] {
				return false
			}
		}

		return true
	}

	for _, entry := range []sampleTest{
		sampleTest{5, []Value{65.4, 79.6, 83.4, 73.6, 82.8}},
		sampleTest{15, []Value{61, 83.5, 49.5, 68, 91, 74, 76.5, 88, 88, 85, 66.5, 75, 79.5, 94.5, 66}},
		sampleTest{30, plotResult.Plots},
		sampleTest{60, plotResult.Plots},
	} {
		result := Result{}
		utils.Clone(&plotResult, &result)

		result.Downsample(entry.Sample)

		if !equalFunc(entry.Result, result.Plots) {
			test.Logf("\nExpected %#v\nbut got  %#v", entry.Result, result.Plots)
			test.Fail()
		}
	}
}

func Test_Result_Summarize(test *testing.T) {
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

	plotResult.Summarize([]float64{20.0, 50.0, 90.0})

	if plotResult.Summary["min"] != minExpectedValue {
		test.Logf("\nExpected min=%g\nbut got %g", minExpectedValue, plotResult.Summary["min"])
		test.Fail()
	}

	if plotResult.Summary["max"] != maxExpectedValue {
		test.Logf("\nExpected max=%g\nbut got %g", maxExpectedValue, plotResult.Summary["max"])
		test.Fail()
	}

	if plotResult.Summary["avg"] != avgExpectedValue {
		test.Logf("\nExpected avg=%g\nbut got %g", avgExpectedValue, plotResult.Summary["avg"])
		test.Fail()
	}

	if !plotResult.Summary["last"].IsNaN() {
		test.Logf("\nExpected last=%g\nbut got %g", lastExpectedValue, plotResult.Summary["last"])
		test.Fail()
	}

	if plotResult.Summary["20th"] != pct20thExpectedValue {
		test.Logf("\nExpected 20th=%g\nbut got %g", pct20thExpectedValue, plotResult.Summary["20th"])
		test.Fail()
	}

	if plotResult.Summary["50th"] != pct50thExpectedValue {
		test.Logf("\nExpected 50th=%g\nbut got %g", pct50thExpectedValue, plotResult.Summary["50th"])
		test.Fail()
	}

	if plotResult.Summary["90th"] != pct90thExpectedValue {
		test.Logf("\nExpected 90th=%g\nbut got %g", pct90thExpectedValue, plotResult.Summary["90th"])
		test.Fail()
	}
}
