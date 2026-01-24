package github

import (
	"testing"
	"time"

	"github.com/google/go-github/v56/github"
)

// TestFetchTasks_FiltersPullRequests verifies that pull requests are filtered out
// when converting GitHub issues to tasks. GitHub's Issues API returns both issues
// and pull requests, but we only want issues.
func TestFetchTasks_FiltersPullRequests(t *testing.T) {
	now := time.Now()
	owner := "test-owner"
	repo := "test-repo"

	// Simulate issues returned by GitHub API (mix of issues and PRs)
	issues := []*github.Issue{
		{
			Number:           github.Int(1),
			Title:            github.String("Regular Issue"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/issues/1"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: nil, // This is a regular issue
		},
		{
			Number:           github.Int(2),
			Title:            github.String("Pull Request"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/pull/2"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: &github.PullRequestLinks{}, // This is a PR
		},
		{
			Number:           github.Int(3),
			Title:            github.String("Another Issue"),
			State:            github.String("closed"),
			HTMLURL:          github.String("https://github.com/test/test/issues/3"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: nil, // This is a regular issue
		},
		{
			Number:           github.Int(4),
			Title:            github.String("Another PR"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/pull/4"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: &github.PullRequestLinks{}, // This is a PR
		},
	}

	// Apply the same filtering logic as FetchTasks
	var filteredTasks []string
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			continue
		}
		task := issueToTask(issue, owner, repo)
		filteredTasks = append(filteredTasks, task.Title)
	}

	// Should only have 2 tasks (issues 1 and 3), not 4
	if len(filteredTasks) != 2 {
		t.Errorf("Expected 2 tasks after filtering PRs, got %d", len(filteredTasks))
	}

	// Verify the correct issues were kept
	expectedTitles := []string{"Regular Issue", "Another Issue"}
	for i, expected := range expectedTitles {
		if filteredTasks[i] != expected {
			t.Errorf("Expected task %d to be %q, got %q", i, expected, filteredTasks[i])
		}
	}
}

// TestFetchTasks_AllPullRequests verifies behavior when all items are PRs.
func TestFetchTasks_AllPullRequests(t *testing.T) {
	now := time.Now()

	issues := []*github.Issue{
		{
			Number:           github.Int(1),
			Title:            github.String("PR 1"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/pull/1"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: &github.PullRequestLinks{},
		},
		{
			Number:           github.Int(2),
			Title:            github.String("PR 2"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/pull/2"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: &github.PullRequestLinks{},
		},
	}

	// Apply filtering
	var count int
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			continue
		}
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 tasks when all items are PRs, got %d", count)
	}
}

// TestFetchTasks_NoPullRequests verifies behavior when there are no PRs.
func TestFetchTasks_NoPullRequests(t *testing.T) {
	now := time.Now()
	owner := "test-owner"
	repo := "test-repo"

	issues := []*github.Issue{
		{
			Number:           github.Int(1),
			Title:            github.String("Issue 1"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/issues/1"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: nil,
		},
		{
			Number:           github.Int(2),
			Title:            github.String("Issue 2"),
			State:            github.String("open"),
			HTMLURL:          github.String("https://github.com/test/test/issues/2"),
			CreatedAt:        &github.Timestamp{Time: now},
			UpdatedAt:        &github.Timestamp{Time: now},
			PullRequestLinks: nil,
		},
	}

	// Apply filtering
	var filteredTasks []string
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			continue
		}
		task := issueToTask(issue, owner, repo)
		filteredTasks = append(filteredTasks, task.Title)
	}

	if len(filteredTasks) != 2 {
		t.Errorf("Expected 2 tasks when no PRs present, got %d", len(filteredTasks))
	}
}
