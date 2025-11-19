package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Manage external task management systems",
	Long: `Manage external task management systems (plugins).

Systems represent external platforms like GitHub, Forgejo, or Todoist
that todu integrates with. Each system uses a plugin to communicate
with the external service.`,
}

var systemListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered systems",
	Long: `List all systems registered in todu.

Shows the systems that have been configured, along with their
identifiers and current configuration status.`,
	RunE: runSystemList,
}

var systemAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new system",
	Long: `Add a new external system to todu.

The identifier must match a registered plugin name. After creating
the system, you'll need to configure it using environment variables.`,
	RunE: runSystemAdd,
}

var systemShowCmd = &cobra.Command{
	Use:   "show <id|identifier>",
	Short: "Show system details",
	Long:  `Display detailed information about a specific system.

You can specify either the system ID (numeric) or the system identifier (e.g., "github").`,
	Args: cobra.ExactArgs(1),
	RunE: runSystemShow,
}

var systemConfigCmd = &cobra.Command{
	Use:   "config <identifier>",
	Short: "Show configuration requirements for a plugin",
	Long: `Display the configuration requirements for a plugin.

Shows the environment variables needed to configure the plugin
and whether they are currently set.`,
	Args: cobra.ExactArgs(1),
	RunE: runSystemConfig,
}

var systemRemoveCmd = &cobra.Command{
	Use:   "remove <id|identifier>",
	Short: "Remove a system",
	Long:  `Remove a system from todu.

You can specify either the system ID (numeric) or the system identifier (e.g., "github").`,
	Args: cobra.ExactArgs(1),
	RunE: runSystemRemove,
}

var (
	systemAddIdentifier string
	systemAddName       string
	systemAddURL        string
	systemAddMetadata   []string
	systemRemoveForce   bool
	outputFormat        string
)

func init() {
	rootCmd.AddCommand(systemCmd)
	systemCmd.AddCommand(systemListCmd)
	systemCmd.AddCommand(systemAddCmd)
	systemCmd.AddCommand(systemShowCmd)
	systemCmd.AddCommand(systemConfigCmd)
	systemCmd.AddCommand(systemRemoveCmd)

	// system add flags
	systemAddCmd.Flags().StringVar(&systemAddIdentifier, "identifier", "", "Plugin identifier (required)")
	systemAddCmd.Flags().StringVar(&systemAddName, "name", "", "Display name for the system (required)")
	systemAddCmd.Flags().StringVar(&systemAddURL, "url", "", "API URL for the system")
	systemAddCmd.Flags().StringArrayVar(&systemAddMetadata, "metadata", []string{}, "Metadata key=value pairs (repeatable)")
	systemAddCmd.MarkFlagRequired("identifier")
	systemAddCmd.MarkFlagRequired("name")

	// system remove flags
	systemRemoveCmd.Flags().BoolVar(&systemRemoveForce, "force", false, "Skip confirmation prompt")

	// system list flags
	systemListCmd.Flags().StringVar(&outputFormat, "format", "table", "Output format (table or json)")
}

func runSystemList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)
	systems, err := client.ListSystems(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list systems: %w", err)
	}

	if outputFormat == "json" {
		// TODO: Implement JSON output
		return fmt.Errorf("JSON output not yet implemented")
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tIDENTIFIER\tNAME\tURL\tCONFIGURED")

	for _, system := range systems {
		url := "(none)"
		if system.URL != nil {
			url = *system.URL
		}

		// Check if plugin is configured
		pluginConfig, _ := registry.LoadPluginConfig(system.Identifier)
		configured := "no"
		if len(pluginConfig) > 0 {
			configured = "yes"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			system.ID,
			system.Identifier,
			system.Name,
			url,
			configured,
		)
	}

	w.Flush()
	return nil
}

func runSystemAdd(cmd *cobra.Command, args []string) error {
	// Validate that plugin exists
	availablePlugins := registry.List()
	pluginExists := false
	for _, p := range availablePlugins {
		if p == systemAddIdentifier {
			pluginExists = true
			break
		}
	}

	if !pluginExists {
		fmt.Fprintf(os.Stderr, "Error: Plugin %q not found\n\n", systemAddIdentifier)
		fmt.Fprintln(os.Stderr, "Available plugins:")
		for _, p := range availablePlugins {
			fmt.Fprintf(os.Stderr, "  - %s\n", p)
		}
		return fmt.Errorf("plugin %q not registered", systemAddIdentifier)
	}

	// Parse metadata
	metadata := make(map[string]string)
	for _, kv := range systemAddMetadata {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format %q: expected key=value", kv)
		}
		metadata[parts[0]] = parts[1]
	}

	// Create system via API
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	var urlPtr *string
	if systemAddURL != "" {
		urlPtr = &systemAddURL
	}

	systemCreate := &types.SystemCreate{
		Identifier: systemAddIdentifier,
		Name:       systemAddName,
		URL:        urlPtr,
		Metadata:   metadata,
	}

	system, err := client.CreateSystem(context.Background(), systemCreate)
	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	fmt.Printf("Created system %d: %s (%s)\n", system.ID, system.Name, system.Identifier)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Configure the plugin by setting environment variables\n")
	fmt.Printf("  2. Run: todu system config %s\n", systemAddIdentifier)

	return nil
}

func runSystemShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	// Try to parse as integer ID first
	var system *types.System
	var numID int
	if _, err := fmt.Sscanf(args[0], "%d", &numID); err == nil {
		system, err = client.GetSystem(context.Background(), numID)
		if err != nil {
			return fmt.Errorf("failed to get system: %w", err)
		}
	} else {
		// Treat as identifier - search through all systems
		identifier := args[0]
		systems, err := client.ListSystems(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list systems: %w", err)
		}

		for _, s := range systems {
			if s.Identifier == identifier {
				system = s
				break
			}
		}

		if system == nil {
			return fmt.Errorf("system with identifier %q not found", identifier)
		}
	}

	fmt.Printf("System ID: %d\n", system.ID)
	fmt.Printf("Identifier: %s\n", system.Identifier)
	fmt.Printf("Name: %s\n", system.Name)

	if system.URL != nil {
		fmt.Printf("URL: %s\n", *system.URL)
	}

	if len(system.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for key, value := range system.Metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Printf("\nCreated: %s\n", system.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", system.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Check configuration
	pluginConfig, _ := registry.LoadPluginConfig(system.Identifier)
	fmt.Println("\nPlugin Configuration:")
	if len(pluginConfig) == 0 {
		fmt.Println("  Not configured")
		fmt.Printf("\n  Run: todu system config %s\n", system.Identifier)
	} else {
		fmt.Println("  Configured")
		for key := range pluginConfig {
			fmt.Printf("  - %s: set\n", key)
		}
	}

	return nil
}

func runSystemConfig(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	// Check if plugin exists
	availablePlugins := registry.List()
	pluginExists := false
	for _, p := range availablePlugins {
		if p == identifier {
			pluginExists = true
			break
		}
	}

	if !pluginExists {
		fmt.Fprintf(os.Stderr, "Error: Plugin %q not found\n\n", identifier)
		fmt.Fprintln(os.Stderr, "Available plugins:")
		for _, p := range availablePlugins {
			fmt.Fprintf(os.Stderr, "  - %s\n", p)
		}
		return fmt.Errorf("plugin %q not registered", identifier)
	}

	fmt.Printf("Configuration for plugin: %s\n\n", identifier)

	// Load current configuration
	currentConfig, _ := registry.LoadPluginConfig(identifier)

	// Common configuration keys (plugins may require different keys)
	fmt.Println("Common configuration variables:")
	fmt.Printf("  TODU_PLUGIN_%s_TOKEN\n", strings.ToUpper(identifier))
	fmt.Printf("  TODU_PLUGIN_%s_URL\n", strings.ToUpper(identifier))
	fmt.Println()

	if len(currentConfig) > 0 {
		fmt.Println("Currently configured:")
		for key, value := range currentConfig {
			// Show first few characters of sensitive values
			displayValue := value
			if key == "token" || key == "api_key" || key == "password" {
				if len(value) > 8 {
					displayValue = value[:8] + "..."
				}
			}
			fmt.Printf("  %s: %s\n", key, displayValue)
		}
	} else {
		fmt.Println("Status: Not configured")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Printf("  export TODU_PLUGIN_%s_TOKEN=your-token-here\n", strings.ToUpper(identifier))
		fmt.Printf("  export TODU_PLUGIN_%s_URL=https://api.example.com\n", strings.ToUpper(identifier))
	}

	return nil
}

func runSystemRemove(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	// Try to parse as integer ID first, otherwise treat as identifier
	var system *types.System
	var id int
	if _, err := fmt.Sscanf(args[0], "%d", &id); err == nil {
		// It's a numeric ID
		system, err = client.GetSystem(context.Background(), id)
		if err != nil {
			return fmt.Errorf("failed to get system: %w", err)
		}
	} else {
		// Treat as identifier - search through all systems
		identifier := args[0]
		systems, err := client.ListSystems(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list systems: %w", err)
		}

		for _, s := range systems {
			if s.Identifier == identifier {
				system = s
				id = s.ID
				break
			}
		}

		if system == nil {
			return fmt.Errorf("system with identifier %q not found", identifier)
		}
	}

	// Confirm deletion unless --force
	if !systemRemoveForce {
		fmt.Printf("Are you sure you want to remove system %d (%s)? [y/N]: ", id, system.Name)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete system
	err = client.DeleteSystem(context.Background(), id)
	if err != nil {
		// Check if it's because of associated projects
		if strings.Contains(err.Error(), "projects") || strings.Contains(err.Error(), "associated") {
			fmt.Fprintf(os.Stderr, "Error: Cannot remove system with associated projects\n\n")
			fmt.Fprintln(os.Stderr, "Please remove all projects from this system first:")
			fmt.Fprintf(os.Stderr, "  todu project list --system %d\n", id)
			return err
		}
		return fmt.Errorf("failed to remove system: %w", err)
	}

	fmt.Printf("Removed system %d: %s\n", id, system.Name)
	return nil
}
