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
	Action        string         `json:"action"`
	Pattern       string         `json:"pattern"`
	Target        string         `json:"target"`
	Into          string         `json:"into"`
	PatternRegexp *regexp.Regexp `json:"-"`
}
