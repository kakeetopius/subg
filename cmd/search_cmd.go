package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/kakeetopius/subg/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	subtitleLang   string
	season         int
	episode        int
	subtitleFormat string
	releaseYear    int
	outputFile     string
	outputDir      string
	imdbID         int

	movie      bool
	serie      bool
	autoSelect bool
)

func SearchCmd() *cobra.Command {
	searchCmd := cobra.Command{
		Use:     "search",
		Short:   "Search and download subtitles for a movie or show.",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			api := viperConfig.GetString("opensubtitles.api_key")
			// Set the outputDir to current working directory if not given
			if outputDir == "" {
				outputDir, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			// Try downloading from open subtitles first
			openSubtitleSearchOptions := opensubtitles.OpenSubSearchOptions{
				Query:         args[0],
				IMDBId:        imdbID,
				SeasonNumber:  season,
				EpisodeNumber: episode,
				Languages:     subtitleLang,
				Year:          releaseYear,

				APIKey:   api,
				CacheDir: viperConfig.GetString("cache_dir"),
			}
			err = searchAndDownloadWithOpenSubtitles(openSubtitleSearchOptions)
			if err == nil || errors.Is(err, ui.ErrUserQuit) {
				return nil
			}

			pterm.Error.Printf("Opensubtitles returned error: %v\n", err)
			pterm.Info.Println("Trying addic7ed")

			// Try addic7ed next if opensubtitles failed
			addic7edSearchOptions := addic7ed.Addic7edSearchOptions{
				Query:    args[0],
				Language: subtitleLang,
				Season:   season,
				Episode:  episode,
			}
			err = searchAndDownloadWithAddic7ed(addic7edSearchOptions)
			if err != nil {
				if errors.Is(err, ui.ErrUserQuit) {
					return nil
				}
				return err
			}
			return nil
		},
	}

	searchCmd.Flags().SortFlags = false
	searchCmd.Flags().StringVar(&subtitleLang, "lang", "en", "The Language for the subtitle to get.")
	searchCmd.Flags().IntVar(&season, "season", 0, "The serie's season if getting subtitles for a serie.")
	searchCmd.Flags().IntVar(&episode, "episode", 0, "The episode number in a serie's season.")
	searchCmd.Flags().StringVar(&subtitleFormat, "format", "srt", "The subtitle format to download.")
	searchCmd.Flags().IntVar(&releaseYear, "year", 0, "The release year of the movie or show to reduce ambiguity.")
	searchCmd.Flags().StringVar(&outputFile, "output-file", "", "The output file name for downloaded subtitle.")
	searchCmd.Flags().StringVar(&outputDir, "output-dir", "", "The output directory name for downloaded subtitle.")
	searchCmd.Flags().IntVar(&imdbID, "imdb-id", 0, "Search for show or movie using imdb ID.")
	searchCmd.Flags().BoolVar(&autoSelect, "auto", false, "Automatically select one subtitle to download without asking user.")
	searchCmd.Flags().BoolVar(&movie, "movie", false, "Specifies that the query is a movie to reduce ambiguity")
	searchCmd.Flags().BoolVar(&serie, "serie", false, "Specifies that the query is for a serie to reduce ambiguity")
	return &searchCmd
}

func searchAndDownloadWithOpenSubtitles(searchOptions opensubtitles.OpenSubSearchOptions) error {
	if searchOptions.APIKey == "" {
		return fmt.Errorf("open subtitle API Key not given. You can provide it with the --api-key flag or in the configuration file or via the environment variable OPENSUBTITLES_API_KEY ")
	}

	// the featureTypes "all", "movie", "episode" is what is required by the opensubtitles wrapper.
	featureType := "all"
	switch {
	case movie:
		featureType = "movie"
	case serie || season != 0 || episode != 0:
		// if a season or episode is given we assume it is a serie
		featureType = "episode"
	}
	searchOptions.Type = featureType

	subtitles, err := opensubtitles.SearchSubtitle(searchOptions)
	if err != nil {
		return err
	}

	if len(subtitles) < 1 {
		if episode != 0 || season != 0 {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", searchOptions.Query, season, episode)
		}
		return fmt.Errorf("no Results returned for %v", searchOptions.Query)
	}
	selectedSubtitle, err := ui.DisplayOpenSubTable(subtitles)
	if err != nil {
		if errors.Is(err, ui.ErrUserQuit) {
			return nil
		}
		return err
	}
	err = opensubtitles.DownloadSubtitle(opensubtitles.OpenSubDownloadOptions{
		Subtitle:   selectedSubtitle,
		Format:     subtitleFormat,
		OutPutFile: outputFile,
		OutPutDir:  outputDir,

		APIKey:   viperConfig.GetString("opensubtitles.api_key"),
		CacheDir: viperConfig.GetString("cache_dir"),
	})
	if err != nil {
		return err
	}

	return nil
}

func searchAndDownloadWithAddic7ed(searchOptions addic7ed.Addic7edSearchOptions) error {
	subs, err := addic7ed.SearchSubtitle(searchOptions)
	if err != nil {
		return err
	}
	selected, err := ui.DisplayAddic7edTable(&subs)
	if err != nil {
		return err
	}

	err = addic7ed.DownloadSubtitle(addic7ed.Addic7edDownloadOptions{
		Subtitle:   *selected,
		OutPutFile: subs.Name,
		OutPutDir:  outputDir,
	})
	if err != nil {
		return err
	}

	return nil
}
