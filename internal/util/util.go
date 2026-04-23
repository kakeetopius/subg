// Package util contains some helper functions.
package util

import (
	"errors"
	"os"
	"path"
)

func CreateFileIfNotExists(fileName string) (*os.File, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0o644)
	if err == nil {
		return file, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	parentDirs := path.Dir(fileName)
	err = os.MkdirAll(parentDirs, 0o755)
	if err != nil {
		return nil, err
	}

	return os.Create(fileName)
}
