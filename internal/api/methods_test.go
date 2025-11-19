package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// System Tests

func TestListSystems(t *testing.T) {
	now := time.Now()
	systems := []*types.System{
		{
			ID:         1,
			Identifier: "github",
			Name:       "GitHub",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         2,
			Identifier: "jira",
			Name:       "Jira",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/systems/" {
			t.Errorf("Expected path '/api/v1/systems/', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got '%s'", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(systems)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ListSystems(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 systems, got %d", len(result))
	}
	if result[0].Identifier != "github" {
		t.Errorf("Expected first system identifier to be 'github', got '%s'", result[0].Identifier)
	}
}

func TestGetSystem(t *testing.T) {
	now := time.Now()
	system := &types.System{
		ID:         1,
		Identifier: "github",
		Name:       "GitHub",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/systems/1" {
			t.Errorf("Expected path '/api/v1/systems/1', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got '%s'", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(system)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetSystem(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Errorf("Expected system ID 1, got %d", result.ID)
	}
	if result.Identifier != "github" {
		t.Errorf("Expected identifier 'github', got '%s'", result.Identifier)
	}
}

func TestGetSystem404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("system not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetSystem(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for 404 response")
	}
}

func TestCreateSystem(t *testing.T) {
	now := time.Now()
	create := &types.SystemCreate{
		Identifier: "github",
		Name:       "GitHub",
	}
	created := &types.System{
		ID:         1,
		Identifier: "github",
		Name:       "GitHub",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/systems/" {
			t.Errorf("Expected path '/api/v1/systems/', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got '%s'", r.Method)
		}

		var received types.SystemCreate
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if received.Identifier != "github" {
			t.Errorf("Expected identifier 'github', got '%s'", received.Identifier)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.CreateSystem(context.Background(), create)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.Identifier != "github" {
		t.Errorf("Expected identifier 'github', got '%s'", result.Identifier)
	}
}

func TestUpdateSystem(t *testing.T) {
	now := time.Now()
	newName := "GitHub Enterprise"
	update := &types.SystemUpdate{
		Name: &newName,
	}
	updated := &types.System{
		ID:         1,
		Identifier: "github",
		Name:       "GitHub Enterprise",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/systems/1" {
			t.Errorf("Expected path '/api/v1/systems/1', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got '%s'", r.Method)
		}

		var received types.SystemUpdate
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if received.Name == nil || *received.Name != "GitHub Enterprise" {
			t.Errorf("Expected name 'GitHub Enterprise'")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.UpdateSystem(context.Background(), 1, update)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Name != "GitHub Enterprise" {
		t.Errorf("Expected name 'GitHub Enterprise', got '%s'", result.Name)
	}
}

func TestDeleteSystem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/systems/1" {
			t.Errorf("Expected path '/api/v1/systems/1', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteSystem(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// Project Tests

func TestListProjects(t *testing.T) {
	now := time.Now()
	projects := []*types.Project{
		{
			ID:           1,
			Name:         "My Project",
			SystemID:     1,
			ExternalID:   "owner/repo",
			Status:       "active",
			SyncStrategy: "pull",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/" {
			t.Errorf("Expected path '/api/v1/projects/', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got '%s'", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ListProjects(context.Background(), nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 project, got %d", len(result))
	}
}

func TestListProjectsWithFilter(t *testing.T) {
	now := time.Now()
	systemID := 1
	projects := []*types.Project{
		{
			ID:           1,
			Name:         "My Project",
			SystemID:     1,
			ExternalID:   "owner/repo",
			Status:       "active",
			SyncStrategy: "pull",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/projects/?system_id=1"
		if r.URL.Path+"?"+r.URL.RawQuery != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path+"?"+r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ListProjects(context.Background(), &systemID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 project, got %d", len(result))
	}
}

func TestGetProject(t *testing.T) {
	now := time.Now()
	project := &types.Project{
		ID:           1,
		Name:         "My Project",
		SystemID:     1,
		ExternalID:   "owner/repo",
		Status:       "active",
		SyncStrategy: "pull",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/1" {
			t.Errorf("Expected path '/api/v1/projects/1', got '%s'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetProject(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Name != "My Project" {
		t.Errorf("Expected name 'My Project', got '%s'", result.Name)
	}
}

func TestCreateProject(t *testing.T) {
	now := time.Now()
	create := &types.ProjectCreate{
		Name:         "My Project",
		SystemID:     1,
		ExternalID:   "owner/repo",
		Status:       "active",
		SyncStrategy: "pull",
	}
	created := &types.Project{
		ID:           1,
		Name:         "My Project",
		SystemID:     1,
		ExternalID:   "owner/repo",
		Status:       "active",
		SyncStrategy: "pull",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got '%s'", r.Method)
		}

		var received types.ProjectCreate
		json.NewDecoder(r.Body).Decode(&received)
		if received.Name != "My Project" {
			t.Errorf("Expected name 'My Project', got '%s'", received.Name)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.CreateProject(context.Background(), create)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestUpdateProject(t *testing.T) {
	now := time.Now()
	newName := "Updated Project"
	update := &types.ProjectUpdate{
		Name: &newName,
	}
	updated := &types.Project{
		ID:           1,
		Name:         "Updated Project",
		SystemID:     1,
		ExternalID:   "owner/repo",
		Status:       "active",
		SyncStrategy: "pull",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got '%s'", r.Method)
		}
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.UpdateProject(context.Background(), 1, update)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got '%s'", result.Name)
	}
}

func TestDeleteProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteProject(context.Background(), 1, false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// Task Tests

func TestListTasks(t *testing.T) {
	now := time.Now()
	tasksResp := TasksResponse{
		Items: []*types.Task{
			{
				ID:          1,
				ExternalID:  "123",
				Title:       "Test Task",
				ProjectID:   1,
				Status:      "open",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Total: 1,
		Skip:  0,
		Limit: 50,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasksResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ListTasks(context.Background(), nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 task, got %d", len(result))
	}
}

func TestListTasksWithFilter(t *testing.T) {
	now := time.Now()
	projectID := 1
	tasksResp := TasksResponse{
		Items: []*types.Task{
			{
				ID:          1,
				ExternalID:  "123",
				Title:       "Test Task",
				ProjectID:   1,
				Status:      "open",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Total: 1,
		Skip:  0,
		Limit: 50,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that project_id is present in query (limit is also added by default)
		if !strings.Contains(r.URL.RawQuery, "project_id=1") {
			t.Errorf("Expected query to contain 'project_id=1', got '%s'", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasksResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	opts := &TaskListOptions{
		ProjectID: &projectID,
	}
	result, err := client.ListTasks(context.Background(), opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 task, got %d", len(result))
	}
}

func TestGetTask(t *testing.T) {
	now := time.Now()
	task := &types.Task{
		ID:          1,
		ExternalID:  "123",
		Title:       "Test Task",
		ProjectID:   1,
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetTask(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", result.Title)
	}
}

func TestCreateTask(t *testing.T) {
	now := time.Now()
	create := &types.TaskCreate{
		ExternalID:  "123",
		Title:       "Test Task",
		ProjectID:   1,
		Status:      "open",
	}
	created := &types.Task{
		ID:          1,
		ExternalID:  "123",
		Title:       "Test Task",
		ProjectID:   1,
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.CreateTask(context.Background(), create)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestUpdateTask(t *testing.T) {
	now := time.Now()
	newTitle := "Updated Task"
	update := &types.TaskUpdate{
		Title: &newTitle,
	}
	updated := &types.Task{
		ID:          1,
		ExternalID:  "123",
		Title:       "Updated Task",
		ProjectID:   1,
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got '%s'", r.Method)
		}
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.UpdateTask(context.Background(), 1, update)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Title != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got '%s'", result.Title)
	}
}

func TestDeleteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteTask(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// Comment Tests

func TestListComments(t *testing.T) {
	now := time.Now()
	comments := []*types.Comment{
		{
			ID:        1,
			TaskID:    1,
			Content:   "Test comment",
			Author:    "user1",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tasks/1/comments" {
			t.Errorf("Expected path '/api/v1/tasks/1/comments', got '%s'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ListComments(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(result))
	}
}

func TestGetComment(t *testing.T) {
	now := time.Now()
	comment := &types.Comment{
		ID:        1,
		TaskID:    1,
		Content:   "Test comment",
		Author:    "user1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/comments/1" {
			t.Errorf("Expected path '/api/v1/comments/1', got '%s'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comment)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetComment(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Content != "Test comment" {
		t.Errorf("Expected content 'Test comment', got '%s'", result.Content)
	}
}

func TestCreateComment(t *testing.T) {
	now := time.Now()
	create := &types.CommentCreate{
		TaskID:  1,
		Content: "Test comment",
		Author:  "user1",
	}
	created := &types.Comment{
		ID:        1,
		TaskID:    1,
		Content:   "Test comment",
		Author:    "user1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got '%s'", r.Method)
		}

		var received types.CommentCreate
		json.NewDecoder(r.Body).Decode(&received)
		if received.Content != "Test comment" {
			t.Errorf("Expected content 'Test comment', got '%s'", received.Content)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.CreateComment(context.Background(), create)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestDeleteComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/comments/1" {
			t.Errorf("Expected path '/api/v1/comments/1', got '%s'", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteComment(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// Error Handling Tests

func TestServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.ListSystems(context.Background())

	if err == nil {
		t.Fatal("Expected error for 500 response")
	}
}

func TestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.ListSystems(context.Background())

	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}
