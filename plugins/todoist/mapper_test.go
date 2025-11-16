package todoist

import (
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// TestMapTodoistPriorityToTodu tests the mapping from Todoist priority to Todu priority.
func TestMapTodoistPriorityToTodu(t *testing.T) {
	tests := []struct {
		name             string
		todoistPriority  int
		expectedPriority *string
	}{
		{
			name:             "urgent (4) maps to high",
			todoistPriority:  4,
			expectedPriority: strPtr("high"),
		},
		{
			name:             "high (3) maps to medium",
			todoistPriority:  3,
			expectedPriority: strPtr("medium"),
		},
		{
			name:             "medium (2) maps to low",
			todoistPriority:  2,
			expectedPriority: strPtr("low"),
		},
		{
			name:             "normal (1) maps to nil",
			todoistPriority:  1,
			expectedPriority: nil,
		},
		{
			name:             "zero maps to nil",
			todoistPriority:  0,
			expectedPriority: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := mapTodoistPriorityToTodu(tt.todoistPriority)

			if tt.expectedPriority == nil {
				if priority != nil {
					t.Errorf("Expected nil priority, got %s", *priority)
				}
			} else {
				if priority == nil {
					t.Fatalf("Expected priority %s, got nil", *tt.expectedPriority)
				}
				if *priority != *tt.expectedPriority {
					t.Errorf("Expected priority %s, got %s", *tt.expectedPriority, *priority)
				}
			}
		})
	}
}

// TestMapToduPriorityToTodoist tests the mapping from Todu priority to Todoist priority.
func TestMapToduPriorityToTodoist(t *testing.T) {
	tests := []struct {
		name             string
		toduPriority     *string
		expectedPriority int
	}{
		{
			name:             "high maps to 4",
			toduPriority:     strPtr("high"),
			expectedPriority: 4,
		},
		{
			name:             "medium maps to 3",
			toduPriority:     strPtr("medium"),
			expectedPriority: 3,
		},
		{
			name:             "low maps to 2",
			toduPriority:     strPtr("low"),
			expectedPriority: 2,
		},
		{
			name:             "nil maps to 1",
			toduPriority:     nil,
			expectedPriority: 1,
		},
		{
			name:             "unknown maps to 1",
			toduPriority:     strPtr("unknown"),
			expectedPriority: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := mapToduPriorityToTodoist(tt.toduPriority)

			if priority != tt.expectedPriority {
				t.Errorf("Expected priority %d, got %d", tt.expectedPriority, priority)
			}
		})
	}
}

// TestTaskToTask tests the conversion of Todoist task to Todu task.
func TestTaskToTask(t *testing.T) {
	tests := []struct {
		name           string
		todoistTask    *Task
		expectedStatus string
		expectedPriority *string
	}{
		{
			name: "completed task maps to done",
			todoistTask: &Task{
				ID:          "123",
				ProjectID:   "456",
				Content:     "Test Task",
				Description: "Test description",
				IsCompleted: true,
				Priority:    1,
				Labels:      []string{},
				CreatedAt:   "2025-01-01T10:00:00Z",
				URL:         "https://todoist.com/app/task/123",
			},
			expectedStatus:   "done",
			expectedPriority: nil,
		},
		{
			name: "active task with high priority",
			todoistTask: &Task{
				ID:          "124",
				ProjectID:   "456",
				Content:     "Important Task",
				Description: "",
				IsCompleted: false,
				Priority:    4,
				Labels:      []string{"work", "urgent"},
				CreatedAt:   "2025-01-01T10:00:00Z",
				URL:         "https://todoist.com/app/task/124",
			},
			expectedStatus:   "active",
			expectedPriority: strPtr("high"),
		},
		{
			name: "task with due date",
			todoistTask: &Task{
				ID:          "125",
				ProjectID:   "456",
				Content:     "Task with due",
				IsCompleted: false,
				Priority:    2,
				Labels:      []string{},
				CreatedAt:   "2025-01-01T10:00:00Z",
				URL:         "https://todoist.com/app/task/125",
				Due: &Due{
					Date:   "2025-12-31",
					String: "Dec 31",
				},
			},
			expectedStatus:   "active",
			expectedPriority: strPtr("low"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := taskToTask(tt.todoistTask)

			if task.ExternalID != tt.todoistTask.ID {
				t.Errorf("Expected external_id %s, got %s", tt.todoistTask.ID, task.ExternalID)
			}

			if task.Title != tt.todoistTask.Content {
				t.Errorf("Expected title %s, got %s", tt.todoistTask.Content, task.Title)
			}

			if task.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, task.Status)
			}

			if tt.expectedPriority == nil {
				if task.Priority != nil {
					t.Errorf("Expected nil priority, got %s", *task.Priority)
				}
			} else {
				if task.Priority == nil {
					t.Errorf("Expected priority %s, got nil", *tt.expectedPriority)
				} else if *task.Priority != *tt.expectedPriority {
					t.Errorf("Expected priority %s, got %s", *tt.expectedPriority, *task.Priority)
				}
			}

			// Check source URL
			if task.SourceURL == nil || *task.SourceURL != tt.todoistTask.URL {
				t.Errorf("Expected source_url %s, got %v", tt.todoistTask.URL, task.SourceURL)
			}

			// Check labels
			if len(task.Labels) != len(tt.todoistTask.Labels) {
				t.Errorf("Expected %d labels, got %d", len(tt.todoistTask.Labels), len(task.Labels))
			}

			// Assignees should always be empty
			if len(task.Assignees) != 0 {
				t.Errorf("Expected empty assignees, got %d", len(task.Assignees))
			}
		})
	}
}

