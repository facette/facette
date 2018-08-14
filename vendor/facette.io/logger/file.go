package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mgutz/ansi"
)

var (
	fileColors = map[int]string{
		LevelError:   "red",
		LevelWarning: "yellow",
		LevelNotice:  "magenta",
		LevelInfo:    "blue",
		LevelDebug:   "cyan",
	}

	fileLabels map[int]string
)

type fileBackend struct {
	logger *Logger
	output *os.File
	writer *log.Logger
	level  int
}

func newFileBackend(config FileConfig, logger *Logger) (backend, error) {
	var (
		output    *os.File
		useColors bool
		err       error
	)

	if config.Level == "" {
		config.Level = defaultLevel
	} else if _, ok := levelMap[config.Level]; !ok {
		return nil, ErrInvalidLevel
	}

	if config.Path != "" && config.Path != "-" {
		// Create parent folders if needed
		dirPath, _ := path.Split(config.Path)

		if dirPath != "" {
			if err = os.MkdirAll(dirPath, 0755); err != nil {
				return nil, err
			}
		}

		// Open logging output file
		if output, err = os.OpenFile(config.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return nil, fmt.Errorf("failed to open logging file: %s", err)
		}
	} else {
		// Set logging output to stderr
		output = os.Stderr
		useColors = true
	}

	writer := log.New(output, "", log.LstdFlags|log.Lmicroseconds)

	// Initialize labels
	fileLabels = map[int]string{}

	for name, level := range levelMap {
		if useColors {
			fileLabels[level] = ansi.Color(strings.ToUpper(name), fileColors[level])
		} else {
			fileLabels[level] = strings.ToUpper(name) + ":"
		}
	}

	return &fileBackend{
		logger: logger,
		output: output,
		writer: writer,
		level:  levelMap[config.Level],
	}, nil
}

func (b fileBackend) Close() {
	b.output.Close()
}

func (b fileBackend) Write(level int, mesg string) {
	if level <= b.level {
		b.writer.Printf("%s %s", fileLabels[level], mesg)
	}
}
