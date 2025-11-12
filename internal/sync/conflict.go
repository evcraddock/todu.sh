package sync

import (
	"log"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// ResolveConflict determines which task should be used when both the Todu
// task and the external task have been modified.
//
// Uses last-write-wins conflict resolution based on the UpdatedAt timestamp.
// Returns the task with the most recent UpdatedAt value.
//
// Logs a warning when a conflict is detected to inform the user.
func ResolveConflict(toduTask, externalTask *types.Task) *types.Task {
	// Compare timestamps
	if toduTask.UpdatedAt.After(externalTask.UpdatedAt) {
		log.Printf("WARNING: Conflict detected for task %q (external_id: %s). Using Todu version (updated at %s) over external version (updated at %s)",
			toduTask.Title, toduTask.ExternalID, toduTask.UpdatedAt, externalTask.UpdatedAt)
		return toduTask
	}

	log.Printf("WARNING: Conflict detected for task %q (external_id: %s). Using external version (updated at %s) over Todu version (updated at %s)",
		externalTask.Title, externalTask.ExternalID, externalTask.UpdatedAt, toduTask.UpdatedAt)
	return externalTask
}

// NeedsUpdate determines if the source task should be used to update the destination task.
// Returns true if the source task is newer than the destination task.
func NeedsUpdate(source, dest *types.Task) bool {
	return source.UpdatedAt.After(dest.UpdatedAt)
}
