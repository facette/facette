package plot

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/facette/facette/pkg/utils"
)

func Test_plotBucketConsolidate(test *testing.T) {
	testBucket := plotBucket{
		startTime: time.Unix(0, 0),
		plots: []Plot{
			{Time: time.Unix(0, 0), Value: 17}, {Time: time.Unix(30, 0), Value: 25},
			{Time: time.Unix(60, 0), Value: 3}, {Time: time.Unix(90, 0), Value: Value(math.NaN())},
			{Time: time.Unix(120, 0), Value: 2},
		},
	}

	// "Average" bucket consolidation
	expectedBucketAverage := Plot{Time: time.Unix(60, 0), Value: 11.75}
	actualBucketAverage := testBucket.Consolidate(ConsolidateAverage)
	if !reflect.DeepEqual(expectedBucketAverage, actualBucketAverage) {
		test.Logf("\nExpected %s\nbut got  %s", &expectedBucketAverage, &actualBucketAverage)
		test.Fail()
		return
	}

	// "Sum" bucket consolidation
	expectedBucketSum := Plot{Time: time.Unix(120, 0), Value: 47}
	actualBucketSum := testBucket.Consolidate(ConsolidateSum)
	if !reflect.DeepEqual(expectedBucketSum, actualBucketSum) {
		test.Logf("\nExpected %s\nbut got  %s", &expectedBucketSum, &actualBucketSum)
		test.Fail()
		return
	}

	// "Last" bucket consolidation
	expectedBucketLast := Plot{Time: time.Unix(120, 0), Value: 2}
	actualBucketLast := testBucket.Consolidate(ConsolidateLast)
	if !reflect.DeepEqual(expectedBucketLast, actualBucketLast) {
		test.Logf("\nExpected %s\nbut got  %s", &expectedBucketLast, &actualBucketLast)
		test.Fail()
		return
	}

	// "Min" bucket consolidation
	expectedBucketMin := Plot{Time: time.Unix(120, 0), Value: 2}
	actualBucketMin := testBucket.Consolidate(ConsolidateMin)
	if !reflect.DeepEqual(expectedBucketMin, actualBucketMin) {
		test.Logf("\nExpected %s\nbut got  %s", &expectedBucketMin, &actualBucketMin)
		test.Fail()
		return
	}

	// "Max" bucket consolidation
	expectedBucketMax := Plot{Time: time.Unix(30, 0), Value: 25}
	actualBucketMax := testBucket.Consolidate(ConsolidateMax)
	if !reflect.DeepEqual(expectedBucketMax, actualBucketMax) {
		test.Logf("\nExpected %s\nbut got  %s", &expectedBucketMax, &actualBucketMax)
		test.Fail()
		return
	}
}

