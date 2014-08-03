package plot

import (
	"math"
	"testing"
	"time"

	"github.com/facette/facette/pkg/utils"
)

// import (
// 	"fmt"
// 	"math"
// 	"reflect"
// 	"testing"
// )

// func Test_FuncConsolidateSeries(test *testing.T) {
// 	endTime := time.Now()
// 	startTime := endTime.Add(-1 * time.Hour)

// 	testSeries := []Series{
// 		{Plots: []Plot{
// 			{Value: 12}, {Value: 4}, {Value: 22}, {Value: Value(math.NaN())}, {Value: 8},
// 			{Value: 6}, {Value: 8}, {Value: Value(math.NaN())}, {Value: 1}, {Value: 56},
// 			{Value: 2}, {Value: 32}, {Value: 22}, {Value: 30}, {Value: 3},
// 			{Value: 2}, {Value: 3}, {Value: 15}, {Value: 26}, {Value: 31},
// 			{Value: 22}, {Value: Value(math.NaN())}, {Value: 28}, {Value: 1},
// 		}, Step: 120},
// 		{Plots: []Plot{
// 			{Value: 2}, {Value: 6}, {Value: 4}, {Value: Value(math.NaN())},
// 			{Value: 11}, {Value: 9}, {Value: 8}, {Value: 8},
// 			{Value: 22}, {Value: Value(math.NaN())}, {Value: 16}, {Value: 4},
// 		}, Step: 400},
// 		{Plots: []Plot{
// 			{Value: 7}, {Value: 12}, {Value: 5},
// 		}, Step: 1600},
// 	}

// 	expectedSeries := []Series{
// 		{Plots: []Plot{
// 			{Value: 10}, {Value: 18.5}, {Value: 18},
// 		}, Step: 1600},
// 		{Plots: []Plot{
// 			{Value: 4}, {Value: 9}, {Value: 14},
// 		}, Step: 1600},
// 		{Plots: []Plot{
// 			{Value: 7}, {Value: 12}, {Value: 5},
// 		}, Step: 1600},
// 	}

// 	consolidatedSeries, _ := ConsolidateSeries(testSeries, startTime, endTime, 1600, ConsolidateAverage)

// 	if !reflect.DeepEqual(expectedSeries, consolidatedSeries) {
// 		test.Logf("\nExpected %#v\nbut got  %#v", expectedSeries, consolidatedSeries)
// 		test.Fail()
// 	}
// }

// func Test_FuncAverageSeries(test *testing.T) {
// 	var (
// 		// Valid series
// 		testFull = []Series{
// 			{Step: 10, Plots: []Plot{{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}}},
// 			{Step: 10, Plots: []Plot{{Value: 68}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79}}},
// 			{Step: 10, Plots: []Plot{{Value: 99}, {Value: 54}, {Value: 88}, {Value: 99}, {Value: 77}}},
// 			{Step: 10, Plots: []Plot{{Value: 85}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}}},
// 			{Step: 10, Plots: []Plot{{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}}},
// 		}

// 		expectedFull = Series{
// 			Step: 10, Plots: []Plot{{Value: 80.4}, {Value: 68.4}, {Value: 89.6}, {Value: 79}, {Value: 67.4}},
// 		}

// 		// Valid series featuring NaN plot values
// 		testNaN = []Series{
// 			{Step: 10, Plots: []Plot{
// 				{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}},
// 			},
// 			{Step: 10, Plots: []Plot{
// 				{Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}},
// 			},
// 			{Step: 10, Plots: []Plot{
// 				{Value: 89}, {Value: 70}, {Value: Value(math.NaN())}, {Value: 93}, {Value: 66}},
// 			},
// 		}

// 		expectedNaN = Series{
// 			Step: 10,
// 			Plots: []Plot{
// 				{Value: 75}, {Value: 67}, {Value: 84.5}, {Value: 75.66666666666667}, {Value: 60.333333333333336},
// 			},
// 		}

