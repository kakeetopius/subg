package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func SearchCmd() *cobra.Command {
	var (
		subtitleLang   string
		season         int
		episode        int
		subtitleFormat string
		resultLimit    int
		releaseYear    string
		outputFile     string
		imdbID         string
		autoSelect     bool
		all            bool
	)
	searchCmd := cobra.Command{
		Use:     "search",
		Short:   "Search and download subtitles for a movie or show.",
		Aliases: []string{"s"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running search")
			return nil
		},
	}

	searchCmd.Flags().SortFlags = false
	searchCmd.Flags().StringVar(&subtitleLang, "lang", "en", "The Language for the subtitle to get.")
	searchCmd.Flags().IntVar(&season, "season", 1, "The serie's season if getting subtitles for a serie.")
	searchCmd.Flags().IntVar(&episode, "episode", 1, "The episode number in a serie's season.")
	searchCmd.Flags().StringVar(&subtitleFormat, "format", "srt", "The subtitle format to download.")
	searchCmd.Flags().StringVar(&releaseYear, "year", "", "The release year of the movie or show to reduce ambiguity.")
	searchCmd.Flags().StringVar(&outputFile, "output", "", "The output file name for downloaded subtitle.")
	searchCmd.Flags().StringVar(&imdbID, "imdb-id", "", "Search for show or movie using imdb ID.")
	searchCmd.Flags().IntVar(&resultLimit, "limit", 5, "The number of subtitle results to return.")
	searchCmd.Flags().BoolVar(&all, "all", false, "Download all the subtitles returned.(The number downloaded equals to --limit)")
	searchCmd.Flags().BoolVar(&autoSelect, "auto", false, "Automatically select one subtitle to download without asking user.")
	return &searchCmd
}
