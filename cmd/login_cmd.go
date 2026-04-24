package cmd

import (
	"fmt"

	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/spf13/cobra"
)

func LoginCmd() *cobra.Command {
	var (
		provider string
		userName string
		password string
	)

	loginCmd := cobra.Command{
		Use:     "login",
		Short:   "Authenticate to a subtitle provider",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if provider == "" {
				return fmt.Errorf("please specify provider to authenticate to. Use subg login --help for more information")
			}
			switch provider {
			case "os":
				return opensubtitles.Login(opensubtitles.OpenSubLoginOptions{
					UserName: config.GetString("opensubtitles.username"),
					Password: config.GetString("opensubtitles.password"),
					APIKey:   config.GetString("opensubtitles.api_key"),
					CacheDir: config.GetString("cache_dir"),
				})

			default:
				return fmt.Errorf("unsupported provider: %v", provider)
			}
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVarP(&provider, "provider", "p", "", "The provider to authenticate to.")
	loginCmd.Flags().StringVarP(&userName, "username", "u", "", "The Account username for the specified provider.")
	loginCmd.Flags().StringVarP(&password, "password", "P", "", "The Account password for the specified provider.")

	userNamePflag := loginCmd.Flags().Lookup("username")
	passwordPflag := loginCmd.Flags().Lookup("password")

	config.BindPFlag("opensubtitles.username", userNamePflag)
	config.BindPFlag("opensubtitles.password", passwordPflag)

	return &loginCmd
}
