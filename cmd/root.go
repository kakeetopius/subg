// Package cmd is used for command line argument passing
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  *viper.Viper
	apiKey  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "subg",
	Short:        "A tool for searching and downloading movie, series subtitles.",
	SilenceUsage: true,

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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/subg.toml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API Key to subtitle provider.")

	viper.BindPFlag("api_key", rootCmd.Flags().Lookup("api-key"))

	rootCmd.AddCommand(
		SearchCmd(),
		LoginCmd(),
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	config = viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		config.SetConfigFile(cfgFile)
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
		config.AddConfigPath(home)
		config.AddConfigPath(configDir)
		config.SetConfigType("toml")
		config.SetConfigName("subg")
	}

	config.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Using config file:", config.ConfigFileUsed())
	return nil
}
