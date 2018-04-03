package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"facette/timerange"

	"github.com/facette/maputil"
	"gopkg.in/yaml.v2"
)

const (
	defaultListen            = "localhost:12003"
	defaultGracefulTimeout   = 30
	defaultRootPath          = "/"
	defaultLogPath           = ""
	defaultLogLevel          = "info"
	defaultSyslogFacility    = "daemon"
	defaultSyslogTag         = "facette"
	defaultFrontendEnabled   = true
	defaultFrontendAssetsDir = "assets"
	defaultTimeRange         = "-1h"
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
	SyslogLevel      string         `yaml:"syslog_level"`
	SyslogFacility   string         `yaml:"syslog_facility"`
	SyslogTag        string         `yaml:"syslog_tag"`
	SyslogAddress    string         `yaml:"syslog_address"`
	SyslogTransport  string         `yaml:"syslog_transport"`
	Frontend         frontendConfig `yaml:"frontend"`
	Backend          *maputil.Map   `yaml:"backend"`
	DefaultTimeRange string         `yaml:"default_time_range"`
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
			SyslogFacility:  defaultSyslogFacility,
			SyslogTag:       defaultSyslogTag,
			Frontend: frontendConfig{
				Enabled:   defaultFrontendEnabled,
				AssetsDir: defaultFrontendAssetsDir,
			},
			DefaultTimeRange: defaultTimeRange,
			HideBuildDetails: defaultHideBuildDetails,
		}
	)

	if path != "" {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}

	config.RootPath = strings.TrimSuffix(config.RootPath, "/")
	config.Frontend.AssetsDir = strings.TrimSuffix(config.Frontend.AssetsDir, "/")

	// Check for settings validity
	if !timerange.IsValid(config.DefaultTimeRange) {
		return nil, fmt.Errorf("invalid default time range %q", config.DefaultTimeRange)
	}

	return &config, nil
}
