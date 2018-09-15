package series

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testBucket                      bucket
	testSeries, testSeriesNormalize []Series
)

func init() {
	testBucket = bucket{
		startTime: time.Unix(0, 0),
		points: []Point{
			{Time: time.Unix(0, 0), Value: 17},
			{Time: time.Unix(30, 0), Value: 25},
			{Time: time.Unix(60, 0), Value: 3},
			{Time: time.Unix(90, 0), Value: Value(math.NaN())},
			{Time: time.Unix(120, 0), Value: 2},
		},
	}

	testSeries = []Series{
		{Points: []Point{{Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43}}},
		{Points: []Point{{Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72}}},
		{Points: []Point{{Value: 89}, {Value: 70}, {Value: Value(math.NaN())}, {Value: 93}, {Value: 66}}},
	}

	testSeriesNormalize = []Series{
		{Points: []Point{
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
		{Points: []Point{
			{Time: time.Unix(1, 0), Value: 17}, {Time: time.Unix(31, 0), Value: 25},
			{Time: time.Unix(61, 0), Value: 3}, {Time: time.Unix(91, 0), Value: 2},
			{Time: time.Unix(121, 0), Value: Value(math.NaN())}, {Time: time.Unix(151, 0), Value: 5},
			{Time: time.Unix(181, 0), Value: 49}, {Time: time.Unix(211, 0), Value: 0},
			{Time: time.Unix(241, 0), Value: 19}, {Time: time.Unix(271, 0), Value: 22},
		}},
		{Points: []Point{
			{Time: time.Unix(2, 0), Value: 24}, {Time: time.Unix(62, 0), Value: 16},
			{Time: time.Unix(122, 0), Value: 37}, {Time: time.Unix(182, 0), Value: 40},
			{Time: time.Unix(242, 0), Value: 43},
		}},
	}
}

func Test_Consolidate_Average(t *testing.T) {
	assert.Equal(t, Point{Time: time.Unix(60, 0), Value: 11.75}, testBucket.Consolidate(ConsolidateAverage))
}

func Test_Consolidate_Sum(t *testing.T) {
	assert.Equal(t, Point{Time: time.Unix(120, 0), Value: 47}, testBucket.Consolidate(ConsolidateSum))
}

func Test_Consolidate_First(t *testing.T) {
	assert.Equal(t, Point{Time: time.Unix(0, 0), Value: 17}, testBucket.Consolidate(ConsolidateFirst))
}

func Test_Consolidate_Last(t *testing.T) {
	assert.Equal(t, Point{Time: time.Unix(120, 0), Value: 2}, testBucket.Consolidate(ConsolidateLast))
}

func Test_Consolidate_Min(t *testing.T) {
	assert.Equal(t, Point{Time: time.Unix(120, 0), Value: 2}, testBucket.Consolidate(ConsolidateMin))
}

func Test_Consolidate_Max(t *testing.T) {
	assert.Equal(t, Point{Time: time.Unix(30, 0), Value: 25}, testBucket.Consolidate(ConsolidateMax))
}

func Test_Normalize_Average(t *testing.T) {
	testNormalize([]Series{
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 4},
				{Time: time.Unix(30, 0), Value: 27.666666666666668},
				{Time: time.Unix(60, 0), Value: 36.666666666666664},
				{Time: time.Unix(90, 0), Value: 25.333333333333332},
				{Time: time.Unix(120, 0), Value: 30.666666666666668},
				{Time: time.Unix(150, 0), Value: 24.333333333333332},
				{Time: time.Unix(180, 0), Value: 27},
				{Time: time.Unix(210, 0), Value: 33},
				{Time: time.Unix(240, 0), Value: 31.333333333333332},
				{Time: time.Unix(270, 0), Value: 26.333333333333332},
			},
		},
		{
			Points: []Point{
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
			},
		},
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 24},
				{Time: time.Unix(30, 0), Value: Value(math.NaN())},
				{Time: time.Unix(60, 0), Value: 16},
				{Time: time.Unix(90, 0), Value: Value(math.NaN())},
				{Time: time.Unix(120, 0), Value: 37},
				{Time: time.Unix(150, 0), Value: Value(math.NaN())},
				{Time: time.Unix(180, 0), Value: 40},
				{Time: time.Unix(210, 0), Value: Value(math.NaN())},
				{Time: time.Unix(240, 0), Value: 43},
				{Time: time.Unix(270, 0), Value: Value(math.NaN())},
			},
		},
	}, ConsolidateAverage, t)
}

