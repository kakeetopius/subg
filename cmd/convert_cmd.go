package cmd

import (
	"github.com/spf13/cobra"
)

func ConvertCmd() *cobra.Command {
	convertCmd := cobra.Command{
		Use:     "download query",
		Short:   "Search and download subtitles for a movie or tv show.",
		Aliases: []string{"dl", "d"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}

	convertCmd.Flags().SortFlags = false
	return &convertCmd
}
