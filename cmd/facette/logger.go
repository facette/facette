package main

import (
	"facette.io/facette/config"
	"facette.io/logger"
)

func newLogger(config *config.Config) (*logger.Logger, error) {
	var loggers []interface{}

	if config.Logger.File != nil {
		loggers = append(loggers, logger.FileConfig{
			Level: config.Logger.File.Level,
			Path:  config.Logger.File.Path,
		})
	}

	if config.Logger.Syslog != nil {
		loggers = append(loggers, logger.SyslogConfig{
			Level:     config.Logger.Syslog.Level,
			Facility:  config.Logger.Syslog.Facility,
			Tag:       config.Logger.Syslog.Tag,
			Address:   config.Logger.Syslog.Address,
			Transport: config.Logger.Syslog.Transport,
		})
	}

	return logger.NewLogger(loggers...)
}
