package types

import "time"

// Project represents a full project
type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	SystemID    int       `json:"system_id"`
	ExternalID  string    `json:"external_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectCreate represents data for creating a project
type ProjectCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	SystemID    int     `json:"system_id"`
	ExternalID  string  `json:"external_id"`
	Status      string  `json:"status"`
}

// ProjectUpdate represents data for updating a project
type ProjectUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}
