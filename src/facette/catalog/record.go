package catalog

import "fmt"

// Record represents a catalog record.
type Record struct {
	Origin         string
	Source         string
	Metric         string
	OriginalOrigin string
	OriginalSource string
	OriginalMetric string
	Connector      interface{}
}

func (r Record) String() string {
	return fmt.Sprintf("{Origin: %q, Source: %q, Metric: %q}", r.Origin, r.Source, r.Metric)
}
