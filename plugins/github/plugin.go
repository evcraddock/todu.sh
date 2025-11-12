package github

import (
	"context"
	"fmt"
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
	return nil, plugin.ErrNotSupported
}

// FetchProject retrieves a single project by its external ID.
func (p *Plugin) FetchProject(ctx context.Context, externalID string) (*types.Project, error) {
	return nil, plugin.ErrNotSupported
}

// FetchTasks retrieves issues from GitHub.
func (p *Plugin) FetchTasks(ctx context.Context, projectExternalID *string, since *time.Time) ([]*types.Task, error) {
	return nil, plugin.ErrNotSupported
}

// FetchTask retrieves a single task by its external ID.
func (p *Plugin) FetchTask(ctx context.Context, projectExternalID *string, taskExternalID string) (*types.Task, error) {
	return nil, plugin.ErrNotSupported
}

// CreateTask creates a new issue in GitHub.
func (p *Plugin) CreateTask(ctx context.Context, projectExternalID *string, task *types.TaskCreate) (*types.Task, error) {
	return nil, plugin.ErrNotSupported
}

// UpdateTask updates an existing issue in GitHub.
func (p *Plugin) UpdateTask(ctx context.Context, projectExternalID *string, taskExternalID string, task *types.TaskUpdate) (*types.Task, error) {
	return nil, plugin.ErrNotSupported
}

// FetchComments retrieves all comments for an issue.
func (p *Plugin) FetchComments(ctx context.Context, projectExternalID *string, taskExternalID string) ([]*types.Comment, error) {
	return nil, plugin.ErrNotSupported
}

// CreateComment creates a new comment on an issue.
func (p *Plugin) CreateComment(ctx context.Context, projectExternalID *string, taskExternalID string, comment *types.CommentCreate) (*types.Comment, error) {
	return nil, plugin.ErrNotSupported
}
