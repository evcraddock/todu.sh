//go:build integration
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

	url := "https://example.com"
	create := &types.SystemCreate{
		Identifier: "test-integration",
		Name:       "Integration Test System",
		URL:        &url,
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
	tasks, err := client.ListTasks(context.Background(), &TaskListOptions{ProjectID: &projectID})
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

// Template Integration Tests

func TestIntegrationListTemplates(t *testing.T) {
	client := NewClient(getAPIURL())
	templates, err := client.ListTemplates(context.Background(), nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.Logf("Found %d templates", len(templates))
	if len(templates) > 0 {
		t.Logf("First template: ID=%d, Title=%s, Type=%s, Active=%v",
			templates[0].ID, templates[0].Title, templates[0].TemplateType, templates[0].IsActive)
	}
}

func TestIntegrationTemplateLifecycle(t *testing.T) {
	client := NewClient(getAPIURL())
	ctx := context.Background()

	// List projects to get a project ID
	projects, err := client.ListProjects(ctx, nil)
	if err != nil {
		t.Fatalf("Expected no error listing projects, got %v", err)
	}

	if len(projects) == 0 {
		t.Skip("No projects available for testing")
	}

	projectID := projects[0].ID

	// Create a template
	create := &types.RecurringTaskTemplateCreate{
		ProjectID:      projectID,
		Title:          "Integration Test Template",
		RecurrenceRule: "FREQ=DAILY",
		StartDate:      "2024-01-01",
		Timezone:       "UTC",
		TemplateType:   "task",
		IsActive:       true,
	}

	template, err := client.CreateTemplate(ctx, create)
	if err != nil {
		t.Fatalf("Expected no error creating template, got %v", err)
	}

	t.Logf("Created template: ID=%d, Title=%s", template.ID, template.Title)

	// Verify the template was created correctly
	if template.Title != "Integration Test Template" {
		t.Errorf("Expected title 'Integration Test Template', got '%s'", template.Title)
	}
	if template.RecurrenceRule != "FREQ=DAILY" {
		t.Errorf("Expected recurrence 'FREQ=DAILY', got '%s'", template.RecurrenceRule)
	}
	if !template.IsActive {
		t.Errorf("Expected template to be active")
	}

	// Get the template
	fetched, err := client.GetTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("Expected no error getting template, got %v", err)
	}

	if fetched.ID != template.ID {
		t.Errorf("Expected template ID %d, got %d", template.ID, fetched.ID)
	}

	// Update the template
	newTitle := "Updated Integration Test Template"
	update := &types.RecurringTaskTemplateUpdate{
		Title: &newTitle,
	}

	updated, err := client.UpdateTemplate(ctx, template.ID, update)
	if err != nil {
		t.Fatalf("Expected no error updating template, got %v", err)
	}

	if updated.Title != newTitle {
		t.Errorf("Expected title '%s', got '%s'", newTitle, updated.Title)
	}

	// Deactivate the template
	active := false
	deactivate := &types.RecurringTaskTemplateUpdate{
		IsActive: &active,
	}

	deactivated, err := client.UpdateTemplate(ctx, template.ID, deactivate)
	if err != nil {
		t.Fatalf("Expected no error deactivating template, got %v", err)
	}

	if deactivated.IsActive {
		t.Errorf("Expected template to be inactive")
	}

	// Delete the template
	err = client.DeleteTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("Expected no error deleting template, got %v", err)
	}

	t.Logf("Successfully completed template lifecycle test")
}

func TestIntegrationListTemplatesWithFilters(t *testing.T) {
	client := NewClient(getAPIURL())
	ctx := context.Background()

	// Test filtering by active status
	active := true
	templates, err := client.ListTemplates(ctx, &TemplateListOptions{
		Active: &active,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.Logf("Found %d active templates", len(templates))

	// Verify all returned templates are active
	for _, tmpl := range templates {
		if !tmpl.IsActive {
			t.Errorf("Expected all templates to be active, but ID=%d is inactive", tmpl.ID)
		}
	}

	// Test filtering by template type
	templates, err = client.ListTemplates(ctx, &TemplateListOptions{
		TemplateType: "habit",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.Logf("Found %d habit templates", len(templates))

	// Verify all returned templates are habits
	for _, tmpl := range templates {
		if tmpl.TemplateType != "habit" {
			t.Errorf("Expected all templates to be habits, but ID=%d is '%s'", tmpl.ID, tmpl.TemplateType)
		}
	}
}
