package timerange

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const durationRegexp = "^([-+])?\\s*" +
	"(?:(\\d+)\\s*y(?:ears?)?)?\\s*" +
	"(?:(\\d+)\\s*mo(?:nths?)?)?\\s*" +
	"(?:(\\d+)\\s*d(?:ays?)?)?\\s*" +
	"(?:(\\d+)\\s*h(?:ours?)?)?\\s*" +
	"(?:(\\d+)\\s*m(?:inutes?)?)?\\s*" +
	"(?:(\\d+)\\s*s(?:econds?)?)?" +
	"$"

type durationUnit struct {
	value int
	text  string
}

// FromDuration converts a duration into a string-defined time range.
func FromDuration(d time.Duration) string {
	units := []durationUnit{
		{86400, "d"},
		{3600, "h"},
		{60, "m"},
		{1, "s"},
	}

	parts := []string{}
	seconds := int(math.Abs(d.Seconds()))

	for _, unit := range units {
		count := int(math.Floor(float64(seconds / unit.value)))
		if count > 0 {
			parts = append(parts, fmt.Sprintf("%d%s", count, unit.text))
			seconds %= unit.value
		}
	}

	result := strings.Join(parts, " ")

	if d < 0 {
		result = "-" + result
	}

	return result
}

// Apply applies a string-defined time range to a specific date.
func Apply(t time.Time, input string) (time.Time, error) {
	modifier := 1
	parts := []int{}

	for i, value := range regexp.MustCompile(durationRegexp).FindStringSubmatch(input) {
		if i == 0 {
			continue
		} else if i == 1 {
			if value == "-" {
				modifier = -1
			}

			continue
		}

		if value == "" {
			parts = append(parts, 0)
			continue
		}

		v, err := strconv.Atoi(value)
		if err != nil {
			return t, ErrInvalidRange
		}
		parts = append(parts, v*modifier)
	}

	if len(parts) == 0 {
		return t, ErrInvalidRange
	}

	return t.AddDate(parts[0], parts[1], parts[2]).
		Add(time.Duration(parts[3]) * time.Hour).
		Add(time.Duration(parts[4]) * time.Minute).
		Add(time.Duration(parts[5]) * time.Second), nil
}
