package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// MockPlugin is a test implementation of the Plugin interface.
//
// It stores all data in memory and can be configured to return errors
// for testing error handling. This is useful for testing code that
// depends on the Plugin interface without requiring a real external system.
//
// Example usage:
//
//	mock := plugin.NewMockPlugin("test-system")
//	mock.Configure(map[string]string{"token": "test-token"})
//	projects, err := mock.FetchProjects(ctx)
type MockPlugin struct {
	mu sync.RWMutex

	name    string
	version string
	config  map[string]string

	// In-memory storage
	projects map[string]*types.Project
	tasks    map[string]*types.Task
	comments map[string][]*types.Comment

	// Error injection for testing
	FetchProjectsError  error
	FetchProjectError   error
	FetchTasksError     error
	FetchTaskError      error
	CreateTaskError     error
	UpdateTaskError     error
	FetchCommentsError  error
	CreateCommentError  error
	ConfigureError      error
	ValidateConfigError error
}

// NewMockPlugin creates a new mock plugin with the given name.
func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		name:     name,
		version:  "1.0.0",
		config:   make(map[string]string),
		projects: make(map[string]*types.Project),
		tasks:    make(map[string]*types.Task),
		comments: make(map[string][]*types.Comment),
	}
}

// Name returns the plugin name.
func (m *MockPlugin) Name() string {
	return m.name
}

// Version returns the plugin version.
func (m *MockPlugin) Version() string {
	return m.version
}

// Configure stores the configuration.
func (m *MockPlugin) Configure(config map[string]string) error {
	if m.ConfigureError != nil {
		return m.ConfigureError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	return nil
}

// ValidateConfig checks that required configuration is present.
func (m *MockPlugin) ValidateConfig() error {
	if m.ValidateConfigError != nil {
		return m.ValidateConfigError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.config["token"]; !ok {
		return NewErrNotConfigured("missing required 'token' configuration")
	}

	return nil
}

// FetchProjects returns all stored projects.
func (m *MockPlugin) FetchProjects(ctx context.Context) ([]*types.Project, error) {
	if m.FetchProjectsError != nil {
		return nil, m.FetchProjectsError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	projects := make([]*types.Project, 0, len(m.projects))
	for _, p := range m.projects {
		projects = append(projects, p)
	}

	return projects, nil
}

// FetchProject returns a single project by external ID.
func (m *MockPlugin) FetchProject(ctx context.Context, externalID string) (*types.Project, error) {
	if m.FetchProjectError != nil {
		return nil, m.FetchProjectError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	project, ok := m.projects[externalID]
	if !ok {
		return nil, NewErrNotFound(fmt.Sprintf("project %s not found", externalID))
	}

	return project, nil
}

// FetchTasks returns all stored tasks, optionally filtered.
func (m *MockPlugin) FetchTasks(ctx context.Context, projectExternalID *string, since *time.Time) ([]*types.Task, error) {
	if m.FetchTasksError != nil {
		return nil, m.FetchTasksError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*types.Task, 0)
	for _, t := range m.tasks {
		// Filter by project if specified
		if projectExternalID != nil {
			// Find the project for this task
			var projectExtID string
			for extID, proj := range m.projects {
				if proj.ID == t.ProjectID {
					projectExtID = extID
					break
				}
			}
			if projectExtID != *projectExternalID {
				continue
			}
		}

		// Filter by time if specified
		if since != nil && t.UpdatedAt.Before(*since) {
			continue
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// FetchTask returns a single task by external ID.
func (m *MockPlugin) FetchTask(ctx context.Context, projectExternalID *string, taskExternalID string) (*types.Task, error) {
	if m.FetchTaskError != nil {
		return nil, m.FetchTaskError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskExternalID]
	if !ok {
		return nil, NewErrNotFound(fmt.Sprintf("task %s not found", taskExternalID))
	}

	return task, nil
}

// CreateTask creates a new task in memory.
func (m *MockPlugin) CreateTask(ctx context.Context, projectExternalID *string, task *types.TaskCreate) (*types.Task, error) {
	if m.CreateTaskError != nil {
		return nil, m.CreateTaskError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	externalID := fmt.Sprintf("task-%d", len(m.tasks)+1)

	created := &types.Task{
		ID:          len(m.tasks) + 1,
		ExternalID:  externalID,
		SourceURL:   task.SourceURL,
		Title:       task.Title,
		Description: task.Description,
		ProjectID:   task.ProjectID,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
		Labels:      task.Labels,
		Assignees:   task.Assignees,
	}

	m.tasks[externalID] = created
	return created, nil
}

// UpdateTask updates an existing task in memory.
func (m *MockPlugin) UpdateTask(ctx context.Context, projectExternalID *string, taskExternalID string, task *types.TaskUpdate) (*types.Task, error) {
	if m.UpdateTaskError != nil {
		return nil, m.UpdateTaskError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.tasks[taskExternalID]
	if !ok {
		return nil, NewErrNotFound(fmt.Sprintf("task %s not found", taskExternalID))
	}

	// Update only provided fields
	if task.Title != nil {
		existing.Title = *task.Title
	}
	if task.Description != nil {
		existing.Description = task.Description
	}
	if task.Status != nil {
		existing.Status = *task.Status
	}
	if task.Priority != nil {
		existing.Priority = task.Priority
	}
	if task.DueDate != nil {
		existing.DueDate = task.DueDate
	}
	if task.Labels != nil {
		existing.Labels = task.Labels
	}
	if task.Assignees != nil {
		existing.Assignees = task.Assignees
	}

	existing.UpdatedAt = time.Now()

	return existing, nil
}

// FetchComments returns all comments for a task.
func (m *MockPlugin) FetchComments(ctx context.Context, projectExternalID *string, taskExternalID string) ([]*types.Comment, error) {
	if m.FetchCommentsError != nil {
		return nil, m.FetchCommentsError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify task exists
	if _, ok := m.tasks[taskExternalID]; !ok {
		return nil, NewErrNotFound(fmt.Sprintf("task %s not found", taskExternalID))
	}

	comments, ok := m.comments[taskExternalID]
	if !ok {
		return []*types.Comment{}, nil
	}

	return comments, nil
}

// CreateComment creates a new comment in memory.
func (m *MockPlugin) CreateComment(ctx context.Context, projectExternalID *string, taskExternalID string, comment *types.CommentCreate) (*types.Comment, error) {
	if m.CreateCommentError != nil {
		return nil, m.CreateCommentError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify task exists
	if _, ok := m.tasks[taskExternalID]; !ok {
		return nil, NewErrNotFound(fmt.Sprintf("task %s not found", taskExternalID))
	}

	now := time.Now()
	created := &types.Comment{
		ID:        len(m.comments[taskExternalID]) + 1,
		TaskID:    comment.TaskID,
		Content:   comment.Content,
		Author:    comment.Author,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.comments[taskExternalID] = append(m.comments[taskExternalID], created)
	return created, nil
}

// AddProject adds a project to the mock plugin's storage.
// This is a test helper method not part of the Plugin interface.
func (m *MockPlugin) AddProject(externalID string, project *types.Project) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[externalID] = project
}

// AddTask adds a task to the mock plugin's storage.
// This is a test helper method not part of the Plugin interface.
func (m *MockPlugin) AddTask(externalID string, task *types.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[externalID] = task
}

// Reset clears all stored data.
// This is a test helper method not part of the Plugin interface.
func (m *MockPlugin) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.projects = make(map[string]*types.Project)
	m.tasks = make(map[string]*types.Task)
	m.comments = make(map[string][]*types.Comment)
}
