package utils

// Round rounds a float64 value into a int64.
func Round(input float64) int64 {
	if input < 0.0 {
		input -= 0.5
	} else {
		input += 0.5
	}

	return int64(input)
}
