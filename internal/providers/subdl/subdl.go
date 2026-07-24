package subdl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/kakeetopius/subg/internal/zip"
	"github.com/pterm/pterm"
)

func (p *SubDL) searchSubtitles(searchOptions providers.SearchOptions) ([]providers.Subtitle, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf(" An API Key is required to use subdl.com")
	}
	c, err := NewClient(Config{
		APIKey: p.APIKey,
	})
	if err != nil {
		return nil, err
	}

	searchParams := SearchParams{}
	searchParams.Query = &searchOptions.Query
	searchParams.APIKey = p.APIKey

	if searchOptions.Season != 0 {
		searchParams.SeasonNumber = &searchOptions.Season
	}
	if searchOptions.Episode != 0 {
		searchParams.EpisodeNumber = &searchOptions.Episode
	}
	if searchOptions.IMDBId != 0 {
		searchParams.IMDBId = &searchOptions.IMDBId
	}
	if searchOptions.Year != 0 {
		searchParams.Year = &searchOptions.Year
	}
	if searchOptions.Language != "" {
		searchParams.Languages = &searchOptions.Language
	}

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on subdl.com.........")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			spinner.Success("Search Complete")
		} else {
			spinner.Fail()
		}
	}()

	results, err := c.SearchSubtitles(context.Background(), searchParams)
	if err != nil {
		return nil, err
	}

	id := 1000
	subs := make([]providers.Subtitle, 0, len(results.Subtitles))
	for _, sub := range results.Subtitles {
		sub.SubID = id
		subs = append(subs, sub)
		id++
	}

	return subs, nil
}

func downloadSubtitle(ctx context.Context, subtitle *SDSubtitle) (subtitleFile providers.SubtitleFile, err error) {
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
		return
	}
	if len(subFiles) == 0 {
		err = fmt.Errorf("no subtitle files found in the zip file")
		return
	}

	subFile := subFiles[0]
	format, err := subformat.SubFormatFromFileName(subFile.Name)
	if err != nil {
		return
	}
	subBytes, err := subFile.Open()
	if err != nil {
		return
	}

	return providers.SubtitleFile{
		Name:       util.StripExtension(subFile.Name),
		Type:       format,
		ReadCloser: subBytes,
	}, nil
}
