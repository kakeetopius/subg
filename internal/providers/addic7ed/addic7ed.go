// Package addic7ed is used to interface with the addic7ed subtitle provider using a wrapper.
package addic7ed

import (
	"fmt"
	"path"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/matcornic/addic7ed"
	"github.com/pterm/pterm"
)

type A7Subtitle struct {
	ID       int
	Language string
	Version  string
	Link     string
}

func searchSubtitle(opts providers.Options) ([]A7Subtitle, error) {
	client := addic7ed.New()

	showName := opts.Query
	if opts.Season != 0 && opts.Episode != 0 {
		// Example format for show name is "Game of Thrones 4 x 9" - Season 4 episode 9 of GOT
		showName = fmt.Sprintf("%v - %v x %v", showName, opts.Season, opts.Episode)
	}

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on addic7ed.com..........")
	if err != nil {
		spinner.Fail()
		return nil, err
	}

	show, err := client.SearchAll(showName)
	if err != nil {
		spinner.Fail()
		return nil, err
	}
	if opts.Language != "" {
		show.Subtitles = show.Subtitles.Filter(addic7ed.WithLanguage(LanguageFullForm(opts.Language)))
	}

	subtitles := make([]A7Subtitle, 0, len(show.Subtitles))

	id := 1000
	for _, sub := range show.Subtitles {
		subtitles = append(subtitles, A7Subtitle{
			ID:       id,
			Language: sub.Language,
			Version:  sub.Version,
			Link:     sub.Link,
		})
		id++
	}

	spinner.Success("Search Done")
	return subtitles, nil
}

func downloadSubtitle(sub *A7Subtitle, opts providers.Options) error {
	subtitle := addic7ed.Subtitle{
		Language: opts.Language,
		Version:  sub.Version,
		Link:     sub.Link,
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
