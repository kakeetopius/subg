// Package formats is used to manipulate different subtitle formats.
package formats

import (
	"fmt"
)

type FormatType int

const (
	FormatTypeSRT FormatType = iota
	FormatTypeSTL
	FormatTypeTTML
	FormatTypeSSA
	FormatTypeVTT
	FormatTypeASS
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

func SubFormatTypeFromString(s string) (FormatType, error) {
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
