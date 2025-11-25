package forgejo

import (
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// TestMapToduStatusToForgejo tests the mapping from todu status to Forgejo state and state_reason.
func TestMapToduStatusToForgejo(t *testing.T) {
	tests := []struct {
		name                string
		toduStatus          string
		expectedState       string
		expectedStateReason string
	}{
		{
			name:                "done maps to closed+completed",
			toduStatus:          "done",
			expectedState:       "closed",
			expectedStateReason: "completed",
		},
		{
			name:                "canceled maps to closed+not_planned",
			toduStatus:          "canceled",
			expectedState:       "closed",
			expectedStateReason: "not_planned",
		},
		{
			name:                "active maps to open",
			toduStatus:          "active",
			expectedState:       "open",
			expectedStateReason: "",
		},
		{
			name:                "inprogress maps to open",
			toduStatus:          "inprogress",
			expectedState:       "open",
			expectedStateReason: "",
		},
		{
			name:                "waiting maps to open",
			toduStatus:          "waiting",
			expectedState:       "open",
			expectedStateReason: "",
		},
		{
			name:                "unknown status maps to open",
			toduStatus:          "some-unknown-status",
			expectedState:       "open",
			expectedStateReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, stateReason := mapToduStatusToForgejo(tt.toduStatus)
			if state != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, state)
			}
			if stateReason != tt.expectedStateReason {
				t.Errorf("Expected stateReason %s, got %s", tt.expectedStateReason, stateReason)
			}
		})
	}
}

// TestMapForgejoStatusToTodu tests the mapping from Forgejo state and state_reason to todu status.
func TestMapForgejoStatusToTodu(t *testing.T) {
	tests := []struct {
		name               string
		forgejoState       string
		forgejoStateReason string
		expectedStatus     string
	}{
		{
			name:               "closed+completed maps to done",
			forgejoState:       "closed",
			forgejoStateReason: "completed",
			expectedStatus:     "done",
		},
		{
			name:               "closed+not_planned maps to canceled",
			forgejoState:       "closed",
			forgejoStateReason: "not_planned",
			expectedStatus:     "canceled",
		},
		{
			name:               "closed with no state_reason maps to done (backward compatibility)",
			forgejoState:       "closed",
			forgejoStateReason: "",
			expectedStatus:     "done",
		},
		{
			name:               "closed+reopened maps to done",
			forgejoState:       "closed",
			forgejoStateReason: "reopened",
			expectedStatus:     "done",
		},
		{
			name:               "open maps to active",
			forgejoState:       "open",
			forgejoStateReason: "",
			expectedStatus:     "active",
		},
		{
			name:               "open with any state_reason maps to active",
			forgejoState:       "open",
			forgejoStateReason: "reopened",
			expectedStatus:     "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := mapForgejoStatusToTodu(tt.forgejoState, tt.forgejoStateReason)
			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

// TestIssueToTask_StatusMapping tests that issueToTask correctly maps Forgejo state_reason.
func TestIssueToTask_StatusMapping(t *testing.T) {
	now := time.Now()
	owner := "test-owner"
	repo := "test-repo"

	tests := []struct {
		name               string
		forgejoState       string
		forgejoStateReason string
		expectedStatus     string
	}{
		{
			name:               "closed issue with completed reason",
			forgejoState:       "closed",
			forgejoStateReason: "completed",
			expectedStatus:     "done",
		},
		{
			name:               "closed issue with not_planned reason",
			forgejoState:       "closed",
			forgejoStateReason: "not_planned",
			expectedStatus:     "canceled",
		},
		{
			name:               "closed issue without state_reason (backward compatibility)",
			forgejoState:       "closed",
			forgejoStateReason: "",
			expectedStatus:     "done",
		},
		{
			name:               "open issue",
			forgejoState:       "open",
			forgejoStateReason: "",
			expectedStatus:     "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{
				Number:      1,
				Title:       "Test Issue",
				Body:        "Test body",
				State:       tt.forgejoState,
				StateReason: tt.forgejoStateReason,
				HTMLURL:     "https://forgejo.example.com/test/test/issues/1",
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			task := issueToTask(issue, owner, repo)

			if task.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, task.Status)
			}
		})
	}
}

