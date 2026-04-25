// Package addic7ed is used to interface with the addic7ed subtitle provider using a wrapper.
package addic7ed

import (
	"fmt"
	"path"

	"github.com/matcornic/addic7ed"
	"github.com/pterm/pterm"
)

type SearchOptions struct {
	Query    string
	Episode  int
	Season   int
	Language string
}

type DownloadOptions struct {
	Subtitle SubtitleOption

	OutPutFile string
	OutPutDir  string
}

type SubtitleOption struct {
	ID       int
	Language string
	Version  string
	Link     string
}

type Subtitle struct {
	Name         string
	SubtitleOpts []SubtitleOption
}

type SearchResult Subtitle

func (r SearchResult) SubtitleByID(id string) (any, error) {
	for _, sub := range r.SubtitleOpts {
		idStr := fmt.Sprint(sub.ID)
		if idStr == id {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}

func SearchSubtitle(opts SearchOptions) (SearchResult, error) {
	client := addic7ed.New()

	showName := opts.Query
	if opts.Season != 0 && opts.Episode != 0 {
		// Example format for show name is "Game of Thrones 4 x 9" - Season 4 episode 9 of GOT
		showName = fmt.Sprintf("%v - %v x %v", showName, opts.Season, opts.Episode)
	}

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on addic7ed.com..........")
	if err != nil {
		spinner.Fail()
		return SearchResult{}, err
	}

	show, err := client.SearchAll(showName)
	if err != nil {
		spinner.Fail()
		return SearchResult{}, err
	}
	if opts.Language != "" {
		show.Subtitles = show.Subtitles.Filter(addic7ed.WithLanguage(LanguageFullForm(opts.Language)))
	}

	subtitle := SearchResult{
		Name:         show.Name,
		SubtitleOpts: make([]SubtitleOption, 0, len(show.Subtitles)),
	}

	id := 1000
	for _, sub := range show.Subtitles {
		subtitle.SubtitleOpts = append(subtitle.SubtitleOpts, SubtitleOption{
			ID:       id,
			Language: sub.Language,
			Version:  sub.Version,
			Link:     sub.Link,
		})
		id++
	}

	spinner.Success("Search Done")
	return subtitle, nil
}

func DownloadSubtitle(opts DownloadOptions) error {
	subtitle := addic7ed.Subtitle{
		Language: opts.Subtitle.Language,
		Version:  opts.Subtitle.Version,
		Link:     opts.Subtitle.Link,
	}

	fileName := fmt.Sprintf("%v.%v", opts.OutPutFile, "srt")
	outPath := path.Join(opts.OutPutDir, fileName)
	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle.........")
	if err != nil {
		spinner.Fail()
		return err
	}

	err = subtitle.DownloadTo(outPath)
	if err != nil {
		return err
	}

	spinner.Success("Download Done.")
	pterm.Info.Println("Subtitle saved at: ", outPath)
	return nil
}

func LanguageFullForm(s string) string {
	// TODO: Obviously add more.
	langs := map[string]string{
		"en": "English",
		"fr": "French",
	}

	return langs[s]
}
