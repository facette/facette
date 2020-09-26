// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package config provides main configuration management.
package config

import (
	"os"

	"gopkg.in/yaml.v3"

	httpserver "facette.io/facette/pkg/http/server"
	"facette.io/facette/pkg/logger"
	"facette.io/facette/pkg/poller"
	"facette.io/facette/pkg/store"
)

// DefaultPath is the default main configuration file path.
const DefaultPath = "/etc/facette/facette.yaml"

// Config is a main configuration.
type Config struct {
	Log    *logger.Config     `yaml:"log"`
	HTTP   *httpserver.Config `yaml:"http"`
	Poller *poller.Config     `yaml:"poller"`
	Store  *store.Config      `yaml:"store"`
}

// DefaultConfig returns a default main configuration.
func DefaultConfig() *Config {
	return &Config{
		Log:    logger.DefaultConfig(),
		HTTP:   httpserver.DefaultConfig(),
		Poller: poller.DefaultConfig(),
		Store:  store.DefaultConfig(),
	}
}

// Load loads main configuration from path, and merges it into the defaults. If
// path doesn't exist, only the default configuration is returned.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		f, err := os.OpenFile(path, os.O_RDONLY, 0) // nolint:gosec
		if os.IsNotExist(err) {
			return cfg, nil
		} else if err != nil {
			return nil, err
		}
		defer f.Close() // nolint:errcheck,gosec

		err = yaml.NewDecoder(f).Decode(cfg)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
