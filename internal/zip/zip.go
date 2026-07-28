// Package zip is used to work with zip files returned by subtitle providers.
package zip

import (
	"archive/zip"
	"bytes"
	"io"

	"github.com/kakeetopius/subg/internal/subformat"
)

func SubtitleFilesFromZip(zipfile io.Reader) ([]*zip.File, error) {
	responseBytes, err := io.ReadAll(zipfile)
	if err != nil {
		return nil, err
	}
	byteReader := bytes.NewReader(responseBytes)
	byteReader.Seek(0, 0)

	zipReader, err := zip.NewReader(byteReader, byteReader.Size())
	if err != nil {
		return nil, err
	}

	subtitleFiles := make([]*zip.File, 0, len(zipReader.File))
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !subformat.FileHasSubtitleExtension(f.Name) {
			continue
		}
		subtitleFiles = append(subtitleFiles, f)
		break
	}

	return subtitleFiles, nil
}