func Test_FuncNormalize(test *testing.T) {
	testSeries := []Series{
		{Name: "series0", Step: 10, Plots: []Plot{
			{Time: time.Unix(0, 0), Value: Value(math.NaN())}, {Time: time.Unix(10, 0), Value: 1},
			{Time: time.Unix(20, 0), Value: 7}, {Time: time.Unix(30, 0), Value: 29},
			{Time: time.Unix(40, 0), Value: 27}, {Time: time.Unix(50, 0), Value: 27},
			{Time: time.Unix(60, 0), Value: 46}, {Time: time.Unix(70, 0), Value: 21},
			{Time: time.Unix(80, 0), Value: 43}, {Time: time.Unix(90, 0), Value: 31},
			{Time: time.Unix(100, 0), Value: 37}, {Time: time.Unix(110, 0), Value: 8},
			{Time: time.Unix(120, 0), Value: 20}, {Time: time.Unix(130, 0), Value: 28},
			{Time: time.Unix(140, 0), Value: 44}, {Time: time.Unix(150, 0), Value: 27},
			{Time: time.Unix(160, 0), Value: 33}, {Time: time.Unix(170, 0), Value: 13},
			{Time: time.Unix(180, 0), Value: 28}, {Time: time.Unix(190, 0), Value: 12},
			{Time: time.Unix(200, 0), Value: 41}, {Time: time.Unix(210, 0), Value: 41},
			{Time: time.Unix(220, 0), Value: 11}, {Time: time.Unix(230, 0), Value: 47},
			{Time: time.Unix(240, 0), Value: 31}, {Time: time.Unix(250, 0), Value: 46},
			{Time: time.Unix(260, 0), Value: 17}, {Time: time.Unix(270, 0), Value: 34},
			{Time: time.Unix(280, 0), Value: 24}, {Time: time.Unix(290, 0), Value: 21},
		}},
		{Name: "series1", Step: 30, Plots: []Plot{
			{Time: time.Unix(1, 0), Value: 17}, {Time: time.Unix(31, 0), Value: 25},
			{Time: time.Unix(61, 0), Value: 3}, {Time: time.Unix(91, 0), Value: 2},
			{Time: time.Unix(121, 0), Value: Value(math.NaN())}, {Time: time.Unix(151, 0), Value: 5},
			{Time: time.Unix(181, 0), Value: 49}, {Time: time.Unix(211, 0), Value: 0},
			{Time: time.Unix(241, 0), Value: 19}, {Time: time.Unix(271, 0), Value: 22},
		}},
		{Name: "series2", Step: 60, Plots: []Plot{
			{Time: time.Unix(2, 0), Value: 24}, {Time: time.Unix(62, 0), Value: 16},
			{Time: time.Unix(122, 0), Value: 37}, {Time: time.Unix(182, 0), Value: 40},
			{Time: time.Unix(242, 0), Value: 43},
		}},
	}

	startTime := time.Unix(0, 0)
	endTime := startTime.Add(300 * time.Second)

	expected := []Series{
		{Name: "series0", Step: 30, Plots: []Plot{
			{Time: time.Unix(0, 0), Value: 4},                    // (NaN + 1 + 7) / 2
			{Time: time.Unix(30, 0), Value: 27.666666666666668},  // (29 + 27 + 27) / 3
			{Time: time.Unix(60, 0), Value: 36.666666666666664},  // (46 + 21 + 43) / 3
			{Time: time.Unix(90, 0), Value: 25.333333333333332},  // (31 + 37 + 8) / 3
			{Time: time.Unix(120, 0), Value: 30.666666666666668}, // (20 + 28 + 44) / 3
			{Time: time.Unix(150, 0), Value: 24.333333333333332}, // (27 + 33 + 13) / 3
			{Time: time.Unix(180, 0), Value: 27},                 // (28 + 12 + 41) / 3
			{Time: time.Unix(210, 0), Value: 33},                 // (41 + 11 + 47) / 3
			{Time: time.Unix(240, 0), Value: 31.333333333333332}, // (31 + 46 + 17) / 3
			{Time: time.Unix(270, 0), Value: 26.333333333333332}, // (34 + 24 + 21) / 3
		}},
		{Name: "series1", Step: 30, Plots: []Plot{
			{Time: time.Unix(0, 0), Value: 17},
			{Time: time.Unix(30, 0), Value: 25},
			{Time: time.Unix(60, 0), Value: 3},
			{Time: time.Unix(90, 0), Value: 2},
			{Time: time.Unix(120, 0), Value: Value(math.NaN())},
			{Time: time.Unix(150, 0), Value: 5},
			{Time: time.Unix(180, 0), Value: 49},
			{Time: time.Unix(210, 0), Value: 0},
			{Time: time.Unix(240, 0), Value: 19},
			{Time: time.Unix(270, 0), Value: 22},
		}},
		{Name: "series2", Step: 30, Plots: []Plot{
			{Time: time.Unix(0, 0), Value: 24},
			{Time: time.Unix(30, 0), Value: 20},
			{Time: time.Unix(60, 0), Value: 16},
			{Time: time.Unix(90, 0), Value: 26.5},
			{Time: time.Unix(120, 0), Value: 37},
			{Time: time.Unix(150, 0), Value: 38.5},
			{Time: time.Unix(180, 0), Value: 40},
			{Time: time.Unix(210, 0), Value: 41.5},
			{Time: time.Unix(240, 0), Value: 43},
			{Time: time.Unix(270, 0), Value: Value(math.NaN())},
		}},
	}

	actual, err := Normalize(testSeries, startTime, endTime, 10, ConsolidateAverage)
	if err != nil {
		test.Logf("\nNormalize() returned an error: %s", err)
		test.Fail()
		return
	}

	if len(expected) != len(actual) {
		test.Logf("\nNormalized series count does not match original series list")
		test.Fail()
		return
	}

	for seriesIndex := range expected {
		if len(expected[seriesIndex].Plots) != len(actual[seriesIndex].Plots) {
			test.Logf(
				"\nPlots number mismatch for series #%d: expected %d, got %d",
				len(expected[seriesIndex].Plots),
				len(actual[seriesIndex].Plots),
			)
			test.Fail()
			return
		}

		if !seriesEqual(expected[seriesIndex].Plots, actual[seriesIndex].Plots, true) {
			test.Logf("\nExpected %+v\nbut got  %+v", expected[seriesIndex].Plots, actual[seriesIndex].Plots)
			test.Fail()
			return
		}
	}
}

