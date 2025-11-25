package forgejo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// mapper.go contains functions for converting between Forgejo API types and Todu types.
//
// Mappings:
//   - Forgejo Repository → Todu Project (external_id = "owner/repo")
//   - Forgejo Issue → Todu Task (external_id = issue number as string)
//   - Forgejo Issue State + StateReason → Todu Status (bidirectional with semantic preservation)
//   - Forgejo Labels → Todu Labels (priority extracted from priority:* labels)
//   - Forgejo Issue Comments → Todu Comments (1:1 mapping)
//
// Status Mapping (Todu → Forgejo):
//   - done       → state: "closed", state_reason: "completed"
//   - canceled  → state: "closed", state_reason: "not_planned"
//   - active     → state: "open"
//   - inprogress → state: "open"
//   - waiting    → state: "open"
//
// Status Mapping (Forgejo → Todu):
//   - state: "closed", state_reason: "completed"   → done
//   - state: "closed", state_reason: "not_planned" → canceled
//   - state: "closed", state_reason: null/other    → done (backward compatibility)
//   - state: "open"                                → active
//
// Priority Mapping:
//   - Labels matching "priority:high" → high priority
//   - Labels matching "priority:medium" → medium priority
//   - Labels matching "priority:low" → low priority
//   - Labels matching "priority:*" (invalid values) → high priority (e.g., urgent, critical, p0)
//   - No priority label → no priority set
//
// Note: Invalid priority values are normalized to "high" to prevent accidental deprioritization
// and to increase visibility so users are more likely to correct the labels in Forgejo.

// repoToProject converts a Forgejo repository to a Todu project.
func repoToProject(repo *Repository) *types.Project {
	externalID := repo.FullName
	if externalID == "" && repo.Owner != nil {
		externalID = fmt.Sprintf("%s/%s", repo.Owner.Login, repo.Name)
	}

	var description *string
	if repo.Description != "" {
		description = &repo.Description
	}

	return &types.Project{
		ExternalID:  externalID,
		Name:        repo.Name,
		Description: description,
		Status:      "active",
	}
}

// issueToTask converts a Forgejo issue to a Todu task.
func issueToTask(issue *Issue, repoOwner, repoName string) *types.Task {
	externalID := strconv.Itoa(issue.Number)

	var description *string
	if issue.Body != "" {
		description = &issue.Body
	}

	// Map status using state and state_reason
	status := mapForgejoStatusToTodu(issue.State, issue.StateReason)

	// Extract priority from labels
	priority := extractPriority(issue.Labels)

	// Extract non-priority labels
	labels := extractLabels(issue.Labels)

	// Extract assignees
	assignees := extractAssignees(issue.Assignees)

	// Get source URL
	sourceURL := issue.HTMLURL

	return &types.Task{
		ExternalID:  externalID,
		SourceURL:   &sourceURL,
		Title:       issue.Title,
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     nil, // Forgejo doesn't have milestones with due dates in the same way
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
		Labels:      labels,
		Assignees:   assignees,
	}
}

// extractPriority extracts priority from Forgejo labels and normalizes to valid values.
// Valid priorities are: high, medium, low
// Any other priority value is mapped to "high" to avoid accidental deprioritization.
func extractPriority(labels []*Label) *string {
	for _, label := range labels {
		name := strings.ToLower(label.Name)
		if strings.HasPrefix(name, "priority:") {
			priority := strings.TrimPrefix(name, "priority:")

			// Normalize to valid priority values
			switch priority {
			case "high", "medium", "low":
				return &priority
			default:
				// Map any invalid priority to "high" to avoid deprioritization
				// This includes values like: urgent, critical, p0, p1, etc.
				high := "high"
				return &high
			}
		}
	}
	return nil
}

// extractLabels extracts non-priority labels.
func extractLabels(labels []*Label) []types.Label {
	var result []types.Label
	for _, label := range labels {
		name := strings.ToLower(label.Name)
		// Skip priority labels
		if strings.HasPrefix(name, "priority:") {
			continue
		}
		result = append(result, types.Label{
			Name: label.Name,
		})
	}
	return result
}

