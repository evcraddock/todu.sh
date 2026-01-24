package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

// Plugin implements the plugin.Plugin interface for GitHub.
type Plugin struct {
	client *client
	config map[string]string
}

// init registers the GitHub plugin with the global registry.
func init() {
	registry.Register("github", func() plugin.Plugin {
		return &Plugin{}
	})
}

// Name returns the unique identifier for this plugin.
func (p *Plugin) Name() string {
	return "github"
}

// Version returns the version of this plugin implementation.
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Configure provides configuration to the plugin.
// Required configuration keys:
//   - token: GitHub personal access token
//
// Optional configuration keys:
//   - url: GitHub API URL (defaults to "https://api.github.com")
func (p *Plugin) Configure(config map[string]string) error {
	p.config = config

	// Validate required configuration
	if err := p.ValidateConfig(); err != nil {
		return err
	}

	// Create GitHub API client
	var err error
	p.client, err = newClient(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return nil
}

// ValidateConfig checks that the plugin has been properly configured.
func (p *Plugin) ValidateConfig() error {
	if p.config == nil {
		return plugin.ErrNotConfigured
	}

	// Check required fields
	token := p.config["token"]
	if token == "" {
		return fmt.Errorf("%w: missing required field 'token'", plugin.ErrNotConfigured)
	}

	// Validate URL format if provided
	if url := p.config["url"]; url != "" {
		// Basic validation - just check it's not empty
		// The actual URL validation will happen when creating the client
		if len(url) == 0 {
			return fmt.Errorf("%w: 'url' cannot be empty", plugin.ErrNotConfigured)
		}
	}

	return nil
}

// FetchProjects retrieves all repositories accessible by this plugin.
func (p *Plugin) FetchProjects(ctx context.Context) ([]*types.Project, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	repos, err := p.client.listRepositories(ctx)
	if err != nil {
		return nil, handleGitHubError(err, "failed to list repositories")
	}

	projects := make([]*types.Project, len(repos))
	for i, repo := range repos {
		projects[i] = repoToProject(repo)
	}

	return projects, nil
}

// FetchProject retrieves a single project by its external ID.
func (p *Plugin) FetchProject(ctx context.Context, externalID string) (*types.Project, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	owner, repo, err := parseRepoExternalID(externalID)
	if err != nil {
		return nil, err
	}

	repository, err := p.client.getRepository(ctx, owner, repo)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to fetch repository %s", externalID))
	}

	return repoToProject(repository), nil
}

// FetchTasks retrieves issues from GitHub.
func (p *Plugin) FetchTasks(ctx context.Context, projectExternalID *string, since *time.Time) ([]*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	if projectExternalID == nil {
		return nil, fmt.Errorf("projectExternalID is required for GitHub")
	}

	owner, repo, err := parseRepoExternalID(*projectExternalID)
	if err != nil {
		return nil, err
	}

	issues, err := p.client.listIssues(ctx, owner, repo, since)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to list issues for %s", *projectExternalID))
	}

	// Filter out pull requests (GitHub's Issues API returns both issues and PRs)
	var tasks []*types.Task
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			continue
		}
		tasks = append(tasks, issueToTask(issue, owner, repo))
	}

	return tasks, nil
}

// FetchTask retrieves a single task by its external ID.
func (p *Plugin) FetchTask(ctx context.Context, projectExternalID *string, taskExternalID string) (*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	if projectExternalID == nil {
		return nil, fmt.Errorf("projectExternalID is required for GitHub")
	}

	owner, repo, err := parseRepoExternalID(*projectExternalID)
	if err != nil {
		return nil, err
	}

	issueNumber, err := strconv.Atoi(taskExternalID)
	if err != nil {
		return nil, fmt.Errorf("invalid task external_id: %w", err)
	}

	issue, err := p.client.getIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to fetch issue %s#%s", *projectExternalID, taskExternalID))
	}

	return issueToTask(issue, owner, repo), nil
}

// CreateTask creates a new issue in GitHub.
func (p *Plugin) CreateTask(ctx context.Context, projectExternalID *string, task *types.TaskCreate) (*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	if projectExternalID == nil {
		return nil, fmt.Errorf("projectExternalID is required for GitHub")
	}

	owner, repo, err := parseRepoExternalID(*projectExternalID)
	if err != nil {
		return nil, err
	}

	req := taskCreateToIssueRequest(task)
	issue, err := p.client.createIssue(ctx, owner, repo, req)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to create issue in %s", *projectExternalID))
	}

	return issueToTask(issue, owner, repo), nil
}

// UpdateTask updates an existing issue in GitHub.
func (p *Plugin) UpdateTask(ctx context.Context, projectExternalID *string, taskExternalID string, task *types.TaskUpdate) (*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	if projectExternalID == nil {
		return nil, fmt.Errorf("projectExternalID is required for GitHub")
	}

	owner, repo, err := parseRepoExternalID(*projectExternalID)
	if err != nil {
		return nil, err
	}

	issueNumber, err := strconv.Atoi(taskExternalID)
	if err != nil {
		return nil, fmt.Errorf("invalid task external_id: %w", err)
	}

	req := taskUpdateToIssueRequest(task)
	issue, err := p.client.updateIssue(ctx, owner, repo, issueNumber, req)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to update issue %s#%s", *projectExternalID, taskExternalID))
	}

	return issueToTask(issue, owner, repo), nil
}

// FetchComments retrieves all comments for an issue.
func (p *Plugin) FetchComments(ctx context.Context, projectExternalID *string, taskExternalID string) ([]*types.Comment, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	if projectExternalID == nil {
		return nil, fmt.Errorf("projectExternalID is required for GitHub")
	}

	owner, repo, err := parseRepoExternalID(*projectExternalID)
	if err != nil {
		return nil, err
	}

	issueNumber, err := strconv.Atoi(taskExternalID)
	if err != nil {
		return nil, fmt.Errorf("invalid task external_id: %w", err)
	}

	comments, err := p.client.listComments(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to list comments for %s#%s", *projectExternalID, taskExternalID))
	}

	result := make([]*types.Comment, len(comments))
	for i, comment := range comments {
		result[i] = commentToComment(comment)
	}

	return result, nil
}

// CreateComment creates a new comment on an issue.
func (p *Plugin) CreateComment(ctx context.Context, projectExternalID *string, taskExternalID string, comment *types.CommentCreate) (*types.Comment, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	if projectExternalID == nil {
		return nil, fmt.Errorf("projectExternalID is required for GitHub")
	}

	owner, repo, err := parseRepoExternalID(*projectExternalID)
	if err != nil {
		return nil, err
	}

	issueNumber, err := strconv.Atoi(taskExternalID)
	if err != nil {
		return nil, fmt.Errorf("invalid task external_id: %w", err)
	}

	ghComment, err := p.client.createComment(ctx, owner, repo, issueNumber, comment.Content)
	if err != nil {
		return nil, handleGitHubError(err, fmt.Sprintf("failed to create comment on %s#%s", *projectExternalID, taskExternalID))
	}

	return commentToComment(ghComment), nil
}

// handleGitHubError converts GitHub API errors to plugin errors.
func handleGitHubError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Check for GitHub API error response
	errMsg := err.Error()

	// Check for 404 Not Found
	if strings.Contains(errMsg, "404") {
		return plugin.NewErrNotFound(context)
	}

	// Check for 401/403 Unauthorized/Forbidden
	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") {
		return plugin.NewErrUnauthorized(context)
	}

	// Return generic error with context
	return fmt.Errorf("%s: %w", context, err)
}
