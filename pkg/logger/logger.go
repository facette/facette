// Package logger implements the internal logging facilities.
package logger

import (
	"fmt"
	"io"
	"log"
)

const (
	_ = iota
	// LevelError represents a error logging level.
	LevelError
	// LevelWarning represents a warning logging level.
	LevelWarning
	// LevelNotice represents a notice logging level.
	LevelNotice
	// LevelInfo represents a info logging level.
	LevelInfo
	// LevelDebug represents a debug logging level.
	LevelDebug
)

var (
	logLevel   = LevelWarning
	levelNames = map[string]int{
		"error":   LevelError,
		"warning": LevelWarning,
		"notice":  LevelNotice,
		"info":    LevelInfo,
		"debug":   LevelDebug,
	}
)

// GetLevelByName returns the numeric value of logging level matching a level name.
func GetLevelByName(name string) (int, error) {
	level, ok := levelNames[name]
	if !ok {
		return 0, fmt.Errorf("invalid level `%s'", name)
	}

	return level, nil
}

// SetLevel sets the global logging level.
func SetLevel(level int) {
	logLevel = level

	if logLevel < LevelError || logLevel > LevelDebug {
		logLevel = LevelInfo
	}
}

// SetOutput sets the global logging output.
func SetOutput(output io.Writer) {
	log.SetOutput(output)
}

// Log logs a message.
func Log(level int, context, format string, v ...interface{}) {
	var criticity string

	if level > logLevel {
		return
	}

	switch level {
	case LevelError:
		criticity = "ERROR"
	case LevelWarning:
		criticity = "WARNING"
	case LevelNotice:
		criticity = "NOTICE"
	case LevelInfo:
		criticity = "INFO"
	case LevelDebug:
		criticity = "DEBUG"
	}

	log.Printf(
		"%s: %s",
		fmt.Sprintf("%s: %s", criticity, context),
		fmt.Sprintf(format, v...),
	)
}
