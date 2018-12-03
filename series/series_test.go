package series

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Scale(t *testing.T) {
	series := Series{
		Points: []Point{{Value: 0.61}, {Value: 0.69}, {Value: 0.98}, {Value: Value(math.NaN())}, {Value: 0.43}},
	}

	expected := Series{
		Points: []Point{{Value: 61}, {Value: 69}, {Value: 98}, {Value: Value(math.NaN())}, {Value: 43}},
	}

	series.Scale(Value(100))
	if !compareSeries(expected, series) {
		assert.Fail(t, fmt.Sprintf("Not equal: \nexpected: %#v\nactual  : %#v", expected, series))
	}
}

func Test_Summarize(t *testing.T) {
	series := Series{
		Points: []Point{
			{Value: Value(math.NaN())}, {Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43},
			{Value: 68}, {Value: Value(math.NaN())}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79},
			{Value: 99}, {Value: 54}, {Value: 88}, {Value: Value(math.NaN())}, {Value: 99}, {Value: 77},
			{Value: 85}, {Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72},
			{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}, {Value: Value(math.NaN())},
		},
		Summary: make(map[string]Value),
	}

	seriesNeg := Series{
		Points: []Point{
			{Value: Value(math.NaN())}, {Value: -61}, {Value: -69}, {Value: -98}, {Value: -56}, {Value: -43},
			{Value: -68}, {Value: Value(math.NaN())}, {Value: -87}, {Value: -95}, {Value: -69}, {Value: -79},
			{Value: -99}, {Value: -54}, {Value: -88}, {Value: Value(math.NaN())}, {Value: -99}, {Value: -77},
			{Value: -85}, {Value: Value(math.NaN())}, {Value: -62}, {Value: -71}, {Value: -78}, {Value: -72},
			{Value: -89}, {Value: -70}, {Value: -96}, {Value: -93}, {Value: -66}, {Value: Value(math.NaN())},
		},
		Summary: make(map[string]Value),
	}

	startTime := time.Now().UTC()
	for i := range series.Points {
		series.Points[i].Time = startTime.Add(time.Duration(i) * time.Second)
		seriesNeg.Points[i].Time = startTime.Add(time.Duration(i) * time.Second)
	}

	checks := []struct {
		series *Series
		label  string
		value  Value
	}{
		{&series, "min", 43},
		{&series, "max", 99},
		{&series, "avg", 76.96},
		{&series, "last", 66},
		{&series, "20th", 62.8},
		{&series, "50th", 77},
		{&series, "90th", 98.4},
		{&seriesNeg, "min", -99},
		{&seriesNeg, "max", -43},
		{&seriesNeg, "avg", -76.96},
		{&seriesNeg, "last", -66},
		{&seriesNeg, "20th", -94.6},
		{&seriesNeg, "50th", -77},
		{&seriesNeg, "90th", -55.199999999999996},
	}

	series.Summarize([]float64{20, 50, 90})
	seriesNeg.Summarize([]float64{20, 50, 90})

	for _, c := range checks {
		assert.Equal(t, c.value, c.series.Summary[c.label])
	}
}

func compareSeries(expected, actual Series) bool {
	if len(actual.Points) != len(expected.Points) {
		return false
	}

	for i := range expected.Points {
		if actual.Points[i].Value.IsNaN() && !expected.Points[i].Value.IsNaN() ||
			!actual.Points[i].Value.IsNaN() && actual.Points[i].Value != expected.Points[i].Value ||
			!actual.Points[i].Time.Equal(expected.Points[i].Time) {
			return false
		}
	}

	return true
}
