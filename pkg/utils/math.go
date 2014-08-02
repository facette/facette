package utils

import (
	"math"
)

// Round rounds a float64 value into a int64.
func Round(input float64) int64 {
	integer, fraction := math.Modf(input)

	if math.Abs(fraction) >= 0.5 {
		if integer >= 1 {
			return int64(integer) + 1
		} else if integer <= -1 {
			return int64(integer) - 1
		}
	}

	return int64(integer)
}
