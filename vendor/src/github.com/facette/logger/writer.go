package logger

import "io"

type writer struct {
	logger *Logger
	level  int
}

func (w writer) Write(p []byte) (int, error) {
	w.logger.write(w.level, string(p))
	return len(p), nil
}

func newWriter(logger *Logger, level int) io.Writer {
	return &writer{
		logger: logger,
		level:  level,
	}
}
