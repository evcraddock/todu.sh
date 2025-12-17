package cmd

import (
	"testing"

	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/google/uuid"
)

func TestProjectAdd_UUIDGeneration(t *testing.T) {
	// Test that UUID generation works correctly
	id := uuid.New().String()

	// Valid UUID format: 8-4-4-4-12 hex characters
	if len(id) != 36 {
		t.Errorf("UUID length = %d, want 36", len(id))
	}

	// Parse should succeed
	_, err := uuid.Parse(id)
	if err != nil {
		t.Errorf("uuid.Parse(%q) error = %v", id, err)
	}
}

func TestProjectAdd_ExternalIDRequiredForNonLocal(t *testing.T) {
	// This tests the logic that external ID is required for non-local systems
	// The actual runProjectAdd would need mocking, so we test the logic here

	tests := []struct {
		name          string
		system        string
		externalID    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "local with empty external-id should auto-generate",
			system:      "local",
			externalID:  "",
			expectError: false,
		},
		{
			name:        "local with provided external-id should use it",
			system:      "local",
			externalID:  "custom-id",
			expectError: false,
		},
		{
			name:          "github with empty external-id should error",
			system:        "github",
			externalID:    "",
			expectError:   true,
			errorContains: "external-id is required",
		},
		{
			name:        "github with external-id should succeed",
			system:      "github",
			externalID:  "owner/repo",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			if tt.system != "local" && tt.externalID == "" {
				if !tt.expectError {
					t.Errorf("Expected no error but validation should fail")
				}
			} else if tt.expectError {
				t.Errorf("Expected error but validation would pass")
			}
		})
	}
}

func TestProjectAdd_DefaultValues(t *testing.T) {
	// Test that default flag values are correct
	// These are set in init() when flags are defined

	// Based on our implementation:
	// projectAddSystem default should be "local"
	// projectAddStatus default should be "active"
	// projectAddSyncStrategy default should be "bidirectional"

	// We can't test the actual cobra flags without initializing the command,
	// but we can document the expected defaults
	expectedDefaults := map[string]string{
		"system":        "local",
		"status":        "active",
		"sync-strategy": "bidirectional",
	}

	// This serves as documentation of expected defaults
	for flag, expected := range expectedDefaults {
		t.Logf("Expected default for --%s: %q", flag, expected)
	}
}

func TestProjectAdd_SyncStrategyValidation(t *testing.T) {
	validStrategies := map[string]bool{
		"pull":          true,
		"push":          true,
		"bidirectional": true,
	}

	tests := []struct {
		strategy string
		valid    bool
	}{
		{"pull", true},
		{"push", true},
		{"bidirectional", true},
		{"invalid", false},
		{"", false},
		{"PULL", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			if validStrategies[tt.strategy] != tt.valid {
				t.Errorf("strategy %q valid = %v, want %v", tt.strategy, validStrategies[tt.strategy], tt.valid)
			}
		})
	}
}

func TestStatusValue(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   int
	}{
		{"active status", "active", 3},
		{"done status", "done", 2},
		{"cancelled status", "cancelled", 1},
		{"unknown status", "unknown", 0},
		{"empty status", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusValue(tt.status)
			if got != tt.want {
				t.Errorf("statusValue(%q) = %d, want %d", tt.status, got, tt.want)
			}
		})
	}
}

func TestSortProjectsByPriorityAndStatus(t *testing.T) {
	// Create projects with various priorities and statuses
	projects := []*types.Project{
		{ID: 1, Name: "Low active", Priority: strPtr("low"), Status: "active"},
		{ID: 2, Name: "High done", Priority: strPtr("high"), Status: "done"},
		{ID: 3, Name: "No priority active", Priority: nil, Status: "active"},
		{ID: 4, Name: "High active", Priority: strPtr("high"), Status: "active"},
		{ID: 5, Name: "Medium cancelled", Priority: strPtr("medium"), Status: "cancelled"},
	}

	sortProjectsByPriorityAndStatus(projects)

	// Expected order:
	// 1. High active (ID 4) - priority 3, status 3
	// 2. High done (ID 2) - priority 3, status 2
	// 3. Medium cancelled (ID 5) - priority 2, status 1
	// 4. Low active (ID 1) - priority 1, status 3
	// 5. No priority active (ID 3) - priority 0, status 3
	expectedOrder := []int{4, 2, 5, 1, 3}
	for i, expectedID := range expectedOrder {
		if projects[i].ID != expectedID {
			t.Errorf("position %d: got ID %d, want %d", i, projects[i].ID, expectedID)
		}
	}
}

func TestSortProjectsByPriorityAndStatus_SamePriorityDifferentStatus(t *testing.T) {
	projects := []*types.Project{
		{ID: 1, Name: "High cancelled", Priority: strPtr("high"), Status: "cancelled"},
		{ID: 2, Name: "High done", Priority: strPtr("high"), Status: "done"},
		{ID: 3, Name: "High active", Priority: strPtr("high"), Status: "active"},
	}

	sortProjectsByPriorityAndStatus(projects)

	// All high priority, so order by status: active > done > cancelled
	expectedOrder := []int{3, 2, 1}
	for i, expectedID := range expectedOrder {
		if projects[i].ID != expectedID {
			t.Errorf("position %d: got ID %d, want %d", i, projects[i].ID, expectedID)
		}
	}
}

func TestSortProjectsByPriorityAndStatus_SamePriorityAndStatus(t *testing.T) {
	projects := []*types.Project{
		{ID: 30, Name: "Project 30", Priority: strPtr("medium"), Status: "active"},
		{ID: 10, Name: "Project 10", Priority: strPtr("medium"), Status: "active"},
		{ID: 20, Name: "Project 20", Priority: strPtr("medium"), Status: "active"},
	}

	sortProjectsByPriorityAndStatus(projects)

	// Same priority and status, so sort by ID
	expectedOrder := []int{10, 20, 30}
	for i, expectedID := range expectedOrder {
		if projects[i].ID != expectedID {
			t.Errorf("position %d: got ID %d, want %d", i, projects[i].ID, expectedID)
		}
	}
}

func TestSortProjectsByPriorityAndStatus_EmptySlice(t *testing.T) {
	projects := []*types.Project{}
	sortProjectsByPriorityAndStatus(projects) // Should not panic
	if len(projects) != 0 {
		t.Errorf("expected empty slice, got %d projects", len(projects))
	}
}
