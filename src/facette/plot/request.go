package plot

import (
	"time"

	"facette/backend"

	"github.com/facette/maputil"
)

// Request represents a plot request instance.
type Request struct {
	StartTime  time.Time      `json:"start_time,omitempty"`
	EndTime    time.Time      `json:"end_time,omitempty"`
	Time       time.Time      `json:"time,omitempty"`
	Range      string         `json:"range,omitempty"`
	Sample     int            `json:"sample"`
	ID         string         `json:"id"`
	Graph      *backend.Graph `json:"graph"`
	Attributes maputil.Map    `json:"attributes,omitempty"`
}
