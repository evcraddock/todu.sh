package todoist

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

// Plugin implements the plugin.Plugin interface for Todoist.
type Plugin struct {
	client        *client
	config        map[string]string
	defaultAuthor string
}

// init registers the Todoist plugin with the global registry.
func init() {
	registry.Register("todoist", func() plugin.Plugin {
		return &Plugin{}
	})
}

// Name returns the unique identifier for this plugin.
func (p *Plugin) Name() string {
	return "todoist"
}

// Version returns the version of this plugin implementation.
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Configure provides configuration to the plugin.
// Required configuration keys:
//   - token: Todoist API token
//
// Optional configuration keys:
//   - default_author: Default author for comments (defaults to "todoist")
func (p *Plugin) Configure(config map[string]string) error {
	p.config = config

	// Validate required configuration
	if err := p.ValidateConfig(); err != nil {
		return err
	}

	// Set default author (Todoist REST API doesn't provide author info)
	p.defaultAuthor = "todoist"
	if author := strings.TrimSpace(config["default_author"]); author != "" {
		p.defaultAuthor = author
	}

	// Create Todoist API client
	var err error
	p.client, err = newClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Todoist client: %w", err)
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

	return nil
}

// FetchProjects retrieves all projects accessible by this plugin.
func (p *Plugin) FetchProjects(ctx context.Context) ([]*types.Project, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	projects, err := p.client.listProjects(ctx)
	if err != nil {
		return nil, handleTodoistError(err, "failed to list projects")
	}

	result := make([]*types.Project, len(projects))
	for i, project := range projects {
		result[i] = projectToProject(project)
	}

	return result, nil
}

// FetchProject retrieves a single project by its external ID.
func (p *Plugin) FetchProject(ctx context.Context, externalID string) (*types.Project, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	project, err := p.client.getProject(ctx, externalID)
	if err != nil {
		return nil, handleTodoistError(err, fmt.Sprintf("failed to fetch project %s", externalID))
	}

	return projectToProject(project), nil
}

// FetchTasks retrieves tasks from Todoist.
func (p *Plugin) FetchTasks(ctx context.Context, projectExternalID *string, since *time.Time) ([]*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	tasks, err := p.client.listTasks(ctx, projectExternalID, since)
	if err != nil {
		return nil, handleTodoistError(err, "failed to list tasks")
	}

	result := make([]*types.Task, len(tasks))
	for i, task := range tasks {
		result[i] = taskToTask(task)
	}

	return result, nil
}

// FetchTask retrieves a single task by its external ID.
func (p *Plugin) FetchTask(ctx context.Context, projectExternalID *string, taskExternalID string) (*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	// Todoist doesn't require project ID to fetch a task
	task, err := p.client.getTask(ctx, taskExternalID)
	if err != nil {
		return nil, handleTodoistError(err, fmt.Sprintf("failed to fetch task %s", taskExternalID))
	}

	return taskToTask(task), nil
}

// CreateTask creates a new task in Todoist.
func (p *Plugin) CreateTask(ctx context.Context, projectExternalID *string, task *types.TaskCreate) (*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	req := taskCreateToRequest(task, projectExternalID)
	created, err := p.client.createTask(ctx, req)
	if err != nil {
		return nil, handleTodoistError(err, "failed to create task")
	}

	return taskToTask(created), nil
}

// UpdateTask updates an existing task in Todoist.
func (p *Plugin) UpdateTask(ctx context.Context, projectExternalID *string, taskExternalID string, task *types.TaskUpdate) (*types.Task, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	// Handle status changes (close/reopen)
	if task.Status != nil {
		if shouldClose(task.Status) {
			if err := p.client.closeTask(ctx, taskExternalID); err != nil {
				return nil, handleTodoistError(err, fmt.Sprintf("failed to close task %s", taskExternalID))
			}
		} else if shouldReopen(task.Status) {
			if err := p.client.reopenTask(ctx, taskExternalID); err != nil {
				return nil, handleTodoistError(err, fmt.Sprintf("failed to reopen task %s", taskExternalID))
			}
		}
	}

	// Update other fields if any are specified
	req := taskUpdateToRequest(task)
	hasUpdates := req.Content != nil || req.Description != nil || req.Priority != nil || len(req.Labels) > 0 || req.DueDate != nil

	var updated *Task
	var err error

	if hasUpdates {
		updated, err = p.client.updateTask(ctx, taskExternalID, req)
		if err != nil {
			return nil, handleTodoistError(err, fmt.Sprintf("failed to update task %s", taskExternalID))
		}
	} else {
		// Just fetch the current state if only status was changed
		updated, err = p.client.getTask(ctx, taskExternalID)
		if err != nil {
			return nil, handleTodoistError(err, fmt.Sprintf("failed to fetch task %s after status change", taskExternalID))
		}
	}

	return taskToTask(updated), nil
}

// FetchComments retrieves all comments for a task.
func (p *Plugin) FetchComments(ctx context.Context, projectExternalID *string, taskExternalID string) ([]*types.Comment, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	comments, err := p.client.listComments(ctx, taskExternalID)
	if err != nil {
		return nil, handleTodoistError(err, fmt.Sprintf("failed to list comments for task %s", taskExternalID))
	}

	result := make([]*types.Comment, len(comments))
	for i, comment := range comments {
		result[i] = commentToComment(comment, p.defaultAuthor)
	}

	return result, nil
}

// CreateComment creates a new comment on a task.
func (p *Plugin) CreateComment(ctx context.Context, projectExternalID *string, taskExternalID string, comment *types.CommentCreate) (*types.Comment, error) {
	if p.client == nil {
		return nil, plugin.ErrNotConfigured
	}

	created, err := p.client.createComment(ctx, taskExternalID, comment.Content)
	if err != nil {
		return nil, handleTodoistError(err, fmt.Sprintf("failed to create comment on task %s", taskExternalID))
	}

	return commentToComment(created, p.defaultAuthor), nil
}

// handleTodoistError converts Todoist API errors to plugin errors.
func handleTodoistError(err error, context string) error {
	if err == nil {
		return nil
	}

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
