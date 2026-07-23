package cmd

import (
	"errors"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	opensubtitlesorg "github.com/kakeetopius/subg/internal/providers/opensubtitles_org"
	"github.com/kakeetopius/subg/internal/providers/subdl"
	"github.com/kakeetopius/subg/internal/subformat"
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

			formatType, err := subformat.SubFormatFromFileNameOrFormatString(outputFile, subtitleFormat)
			if err != nil {
				if errors.Is(err, subformat.ErrCouldNotDetermineFormat) {
					return fmt.Errorf("could not determine which format to download. use either the --format flag or give an output file with the correct extension")
				}
				return err
			}

			providerSet := providers.NewProviderSet(args)
			providerSet.WithOptions(providers.Options{
				IMDBId:         imdbID,
				Season:         season,
				Episode:        episode,
				Language:       subtitleLang,
				Year:           releaseYear,
				SubtitleFormat: formatType,
				OutputFile:     outputFile,
				OutputDir:      outputDir,
				IsMovie:        isMovie,
				IsSerie:        isSerie,
				AutoSelect:     autoSelect,
			})

			for _, provider := range providersToUse {
				switch provider {
				case "os":
					providerSet.Append(opensubtitles.NewProvider(opensubtitles.OSOptions{
						APIKey:   appConfig.GetString("opensubtitles.api_key"),
						CacheDir: appConfig.GetString("cache_dir"),
					}))
				case "sd":
					providerSet.Append(subdl.NewProvider(subdl.SubDLOpts{
						APIKey: appConfig.GetString("subdl.api_key"),
					}))
				case "os_org":
					providerSet.Append(opensubtitlesorg.NewProvider(opensubtitlesorg.OSOrgOptions{
						CacheDir: appConfig.GetString("cache_dir"),
					}))
				case "a7":
					providerSet.Append(addic7ed.NewProvider())

				default:
					return fmt.Errorf("unknown provider code: %s", provider)
				}
			}

			if len(providersToUse) == 0 {
				addAllProvidersToSet(providerSet)
			}

			return providerSet.StartSearchAndDownload()
		},
	}

	dlCmd.Flags().SortFlags = false
	dlCmd.Flags().StringVarP(&subtitleLang, "lang", "l", "en", "The Language for the subtitle to download")
	dlCmd.Flags().IntVarP(&season, "season", "s", 0, "The season if getting subtitles for a serie.")
	dlCmd.Flags().IntVarP(&episode, "episode", "e", 0, "The episode number in a serie's season.")
	dlCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The format to save the subtitle file(s) as.")
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
	providerSet.Append(
		opensubtitles.NewProvider(opensubtitles.OSOptions{
			APIKey:   appConfig.GetString("opensubtitles.api_key"),
			CacheDir: appConfig.GetString("cache_dir"),
		}),
		subdl.NewProvider(subdl.SubDLOpts{
			APIKey: appConfig.GetString("subdl.api_key"),
		}),
		opensubtitlesorg.NewProvider(opensubtitlesorg.OSOrgOptions{
			CacheDir: appConfig.GetString("cache_dir"),
		}),
		addic7ed.NewProvider(),
	)
}
