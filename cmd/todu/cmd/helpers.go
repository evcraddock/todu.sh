package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/google/uuid"
)

// parseDateToUTCStart parses a date string (YYYY-MM-DD) in local timezone
// and returns the start of that day (00:00:00) converted to UTC as RFC3339.
// This is useful for "after" filters where we want everything from the start of the local day.
func parseDateToUTCStart(dateStr string) (string, error) {
	// Parse the date in local timezone
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return "", fmt.Errorf("invalid date format %q, expected YYYY-MM-DD: %w", dateStr, err)
	}
	// Convert to UTC and format as RFC3339
	return t.UTC().Format(time.RFC3339), nil
}

// parseDateToUTCEnd parses a date string (YYYY-MM-DD) in local timezone
// and returns the end of that day (23:59:59.999999999) converted to UTC as RFC3339.
// This is useful for "before" filters where we want everything up to the end of the local day.
func parseDateToUTCEnd(dateStr string) (string, error) {
	// Parse the date in local timezone
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return "", fmt.Errorf("invalid date format %q, expected YYYY-MM-DD: %w", dateStr, err)
	}
	// Add 1 day and subtract 1 nanosecond to get end of day
	endOfDay := t.AddDate(0, 0, 1).Add(-time.Nanosecond)
	// Convert to UTC and format as RFC3339
	return endOfDay.UTC().Format(time.RFC3339), nil
}

// ensureLocalSystem checks if the "local" system exists and creates it if not.
// This allows local-only projects to be created without manual system setup.
// Returns the system ID of the local system.
func ensureLocalSystem(client *api.Client) (int, error) {
	ctx := context.Background()

	// Check if local system already exists
	systems, err := client.ListSystems(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list systems: %w", err)
	}

	for _, sys := range systems {
		if sys.Identifier == "local" {
			return sys.ID, nil
		}
	}

	// Create local system
	// Use a placeholder URL since the API requires a valid http(s) URL
	localURL := "https://local.todu.sh"
	systemCreate := &types.SystemCreate{
		Identifier: "local",
		Name:       "Local Tasks",
		URL:        &localURL,
		Metadata:   nil,
	}

	system, err := client.CreateSystem(ctx, systemCreate)
	if err != nil {
		return 0, fmt.Errorf("failed to create local system: %w", err)
	}

	fmt.Printf("Auto-registered local system (ID: %d)\n", system.ID)
	return system.ID, nil
}

// resolveSystemID resolves a system identifier (ID or name) to its numeric ID.
// Accepts either an integer string (e.g., "1") or system identifier (e.g., "github").
func resolveSystemID(client *api.Client, systemArg string) (int, error) {
	if systemArg == "" {
		return 0, nil
	}

	// Try parsing as integer first
	if id, err := strconv.Atoi(systemArg); err == nil {
		return id, nil
	}

	// Otherwise, look up by identifier
	systems, err := client.ListSystems(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to list systems: %w", err)
	}

	for _, sys := range systems {
		if sys.Identifier == systemArg {
			return sys.ID, nil
		}
	}

	return 0, fmt.Errorf("system %q not found", systemArg)
}

// loadConfig loads the configuration using the global --config flag if set
func loadConfig() (*config.Config, error) {
	return config.Load(GetConfigFile())
}

// ensureDefaultProject ensures the default project exists, creating it if needed.
// Returns the project ID for the default project.
// The project name is resolved from config (defaults.project).
// If the project doesn't exist, it's auto-created using the local system.
func ensureDefaultProject(ctx context.Context, client *api.Client, projectName string) (int, error) {
	// Check if project already exists (case-insensitive)
	projects, err := client.ListProjects(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to list projects: %w", err)
	}

	lowerName := strings.ToLower(projectName)
	for _, project := range projects {
		if strings.ToLower(project.Name) == lowerName {
			return project.ID, nil
		}
	}

	// Project doesn't exist, auto-create it using the local system
	systemID, err := ensureLocalSystem(client)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure local system: %w", err)
	}

	description := "Default project for quick task capture"
	projectCreate := &types.ProjectCreate{
		Name:         projectName,
		Description:  &description,
		SystemID:     systemID,
		ExternalID:   uuid.New().String(),
		Status:       "active",
		SyncStrategy: "bidirectional",
	}

	project, err := client.CreateProject(ctx, projectCreate)
	if err != nil {
		return 0, fmt.Errorf("failed to create default project: %w", err)
	}

	fmt.Printf("Auto-created default project %q (ID: %d)\n", project.Name, project.ID)
	return project.ID, nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// formatTimezone returns a timezone abbreviation based on UTC offset
func formatTimezone(offsetSeconds int) string {
	hours := offsetSeconds / 3600
	switch hours {
	case -5:
		return "ET" // Eastern Time (EST/EDT)
	case -6:
		return "CT" // Central Time (CST/CDT)
	case -7:
		return "MT" // Mountain Time (MST/MDT)
	case -8:
		return "PT" // Pacific Time (PST/PDT)
	case 0:
		return "UTC"
	default:
		if hours >= 0 {
			return fmt.Sprintf("UTC+%d", hours)
		}
		return fmt.Sprintf("UTC%d", hours)
	}
}
