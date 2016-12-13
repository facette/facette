package osutil

import (
	"os"
	"path/filepath"
	"strings"
)

// Walk browses a directory executing a callback function for each files found.
func Walk(path string, walkFunc filepath.WalkFunc) chan error {
	errChan := make(chan error)
	go walk(path, "", walkFunc, errChan, true)
	return errChan
}

func walk(root, originalRoot string, walkFunc filepath.WalkFunc, errChan chan error, isRoot bool) {
	if _, err := os.Stat(root); err != nil {
		errChan <- err
		return
	}

	// Walk root directory
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errChan <- err
			return nil
		}

		mode := info.Mode() & os.ModeType
		if mode == os.ModeSymlink {
			// Follow symbolic link if evaluation succeeds
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				errChan <- err
				return nil
			}

			walk(realPath, path, walkFunc, errChan, false)
			return nil
		}

		if originalRoot != "" {
			path = originalRoot + strings.TrimPrefix(path, root)
		}

		walkFunc(path, info, err)
		return nil
	})
	if err != nil {
		errChan <- err
	}

	if isRoot {
		close(errChan)
	}
}
