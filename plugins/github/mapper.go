package github

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/google/go-github/v56/github"
)

// mapper.go contains functions for converting between GitHub API types and Todu types.
//
// Mappings:
//   - GitHub Repository → Todu Project (external_id = "owner/repo")
//   - GitHub Issue → Todu Task (external_id = issue number as string)
//   - GitHub Issue State + StateReason → Todu Status (bidirectional with semantic preservation)
//   - GitHub Labels → Todu Labels (priority extracted from priority:* labels)
//   - GitHub Issue Comments → Todu Comments (1:1 mapping)
//
// Status Mapping (Todu → GitHub):
//   - done       → state: "closed", state_reason: "completed"
//   - canceled  → state: "closed", state_reason: "not_planned"
//   - active     → state: "open"
//   - inprogress → state: "open"
//   - waiting    → state: "open"
//
// Status Mapping (GitHub → Todu):
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
// and to increase visibility so users are more likely to correct the labels in GitHub.

// repoToProject converts a GitHub repository to a Todu project.
func repoToProject(repo *github.Repository) *types.Project {
	externalID := fmt.Sprintf("%s/%s", repo.Owner.GetLogin(), repo.GetName())

	var description *string
	if desc := repo.GetDescription(); desc != "" {
		description = &desc
	}

	return &types.Project{
		ExternalID:  externalID,
		Name:        repo.GetName(),
		Description: description,
		Status:      "active",
	}
}

// issueToTask converts a GitHub issue to a Todu task.
func issueToTask(issue *github.Issue, repoOwner, repoName string) *types.Task {
	externalID := strconv.Itoa(issue.GetNumber())

	var description *string
	if body := issue.GetBody(); body != "" {
		description = &body
	}

	// Map status using state and state_reason
	status := mapGitHubStatusToTodu(issue.GetState(), issue.GetStateReason())

	// Extract priority from labels
	priority := extractPriority(issue.Labels)

	// Extract non-priority labels
	labels := extractLabels(issue.Labels)

	// Extract assignees
	assignees := extractAssignees(issue.Assignees)

	// Get source URL
	sourceURL := issue.GetHTMLURL()

	// Extract due date from milestone
	var dueDate *time.Time
	if milestone := issue.Milestone; milestone != nil && milestone.DueOn != nil {
		dueDate = &milestone.DueOn.Time
	}

	return &types.Task{
		ExternalID:  externalID,
		SourceURL:   &sourceURL,
		Title:       issue.GetTitle(),
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		CreatedAt:   issue.GetCreatedAt().Time,
		UpdatedAt:   issue.GetUpdatedAt().Time,
		Labels:      labels,
		Assignees:   assignees,
	}
}

// extractPriority extracts priority from GitHub labels and normalizes to valid values.
// Valid priorities are: high, medium, low
// Any other priority value is mapped to "high" to avoid accidental deprioritization.
func extractPriority(labels []*github.Label) *string {
	for _, label := range labels {
		name := strings.ToLower(label.GetName())
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
func extractLabels(labels []*github.Label) []types.Label {
	var result []types.Label
	for _, label := range labels {
		name := strings.ToLower(label.GetName())
		// Skip priority labels
		if strings.HasPrefix(name, "priority:") {
			continue
		}
		result = append(result, types.Label{
			Name: label.GetName(),
		})
	}
	return result
}

// extractAssignees extracts assignees from GitHub issue.
func extractAssignees(assignees []*github.User) []types.Assignee {
	var result []types.Assignee
	for _, assignee := range assignees {
		result = append(result, types.Assignee{
			Name: assignee.GetLogin(),
		})
	}
	return result
}

// commentToComment converts a GitHub issue comment to a Todu comment.
func commentToComment(comment *github.IssueComment) *types.Comment {
	return &types.Comment{
		ExternalID: fmt.Sprintf("%d", comment.GetID()),
		Content:    comment.GetBody(),
		Author:     comment.User.GetLogin(),
		CreatedAt:  comment.GetCreatedAt().Time,
		UpdatedAt:  comment.GetUpdatedAt().Time,
	}
}

// mapToduStatusToGitHub maps todu status to GitHub state and state_reason.
//
// Mappings:
//   - done       → state: "closed", state_reason: "completed"
//   - canceled  → state: "closed", state_reason: "not_planned"
//   - active     → state: "open", state_reason: ""
//   - inprogress → state: "open", state_reason: ""
//   - waiting    → state: "open", state_reason: ""
func mapToduStatusToGitHub(status string) (state string, stateReason string) {
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

// mapGitHubStatusToTodu maps GitHub state and state_reason to todu status.
//
// Mappings:
//   - state: "closed", state_reason: "completed"   → done
//   - state: "closed", state_reason: "not_planned" → canceled
//   - state: "closed", state_reason: ""            → done (backward compatibility)
//   - state: "open"                                → active
func mapGitHubStatusToTodu(state string, stateReason string) string {
	if state == "closed" {
		if stateReason == "not_planned" {
			return "canceled"
		}
		// Default closed issues to "done" (includes "completed" and nil for backward compatibility)
		return "done"
	}
	return "active"
}

// taskCreateToIssueRequest converts a Todu TaskCreate to a GitHub IssueRequest.
func taskCreateToIssueRequest(task *types.TaskCreate) *github.IssueRequest {
	req := &github.IssueRequest{
		Title: &task.Title,
	}

	if task.Description != nil {
		req.Body = task.Description
	}

	// Build labels including priority
	var labels []string
	if task.Priority != nil {
		labels = append(labels, fmt.Sprintf("priority:%s", *task.Priority))
	}
	labels = append(labels, task.Labels...)
	if len(labels) > 0 {
		req.Labels = &labels
	}

	// Set assignees
	if len(task.Assignees) > 0 {
		req.Assignees = &task.Assignees
	}

	return req
}

// taskUpdateToIssueRequest converts a Todu TaskUpdate to a GitHub IssueRequest.
func taskUpdateToIssueRequest(task *types.TaskUpdate) *github.IssueRequest {
	req := &github.IssueRequest{}

	if task.Title != nil {
		req.Title = task.Title
	}

	if task.Description != nil {
		req.Body = task.Description
	}

	// Handle status change
	if task.Status != nil {
		state, stateReason := mapToduStatusToGitHub(*task.Status)
		req.State = &state
		if stateReason != "" {
			req.StateReason = &stateReason
		}
	}

	// Build labels including priority
	if task.Priority != nil || len(task.Labels) > 0 {
		var labels []string
		if task.Priority != nil {
			labels = append(labels, fmt.Sprintf("priority:%s", *task.Priority))
		}
		labels = append(labels, task.Labels...)
		req.Labels = &labels
	}

	// Set assignees
	if len(task.Assignees) > 0 {
		req.Assignees = &task.Assignees
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
