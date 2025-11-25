package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/google/uuid"
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
with the external system.

For local-only projects (no external sync), you can simply use:
  todu project add --name "My Project"

This will auto-register the local system and generate an external ID.`,
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
	projectListSystem       string
	projectListPriority     []string
	projectAddSystem        string
	projectAddExternalID    string
	projectAddName          string
	projectAddDescription   string
	projectAddStatus        string
	projectAddPriority      string
	projectAddSyncStrategy  string
	projectUpdateName        string
	projectUpdateDescription string
	projectUpdateStatus      string
	projectUpdatePriority    string
	projectUpdateSyncStrategy string
	projectRemoveForce       bool
	projectRemoveCascade     bool
	projectDiscoverSystem    string
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
	projectListCmd.Flags().StringVar(&projectListSystem, "system", "", "Filter by system ID or name")
	projectListCmd.Flags().StringSliceVar(&projectListPriority, "priority", nil, "Filter by priority (low, medium, high) - can be specified multiple times")

	// project add flags
	projectAddCmd.Flags().StringVar(&projectAddSystem, "system", "local", "System ID or name (default: local)")
	projectAddCmd.Flags().StringVar(&projectAddExternalID, "external-id", "", "External ID in the system (auto-generated for local)")
	projectAddCmd.Flags().StringVar(&projectAddName, "name", "", "Project name (required)")
	projectAddCmd.Flags().StringVar(&projectAddDescription, "description", "", "Project description")
	projectAddCmd.Flags().StringVar(&projectAddStatus, "status", "active", "Project status")
	projectAddCmd.Flags().StringVar(&projectAddPriority, "priority", "", "Project priority (low, medium, high)")
	projectAddCmd.Flags().StringVar(&projectAddSyncStrategy, "sync-strategy", "bidirectional", "Sync strategy (pull, push, or bidirectional)")
	projectAddCmd.MarkFlagRequired("name")

	// project update flags
	projectUpdateCmd.Flags().StringVar(&projectUpdateName, "name", "", "Project name")
	projectUpdateCmd.Flags().StringVar(&projectUpdateDescription, "description", "", "Project description")
	projectUpdateCmd.Flags().StringVar(&projectUpdateStatus, "status", "", "Project status")
	projectUpdateCmd.Flags().StringVar(&projectUpdatePriority, "priority", "", "Project priority (low, medium, high)")
	projectUpdateCmd.Flags().StringVar(&projectUpdateSyncStrategy, "sync-strategy", "", "Sync strategy (pull, push, or bidirectional)")

	// project remove flags
	projectRemoveCmd.Flags().BoolVar(&projectRemoveForce, "force", false, "Skip confirmation prompt")
	projectRemoveCmd.Flags().BoolVar(&projectRemoveCascade, "cascade", false, "Delete associated tasks as well")

	// project discover flags
	projectDiscoverCmd.Flags().StringVar(&projectDiscoverSystem, "system", "", "System ID or name (required)")
	projectDiscoverCmd.Flags().BoolVar(&projectDiscoverAutoImport, "auto-import", false, "Automatically import all discovered projects")
	projectDiscoverCmd.MarkFlagRequired("system")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	opts := &api.ProjectListOptions{}
	if projectListSystem != "" {
		systemID, err := resolveSystemID(client, projectListSystem)
		if err != nil {
			return err
		}
		opts.SystemID = &systemID
	}
	if len(projectListPriority) > 0 {
		opts.Priority = projectListPriority
	}

	projects, err := client.ListProjects(context.Background(), opts)
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

	if GetOutputFormat() == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(projects)
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSYSTEM\tSTATUS\tPRIORITY\tLAST SYNCED")

	for _, project := range projects {
		systemName := systemNames[project.SystemID]
		if systemName == "" {
			systemName = fmt.Sprintf("(unknown:%d)", project.SystemID)
		}

		lastSynced := "Never"
		if project.LastSyncedAt != nil {
			lastSynced = project.LastSyncedAt.Format("2006-01-02 15:04")
		}

		priority := ""
		if project.Priority != nil {
			priority = *project.Priority
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			project.ID,
			truncateString(project.Name, 30),
			systemName,
			project.Status,
			priority,
			lastSynced,
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

	// Validate priority if provided
	if projectAddPriority != "" {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
		if !validPriorities[projectAddPriority] {
			return fmt.Errorf("invalid priority %q: must be low, medium, or high", projectAddPriority)
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	var systemID int

	// Handle local system auto-registration
	if projectAddSystem == "local" {
		systemID, err = ensureLocalSystem(client)
		if err != nil {
			return err
		}

		// Auto-generate external ID for local projects if not provided
		if projectAddExternalID == "" {
			projectAddExternalID = uuid.New().String()
		}
	} else {
		// Resolve system ID for non-local systems
		systemID, err = resolveSystemID(client, projectAddSystem)
		if err != nil {
			return err
		}

		// Validate system exists
		_, err = client.GetSystem(context.Background(), systemID)
		if err != nil {
			return fmt.Errorf("system %q not found: %w", projectAddSystem, err)
		}

		// External ID is required for non-local systems
		if projectAddExternalID == "" {
			return fmt.Errorf("--external-id is required for non-local systems")
		}
	}

	var descPtr *string
	if projectAddDescription != "" {
		descPtr = &projectAddDescription
	}

	var priorityPtr *string
	if projectAddPriority != "" {
		priorityPtr = &projectAddPriority
	}

	projectCreate := &types.ProjectCreate{
		Name:         projectAddName,
		Description:  descPtr,
		SystemID:     systemID,
		ExternalID:   projectAddExternalID,
		Status:       projectAddStatus,
		Priority:     priorityPtr,
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
	if project.Priority != nil {
		fmt.Printf("  Priority: %s\n", *project.Priority)
	}
	fmt.Printf("  Sync Strategy: %s\n", project.SyncStrategy)

	return nil
}

func runProjectShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Resolve project ID from name or ID
	projectID, err := resolveProjectID(ctx, client, args[0])
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Fetch system details
	system, err := client.GetSystem(ctx, project.SystemID)
	if err != nil {
		return fmt.Errorf("failed to get system: %w", err)
	}

	// Display results
	if GetOutputFormat() == "json" {
		output := map[string]interface{}{
			"project": project,
			"system":  system,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Project ID: %d\n", project.ID)
	fmt.Printf("Name: %s\n", project.Name)

	if project.Description != nil {
		fmt.Printf("Description: %s\n", *project.Description)
	}

	fmt.Printf("System: %s (ID: %d)\n", system.Identifier, system.ID)
	fmt.Printf("External ID: %s\n", project.ExternalID)
	fmt.Printf("Status: %s\n", project.Status)
	if project.Priority != nil {
		fmt.Printf("Priority: %s\n", *project.Priority)
	}
	fmt.Printf("Sync Strategy: %s\n", project.SyncStrategy)
	fmt.Printf("\nCreated: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runProjectUpdate(cmd *cobra.Command, args []string) error {
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

	// Validate priority if provided
	if projectUpdatePriority != "" {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
		if !validPriorities[projectUpdatePriority] {
			return fmt.Errorf("invalid priority %q: must be low, medium, or high", projectUpdatePriority)
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Resolve project ID from name or ID
	projectID, err := resolveProjectID(ctx, client, args[0])
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

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
	if projectUpdatePriority != "" {
		projectUpdate.Priority = &projectUpdatePriority
	}
	if projectUpdateSyncStrategy != "" {
		projectUpdate.SyncStrategy = &projectUpdateSyncStrategy
	}

	project, err := client.UpdateProject(ctx, projectID, projectUpdate)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("Updated project %d: %s\n", project.ID, project.Name)
	fmt.Printf("  Status: %s\n", project.Status)
	if project.Priority != nil {
		fmt.Printf("  Priority: %s\n", *project.Priority)
	}
	fmt.Printf("  Sync Strategy: %s\n", project.SyncStrategy)

	return nil
}

func runProjectRemove(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Resolve project ID from name or ID
	projectID, err := resolveProjectID(ctx, client, args[0])
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	// Fetch project for confirmation
	project, err := client.GetProject(ctx, projectID)
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
	err = client.DeleteProject(ctx, projectID, projectRemoveCascade)
	if err != nil {
		// Check if it's because of associated tasks
		if strings.Contains(err.Error(), "tasks") || strings.Contains(err.Error(), "associated") {
			fmt.Fprintf(os.Stderr, "\nError: Cannot remove project with associated tasks\n\n")

			// Extract task count from error message if available
			errorMsg := err.Error()
			fmt.Fprintf(os.Stderr, "%s\n\n", errorMsg)

			// Offer to cascade delete
			fmt.Fprintln(os.Stderr, "Options:")
			fmt.Fprintln(os.Stderr, "  1. Remove tasks first:")
			fmt.Fprintf(os.Stderr, "     todu task list --project %d\n", projectID)
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  2. Delete project and all associated tasks:")
			fmt.Fprintf(os.Stderr, "     todu project remove %d --cascade\n", projectID)

			// If not using --force, offer to cascade delete now
			if !projectRemoveForce && !projectRemoveCascade {
				fmt.Fprintf(os.Stderr, "\nWould you like to delete the project and all its tasks now? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "y" || response == "yes" {
					// Retry with cascade
					err = client.DeleteProject(ctx, projectID, true)
					if err != nil {
						return fmt.Errorf("failed to remove project with cascade: %w", err)
					}
					fmt.Printf("\nRemoved project %d and all associated tasks: %s\n", projectID, project.Name)
					return nil
				}
			}

			return err
		}
		return fmt.Errorf("failed to remove project: %w", err)
	}

	if projectRemoveCascade {
		fmt.Printf("Removed project %d and all associated tasks: %s\n", projectID, project.Name)
	} else {
		fmt.Printf("Removed project %d: %s\n", projectID, project.Name)
	}
	return nil
}

func runProjectDiscover(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(cfg.APIURL)

	// Resolve system ID
	systemID, err := resolveSystemID(client, projectDiscoverSystem)
	if err != nil {
		return err
	}

	// Get system details
	system, err := client.GetSystem(context.Background(), systemID)
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
	existingProjects, err := client.ListProjects(context.Background(), &api.ProjectListOptions{SystemID: &systemID})
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
				SystemID:     systemID,
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
		fmt.Printf("  todu project add --system %s --external-id <id> --name <name>\n", projectDiscoverSystem)
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
