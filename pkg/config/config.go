// Package config implements the service configuration handling.
package config

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/facette/facette/pkg/utils"
)

const (
	// DefaultConfigFile represents the default configuration file location.
	DefaultConfigFile string = "/etc/facette/facette.json"
	// DefaultLogFile represents the default log file location.
	DefaultLogFile string = ""
	// DefaultPlotSample represents the default plot sample for graph querying.
	DefaultPlotSample int = 400
)

// Config represents the global configuration of the instance.
type Config struct {
	Path         string                     `json:"-"`
	LogFile      string                     `json:"-"`
	BindAddr     string                     `json:"bind"`
	BaseDir      string                     `json:"base_dir"`
	DataDir      string                     `json:"data_dir"`
	ProvidersDir string                     `json:"providers_dir"`
	PidFile      string                     `json:"pid_file"`
	URLPrefix    string                     `json:"url_prefix"`
	Providers    map[string]*ProviderConfig `json:"-"`
}

// Load loads the configuration from the filesystem.
func (config *Config) Load(filePath string) error {
	var errOutput error

	_, err := utils.JSONLoad(filePath, &config)
	if err != nil {
		return err
	}

	// Load provider definitions
	config.Providers = make(map[string]*ProviderConfig)

	walkFunc := func(filePath string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() || !strings.HasSuffix(filePath, ".json") {
			return nil
		}

		_, providerName := path.Split(strings.TrimSuffix(filePath, ".json"))

		config.Providers[providerName] = &ProviderConfig{}

		if fileInfo, err = utils.JSONLoad(filePath, config.Providers[providerName]); err != nil {
			err = fmt.Errorf("in %s, %s", filePath, err.Error())
			if errOutput == nil {
				errOutput = err
			}

			return err
		}

		return nil
	}

	if err := utils.WalkDir(config.ProvidersDir, walkFunc); err != nil {
		return fmt.Errorf("unable to load provider definitions: %s", err)
	}

	if errOutput != nil {
		return errOutput
	}

	config.Path = filePath

	return nil
}

// Reload reloads the configuration.
func (config *Config) Reload() error {
	return config.Load(config.Path)
}

func getSetting(config map[string]interface{}, setting string, kind reflect.Kind,
	mandatory bool, fallbackValue interface{}) (interface{}, error) {
	if _, ok := config[setting]; !ok {
		if mandatory {
			return fallbackValue, fmt.Errorf("missing `%s' mandatory setting", setting)
		}

		return fallbackValue, nil
	}

	if reflect.ValueOf(config[setting]).Kind() != kind {
		return fallbackValue, fmt.Errorf("setting `%s' value should be a %s", setting, kind.String())
	}

	return config[setting], nil
}

// GetString returns the string value of a configuration setting.
func GetString(config map[string]interface{}, setting string, mandatory bool) (string, error) {
	value, err := getSetting(config, setting, reflect.String, mandatory, "")
	return value.(string), err
}

// GetInt returns the int value of a configuration setting.
func GetInt(config map[string]interface{}, setting string, mandatory bool) (int, error) {
	value, err := getSetting(config, setting, reflect.Int, mandatory, 0)
	return value.(int), err
}

// GetFloat returns the float value of a configuration setting.
func GetFloat(config map[string]interface{}, setting string, mandatory bool) (float64, error) {
	value, err := getSetting(config, setting, reflect.Float64, mandatory, 0.0)
	return value.(float64), err
}

// GetBool returns the bool value of a configuration setting.
func GetBool(config map[string]interface{}, setting string, mandatory bool) (bool, error) {
	value, err := getSetting(config, setting, reflect.Bool, mandatory, false)
	return value.(bool), err
}
