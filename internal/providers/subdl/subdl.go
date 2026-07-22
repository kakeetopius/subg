package subdl

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/kakeetopius/subg/internal/zip"
	"github.com/pterm/pterm"
)

func searchSubtitles(opts options) ([]SDSubtitle, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf(" An API Key is required to use subdl.com")
	}
	c, err := NewClient(Config{
		APIKey: opts.APIKey,
	})
	if err != nil {
		return nil, err
	}

	searchParams := SearchParams{}
	searchParams.Query = &opts.Query
	searchParams.APIKey = opts.APIKey

	if opts.Season != 0 {
		searchParams.SeasonNumber = &opts.Season
	}
	if opts.Episode != 0 {
		searchParams.EpisodeNumber = &opts.Episode
	}
	if opts.IMDBId != 0 {
		searchParams.IMDBId = &opts.IMDBId
	}
	if opts.Year != 0 {
		searchParams.Year = &opts.Year
	}
	if opts.Language != "" {
		searchParams.Languages = &opts.Language
	}

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on subdl.com.........")
	if err != nil {
		return nil, err
	}
	results, err := c.SearchSubtitles(context.Background(), searchParams)
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
	return results.Subtitles, nil
}

func downloadSubtitle(subtitle *SDSubtitle) (name string, subBytes io.ReadCloser, format subformat.FormatType, err error) {
	if subtitle == nil {
		err = fmt.Errorf("no subtitle provided for download")
		return
	}
	url := SUBDLDOWNLOADURL + subtitle.URL

	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle.........")
	if err != nil {
		return
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
		return
	}
	defer resp.Body.Close()

	subFiles, err := zip.SubtitleFilesFromZip(resp.Body)
	if err != nil {
		return "", nil, 0, err
	}
	if len(subFiles) == 0 {
		err = fmt.Errorf("no subtitle files found in the zip file")
		return
	}

	subFile := subFiles[0]
	format, err = subformat.SubFormatFromFileName(subFile.Name)
	if err != nil {
		return "", nil, 0, err
	}
	subBytes, err = subFile.Open()
	if err != nil {
		return "", nil, 0, err
	}

	return util.StripExtension(subFile.Name), subBytes, format, nil
}