// TestProjectToProject tests the conversion of Todoist project to Todu project.
func TestProjectToProject(t *testing.T) {
	project := &Project{
		ID:             "abc123",
		Name:           "My Project",
		Color:          "red",
		IsInboxProject: false,
		IsFavorite:     true,
	}

	toduProject := projectToProject(project)

	if toduProject.ExternalID != project.ID {
		t.Errorf("Expected external_id %s, got %s", project.ID, toduProject.ExternalID)
	}

	if toduProject.Name != project.Name {
		t.Errorf("Expected name %s, got %s", project.Name, toduProject.Name)
	}

	if toduProject.Description != nil {
		t.Errorf("Expected nil description, got %v", toduProject.Description)
	}

	if toduProject.Status != "active" {
		t.Errorf("Expected status active, got %s", toduProject.Status)
	}
}

// TestParseDueDate tests parsing of Todoist due dates.
func TestParseDueDate(t *testing.T) {
	tests := []struct {
		name        string
		due         *Due
		expectNil   bool
		expectedDay int
	}{
		{
			name:      "nil due returns nil",
			due:       nil,
			expectNil: true,
		},
		{
			name: "date only",
			due: &Due{
				Date: "2025-12-25",
			},
			expectNil:   false,
			expectedDay: 25,
		},
		{
			name: "datetime format",
			due: &Due{
				Datetime: "2025-12-25T14:30:00Z",
			},
			expectNil:   false,
			expectedDay: 25,
		},
		{
			name: "prefers datetime over date",
			due: &Due{
				Date:     "2025-12-24",
				Datetime: "2025-12-25T14:30:00Z",
			},
			expectNil:   false,
			expectedDay: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDueDate(tt.due)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected non-nil date")
				}
				if result.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.Day())
				}
			}
		})
	}
}

// TestParseTimestamp tests parsing of Todoist timestamps.
func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expectZero bool
	}{
		{
			name:       "empty string returns zero time",
			timestamp:  "",
			expectZero: true,
		},
		{
			name:       "RFC3339 format",
			timestamp:  "2025-01-15T10:30:00Z",
			expectZero: false,
		},
		{
			name:       "RFC3339 with timezone offset",
			timestamp:  "2025-01-15T10:30:00+05:00",
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimestamp(tt.timestamp)

			if tt.expectZero {
				if !result.IsZero() {
					t.Errorf("Expected zero time, got %v", result)
				}
			} else {
				if result.IsZero() {
					t.Error("Expected non-zero time")
				}
			}
		})
	}
}

// TestCommentToComment tests the conversion of Todoist comment to Todu comment.
func TestCommentToComment(t *testing.T) {
	comment := &Comment{
		ID:       "comment123",
		TaskID:   "task456",
		Content:  "This is a test comment",
		PostedAt: "2025-01-01T12:00:00Z",
	}

	tests := []struct {
		name           string
		defaultAuthor  string
		expectedAuthor string
	}{
		{
			name:           "custom author",
			defaultAuthor:  "erik",
			expectedAuthor: "erik",
		},
		{
			name:           "default fallback",
			defaultAuthor:  "",
			expectedAuthor: "todoist",
		},
		{
			name:           "system default",
			defaultAuthor:  "todoist",
			expectedAuthor: "todoist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toduComment := commentToComment(comment, tt.defaultAuthor)

			if toduComment.ExternalID != comment.ID {
				t.Errorf("Expected external_id %s, got %s", comment.ID, toduComment.ExternalID)
			}

			if toduComment.Content != comment.Content {
				t.Errorf("Expected content %s, got %s", comment.Content, toduComment.Content)
			}

			if toduComment.Author != tt.expectedAuthor {
				t.Errorf("Expected author '%s', got %s", tt.expectedAuthor, toduComment.Author)
			}

			if toduComment.CreatedAt.IsZero() {
				t.Error("Expected non-zero created_at")
			}
		})
	}
}

