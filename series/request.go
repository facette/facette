package series

import (
	"time"

	"facette.io/facette/storage"
	"facette.io/maputil"
)

// Request represents a point request instance.
type Request struct {
	StartTime  time.Time      `json:"start_time,omitempty"`
	EndTime    time.Time      `json:"end_time,omitempty"`
	Time       time.Time      `json:"time,omitempty"`
	Range      string         `json:"range,omitempty"`
	Sample     int            `json:"sample"`
	ID         string         `json:"id"`
	Graph      *storage.Graph `json:"graph"`
	Attributes maputil.Map    `json:"attributes,omitempty"`
}
