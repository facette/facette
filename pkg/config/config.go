package config

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/facette/facette/pkg/utils"
)

const (
	// DefaultConfigFile represents the default configuration file location.
	DefaultConfigFile = "/etc/facette/facette.json"
)

// Config represents the main configuration system structure.
type Config struct {
	Path      string                   `json:"-"`
	BindAddr  string                   `json:"bind"`
	BaseDir   string                   `json:"base_dir"`
	DataDir   string                   `json:"data_dir"`
	OriginDir string                   `json:"origin_dir"`
	PidFile   string                   `json:"pid_file"`
	ServerLog string                   `json:"server_log"`
	AccessLog string                   `json:"access_log"`
	URLPrefix string                   `json:"url_prefix"`
	Auth      map[string]string        `json:"auth"`
	Scales    [][2]interface{}         `json:"scales"`
	Origins   map[string]*OriginConfig `json:"-"`
}

// Load loads the configuration from the filesystem using the filePath paramater as origin path.
func (config *Config) Load(filePath string) error {
	var errOutput error

	_, err := utils.JSONLoad(filePath, &config)
	if err != nil {
		return err
	}

	// Load origin definitions
	config.Origins = make(map[string]*OriginConfig)

	walkFunc := func(filePath string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() || !strings.HasSuffix(filePath, ".json") {
			return nil
		}

		_, originName := path.Split(filePath[:len(filePath)-5])

		config.Origins[originName] = &OriginConfig{}

		if fileInfo, err = utils.JSONLoad(filePath, config.Origins[originName]); err != nil {
			err = fmt.Errorf("in %s, %s", filePath, err.Error())

			if errOutput == nil {
				errOutput = err
			}

			return err
		}

		config.Origins[originName].Modified = fileInfo.ModTime()

		return nil
	}

	utils.WalkDir(config.OriginDir, walkFunc)

	if errOutput != nil {
		return errOutput
	}

	// Pre-compile Regexp items
	for _, origin := range config.Origins {
		for _, filter := range origin.Filters {
			filter.PatternRegexp = regexp.MustCompile(filter.Pattern)
		}

		for _, template := range origin.Templates {
			template.SplitRegexp = regexp.MustCompile(template.SplitPattern)
		}
	}

	config.Path = filePath

	return nil
}

// Reload reloads the configuration from the filesystem.
func (config *Config) Reload() error {
	return config.Load(config.Path)
}
