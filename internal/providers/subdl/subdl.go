package subdl

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pterm/pterm"
)

type SubDLDownloadOptions struct {
	Subtitle   *SubDLSubtitle
	OutPutFile string
	OutPutDir  string
}

func SearchSubtitles(opts SubDLSearchParams) (*SubDLSearchResults, error) {
	c, err := NewClient(Config{
		APIKey: opts.APIKey,
	})
	if err != nil {
		return nil, err
	}

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on subdl.com.........")
	if err != nil {
		return nil, err
	}
	results, err := c.SearchSubtitles(context.Background(), opts)
	if err != nil {
		spinner.Fail()
		return nil, err
	}
	id := 1000
	for i := range results.Subtitles {
		results.Subtitles[i].ID = id
		id++
	}

	spinner.Success("Search Done")
	return results, nil
}

func DownloadSubtitle(opts SubDLDownloadOptions) (err error) {
	if opts.Subtitle == nil {
		return fmt.Errorf("no subtitle provided for download")
	}
	url := SUBDLDOWNLOADURL + opts.Subtitle.URL

	zipOutfile := opts.OutPutFile
	if zipOutfile == "" {
		zipOutfile = fmt.Sprintf("%v.%v", opts.Subtitle.ReleaseName, "zip")
	}

	zipOutfile = path.Join(opts.OutPutDir, zipOutfile)
	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle.........")
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			spinner.Success("Download Done")
		} else {
			spinner.Fail()
		}
	}()

	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outFile, err := os.OpenFile(zipOutfile, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipBytesReader := bytes.NewReader(zipBytes)
	_, err = io.Copy(outFile, zipBytesReader)
	if err != nil {
		spinner.Fail()
		return err
	}

	pterm.Info.Printf("Zip file downloaded successfully to: %v \n", zipOutfile)
	pterm.Info.Printf("Extracting zip file...\n")
	return extractSubtitlesFromZip(zipBytesReader, opts.OutPutDir)
}

func extractSubtitlesFromZip(zipBytes *bytes.Reader, outDir string) error {
	zipBytes.Seek(0, 0) // reset to start
	zipper, err := zip.NewReader(zipBytes, zipBytes.Size())
	if err != nil {
		return err
	}

	var allFiles []*zip.File
	for _, f := range zipper.File {
		// get files only
		if !f.FileInfo().IsDir() {
			allFiles = append(allFiles, f)
		}
	}

	var srtFiles []*zip.File
	for _, f := range allFiles {
		// find all .srt files
		if strings.HasSuffix(f.Name, ".srt") {
			srtFiles = append(srtFiles, f)
		}
	}

	if len(srtFiles) == 0 {
		// if no srt files found consider any file
		srtFiles = allFiles
	}

	for _, f := range srtFiles {
		subtitleFile, err := f.Open()
		if err != nil {
			return err
		}
		defer subtitleFile.Close()

		outFileName := path.Join(outDir, f.Name)
		outFile, err := os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, subtitleFile)
		if err != nil {
			return err
		}
	}

	pterm.Info.Printf("Extracted %v files to: %v\n", len(srtFiles), outDir)
	return nil
}
