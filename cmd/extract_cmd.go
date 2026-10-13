package cmd

import (
	"context"

	"github.com/kakeetopius/subg/internal/extract"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/spf13/cobra"
)

func ExtractCmd() *cobra.Command {
	var (
		verbose        bool
		outputDir      string
		directory      string
		subtitleFormat string
	)

	extractCmd := cobra.Command{
		Use:     "extract file",
		Short:   "Extract subtitles from a video file.",
		Aliases: []string{"ex"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			format, err := subformat.SubFormatFromString(subtitleFormat)
			if err != nil {
				return err
			}
			return extract.ExtractSubtitles(context.Background(), extract.ExtractOptions{
				Files:     args,
				Verbose:   verbose,
				OutputDir: outputDir,
				Directory: directory,
				Format:    format,
			})
		},
	}

	extractCmd.Flags().SortFlags = false
	extractCmd.Flags().StringVarP(&directory, "dir", "d", "", "Extract subtitles for all video files in a directory.")
	extractCmd.Flags().StringVar(&outputDir, "output-dir", "", "The directory to save extracted subtitles files to.")
	extractCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The format to save the extracted subtitle files as.")
	extractCmd.Flags().BoolVar(&verbose, "verbose", false, "Print extra debug information from ffmpeg and ffprobe.")

	return &extractCmd
}
