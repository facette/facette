// Package connector implements the connectors handling third-party data sources.
package connector

import (
	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/plot"
)

const (
	_ = iota
	// OperGroupTypeNone represents a null operation group mode.
	OperGroupTypeNone
	// OperGroupTypeAvg represents a AVG operation group mode.
	OperGroupTypeAvg
	// OperGroupTypeSum represents a SUM operation group mode.
	OperGroupTypeSum
)

// Connector represents the main interface of a connector handler.
type Connector interface {
	GetPlots(query *plot.Query) ([]*plot.Series, error)
	Refresh(string, chan *catalog.Record) error
}

var (
	// Connectors represents the list of all available connector handlers.
	Connectors = make(map[string]func(map[string]interface{}) (Connector, error))
)
