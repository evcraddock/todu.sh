package todoist

import (
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// mapper.go contains functions for converting between Todoist API types and Todu types.
//
// Mappings:
//   - Todoist Project → Todu Project (external_id = project ID)
//   - Todoist Task → Todu Task (external_id = task ID)
//   - Todoist is_completed → Todu Status (done/active)
//   - Todoist priority (1-4) → Todu Priority (nil/low/medium/high)
//   - Todoist labels → Todu Labels (1:1 mapping)
//   - Todoist Comments → Todu Comments (1:1 mapping)
//
// Priority Mapping (Todoist → Todu):
//   - 4 (urgent in Todoist) → high
//   - 3 (high in Todoist)   → medium
//   - 2 (medium in Todoist) → low
//   - 1 (normal in Todoist) → nil (no priority)
//
// Priority Mapping (Todu → Todoist):
//   - high   → 4
//   - medium → 3
//   - low    → 2
//   - nil    → 1

// projectToProject converts a Todoist project to a Todu project.
func projectToProject(p *Project) *types.Project {
	return &types.Project{
		ExternalID:  p.ID,
		Name:        p.Name,
		Description: nil, // Todoist projects don't have descriptions
		Status:      "active",
	}
}

// taskToTask converts a Todoist task to a Todu task.
func taskToTask(t *Task) *types.Task {
	var description *string
	if t.Description != "" {
		description = &t.Description
	}

	// Map completion status
	status := "active"
	if t.IsCompleted {
		status = "done"
	}

	// Map priority (Todoist uses 4=urgent, 3=high, 2=medium, 1=normal)
	priority := mapTodoistPriorityToTodu(t.Priority)

	// Extract labels
	labels := extractLabels(t.Labels)

	// Map due date
	var dueDate *time.Time
	if t.Due != nil {
		dueDate = parseDueDate(t.Due)
	}

	// Parse timestamps
	createdAt := parseTimestamp(t.CreatedAt)
	// Todoist REST API v2 doesn't provide updated_at for tasks
	// Use created_at as fallback
	updatedAt := createdAt

	// Source URL
	sourceURL := t.URL

	return &types.Task{
		ExternalID:  t.ID,
		SourceURL:   &sourceURL,
		Title:       t.Content,
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Labels:      labels,
		Assignees:   []types.Assignee{}, // Todoist is a personal task manager
	}
}

// mapTodoistPriorityToTodu converts Todoist priority (1-4) to Todu priority string.
// Todoist: 4=urgent, 3=high, 2=medium, 1=normal
// Todu: high, medium, low, nil
func mapTodoistPriorityToTodu(priority int) *string {
	switch priority {
	case 4:
		p := "high"
		return &p
	case 3:
		p := "medium"
		return &p
	case 2:
		p := "low"
		return &p
	default:
		return nil
	}
}

// mapToduPriorityToTodoist converts Todu priority string to Todoist priority (1-4).
func mapToduPriorityToTodoist(priority *string) int {
	if priority == nil {
		return 1 // normal
	}

	switch *priority {
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	default:
		return 1
	}
}

// extractLabels converts Todoist label strings to Todu labels.
func extractLabels(labels []string) []types.Label {
	result := make([]types.Label, 0, len(labels))
	for _, label := range labels {
		result = append(result, types.Label{
			Name: label,
		})
	}
	return result
}

// parseDueDate parses a Todoist Due object into a time.Time.
func parseDueDate(due *Due) *time.Time {
	if due == nil {
		return nil
	}

	// Try to parse datetime first (more precise)
	if due.Datetime != "" {
		if t, err := time.Parse(time.RFC3339, due.Datetime); err == nil {
			return &t
		}
	}

	// Fall back to date only
	if due.Date != "" {
		// Date can be "YYYY-MM-DD" or "YYYY-MM-DDTHH:MM:SS"
		if t, err := time.Parse("2006-01-02", due.Date); err == nil {
			return &t
		}
		if t, err := time.Parse(time.RFC3339, due.Date); err == nil {
			return &t
		}
	}

	return nil
}

// parseTimestamp parses a Todoist timestamp string.
func parseTimestamp(timestamp string) time.Time {
	if timestamp == "" {
		return time.Time{}
	}

	// Todoist uses RFC3339 format
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t
	}

	return time.Time{}
}

// commentToComment converts a Todoist comment to a Todu comment.
func commentToComment(c *Comment, defaultAuthor string) *types.Comment {
	postedAt := parseTimestamp(c.PostedAt)

	author := defaultAuthor
	if author == "" {
		author = "todoist" // fallback default
	}

	return &types.Comment{
		ExternalID: c.ID,
		Content:    c.Content,
		Author:     author,
		CreatedAt:  postedAt,
		UpdatedAt:  postedAt, // Todoist comments are immutable
	}
}

// taskCreateToRequest converts a Todu TaskCreate to a Todoist TaskCreateRequest.
func taskCreateToRequest(task *types.TaskCreate, projectID *string) *TaskCreateRequest {
	req := &TaskCreateRequest{
		Content:  task.Title,
		Priority: mapToduPriorityToTodoist(task.Priority),
	}

	if task.Description != nil {
		req.Description = *task.Description
	}

	if projectID != nil && *projectID != "" {
		req.ProjectID = *projectID
	}

	// Convert labels
	if len(task.Labels) > 0 {
		req.Labels = task.Labels
	}

	// Convert due date
	if task.DueDate != nil {
		dueStr := task.DueDate.Format("2006-01-02")
		req.DueDate = dueStr
	}

	return req
}

// taskUpdateToRequest converts a Todu TaskUpdate to a Todoist TaskUpdateRequest.
func taskUpdateToRequest(task *types.TaskUpdate) *TaskUpdateRequest {
	req := &TaskUpdateRequest{}

	if task.Title != nil {
		req.Content = task.Title
	}

	if task.Description != nil {
		req.Description = task.Description
	}

	if task.Priority != nil {
		priority := mapToduPriorityToTodoist(task.Priority)
		req.Priority = &priority
	}

	if len(task.Labels) > 0 {
		req.Labels = task.Labels
	}

	if task.DueDate != nil {
		dueStr := task.DueDate.Format("2006-01-02")
		req.DueDate = &dueStr
	}

	return req
}

// shouldClose determines if a task should be closed based on status change.
func shouldClose(status *string) bool {
	if status == nil {
		return false
	}
	return *status == "done" || *status == "cancelled"
}

// shouldReopen determines if a task should be reopened based on status change.
func shouldReopen(status *string) bool {
	if status == nil {
		return false
	}
	return *status == "active" || *status == "inprogress" || *status == "waiting"
}
