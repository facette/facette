package config

import "regexp"

// ProviderConfig represents a provider definition in the configuration system.
type ProviderConfig struct {
	Connector       map[string]interface{}  `json:"connector"`
	Filters         []*ProviderFilterConfig `json:"filters"`
	RefreshInterval int                     `json:"refresh_interval"`
}

// ProviderFilterConfig represents a filtering rule in an ProviderConfig instance.
type ProviderFilterConfig struct {
	Pattern       string         `json:"pattern"`
	Rewrite       string         `json:"rewrite"`
	Discard       bool           `json:"discard"`
	Sieve         bool           `json:"sieve"`
	Target        string         `json:"target"`
	PatternRegexp *regexp.Regexp `json:"-"`
}
