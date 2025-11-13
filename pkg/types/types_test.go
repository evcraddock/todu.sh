package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskMarshalJSON(t *testing.T) {
	now := time.Now()
	sourceURL := "https://example.com/task/1"
	desc := "Test description"
	priority := "high"

	task := Task{
		ID:          1,
		ExternalID:  "ext-123",
		SourceURL:   &sourceURL,
		Title:       "Test Task",
		Description: &desc,
		ProjectID:   10,
		Status:      "open",
		Priority:    &priority,
		DueDate:     &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Labels: []Label{
			{ID: 1, Name: "bug"},
			{ID: 2, Name: "feature"},
		},
		Assignees: []Assignee{
			{ID: 1, Name: "John Doe"},
		},
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	var unmarshaled Task
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal task: %v", err)
	}

	if unmarshaled.ID != task.ID {
		t.Errorf("Expected ID %d, got %d", task.ID, unmarshaled.ID)
	}
	if unmarshaled.ExternalID != task.ExternalID {
		t.Errorf("Expected ExternalID %s, got %s", task.ExternalID, unmarshaled.ExternalID)
	}
	if unmarshaled.Title != task.Title {
		t.Errorf("Expected Title %s, got %s", task.Title, unmarshaled.Title)
	}
	if *unmarshaled.Description != *task.Description {
		t.Errorf("Expected Description %s, got %s", *task.Description, *unmarshaled.Description)
	}
	if unmarshaled.Status != task.Status {
		t.Errorf("Expected Status %s, got %s", task.Status, unmarshaled.Status)
	}
	if len(unmarshaled.Labels) != len(task.Labels) {
		t.Errorf("Expected %d labels, got %d", len(task.Labels), len(unmarshaled.Labels))
	}
	if len(unmarshaled.Assignees) != len(task.Assignees) {
		t.Errorf("Expected %d assignees, got %d", len(task.Assignees), len(unmarshaled.Assignees))
	}
}

func TestTaskCreateMarshalJSON(t *testing.T) {
	now := time.Now()
	sourceURL := "https://example.com/task/1"
	desc := "Test description"
	priority := "high"

	taskCreate := TaskCreate{
		ExternalID:  "ext-123",
		SourceURL:   &sourceURL,
		Title:       "Test Task",
		Description: &desc,
		ProjectID:   10,
		Status:      "open",
		Priority:    &priority,
		DueDate:     &now,
		Labels: []string{
			"bug",
		},
		Assignees: []string{
			"John Doe",
		},
	}

	data, err := json.Marshal(taskCreate)
	if err != nil {
		t.Fatalf("Failed to marshal task create: %v", err)
	}

	var unmarshaled TaskCreate
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal task create: %v", err)
	}

	if unmarshaled.ExternalID != taskCreate.ExternalID {
		t.Errorf("Expected ExternalID %s, got %s", taskCreate.ExternalID, unmarshaled.ExternalID)
	}
	if unmarshaled.Title != taskCreate.Title {
		t.Errorf("Expected Title %s, got %s", taskCreate.Title, unmarshaled.Title)
	}
}

func TestTaskUpdateMarshalJSON(t *testing.T) {
	title := "Updated Title"
	status := "closed"
	priority := "low"

	taskUpdate := TaskUpdate{
		Title:    &title,
		Status:   &status,
		Priority: &priority,
	}

	data, err := json.Marshal(taskUpdate)
	if err != nil {
		t.Fatalf("Failed to marshal task update: %v", err)
	}

	var unmarshaled TaskUpdate
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal task update: %v", err)
	}

	if *unmarshaled.Title != *taskUpdate.Title {
		t.Errorf("Expected Title %s, got %s", *taskUpdate.Title, *unmarshaled.Title)
	}
	if *unmarshaled.Status != *taskUpdate.Status {
		t.Errorf("Expected Status %s, got %s", *taskUpdate.Status, *unmarshaled.Status)
	}
}

