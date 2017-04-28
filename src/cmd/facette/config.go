package main

import (
	"strings"

	"facette/yamlutil"

	"github.com/facette/maputil"
)

const (
	defaultListen            = "localhost:12003"
	defaultGracefulTimeout   = 30
	defaultRootPath          = "/"
	defaultLogPath           = ""
	defaultLogLevel          = "info"
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
	GracefulTimeout  int            `yaml:"graceful_timeout"`
	RootPath         string         `yaml:"root_path"`
	LogPath          string         `yaml:"log_path"`
	LogLevel         string         `yaml:"log_level"`
	Frontend         frontendConfig `yaml:"frontend"`
	Backend          *maputil.Map   `yaml:"backend"`
	HideBuildDetails bool           `yaml:"hide_build_details"`
	ReadOnly         bool           `yaml:"read_only"`
}

func initConfig(path string) (*config, error) {
	var (
		config = config{
			Listen:          defaultListen,
			GracefulTimeout: defaultGracefulTimeout,
			RootPath:        defaultRootPath,
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

	config.RootPath = strings.TrimSuffix(config.RootPath, "/")
	config.Frontend.AssetsDir = strings.TrimSuffix(config.Frontend.AssetsDir, "/")

	return &config, nil
}
