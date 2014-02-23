package types

import (
	"encoding/json"
	"math"
)

// PlotValue represents a graph plot value.
type PlotValue float64

// MarshalJSON handles JSON marshaling of the PlotValue type.
func (value PlotValue) MarshalJSON() ([]byte, error) {
	// Handle NaN and near-zero values
	if math.IsNaN(float64(value)) {
		return json.Marshal(nil)
	} else if math.Exp(float64(value)) == 1 {
		return json.Marshal(0)
	}

	return json.Marshal(float64(value))
}
