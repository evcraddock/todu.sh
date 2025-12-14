package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	APIURL         string               `mapstructure:"api_url"`
	APIKey         string               `mapstructure:"api_key"`
	Author         string               `mapstructure:"author"`
	LocalReports   string               `mapstructure:"local_reports"`
	Daemon         DaemonConfig         `mapstructure:"daemon"`
	Output         OutputConfig         `mapstructure:"output"`
	Defaults       DefaultsConfig       `mapstructure:"defaults"`
	RecurringTasks RecurringTasksConfig `mapstructure:"recurring_tasks"`
}

// DefaultsConfig contains default values for commands
type DefaultsConfig struct {
	Project string `mapstructure:"project"`
}

// DaemonConfig contains daemon-specific settings
type DaemonConfig struct {
	Interval      string `mapstructure:"interval"`
	Projects      []int  `mapstructure:"projects"`
	LogLevel      string `mapstructure:"log_level"`
	LogMaxSizeMB  int    `mapstructure:"log_max_size_mb"`
	LogMaxBackups int    `mapstructure:"log_max_backups"`
	LogMaxAgeDays int    `mapstructure:"log_max_age_days"`
}

// RecurringTasksConfig contains recurring task processing settings
type RecurringTasksConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	Format string `mapstructure:"format"`
	Color  bool   `mapstructure:"color"`
}

// Load loads configuration from file and environment variables
// If configPath is provided, it will be used exclusively.
// Otherwise, searches in order: ./config.yaml, ~/.config/todu/config.yaml, ~/.todu/config.yaml
func Load(configPath string) (*Config, error) {
	// If a specific config path is provided, use it exclusively
	if configPath != "" {
		return loadFromFile(configPath, true)
	}

	// Otherwise use the default search paths
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

// loadFromFile loads configuration from a specific file path
// enableEnv controls whether environment variables are read
func loadFromFile(filePath string, enableEnv bool) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("api_url", "http://localhost:8000")
	v.SetDefault("api_key", "")
	v.SetDefault("author", "")
	v.SetDefault("local_reports", "")
	v.SetDefault("daemon.interval", "5m")
	v.SetDefault("daemon.projects", []int{})
	v.SetDefault("daemon.log_level", "info")
	v.SetDefault("daemon.log_max_size_mb", 10)
	v.SetDefault("daemon.log_max_backups", 5)
	v.SetDefault("daemon.log_max_age_days", 7)
	v.SetDefault("recurring_tasks.enabled", true)
	v.SetDefault("output.format", "text")
	v.SetDefault("output.color", true)
	v.SetDefault("defaults.project", "")

	// Enable environment variable support with TODU_ prefix
	if enableEnv {
		v.SetEnvPrefix("TODU")
		// Replace dots with underscores for nested config keys
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	// Set the specific config file
	v.SetConfigFile(filePath)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal into Config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// loadFromPaths loads configuration from specified paths
// enableEnv controls whether environment variables are read
func loadFromPaths(paths []string, enableEnv bool) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("api_url", "http://localhost:8000")
	v.SetDefault("api_key", "")
	v.SetDefault("author", "")
	v.SetDefault("local_reports", "")
	v.SetDefault("daemon.interval", "5m")
	v.SetDefault("daemon.projects", []int{})
	v.SetDefault("daemon.log_level", "info")
	v.SetDefault("daemon.log_max_size_mb", 10)
	v.SetDefault("daemon.log_max_backups", 5)
	v.SetDefault("daemon.log_max_age_days", 7)
	v.SetDefault("recurring_tasks.enabled", true)
	v.SetDefault("output.format", "text")
	v.SetDefault("output.color", true)
	v.SetDefault("defaults.project", "")

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

// GetConfigPath returns the path to the config file that should be used for writing.
// It prefers ~/.config/todu/config.yaml, creating the directory if needed.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Prefer XDG config directory
	configDir := filepath.Join(homeDir, ".config", "todu")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// SetAPIKey updates the api_key in the config file.
// If configPath is empty, uses the default config path.
// It preserves other settings in the file.
func SetAPIKey(configPath, apiKey string) error {
	var err error
	if configPath == "" {
		configPath, err = GetConfigPath()
		if err != nil {
			return err
		}
	}

	// Read existing config file or start with empty map
	configData := make(map[string]interface{})

	data, err := os.ReadFile(configPath)
	if err == nil {
		// File exists, parse it
		if err := yaml.Unmarshal(data, &configData); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Update api_key
	configData["api_key"] = apiKey

	// Write back to file
	newData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
