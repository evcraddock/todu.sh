package types

import "time"

// Project represents a full project
type Project struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	SystemID     int        `json:"system_id"`
	ExternalID   string     `json:"external_id"`
	Status       string     `json:"status"`
	Priority     *string    `json:"priority,omitempty"` // "low", "medium", or "high"
	SyncStrategy string     `json:"sync_strategy"`      // "pull", "push", or "bidirectional"
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ProjectCreate represents data for creating a project
type ProjectCreate struct {
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	SystemID     int     `json:"system_id"`
	ExternalID   string  `json:"external_id"`
	Status       string  `json:"status"`
	Priority     *string `json:"priority,omitempty"` // "low", "medium", or "high"
	SyncStrategy string  `json:"sync_strategy"`      // "pull", "push", or "bidirectional"
}

// ProjectUpdate represents data for updating a project
type ProjectUpdate struct {
	Name         *string    `json:"name,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Status       *string    `json:"status,omitempty"`
	Priority     *string    `json:"priority,omitempty"`      // "low", "medium", or "high"
	SyncStrategy *string    `json:"sync_strategy,omitempty"` // "pull", "push", or "bidirectional"
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}
