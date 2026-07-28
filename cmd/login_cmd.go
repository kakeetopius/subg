package cmd

import (
	"fmt"

	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/spf13/cobra"
)

func LoginCmd() *cobra.Command {
	var (
		userNameGiven string
		passwordGiven string
		providerGiven string
	)

	loginCmd := cobra.Command{
		Use:   "login",
		Short: "Authenticate to a subtitle provider",
		Long: `Authenticate to a subtitle provider.
Note that only opensubtitles.com (code: os_api) requires authentication`,
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if providerGiven == "" {
				return fmt.Errorf("please specify provider to authenticate to. Use subg login --help for more information")
			}
			if providerGiven != "os_api" {
				return fmt.Errorf("only the provider os_api requires logging in")
			}

			switch providerGiven {
			case "os_api":
				return opensubtitles.Login(opensubtitles.LoginOptions{
					UserName: appConfig.GetString("opensubtitles.username"),
					Password: appConfig.GetString("opensubtitles.password"),
					APIKey:   appConfig.GetString("opensubtitles.api_key"),
					CacheDir: appConfig.GetString("cache_dir"),
				})
			}

			return nil
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVarP(&userNameGiven, "username", "u", "", "The Account username for the specified provider.")
	loginCmd.Flags().StringVarP(&passwordGiven, "password", "P", "", "The Account password for the specified provider.")

	userNamePflag := loginCmd.Flags().Lookup("username")
	passwordPflag := loginCmd.Flags().Lookup("password")

	appConfig.BindPFlag("opensubtitles.username", userNamePflag)
	appConfig.BindPFlag("opensubtitles.password", passwordPflag)

	loginCmd.Flags().StringVarP(&providerGiven, "provider", "p", "", "The provider to authenticate to.")
	return &loginCmd
}
