package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/kakeetopius/subg/internal/providers/subdl"
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

	isMovie    bool
	isSerie    bool
	autoSelect bool
)

func SearchCmd() *cobra.Command {
	searchCmd := cobra.Command{
		Use:     "search",
		Short:   "Search and download subtitles for a movie or show.",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				// if error is ui.ErrUserQuit no need to return it
				if err != nil && errors.Is(err, ui.ErrUserQuit) {
					err = nil
				}
			}()

			providerToUse := viperConfig.GetString("provider")
			query := args[0]
			// Set the outputDir to current working directory if not given
			if outputDir == "" {
				outputDir, err = os.Getwd()
				if err != nil {
					return
				}
			}

			switch providerToUse {
			case "os":
				return searchAndDownloadWithOpenSubtitles(query)
			case "sd":
				return searchAndDownloadWithSubdl(query)
			case "a7":
				return searchAndDownloadWithAddic7ed(query)
			default:
				if providerToUse != "" {
					return fmt.Errorf("unknown subtitle provider: %v", providerToUse)
				}
			}

			// If no provider was given try downloading from open subtitles first
			err = searchAndDownloadWithOpenSubtitles(query)
			if err == nil || errors.Is(err, ui.ErrUserQuit) {
				return
			}

			pterm.Error.Printf("Opensubtitles returned error: %v\n", err)
			pterm.Info.Println("Trying subdl.com")

			// Try subdl next
			err = searchAndDownloadWithSubdl(query)
			if err == nil || errors.Is(err, ui.ErrUserQuit) {
				return
			}

			pterm.Error.Printf("subdl.com returned error: %v\n", err)
			pterm.Info.Println("Trying addic7ed.com")

			// Try addic7ed next
			err = searchAndDownloadWithAddic7ed(query)
			if err != nil {
				pterm.Error.Println(err)
				// We dont return the error because it is already printed
			}
			return nil
		},
	}

	searchCmd.Flags().SortFlags = false
	searchCmd.Flags().StringVarP(&subtitleLang, "lang", "l", "en", "The Language for the subtitle to get.")
	searchCmd.Flags().IntVarP(&season, "season", "s", 0, "The serie's season if getting subtitles for a serie.")
	searchCmd.Flags().IntVarP(&episode, "episode", "e", 0, "The episode number in a serie's season.")
	searchCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The subtitle format to download.")
	searchCmd.Flags().IntVarP(&releaseYear, "year", "y", 0, "The release year of the movie or show to reduce ambiguity.")
	searchCmd.Flags().StringVar(&outputFile, "output-file", "", "The output file name for downloaded subtitle.")
	searchCmd.Flags().StringVar(&outputDir, "output-dir", "", "The output directory name for downloaded subtitle.")
	searchCmd.Flags().IntVar(&imdbID, "imdb-id", 0, "Search for show or movie using imdb ID.")
	searchCmd.Flags().BoolVar(&autoSelect, "auto", false, "Automatically select one subtitle to download without asking user.")
	searchCmd.Flags().BoolVar(&isMovie, "movie", false, "Specifies that the query is a movie to reduce ambiguity")
	searchCmd.Flags().BoolVar(&isSerie, "serie", false, "Specifies that the query is for a serie to reduce ambiguity")
	return &searchCmd
}

func searchAndDownloadWithOpenSubtitles(query string) error {
	api := viperConfig.GetString("opensubtitles.api_key")

	if api == "" {
		return fmt.Errorf("open subtitle API Key not given. You can provide it with the --api-key flag or in the configuration file or via the environment variable OPENSUBTITLES_API_KEY ")
	}
	searchOptions := opensubtitles.SearchOptions{
		Query:         query,
		IMDBId:        imdbID,
		SeasonNumber:  season,
		EpisodeNumber: episode,
		Languages:     subtitleLang,
		Year:          releaseYear,

		APIKey:   api,
		CacheDir: viperConfig.GetString("cache_dir"),
	}

	// the featureTypes "all", "movie", "episode" is what is required by the opensubtitles wrapper.
	featureType := "all"
	switch {
	case isMovie:
		featureType = "movie"
	case isSerie || season != 0 || episode != 0:
		// if a season or episode is given we assume it is a serie
		featureType = "episode"
	}
	searchOptions.Type = featureType

	subtitles, err := opensubtitles.SearchSubtitle(searchOptions)
	if err != nil {
		return err
	}

	if len(subtitles) < 1 {
		if isSerie || episode != 0 || season != 0 {
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
	err = selectedSubtitle.Download(opensubtitles.DownloadOptions{
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

func searchAndDownloadWithAddic7ed(query string) error {
	searchOptions := addic7ed.SearchOptions{
		Query:    query,
		Language: subtitleLang,
		Season:   season,
		Episode:  episode,
	}
	subs, err := addic7ed.SearchSubtitle(searchOptions)
	if err != nil {
		return err
	}
	selected, err := ui.DisplayAddic7edTable(&subs)
	if err != nil {
		return err
	}

	err = selected.Download(addic7ed.DownloadOptions{
		OutPutFile: subs.Name,
		OutPutDir:  outputDir,
	})
	if err != nil {
		return err
	}

	return nil
}

func searchAndDownloadWithSubdl(query string) error {
	searchOptions := subdl.SearchParams{
		APIKey: viperConfig.GetString("subdl.api_key"),
		Query:  &query,
	}

	// keyword "movie" or "tv" is what is required by the subdl API.
	featureType := "movie"
	if isSerie || season != 0 || episode != 0 {
		// if a season or episode is given we assume it is a serie
		featureType = "tv"
	}
	searchOptions.Type = &featureType

	if season != 0 {
		searchOptions.SeasonNumber = &season
	}
	if episode != 0 {
		searchOptions.EpisodeNumber = &episode
	}
	if imdbID != 0 {
		searchOptions.IMDBId = &imdbID
	}
	if releaseYear != 0 {
		searchOptions.Year = &releaseYear
	}
	if subtitleLang != "" {
		searchOptions.Languages = &subtitleLang
	}

	results, err := subdl.SearchSubtitles(searchOptions)
	if err != nil {
		return err
	}
	if !results.Status || len(results.Results) == 0 || len(results.Subtitles) == 0 {
		if isSerie || episode != 0 || season != 0 {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", query, season, episode)
		}
		return fmt.Errorf("no Results returned for %v", query)
	}

	selectedSub, err := ui.DisplaySubDLTable(results)
	if err != nil {
		return err
	}
	err = selectedSub.Download(subdl.DownloadOptions{
		OutPutDir:  outputDir,
		OutPutFile: outputFile,
	})
	if err != nil {
		return err
	}
	return nil
}
