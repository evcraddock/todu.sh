package plugin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestMockPluginImplementsInterface(t *testing.T) {
	var _ Plugin = (*MockPlugin)(nil)
}

func TestMockPluginName(t *testing.T) {
	mock := NewMockPlugin("test-system")
	if mock.Name() != "test-system" {
		t.Errorf("Expected name 'test-system', got '%s'", mock.Name())
	}
}

func TestMockPluginVersion(t *testing.T) {
	mock := NewMockPlugin("test-system")
	if mock.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", mock.Version())
	}
}

func TestMockPluginConfigure(t *testing.T) {
	mock := NewMockPlugin("test-system")
	config := map[string]string{
		"token": "test-token",
		"url":   "https://example.com",
	}

	err := mock.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMockPluginValidateConfig(t *testing.T) {
	mock := NewMockPlugin("test-system")

	// Should fail without token
	err := mock.ValidateConfig()
	if err == nil {
		t.Fatal("Expected error when token not configured")
	}
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("Expected ErrNotConfigured, got %v", err)
	}

	// Should succeed with token
	_ = mock.Configure(map[string]string{"token": "test-token"})
	err = mock.ValidateConfig()
	if err != nil {
		t.Errorf("Expected no error with token configured, got %v", err)
	}
}

func TestMockPluginFetchProjects(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	// Add a project
	project := &types.Project{
		ID:         1,
		Name:       "Test Project",
		ExternalID: "test-proj-1",
		Status:     "active",
	}
	mock.AddProject("test-proj-1", project)

	projects, err := mock.FetchProjects(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}

	if projects[0].Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", projects[0].Name)
	}
}

func TestMockPluginFetchProject(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	project := &types.Project{
		ID:         1,
		Name:       "Test Project",
		ExternalID: "test-proj-1",
		Status:     "active",
	}
	mock.AddProject("test-proj-1", project)

	// Fetch existing project
	fetched, err := mock.FetchProject(ctx, "test-proj-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if fetched.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", fetched.Name)
	}

	// Fetch non-existent project
	_, err = mock.FetchProject(ctx, "non-existent")
	if err == nil {
		t.Fatal("Expected error for non-existent project")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMockPluginCreateTask(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	taskCreate := &types.TaskCreate{
		Title:     "Test Task",
		ProjectID: 1,
		Status:    "open",
	}

	created, err := mock.CreateTask(ctx, nil, taskCreate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if created.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", created.Title)
	}
	if created.ExternalID == "" {
		t.Error("Expected external_id to be set")
	}
	if created.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestMockPluginUpdateTask(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	// Create a task first
	task := &types.Task{
		ID:         1,
		ExternalID: "task-1",
		Title:      "Original Title",
		ProjectID:  1,
		Status:     "open",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	mock.AddTask("task-1", task)

	// Update it
	newTitle := "Updated Title"
	taskUpdate := &types.TaskUpdate{
		Title: &newTitle,
	}

	updated, err := mock.UpdateTask(ctx, nil, "task-1", taskUpdate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}

	// Update non-existent task
	_, err = mock.UpdateTask(ctx, nil, "non-existent", taskUpdate)
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMockPluginFetchComments(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	// Add a task
	task := &types.Task{
		ID:         1,
		ExternalID: "task-1",
		Title:      "Test Task",
		ProjectID:  1,
		Status:     "open",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	mock.AddTask("task-1", task)

	// Fetch comments for task with no comments
	comments, err := mock.FetchComments(ctx, nil, "task-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(comments))
	}

	// Fetch comments for non-existent task
	_, err = mock.FetchComments(ctx, nil, "non-existent")
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMockPluginCreateComment(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	// Add a task
	task := &types.Task{
		ID:         1,
		ExternalID: "task-1",
		Title:      "Test Task",
		ProjectID:  1,
		Status:     "open",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	mock.AddTask("task-1", task)

	// Create a comment
	taskID := 1
	commentCreate := &types.CommentCreate{
		TaskID:  &taskID,
		Content: "Test comment",
		Author:  "testuser",
	}

	created, err := mock.CreateComment(ctx, nil, "task-1", commentCreate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if created.Content != "Test comment" {
		t.Errorf("Expected content 'Test comment', got '%s'", created.Content)
	}
	if created.Author != "testuser" {
		t.Errorf("Expected author 'testuser', got '%s'", created.Author)
	}

	// Verify comment is stored
	comments, _ := mock.FetchComments(ctx, nil, "task-1")
	if len(comments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(comments))
	}
}

func TestMockPluginErrorInjection(t *testing.T) {
	mock := NewMockPlugin("test-system")
	ctx := context.Background()

	testErr := errors.New("injected error")

	// Test FetchProjects error
	mock.FetchProjectsError = testErr
	_, err := mock.FetchProjects(ctx)
	if err != testErr {
		t.Errorf("Expected injected error, got %v", err)
	}

	// Test Configure error
	mock.ConfigureError = testErr
	err = mock.Configure(map[string]string{})
	if err != testErr {
		t.Errorf("Expected injected error, got %v", err)
	}
}

func TestMockPluginReset(t *testing.T) {
	mock := NewMockPlugin("test-system")

	// Add some data
	project := &types.Project{
		ID:         1,
		Name:       "Test Project",
		ExternalID: "test-proj-1",
		Status:     "active",
	}
	mock.AddProject("test-proj-1", project)

	// Verify data exists
	projects, _ := mock.FetchProjects(context.Background())
	if len(projects) != 1 {
		t.Errorf("Expected 1 project before reset, got %d", len(projects))
	}

	// Reset
	mock.Reset()

	// Verify data cleared
	projects, _ = mock.FetchProjects(context.Background())
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after reset, got %d", len(projects))
	}
}
