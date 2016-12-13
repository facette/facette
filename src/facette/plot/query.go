package plot

import "time"

// Query represents a time series plot query instance.
type Query struct {
	StartTime time.Time
	EndTime   time.Time
	Sample    int
	Series    []QuerySeries
}

// QuerySeries represents a series instance in a time series plot query.
type QuerySeries struct {
	Origin string
	Source string
	Metric string
}
