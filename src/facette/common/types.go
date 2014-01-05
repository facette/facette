package common

import (
	"encoding/json"
	"math"
)

// PlotValue represents a graph plot value.
type PlotValue float64

// MarshalJSON handles JSON marshaling of the PlotValue type.
func (value PlotValue) MarshalJSON() ([]byte, error) {
	// Make NaN and very small values null
	if math.IsNaN(float64(value)) || math.Exp(float64(value)) == 1 {
		return json.Marshal(nil)
	}

	return json.Marshal(float64(value))
}