func Test_Normalize_First(t *testing.T) {
	testNormalize([]Series{
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: Value(math.NaN())},
				{Time: time.Unix(30, 0), Value: 29},
				{Time: time.Unix(60, 0), Value: 46},
				{Time: time.Unix(90, 0), Value: 31},
				{Time: time.Unix(120, 0), Value: 20},
				{Time: time.Unix(150, 0), Value: 27},
				{Time: time.Unix(180, 0), Value: 28},
				{Time: time.Unix(210, 0), Value: 41},
				{Time: time.Unix(240, 0), Value: 31},
				{Time: time.Unix(270, 0), Value: 34},
			},
		},
		{
			Points: []Point{
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
			},
		},
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 24},
				{Time: time.Unix(30, 0), Value: Value(math.NaN())},
				{Time: time.Unix(60, 0), Value: 16},
				{Time: time.Unix(90, 0), Value: Value(math.NaN())},
				{Time: time.Unix(120, 0), Value: 37},
				{Time: time.Unix(150, 0), Value: Value(math.NaN())},
				{Time: time.Unix(180, 0), Value: 40},
				{Time: time.Unix(210, 0), Value: Value(math.NaN())},
				{Time: time.Unix(240, 0), Value: 43},
				{Time: time.Unix(270, 0), Value: Value(math.NaN())},
			},
		},
	}, ConsolidateFirst, t)
}

func Test_Normalize_Last(t *testing.T) {
	testNormalize([]Series{
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 7},
				{Time: time.Unix(30, 0), Value: 27},
				{Time: time.Unix(60, 0), Value: 43},
				{Time: time.Unix(90, 0), Value: 8},
				{Time: time.Unix(120, 0), Value: 44},
				{Time: time.Unix(150, 0), Value: 13},
				{Time: time.Unix(180, 0), Value: 41},
				{Time: time.Unix(210, 0), Value: 47},
				{Time: time.Unix(240, 0), Value: 17},
				{Time: time.Unix(270, 0), Value: 21},
			},
		},
		{
			Points: []Point{
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
			},
		},
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 24},
				{Time: time.Unix(30, 0), Value: Value(math.NaN())},
				{Time: time.Unix(60, 0), Value: 16},
				{Time: time.Unix(90, 0), Value: Value(math.NaN())},
				{Time: time.Unix(120, 0), Value: 37},
				{Time: time.Unix(150, 0), Value: Value(math.NaN())},
				{Time: time.Unix(180, 0), Value: 40},
				{Time: time.Unix(210, 0), Value: Value(math.NaN())},
				{Time: time.Unix(240, 0), Value: 43},
				{Time: time.Unix(270, 0), Value: Value(math.NaN())},
			},
		},
	}, ConsolidateLast, t)
}

func Test_Normalize_Max(t *testing.T) {
	testNormalize([]Series{
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 7},
				{Time: time.Unix(30, 0), Value: 29},
				{Time: time.Unix(60, 0), Value: 46},
				{Time: time.Unix(90, 0), Value: 37},
				{Time: time.Unix(120, 0), Value: 44},
				{Time: time.Unix(150, 0), Value: 33},
				{Time: time.Unix(180, 0), Value: 41},
				{Time: time.Unix(210, 0), Value: 47},
				{Time: time.Unix(240, 0), Value: 46},
				{Time: time.Unix(270, 0), Value: 34},
			},
		},
		{
			Points: []Point{
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
			},
		},
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 24},
				{Time: time.Unix(30, 0), Value: Value(math.NaN())},
				{Time: time.Unix(60, 0), Value: 16},
				{Time: time.Unix(90, 0), Value: Value(math.NaN())},
				{Time: time.Unix(120, 0), Value: 37},
				{Time: time.Unix(150, 0), Value: Value(math.NaN())},
				{Time: time.Unix(180, 0), Value: 40},
				{Time: time.Unix(210, 0), Value: Value(math.NaN())},
				{Time: time.Unix(240, 0), Value: 43},
				{Time: time.Unix(270, 0), Value: Value(math.NaN())},
			},
		},
	}, ConsolidateMax, t)
}

