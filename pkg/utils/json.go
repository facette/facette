package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// JSONDump dumps the data structure using JSON format on the filesystem.
func JSONDump(filePath string, data interface{}, modTime time.Time) error {
	output, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	dirPath, _ := path.Split(filePath)

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	fd.Write(output)
	fd.Write([]byte("\n"))

	os.Chtimes(filePath, modTime, modTime)

	return nil
}

// JSONLoad loads the JSON formatted data in result from the filesystem.
func JSONLoad(filePath string, result interface{}) (os.FileInfo, error) {
	// Load JSON data from file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil, jsonError(string(data), err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return fileInfo, nil
}

func jsonError(data string, err error) error {
	syntax, ok := err.(*json.SyntaxError)
	if !ok {
		return err
	}

	lineStart := strings.LastIndex(data[:syntax.Offset], "\n")
	line, position := strings.Count(data[:syntax.Offset], "\n")+1, int(syntax.Offset)-lineStart-1

	return fmt.Errorf("%s (line: %d, pos: %d)", err, line, position)
}
