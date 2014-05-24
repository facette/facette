package config

import (
	"regexp"
	"time"
)

// OriginConfig represents an origin entry in the configuration system.
type OriginConfig struct {
	Connector   map[string]string          `json:"connector"`
	Filters     []*OriginFilterConfig      `json:"filters"`
	Templates   map[string]*TemplateConfig `json:"templates"`
	SelfRefresh int                        `json:"self_refresh"`
	Modified    time.Time                  `json:"-"`
}

// OriginFilterConfig represents a filter entry in an OriginConfig instance.
type OriginFilterConfig struct {
	Pattern       string         `json:"pattern"`
	Rewrite       string         `json:"rewrite"`
	Discard       bool           `json:"discard"`
	Target        string         `json:"target"`
	PatternRegexp *regexp.Regexp `json:"-"`
}
