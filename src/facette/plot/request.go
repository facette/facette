package plot

import (
	"time"

	"facette/backend"
	"facette/mapper"
)

// Request represents a plot request instance.
type Request struct {
	Time       time.Time      `json:"time"`
	Range      string         `json:"range"`
	Sample     int            `json:"sample"`
	ID         string         `json:"id"`
	Graph      *backend.Graph `json:"graph"`
	Attributes mapper.Map     `json:"attributes,omitempty"`
}