// 		// Valid series: not consolidated
// 		testNotNormalized = []Series{
// 			{Plots: []Plot{
// 				{Value: 12}, {Value: 4}, {Value: 22}, {Value: Value(math.NaN())}, {Value: 8},
// 				{Value: 6}, {Value: 8}, {Value: Value(math.NaN())}, {Value: 1}, {Value: 56},
// 				{Value: 2}, {Value: 32}, {Value: 22}, {Value: 30}, {Value: 3},
// 				{Value: 2}, {Value: 3}, {Value: 15}, {Value: 26}, {Value: 31},
// 				{Value: 22}, {Value: Value(math.NaN())}, {Value: 28}, {Value: 1},
// 			}, Step: 200},
// 			{Plots: []Plot{
// 				{Value: 2}, {Value: 6}, {Value: 4}, {Value: Value(math.NaN())},
// 				{Value: 11}, {Value: 9}, {Value: 8}, {Value: 8},
// 				{Value: 22}, {Value: Value(math.NaN())}, {Value: 16}, {Value: 4},
// 			}, Step: 400},
// 			{Plots: []Plot{
// 				{Value: 7}, {Value: 12}, {Value: 5},
// 			}, Step: 1600},
// 		}

// 		expectedNotNormalized = Series{
// 			Step:  1600,
// 			Plots: []Plot{{Value: 7}, {Value: 13.166666666666666}, {Value: 12.333333333333334}},
// 		}
// 	)

// 	avgFull, err := AverageSeries(testFull)
// 	if err != nil {
// 		test.Logf("AverageSeries(testFull) returned an error: %s", err)
// 		test.Fail()
// 	}

// 	if err = compareSeries(expectedFull, avgFull); err != nil {
// 		test.Logf(fmt.Sprintf("AverageSeries(testFull): %s", err))
// 		test.Fail()
// 		return
// 	}

// 	avgNaN, err := AverageSeries(testNaN)
// 	if err != nil {
// 		test.Logf("AverageSeries(testNaN) returned an error: %s", err)
// 		test.Fail()
// 	}

// 	if err = compareSeries(expectedNaN, avgNaN); err != nil {
// 		test.Logf(fmt.Sprintf("AverageSeries(testNaN): %s", err))
// 		test.Fail()
// 		return
// 	}

// 	avgNotNormalized, err := AverageSeries(testNotNormalized)
// 	if err != nil {
// 		test.Logf("AverageSeries(testNotNormalized) returned an error: %s", err)
// 		test.Fail()
// 	}

// 	if err = compareSeries(expectedNotNormalized, avgNotNormalized); err != nil {
// 		test.Logf(fmt.Sprintf("AverageSeries(testNotNormalized): %s", err))
// 		test.Fail()
// 		return
// 	}
// }

// func Test_FuncSumSeries(test *testing.T) {
// 	var (
// 		// Valid series
// 		testFull = []Series{
// 			{Step: 10, Plots: []Plot{{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}}},
// 			{Step: 10, Plots: []Plot{{Value: 68}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79}}},
// 			{Step: 10, Plots: []Plot{{Value: 99}, {Value: 54}, {Value: 88}, {Value: 99}, {Value: 77}}},
// 			{Step: 10, Plots: []Plot{{Value: 85}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}}},
// 			{Step: 10, Plots: []Plot{{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}}},
// 		}

// 		expectedFull = Series{
// 			Step: 10, Plots: []Plot{{Value: 402}, {Value: 342}, {Value: 448}, {Value: 395}, {Value: 337}},
// 		}

// 		// Valid series featuring NaN plot values
// 		testNaN = []Series{
// 			{Step: 10, Plots: []Plot{
// 				{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}},
// 			},
// 			{Step: 10, Plots: []Plot{
// 				{Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}},
// 			},
// 			{Step: 10, Plots: []Plot{
// 				{Value: 89}, {Value: 70}, {Value: Value(math.NaN())}, {Value: 93}, {Value: 66}},
// 			},
// 		}

// 		expectedNaN = Series{
// 			Step:  10,
// 			Plots: []Plot{{Value: 150}, {Value: 201}, {Value: 169}, {Value: 227}, {Value: 181}},
// 		}

// 		// Valid series: not consolidated
// 		testNotNormalized = []Series{
// 			{Plots: []Plot{
// 				{Value: 12}, {Value: 4}, {Value: 22}, {Value: Value(math.NaN())}, {Value: 8},
// 				{Value: 6}, {Value: 8}, {Value: Value(math.NaN())}, {Value: 1}, {Value: 56},
// 				{Value: 2}, {Value: 32}, {Value: 22}, {Value: 30}, {Value: 3},
// 				{Value: 2}, {Value: 3}, {Value: 15}, {Value: 26}, {Value: 31},
// 				{Value: 22}, {Value: Value(math.NaN())}, {Value: 28}, {Value: 1},
// 			}, Step: 200},
// 			{Plots: []Plot{
// 				{Value: 2}, {Value: 6}, {Value: 4}, {Value: Value(math.NaN())},
// 				{Value: 11}, {Value: 9}, {Value: 8}, {Value: 8},
// 				{Value: 22}, {Value: Value(math.NaN())}, {Value: 16}, {Value: 4},
// 			}, Step: 400},
// 			{Plots: []Plot{
// 				{Value: 7}, {Value: 12}, {Value: 5},
// 			}, Step: 1600},
// 		}

