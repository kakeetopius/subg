// Package subformat is used to manipulate different subtitle formats.
package subformat

import (
	"cmp"
	"errors"
	"fmt"
	"strings"

	"github.com/kakeetopius/subg/internal/util"
)

type FormatType int

const (
	FormatTypeSRT FormatType = iota + 1
	FormatTypeSTL
	FormatTypeTTML
	FormatTypeSSA
	FormatTypeVTT
	FormatTypeASS
)

var (
	ErrCouldNotDetermineFormat = errors.New("could not determine format to use")
	ErrFormatMismatch          = errors.New("format mismatch between the output file extension and the format given. use one or the other")
)

func (t FormatType) String() string {
	switch t {
	case FormatTypeSRT:
		return "srt"
	case FormatTypeSTL:
		return "stl"
	case FormatTypeTTML:
		return "ttml"
	case FormatTypeSSA:
		return "ssa"
	case FormatTypeVTT:
		return "vtt"
	case FormatTypeASS:
		return "ass"
	}

	return "unknown"
}

func SubFormatFromString(s string) (FormatType, error) {
	if s == "" {
		return 0, ErrCouldNotDetermineFormat
	}
	if !strings.HasPrefix(s, ".") {
		s = fmt.Sprintf(".%v", s)
	}

	switch s {
	case ".srt":
		return FormatTypeSRT, nil
	case ".stl":
		return FormatTypeSTL, nil
	case ".ttml":
		return FormatTypeTTML, nil
	case ".ssa":
		return FormatTypeSSA, nil
	case ".vtt":
		return FormatTypeVTT, nil
	case ".ass":
		return FormatTypeASS, nil
	}

	return 0, fmt.Errorf("unsupported format: %v", s)
}

func SubFormatFromFileName(fileName string) (FormatType, error) {
	return SubFormatFromString(util.ExtensionOf(fileName))
}

func SubFormatFromFileNameOrFormatString(fileName string, format string) (FormatType, error) {
	if fileName == "" && format == "" {
		return 0, ErrCouldNotDetermineFormat
	}

	var outFileFormat FormatType
	if fileName != "" {
		outFileExt := util.ExtensionOf(fileName)
		fType, err := SubFormatFromString(outFileExt)
		if err != nil {
			return 0, err
		}
		outFileFormat = fType
	}

	var convertToFormat FormatType
	if format != "" {
		fType, err := SubFormatFromString(format)
		if err != nil {
			return 0, err
		}
		convertToFormat = fType
	}

	if fileName != "" && format != "" && convertToFormat != outFileFormat {
		return 0, ErrFormatMismatch
	}

	return cmp.Or(convertToFormat, outFileFormat), nil
}

func AddExtensionToSubFile(file string, fileFormat FormatType) string {
	return fmt.Sprintf("%s.%s", file, fileFormat)
}

func FileHasSubtitleExtension(fileName string) bool {
	_, err := SubFormatFromString(util.ExtensionOf(fileName))

	return err == nil
}
