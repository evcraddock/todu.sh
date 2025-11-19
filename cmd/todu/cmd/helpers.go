package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/pkg/types"
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
