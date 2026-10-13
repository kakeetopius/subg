package cmd

import (
	"errors"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	opensubtitlesorg "github.com/kakeetopius/subg/internal/providers/opensubtitles_org"
	"github.com/kakeetopius/subg/internal/providers/subdl"
	subdlapi "github.com/kakeetopius/subg/internal/providers/subdl_api"
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
		directory      string
		fileName       string
		embed          bool
		keep           bool

		isMovie        bool
		isSerie        bool
		noAutoSelect   bool
		providersGiven []string
	)

	dlCmd := cobra.Command{
		Use:   "download query",
		Short: "Search and download subtitles for a movie or tv show.",
		Long: `Search and download subtitles for a movie or tv show.

subg is capable of downloading subtitles from various subtitle providers.
The following is the list of supported providers so far in order of priority.
  os:   opensubtitles.org
  sd:	subdl.com
  a7:   addic7ed.com
`,
		Aliases: []string{"dl", "d"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			providersToUse := appConfig.GetStringSlice("providers")

			formatType, err := subformat.SubFormatFromFileNameOrFormatString(outputFile, subtitleFormat)
			if err != nil {
				if errors.Is(err, subformat.ErrCouldNotDetermineFormat) {
					return fmt.Errorf("could not determine which format to download. use either the --format flag or give an output file with the correct extension")
				}
				return err
			}

			providerSet := providers.NewProviderSet(providers.ProviderSetOptions{
				Directory:             directory,
				DefaultEmbedSubtitles: embed,
				AutoSelect:            !noAutoSelect,
				DefaultSubFormat:      formatType,
			})

			searchOpts := providers.SearchOptions{
				IMDBId:         imdbID,
				Season:         season,
				Episode:        episode,
				Language:       subtitleLang,
				Year:           releaseYear,
				SubtitleFormat: formatType,
				OutputFile:     outputFile,
				OutputDir:      outputDir,
				Embed:          embed,
				KeepSub:        keep,
				IsMovie:        isMovie,
				IsSerie:        isSerie,
			}
			if len(args) > 0 {
				searchOpts.Query = args[0]
			}
			if fileName != "" {
				searchOpts.InputFileName = fileName
			}

			providerSet.AppendQuery(searchOpts)

			cacheDir := appConfig.GetString("cache_dir")
			for _, provider := range providersToUse {
				switch provider {
				case "os":
					providerSet.Append(opensubtitlesorg.NewProvider(opensubtitlesorg.OSOrgOptions{
						CacheDir: cacheDir,
					}))
				case "sd":
					providerSet.Append(subdl.NewProvider())
				case "a7":
					providerSet.Append(addic7ed.NewProvider())
				case "os_api":
					apiKey := appConfig.GetString("opensubtitles.api_key")
					if apiKey == "" {
						return fmt.Errorf("cannot use os_api provider without an api key from opensubtitles.com")
					}
					providerSet.Append(opensubtitles.NewProvider(opensubtitles.OSOptions{
						APIKey:   apiKey,
						CacheDir: cacheDir,
					}))
				case "sd_api":
					apiKey := appConfig.GetString("subdl.api_key")
					if apiKey == "" {
						return fmt.Errorf("cannot use sd_api provider without an api key subdl.com")
					}
					providerSet.Append(subdlapi.NewProvider(subdlapi.SubDLOpts{
						APIKey: apiKey,
					}))
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
	dlCmd.Flags().IntVarP(&releaseYear, "year", "y", 0, "The release year of the movie or show")
	dlCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The format to save the subtitle file as.")
	dlCmd.Flags().IntVar(&imdbID, "imdb-id", 0, "Download a show or movie using imdb ID.")
	dlCmd.Flags().BoolVar(&isMovie, "movie", false, "Specifies that the query is for a movie to reduce ambiguity")
	dlCmd.Flags().BoolVar(&isSerie, "serie", false, "Specifies that the query is for a serie to reduce ambiguity")

	dlCmd.Flags().StringVarP(&fileName, "file", "F", "", "Download subtitles for a movie / tv show video file")
	dlCmd.Flags().StringVarP(&directory, "dir", "d", "", "Download subtitles for all movies / tv shows in a a directory.")
	dlCmd.Flags().BoolVar(&embed, "embed", false, "Embed the downloaded subtitle(s) into the video file(s) given via --file or --dir")
	dlCmd.Flags().BoolVar(&keep, "keep", false, "Keep subtitle file even after embedding.")

	dlCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "The output file name for downloaded subtitle.")
	dlCmd.Flags().StringVar(&outputDir, "output-dir", "", "The output directory name for downloaded subtitle.")
	dlCmd.Flags().BoolVar(&noAutoSelect, "no-auto", false, "Do not automatically select subtitle to download rather show results to allow the user to select.")

	dlCmd.Flags().StringSliceVarP(&providersGiven, "providers", "p", nil, "The provider(s) to use.")
	appConfig.BindPFlag("providers", dlCmd.Flags().Lookup("providers"))
	return &dlCmd
}

func addAllProvidersToSet(providerSet *providers.ProviderSet) {
	cacheDir := appConfig.GetString("cache_dir")

	providerSet.Append(
		opensubtitlesorg.NewProvider(opensubtitlesorg.OSOrgOptions{
			CacheDir: cacheDir,
		}),
		subdl.NewProvider(),
		addic7ed.NewProvider(),
	)

	openSubAPIKey := appConfig.GetString("opensubtitles.api_key")
	if openSubAPIKey != "" {
		providerSet.Append(
			opensubtitles.NewProvider(opensubtitles.OSOptions{
				APIKey:   appConfig.GetString("opensubtitles.api_key"),
				CacheDir: cacheDir,
			}),
		)
	}

	subDLAPIKey := appConfig.GetString("subdl.api_key")
	if subDLAPIKey != "" {
		providerSet.Append(
			subdlapi.NewProvider(subdlapi.SubDLOpts{
				APIKey: appConfig.GetString("subdl.api_key"),
			}),
		)
	}
}
