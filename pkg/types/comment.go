package types

import "time"

// Comment represents a task comment with full fields
type Comment struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentCreate represents data for creating a comment
type CommentCreate struct {
	TaskID  int    `json:"task_id"`
	Content string `json:"content"`
	Author  string `json:"author"`
}
