package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Use a temporary directory to ensure no config file exists
	tmpDir := t.TempDir()

	// Load configuration from temp dir only (no env vars)
	config, err := loadFromPaths([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Expected no error when loading defaults, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify defaults
	if config.APIURL != "http://localhost:8000" {
		t.Errorf("Expected APIURL to be 'http://localhost:8000', got '%s'", config.APIURL)
	}

	if config.Daemon.Interval != "5m" {
		t.Errorf("Expected Daemon.Interval to be '5m', got '%s'", config.Daemon.Interval)
	}

	if len(config.Daemon.Projects) != 0 {
		t.Errorf("Expected Daemon.Projects to be empty, got %d items", len(config.Daemon.Projects))
	}

	if config.Output.Format != "text" {
		t.Errorf("Expected Output.Format to be 'text', got '%s'", config.Output.Format)
	}

	if config.Output.Color != true {
		t.Errorf("Expected Output.Color to be true, got %v", config.Output.Color)
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a config file
	configContent := `api_url: http://example.com:9000
daemon:
  interval: 10m
  projects:
    - 1
    - 2
    - 3
output:
  format: json
  color: false
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load configuration from temp dir only (no env vars)
	config, err := loadFromPaths([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Expected no error when loading config file, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify values from config file
	if config.APIURL != "http://example.com:9000" {
		t.Errorf("Expected APIURL to be 'http://example.com:9000', got '%s'", config.APIURL)
	}

	if config.Daemon.Interval != "10m" {
		t.Errorf("Expected Daemon.Interval to be '10m', got '%s'", config.Daemon.Interval)
	}

	if len(config.Daemon.Projects) != 3 {
		t.Errorf("Expected Daemon.Projects to have 3 items, got %d", len(config.Daemon.Projects))
	}

	if config.Output.Format != "json" {
		t.Errorf("Expected Output.Format to be 'json', got '%s'", config.Output.Format)
	}

	if config.Output.Color != false {
		t.Errorf("Expected Output.Color to be false, got %v", config.Output.Color)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Use a temporary directory to ensure no config file exists
	tmpDir := t.TempDir()

	// Set environment variables
	os.Setenv("TODU_API_URL", "http://env.example.com:8080")
	defer os.Unsetenv("TODU_API_URL")

	os.Setenv("TODU_OUTPUT_FORMAT", "json")
	defer os.Unsetenv("TODU_OUTPUT_FORMAT")

	// Load configuration from temp dir with env vars enabled
	config, err := loadFromPaths([]string{tmpDir}, true)
	if err != nil {
		t.Fatalf("Expected no error when loading with env vars, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify environment variables override defaults
	if config.APIURL != "http://env.example.com:8080" {
		t.Errorf("Expected APIURL from env to be 'http://env.example.com:8080', got '%s'", config.APIURL)
	}

	if config.Output.Format != "json" {
		t.Errorf("Expected Output.Format from env to be 'json', got '%s'", config.Output.Format)
	}

	// Verify defaults are still used for non-overridden values
	if config.Daemon.Interval != "5m" {
		t.Errorf("Expected Daemon.Interval default to be '5m', got '%s'", config.Daemon.Interval)
	}
}

func TestLoadNoErrorOnMissingFile(t *testing.T) {
	// Use a temporary directory to ensure no config file exists
	tmpDir := t.TempDir()

	// Load configuration - should not error
	config, err := loadFromPaths([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Expected no error when config file is missing, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil even when file is missing")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create an invalid config file
	invalidContent := `api_url: http://example.com
daemon:
  interval: 10m
  projects: [1, 2, 3
output:
  format: json
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load configuration - should error because of invalid YAML
	_, err := loadFromPaths([]string{tmpDir}, false)
	if err == nil {
		t.Fatal("Expected error when loading invalid YAML, got nil")
	}
}

func TestLocalConfigOverridesGlobal(t *testing.T) {
	// Create two temporary directories: one for "global" config, one for "local"
	globalDir := t.TempDir()
	localDir := t.TempDir()

	// Create global config
	globalContent := `api_url: http://global.example.com:8000
output:
  format: json
`
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("Failed to write global config file: %v", err)
	}

	// Create local config with different values
	localContent := `api_url: http://local.example.com:9000
output:
  format: text
`
	if err := os.WriteFile(filepath.Join(localDir, "config.yaml"), []byte(localContent), 0644); err != nil {
		t.Fatalf("Failed to write local config file: %v", err)
	}

	// Load configuration with local dir first, then global dir (no env vars)
	// This simulates the behavior where local config should override global
	config, err := loadFromPaths([]string{localDir, globalDir}, false)
	if err != nil {
		t.Fatalf("Expected no error when loading config files, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify that local config values are used (not global)
	if config.APIURL != "http://local.example.com:9000" {
		t.Errorf("Expected APIURL from local config to be 'http://local.example.com:9000', got '%s'", config.APIURL)
	}

	if config.Output.Format != "text" {
		t.Errorf("Expected Output.Format from local config to be 'text', got '%s'", config.Output.Format)
	}
}
