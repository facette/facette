package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"facette.io/facette/timerange"
	"facette.io/maputil"
	"gopkg.in/yaml.v2"
)

// Config represents a configuration instance.
type Config struct {
	Logger   *LoggerConfig   `yaml:"logger"`
	HTTP     *HTTPConfig     `yaml:"http"`
	Storage  *maputil.Map    `yaml:"storage"`
	Defaults *DefaultsConfig `yaml:"defaults"`
}

// New creates a new configuration instance, initializing its content based on a provided configuration file.
func New(path string) (*Config, error) {
	config := &Config{
		Logger:   newLoggerConfig(),
		HTTP:     newHTTPConfig(),
		Storage:  newStorageConfig(),
		Defaults: newDefaultsConfig(),
	}

	if path != "" {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(data, config)
		if err != nil {
			return nil, err
		}
	}

	err := normalizeHTTPConfig(config.HTTP)
	if err != nil {
		return nil, err
	}

	// Normalize settings and check for their validity
	config.HTTP.BasePath = strings.TrimSuffix(config.HTTP.BasePath, "/")

	if !timerange.IsValid(config.Defaults.TimeRange) {
		return nil, fmt.Errorf("invalid default time range %q", config.Defaults.TimeRange)
	}

	return config, nil
}
