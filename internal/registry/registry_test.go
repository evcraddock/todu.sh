package registry

import (
	"os"
	"testing"

	"github.com/evcraddock/todu.sh/pkg/plugin"
)

func TestNewRegistry(t *testing.T) {
	reg := New()
	if reg == nil {
		t.Fatal("Expected non-nil registry")
	}

	if len(reg.factories) != 0 {
		t.Errorf("Expected empty registry, got %d plugins", len(reg.factories))
	}
}

func TestRegisterPlugin(t *testing.T) {
	reg := New()

	factory := func() plugin.Plugin {
		return plugin.NewMockPlugin("test")
	}

	err := reg.Register("test-plugin", factory)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify plugin was registered
	names := reg.List()
	if len(names) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(names))
	}
	if names[0] != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", names[0])
	}
}

func TestRegisterDuplicateName(t *testing.T) {
	reg := New()

	factory := func() plugin.Plugin {
		return plugin.NewMockPlugin("test")
	}

	// Register first time - should succeed
	err := reg.Register("test-plugin", factory)
	if err != nil {
		t.Fatalf("Expected no error on first registration, got %v", err)
	}

	// Register second time - should fail
	err = reg.Register("test-plugin", factory)
	if err == nil {
		t.Fatal("Expected error when registering duplicate plugin name")
	}
}

func TestRegisterInvalidNames(t *testing.T) {
	reg := New()
	factory := func() plugin.Plugin {
		return plugin.NewMockPlugin("test")
	}

	invalidNames := []string{
		"Test",        // uppercase
		"test_plugin", // underscore
		"test plugin", // space
		"test.plugin", // dot
		"test/plugin", // slash
		"test@plugin", // special char
		"",            // empty
		"123",         // valid (numbers ok)
		"test-123",    // valid (hyphen ok)
	}

	for _, name := range invalidNames {
		err := reg.Register(name, factory)
		// Only some should fail
		if name == "Test" || name == "test_plugin" || name == "test plugin" ||
			name == "test.plugin" || name == "test/plugin" || name == "test@plugin" || name == "" {
			if err == nil {
				t.Errorf("Expected error for invalid name %q, got none", name)
			}
		}
	}
}

func TestListPlugins(t *testing.T) {
	reg := New()

	// Empty registry
	names := reg.List()
	if len(names) != 0 {
		t.Errorf("Expected empty list, got %d plugins", len(names))
	}

	// Add plugins
	plugins := []string{"github", "forgejo", "todoist"}
	for _, name := range plugins {
		_ = reg.Register(name, func() plugin.Plugin {
			return plugin.NewMockPlugin(name)
		})
	}

	// List should be sorted
	names = reg.List()
	if len(names) != 3 {
		t.Fatalf("Expected 3 plugins, got %d", len(names))
	}

	// Check sorted order
	expected := []string{"forgejo", "github", "todoist"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("Expected plugin %d to be %q, got %q", i, name, names[i])
		}
	}
}

