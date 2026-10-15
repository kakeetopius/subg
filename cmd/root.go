// Package cmd is used for command line argument passing
package cmd

import (
	"fmt"
	"os"
	"path"

	goversion "github.com/caarlos0/go-version"
	"github.com/kakeetopius/subg/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Command() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:          "subg",
		Short:        "A tool for downloading, generating, and manipulating subtitles.",
		SilenceUsage: true,
		Version:      buildVersion().GitVersion,
	}

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/subg.toml)")
	rootCmd.PersistentFlags().String("cache-dir", "", "The directory to store cached information like JWT token")
	rootCmd.PersistentFlags().Bool("debug", false, "Run in debug mode.")

	rootCmd.AddCommand(
		DownloadCmd(),
		ConvertCmd(),
		GenCmd(),
		EmbedCmd(),
		ExtractCmd(),
		LoginCmd(),
		versionCmd(),
	)

	return rootCmd
}

func rootCmdConfigFlagBinder(cmd *cobra.Command) config.FlagBinder {
	return func(cfg *viper.Viper) {
		userCacheDir, err := os.UserCacheDir()
		cobra.CheckErr(err)

		cfg.BindPFlag("cache_dir", cmd.PersistentFlags().Lookup("cache-dir"))
		cfg.SetDefault("cache_dir", path.Join(userCacheDir, "subg"))
	}
}

func configPath(cmd *cobra.Command) string {
	config, _ := cmd.PersistentFlags().GetString("config")
	return config
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show detailed version information",
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(buildVersion().String())
		},
	}
}

// buildVersion constructs and returns the application's version information.
func buildVersion() goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("subg", "Tool to work with subtitles.", ""),
	)
}
