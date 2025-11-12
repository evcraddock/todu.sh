package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long: `Manage projects linked to external task management systems.

Projects represent external resources (like GitHub repositories) that
have been registered with todu for task synchronization.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long: `List all projects registered in todu.

Shows projects that have been linked to external systems, along with
their sync status and configuration.`,
	RunE: runProjectList,
}

var projectAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new project",
	Long: `Add a new project linked to an external system.

The project will be registered in todu and can then be synchronized
with the external system.`,
	RunE: runProjectAdd,
}

var projectShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show project details",
	Long:  `Display detailed information about a specific project.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectShow,
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a project",
	Long: `Update project fields.

Only the specified fields will be updated. Other fields will remain unchanged.`,
	Args: cobra.ExactArgs(1),
	RunE: runProjectUpdate,
}

var projectRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a project",
	Long:  `Remove a project from todu.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectRemove,
}

var projectDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover projects from external system",
	Long: `Discover available projects from an external system.

This queries the external system (via its plugin) for all accessible
projects and shows which ones are already registered in todu.`,
	RunE: runProjectDiscover,
}

var (
	projectListSystem     int
	projectListFormat     string
	projectAddSystem      int
	projectAddExternalID  string
	projectAddName        string
	projectAddDescription string
	projectAddStatus      string
	projectAddSyncStrategy string
	projectUpdateName        string
	projectUpdateDescription string
	projectUpdateStatus      string
	projectUpdateSyncStrategy string
	projectRemoveForce       bool
	projectDiscoverSystem    int
	projectDiscoverAutoImport bool
)

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectRemoveCmd)
	projectCmd.AddCommand(projectDiscoverCmd)

	// project list flags
	projectListCmd.Flags().IntVar(&projectListSystem, "system", 0, "Filter by system ID")
	projectListCmd.Flags().StringVar(&projectListFormat, "format", "table", "Output format (table or json)")

	// project add flags
	projectAddCmd.Flags().IntVar(&projectAddSystem, "system", 0, "System ID (required)")
	projectAddCmd.Flags().StringVar(&projectAddExternalID, "external-id", "", "External ID in the system (required)")
	projectAddCmd.Flags().StringVar(&projectAddName, "name", "", "Project name (required)")
	projectAddCmd.Flags().StringVar(&projectAddDescription, "description", "", "Project description")
	projectAddCmd.Flags().StringVar(&projectAddStatus, "status", "active", "Project status")
	projectAddCmd.Flags().StringVar(&projectAddSyncStrategy, "sync-strategy", "bidirectional", "Sync strategy (pull, push, or bidirectional)")
	projectAddCmd.MarkFlagRequired("system")
	projectAddCmd.MarkFlagRequired("external-id")
	projectAddCmd.MarkFlagRequired("name")

	// project update flags
	projectUpdateCmd.Flags().StringVar(&projectUpdateName, "name", "", "Project name")
	projectUpdateCmd.Flags().StringVar(&projectUpdateDescription, "description", "", "Project description")
	projectUpdateCmd.Flags().StringVar(&projectUpdateStatus, "status", "", "Project status")
	projectUpdateCmd.Flags().StringVar(&projectUpdateSyncStrategy, "sync-strategy", "", "Sync strategy (pull, push, or bidirectional)")

	// project remove flags
	projectRemoveCmd.Flags().BoolVar(&projectRemoveForce, "force", false, "Skip confirmation prompt")

	// project discover flags
	projectDiscoverCmd.Flags().IntVar(&projectDiscoverSystem, "system", 0, "System ID (required)")
	projectDiscoverCmd.Flags().BoolVar(&projectDiscoverAutoImport, "auto-import", false, "Automatically import all discovered projects")
	projectDiscoverCmd.MarkFlagRequired("system")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	var systemIDPtr *int
	if projectListSystem != 0 {
		systemIDPtr = &projectListSystem
	}

	projects, err := client.ListProjects(context.Background(), systemIDPtr)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	// Fetch all systems to map IDs to names
	systems, err := client.ListSystems(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list systems: %w", err)
	}

	systemNames := make(map[int]string)
	for _, sys := range systems {
		systemNames[sys.ID] = sys.Identifier
	}

	if projectListFormat == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(projects)
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSYSTEM\tEXTERNAL ID\tSTATUS\tSYNC STRATEGY")

	for _, project := range projects {
		systemName := systemNames[project.SystemID]
		if systemName == "" {
			systemName = fmt.Sprintf("(unknown:%d)", project.SystemID)
		}

		syncStrategy := project.SyncStrategy
		if syncStrategy == "" {
			syncStrategy = "bidirectional"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			project.ID,
			truncateString(project.Name, 30),
			systemName,
			truncateString(project.ExternalID, 30),
			project.Status,
			syncStrategy,
		)
	}

	w.Flush()
	return nil
}

func runProjectAdd(cmd *cobra.Command, args []string) error {
	// Validate sync strategy
	validStrategies := map[string]bool{
		"pull":          true,
		"push":          true,
		"bidirectional": true,
	}
	if !validStrategies[projectAddSyncStrategy] {
		return fmt.Errorf("invalid sync strategy %q: must be pull, push, or bidirectional", projectAddSyncStrategy)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	// Validate system exists
	_, err = client.GetSystem(context.Background(), projectAddSystem)
	if err != nil {
		return fmt.Errorf("system %d not found: %w", projectAddSystem, err)
	}

	var descPtr *string
	if projectAddDescription != "" {
		descPtr = &projectAddDescription
	}

	projectCreate := &types.ProjectCreate{
		Name:         projectAddName,
		Description:  descPtr,
		SystemID:     projectAddSystem,
		ExternalID:   projectAddExternalID,
		Status:       projectAddStatus,
		SyncStrategy: projectAddSyncStrategy,
	}

	project, err := client.CreateProject(context.Background(), projectCreate)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created project %d: %s\n", project.ID, project.Name)
	fmt.Printf("  System: %d\n", project.SystemID)
	fmt.Printf("  External ID: %s\n", project.ExternalID)
	fmt.Printf("  Status: %s\n", project.Status)
	fmt.Printf("  Sync Strategy: %s\n", project.SyncStrategy)

	return nil
}

func runProjectShow(cmd *cobra.Command, args []string) error {
	var projectID int
	if _, err := fmt.Sscanf(args[0], "%d", &projectID); err != nil {
		return fmt.Errorf("invalid project ID %q: must be a number", args[0])
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	project, err := client.GetProject(context.Background(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Fetch system details
	system, err := client.GetSystem(context.Background(), project.SystemID)
	if err != nil {
		return fmt.Errorf("failed to get system: %w", err)
	}

	fmt.Printf("Project ID: %d\n", project.ID)
	fmt.Printf("Name: %s\n", project.Name)

	if project.Description != nil {
		fmt.Printf("Description: %s\n", *project.Description)
	}

	syncStrategy := project.SyncStrategy
	if syncStrategy == "" {
		syncStrategy = "bidirectional"
	}

	fmt.Printf("System: %s (ID: %d)\n", system.Identifier, system.ID)
	fmt.Printf("External ID: %s\n", project.ExternalID)
	fmt.Printf("Status: %s\n", project.Status)
	fmt.Printf("Sync Strategy: %s\n", syncStrategy)
	fmt.Printf("\nCreated: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runProjectUpdate(cmd *cobra.Command, args []string) error {
	var projectID int
	if _, err := fmt.Sscanf(args[0], "%d", &projectID); err != nil {
		return fmt.Errorf("invalid project ID %q: must be a number", args[0])
	}

	// Validate sync strategy if provided
	if projectUpdateSyncStrategy != "" {
		validStrategies := map[string]bool{
			"pull":          true,
			"push":          true,
			"bidirectional": true,
		}
		if !validStrategies[projectUpdateSyncStrategy] {
			return fmt.Errorf("invalid sync strategy %q: must be pull, push, or bidirectional", projectUpdateSyncStrategy)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	projectUpdate := &types.ProjectUpdate{}

	if projectUpdateName != "" {
		projectUpdate.Name = &projectUpdateName
	}
	if projectUpdateDescription != "" {
		projectUpdate.Description = &projectUpdateDescription
	}
	if projectUpdateStatus != "" {
		projectUpdate.Status = &projectUpdateStatus
	}
	if projectUpdateSyncStrategy != "" {
		projectUpdate.SyncStrategy = &projectUpdateSyncStrategy
	}

	project, err := client.UpdateProject(context.Background(), projectID, projectUpdate)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("Updated project %d: %s\n", project.ID, project.Name)
	fmt.Printf("  Status: %s\n", project.Status)
	fmt.Printf("  Sync Strategy: %s\n", project.SyncStrategy)

	return nil
}

func runProjectRemove(cmd *cobra.Command, args []string) error {
	var projectID int
	if _, err := fmt.Sscanf(args[0], "%d", &projectID); err != nil {
		return fmt.Errorf("invalid project ID %q: must be a number", args[0])
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	// Fetch project for confirmation
	project, err := client.GetProject(context.Background(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Confirm deletion unless --force
	if !projectRemoveForce {
		fmt.Printf("Are you sure you want to remove project %d (%s)? [y/N]: ", projectID, project.Name)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete project
	err = client.DeleteProject(context.Background(), projectID)
	if err != nil {
		// Check if it's because of associated tasks
		if strings.Contains(err.Error(), "tasks") || strings.Contains(err.Error(), "associated") {
			fmt.Fprintf(os.Stderr, "Error: Cannot remove project with associated tasks\n\n")
			fmt.Fprintln(os.Stderr, "Please remove all tasks from this project first:")
			fmt.Fprintf(os.Stderr, "  todu task list --project %d\n", projectID)
			return err
		}
		return fmt.Errorf("failed to remove project: %w", err)
	}

	fmt.Printf("Removed project %d: %s\n", projectID, project.Name)
	return nil
}

func runProjectDiscover(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	// Get system details
	system, err := client.GetSystem(context.Background(), projectDiscoverSystem)
	if err != nil {
		return fmt.Errorf("failed to get system: %w", err)
	}

	// Load plugin configuration
	pluginConfig, err := registry.LoadPluginConfig(system.Identifier)
	if err != nil {
		return fmt.Errorf("failed to load plugin config: %w", err)
	}

	// Get plugin instance (creates and configures it)
	plugin, err := registry.Create(system.Identifier, pluginConfig)
	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	// Validate configuration
	if err := plugin.ValidateConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Plugin %s is not properly configured\n\n", system.Identifier)
		fmt.Fprintf(os.Stderr, "Please configure the plugin:\n")
		fmt.Fprintf(os.Stderr, "  todu system config %s\n", system.Identifier)
		return err
	}

	fmt.Printf("Discovering projects from %s...\n\n", system.Name)

	// Fetch projects from plugin
	externalProjects, err := plugin.FetchProjects(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch projects from plugin: %w", err)
	}

	// Get existing projects for this system
	existingProjects, err := client.ListProjects(context.Background(), &projectDiscoverSystem)
	if err != nil {
		return fmt.Errorf("failed to list existing projects: %w", err)
	}

	// Map external IDs to existing projects
	existingMap := make(map[string]bool)
	for _, p := range existingProjects {
		existingMap[p.ExternalID] = true
	}

	// Display results
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EXTERNAL ID\tNAME\tSTATUS")

	var newProjects []*types.Project
	for _, project := range externalProjects {
		status := "NEW"
		if existingMap[project.ExternalID] {
			status = "Already imported"
		} else {
			newProjects = append(newProjects, project)
		}

		desc := project.Name
		if project.Description != nil && *project.Description != "" {
			desc = truncateString(*project.Description, 40)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n",
			truncateString(project.ExternalID, 30),
			truncateString(desc, 40),
			status,
		)
	}

	w.Flush()

	// Auto-import if requested
	if projectDiscoverAutoImport && len(newProjects) > 0 {
		fmt.Printf("\nImporting %d new projects...\n", len(newProjects))

		for _, project := range newProjects {
			projectCreate := &types.ProjectCreate{
				Name:         project.Name,
				Description:  project.Description,
				SystemID:     projectDiscoverSystem,
				ExternalID:   project.ExternalID,
				Status:       "active",
				SyncStrategy: "bidirectional",
			}

			created, err := client.CreateProject(context.Background(), projectCreate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to import %s: %v\n", project.ExternalID, err)
				continue
			}

			fmt.Printf("  Imported: %s (ID: %d)\n", created.Name, created.ID)
		}
	} else if len(newProjects) > 0 {
		fmt.Printf("\nFound %d new projects. Use --auto-import to import them all.\n", len(newProjects))
		fmt.Println("\nTo import individual projects:")
		fmt.Printf("  todu project add --system %d --external-id <id> --name <name>\n", projectDiscoverSystem)
	}

	return nil
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
