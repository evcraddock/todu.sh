package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/evcraddock/todu.sh/internal/api"
)

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
