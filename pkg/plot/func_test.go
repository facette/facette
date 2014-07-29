package plot

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func Test_FuncAvgSeries(test *testing.T) {
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
			Plots: []Plot{{Value: 80.4}, {Value: 68.4}, {Value: 89.6}, {Value: 79}, {Value: 67.4}},
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
			Plots: []Plot{
				{Value: 75}, {Value: 67}, {Value: 84.5}, {Value: 75.66666666666667}, {Value: 60.333333333333336},
			},
		}

		// Valid series: not normalized
		testNotNormalized = []Series{
			Series{Plots: []Plot{{Value: 85}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}}},
			Series{Plots: []Plot{{Value: 70}, {Value: 96}, {Value: 93}}},
			Series{Plots: []Plot{{Value: 55}, {Value: 48}, {Value: 39}, {Value: 53}}},
		}

		expectedNotNormalized = Series{
			Plots: []Plot{
				{Value: 70}, {Value: 68.66666666666667}, {Value: 67.66666666666667}, {Value: 65.5}, {Value: 72},
			},
		}
	)

	avgFull, err := AvgSeries(testFull)
	if err != nil {
		test.Logf("AvgSeries(testFull) returned an error: %s", err)
		test.Fail()
	}

	if err = compareSeries(expectedFull, avgFull); err != nil {
		test.Logf(fmt.Sprintf("AvgSeries(testFull): %s", err))
		test.Fail()
		return
	}

	avgNaN, err := AvgSeries(testNaN)
	if err != nil {
		test.Logf("AvgSeries(testNaN) returned an error: %s", err)
		test.Fail()
	}

	if err = compareSeries(expectedNaN, avgNaN); err != nil {
		test.Logf(fmt.Sprintf("AvgSeries(testNaN): %s", err))
		test.Fail()
		return
	}

	avgNotNormalized, err := AvgSeries(testNotNormalized)
	if err != nil {
		test.Logf("AvgSeries(testNotNormalized) returned an error: %s", err)
		test.Fail()
	}

	if err = compareSeries(expectedNotNormalized, avgNotNormalized); err != nil {
		test.Logf(fmt.Sprintf("AvgSeries(testNotNormalized): %s", err))
		test.Fail()
		return
	}
}

func Test_FuncNormalize(test *testing.T) {
	testSeries := []Series{
		{Plots: []Plot{
			{Value: 12}, {Value: 4}, {Value: 22}, {Value: Value(math.NaN())}, {Value: 8},
			{Value: 6}, {Value: 8}, {Value: Value(math.NaN())}, {Value: 1}, {Value: 56},
			{Value: 2}, {Value: 32}, {Value: 22}, {Value: 30}, {Value: 3},
			{Value: 2}, {Value: 3}, {Value: 15}, {Value: 26}, {Value: 31},
			{Value: 22}, {Value: Value(math.NaN())}, {Value: 28}, {Value: 1},
		}, Step: 200},
		{Plots: []Plot{
			{Value: 2}, {Value: 6}, {Value: 4}, {Value: Value(math.NaN())},
			{Value: 11}, {Value: 9}, {Value: 8}, {Value: 8},
			{Value: 22}, {Value: Value(math.NaN())}, {Value: 16}, {Value: 4},
		}, Step: 400},
		{Plots: []Plot{
			{Value: 7}, {Value: 12}, {Value: 5},
		}, Step: 1600},
	}

	expectedSeries := []Series{
		{Plots: []Plot{
			{Value: 10}, {Value: 18.5}, {Value: 18},
		}, Step: 1600},
		{Plots: []Plot{
			{Value: 4}, {Value: 9}, {Value: 14},
		}, Step: 1600},
		{Plots: []Plot{
			{Value: 7}, {Value: 12}, {Value: 5},
		}, Step: 1600},
	}

	normalizedSeries, _ := Normalize(testSeries, ConsolidateAverage)

	if !reflect.DeepEqual(expectedSeries, normalizedSeries) {
		test.Logf("\nExpected %#v\nbut got  %#v", expectedSeries, normalizedSeries)
		test.Fail()
	}
}

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
