package cmd

import (
	"context"
	"os"

	"github.com/kakeetopius/subg/internal/embed"
	"github.com/spf13/cobra"
)

func EmbedCmd() *cobra.Command {
	var (
		verbose          bool
		outputFile       string
		subtitleFileName string
		overwrite        bool
	)

	embedCmd := cobra.Command{
		Use:     "embed file",
		Short:   "Embed subtitle into a video file.",
		Aliases: []string{"e"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			subtitle, err := os.Open(subtitleFileName)
			if err != nil {
				return err
			}
			defer subtitle.Close()

			return embed.EmbedSubtitle(context.Background(), embed.Options{
				File:           args[0],
				Subtitle:       subtitle,
				OutputFileName: outputFile,
				Verbose:        verbose,
				Overwrite:      overwrite,
			})
		},
	}

	embedCmd.Flags().SortFlags = false
	embedCmd.Flags().StringVarP(&subtitleFileName, "subtitle", "s", "", "The subtitle to embed into the video file.")
	embedCmd.Flags().StringVarP(&outputFile, "out", "o", "", "The output file name of the new video.")
	embedCmd.Flags().BoolVar(&overwrite, "overwrite", true, "Overwrite the input file if the output file is not given.")
	embedCmd.Flags().BoolVar(&verbose, "verbose", false, "Print extra debug information from ffmpeg.")

	return &embedCmd
}