func TestProjectMarshalJSON(t *testing.T) {
	now := time.Now()
	desc := "Test project description"

	project := Project{
		ID:          1,
		Name:        "Test Project",
		Description: &desc,
		SystemID:    5,
		ExternalID:  "proj-123",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("Failed to marshal project: %v", err)
	}

	var unmarshaled Project
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal project: %v", err)
	}

	if unmarshaled.ID != project.ID {
		t.Errorf("Expected ID %d, got %d", project.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != project.Name {
		t.Errorf("Expected Name %s, got %s", project.Name, unmarshaled.Name)
	}
	if *unmarshaled.Description != *project.Description {
		t.Errorf("Expected Description %s, got %s", *project.Description, *unmarshaled.Description)
	}
	if unmarshaled.SystemID != project.SystemID {
		t.Errorf("Expected SystemID %d, got %d", project.SystemID, unmarshaled.SystemID)
	}
	if unmarshaled.ExternalID != project.ExternalID {
		t.Errorf("Expected ExternalID %s, got %s", project.ExternalID, unmarshaled.ExternalID)
	}
	if unmarshaled.Status != project.Status {
		t.Errorf("Expected Status %s, got %s", project.Status, unmarshaled.Status)
	}
}

func TestProjectCreateMarshalJSON(t *testing.T) {
	desc := "Test project description"

	projectCreate := ProjectCreate{
		Name:        "Test Project",
		Description: &desc,
		SystemID:    5,
		ExternalID:  "proj-123",
		Status:      "active",
	}

	data, err := json.Marshal(projectCreate)
	if err != nil {
		t.Fatalf("Failed to marshal project create: %v", err)
	}

	var unmarshaled ProjectCreate
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal project create: %v", err)
	}

	if unmarshaled.Name != projectCreate.Name {
		t.Errorf("Expected Name %s, got %s", projectCreate.Name, unmarshaled.Name)
	}
	if unmarshaled.ExternalID != projectCreate.ExternalID {
		t.Errorf("Expected ExternalID %s, got %s", projectCreate.ExternalID, unmarshaled.ExternalID)
	}
}

func TestProjectUpdateMarshalJSON(t *testing.T) {
	name := "Updated Project Name"
	status := "archived"

	projectUpdate := ProjectUpdate{
		Name:   &name,
		Status: &status,
	}

	data, err := json.Marshal(projectUpdate)
	if err != nil {
		t.Fatalf("Failed to marshal project update: %v", err)
	}

	var unmarshaled ProjectUpdate
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal project update: %v", err)
	}

	if *unmarshaled.Name != *projectUpdate.Name {
		t.Errorf("Expected Name %s, got %s", *projectUpdate.Name, *unmarshaled.Name)
	}
	if *unmarshaled.Status != *projectUpdate.Status {
		t.Errorf("Expected Status %s, got %s", *projectUpdate.Status, *unmarshaled.Status)
	}
}

func TestSystemMarshalJSON(t *testing.T) {
	now := time.Now()
	url := "https://api.example.com"
	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	system := System{
		ID:         1,
		Identifier: "test-system",
		Name:       "Test System",
		URL:        &url,
		Metadata:   metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	data, err := json.Marshal(system)
	if err != nil {
		t.Fatalf("Failed to marshal system: %v", err)
	}

	var unmarshaled System
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal system: %v", err)
	}

	if unmarshaled.ID != system.ID {
		t.Errorf("Expected ID %d, got %d", system.ID, unmarshaled.ID)
	}
	if unmarshaled.Identifier != system.Identifier {
		t.Errorf("Expected Identifier %s, got %s", system.Identifier, unmarshaled.Identifier)
	}
	if unmarshaled.Name != system.Name {
		t.Errorf("Expected Name %s, got %s", system.Name, unmarshaled.Name)
	}
	if len(unmarshaled.Metadata) != len(system.Metadata) {
		t.Errorf("Expected %d metadata entries, got %d", len(system.Metadata), len(unmarshaled.Metadata))
	}
}

func TestJSONTagCorrectness(t *testing.T) {
	// Test that JSON tags are using snake_case
	task := Task{
		ID:         1,
		ExternalID: "ext-123",
		Title:      "Test",
		ProjectID:  10,
		Status:     "open",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	// Unmarshal into a map to check the actual JSON field names
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal into map: %v", err)
	}

	// Verify snake_case fields exist
	requiredFields := []string{"id", "external_id", "title", "project_id", "status", "created_at", "updated_at"}
	for _, field := range requiredFields {
		if _, ok := jsonMap[field]; !ok {
			t.Errorf("Expected JSON field %s to exist", field)
		}
	}
}