func Test_FuncAverageSeries(test *testing.T) {
	testSeries := []Series{
		{Step: 10, Plots: []Plot{
			{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}},
		},
		{Step: 10, Plots: []Plot{
			{Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}},
		},
		{Step: 10, Plots: []Plot{
			{Value: 89}, {Value: 70}, {Value: Value(math.NaN())}, {Value: 93}, {Value: 66}},
		},
	}

	expected := Series{
		Step: 10,
		Plots: []Plot{
			{Value: 75}, {Value: 67}, {Value: 84.5}, {Value: 75.66666666666667}, {Value: 60.333333333333336},
		},
	}

	actual, err := AverageSeries(testSeries)
	if err != nil {
		test.Logf("AverageSeries() returned an error: %s", err)
		test.Fail()
		return
	}

	if !seriesEqual(expected.Plots, actual.Plots, false) {
		test.Logf("\nExpected %+v\nbut got  %+v", expected.Plots, actual.Plots)
		test.Fail()
		return
	}
}

func Test_FuncSumSeries(test *testing.T) {
	testSeries := []Series{
		{Step: 10, Plots: []Plot{
			{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}},
		},
		{Step: 10, Plots: []Plot{
			{Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}},
		},
		{Step: 10, Plots: []Plot{
			{Value: 89}, {Value: 70}, {Value: Value(math.NaN())}, {Value: 93}, {Value: 66}},
		},
	}

	expected := Series{
		Step:  10,
		Plots: []Plot{{Value: 150}, {Value: 201}, {Value: 169}, {Value: 227}, {Value: 181}},
	}

	actual, err := SumSeries(testSeries)
	if err != nil {
		test.Logf("SumSeries() returned an error: %s", err)
		test.Fail()
		return
	}

	if !seriesEqual(expected.Plots, actual.Plots, false) {
		test.Logf("\nExpected %+v\nbut got  %+v", expected.Plots, actual.Plots)
		test.Fail()
		return
	}
}

func Test_DownsampleAverage(test *testing.T) {
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

	testDownsampling(test, testSlice, ConsolidateAverage)
}

func Test_DownsampleMax(test *testing.T) {
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

	testDownsampling(test, testSlice, ConsolidateMax)
}

func Test_DownsampleMin(test *testing.T) {
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

	testDownsampling(test, testSlice, ConsolidateMin)
}

func Test_DownsampleLast(test *testing.T) {
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

	testDownsampling(test, testSlice, ConsolidateLast)
}

func Test_DownsampleSum(test *testing.T) {
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

	testDownsampling(test, testSlice, ConsolidateSum)
}

func testDownsampling(test *testing.T, testSlice []sampleTest, consolidationType int) {
	for _, entry := range testSlice {
		series := Series{}
		utils.Clone(&plotSeries, &series)

		series.Downsample(startTime, endTime, entry.Sample, consolidationType)

		if !seriesEqual(entry.Plots, series.Plots, false) {
			test.Logf("\nExpected %+v\nbut got  %+v", entry.Plots, series.Plots)
			test.Fail()
			return
		}
	}
}

func seriesEqual(a, b []Plot, compareTimetamp bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Value.IsNaN() && !b[i].Value.IsNaN() || !a[i].Value.IsNaN() && a[i].Value != b[i].Value {
			return false
		}

		if compareTimetamp {
			if !a[i].Time.Equal(b[i].Time) {
				return false
			}
		}
	}

	return true
}
