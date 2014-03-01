package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	durationRegexp = "^([-+])?\\s*" +
		"(?:(\\d+)\\s*y(?:ears?)?)?\\s*" +
		"(?:(\\d+)\\s*mo(?:nths?)?)?\\s*" +
		"(?:(\\d+)\\s*d(?:ays?)?)?\\s*" +
		"(?:(\\d+)\\s*h(?:ours?)?)?\\s*" +
		"(?:(\\d+)\\s*m(?:inutes?)?)?\\s*" +
		"(?:(\\d+)\\s*s(?:econds?)?)?" +
		"$"
)

// TimeApplyRange applies a string defined time range to a specific date.
func TimeApplyRange(refTime time.Time, input string) (time.Time, error) {
	re := regexp.MustCompile(durationRegexp)

	modifier := 1

	chunks := make([]int, 0)

	for key, value := range re.FindStringSubmatch(strings.Trim(input, " ")) {
		var intVal int

		if key == 0 {
			continue
		} else if key == 1 {
			if value == "-" {
				modifier = -1
			}

			continue
		}

		if value == "" {
			chunks = append(chunks, 0)
			continue
		}

		intVal, err := strconv.Atoi(value)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid range")
		}

		chunks = append(chunks, intVal*modifier)
	}

	if len(chunks) == 0 {
		return time.Time{}, fmt.Errorf("invalid range")
	}

	newTime := refTime.
		AddDate(chunks[0], chunks[1], chunks[2]).
		Add(time.Duration(chunks[3]) * time.Hour).
		Add(time.Duration(chunks[4]) * time.Minute).
		Add(time.Duration(chunks[5]) * time.Second)

	return newTime, nil
}