func TestCreatePlugin(t *testing.T) {
	reg := New()

	// Register a plugin
	_ = reg.Register("test-plugin", func() plugin.Plugin {
		return plugin.NewMockPlugin("test")
	})

	// Create instance
	config := map[string]string{
		"token": "test-token",
		"url":   "https://example.com",
	}

	p, err := reg.Create("test-plugin", config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if p == nil {
		t.Fatal("Expected non-nil plugin")
	}

	if p.Name() != "test" {
		t.Errorf("Expected plugin name 'test', got '%s'", p.Name())
	}
}

func TestCreateUnregisteredPlugin(t *testing.T) {
	reg := New()

	config := map[string]string{"token": "test-token"}

	_, err := reg.Create("non-existent", config)
	if err == nil {
		t.Fatal("Expected error when creating unregistered plugin")
	}
}

func TestCreateWithConfigurationError(t *testing.T) {
	reg := New()

	// Register a plugin
	_ = reg.Register("test-plugin", func() plugin.Plugin {
		mock := plugin.NewMockPlugin("test")
		// Inject a configuration error
		mock.ConfigureError = plugin.NewErrNotConfigured("test error")
		return mock
	})

	config := map[string]string{"token": "test-token"}

	_, err := reg.Create("test-plugin", config)
	if err == nil {
		t.Fatal("Expected error when plugin configuration fails")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Note: This test uses the global Default registry
	// We need to be careful about test pollution

	initialCount := len(List())

	// Register via package-level function
	err := Register("test-global", func() plugin.Plugin {
		return plugin.NewMockPlugin("global")
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify via package-level function
	names := List()
	if len(names) != initialCount+1 {
		t.Errorf("Expected %d plugins, got %d", initialCount+1, len(names))
	}

	// Create via package-level function
	config := map[string]string{"token": "test-token"}
	p, err := Create("test-global", config)
	if err != nil {
		t.Fatalf("Expected no error creating plugin, got %v", err)
	}

	if p.Name() != "global" {
		t.Errorf("Expected plugin name 'global', got '%s'", p.Name())
	}
}

func TestLoadPluginConfigFromEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("TODU_PLUGIN_GITHUB_TOKEN", "ghp_test123")
	os.Setenv("TODU_PLUGIN_GITHUB_URL", "https://api.github.com")
	os.Setenv("TODU_PLUGIN_GITHUB_TIMEOUT", "30")
	defer os.Unsetenv("TODU_PLUGIN_GITHUB_TOKEN")
	defer os.Unsetenv("TODU_PLUGIN_GITHUB_URL")
	defer os.Unsetenv("TODU_PLUGIN_GITHUB_TIMEOUT")

	config, err := LoadPluginConfig("github")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(config) != 3 {
		t.Errorf("Expected 3 config keys, got %d", len(config))
	}

	if config["token"] != "ghp_test123" {
		t.Errorf("Expected token 'ghp_test123', got '%s'", config["token"])
	}

	if config["url"] != "https://api.github.com" {
		t.Errorf("Expected url 'https://api.github.com', got '%s'", config["url"])
	}

	if config["timeout"] != "30" {
		t.Errorf("Expected timeout '30', got '%s'", config["timeout"])
	}
}

func TestLoadPluginConfigEmptyName(t *testing.T) {
	_, err := LoadPluginConfig("")
	if err == nil {
		t.Fatal("Expected error for empty plugin name")
	}
}

func TestLoadPluginConfigNoEnvironment(t *testing.T) {
	config, err := LoadPluginConfig("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(config) != 0 {
		t.Errorf("Expected empty config, got %d keys", len(config))
	}
}

func TestLoadPluginConfigCaseInsensitive(t *testing.T) {
	// Set with uppercase
	os.Setenv("TODU_PLUGIN_TEST_API_KEY", "test123")
	defer os.Unsetenv("TODU_PLUGIN_TEST_API_KEY")

	config, err := LoadPluginConfig("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be lowercase in returned map
	if config["api_key"] != "test123" {
		t.Errorf("Expected api_key 'test123', got '%s'", config["api_key"])
	}
}

func TestLoadPluginConfigFiltersOtherPlugins(t *testing.T) {
	// Set variables for multiple plugins
	os.Setenv("TODU_PLUGIN_GITHUB_TOKEN", "github_token")
	os.Setenv("TODU_PLUGIN_FORGEJO_TOKEN", "forgejo_token")
	defer os.Unsetenv("TODU_PLUGIN_GITHUB_TOKEN")
	defer os.Unsetenv("TODU_PLUGIN_FORGEJO_TOKEN")

	// Load only github config
	config, err := LoadPluginConfig("github")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(config) != 1 {
		t.Errorf("Expected 1 config key, got %d", len(config))
	}

	if config["token"] != "github_token" {
		t.Errorf("Expected token 'github_token', got '%s'", config["token"])
	}

	// Should not include forgejo token
	if _, exists := config["forgejo_token"]; exists {
		t.Error("Should not include forgejo token in github config")
	}
}
