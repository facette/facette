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

// JSONDump dumps the data structure in JSON format in filePath on the filesystem.
func JSONDump(filePath string, data interface{}, modTime time.Time) error {
	var (
		dirPath string
		err     error
		fd      *os.File
		output  []byte
	)

	if output, err = json.MarshalIndent(data, "", "    "); err != nil {
		return err
	}

	dirPath, _ = path.Split(filePath)
	os.MkdirAll(dirPath, 0755)

	fd, _ = os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	defer fd.Close()

	fd.Write(output)
	fd.Write([]byte("\n"))

	fd.Close()

	os.Chtimes(filePath, modTime, modTime)

	return nil
}

// JSONLoad loads the JSON formatted data in result from filePath on the filesystem.
func JSONLoad(filePath string, result interface{}) (os.FileInfo, error) {
	var (
		data     []byte
		err      error
		fd       *os.File
		fileInfo os.FileInfo
	)

	if _, err = os.Stat(filePath); err != nil {
		return nil, err
	}

	// Load JSON data from file
	fd, _ = os.OpenFile(filePath, os.O_RDONLY, 0644)
	defer fd.Close()

	if data, err = ioutil.ReadFile(filePath); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, result); err != nil {
		return nil, jsonError(string(data), err)
	}

	fd.Close()

	if fileInfo, err = os.Stat(filePath); err != nil {
		return nil, err
	}

	return fileInfo, nil
}

func jsonError(data string, err error) error {
	var (
		line      int
		lineStart int
		position  int
		ok        bool
		syntax    *json.SyntaxError
	)

	if syntax, ok = err.(*json.SyntaxError); !ok {
		return err
	}

	lineStart = strings.LastIndex(data[:syntax.Offset], "\n")
	line, position = strings.Count(data[:syntax.Offset], "\n")+1, int(syntax.Offset)-lineStart-1

	return fmt.Errorf("%s (line: %d, pos: %d)", err.Error(), line, position)
}
