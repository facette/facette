package backend

import (
	"fmt"
	"log"
)

// A Origin represents an origin entry.
type Origin struct {
	Name      string
	Backend   BackendHandler
	Sources   map[string]*Source
	catalog   *Catalog
	inputChan chan [2]string
}

// AppendSource adds a new Source entry into the Origin instance.
func (origin *Origin) AppendSource(name, origName string) *Source {
	if origin.catalog.debugLevel > 2 {
		log.Printf("DEBUG: appending `%s' source into origin...\n", name)
	}

	// Append new source instance into origin
	source := &Source{Name: name, OriginalName: origName, Metrics: make(map[string]*Metric), origin: origin}
	origin.Sources[name] = source

	return source
}

// Update updates the current Origin by parsing the filesystem for sources or metrics.
func (origin *Origin) Update() error {
	if origin.catalog.debugLevel > 1 {
		log.Printf("DEBUG: updating origin `%s'...\n", origin.Name)
	}

	origin.Sources = make(map[string]*Source)

	if origin.Backend == nil {
		return fmt.Errorf("backend for `%s' origin is not initialized", origin.Name)
	}

	return origin.Backend.Update()
}
