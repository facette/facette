package main

import (
	"facette.io/facette/config"
	"facette.io/logger"
)

func newLogger(config *config.Config) (*logger.Logger, error) {
	var loggers []interface{}

	if config.LogPath != "" {
		loggers = append(loggers, logger.FileConfig{
			Level: config.LogLevel,
			Path:  config.LogPath,
		})
	}

	if config.SyslogLevel != "" {
		loggers = append(loggers, logger.SyslogConfig{
			Level:     config.SyslogLevel,
			Facility:  config.SyslogFacility,
			Tag:       config.SyslogTag,
			Address:   config.SyslogAddress,
			Transport: config.SyslogTransport,
		})
	}

	return logger.NewLogger(loggers...)
}
