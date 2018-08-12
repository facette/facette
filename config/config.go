package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"facette.io/facette/timerange"
	"facette.io/maputil"
	"gopkg.in/yaml.v2"
)

const (
	defaultListen           = "localhost:12003"
	defaultGracefulTimeout  = 30
	defaultRootPath         = "/"
	defaultLogPath          = "-"
	defaultLogLevel         = "info"
	defaultSyslogFacility   = "daemon"
	defaultSyslogTag        = "facette"
	defaultFrontendEnabled  = true
	defaultTimeRange        = "-1h"
	defaultHideBuildDetails = false
)

// Config represents a configuration instance.
type Config struct {
	Listen           string          `yaml:"listen"`
	SocketMode       string          `yaml:"socket_mode"`
	SocketUser       string          `yaml:"socket_user"`
	SocketGroup      string          `yaml:"socket_group"`
	GracefulTimeout  int             `yaml:"graceful_timeout"`
	RootPath         string          `yaml:"root_path"`
	LogPath          string          `yaml:"log_path"`
	LogLevel         string          `yaml:"log_level"`
	SyslogLevel      string          `yaml:"syslog_level"`
	SyslogFacility   string          `yaml:"syslog_facility"`
	SyslogTag        string          `yaml:"syslog_tag"`
	SyslogAddress    string          `yaml:"syslog_address"`
	SyslogTransport  string          `yaml:"syslog_transport"`
	Frontend         *FrontendConfig `yaml:"frontend"`
	Backend          *maputil.Map    `yaml:"backend"`
	DefaultTimeRange string          `yaml:"default_time_range"`
	HideBuildDetails bool            `yaml:"hide_build_details"`
	ReadOnly         bool            `yaml:"read_only"`
}

// New creates a new configuration instance, initializing its content based on a provided configuration file.
func New(path string) (*Config, error) {
	config := &Config{
		Listen:          defaultListen,
		GracefulTimeout: defaultGracefulTimeout,
		RootPath:        defaultRootPath,
		LogPath:         defaultLogPath,
		LogLevel:        defaultLogLevel,
		SyslogFacility:  defaultSyslogFacility,
		SyslogTag:       defaultSyslogTag,
		Frontend: &FrontendConfig{
			Enabled: defaultFrontendEnabled,
		},
		DefaultTimeRange: defaultTimeRange,
		HideBuildDetails: defaultHideBuildDetails,
	}

	if path != "" {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	config.RootPath = strings.TrimSuffix(config.RootPath, "/")

	// Check for settings validity
	if !timerange.IsValid(config.DefaultTimeRange) {
		return nil, fmt.Errorf("invalid default time range %q", config.DefaultTimeRange)
	}

	return config, nil
}
