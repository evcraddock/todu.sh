package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/google/uuid"
)

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