// extractAssignees extracts assignees from Forgejo issue.
func extractAssignees(assignees []*User) []types.Assignee {
	var result []types.Assignee
	for _, assignee := range assignees {
		result = append(result, types.Assignee{
			Name: assignee.Login,
		})
	}
	return result
}

// commentToComment converts a Forgejo issue comment to a Todu comment.
func commentToComment(comment *Comment) *types.Comment {
	author := ""
	if comment.User != nil {
		author = comment.User.Login
	}

	return &types.Comment{
		ExternalID: fmt.Sprintf("%d", comment.ID),
		Content:    comment.Body,
		Author:     author,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
	}
}

// mapToduStatusToForgejo maps todu status to Forgejo state and state_reason.
//
// Mappings:
//   - done       → state: "closed", state_reason: "completed"
//   - canceled  → state: "closed", state_reason: "not_planned"
//   - active     → state: "open", state_reason: ""
//   - inprogress → state: "open", state_reason: ""
//   - waiting    → state: "open", state_reason: ""
func mapToduStatusToForgejo(status string) (state string, stateReason string) {
	switch status {
	case "done":
		return "closed", "completed"
	case "canceled":
		return "closed", "not_planned"
	default:
		// active, inprogress, waiting, and any other status map to open
		return "open", ""
	}
}

// mapForgejoStatusToTodu maps Forgejo state and state_reason to todu status.
//
// Mappings:
//   - state: "closed", state_reason: "completed"   → done
//   - state: "closed", state_reason: "not_planned" → canceled
//   - state: "closed", state_reason: ""            → done (backward compatibility)
//   - state: "open"                                → active
func mapForgejoStatusToTodu(state string, stateReason string) string {
	if state == "closed" {
		if stateReason == "not_planned" {
			return "canceled"
		}
		// Default closed issues to "done" (includes "completed" and nil for backward compatibility)
		return "done"
	}
	return "active"
}

// taskCreateToIssueRequest converts a Todu TaskCreate to a Forgejo CreateIssueRequest.
// Note: labelIDs must be pre-resolved using client.resolveLabelIDs.
func taskCreateToIssueRequest(task *types.TaskCreate, labelIDs []int64) *CreateIssueRequest {
	req := &CreateIssueRequest{
		Title:  task.Title,
		Labels: labelIDs,
	}

	if task.Description != nil {
		req.Body = *task.Description
	}

	// Set first assignee (Forgejo API only accepts one assignee in create request)
	if len(task.Assignees) > 0 {
		req.Assignee = task.Assignees[0]
	}

	return req
}

// taskUpdateToIssueRequest converts a Todu TaskUpdate to a Forgejo UpdateIssueRequest.
// Note: Labels are handled separately via updateIssueLabels.
func taskUpdateToIssueRequest(task *types.TaskUpdate) *UpdateIssueRequest {
	req := &UpdateIssueRequest{}

	if task.Title != nil {
		req.Title = task.Title
	}

	if task.Description != nil {
		req.Body = task.Description
	}

	// Handle status change
	if task.Status != nil {
		state, _ := mapToduStatusToForgejo(*task.Status)
		req.State = &state
		// Note: Forgejo may not support state_reason in the same way as GitHub
		// The state change should be sufficient
	}

	return req
}

// parseRepoExternalID parses "owner/repo" format into owner and repo.
func parseRepoExternalID(externalID string) (owner, repo string, err error) {
	parts := strings.Split(externalID, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository external_id format: expected 'owner/repo', got %q", externalID)
	}
	return parts[0], parts[1], nil
}

// buildLabelsWithPriority creates a list of label names including priority.
// This is used before resolving to IDs.
func buildLabelsWithPriority(labels []string, priority *string) []string {
	var result []string
	if priority != nil {
		result = append(result, fmt.Sprintf("priority:%s", *priority))
	}
	result = append(result, labels...)
	return result
}
