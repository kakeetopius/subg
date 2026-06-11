package cmd

import (
	"github.com/kakeetopius/subg/internal/transcribe"
	"github.com/spf13/cobra"
)

func GenCmd() *cobra.Command {
	genCmd := cobra.Command{
		Use:     "generate",
		Short:   "Generate subtitles from video or audio file",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return transcribe.Transcribe()
		},
	}

	genCmd.Flags().SortFlags = false
	return &genCmd
}
