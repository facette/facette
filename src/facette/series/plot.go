package series

import (
	"encoding/json"
	"math"
	"time"
)

const (
	// DefaultSample represents the default sample value.
	DefaultSample = 400
)

// Point represents a time series point instance.
type Point struct {
	Time  time.Time `json:"time"`
	Value Value     `json:"value"`

	prev *Point
	next *Point
}

// MarshalJSON implements the json.Marshaler interface.
func (point Point) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]interface{}{int(point.Time.Unix()), point.Value})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (point *Point) UnmarshalJSON(data []byte) error {
	input := [2]float64{}
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	point.Time = time.Unix(int64(input[0]), 0)
	point.Value = Value(input[1])

	return nil
}

// Value represents a time series point value.
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
