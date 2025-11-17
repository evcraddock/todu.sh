// Package local provides a no-op plugin for local-only projects.
//
// The local plugin allows projects to exist in todu without requiring
// synchronization to an external system. All sync operations are no-ops,
// meaning tasks remain purely local.
package local

import (
	"context"
	"time"

	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

// Plugin implements the plugin.Plugin interface for local-only projects.
// All operations are no-ops since there is no external system to sync with.
type Plugin struct {
	config map[string]string
}

func init() {
	err := registry.Register("local", func() plugin.Plugin {
		return &Plugin{}
	})
	if err != nil {
		panic(err)
	}
}

// Name returns the unique identifier for this plugin.
func (p *Plugin) Name() string {
	return "local"
}

// Version returns the version of this plugin implementation.
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Configure provides configuration to the plugin.
// The local plugin requires no configuration.
func (p *Plugin) Configure(config map[string]string) error {
	p.config = config
	return nil
}

// ValidateConfig checks that the plugin has been properly configured.
// The local plugin requires no configuration, so this always succeeds.
func (p *Plugin) ValidateConfig() error {
	return nil
}

// FetchProjects returns an empty slice since local projects are not fetched
// from any external system.
func (p *Plugin) FetchProjects(_ context.Context) ([]*types.Project, error) {
	return []*types.Project{}, nil
}

// FetchProject returns ErrNotFound since local projects don't exist in an
// external system.
func (p *Plugin) FetchProject(_ context.Context, _ string) (*types.Project, error) {
	return nil, plugin.NewErrNotFound("local projects are not fetched from external systems")
}

// FetchTasks returns an empty slice since local tasks are not fetched from
// any external system.
func (p *Plugin) FetchTasks(_ context.Context, _ *string, _ *time.Time) ([]*types.Task, error) {
	return []*types.Task{}, nil
}

// FetchTask returns ErrNotFound since local tasks don't exist in an external
// system.
func (p *Plugin) FetchTask(_ context.Context, _ *string, _ string) (*types.Task, error) {
	return nil, plugin.NewErrNotFound("local tasks are not fetched from external systems")
}

// CreateTask returns ErrNotSupported for local projects. Tasks are created
// directly in todu without syncing to an external system.
func (p *Plugin) CreateTask(_ context.Context, _ *string, _ *types.TaskCreate) (*types.Task, error) {
	return nil, plugin.ErrNotSupported
}

// UpdateTask returns ErrNotSupported for local projects. Tasks are updated
// directly in todu without syncing to an external system.
func (p *Plugin) UpdateTask(_ context.Context, _ *string, _ string, _ *types.TaskUpdate) (*types.Task, error) {
	return nil, plugin.ErrNotSupported
}

// FetchComments returns an empty slice since local comments are not fetched
// from any external system.
func (p *Plugin) FetchComments(_ context.Context, _ *string, _ string) ([]*types.Comment, error) {
	return []*types.Comment{}, nil
}

// CreateComment returns ErrNotSupported for local projects. Comments are
// created directly in todu without syncing to an external system.
func (p *Plugin) CreateComment(_ context.Context, _ *string, _ string, _ *types.CommentCreate) (*types.Comment, error) {
	return nil, plugin.ErrNotSupported
}
