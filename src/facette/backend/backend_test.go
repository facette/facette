package backend

import "github.com/facette/logger"

var log *logger.Logger

func init() {
	var err error

	log, err = logger.NewLogger(logger.FileConfig{})
	if err != nil {
		panic("failed to initialize logger")
	}
}
