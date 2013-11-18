package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
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
		return nil, err
	}

	fd.Close()

	if fileInfo, err = os.Stat(filePath); err != nil {
		return nil, err
	}

	return fileInfo, nil
}
