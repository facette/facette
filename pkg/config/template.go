package config

import (
	"regexp"
)

// TemplateConfig represents a template entry in the configuration system.
type TemplateConfig struct {
	SplitPattern string                 `json:"split_pattern"`
	StackMode    int                    `json:"stack_mode"`
	Stacks       []*TemplateStackConfig `json:"stacks"`
	Options      map[string]string      `json:"options"`
	SplitRegexp  *regexp.Regexp         `json:"-"`
}

// TemplateStackConfig represents a stack entry in a TemplateConfig instance.
type TemplateStackConfig struct {
	Groups map[string]*TemplateGroupConfig `json:"groups"`
}

// TemplateGroupConfig represents a group entry in a TemplateStackConfig instance.
type TemplateGroupConfig struct {
	Type    int    `json:"type"`
	Pattern string `json:"pattern"`
}
