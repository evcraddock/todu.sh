package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize tasks with external systems",
	Long: `Synchronize tasks between todu and external task management systems.

Supports bidirectional sync, allowing changes in external systems to be
pulled into todu, and changes in todu to be pushed to external systems.`,
	RunE: runSync,
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status for projects",
	Long: `Display the last sync time for all projects.

Shows when each project was last synchronized, making it easy to identify
projects that need attention.`,
	RunE: runSyncStatus,
}

var (
	syncProject      string
	syncSystem       string
	syncAll          bool
	syncStrategy     string
	syncDryRun       bool
	syncStatusSystem string
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncStatusCmd)

	// Sync flags
	syncCmd.Flags().StringVarP(&syncProject, "project", "p", "", "Sync specific project by ID or name")
	syncCmd.Flags().StringVarP(&syncSystem, "system", "s", "", "Sync all projects for a system (ID or name)")
	syncCmd.Flags().BoolVarP(&syncAll, "all", "a", false, "Sync all projects (default if no filters)")
	syncCmd.Flags().StringVar(&syncStrategy, "strategy", "", "Override sync strategy (pull/push/bidirectional)")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without making them")

	// Sync status flags
	syncStatusCmd.Flags().StringVarP(&syncStatusSystem, "system", "s", "", "Filter by system ID or name")
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured. Run 'todu config show' to see configuration")
	}

	// Create API client
	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Create sync engine
	engine := sync.NewEngine(apiClient, registry.Default)

	// Build sync options
	options := sync.Options{
		DryRun: syncDryRun,
	}

	// Handle strategy override
	if syncStrategy != "" {
		strategy := sync.Strategy(syncStrategy)
		if !strategy.IsValid() {
			return fmt.Errorf("invalid strategy %q. Must be: pull, push, or bidirectional", syncStrategy)
		}
		options.StrategyOverride = &strategy
	}

	// Handle project/system filters
	if syncProject != "" {
		// Resolve project ID from name or ID
		projectID, err := resolveProjectID(ctx, apiClient, syncProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}
		options.ProjectIDs = []int{projectID}
	} else if syncSystem != "" {
		systemID, err := resolveSystemID(apiClient, syncSystem)
		if err != nil {
			return err
		}
		options.SystemID = &systemID
	}
	// If neither project nor system is specified, sync all (default behavior)

	// Display dry run notice
	if syncDryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Println("No changes will be made")
		fmt.Println()
	}

	// Run sync
	result, err := engine.Sync(ctx, options)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Display results
	fmt.Println()
	displaySyncResults(result, syncDryRun)

	// Exit with error code if there were errors
	if result.HasErrors() {
		return fmt.Errorf("sync completed with %d error(s)", result.TotalErrors)
	}

	return nil
}

func runSyncStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured. Run 'todu config show' to see configuration")
	}

	// Create API client
	apiClient := api.NewClient(cfg.APIURL)

	// Get projects
	ctx := context.Background()
	opts := &api.ProjectListOptions{}
	if syncStatusSystem != "" {
		systemID, err := resolveSystemID(apiClient, syncStatusSystem)
		if err != nil {
			return err
		}
		opts.SystemID = &systemID
	}

	projects, err := apiClient.ListProjects(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	// Display projects with sync status
	fmt.Println("Project Sync Status:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSYSTEM ID\tLAST SYNC\tSTATUS")
	fmt.Fprintln(w, "--\t----\t---------\t---------\t------")

	now := time.Now()
	for _, project := range projects {
		lastSync := "Never"
		status := "⚠️  Not synced"

		// Note: Last sync time would be stored in project metadata
		// For now, we'll use the UpdatedAt field as a proxy
		timeSince := now.Sub(project.UpdatedAt)
		if timeSince < 1*time.Hour {
			lastSync = formatDuration(timeSince) + " ago"
			status = "✓ Recent"
		} else if timeSince < 24*time.Hour {
			lastSync = formatDuration(timeSince) + " ago"
			status = "✓ Up to date"
		} else {
			lastSync = formatDuration(timeSince) + " ago"
			status = "⚠️  Stale"
		}

		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\n",
			project.ID,
			project.Name,
			project.SystemID,
			lastSync,
			status,
		)
	}

	w.Flush()
	return nil
}

func displaySyncResults(result *sync.Result, dryRun bool) {
	if dryRun {
		fmt.Println("=== DRY RUN RESULTS ===")
		fmt.Println()
	} else {
		fmt.Println("=== SYNC RESULTS ===")
		fmt.Println()
	}

	// Display per-project results
	if len(result.ProjectResults) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PROJECT\tCREATED\tUPDATED\tSKIPPED\tERRORS")
		fmt.Fprintln(w, "-------\t-------\t-------\t-------\t------")

		for _, pr := range result.ProjectResults {
			errCount := len(pr.Errors)
			fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\n",
				pr.ProjectName,
				pr.Created,
				pr.Updated,
				pr.Skipped,
				errCount,
			)

			// Show errors for this project
			if errCount > 0 {
				for _, err := range pr.Errors {
					fmt.Fprintf(w, "  └─ Error: %v\n", err)
				}
			}
		}

		w.Flush()
		fmt.Println()
	}

	// Display totals
	fmt.Printf("Total: %d created, %d updated, %d skipped",
		result.TotalCreated,
		result.TotalUpdated,
		result.TotalSkipped,
	)

	if result.TotalErrors > 0 {
		fmt.Printf(", %d errors", result.TotalErrors)
	}

	fmt.Println()
	fmt.Printf("Duration: %s\n", result.Duration.Round(time.Millisecond))

	if dryRun {
		fmt.Println("\nNo changes were made (dry run)")
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else {
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

// Helper to parse int from string argument
func parseIntArg(arg string, name string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, arg)
	}
	return id, nil
}
