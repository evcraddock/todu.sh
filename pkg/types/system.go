package types

import "time"

// System represents an external task management system
type System struct {
	ID         int               `json:"id"`
	Identifier string            `json:"identifier"`
	Name       string            `json:"name"`
	URL        *string           `json:"url,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// SystemCreate represents data for creating a system
type SystemCreate struct {
	Identifier string            `json:"identifier"`
	Name       string            `json:"name"`
	URL        *string           `json:"url,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// SystemUpdate represents data for updating a system
type SystemUpdate struct {
	Name     *string           `json:"name,omitempty"`
	URL      *string           `json:"url,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
