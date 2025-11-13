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
//   - GitHub Issue State (open/closed) → Todu Status (open/done)
//   - GitHub Labels → Todu Labels (priority extracted from priority:* labels)
//   - GitHub Issue Comments → Todu Comments (1:1 mapping)
//
// Priority Mapping:
//   - Labels matching "priority:high" → high priority
//   - Labels matching "priority:medium" → medium priority
//   - Labels matching "priority:low" → low priority
//   - No priority label → no priority set

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

	// Map status
	status := "active"
	if issue.GetState() == "closed" {
		status = "done"
	}

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

// extractPriority extracts priority from GitHub labels.
func extractPriority(labels []*github.Label) *string {
	for _, label := range labels {
		name := strings.ToLower(label.GetName())
		if strings.HasPrefix(name, "priority:") {
			priority := strings.TrimPrefix(name, "priority:")
			return &priority
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
		Content:   comment.GetBody(),
		Author:    comment.User.GetLogin(),
		CreatedAt: comment.GetCreatedAt().Time,
		UpdatedAt: comment.GetUpdatedAt().Time,
	}
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
	for _, label := range task.Labels {
		labels = append(labels, label.Name)
	}
	if len(labels) > 0 {
		req.Labels = &labels
	}

	// Set assignees
	if len(task.Assignees) > 0 {
		assignees := make([]string, len(task.Assignees))
		for i, assignee := range task.Assignees {
			assignees[i] = assignee.Name
		}
		req.Assignees = &assignees
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
		state := "open"
		if *task.Status == "done" || *task.Status == "cancelled" {
			state = "closed"
		}
		req.State = &state
	}

	// Build labels including priority
	if task.Priority != nil || len(task.Labels) > 0 {
		var labels []string
		if task.Priority != nil {
			labels = append(labels, fmt.Sprintf("priority:%s", *task.Priority))
		}
		for _, label := range task.Labels {
			labels = append(labels, label.Name)
		}
		req.Labels = &labels
	}

	// Set assignees
	if len(task.Assignees) > 0 {
		assignees := make([]string, len(task.Assignees))
		for i, assignee := range task.Assignees {
			assignees[i] = assignee.Name
		}
		req.Assignees = &assignees
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
