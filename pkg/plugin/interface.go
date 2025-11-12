// Package plugin defines the interface for external task management system integrations.
//
// Plugins are the primary mechanism for integrating with external systems like GitHub,
// Forgejo, and Todoist. Each plugin implements the Plugin interface to provide a
// consistent way to fetch and manage tasks across different platforms.
//
// # Plugin Architecture
//
// Plugins are designed to be:
//   - Independent: Each plugin is self-contained and doesn't depend on other plugins
//   - Configurable: Plugins receive configuration via the Configure method
//   - Stateless: Plugins should not maintain state between method calls (use the API for persistence)
//
// # External IDs
//
// The external_id field is critical for mapping resources between the external system
// and Todu:
//   - For projects: The unique identifier in the external system (e.g., "owner/repo" for GitHub)
//   - For tasks: The unique identifier in the external system (e.g., issue number, task ID)
//
// # Optional Operations
//
// Not all plugins support all operations. For example:
//   - A read-only integration might not support CreateTask or UpdateTask
//   - Some systems might not support comments
//
// Plugins should return ErrNotSupported for operations they don't implement.
//
// # Configuration
//
// Plugins receive configuration as a map[string]string. Common configuration keys include:
//   - "token" or "api_key": Authentication credentials
//   - "url" or "base_url": API endpoint URL
//   - "username": Username for authentication
//
// Plugins should document their required and optional configuration keys.
package plugin

import (
	"context"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// Plugin defines the interface that all external system integrations must implement.
//
// Plugins provide a bridge between Todu and external task management systems,
// allowing Todu to fetch, create, and update tasks across different platforms.
type Plugin interface {
	// Name returns the unique identifier for this plugin.
	// This should be a stable, lowercase identifier (e.g., "github", "forgejo", "todoist").
	Name() string

	// Version returns the version of this plugin implementation.
	// Should follow semantic versioning (e.g., "1.0.0").
	Version() string

	// Configure provides configuration to the plugin.
	// The config map contains key-value pairs specific to this plugin.
	// This method should validate and store the configuration for later use.
	// Returns an error if the configuration is invalid.
	Configure(config map[string]string) error

	// ValidateConfig checks that the plugin has been properly configured.
	// This should verify that all required configuration keys are present
	// and that authentication credentials are valid.
	// Returns ErrNotConfigured if configuration is missing or invalid.
	ValidateConfig() error

	// FetchProjects retrieves all projects accessible by this plugin.
	// For GitHub, this might return all repositories the user has access to.
	// For Todoist, this might return all projects.
	// Returns a slice of projects with external_id populated.
	FetchProjects(ctx context.Context) ([]*types.Project, error)

	// FetchProject retrieves a single project by its external ID.
	// The externalID should match the external_id format for this plugin.
	// Returns ErrNotFound if the project doesn't exist.
	FetchProject(ctx context.Context, externalID string) (*types.Project, error)

	// FetchTasks retrieves tasks from the external system.
	//
	// Parameters:
	//   - projectExternalID: Optional filter to fetch tasks for a specific project.
	//     If nil, fetch tasks across all projects the plugin has access to.
	//   - since: Optional filter to fetch only tasks modified after this time.
	//     If nil, fetch all tasks.
	//
	// Returns a slice of tasks with external_id and project_id populated.
	FetchTasks(ctx context.Context, projectExternalID *string, since *time.Time) ([]*types.Task, error)

	// FetchTask retrieves a single task by its external ID.
	//
	// Parameters:
	//   - projectExternalID: Optional project identifier for systems that require it.
	//     For GitHub, this would be the "owner/repo". Can be nil if not needed.
	//   - taskExternalID: The external identifier for the task (e.g., issue number).
	//
	// Returns ErrNotFound if the task doesn't exist.
	FetchTask(ctx context.Context, projectExternalID *string, taskExternalID string) (*types.Task, error)

	// CreateTask creates a new task in the external system.
	//
	// Parameters:
	//   - projectExternalID: Optional project identifier. Required by some systems.
	//   - task: The task to create. The external_id field will be ignored and set by the plugin.
	//
	// Returns the created task with external_id populated.
	// Returns ErrNotSupported if the plugin doesn't support task creation.
	CreateTask(ctx context.Context, projectExternalID *string, task *types.TaskCreate) (*types.Task, error)

	// UpdateTask updates an existing task in the external system.
	//
	// Parameters:
	//   - projectExternalID: Optional project identifier. Required by some systems.
	//   - taskExternalID: The external identifier for the task to update.
	//   - task: The fields to update. Only non-nil fields should be updated.
	//
	// Returns the updated task.
	// Returns ErrNotSupported if the plugin doesn't support task updates.
	// Returns ErrNotFound if the task doesn't exist.
	UpdateTask(ctx context.Context, projectExternalID *string, taskExternalID string, task *types.TaskUpdate) (*types.Task, error)

	// FetchComments retrieves all comments for a task.
	//
	// Parameters:
	//   - projectExternalID: Optional project identifier. Required by some systems.
	//   - taskExternalID: The external identifier for the task.
	//
	// Returns a slice of comments for the task.
	// Returns ErrNotSupported if the plugin doesn't support comments.
	// Returns ErrNotFound if the task doesn't exist.
	FetchComments(ctx context.Context, projectExternalID *string, taskExternalID string) ([]*types.Comment, error)

	// CreateComment creates a new comment on a task.
	//
	// Parameters:
	//   - projectExternalID: Optional project identifier. Required by some systems.
	//   - taskExternalID: The external identifier for the task.
	//   - comment: The comment to create.
	//
	// Returns the created comment.
	// Returns ErrNotSupported if the plugin doesn't support comments.
	// Returns ErrNotFound if the task doesn't exist.
	CreateComment(ctx context.Context, projectExternalID *string, taskExternalID string, comment *types.CommentCreate) (*types.Comment, error)
}
