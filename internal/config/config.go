package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the main configuration structure
type Config struct {
	APIURL string       `mapstructure:"api_url"`
	Daemon DaemonConfig `mapstructure:"daemon"`
	Output OutputConfig `mapstructure:"output"`
}

// DaemonConfig contains daemon-specific settings
type DaemonConfig struct {
	Interval string `mapstructure:"interval"`
	Projects []int  `mapstructure:"projects"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	Format string `mapstructure:"format"`
	Color  bool   `mapstructure:"color"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	var paths []string
	if err == nil {
		// Search local config first, then global configs
		// This allows local development config to override global config
		paths = []string{
			".",
			filepath.Join(homeDir, ".config", "todu"),
			filepath.Join(homeDir, ".todu"),
		}
	} else {
		paths = []string{"."}
	}
	return loadFromPaths(paths, true)
}

// loadFromPaths loads configuration from specified paths
// enableEnv controls whether environment variables are read
func loadFromPaths(paths []string, enableEnv bool) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("api_url", "http://localhost:8000")
	v.SetDefault("daemon.interval", "5m")
	v.SetDefault("daemon.projects", []int{})
	v.SetDefault("output.format", "text")
	v.SetDefault("output.color", true)

	// Set config file name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths to search
	for _, path := range paths {
		v.AddConfigPath(path)
	}

	// Enable environment variable support with TODU_ prefix
	if enableEnv {
		v.SetEnvPrefix("TODU")
		// Replace dots with underscores for nested config keys
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	// Read config file, but don't error if it doesn't exist
	if err := v.ReadInConfig(); err != nil {
		// Only return error if it's not a "file not found" error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// File not found is fine, we'll use defaults
	}

	// Unmarshal into Config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
