package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kakeetopius/subg/cmd/search"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/spf13/cobra"
)

func SearchCmd() *cobra.Command {
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
	searchCmd := cobra.Command{
		Use:     "search",
		Short:   "Search and download subtitles for a movie or show.",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api := config.GetString("opensubtitles.api_key")
			if api == "" {
				return fmt.Errorf("open subtitle API Key not given. You can provide it with the --api-key flag or in the configuration file or via the environment variable OPENSUBTITLES_API_KEY ")
			}
			searchOptions := opensubtitles.SearchOptions{
				Query:         args[0],
				IMDBId:        imdbID,
				SeasonNumber:  season,
				EpisodeNumber: episode,
				Languages:     subtitleLang,
				Year:          releaseYear,

				APIKey:   api,
				CacheDir: config.GetString("cache_dir"),
			}
			featureType := "all"
			switch {
			case movie:
				featureType = "movie"
			case serie:
				featureType = "episode"
			}
			searchOptions.Type = featureType

			subtitles, err := opensubtitles.SearchSubtitle(searchOptions)
			if len(subtitles) < 1 {
				if episode != 0 || season != 0 {
					return fmt.Errorf("no results returned for %v Season %v Episode %v", args[0], season, episode)
				}
				return fmt.Errorf("no Results returned for %v", args[0])
			}
			if err != nil {
				return err
			}
			selectedSubtitle, err := search.DisplaySubtitleTable(subtitles)
			if err != nil {
				if errors.Is(err, search.ErrUserQuit) {
					return nil
				}
				return err
			}

			if outputDir == "" {
				outputDir, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			err = opensubtitles.DownloadSubtitle(opensubtitles.DownloadOptions{
				Subtitle:   selectedSubtitle,
				Format:     subtitleFormat,
				OutPutFile: outputFile,
				OutPutDir:  outputDir,

				APIKey:   config.GetString("opensubtitles.api_key"),
				CacheDir: config.GetString("cache_dir"),
			})
			if err != nil {
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
