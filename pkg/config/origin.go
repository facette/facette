package config

import (
	"regexp"
	"time"
)

// OriginConfig represents an origin entry in the configuration system.
type OriginConfig struct {
	Connector       map[string]interface{} `json:"connector"`
	Filters         []*OriginFilterConfig  `json:"filters"`
	RefreshInterval int                    `json:"refresh_interval"`
	Modified        time.Time              `json:"-"`
}

// OriginFilterConfig represents a filtering rule in an OriginConfig instance.
type OriginFilterConfig struct {
	Pattern       string         `json:"pattern"`
	Rewrite       string         `json:"rewrite"`
	Discard       bool           `json:"discard"`
	Target        string         `json:"target"`
	PatternRegexp *regexp.Regexp `json:"-"`
}
