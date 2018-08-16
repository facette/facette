package catalog

import (
	"fmt"

	"facette.io/maputil"
)

// Record represents a catalog record instance.
type Record struct {
	Origin     string
	Source     string
	Metric     string
	Attributes *maputil.Map
}

func (r Record) String() string {
	return fmt.Sprintf("{Origin: %q, Source: %q, Metric: %q}", r.Origin, r.Source, r.Metric)
}
