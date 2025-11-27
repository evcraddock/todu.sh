package sync

import (
	"github.com/rs/zerolog"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// ResolveConflict determines which task should be used when both the Todu
// task and the external task have been modified.
//
// Uses last-write-wins conflict resolution based on the UpdatedAt timestamp.
// Returns the task with the most recent UpdatedAt value.
//
// Logs a warning when a conflict is detected to inform the user.
func ResolveConflict(logger zerolog.Logger, toduTask, externalTask *types.Task) *types.Task {
	// Compare timestamps
	if toduTask.UpdatedAt.After(externalTask.UpdatedAt) {
		logger.Warn().
			Str("task", toduTask.Title).
			Str("external_id", toduTask.ExternalID).
			Time("todu_updated", toduTask.UpdatedAt).
			Time("external_updated", externalTask.UpdatedAt).
			Msg("Conflict detected - using Todu version")
		return toduTask
	}

	logger.Warn().
		Str("task", externalTask.Title).
		Str("external_id", externalTask.ExternalID).
		Time("external_updated", externalTask.UpdatedAt).
		Time("todu_updated", toduTask.UpdatedAt).
		Msg("Conflict detected - using external version")
	return externalTask
}

// NeedsUpdate determines if the source task should be used to update the destination task.
// Returns true if the source task is newer than the destination task.
func NeedsUpdate(source, dest *types.Task) bool {
	return source.UpdatedAt.After(dest.UpdatedAt)
}