// 		expectedNotNormalized = Series{
// 			Step:  1600,
// 			Plots: []Plot{{Value: 21}, {Value: 39.5}, {Value: 37}},
// 		}
// 	)

// 	sumFull, err := SumSeries(testFull)
// 	if err != nil {
// 		test.Logf("SumSeries(testFull) returned an error: %s", err)
// 		test.Fail()
// 	}

// 	if err = compareSeries(expectedFull, sumFull); err != nil {
// 		test.Logf(fmt.Sprintf("SumSeries(testFull): %s", err))
// 		test.Fail()
// 		return
// 	}

// 	sumNaN, err := SumSeries(testNaN)
// 	if err != nil {
// 		test.Logf("SumSeries(testNaN) returned an error: %s", err)
// 		test.Fail()
// 	}

// 	if err = compareSeries(expectedNaN, sumNaN); err != nil {
// 		test.Logf(fmt.Sprintf("SumSeries(testNaN): %s", err))
// 		test.Fail()
// 		return
// 	}

// 	sumNotNormalized, err := SumSeries(testNotNormalized)
// 	if err != nil {
// 		test.Logf("SumSeries(testNotNormalized) returned an error: %s", err)
// 		test.Fail()
// 	}

// 	if err = compareSeries(expectedNotNormalized, sumNotNormalized); err != nil {
// 		test.Logf(fmt.Sprintf("SumSeries(testNotNormalized): %s", err))
// 		test.Fail()
// 		return
// 	}
// }

func Test_ConsolidateAverage(test *testing.T) {
	testSlice := []sampleTest{
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
	}

	for testIndex := range testSlice {
		if testSlice[testIndex].Sample > len(plotSeries.Plots) {
			continue
		}

		chunkSize := len(plotSeries.Plots) / testSlice[testIndex].Sample

		sampleIndex := 0

		for plotIndex := chunkSize; plotIndex < len(plotSeries.Plots); plotIndex += chunkSize {
			testSlice[testIndex].Plots[sampleIndex].Time = plotSeries.Plots[plotIndex-chunkSize].Time.Add(
				plotSeries.Plots[plotIndex].Time.Sub(plotSeries.Plots[plotIndex-chunkSize].Time) / 2,
			)

			sampleIndex++
		}
	}

	consolidateHandle(test, testSlice, ConsolidateAverage)
}

func Test_ConsolidateMax(test *testing.T) {
	testSlice := []sampleTest{
		sampleTest{5, []Plot{
			{Value: 98}, {Value: 95}, {Value: 99}, {Value: 85}, {Value: 96},
		}},
		sampleTest{15, []Plot{
			{Value: 61}, {Value: 98}, {Value: 56}, {Value: 68}, {Value: 95},
			{Value: 79}, {Value: 99}, {Value: 88}, {Value: 99}, {Value: 85},
			{Value: 71}, {Value: 78}, {Value: 89}, {Value: 96}, {Value: 66},
		}},
		sampleTest{30, plotSeries.Plots},
		sampleTest{60, plotSeries.Plots},
	}

	for testIndex := range testSlice {
		if testSlice[testIndex].Sample > len(plotSeries.Plots) {
			continue
		}

		chunkSize := len(plotSeries.Plots) / testSlice[testIndex].Sample

		maxTime := time.Time{}

		sampleIndex := 0

		for plotIndex := range plotSeries.Plots {
			if plotIndex%chunkSize == 0 {
				testSlice[testIndex].Plots[sampleIndex].Time = maxTime
				maxTime = time.Time{}
				continue
			}

			if plotSeries.Plots[plotIndex].Time.After(maxTime) {
				maxTime = plotSeries.Plots[plotIndex].Time
			}
		}
	}

	consolidateHandle(test, testSlice, ConsolidateMax)
}

