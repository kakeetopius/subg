// Package cmd is used for command line argument passing
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	debug       bool
	cfgFile     string
	subgVersion = "subg v.0.1.3"
	viperConfig *viper.Viper
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "subg",
	Short:        "A tool for searching and downloading movie, series subtitles.",
	SilenceUsage: true,
	Long: `A tool for searching and downloading movie, series subtitles.

subg is capable of downloading subtitles from various subtitle providers.

The following is the list of supported providers so far in order of priority.
  os:   opensubtitles.com
  sd:	subdl.com
  a7:   addic7ed.com
`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var (
		apiKey   string
		cacheDir string
		provider string
	)

	viperConfig = viper.New()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/subg.toml)")

	userConfigDir, err := os.UserCacheDir()
	cobra.CheckErr(err)
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "The directory to store cached information like JWT token")
	rootCmd.PersistentFlags().MarkHidden("cache-dir")
	viperConfig.BindPFlag("cache_dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	viperConfig.SetDefault("cache_dir", path.Join(userConfigDir, "subg"))

	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API Key to subtitle provider.")
	rootCmd.PersistentFlags().MarkHidden("api-key")
	apiKeyPflag := rootCmd.PersistentFlags().Lookup("api-key")
	viperConfig.BindPFlag("opensubtitles.api_key", apiKeyPflag)
	viperConfig.BindEnv("opensubtitles.api_key", "OPENSUBTITLES_API_KEY")
	viperConfig.BindPFlag("subdl.api_key", apiKeyPflag)
	viperConfig.BindEnv("subdl.api_key", "SUBDL_API_KEY")

	rootCmd.PersistentFlags().StringVar(&provider, "provider", "", "The provider to use.")
	viperConfig.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Run in debug mode.")
	rootCmd.AddCommand(
		SearchCmd(),
		LoginCmd(),
		versionCmd(),
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viperConfig.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir, err := os.UserConfigDir()
		if err != nil {
			return err
		}

		// Search config in home directory with name "subg" (without extension).
		viperConfig.AddConfigPath(home)
		viperConfig.AddConfigPath(configDir)
		viperConfig.AddConfigPath(path.Join(configDir, "subg"))
		viperConfig.SetConfigName("subg")
		viperConfig.SetConfigType("toml")
	}

	viperConfig.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viperConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// if file not found no need to error.
			return nil
		}
		return fmt.Errorf("error reading config file %v: %w", viper.ConfigFileUsed(), err)
	}

	if debug {
		fmt.Fprintln(os.Stderr, "Using config file:", viperConfig.ConfigFileUsed())
	}

	return nil
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Get the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(subgVersion)
		},
	}
}
