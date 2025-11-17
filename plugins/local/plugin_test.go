package local

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	if got := p.Name(); got != "local" {
		t.Errorf("Name() = %q, want %q", got, "local")
	}
}

func TestPlugin_Version(t *testing.T) {
	p := &Plugin{}
	if got := p.Version(); got != "1.0.0" {
		t.Errorf("Version() = %q, want %q", got, "1.0.0")
	}
}

func TestPlugin_Configure(t *testing.T) {
	p := &Plugin{}
	config := map[string]string{
		"key": "value",
	}
	err := p.Configure(config)
	if err != nil {
		t.Errorf("Configure() error = %v, want nil", err)
	}
	if p.config["key"] != "value" {
		t.Errorf("Configure() did not store config, got %v", p.config)
	}
}

func TestPlugin_Configure_EmptyConfig(t *testing.T) {
	p := &Plugin{}
	err := p.Configure(map[string]string{})
	if err != nil {
		t.Errorf("Configure() with empty config error = %v, want nil", err)
	}
}

func TestPlugin_ValidateConfig(t *testing.T) {
	p := &Plugin{}
	err := p.ValidateConfig()
	if err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestPlugin_ValidateConfig_WithoutConfigure(t *testing.T) {
	p := &Plugin{}
	// Should succeed even without Configure being called
	err := p.ValidateConfig()
	if err != nil {
		t.Errorf("ValidateConfig() without Configure error = %v, want nil", err)
	}
}

func TestPlugin_FetchProjects(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projects, err := p.FetchProjects(ctx)
	if err != nil {
		t.Errorf("FetchProjects() error = %v, want nil", err)
	}
	if len(projects) != 0 {
		t.Errorf("FetchProjects() returned %d projects, want 0", len(projects))
	}
}

func TestPlugin_FetchProject(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	project, err := p.FetchProject(ctx, "any-id")
	if project != nil {
		t.Errorf("FetchProject() returned non-nil project")
	}
	if !errors.Is(err, plugin.ErrNotFound) {
		t.Errorf("FetchProject() error = %v, want ErrNotFound", err)
	}
}

func TestPlugin_FetchTasks(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projectID := "test-project"
	since := time.Now()

	tasks, err := p.FetchTasks(ctx, &projectID, &since)
	if err != nil {
		t.Errorf("FetchTasks() error = %v, want nil", err)
	}
	if len(tasks) != 0 {
		t.Errorf("FetchTasks() returned %d tasks, want 0", len(tasks))
	}
}

func TestPlugin_FetchTasks_NilParams(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	tasks, err := p.FetchTasks(ctx, nil, nil)
	if err != nil {
		t.Errorf("FetchTasks() with nil params error = %v, want nil", err)
	}
	if len(tasks) != 0 {
		t.Errorf("FetchTasks() returned %d tasks, want 0", len(tasks))
	}
}

func TestPlugin_FetchTask(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projectID := "test-project"

	task, err := p.FetchTask(ctx, &projectID, "task-123")
	if task != nil {
		t.Errorf("FetchTask() returned non-nil task")
	}
	if !errors.Is(err, plugin.ErrNotFound) {
		t.Errorf("FetchTask() error = %v, want ErrNotFound", err)
	}
}

func TestPlugin_CreateTask(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projectID := "test-project"
	taskCreate := &types.TaskCreate{
		Title:  "Test Task",
		Status: "active",
	}

	task, err := p.CreateTask(ctx, &projectID, taskCreate)
	if task != nil {
		t.Errorf("CreateTask() returned non-nil task")
	}
	if !errors.Is(err, plugin.ErrNotSupported) {
		t.Errorf("CreateTask() error = %v, want ErrNotSupported", err)
	}
}

func TestPlugin_UpdateTask(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projectID := "test-project"
	title := "Updated Title"
	taskUpdate := &types.TaskUpdate{
		Title: &title,
	}

	task, err := p.UpdateTask(ctx, &projectID, "task-123", taskUpdate)
	if task != nil {
		t.Errorf("UpdateTask() returned non-nil task")
	}
	if !errors.Is(err, plugin.ErrNotSupported) {
		t.Errorf("UpdateTask() error = %v, want ErrNotSupported", err)
	}
}

func TestPlugin_FetchComments(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projectID := "test-project"

	comments, err := p.FetchComments(ctx, &projectID, "task-123")
	if err != nil {
		t.Errorf("FetchComments() error = %v, want nil", err)
	}
	if len(comments) != 0 {
		t.Errorf("FetchComments() returned %d comments, want 0", len(comments))
	}
}

func TestPlugin_CreateComment(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()
	projectID := "test-project"
	commentCreate := &types.CommentCreate{
		Content: "Test comment",
		Author:  "test-user",
	}

	comment, err := p.CreateComment(ctx, &projectID, "task-123", commentCreate)
	if comment != nil {
		t.Errorf("CreateComment() returned non-nil comment")
	}
	if !errors.Is(err, plugin.ErrNotSupported) {
		t.Errorf("CreateComment() error = %v, want ErrNotSupported", err)
	}
}

func TestPlugin_ImplementsInterface(t *testing.T) {
	// Compile-time check that Plugin implements plugin.Plugin
	var _ plugin.Plugin = (*Plugin)(nil)
}
