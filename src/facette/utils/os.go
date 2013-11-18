package utils

import (
	"os"
	"path/filepath"
)

func walkDir(dirPath string, linkPath string, walkFunc filepath.WalkFunc) error {
	var (
		err error
	)

	if _, err = os.Stat(dirPath); err != nil {
		return err
	}

	// Search for files recursively
	internalFunc := func(filePath string, fileInfo os.FileInfo, err error) error {
		var (
			mode     os.FileMode
			realPath string
		)

		mode = fileInfo.Mode() & os.ModeType

		if mode == os.ModeSymlink {
			if realPath, err = filepath.EvalSymlinks(filePath); err != nil {
				return err
			}

			return walkDir(realPath, filePath, walkFunc)
		} else if linkPath != "" {
			return walkFunc(linkPath+filePath[len(dirPath):], fileInfo, err)
		} else {
			return walkFunc(filePath, fileInfo, err)
		}
	}

	return filepath.Walk(dirPath, internalFunc)
}

// WalkDir browses dirPath on the filesystem executing walkFunc for each found files.
func WalkDir(dirPath string, walkFunc filepath.WalkFunc) error {
	return walkDir(dirPath, "", walkFunc)
}
