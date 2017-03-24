package main

import (
	"facette/mapper"
	"facette/yamlutil"
)

const (
	defaultListen            = "localhost:12003"
	defaultLogPath           = ""
	defaultLogLevel          = "info"
	defaultGracefulTimeout   = 30
	defaultFrontendEnabled   = true
	defaultFrontendAssetsDir = "assets"
	defaultHideBuildDetails  = false
)

type frontendConfig struct {
	Enabled   bool   `yaml:"enabled"`
	AssetsDir string `yaml:"assets_dir"`
}

type config struct {
	Listen           string         `yaml:"listen"`
	SocketMode       string         `yaml:"socket_mode"`
	SocketUser       string         `yaml:"socket_user"`
	SocketGroup      string         `yaml:"socket_group"`
	LogPath          string         `yaml:"log_path"`
	LogLevel         string         `yaml:"log_level"`
	GracefulTimeout  int            `yaml:"graceful_timeout"`
	Frontend         frontendConfig `yaml:"frontend"`
	Backend          *mapper.Map    `yaml:"backend,omitempty"`
	HideBuildDetails bool           `yaml:"hide_build_details"`
}

func initConfig(path string) (*config, error) {
	var (
		config = config{
			Listen:          defaultListen,
			GracefulTimeout: defaultGracefulTimeout,
			LogPath:         defaultLogPath,
			LogLevel:        defaultLogLevel,
			Frontend: frontendConfig{
				Enabled:   defaultFrontendEnabled,
				AssetsDir: defaultFrontendAssetsDir,
			},
			HideBuildDetails: defaultHideBuildDetails,
		}
	)

	if path != "" {
		if err := yamlutil.UnmarshalFile(path, &config); err != nil {
			return nil, err
		}
	}

	return &config, nil
}
