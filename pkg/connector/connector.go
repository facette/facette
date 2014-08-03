// Package connector implements the connectors handling third-party data sources.
package connector

import (
	"fmt"
	"regexp"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/plot"
)

// Connector represents the main interface of a connector handler.
type Connector interface {
	GetName() string
	GetPlots(query *plot.Query) ([]plot.Series, error)
	Refresh(string, chan *catalog.Record) error
}

var (
	// Connectors represents the list of all available connector handlers.
	Connectors = make(map[string]func(string, map[string]interface{}) (Connector, error))
)

func compilePattern(pattern string) (*regexp.Regexp, error) {
	var (
		re  *regexp.Regexp
		err error
	)

	// Compile regexp pattern
	if re, err = regexp.Compile(pattern); err != nil {
		return nil, err
	}

	// Validate pattern keywords
	groups := make(map[string]bool)

	for _, key := range re.SubexpNames() {
		if key == "" {
			continue
		} else if key == "source" || key == "metric" {
			groups[key] = true
		} else {
			return nil, fmt.Errorf("invalid pattern keyword `%s'", key)
		}
	}

	if !groups["source"] {
		return nil, fmt.Errorf("missing pattern keyword `source'")
	} else if !groups["metric"] {
		return nil, fmt.Errorf("missing pattern keyword `metric'")
	}

	return re, nil
}

func matchSeriesPattern(re *regexp.Regexp, series string) ([2]string, error) {
	var sourceName, metricName string

	submatch := re.FindStringSubmatch(series)
	if len(submatch) == 0 {
		return [2]string{}, fmt.Errorf("series `%s' does not match pattern", series)
	}

	if re.SubexpNames()[1] == "source" {
		sourceName = submatch[1]
		metricName = submatch[2]
	} else {
		sourceName = submatch[2]
		metricName = submatch[1]
	}

	return [2]string{sourceName, metricName}, nil
}
