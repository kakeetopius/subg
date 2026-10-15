// Package config loads and validates the configuration
package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	Providers     []string
	CacheDir      string
	Transcriber   TranscriberConfig
	OpenSubtitles OpenSubtitlesConfig
	SubDL         SubDLConfig

	cfg        *viper.Viper
	flagBinder FlagBinder
}

type TranscriberConfig struct {
	HFToken string
}

type OpenSubtitlesConfig struct {
	UserName string
	Password string
	APIKey   string
}

type SubDLConfig struct {
	APIKey string
}

// FlagBinder is a function that binds flags to a viper config.
type FlagBinder func(*viper.Viper)

func New(path string, flagBinder FlagBinder) (*Config, error) {
	appConfig, err := initConfig(path)
	if err != nil {
		return nil, err
	}

	return new(Config{
		cfg:        appConfig,
		flagBinder: flagBinder,
	}), nil
}

func (c *Config) SetViperConfig(v *viper.Viper) {
	c.cfg = v
}

func (c *Config) Load() error {
	if c.flagBinder != nil {
		c.flagBinder(c.cfg)
	}

	c.Providers = c.cfg.GetStringSlice("providers")
	c.CacheDir = c.cfg.GetString("cache_dir")

	c.Transcriber.HFToken = c.cfg.GetString("transcriber.hf_token")

	c.OpenSubtitles.UserName = c.cfg.GetString("opensubtitles.username")
	c.OpenSubtitles.Password = c.cfg.GetString("opensubtitles.password")
	c.OpenSubtitles.APIKey = c.cfg.GetString("opensubtitles.api_key")

	c.SubDL.APIKey = c.cfg.GetString("subdl.api_key")

	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string) (*viper.Viper, error) {
	appConfig := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		appConfig.SetConfigFile(cfgFile)
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}

		// Search config in home directory with name "subg" (without extension).
		appConfig.AddConfigPath(configDir)
		appConfig.AddConfigPath(path.Join(configDir, "subg"))
		appConfig.SetConfigName("subg")
	}

	appConfig.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := appConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// if file not found no need to error.
			return appConfig, nil
		}
		return nil, fmt.Errorf("error reading config file %v: %w", viper.ConfigFileUsed(), err)
	}

	return appConfig, nil
}
