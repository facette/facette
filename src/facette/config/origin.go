package config

import (
	"regexp"
	"time"
)

// OriginFilter represents a filter entry for a Origin item.
type OriginFilter struct {
	Pattern       string         `json:"pattern"`
	Rewrite       string         `json:"rewrite"`
	Discard       bool           `json:"discard"`
	PatternRegexp *regexp.Regexp `json:"-"`
}

// OriginConfig represents a Origin entry in the configuration system.
type OriginConfig struct {
	Backend   map[string]string          `json:"backend"`
	Filters   []*OriginFilter            `json:"filters"`
	Templates map[string]*TemplateConfig `json:"templates"`
	Modified  time.Time                  `json:"-"`
}
