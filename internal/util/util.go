// Package util contains some helper functions.
package util

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kakeetopius/subg/internal/formats"
)

var (
	ErrCouldNotDetermineFormat = errors.New("could not determine format to use")
	ErrFormatMismatch          = errors.New("format mismatch between the output file extension and the format given. use one or the other")
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

func AddSubFileExtension(file string, fileFormat formats.FormatType) string {
	outfile := strings.TrimSuffix(file, filepath.Ext(file))

	return fmt.Sprintf("%s.%s", outfile, fileFormat)
}

func StripExtension(file string) string {
	return strings.TrimSuffix(file, filepath.Ext(file))
}

func SubFormatFromFileNameOrFormatString(fileName string, format string) (formats.FormatType, error) {
	var outFileFormat formats.FormatType
	if fileName != "" {
		outFileExt := filepath.Ext(fileName)
		fType, err := formats.SubFormatTypeFromString(outFileExt)
		if err != nil {
			return 0, err
		}
		outFileFormat = fType
	}

	var convertToFormat formats.FormatType
	if format != "" {
		fType, err := formats.SubFormatTypeFromString(format)
		if err != nil {
			return 0, err
		}
		convertToFormat = fType
	}

	if fileName != "" && format != "" && convertToFormat != outFileFormat {
		return 0, ErrFormatMismatch
	}

	var formatType formats.FormatType
	if fileName != "" {
		formatType = outFileFormat
	} else if format != "" {
		formatType = convertToFormat
	} else {
		return 0, ErrCouldNotDetermineFormat
	}

	return formatType, nil
}
