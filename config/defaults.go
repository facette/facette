package config

// DefaultsConfig represents a service defaults configuration instance.
type DefaultsConfig struct {
	TimeRange string `json:"time_range" yaml:"time_range"`
}

func newDefaultsConfig() *DefaultsConfig {
	return &DefaultsConfig{
		TimeRange: "-1h",
	}
}
