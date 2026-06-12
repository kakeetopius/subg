package cmd

import (
	"fmt"
	"os"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/kakeetopius/subg/internal/providers/subdl"
	"github.com/spf13/cobra"
)

func DownloadCmd() *cobra.Command {
	var (
		subtitleLang   string
		season         int
		episode        int
		subtitleFormat string
		releaseYear    int
		outputFile     string
		outputDir      string
		imdbID         int

		isMovie        bool
		isSerie        bool
		autoSelect     bool
		providersGiven []string
	)

	dlCmd := cobra.Command{
		Use:   "download query",
		Short: "Search and download subtitles for a movie or tv show.",
		Long: `Search and download subtitles for a movie or tv show.

subg is capable of downloading subtitles from various subtitle providers.
The following is the list of supported providers so far in order of priority.
  os:   opensubtitles.com
  sd:	subdl.com
  a7:   addic7ed.com
`,
		Aliases: []string{"dl", "d"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			providersToUse := appConfig.GetStringSlice("providers")

			// Set the outputDir to current working directory if not given
			if outputDir == "" {
				outputDir, err = os.Getwd()
				if err != nil {
					return
				}
			}

			providerSet := providers.NewProviderSet(args)
			providerSet.WithOptions(providers.Options{
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

			return providerSet.StartSearchAndDownload()
		},
	}

	dlCmd.Flags().SortFlags = false
	dlCmd.Flags().StringVarP(&subtitleLang, "lang", "l", "en", "The Language for the subtitle to download")
	dlCmd.Flags().IntVarP(&season, "season", "s", 0, "The season if getting subtitles for a serie.")
	dlCmd.Flags().IntVarP(&episode, "episode", "e", 0, "The episode number in a serie's season.")
	dlCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The subtitle format to download.")
	dlCmd.Flags().IntVarP(&releaseYear, "year", "y", 0, "The release year of the movie or show")
	dlCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "The output file name for downloaded subtitle.")
	dlCmd.Flags().StringVar(&outputDir, "output-dir", "", "The output directory name for downloaded subtitle.")
	dlCmd.Flags().IntVar(&imdbID, "imdb-id", 0, "Download a show or movie using imdb ID.")
	dlCmd.Flags().BoolVar(&autoSelect, "auto", false, "Automatically select one subtitle to download without asking user.")
	dlCmd.Flags().BoolVar(&isMovie, "movie", false, "Specifies that the query is for a movie to reduce ambiguity")
	dlCmd.Flags().BoolVar(&isSerie, "serie", false, "Specifies that the query is for a serie to reduce ambiguity")

	dlCmd.Flags().StringSliceVarP(&providersGiven, "providers", "p", nil, "The provider(s) to use.")
	appConfig.BindPFlag("providers", dlCmd.Flags().Lookup("providers"))
	return &dlCmd
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
