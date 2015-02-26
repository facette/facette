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
	// DefaultBindAddr represents the default server binding address.
	DefaultBindAddr string = "tcp://localhost:12003"
	// DefaultSocketUser indicates that the socket owning user should not be changed
	DefaultSocketUser int = -1
	// DefaultSocketGroup indicates that the socket owning group should not be changed
	DefaultSocketGroup int = -1
	// DefaultBaseDir represents the default server base directory location.
	DefaultBaseDir string = "/usr/share/facette"
	// DefaultDataDir represents the default internal data files directory location.
	DefaultDataDir string = "/var/lib/facette"
	// DefaultProvidersDir represents the default providers definition files directory location.
	DefaultProvidersDir string = "/etc/facette/providers"
	// DefaultPidFile represents the default server process PID file location.
	DefaultPidFile string = "/var/run/facette/facette.pid"
	// DefaultPlotSample represents the default plot sample for graph querying.
	DefaultPlotSample int = 400
)

// Config represents the global configuration of the instance.
type Config struct {
	BindAddr     string                     `json:"bind"`
	SocketUser   int                        `json:"socket_user,string"`
	SocketGroup  int                        `json:"socket_group,string"`
	SocketMode   *string                    `json:"socket_mode"`
	BaseDir      string                     `json:"base_dir"`
	DataDir      string                     `json:"data_dir"`
	ProvidersDir string                     `json:"providers_dir"`
	PidFile      string                     `json:"pid_file"`
	URLPrefix    string                     `json:"url_prefix"`
	ReadOnly     bool                       `json:"read_only"`
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
			err = fmt.Errorf("in %s, %s", filePath, err)
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

	return nil
}

func getSetting(config map[string]interface{}, setting string, kind reflect.Kind,
	mandatory bool, fallbackValue interface{}) (interface{}, error) {

	if _, ok := config[setting]; !ok {
		if mandatory {
			return fallbackValue, fmt.Errorf("missing mandatory setting `%s'", setting)
		}

		return fallbackValue, nil
	}

	configKind := reflect.ValueOf(config[setting]).Kind()
	if configKind != kind {
		return fallbackValue, fmt.Errorf("setting `%s' value should be of type %s and not %s", setting, kind.String(),
			configKind.String())
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
	value, err := getSetting(config, setting, reflect.Float64, mandatory, 0.0)
	return int(value.(float64)), err
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

// GetStringSlice returns the string slice of a configuration setting.
func GetStringSlice(config map[string]interface{}, setting string, mandatory bool) ([]string, error) {
	value, err := getSetting(config, setting, reflect.Slice, mandatory, nil)
	if err != nil || value == nil {
		return nil, err
	}
	array := make([]string, 0)
	for _, v := range value.([]interface{}) {
		if reflect.ValueOf(v).Kind() != reflect.String {
			return nil, fmt.Errorf("setting `%s' should be slice of strings and not %s", setting,
				reflect.ValueOf(v).Kind().String())
		} else {
			array = append(array, v.(string))
		}
	}
	return array, nil
}

// GetJsonObj returns the JSON Object interface{} of a configuration setting.
func GetJsonObj(config map[string]interface{}, setting string, mandatory bool) (interface{}, error) {
	value, err := getSetting(config, setting, reflect.Map, mandatory, nil)
	return value, err
}

// GetJsonArray returns the JSON Array interface{} of a configuration setting.
func GetJsonArray(config map[string]interface{}, setting string, mandatory bool) (interface{}, error) {
	value, err := getSetting(config, setting, reflect.Slice, mandatory, nil)
	return value, err
}
