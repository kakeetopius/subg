package subformat

import (
	"fmt"
	"io"
	"os"

	"github.com/asticode/go-astisub"
	"github.com/kakeetopius/subg/internal/util"
)

type Formatter struct {
	subtitle   *astisub.Subtitles
	formatType FormatType
}

func NewSubFormatter(subtitleType FormatType, r io.Reader) (Formatter, error) {
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
		return Formatter{}, fmt.Errorf("invalid subtitle format type: %v", subtitleType)
	}

	if err != nil {
		return Formatter{}, err
	}

	subFormat := Formatter{subtitle: sub, formatType: subtitleType}

	return subFormat, nil
}

func NewSubFormatterFromFile(fileName string) (Formatter, error) {
	formatType, err := SubFormatFromString(util.ExtensionOf(fileName))
	if err != nil {
		return Formatter{}, err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return Formatter{}, err
	}
	defer file.Close()
	return NewSubFormatter(formatType, file)
}

func (s *Formatter) Type() FormatType {
	return s.formatType
}

func (s *Formatter) ConvertToAndWrite(newFormatType FormatType, out io.Writer) error {
	s.formatType = newFormatType
	return s.Write(out)
}

func (s *Formatter) Write(w io.Writer) error {
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
