package utils

import (
	"os"
	"path/filepath"

	"github.com/facette/facette/pkg/logger"
)

func walkDir(dirPath string, linkPath string, walkFunc filepath.WalkFunc) error {
	if _, err := os.Stat(dirPath); err != nil {
		return err
	}

	// Search for files recursively
	return filepath.Walk(dirPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		mode := fileInfo.Mode() & os.ModeType

		if mode == os.ModeSymlink {
			realPath, err := filepath.EvalSymlinks(filePath)
			if err != nil {
				logger.Log(logger.LevelWarning, "utils", "failed to resolve symlink %q: %v", filePath, err)
				return nil
			}

			return walkDir(realPath, filePath, walkFunc)
		} else if linkPath != "" {
			return walkFunc(linkPath+filePath[len(dirPath):], fileInfo, err)
		} else {
			return walkFunc(filePath, fileInfo, err)
		}
	})
}

// WalkDir browses a directory on the filesystem executing a callback function for each found files.
func WalkDir(dirPath string, walkFunc filepath.WalkFunc) error {
	return walkDir(dirPath, "", walkFunc)
}
