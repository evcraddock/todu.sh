package types

import "time"

// Task represents a full task with all fields
type Task struct {
	ID          int        `json:"id"`
	ExternalID  string     `json:"external_id"`
	SourceURL   *string    `json:"source_url,omitempty"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	ProjectID   int        `json:"project_id"`
	Status      string     `json:"status"`
	Priority    *string    `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Labels      []Label    `json:"labels,omitempty"`
	Assignees   []Assignee `json:"assignees,omitempty"`
}

// TaskCreate represents data for creating a new task
type TaskCreate struct {
	ExternalID  string     `json:"external_id"`
	SourceURL   *string    `json:"source_url,omitempty"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	ProjectID   int        `json:"project_id"`
	Status      string     `json:"status"`
	Priority    *string    `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Labels      []Label    `json:"labels,omitempty"`
	Assignees   []Assignee `json:"assignees,omitempty"`
}

// TaskUpdate represents data for updating an existing task
type TaskUpdate struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Labels      []Label    `json:"labels,omitempty"`
	Assignees   []Assignee `json:"assignees,omitempty"`
}
