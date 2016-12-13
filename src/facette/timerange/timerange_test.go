package timerange

import (
	"testing"
	"time"
)

func Test_Apply(t *testing.T) {
	ref := time.Now().UTC()

	tests := []struct {
		Range    string
		Expected time.Time
	}{
		{"-1h", ref.Add(-1 * time.Hour)},
		{"2mo", ref.AddDate(0, 2, 0)},
		{"-1y 3h 126s", ref.AddDate(-1, 0, 0).Add(-3*time.Hour - 126*time.Second)},
		{"3d 1h 6m", ref.AddDate(0, 0, 3).Add(time.Hour + 6*time.Minute)},
	}

	for _, entry := range tests {
		result, err := Apply(ref, entry.Range)
		if err != nil {
			t.Log(err)
			t.Fail()
		} else if !result.Equal(entry.Expected) {
			t.Logf("\nExpected %#v\nbut got  %#v", entry.Expected, result)
			t.Fail()
		}
	}
}

func Test_FromDuration(t *testing.T) {
	ref := time.Now().UTC()

	tests := []struct {
		Duration time.Duration
		Expected string
	}{
		{ref.Sub(ref.Add(time.Hour)), "-1h"},
		{ref.Sub(ref.AddDate(0, 0, 60)) * -1, "60d"},
		{ref.Sub(ref.AddDate(0, 0, 1).Add(3*time.Hour + 126*time.Second)), "-1d 3h 2m 6s"},
		{ref.Sub(ref.AddDate(0, 0, 3).Add(time.Hour+6*time.Minute)) * -1, "3d 1h 6m"},
	}

	for _, entry := range tests {
		result := FromDuration(entry.Duration)
		if result != entry.Expected {
			t.Logf("\nExpected %#v\nbut got  %#v", entry.Expected, result)
			t.Fail()
		}
	}
}
