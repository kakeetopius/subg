package cmd

import (
	"github.com/spf13/cobra"
)

func ConvertCmd() *cobra.Command {
	convertCmd := cobra.Command{
		Use:     "convert",
		Short:   "Convert subtitle from one format to another",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}

	convertCmd.Flags().SortFlags = false
	return &convertCmd
}
