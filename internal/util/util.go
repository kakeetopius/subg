// Package util contains some helper functions.
package util

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CreateFileIfNotExists(fileName string) (*os.File, error) {
	if FileExists(fileName) {
		return os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC, 0o755)
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

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
	return filepath.Ext(fileName)
}

func StripExtension(fileName string) string {
	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex == -1 {
		// no dot present
		return fileName
	}
	return fileName[:dotIndex]
}

func IsVideoExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mkv", ".avi", ".mov", ".webm", ".flv", ".wmv", ".m4v":
		return true
	default:
		return false
	}
}

func IsAudioExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp3", ".m4a", ".aac", ".wav", ".flac", ".ogg", ".opus", ".wma", ".aiff", ".alac":
		return true
	default:
		return false
	}
}

func IsVideoOrAudio(ext string) bool {
	return IsAudioExtension(ext) || IsVideoExtension(ext)
}