// TestTaskUpdateToIssueRequest_StatusMapping tests that taskUpdateToIssueRequest sets state.
func TestTaskUpdateToIssueRequest_StatusMapping(t *testing.T) {
	tests := []struct {
		name          string
		toduStatus    string
		expectedState string
	}{
		{
			name:          "done status sets closed state",
			toduStatus:    "done",
			expectedState: "closed",
		},
		{
			name:          "canceled status sets closed state",
			toduStatus:    "canceled",
			expectedState: "closed",
		},
		{
			name:          "active status sets open state",
			toduStatus:    "active",
			expectedState: "open",
		},
		{
			name:          "inprogress status sets open state",
			toduStatus:    "inprogress",
			expectedState: "open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskUpdate := &types.TaskUpdate{
				Status: &tt.toduStatus,
			}

			req := taskUpdateToIssueRequest(taskUpdate)

			if req.State == nil {
				t.Fatal("Expected State to be set")
			}
			if *req.State != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, *req.State)
			}
		})
	}
}

// TestTaskUpdateToIssueRequest_NoStatusChange tests that state is not set when status is nil.
func TestTaskUpdateToIssueRequest_NoStatusChange(t *testing.T) {
	title := "Updated Title"
	taskUpdate := &types.TaskUpdate{
		Title:  &title,
		Status: nil,
	}

	req := taskUpdateToIssueRequest(taskUpdate)

	if req.State != nil {
		t.Errorf("Expected State to be nil when status is not updated, got %s", *req.State)
	}
}

