package config

// LoggerConfig represents a logger configuration instance.
type LoggerConfig struct {
	File   *LoggerFileConfig   `yaml:"file"`
	Syslog *LoggerSyslogConfig `yaml:"syslog"`
}

func newLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		File: &LoggerFileConfig{
			Level: "info",
			Path:  "",
		},
	}
}

// LoggerFileConfig represents a file logger configuration instance.
type LoggerFileConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

// LoggerSyslogConfig represents a Syslog logger configuration instance.
type LoggerSyslogConfig struct {
	Level     string `yaml:"level"`
	Facility  string `yaml:"facility"`
	Tag       string `yaml:"tag"`
	Address   string `yaml:"address"`
	Transport string `yaml:"transport"`
}
