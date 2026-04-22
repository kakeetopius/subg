package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func LoginCmd() *cobra.Command {
	var provider string
	loginCmd := cobra.Command{
		Use:     "login",
		Short:   "Authenticate to a subtitle provider",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running search")
			return nil
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVarP(&provider, "provider", "p", "", "The provider to authenticate to.")

	return &loginCmd
}
