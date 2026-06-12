package formats

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/asticode/go-astisub"
)

type SubFormat struct {
	subtitle   *astisub.Subtitles
	formatType FormatType
}

func NewSubFormat(subtitleType FormatType, r io.Reader) (SubFormat, error) {
	var sub *astisub.Subtitles
	var err error

	switch subtitleType {
	case FormatTypeSRT:
		sub, err = astisub.ReadFromSRT(r)
	case FormatTypeSTL:
		sub, err = astisub.ReadFromSTL(r, astisub.STLOptions{})
	case FormatTypeTTML:
		sub, err = astisub.ReadFromTTML(r)
	case FormatTypeSSA, FormatTypeASS:
		sub, err = astisub.ReadFromSSA(r)
	case FormatTypeVTT:
		sub, err = astisub.ReadFromWebVTT(r)
	default:
		return SubFormat{}, fmt.Errorf("invalid subtitle format type: %v", subtitleType)
	}

	if err != nil {
		return SubFormat{}, err
	}

	srtSub := SubFormat{subtitle: sub, formatType: subtitleType}

	return srtSub, nil
}

func NewSubFormatFromFile(fileName string) (SubFormat, error) {
	fileExt := filepath.Ext(fileName)
	if fileExt == "" {
		return SubFormat{}, fmt.Errorf("could not determine subtitle format of file: %s", fileName)
	}
	formatType, err := SubFormatTypeFromString(strings.ToLower(filepath.Ext(fileName)))
	if err != nil {
		return SubFormat{}, err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return SubFormat{}, err
	}
	return NewSubFormat(formatType, file)
}

func (s *SubFormat) Type() FormatType {
	return s.formatType
}

func (s *SubFormat) ConvertTo(fType FormatType, out io.Writer) error {
	s.formatType = fType
	return s.Write(out)
}

func (s *SubFormat) Write(w io.Writer) error {
	var err error

	switch s.formatType {
	case FormatTypeSRT:
		err = s.subtitle.WriteToSRT(w)
	case FormatTypeSTL:
		err = s.subtitle.WriteToSTL(w)
	case FormatTypeTTML:
		err = s.subtitle.WriteToTTML(w)
	case FormatTypeSSA:
		if s.subtitle.Metadata == nil {
			s.subtitle.Metadata = &astisub.Metadata{SSAScriptType: "v4.00"}
		}
		err = s.subtitle.WriteToSSA(w)
	case FormatTypeASS:
		if s.subtitle.Metadata == nil {
			s.subtitle.Metadata = &astisub.Metadata{SSAScriptType: "v4.00+"}
		}
		err = s.subtitle.WriteToSSA(w)
	case FormatTypeVTT:
		err = s.subtitle.WriteToWebVTT(w)
	default:
		err = fmt.Errorf("invalid subtitle format type: %v", s.formatType)
	}

	return err
}