// TestExtractPriority tests priority extraction and normalization from Forgejo labels.
func TestExtractPriority(t *testing.T) {
	tests := []struct {
		name             string
		labels           []*Label
		expectedPriority *string
	}{
		{
			name: "valid high priority",
			labels: []*Label{
				{Name: "priority:high"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "valid medium priority",
			labels: []*Label{
				{Name: "priority:medium"},
			},
			expectedPriority: strPtr("medium"),
		},
		{
			name: "valid low priority",
			labels: []*Label{
				{Name: "priority:low"},
			},
			expectedPriority: strPtr("low"),
		},
		{
			name: "invalid urgent priority maps to high",
			labels: []*Label{
				{Name: "priority:urgent"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "invalid critical priority maps to high",
			labels: []*Label{
				{Name: "priority:critical"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "invalid p0 priority maps to high",
			labels: []*Label{
				{Name: "priority:p0"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "invalid blocker priority maps to high",
			labels: []*Label{
				{Name: "priority:blocker"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "no priority label returns nil",
			labels: []*Label{
				{Name: "bug"},
				{Name: "enhancement"},
			},
			expectedPriority: nil,
		},
		{
			name: "case insensitive - uppercase HIGH",
			labels: []*Label{
				{Name: "priority:HIGH"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "case insensitive - uppercase URGENT maps to high",
			labels: []*Label{
				{Name: "priority:URGENT"},
			},
			expectedPriority: strPtr("high"),
		},
		{
			name: "mixed case - MixedCase priority maps to medium",
			labels: []*Label{
				{Name: "Priority:Medium"},
			},
			expectedPriority: strPtr("medium"),
		},
		{
			name: "priority label among other labels",
			labels: []*Label{
				{Name: "bug"},
				{Name: "priority:low"},
				{Name: "needs-review"},
			},
			expectedPriority: strPtr("low"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := extractPriority(tt.labels)

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

// TestExtractLabels tests that non-priority labels are extracted correctly.
func TestExtractLabels(t *testing.T) {
	tests := []struct {
		name           string
		labels         []*Label
		expectedLabels []string
	}{
		{
			name: "filters out priority labels",
			labels: []*Label{
				{Name: "bug"},
				{Name: "priority:high"},
				{Name: "enhancement"},
			},
			expectedLabels: []string{"bug", "enhancement"},
		},
		{
			name: "returns all non-priority labels",
			labels: []*Label{
				{Name: "bug"},
				{Name: "documentation"},
			},
			expectedLabels: []string{"bug", "documentation"},
		},
		{
			name:           "empty labels returns empty slice",
			labels:         []*Label{},
			expectedLabels: []string{},
		},
		{
			name: "only priority labels returns empty slice",
			labels: []*Label{
				{Name: "priority:high"},
			},
			expectedLabels: []string{},
		},
		{
			name: "case insensitive priority filtering",
			labels: []*Label{
				{Name: "Priority:HIGH"},
				{Name: "bug"},
			},
			expectedLabels: []string{"bug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := extractLabels(tt.labels)

			if len(labels) != len(tt.expectedLabels) {
				t.Fatalf("Expected %d labels, got %d", len(tt.expectedLabels), len(labels))
			}

			for i, label := range labels {
				if label.Name != tt.expectedLabels[i] {
					t.Errorf("Expected label %s at position %d, got %s", tt.expectedLabels[i], i, label.Name)
				}
			}
		})
	}
}

// TestExtractAssignees tests that assignees are extracted correctly.
func TestExtractAssignees(t *testing.T) {
	tests := []struct {
		name              string
		assignees         []*User
		expectedAssignees []string
	}{
		{
			name: "single assignee",
			assignees: []*User{
				{Login: "user1"},
			},
			expectedAssignees: []string{"user1"},
		},
		{
			name: "multiple assignees",
			assignees: []*User{
				{Login: "user1"},
				{Login: "user2"},
				{Login: "user3"},
			},
			expectedAssignees: []string{"user1", "user2", "user3"},
		},
		{
			name:              "no assignees",
			assignees:         []*User{},
			expectedAssignees: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignees := extractAssignees(tt.assignees)

			if len(assignees) != len(tt.expectedAssignees) {
				t.Fatalf("Expected %d assignees, got %d", len(tt.expectedAssignees), len(assignees))
			}

			for i, assignee := range assignees {
				if assignee.Name != tt.expectedAssignees[i] {
					t.Errorf("Expected assignee %s at position %d, got %s", tt.expectedAssignees[i], i, assignee.Name)
				}
			}
		})
	}
}

// TestParseRepoExternalID tests parsing of "owner/repo" format.
func TestParseRepoExternalID(t *testing.T) {
	tests := []struct {
		name          string
		externalID    string
		expectedOwner string
		expectedRepo  string
		expectError   bool
	}{
		{
			name:          "valid owner/repo",
			externalID:    "myowner/myrepo",
			expectedOwner: "myowner",
			expectedRepo:  "myrepo",
			expectError:   false,
		},
		{
			name:          "owner/repo with dashes",
			externalID:    "my-owner/my-repo",
			expectedOwner: "my-owner",
			expectedRepo:  "my-repo",
			expectError:   false,
		},
		{
			name:        "missing slash",
			externalID:  "invalidformat",
			expectError: true,
		},
		{
			name:        "too many slashes",
			externalID:  "owner/repo/extra",
			expectError: true,
		},
		{
			name:        "empty string",
			externalID:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRepoExternalID(tt.externalID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if owner != tt.expectedOwner {
					t.Errorf("Expected owner %s, got %s", tt.expectedOwner, owner)
				}
				if repo != tt.expectedRepo {
					t.Errorf("Expected repo %s, got %s", tt.expectedRepo, repo)
				}
			}
		})
	}
}

// TestBuildLabelsWithPriority tests building label list with priority included.
func TestBuildLabelsWithPriority(t *testing.T) {
	tests := []struct {
		name           string
		labels         []string
		priority       *string
		expectedLabels []string
	}{
		{
			name:           "priority only",
			labels:         []string{},
			priority:       strPtr("high"),
			expectedLabels: []string{"priority:high"},
		},
		{
			name:           "labels only",
			labels:         []string{"bug", "enhancement"},
			priority:       nil,
			expectedLabels: []string{"bug", "enhancement"},
		},
		{
			name:           "both priority and labels",
			labels:         []string{"bug", "enhancement"},
			priority:       strPtr("medium"),
			expectedLabels: []string{"priority:medium", "bug", "enhancement"},
		},
		{
			name:           "no priority no labels",
			labels:         []string{},
			priority:       nil,
			expectedLabels: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildLabelsWithPriority(tt.labels, tt.priority)

			if len(result) != len(tt.expectedLabels) {
				t.Fatalf("Expected %d labels, got %d: %v", len(tt.expectedLabels), len(result), result)
			}

			for i, label := range result {
				if label != tt.expectedLabels[i] {
					t.Errorf("Expected label %s at position %d, got %s", tt.expectedLabels[i], i, label)
				}
			}
		})
	}
}

// TestTaskCreateToIssueRequest tests conversion of TaskCreate to CreateIssueRequest.
func TestTaskCreateToIssueRequest(t *testing.T) {
	tests := []struct {
		name        string
		task        *types.TaskCreate
		labelIDs    []int64
		expectedReq *CreateIssueRequest
	}{
		{
			name: "basic task with title only",
			task: &types.TaskCreate{
				Title: "Test Issue",
			},
			labelIDs: []int64{},
			expectedReq: &CreateIssueRequest{
				Title:  "Test Issue",
				Labels: []int64{},
			},
		},
		{
			name: "task with description",
			task: &types.TaskCreate{
				Title:       "Test Issue",
				Description: strPtr("Issue description"),
			},
			labelIDs: []int64{},
			expectedReq: &CreateIssueRequest{
				Title:  "Test Issue",
				Body:   "Issue description",
				Labels: []int64{},
			},
		},
		{
			name: "task with labels",
			task: &types.TaskCreate{
				Title:  "Test Issue",
				Labels: []string{"bug", "enhancement"},
			},
			labelIDs: []int64{1, 2},
			expectedReq: &CreateIssueRequest{
				Title:  "Test Issue",
				Labels: []int64{1, 2},
			},
		},
		{
			name: "task with assignee",
			task: &types.TaskCreate{
				Title:     "Test Issue",
				Assignees: []string{"user1", "user2"},
			},
			labelIDs: []int64{},
			expectedReq: &CreateIssueRequest{
				Title:    "Test Issue",
				Labels:   []int64{},
				Assignee: "user1", // Only first assignee is used
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := taskCreateToIssueRequest(tt.task, tt.labelIDs)

			if req.Title != tt.expectedReq.Title {
				t.Errorf("Expected title %s, got %s", tt.expectedReq.Title, req.Title)
			}
			if req.Body != tt.expectedReq.Body {
				t.Errorf("Expected body %s, got %s", tt.expectedReq.Body, req.Body)
			}
			if req.Assignee != tt.expectedReq.Assignee {
				t.Errorf("Expected assignee %s, got %s", tt.expectedReq.Assignee, req.Assignee)
			}
			if len(req.Labels) != len(tt.expectedReq.Labels) {
				t.Errorf("Expected %d labels, got %d", len(tt.expectedReq.Labels), len(req.Labels))
			}
		})
	}
}

// TestRepoToProject tests conversion of Repository to Project.
func TestRepoToProject(t *testing.T) {
	tests := []struct {
		name            string
		repo            *Repository
		expectedProject *types.Project
	}{
		{
			name: "basic repository",
			repo: &Repository{
				FullName:    "owner/repo",
				Name:        "repo",
				Description: "Test repository",
			},
			expectedProject: &types.Project{
				ExternalID:  "owner/repo",
				Name:        "repo",
				Description: strPtr("Test repository"),
				Status:      "active",
			},
		},
		{
			name: "repository without description",
			repo: &Repository{
				FullName: "owner/repo",
				Name:     "repo",
			},
			expectedProject: &types.Project{
				ExternalID: "owner/repo",
				Name:       "repo",
				Status:     "active",
			},
		},
		{
			name: "repository with owner but no full name",
			repo: &Repository{
				Owner: &User{Login: "myowner"},
				Name:  "myrepo",
			},
			expectedProject: &types.Project{
				ExternalID: "myowner/myrepo",
				Name:       "myrepo",
				Status:     "active",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := repoToProject(tt.repo)

			if project.ExternalID != tt.expectedProject.ExternalID {
				t.Errorf("Expected external_id %s, got %s", tt.expectedProject.ExternalID, project.ExternalID)
			}
			if project.Name != tt.expectedProject.Name {
				t.Errorf("Expected name %s, got %s", tt.expectedProject.Name, project.Name)
			}
			if project.Status != tt.expectedProject.Status {
				t.Errorf("Expected status %s, got %s", tt.expectedProject.Status, project.Status)
			}

			if tt.expectedProject.Description == nil {
				if project.Description != nil {
					t.Errorf("Expected nil description, got %s", *project.Description)
				}
			} else {
				if project.Description == nil {
					t.Fatalf("Expected description %s, got nil", *tt.expectedProject.Description)
				}
				if *project.Description != *tt.expectedProject.Description {
					t.Errorf("Expected description %s, got %s", *tt.expectedProject.Description, *project.Description)
				}
			}
		})
	}
}

// TestCommentToComment tests conversion of Forgejo Comment to Todu Comment.
func TestCommentToComment(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		comment         *Comment
		expectedComment *types.Comment
	}{
		{
			name: "basic comment",
			comment: &Comment{
				ID:        123,
				Body:      "Test comment",
				User:      &User{Login: "testuser"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedComment: &types.Comment{
				ExternalID: "123",
				Content:    "Test comment",
				Author:     "testuser",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		{
			name: "comment without user",
			comment: &Comment{
				ID:        456,
				Body:      "Anonymous comment",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedComment: &types.Comment{
				ExternalID: "456",
				Content:    "Anonymous comment",
				Author:     "",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := commentToComment(tt.comment)

			if comment.ExternalID != tt.expectedComment.ExternalID {
				t.Errorf("Expected external_id %s, got %s", tt.expectedComment.ExternalID, comment.ExternalID)
			}
			if comment.Content != tt.expectedComment.Content {
				t.Errorf("Expected content %s, got %s", tt.expectedComment.Content, comment.Content)
			}
			if comment.Author != tt.expectedComment.Author {
				t.Errorf("Expected author %s, got %s", tt.expectedComment.Author, comment.Author)
			}
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
