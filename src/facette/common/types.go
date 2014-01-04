package common

import (
	"encoding/json"
	"math"
)

// PlotValue represents a graph plot value.
type PlotValue float64

// MarshalJSON handles JSON marshaling of the PlotValue type.
func (value PlotValue) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) || math.Floor(float64(value)) == 0 {
		return json.Marshal(nil)
	}

	return json.Marshal(float64(value))
}
