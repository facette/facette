package series

import (
	"time"

	"facette.io/facette/catalog"
)

// Query represents a time series point query instance.
type Query struct {
	StartTime time.Time
	EndTime   time.Time
	Sample    int
	Metrics   []*catalog.Metric
}
