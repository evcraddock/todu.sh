// +build integration

package api

import (
	"context"
	"os"
	"testing"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// Run with: go test -tags=integration ./internal/api -v
// Requires API server running at http://localhost:8000

func getAPIURL() string {
	url := os.Getenv("TODU_API_URL")
	if url == "" {
		url = "http://localhost:8000"
	}
	return url
}

func TestIntegrationListProjects(t *testing.T) {
	client := NewClient(getAPIURL())
	projects, err := client.ListProjects(context.Background(), nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.Logf("Found %d projects", len(projects))
	if len(projects) > 0 {
		t.Logf("First project: ID=%d, Name=%s, Status=%s",
			projects[0].ID, projects[0].Name, projects[0].Status)
	}
}

func TestIntegrationListTasks(t *testing.T) {
	client := NewClient(getAPIURL())
	tasks, err := client.ListTasks(context.Background(), nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.Logf("Found %d tasks", len(tasks))
	if len(tasks) > 0 {
		t.Logf("First task: ID=%d, Title=%s, Status=%s, ProjectID=%d",
			tasks[0].ID, tasks[0].Title, tasks[0].Status, tasks[0].ProjectID)
	}
}

func TestIntegrationCreateSystem(t *testing.T) {
	client := NewClient(getAPIURL())

	create := &types.SystemCreate{
		Identifier: "test-integration",
		Name:       "Integration Test System",
	}

	system, err := client.CreateSystem(context.Background(), create)
	if err != nil {
		t.Fatalf("Expected no error creating system, got %v", err)
	}

	t.Logf("Created system: ID=%d, Identifier=%s, Name=%s",
		system.ID, system.Identifier, system.Name)

	// Clean up
	if err := client.DeleteSystem(context.Background(), system.ID); err != nil {
		t.Logf("Warning: Failed to delete test system: %v", err)
	}
}

func TestIntegrationGetProject(t *testing.T) {
	client := NewClient(getAPIURL())

	// List projects to get an ID
	projects, err := client.ListProjects(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error listing projects, got %v", err)
	}

	if len(projects) == 0 {
		t.Skip("No projects available for testing")
	}

	projectID := projects[0].ID

	// Get the specific project
	project, err := client.GetProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error getting project, got %v", err)
	}

	if project.ID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, project.ID)
	}

	t.Logf("Retrieved project: ID=%d, Name=%s, Status=%s",
		project.ID, project.Name, project.Status)
}

func TestIntegrationTasksWithFilter(t *testing.T) {
	client := NewClient(getAPIURL())

	// List projects to get a project ID
	projects, err := client.ListProjects(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error listing projects, got %v", err)
	}

	if len(projects) == 0 {
		t.Skip("No projects available for testing")
	}

	projectID := projects[0].ID

	// List tasks filtered by project
	tasks, err := client.ListTasks(context.Background(), &projectID)
	if err != nil {
		t.Fatalf("Expected no error listing tasks, got %v", err)
	}

	t.Logf("Found %d tasks for project %d", len(tasks), projectID)

	// Verify all tasks belong to the specified project
	for _, task := range tasks {
		if task.ProjectID != projectID {
			t.Errorf("Expected task to belong to project %d, got %d", projectID, task.ProjectID)
		}
	}
}
