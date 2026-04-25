package cmd

import (
	"fmt"

	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/spf13/cobra"
)

func LoginCmd() *cobra.Command {
	var (
		userName string
		password string
	)

	loginCmd := cobra.Command{
		Use:     "login",
		Short:   "Authenticate to a subtitle provider",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			providerToUse := viperConfig.GetString("provider")
			if providerToUse == "" {
				return fmt.Errorf("please specify provider to authenticate to. Use subg login --help for more information")
			}
			switch providerToUse {
			case "os":
				return opensubtitles.Login(opensubtitles.LoginOptions{
					UserName: viperConfig.GetString("opensubtitles.username"),
					Password: viperConfig.GetString("opensubtitles.password"),
					APIKey:   viperConfig.GetString("opensubtitles.api_key"),
					CacheDir: viperConfig.GetString("cache_dir"),
				})
			case "sd":
				fmt.Println("Provider subdl.com doesn't need any authentication. The provider only requires an api key that can be passed via the --api-key flag or via the SUBDL_API_KEY or in the configuration file.")
				return nil
			case "a7":
				fmt.Println("Provider: addic7ed.com doesn't need any authentication.")
				return nil
			default:
				return fmt.Errorf("unknown provider: %v", providerToUse)
			}
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVarP(&userName, "username", "u", "", "The Account username for the specified provider.")
	loginCmd.Flags().StringVarP(&password, "password", "P", "", "The Account password for the specified provider.")

	userNamePflag := loginCmd.Flags().Lookup("username")
	passwordPflag := loginCmd.Flags().Lookup("password")

	viperConfig.BindPFlag("opensubtitles.username", userNamePflag)
	viperConfig.BindPFlag("opensubtitles.password", passwordPflag)

	return &loginCmd
}
