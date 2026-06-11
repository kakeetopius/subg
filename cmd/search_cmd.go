package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/kakeetopius/subg/internal/providers/subdl"
	"github.com/kakeetopius/subg/internal/ui"
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
		Use:     "search query",
		Short:   "Search and download subtitles for a movie or show.",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				// if error is ui.ErrUserQuit no need to return it
				if err != nil {
					if errors.Is(err, ui.ErrNextProvider) || errors.Is(err, ui.ErrUserQuit) {
						err = nil
					}
				}
			}()

			providersToUse := appConfig.GetStringSlice("providers")
			query := args[0]
			// Set the outputDir to current working directory if not given
			if outputDir == "" {
				outputDir, err = os.Getwd()
				if err != nil {
					return
				}
			}

			providerSet := providers.NewProviderSet()
			providerSet.WithOptions(providers.Options{
				Query:          query,
				IMDBId:         imdbID,
				Season:         season,
				Episode:        episode,
				Language:       subtitleLang,
				Year:           releaseYear,
				SubtitleFormat: subtitleFormat,
				OutPutFile:     outputFile,
				OutPutDir:      outputDir,
				IsMovie:        isMovie,
				IsSerie:        isSerie,
				AutoSelect:     autoSelect,
			})

			for _, provider := range providersToUse {
				switch provider {
				case "os":
					providerSet.WithProvider(&opensubtitles.OpenSubtitles{}, opensubtitles.OSOptions{
						APIKey:   appConfig.GetString("opensubtitles.api_key"),
						CacheDir: appConfig.GetString("cache_dir"),
					})
				case "sd":
					providerSet.WithProvider(&subdl.SubDL{}, subdl.SubDLOpts{
						APIKey: appConfig.GetString("subdl.api_key"),
					})
				case "a7":
					providerSet.WithProvider(&addic7ed.Addic7ed{})
				default:
					return fmt.Errorf("unknown provider code: %s", provider)
				}
			}

			if len(providersToUse) == 0 {
				addAllProvidersToSet(&providerSet)
			}

			return providerSet.Start()
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

func addAllProvidersToSet(providerSet *providers.ProviderSet) {
	providerSet.WithProvider(&opensubtitles.OpenSubtitles{}, opensubtitles.OSOptions{
		APIKey:   appConfig.GetString("opensubtitles.api_key"),
		CacheDir: appConfig.GetString("cache_dir"),
	}).
		WithProvider(&subdl.SubDL{}, subdl.SubDLOpts{
			APIKey: appConfig.GetString("subdl.api_key"),
		}).
		WithProvider(&addic7ed.Addic7ed{})
}
