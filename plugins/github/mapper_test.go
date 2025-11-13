package github

import (
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/google/go-github/v56/github"
)

// TestMapToduStatusToGitHub tests the mapping from todu status to GitHub state and state_reason.
func TestMapToduStatusToGitHub(t *testing.T) {
	tests := []struct {
		name              string
		toduStatus        string
		expectedState     string
		expectedStateReason string
	}{
		{
			name:              "done maps to closed+completed",
			toduStatus:        "done",
			expectedState:     "closed",
			expectedStateReason: "completed",
		},
		{
			name:              "cancelled maps to closed+not_planned",
			toduStatus:        "cancelled",
			expectedState:     "closed",
			expectedStateReason: "not_planned",
		},
		{
			name:              "active maps to open",
			toduStatus:        "active",
			expectedState:     "open",
			expectedStateReason: "",
		},
		{
			name:              "inprogress maps to open",
			toduStatus:        "inprogress",
			expectedState:     "open",
			expectedStateReason: "",
		},
		{
			name:              "waiting maps to open",
			toduStatus:        "waiting",
			expectedState:     "open",
			expectedStateReason: "",
		},
		{
			name:              "unknown status maps to open",
			toduStatus:        "some-unknown-status",
			expectedState:     "open",
			expectedStateReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, stateReason := mapToduStatusToGitHub(tt.toduStatus)
			if state != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, state)
			}
			if stateReason != tt.expectedStateReason {
				t.Errorf("Expected stateReason %s, got %s", tt.expectedStateReason, stateReason)
			}
		})
	}
}

// TestMapGitHubStatusToTodu tests the mapping from GitHub state and state_reason to todu status.
func TestMapGitHubStatusToTodu(t *testing.T) {
	tests := []struct {
		name           string
		githubState    string
		githubStateReason string
		expectedStatus string
	}{
		{
			name:           "closed+completed maps to done",
			githubState:    "closed",
			githubStateReason: "completed",
			expectedStatus: "done",
		},
		{
			name:           "closed+not_planned maps to cancelled",
			githubState:    "closed",
			githubStateReason: "not_planned",
			expectedStatus: "cancelled",
		},
		{
			name:           "closed with no state_reason maps to done (backward compatibility)",
			githubState:    "closed",
			githubStateReason: "",
			expectedStatus: "done",
		},
		{
			name:           "closed+reopened maps to done",
			githubState:    "closed",
			githubStateReason: "reopened",
			expectedStatus: "done",
		},
		{
			name:           "open maps to active",
			githubState:    "open",
			githubStateReason: "",
			expectedStatus: "active",
		},
		{
			name:           "open with any state_reason maps to active",
			githubState:    "open",
			githubStateReason: "reopened",
			expectedStatus: "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := mapGitHubStatusToTodu(tt.githubState, tt.githubStateReason)
			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

// TestIssueToTask_StatusMapping tests that issueToTask correctly maps GitHub state_reason.
func TestIssueToTask_StatusMapping(t *testing.T) {
	now := time.Now()
	owner := "test-owner"
	repo := "test-repo"

	tests := []struct {
		name           string
		githubState    string
		githubStateReason string
		expectedStatus string
	}{
		{
			name:           "closed issue with completed reason",
			githubState:    "closed",
			githubStateReason: "completed",
			expectedStatus: "done",
		},
		{
			name:           "closed issue with not_planned reason",
			githubState:    "closed",
			githubStateReason: "not_planned",
			expectedStatus: "cancelled",
		},
		{
			name:           "closed issue without state_reason (backward compatibility)",
			githubState:    "closed",
			githubStateReason: "",
			expectedStatus: "done",
		},
		{
			name:           "open issue",
			githubState:    "open",
			githubStateReason: "",
			expectedStatus: "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &github.Issue{
				Number:      github.Int(1),
				Title:       github.String("Test Issue"),
				Body:        github.String("Test body"),
				State:       github.String(tt.githubState),
				StateReason: github.String(tt.githubStateReason),
				HTMLURL:     github.String("https://github.com/test/test/issues/1"),
				CreatedAt:   &github.Timestamp{Time: now},
				UpdatedAt:   &github.Timestamp{Time: now},
			}

			task := issueToTask(issue, owner, repo)

			if task.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, task.Status)
			}
		})
	}
}

// TestTaskUpdateToIssueRequest_StatusMapping tests that taskUpdateToIssueRequest sets state_reason.
func TestTaskUpdateToIssueRequest_StatusMapping(t *testing.T) {
	tests := []struct {
		name              string
		toduStatus        string
		expectedState     string
		expectedStateReason *string
	}{
		{
			name:              "done status sets completed state_reason",
			toduStatus:        "done",
			expectedState:     "closed",
			expectedStateReason: github.String("completed"),
		},
		{
			name:              "cancelled status sets not_planned state_reason",
			toduStatus:        "cancelled",
			expectedState:     "closed",
			expectedStateReason: github.String("not_planned"),
		},
		{
			name:              "active status does not set state_reason",
			toduStatus:        "active",
			expectedState:     "open",
			expectedStateReason: nil,
		},
		{
			name:              "inprogress status does not set state_reason",
			toduStatus:        "inprogress",
			expectedState:     "open",
			expectedStateReason: nil,
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

			if tt.expectedStateReason == nil {
				if req.StateReason != nil {
					t.Errorf("Expected StateReason to be nil, got %s", *req.StateReason)
				}
			} else {
				if req.StateReason == nil {
					t.Fatal("Expected StateReason to be set")
				}
				if *req.StateReason != *tt.expectedStateReason {
					t.Errorf("Expected StateReason %s, got %s", *tt.expectedStateReason, *req.StateReason)
				}
			}
		})
	}
}

// TestTaskUpdateToIssueRequest_NoStatusChange tests that state_reason is not set when status is nil.
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
	if req.StateReason != nil {
		t.Errorf("Expected StateReason to be nil when status is not updated, got %s", *req.StateReason)
	}
}
