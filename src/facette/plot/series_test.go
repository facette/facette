package plot

import (
	"math"
	"testing"
	"time"
)

func Test_Scale(t *testing.T) {
	series := Series{
		Plots: []Plot{{Value: 0.61}, {Value: 0.69}, {Value: 0.98}, {Value: Value(math.NaN())}, {Value: 0.43}},
	}

	expected := Series{
		Plots: []Plot{{Value: 61}, {Value: 69}, {Value: 98}, {Value: Value(math.NaN())}, {Value: 43}},
	}

	series.Scale(Value(100))
	if !compareSeries(series, expected) {
		t.Logf("\nExpected %#v\nbut got  %#v", expected, series)
		t.Fail()
	}
}

func Test_Summarize(t *testing.T) {
	series := Series{
		Plots: []Plot{
			{Value: Value(math.NaN())}, {Value: 61}, {Value: 69}, {Value: 98}, {Value: 56}, {Value: 43},
			{Value: 68}, {Value: Value(math.NaN())}, {Value: 87}, {Value: 95}, {Value: 69}, {Value: 79},
			{Value: 99}, {Value: 54}, {Value: 88}, {Value: Value(math.NaN())}, {Value: 99}, {Value: 77},
			{Value: 85}, {Value: Value(math.NaN())}, {Value: 62}, {Value: 71}, {Value: 78}, {Value: 72},
			{Value: 89}, {Value: 70}, {Value: 96}, {Value: 93}, {Value: 66}, {Value: Value(math.NaN())},
		},
		Summary: make(map[string]Value),
	}

	seriesNeg := Series{
		Plots: []Plot{
			{Value: Value(math.NaN())}, {Value: -61}, {Value: -69}, {Value: -98}, {Value: -56}, {Value: -43},
			{Value: -68}, {Value: Value(math.NaN())}, {Value: -87}, {Value: -95}, {Value: -69}, {Value: -79},
			{Value: -99}, {Value: -54}, {Value: -88}, {Value: Value(math.NaN())}, {Value: -99}, {Value: -77},
			{Value: -85}, {Value: Value(math.NaN())}, {Value: -62}, {Value: -71}, {Value: -78}, {Value: -72},
			{Value: -89}, {Value: -70}, {Value: -96}, {Value: -93}, {Value: -66}, {Value: Value(math.NaN())},
		},
		Summary: make(map[string]Value),
	}

	startTime := time.Now().UTC()
	for i := range series.Plots {
		series.Plots[i].Time = startTime.Add(time.Duration(i) * time.Second)
		seriesNeg.Plots[i].Time = startTime.Add(time.Duration(i) * time.Second)
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
		if c.series.Summary[c.label] != c.value {
			t.Logf("\nExpected %s=%g\nbut got  %s=%g", c.label, c.value, c.label, c.series.Summary[c.label])
			t.Fail()
		}
	}
}

func compareSeries(actual, expected Series) bool {
	if len(actual.Plots) != len(expected.Plots) {
		return false
	}

	for i := range expected.Plots {
		if actual.Plots[i].Value.IsNaN() && !expected.Plots[i].Value.IsNaN() ||
			!actual.Plots[i].Value.IsNaN() && actual.Plots[i].Value != expected.Plots[i].Value ||
			!actual.Plots[i].Time.Equal(expected.Plots[i].Time) {
			return false
		}
	}

	return true
}
