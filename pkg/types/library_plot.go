package types

import (
	"encoding/json"
	"math"
	"time"
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

// PlotRequest represents a eplot request struct in the server library backend.
type PlotRequest struct {
	Time        string    `json:"time"`
	Range       string    `json:"range"`
	Sample      int       `json:"sample"`
	Constants   []float64 `json:"constants"`
	Percentiles []float64 `json:"percentiles"`
	Graph       string    `json:"graph"`
	Origin      string    `json:"origin"`
	Source      string    `json:"source"`
	Metric      string    `json:"metric"`
	Template    string    `json:"template"`
	Filter      string    `json:"filter"`
}

// SerieResponse represents a serie response struct in the server library backend.
type SerieResponse struct {
	Name    string                 `json:"name"`
	Plots   []PlotValue            `json:"plots"`
	Info    map[string]PlotValue   `json:"info"`
	Options map[string]interface{} `json:"options"`
}

// StackResponse represents a stack response struct in the server library backend.
type StackResponse struct {
	Name   string           `json:"name"`
	Series []*SerieResponse `json:"series"`
}

// PlotResponse represents a plot response struct in the server library backend.
type PlotResponse struct {
	ID          string           `json:"id"`
	Start       string           `json:"start"`
	End         string           `json:"end"`
	Step        float64          `json:"step"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        int              `json:"type"`
	StackMode   int              `json:"stack_mode"`
	Stacks      []*StackResponse `json:"stacks"`
	Modified    time.Time        `json:"modified"`
}
