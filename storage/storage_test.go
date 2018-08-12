package storage

import "facette.io/logger"

var log *logger.Logger

func init() {
	var err error

	log, err = logger.NewLogger(logger.FileConfig{})
	if err != nil {
		panic("failed to initialize logger")
	}
}
