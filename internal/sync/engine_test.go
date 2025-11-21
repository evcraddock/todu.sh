package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

// sharedMockPlugin is a shared instance used across test invocations
var sharedMockPlugin *plugin.MockPlugin

// setupTestEngine creates a test engine with a mock API server and plugin.
func setupTestEngine(t *testing.T) (*Engine, *httptest.Server, *plugin.MockPlugin) {
	// Set environment variable for plugin config
	//Note: test-system becomes TEST-SYSTEM (preserves hyphens)
	t.Setenv("TODU_PLUGIN_TEST-SYSTEM_TOKEN", "test-token")

	// Initialize shared mock plugin
	if sharedMockPlugin == nil {
		sharedMockPlugin = plugin.NewMockPlugin("test-system")
	} else {
		sharedMockPlugin.Reset()
	}

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMockAPI(t, w, r)
	}))

	// Create API client
	apiClient := api.NewClient(server.URL)

	// Create registry and register mock plugin factory that returns the shared instance
	reg := registry.New()
	reg.Register("test-system", func() plugin.Plugin {
		return sharedMockPlugin
	})

	engine := NewEngine(apiClient, reg)
	return engine, server, sharedMockPlugin
}

// handleMockAPI handles mock API requests for testing.
func handleMockAPI(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "GET" && r.URL.Path == "/api/v1/projects/":
		// List projects
		projects := []*types.Project{
			{
				ID:           1,
				Name:         "Test Project",
				SystemID:     1,
				ExternalID:   "test-repo",
				Status:       "active",
				SyncStrategy: "pull",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}
		json.NewEncoder(w).Encode(projects)

	case r.Method == "GET" && r.URL.Path == "/api/v1/projects/1":
		// Get project
		project := &types.Project{
			ID:           1,
			Name:         "Test Project",
			SystemID:     1,
			ExternalID:   "test-repo",
			Status:       "active",
			SyncStrategy: "pull",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		json.NewEncoder(w).Encode(project)

	case r.Method == "GET" && r.URL.Path == "/api/v1/systems/1":
		// Get system
		system := &types.System{
			ID:         1,
			Identifier: "test-system",
			Name:       "Test System",
			URL:        stringPtr("http://test.example.com"),
			Metadata:   map[string]string{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		json.NewEncoder(w).Encode(system)

	case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/":
		// List tasks
		tasksResp := &api.TasksResponse{
			Items: []*types.Task{},
			Total: 0,
			Skip:  0,
			Limit: 100,
		}
		json.NewEncoder(w).Encode(tasksResp)

	case r.Method == "POST" && r.URL.Path == "/api/v1/tasks/":
		// Create task
		var taskCreate types.TaskCreate
		if err := json.NewDecoder(r.Body).Decode(&taskCreate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Convert string labels to Label structs
		labels := make([]types.Label, len(taskCreate.Labels))
		for i, name := range taskCreate.Labels {
			labels[i] = types.Label{Name: name}
		}

		// Convert string assignees to Assignee structs
		assignees := make([]types.Assignee, len(taskCreate.Assignees))
		for i, name := range taskCreate.Assignees {
			assignees[i] = types.Assignee{Name: name}
		}

		task := &types.Task{
			ID:          1,
			ExternalID:  taskCreate.ExternalID,
			SourceURL:   taskCreate.SourceURL,
			Title:       taskCreate.Title,
			Description: taskCreate.Description,
			ProjectID:   taskCreate.ProjectID,
			Status:      taskCreate.Status,
			Priority:    taskCreate.Priority,
			DueDate:     taskCreate.DueDate,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Labels:      labels,
			Assignees:   assignees,
		}
		json.NewEncoder(w).Encode(task)

	case r.Method == "PUT" && r.URL.Path == "/api/v1/tasks/1":
		// Update task
		var taskUpdate types.TaskUpdate
		if err := json.NewDecoder(r.Body).Decode(&taskUpdate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Convert string labels to Label structs
		labels := make([]types.Label, len(taskUpdate.Labels))
		for i, name := range taskUpdate.Labels {
			labels[i] = types.Label{Name: name}
		}

		// Convert string assignees to Assignee structs
		assignees := make([]types.Assignee, len(taskUpdate.Assignees))
		for i, name := range taskUpdate.Assignees {
			assignees[i] = types.Assignee{Name: name}
		}

		task := &types.Task{
			ID:          1,
			ExternalID:  "task-1",
			Title:       *taskUpdate.Title,
			Description: taskUpdate.Description,
			ProjectID:   1,
			Status:      *taskUpdate.Status,
			Priority:    taskUpdate.Priority,
			DueDate:     taskUpdate.DueDate,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now(),
			Labels:      labels,
			Assignees:   assignees,
		}
		json.NewEncoder(w).Encode(task)

	case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/1/comments":
		// List comments for task
		comments := []*types.Comment{}
		json.NewEncoder(w).Encode(comments)

	case r.Method == "POST" && r.URL.Path == "/api/v1/tasks/1/comments":
		// Create comment
		var commentCreate types.CommentCreate
		if err := json.NewDecoder(r.Body).Decode(&commentCreate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		comment := &types.Comment{
			ID:         1,
			TaskID:     commentCreate.TaskID,
			ExternalID: commentCreate.ExternalID,
			Content:    commentCreate.Content,
			Author:     commentCreate.Author,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		json.NewEncoder(w).Encode(comment)

	case r.Method == "PUT" && r.URL.Path == "/api/v1/comments/1":
		// Update comment
		var commentUpdate types.CommentUpdate
		if err := json.NewDecoder(r.Body).Decode(&commentUpdate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		taskID := 1
		comment := &types.Comment{
			ID:         1,
			TaskID:     &taskID,
			ExternalID: *commentUpdate.ExternalID,
			Content:    "Test comment",
			Author:     "test-user",
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now(),
		}
		json.NewEncoder(w).Encode(comment)

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestNewEngine(t *testing.T) {
	apiClient := api.NewClient("http://test.example.com")
	reg := registry.New()

	engine := NewEngine(apiClient, reg)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}
	if engine.apiClient != apiClient {
		t.Error("Expected apiClient to be set")
	}
	if engine.registry != reg {
		t.Error("Expected registry to be set")
	}
}

func TestSyncPullStrategy(t *testing.T) {
	engine, server, mockPlugin := setupTestEngine(t)
	defer server.Close()

	// Add a project to the mock plugin (needed for task lookup)
	mockPlugin.AddProject("test-repo", &types.Project{
		ID:         1,
		ExternalID: "test-repo",
		Name:       "Test Project",
	})

	// Add a task to the mock plugin
	now := time.Now()
	mockPlugin.AddTask("task-1", &types.Task{
		ID:          1,
		ExternalID:  "task-1",
		Title:       "Test Task",
		Description: stringPtr("Test description"),
		ProjectID:   1,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	ctx := context.Background()
	options := Options{
		ProjectIDs: []int{1},
		DryRun:     false,
	}

	result, err := engine.Sync(ctx, options)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.TotalCreated != 1 {
		t.Errorf("Expected 1 task created, got %d", result.TotalCreated)
	}
	if result.TotalErrors != 0 {
		t.Errorf("Expected 0 errors, got %d", result.TotalErrors)
		if len(result.ProjectResults) > 0 {
			for _, e := range result.ProjectResults[0].Errors {
				t.Logf("  Error: %v", e)
			}
		}
	}
	if len(result.ProjectResults) != 1 {
		t.Errorf("Expected 1 project result, got %d", len(result.ProjectResults))
	}
}

func TestSyncPullStrategyUpdate(t *testing.T) {
	engine, server, mockPlugin := setupTestEngine(t)
	defer server.Close()

	// Add a project to the mock plugin
	mockPlugin.AddProject("test-repo", &types.Project{
		ID:         1,
		ExternalID: "test-repo",
		Name:       "Test Project",
	})

	// Add a task to the mock plugin with newer timestamp
	now := time.Now()
	mockPlugin.AddTask("task-1", &types.Task{
		ID:          1,
		ExternalID:  "task-1",
		Title:       "Updated Task",
		Description: stringPtr("Updated description"),
		ProjectID:   1,
		Status:      "active",
		CreatedAt:   now.Add(-1 * time.Hour),
		UpdatedAt:   now,
	})

	ctx := context.Background()
	options := Options{
		ProjectIDs: []int{1},
		DryRun:     false,
	}

	result, err := engine.Sync(ctx, options)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.TotalCreated != 1 {
		t.Errorf("Expected 1 task created, got %d", result.TotalCreated)
	}
}

func TestSyncDryRun(t *testing.T) {
	engine, server, mockPlugin := setupTestEngine(t)
	defer server.Close()

	// Add a project to the mock plugin
	mockPlugin.AddProject("test-repo", &types.Project{
		ID:         1,
		ExternalID: "test-repo",
		Name:       "Test Project",
	})

	// Add a task to the mock plugin
	now := time.Now()
	mockPlugin.AddTask("task-1", &types.Task{
		ID:          1,
		ExternalID:  "task-1",
		Title:       "Test Task",
		Description: stringPtr("Test description"),
		ProjectID:   1,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	ctx := context.Background()
	options := Options{
		ProjectIDs: []int{1},
		DryRun:     true,
	}

	result, err := engine.Sync(ctx, options)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// In dry run mode, tasks should be counted but not actually created
	if result.TotalCreated != 1 {
		t.Errorf("Expected 1 task to be counted in dry run, got %d", result.TotalCreated)
	}
}

func TestSyncStrategyOverride(t *testing.T) {
	engine, server, mockPlugin := setupTestEngine(t)
	defer server.Close()

	// Add a project to the mock plugin
	mockPlugin.AddProject("test-repo", &types.Project{
		ID:         1,
		ExternalID: "test-repo",
		Name:       "Test Project",
	})

	// Add a task to the mock plugin
	now := time.Now()
	mockPlugin.AddTask("task-1", &types.Task{
		ID:          1,
		ExternalID:  "task-1",
		Title:       "Test Task",
		Description: stringPtr("Test description"),
		ProjectID:   1,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	ctx := context.Background()
	strategyOverride := StrategyPull
	options := Options{
		ProjectIDs:       []int{1},
		StrategyOverride: &strategyOverride,
		DryRun:           false,
	}

	result, err := engine.Sync(ctx, options)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.TotalCreated != 1 {
		t.Errorf("Expected 1 task created, got %d", result.TotalCreated)
	}
}

func TestSyncNoProjects(t *testing.T) {
	engine, server, _ := setupTestEngine(t)
	defer server.Close()

	ctx := context.Background()
	options := Options{
		ProjectIDs: []int{999}, // Non-existent project
		DryRun:     false,
	}

	_, err := engine.Sync(ctx, options)
	if err == nil {
		t.Error("Expected error for non-existent project")
	}
}

func TestDetermineStrategy(t *testing.T) {
	engine := &Engine{}

	tests := []struct {
		name             string
		projectStrategy  string
		strategyOverride *Strategy
		expected         Strategy
	}{
		{
			name:             "Use project strategy when no override",
			projectStrategy:  "pull",
			strategyOverride: nil,
			expected:         StrategyPull,
		},
		{
			name:            "Use override when provided",
			projectStrategy: "pull",
			strategyOverride: func() *Strategy {
				s := StrategyBidirectional
				return &s
			}(),
			expected: StrategyBidirectional,
		},
		{
			name:             "Use project bidirectional strategy",
			projectStrategy:  "bidirectional",
			strategyOverride: nil,
			expected:         StrategyBidirectional,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &types.Project{
				SyncStrategy: tt.projectStrategy,
			}
			options := Options{
				StrategyOverride: tt.strategyOverride,
			}

			result := engine.determineStrategy(project, options)
			if result != tt.expected {
				t.Errorf("Expected strategy %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSyncWithPluginError(t *testing.T) {
	engine, server, mockPlugin := setupTestEngine(t)
	defer server.Close()

	// Inject error into plugin
	mockPlugin.FetchTasksError = plugin.NewErrNotFound("test error")

	ctx := context.Background()
	options := Options{
		ProjectIDs: []int{1},
		DryRun:     false,
	}

	result, err := engine.Sync(ctx, options)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("Expected errors to be recorded")
	}
	if len(result.ProjectResults) == 0 {
		t.Error("Expected project results even with errors")
	}
	if len(result.ProjectResults[0].Errors) == 0 {
		t.Error("Expected errors in project result")
	}
}

func TestResultAddProjectResult(t *testing.T) {
	result := &Result{}

	pr := ProjectResult{
		ProjectID:   1,
		ProjectName: "Test Project",
		Created:     5,
		Updated:     3,
		Skipped:     2,
		Errors:      []error{},
	}

	result.AddProjectResult(pr)

	if result.TotalCreated != 5 {
		t.Errorf("Expected TotalCreated to be 5, got %d", result.TotalCreated)
	}
	if result.TotalUpdated != 3 {
		t.Errorf("Expected TotalUpdated to be 3, got %d", result.TotalUpdated)
	}
	if result.TotalSkipped != 2 {
		t.Errorf("Expected TotalSkipped to be 2, got %d", result.TotalSkipped)
	}
	if len(result.ProjectResults) != 1 {
		t.Errorf("Expected 1 project result, got %d", len(result.ProjectResults))
	}
}

func TestResultHasErrors(t *testing.T) {
	result := &Result{
		TotalErrors: 0,
	}

	if result.HasErrors() {
		t.Error("Expected HasErrors to return false when TotalErrors is 0")
	}

	result.TotalErrors = 1
	if !result.HasErrors() {
		t.Error("Expected HasErrors to return true when TotalErrors > 0")
	}
}

func TestStrategyIsValid(t *testing.T) {
	tests := []struct {
		strategy Strategy
		valid    bool
	}{
		{StrategyPull, true},
		{StrategyPush, true},
		{StrategyBidirectional, true},
		{Strategy("invalid"), false},
		{Strategy(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			if tt.strategy.IsValid() != tt.valid {
				t.Errorf("Expected IsValid() to return %v for %s", tt.valid, tt.strategy)
			}
		})
	}
}
