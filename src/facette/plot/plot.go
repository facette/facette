package plot

import (
	"encoding/json"
	"math"
	"time"
)

const (
	// DefaultSample represents the default sample value.
	DefaultSample = 400
)

// Plot represents a time series plot instance.
type Plot struct {
	Time  time.Time `json:"time"`
	Value Value     `json:"value"`

	prev *Plot
	next *Plot
}

// MarshalJSON implements the json.Marshaler interface.
func (plot Plot) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]interface{}{int(plot.Time.Unix()), plot.Value})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (plot *Plot) UnmarshalJSON(data []byte) error {
	input := [2]float64{}
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	plot.Time = time.Unix(int64(input[0]), 0)
	plot.Value = Value(input[1])

	return nil
}

// Value represents a time series plot value.
type Value float64

// MarshalJSON implements the json.Marshaler interface.
func (value Value) MarshalJSON() ([]byte, error) {
	// Handle NaN and near-zero values marshalling
	if math.IsNaN(float64(value)) {
		return json.Marshal(nil)
	} else if math.Exp(float64(value)) == 1 {
		return json.Marshal(0)
	}

	return json.Marshal(float64(value))
}

// IsNaN reports whether the Value is an IEEE 754 'not-a-number' value.
func (value Value) IsNaN() bool {
	return math.IsNaN(float64(value))
}
