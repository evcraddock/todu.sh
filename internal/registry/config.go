package registry

import (
	"fmt"
	"os"
	"strings"
)

// LoadPluginConfig loads configuration for a plugin from environment variables.
//
// Environment variables are expected to follow the pattern:
//
//	TODU_PLUGIN_{PLUGINNAME}_{KEY}
//
// Where:
//   - PLUGINNAME is the plugin name in uppercase
//   - KEY is the configuration key in uppercase
//
// The function returns a map with lowercase keys.
//
// Example:
//
//	export TODU_PLUGIN_GITHUB_TOKEN=ghp_abc123
//	export TODU_PLUGIN_GITHUB_URL=https://api.github.com
//
//	config, err := LoadPluginConfig("github")
//	// Returns: map[string]string{
//	//     "token": "ghp_abc123",
//	//     "url": "https://api.github.com",
//	// }
func LoadPluginConfig(pluginName string) (map[string]string, error) {
	if pluginName == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}

	config := make(map[string]string)
	prefix := fmt.Sprintf("TODU_PLUGIN_%s_", strings.ToUpper(pluginName))

	// Scan all environment variables
	for _, env := range os.Environ() {
		// Split into key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Check if this is a plugin config variable
		if strings.HasPrefix(key, prefix) {
			// Extract the config key (after the prefix)
			configKey := strings.TrimPrefix(key, prefix)
			// Convert to lowercase
			configKey = strings.ToLower(configKey)

			config[configKey] = value
		}
	}

	return config, nil
}
