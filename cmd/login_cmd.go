package cmd

import (
	"fmt"

	"github.com/kakeetopius/subg/internal/config"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			config, err := config.New(configPath(cmd), loginCmdConfigFlagBinder(cmd))
			if err != nil {
				return err
			}
			config.Load()

			if providerGiven == "" {
				return fmt.Errorf("please specify provider to authenticate to. Use subg login --help for more information")
			}
			if providerGiven != "os_api" {
				return fmt.Errorf("only the provider os_api requires logging in")
			}

			switch providerGiven {
			case "os_api":
				return opensubtitles.Login(opensubtitles.LoginOptions{
					UserName: config.OpenSubtitles.UserName,
					Password: config.OpenSubtitles.Password,
					APIKey:   config.OpenSubtitles.APIKey,
					CacheDir: config.CacheDir,
				})
			}

			return nil
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVarP(&userNameGiven, "username", "u", "", "The Account username for the specified provider.")
	loginCmd.Flags().StringVarP(&passwordGiven, "password", "P", "", "The Account password for the specified provider.")

	loginCmd.Flags().StringVarP(&providerGiven, "provider", "p", "", "The provider to authenticate to.")
	return &loginCmd
}

func loginCmdConfigFlagBinder(cmd *cobra.Command) config.FlagBinder {
	return func(cfg *viper.Viper) {
		rootCmdConfigFlagBinder(cmd)(cfg)

		cfg.BindPFlag("opensubtitles.username", cmd.Flags().Lookup("username"))
		cfg.BindPFlag("opensubtitles.password", cmd.Flags().Lookup("password"))
	}
}