// TestTaskCreateToRequest tests the conversion of Todu TaskCreate to Todoist request.
func TestTaskCreateToRequest(t *testing.T) {
	description := "Test description"
	priority := "high"
	dueDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	taskCreate := &types.TaskCreate{
		Title:       "New Task",
		Description: &description,
		Priority:    &priority,
		Labels:      []string{"work", "important"},
		DueDate:     &dueDate,
	}

	projectID := "project123"
	req := taskCreateToRequest(taskCreate, &projectID)

	if req.Content != taskCreate.Title {
		t.Errorf("Expected content %s, got %s", taskCreate.Title, req.Content)
	}

	if req.Description != description {
		t.Errorf("Expected description %s, got %s", description, req.Description)
	}

	if req.Priority != 4 {
		t.Errorf("Expected priority 4, got %d", req.Priority)
	}

	if req.ProjectID != projectID {
		t.Errorf("Expected project_id %s, got %s", projectID, req.ProjectID)
	}

	if len(req.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(req.Labels))
	}

	if req.DueDate != "2025-12-31" {
		t.Errorf("Expected due_date 2025-12-31, got %s", req.DueDate)
	}
}

// TestTaskUpdateToRequest tests the conversion of Todu TaskUpdate to Todoist request.
func TestTaskUpdateToRequest(t *testing.T) {
	title := "Updated Title"
	description := "Updated description"
	priority := "medium"

	taskUpdate := &types.TaskUpdate{
		Title:       &title,
		Description: &description,
		Priority:    &priority,
		Labels:      []string{"updated"},
	}

	req := taskUpdateToRequest(taskUpdate)

	if req.Content == nil || *req.Content != title {
		t.Errorf("Expected content %s, got %v", title, req.Content)
	}

	if req.Description == nil || *req.Description != description {
		t.Errorf("Expected description %s, got %v", description, req.Description)
	}

	if req.Priority == nil || *req.Priority != 3 {
		t.Errorf("Expected priority 3, got %v", req.Priority)
	}

	if len(req.Labels) != 1 || req.Labels[0] != "updated" {
		t.Errorf("Expected labels [updated], got %v", req.Labels)
	}
}

// TestShouldClose tests the logic for determining if a task should be closed.
func TestShouldClose(t *testing.T) {
	tests := []struct {
		name     string
		status   *string
		expected bool
	}{
		{
			name:     "done should close",
			status:   strPtr("done"),
			expected: true,
		},
		{
			name:     "cancelled should close",
			status:   strPtr("cancelled"),
			expected: true,
		},
		{
			name:     "active should not close",
			status:   strPtr("active"),
			expected: false,
		},
		{
			name:     "inprogress should not close",
			status:   strPtr("inprogress"),
			expected: false,
		},
		{
			name:     "nil should not close",
			status:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldClose(tt.status)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestShouldReopen tests the logic for determining if a task should be reopened.
func TestShouldReopen(t *testing.T) {
	tests := []struct {
		name     string
		status   *string
		expected bool
	}{
		{
			name:     "active should reopen",
			status:   strPtr("active"),
			expected: true,
		},
		{
			name:     "inprogress should reopen",
			status:   strPtr("inprogress"),
			expected: true,
		},
		{
			name:     "waiting should reopen",
			status:   strPtr("waiting"),
			expected: true,
		},
		{
			name:     "done should not reopen",
			status:   strPtr("done"),
			expected: false,
		},
		{
			name:     "nil should not reopen",
			status:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldReopen(tt.status)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestExtractLabels tests the extraction of labels from Todoist.
func TestExtractLabels(t *testing.T) {
	labels := []string{"work", "personal", "urgent"}

	result := extractLabels(labels)

	if len(result) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(result))
	}

	for i, label := range labels {
		if result[i].Name != label {
			t.Errorf("Expected label %s at index %d, got %s", label, i, result[i].Name)
		}
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
