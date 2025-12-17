package cmd

import (
	"testing"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestPriorityValue(t *testing.T) {
	tests := []struct {
		name     string
		priority *string
		want     int
	}{
		{"nil priority", nil, 0},
		{"high priority", strPtr("high"), 3},
		{"medium priority", strPtr("medium"), 2},
		{"low priority", strPtr("low"), 1},
		{"unknown priority", strPtr("unknown"), 0},
		{"empty string", strPtr(""), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityValue(tt.priority)
			if got != tt.want {
				t.Errorf("priorityValue() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSortTasksByPriority(t *testing.T) {
	// Create tasks with various priorities and IDs
	tasks := []*types.Task{
		{ID: 1, Title: "Low priority task", Priority: strPtr("low")},
		{ID: 2, Title: "High priority task", Priority: strPtr("high")},
		{ID: 3, Title: "No priority task", Priority: nil},
		{ID: 4, Title: "Medium priority task", Priority: strPtr("medium")},
		{ID: 5, Title: "Another high priority", Priority: strPtr("high")},
	}

	sortTasksByPriority(tasks)

	// Verify order: high (2, 5), medium (4), low (1), nil (3)
	expectedOrder := []int{2, 5, 4, 1, 3}
	for i, expectedID := range expectedOrder {
		if tasks[i].ID != expectedID {
			t.Errorf("position %d: got ID %d, want %d", i, tasks[i].ID, expectedID)
		}
	}
}

func TestSortTasksByPriority_SamePrioritySortsByID(t *testing.T) {
	// All tasks have the same priority - should sort by ID
	tasks := []*types.Task{
		{ID: 30, Title: "Task 30", Priority: strPtr("medium")},
		{ID: 10, Title: "Task 10", Priority: strPtr("medium")},
		{ID: 20, Title: "Task 20", Priority: strPtr("medium")},
	}

	sortTasksByPriority(tasks)

	expectedOrder := []int{10, 20, 30}
	for i, expectedID := range expectedOrder {
		if tasks[i].ID != expectedID {
			t.Errorf("position %d: got ID %d, want %d", i, tasks[i].ID, expectedID)
		}
	}
}

func TestSortTasksByPriority_AllNilPriority(t *testing.T) {
	// All tasks have nil priority - should sort by ID
	tasks := []*types.Task{
		{ID: 3, Title: "Task 3", Priority: nil},
		{ID: 1, Title: "Task 1", Priority: nil},
		{ID: 2, Title: "Task 2", Priority: nil},
	}

	sortTasksByPriority(tasks)

	expectedOrder := []int{1, 2, 3}
	for i, expectedID := range expectedOrder {
		if tasks[i].ID != expectedID {
			t.Errorf("position %d: got ID %d, want %d", i, tasks[i].ID, expectedID)
		}
	}
}

func TestSortTasksByPriority_EmptySlice(t *testing.T) {
	tasks := []*types.Task{}
	sortTasksByPriority(tasks) // Should not panic
	if len(tasks) != 0 {
		t.Errorf("expected empty slice, got %d tasks", len(tasks))
	}
}

func TestSortTasksByPriority_SingleTask(t *testing.T) {
	tasks := []*types.Task{
		{ID: 1, Title: "Only task", Priority: strPtr("high")},
	}
	sortTasksByPriority(tasks)
	if tasks[0].ID != 1 {
		t.Errorf("expected task ID 1, got %d", tasks[0].ID)
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
