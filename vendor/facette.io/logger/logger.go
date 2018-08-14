// Package logger is a simple wrapper around log.Logger with usual logging levels "error", "warning", "notice", "info"
// and "debug".
package logger

import (
	"fmt"
	"log"
	"sync"
)

const defaultLevel = "info"

const (
	_ = iota
	// LevelError represents the error logging level.
	LevelError
	// LevelWarning represents the warning logging level.
	LevelWarning
	// LevelNotice represents the notice logging level.
	LevelNotice
	// LevelInfo represents the info logging level.
	LevelInfo
	// LevelDebug represents the debug logging level.
	LevelDebug
)

var (
	logger   *Logger
	levelMap = map[string]int{
		"error":   LevelError,
		"warning": LevelWarning,
		"notice":  LevelNotice,
		"info":    LevelInfo,
		"debug":   LevelDebug,
	}
)

// Logger represents a logger instance.
type Logger struct {
	backends []backend
	context  string

	wg sync.WaitGroup

	sync.Mutex
}

func init() {
	logger, _ = NewLogger(FileConfig{Level: "debug"})
}

// NewLogger returns a new Logger instance initialized with the given configuration.
// If no configs are passed are parameter, log messages will effectively be discarded.
func NewLogger(configs ...interface{}) (*Logger, error) {
	// Initialize logger backends
	l := &Logger{
		backends: []backend{},
		wg:       sync.WaitGroup{},
	}

	for _, config := range configs {
		var (
			b   backend
			err error
		)

		switch config.(type) {
		case FileConfig:
			b, err = newFileBackend(config.(FileConfig), l)

		case SyslogConfig:
			b, err = newSyslogBackend(config.(SyslogConfig), l)

		default:
			err = ErrUnsupportedBackend
		}

		if err != nil {
			return nil, err
		}

		l.backends = append(l.backends, b)
	}

	return l, nil
}

// Logger returns a log.Logger instance for a given logging level.
func (l *Logger) Logger(level int) *log.Logger {
	return log.New(newWriter(l, level), "", 0)
}

// Context clones the Logger instance and sets the context to the provided string.
func (l *Logger) Context(context string) *Logger {
	return &Logger{
		backends: l.backends,
		context:  context,
		wg:       sync.WaitGroup{},
	}
}

// CurrentContext returns the current logger context.
func (l *Logger) CurrentContext() string {
	return l.context
}

// Error prints an error message in the logging system.
func (l *Logger) Error(format string, v ...interface{}) *Logger {
	l.write(LevelError, format, v...)
	return l
}

// Error prints an error message using the default logger.
func Error(format string, v ...interface{}) {
	logger.Error(format, v...)
}

// Warning prints a warning message in the logging system.
func (l *Logger) Warning(format string, v ...interface{}) *Logger {
	l.write(LevelWarning, format, v...)
	return l
}

// Warning prints a warning message using the default logger.
func Warning(format string, v ...interface{}) {
	logger.Warning(format, v...)
}

// Notice prints a notice message in the logging system.
func (l *Logger) Notice(format string, v ...interface{}) *Logger {
	l.write(LevelNotice, format, v...)
	return l
}

// Notice prints a notice message using the default logger.
func Notice(format string, v ...interface{}) {
	logger.Notice(format, v...)
}

// Info prints an information message in the logging system.
func (l *Logger) Info(format string, v ...interface{}) *Logger {
	l.write(LevelInfo, format, v...)
	return l
}

// Info prints an information message using the default logger.
func Info(format string, v ...interface{}) {
	logger.Info(format, v...)
}

// Debug prints a debug message in the logging system.
func (l *Logger) Debug(format string, v ...interface{}) *Logger {
	l.write(LevelDebug, format, v...)
	return l
}

// Debug prints a debug message using the default logger.
func Debug(format string, v ...interface{}) {
	logger.Debug(format, v...)
}

// Close closes the logger output file.
func (l *Logger) Close() {
	for _, b := range l.backends {
		b.Close()
	}
}

func (l *Logger) write(level int, format string, v ...interface{}) {
	var mesg string

	l.Lock()
	defer l.Unlock()

	// Set message
	if l.context != "" {
		mesg = l.context + ": "
	}

	if len(v) > 0 {
		mesg += fmt.Sprintf(format, v...)
	} else {
		mesg += format
	}

	// Write messages to backends
	l.wg.Add(len(l.backends))

	for _, b := range l.backends {
		go func(b backend) {
			b.Write(level, mesg)
			l.wg.Done()
		}(b)
	}

	l.wg.Wait()
}
