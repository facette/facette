package config

import (
	"net/url"
	"strings"
)

// HTTPConfig represents a HTTP configuration instance.
type HTTPConfig struct {
	Listen          string `yaml:"listen"`
	GracefulTimeout int    `yaml:"graceful_timeout"`
	BasePath        string `yaml:"base_path"`
	ReadOnly        bool   `yaml:"read_only"`
	EnableUI        bool   `yaml:"enable_ui"`
	ExposeVersion   bool   `yaml:"expose_version"`

	SocketMode  string
	SocketUser  string
	SocketGroup string
}

func newHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Listen:          "localhost:12003",
		GracefulTimeout: 30,
		BasePath:        "/",
		ReadOnly:        false,
		EnableUI:        true,
		ExposeVersion:   true,
	}
}

func normalizeHTTPConfig(config *HTTPConfig) error {
	if !strings.HasPrefix(config.Listen, "unix:") {
		return nil
	}

	idx := strings.IndexByte(config.Listen, '?')
	if idx != -1 {
		values, err := url.ParseQuery(config.Listen[idx+1:])
		if err != nil {
			return err
		}

		config.Listen = config.Listen[:idx]
		config.SocketMode = values.Get("mode")
		config.SocketUser = values.Get("user")
		config.SocketGroup = values.Get("group")
	}

	return nil
}
