// Package util contains some helper functions.
package util

import (
	"os"
	"path"
	"strings"
)

func CreateFileIfNotExists(fileName string) (*os.File, error) {
	if FileExists(fileName) {
		return os.Open(fileName)
	}

	err := CreateDirIfNotExists(path.Dir(fileName))
	if err != nil {
		return nil, err
	}

	return os.Create(fileName)
}

func CreateDirIfNotExists(dirName string) error {
	if DirExists(dirName) {
		return nil
	}
	return os.MkdirAll(dirName, 0o755)
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func ExtensionOf(fileName string) string {
	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex == -1 {
		// no dot present
		return ""
	}
	return fileName[dotIndex:]
}

func StripExtension(fileName string) string {
	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex == -1 {
		// no dot present
		return fileName
	}
	return fileName[:dotIndex]
}
