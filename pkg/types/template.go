package types

import "time"

// RecurringTaskTemplate represents a recurring task template with all fields
type RecurringTaskTemplate struct {
	ID             int        `json:"id"`
	ProjectID      int        `json:"project_id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Priority       *string    `json:"priority,omitempty"`
	RecurrenceRule string     `json:"recurrence_rule"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Timezone       string     `json:"timezone"`
	TemplateType   string     `json:"template_type"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Labels         []Label    `json:"labels,omitempty"`
	Assignees      []Assignee `json:"assignees,omitempty"`
}

// RecurringTaskTemplateCreate represents data for creating a new recurring task template
type RecurringTaskTemplateCreate struct {
	ProjectID      int      `json:"project_id"`
	Title          string   `json:"title"`
	Description    *string  `json:"description,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	RecurrenceRule string   `json:"recurrence_rule"`
	StartDate      string   `json:"start_date"`
	EndDate        *string  `json:"end_date,omitempty"`
	Timezone       string   `json:"timezone"`
	TemplateType   string   `json:"template_type"`
	IsActive       bool     `json:"is_active"`
	Labels         []string `json:"labels,omitempty"`
	Assignees      []string `json:"assignees,omitempty"`
}

// RecurringTaskTemplateUpdate represents data for updating an existing recurring task template
type RecurringTaskTemplateUpdate struct {
	Title          *string  `json:"title,omitempty"`
	Description    *string  `json:"description,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	RecurrenceRule *string  `json:"recurrence_rule,omitempty"`
	EndDate        *string  `json:"end_date,omitempty"`
	Timezone       *string  `json:"timezone,omitempty"`
	IsActive       *bool    `json:"is_active,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	Assignees      []string `json:"assignees,omitempty"`
}
