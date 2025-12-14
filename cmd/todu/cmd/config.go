package cmd

import (
	"fmt"
	"strings"

	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage todu configuration",
	Long: `Manage todu configuration settings.

Use this command to view and manage your todu configuration.
Configuration can be set via config files or environment variables.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Display the current todu configuration.

This shows the effective configuration after merging:
  1. Default values
  2. Configuration file (~/.todu/config.yaml or ~/.config/todu/config.yaml)
  3. Environment variables (TODU_* prefix)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		fmt.Println("Current Configuration:")
		fmt.Println()

		// API Configuration
		fmt.Println("API:")
		fmt.Printf("  URL: %s\n", cfg.APIURL)
		if cfg.APIKey != "" {
			// Mask the API key, showing only first 8 characters
			masked := cfg.APIKey
			if len(masked) > 8 {
				masked = masked[:8] + "..."
			}
			fmt.Printf("  Key: %s (configured)\n", masked)
		} else {
			fmt.Println("  Key: (not set)")
		}
		fmt.Println()

		// Daemon Configuration
		fmt.Println("Daemon:")
		fmt.Printf("  Interval: %s\n", cfg.Daemon.Interval)
		if len(cfg.Daemon.Projects) > 0 {
			projectIDs := make([]string, len(cfg.Daemon.Projects))
			for i, id := range cfg.Daemon.Projects {
				projectIDs[i] = fmt.Sprintf("%d", id)
			}
			fmt.Printf("  Projects: [%s]\n", strings.Join(projectIDs, ", "))
		} else {
			fmt.Println("  Projects: [] (sync all projects)")
		}
		fmt.Println()

		// Output Configuration
		fmt.Println("Output:")
		fmt.Printf("  Format: %s\n", cfg.Output.Format)
		fmt.Printf("  Color:  %t\n", cfg.Output.Color)
		fmt.Println()

		// Defaults Configuration
		fmt.Println("Defaults:")
		if cfg.Author != "" {
			fmt.Printf("  Author:  %s\n", cfg.Author)
		} else {
			fmt.Println("  Author:  (not set)")
		}
		if cfg.Defaults.Project != "" {
			fmt.Printf("  Project: %s\n", cfg.Defaults.Project)
		} else {
			fmt.Println("  Project: (not set)")
		}
		fmt.Println()

		// Paths Configuration
		fmt.Println("Paths:")
		if cfg.LocalReports != "" {
			fmt.Printf("  Local Reports: %s\n", cfg.LocalReports)
		} else {
			fmt.Println("  Local Reports: (not set)")
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long:  `Set configuration values. Use subcommands to set specific settings.`,
}

var configSetAPIKeyCmd = &cobra.Command{
	Use:   "api-key <key>",
	Short: "Set the API key for authentication",
	Long: `Set the API key used for authenticating with the todu-api.

The API key will be stored in ~/.config/todu/config.yaml.

To generate an API key, use the todu-keys CLI tool in the todu-api project:
  uv run todu-keys create --name "CLI"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := args[0]
		configPath := GetConfigFile()

		if err := config.SetAPIKey(configPath, apiKey); err != nil {
			return fmt.Errorf("failed to set API key: %w", err)
		}

		// Get the actual path used for display
		if configPath == "" {
			configPath, _ = config.GetConfigPath()
		}
		fmt.Printf("API key saved to %s\n", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configSetCmd.AddCommand(configSetAPIKeyCmd)
}
