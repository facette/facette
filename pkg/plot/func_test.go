package plot

import (
	"fmt"
	"math"
	"testing"
)

func Test_FuncSumSeries(test *testing.T) {
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