func Test_ConsolidateMin(test *testing.T) {
	testSlice := []sampleTest{
		sampleTest{5, []Plot{
			{Value: 43}, {Value: 68}, {Value: 54}, {Value: 62}, {Value: 66},
		}},
		sampleTest{15, []Plot{
			{Value: 61}, {Value: 69}, {Value: 43}, {Value: 68}, {Value: 87},
			{Value: 69}, {Value: 54}, {Value: 88}, {Value: 77}, {Value: 85},
			{Value: 62}, {Value: 72}, {Value: 70}, {Value: 93}, {Value: 66},
		}},
		sampleTest{30, plotSeries.Plots},
		sampleTest{60, plotSeries.Plots},
	}

	for testIndex := range testSlice {
		if testSlice[testIndex].Sample > len(plotSeries.Plots) {
			continue
		}

		chunkSize := len(plotSeries.Plots) / testSlice[testIndex].Sample

		minTime := time.Time{}

		sampleIndex := 0

		for plotIndex := range plotSeries.Plots {
			if plotIndex%chunkSize == 0 {
				testSlice[testIndex].Plots[sampleIndex].Time = minTime
				minTime = time.Time{}
				continue
			}

			if minTime.IsZero() || plotSeries.Plots[plotIndex].Time.Before(minTime) {
				minTime = plotSeries.Plots[plotIndex].Time
			}
		}
	}

	consolidateHandle(test, testSlice, ConsolidateMin)
}

func Test_ConsolidateLast(test *testing.T) {
	testSlice := []sampleTest{
		sampleTest{5, []Plot{
			{Value: 43}, {Value: 79}, {Value: 77}, {Value: 72}, {Value: Value(math.NaN())},
		}},
		sampleTest{15, []Plot{
			{Value: 61}, {Value: 98}, {Value: 43}, {Value: Value(math.NaN())}, {Value: 95},
			{Value: 79}, {Value: 54}, {Value: Value(math.NaN())}, {Value: 77}, {Value: Value(math.NaN())},
			{Value: 71}, {Value: 72}, {Value: 70}, {Value: 93}, {Value: Value(math.NaN())},
		}},
		sampleTest{30, plotSeries.Plots},
		sampleTest{60, plotSeries.Plots},
	}

	for testIndex := range testSlice {
		if testSlice[testIndex].Sample > len(plotSeries.Plots) {
			continue
		}

		chunkTime := time.Duration(len(plotSeries.Plots)/testSlice[testIndex].Sample) * time.Second

		for plotIndex := range testSlice[testIndex].Plots {
			testSlice[testIndex].Plots[plotIndex].Time = startTime.Add(time.Duration(plotIndex) * chunkTime)
		}
	}

	consolidateHandle(test, testSlice, ConsolidateLast)
}

func Test_ConsolidateSum(test *testing.T) {
	testSlice := []sampleTest{
		sampleTest{5, []Plot{
			{Value: 327}, {Value: 398}, {Value: 417}, {Value: 368}, {Value: 414},
		}},
		sampleTest{15, []Plot{
			{Value: 61}, {Value: 167}, {Value: 99}, {Value: 68}, {Value: 182},
			{Value: 148}, {Value: 153}, {Value: 88}, {Value: 176}, {Value: 85},
			{Value: 133}, {Value: 150}, {Value: 159}, {Value: 189}, {Value: 66},
		}},
		sampleTest{30, plotSeries.Plots},
		sampleTest{60, plotSeries.Plots},
	}

	for testIndex := range testSlice {
		if testSlice[testIndex].Sample > len(plotSeries.Plots) {
			continue
		}

		chunkTime := time.Duration(len(plotSeries.Plots)/testSlice[testIndex].Sample) * time.Second

		for plotIndex := range testSlice[testIndex].Plots {
			testSlice[testIndex].Plots[plotIndex].Time = startTime.Add(time.Duration(plotIndex) * chunkTime)
		}
	}

	consolidateHandle(test, testSlice, ConsolidateSum)
}

func consolidateHandle(test *testing.T, testSlice []sampleTest, consolidationType int) {
	for _, entry := range testSlice {
		series := Series{}
		utils.Clone(&plotSeries, &series)

		series.Downsample(startTime, endTime, entry.Sample, consolidationType)

		if !consolidateEqual(entry.Plots, series.Plots) {
			test.Logf("\nExpected %#v\nbut got  %#v", entry.Plots, series.Plots)
			test.Fail()
		}
	}
}

func consolidateEqual(a, b []Plot) bool {
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
