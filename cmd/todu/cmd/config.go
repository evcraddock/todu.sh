package cmd

import (
	"fmt"
	"strings"

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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
}
