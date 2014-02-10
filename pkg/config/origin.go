package config

import (
	"regexp"
	"time"
)

// OriginFilterConfig represents a filter entry for a Origin item.
type OriginFilterConfig struct {
	Pattern       string         `json:"pattern"`
	Rewrite       string         `json:"rewrite"`
	Discard       bool           `json:"discard"`
	Target        string         `json:"target"`
	PatternRegexp *regexp.Regexp `json:"-"`
}

// OriginConfig represents a Origin entry in the configuration system.
type OriginConfig struct {
	Backend   map[string]string          `json:"backend"`
	Filters   []*OriginFilterConfig      `json:"filters"`
	Templates map[string]*TemplateConfig `json:"templates"`
	Modified  time.Time                  `json:"-"`
}
