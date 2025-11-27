package types

import "time"

// Comment represents a task comment or journal entry with full fields
type Comment struct {
	ID         int       `json:"id"`
	TaskID     *int      `json:"task_id"`               // Nullable - nil for journal entries
	ExternalID string    `json:"external_id,omitempty"` // External system's comment ID for sync
	Content    string    `json:"content"`
	Author     string    `json:"author"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CommentCreate represents data for creating a comment or journal entry
type CommentCreate struct {
	TaskID     *int   `json:"task_id,omitempty"`     // Nullable - omit for journal entries
	ExternalID string `json:"external_id,omitempty"` // Optional external system's comment ID
	Content    string `json:"content"`
	Author     string `json:"author"`
}

// CommentUpdate represents data for updating a comment
type CommentUpdate struct {
	ExternalID *string `json:"external_id,omitempty"` // Update external system's comment ID
	Content    *string `json:"content,omitempty"`
}