func Test_Normalize_Min(t *testing.T) {
	testNormalize([]Series{
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 1},
				{Time: time.Unix(30, 0), Value: 27},
				{Time: time.Unix(60, 0), Value: 21},
				{Time: time.Unix(90, 0), Value: 8},
				{Time: time.Unix(120, 0), Value: 20},
				{Time: time.Unix(150, 0), Value: 13},
				{Time: time.Unix(180, 0), Value: 12},
				{Time: time.Unix(210, 0), Value: 11},
				{Time: time.Unix(240, 0), Value: 17},
				{Time: time.Unix(270, 0), Value: 21},
			},
		},
		{
			Points: []Point{
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
			},
		},
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 24},
				{Time: time.Unix(30, 0), Value: Value(math.NaN())},
				{Time: time.Unix(60, 0), Value: 16},
				{Time: time.Unix(90, 0), Value: Value(math.NaN())},
				{Time: time.Unix(120, 0), Value: 37},
				{Time: time.Unix(150, 0), Value: Value(math.NaN())},
				{Time: time.Unix(180, 0), Value: 40},
				{Time: time.Unix(210, 0), Value: Value(math.NaN())},
				{Time: time.Unix(240, 0), Value: 43},
				{Time: time.Unix(270, 0), Value: Value(math.NaN())},
			},
		},
	}, ConsolidateMin, t)
}

func Test_Normalize_Sum(t *testing.T) {
	testNormalize([]Series{
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 8},
				{Time: time.Unix(30, 0), Value: 83},
				{Time: time.Unix(60, 0), Value: 110},
				{Time: time.Unix(90, 0), Value: 76},
				{Time: time.Unix(120, 0), Value: 92},
				{Time: time.Unix(150, 0), Value: 73},
				{Time: time.Unix(180, 0), Value: 81},
				{Time: time.Unix(210, 0), Value: 99},
				{Time: time.Unix(240, 0), Value: 94},
				{Time: time.Unix(270, 0), Value: 79},
			},
		},
		{
			Points: []Point{
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
			},
		},
		{
			Points: []Point{
				{Time: time.Unix(0, 0), Value: 24},
				{Time: time.Unix(30, 0), Value: Value(math.NaN())},
				{Time: time.Unix(60, 0), Value: 16},
				{Time: time.Unix(90, 0), Value: Value(math.NaN())},
				{Time: time.Unix(120, 0), Value: 37},
				{Time: time.Unix(150, 0), Value: Value(math.NaN())},
				{Time: time.Unix(180, 0), Value: 40},
				{Time: time.Unix(210, 0), Value: Value(math.NaN())},
				{Time: time.Unix(240, 0), Value: 43},
				{Time: time.Unix(270, 0), Value: Value(math.NaN())},
			},
		},
	}, ConsolidateSum, t)
}

func Test_Average(t *testing.T) {
	expected := Series{
		Points: []Point{
			{Value: 75}, {Value: 67}, {Value: 84.5}, {Value: 75.66666666666667}, {Value: 60.333333333333336},
		},
	}

	series, err := Average(testSeries)
	assert.Nil(t, err)
	if !compareSeries(series, expected) {
		assert.Fail(t, fmt.Sprintf("Not equal: \nexpected: %#v\nactual  : %#v", expected, series))
	}
}

func Test_Sum(t *testing.T) {
	expected := Series{
		Points: []Point{{Value: 150}, {Value: 201}, {Value: 169}, {Value: 227}, {Value: 181}},
	}

	series, err := Sum(testSeries)
	assert.Nil(t, err)
	if !compareSeries(series, expected) {
		assert.Fail(t, fmt.Sprintf("Not equal: \nexpected: %#v\nactual  : %#v", expected, series))
	}
}

func testNormalize(expected []Series, consolidation int, t *testing.T) {
	startTime := time.Unix(0, 0)

	series, err := Normalize(testSeriesNormalize, startTime, startTime.Add(300*time.Second), 10, consolidation)
	assert.Nil(t, err)
	assert.Len(t, series, len(expected))

	for i, s := range series {
		if !compareSeries(s, expected[i]) {
			assert.Fail(t, fmt.Sprintf("Not equal: \nexpected: %#v\nactual  : %#v", expected[i], s))
		}
	}
}
