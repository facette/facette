package common

import (
	"regexp"
)

// TemplateGroupConfig represents a Template entry group in the configuration system.
type TemplateGroupConfig struct {
	Type    int    `json:"type"`
	Pattern string `json:"pattern"`
}

// TemplateStackConfig represents a Template entry stack in the configuration system.
type TemplateStackConfig struct {
	Groups map[string]*TemplateGroupConfig `json:"groups"`
}

// TemplateConfig represents a Template entry in the configuration system.
type TemplateConfig struct {
	SplitPattern string                 `json:"split_pattern"`
	StackMode    int                    `json:"stack_mode"`
	Stacks       []*TemplateStackConfig `json:"stacks"`
	Options      map[string]string      `json:"options"`
	SplitRegexp  *regexp.Regexp         `json:"-"`
}
