package review

import (
	"strings"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestBuildNextWeeklySection(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)

	highPriority := []*types.Task{
		{ID: 1, Title: "High priority task", Status: "active"},
	}

	activeTasks := []*types.Task{
		{ID: 1, Title: "High priority task", Status: "active"}, // Duplicate
		{ID: 2, Title: "Overdue task", Status: "active", DueDate: &yesterday},
		{ID: 3, Title: "Future task", Status: "active", DueDate: &tomorrow},
	}

	result := buildNextWeeklySection(highPriority, activeTasks, today)

	// Should have high priority + overdue, but not future task
	if len(result) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(result))
	}

	// Check deduplication - task 1 should only appear once
	ids := make(map[int]int)
	for _, task := range result {
		ids[task.ID]++
	}
	if ids[1] != 1 {
		t.Errorf("Task 1 should appear exactly once, got %d", ids[1])
	}

	// Task 3 (future) should not be included
	if ids[3] != 0 {
		t.Error("Future task should not be included in Next section")
	}
}

func TestBuildActiveSection(t *testing.T) {
	priority := "medium"
	mediumPriority := []*types.Task{
		{ID: 1, Title: "Medium priority task", Status: "active", Priority: &priority},
	}

	highPriority := "high"
	lowPriority := "low"
	activeTasks := []*types.Task{
		{ID: 1, Title: "Medium priority task", Status: "active", Priority: &priority}, // Duplicate
		{ID: 2, Title: "No priority task", Status: "active", Priority: nil},
		{ID: 3, Title: "High priority task", Status: "active", Priority: &highPriority},
		{ID: 4, Title: "Low priority task", Status: "active", Priority: &lowPriority},
	}

	// No exclusions
	excludeIDs := make(map[int]struct{})
	result := buildActiveSection(mediumPriority, activeTasks, excludeIDs)

	// Should have medium priority + no priority only
	if len(result) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(result))
	}

	// Check correct tasks are included
	ids := make(map[int]bool)
	for _, task := range result {
		ids[task.ID] = true
	}

	if !ids[1] {
		t.Error("Medium priority task should be included")
	}
	if !ids[2] {
		t.Error("No priority task should be included")
	}
	if ids[3] {
		t.Error("High priority task should not be included")
	}
	if ids[4] {
		t.Error("Low priority task should not be included")
	}
}

func TestBuildActiveSection_WithExclusions(t *testing.T) {
	priority := "medium"
	mediumPriority := []*types.Task{
		{ID: 1, Title: "Medium priority task", Status: "active", Priority: &priority},
		{ID: 5, Title: "Another medium", Status: "active", Priority: &priority},
	}

	activeTasks := []*types.Task{
		{ID: 2, Title: "No priority task", Status: "active", Priority: nil},
		{ID: 6, Title: "Another no priority", Status: "active", Priority: nil},
	}

	// Exclude task 1 and 2 (as if they were in Next section)
	excludeIDs := map[int]struct{}{
		1: {},
		2: {},
	}
	result := buildActiveSection(mediumPriority, activeTasks, excludeIDs)

	// Should only have tasks 5 and 6
	if len(result) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(result))
	}

	ids := make(map[int]bool)
	for _, task := range result {
		ids[task.ID] = true
	}

	if ids[1] {
		t.Error("Task 1 should be excluded")
	}
	if ids[2] {
		t.Error("Task 2 should be excluded")
	}
	if !ids[5] {
		t.Error("Task 5 should be included")
	}
	if !ids[6] {
		t.Error("Task 6 should be included")
	}
}

func TestSortTasksByDueDate(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tasks := []*types.Task{
		{ID: 1, Title: "No due date", DueDate: nil},
		{ID: 2, Title: "Tomorrow", DueDate: &tomorrow},
		{ID: 3, Title: "Yesterday", DueDate: &yesterday},
	}

	sortTasksByDueDate(tasks)

	// Should be: yesterday, tomorrow, no due date
	if tasks[0].ID != 3 {
		t.Errorf("First task should be ID 3 (yesterday), got %d", tasks[0].ID)
	}
	if tasks[1].ID != 2 {
		t.Errorf("Second task should be ID 2 (tomorrow), got %d", tasks[1].ID)
	}
	if tasks[2].ID != 1 {
		t.Errorf("Third task should be ID 1 (no due date), got %d", tasks[2].ID)
	}
}

func TestGenerateWeeklyMarkdown_Empty(t *testing.T) {
	data := &weeklyData{
		waiting:    nil,
		next:       nil,
		active:     nil,
		someday:    nil,
		projectMap: make(map[int]string),
	}

	result := generateWeeklyMarkdown(data)

	// Check that all sections are present
	expectedSections := []string{
		"# Weekly Review",
		"## Waiting",
		"## Next",
		"## Active",
		"## Someday",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected markdown to contain %q", section)
		}
	}

	// Check that empty sections show "0 tasks"
	if strings.Count(result, "0 tasks") != 4 {
		t.Errorf("Expected 4 '0 tasks' entries for empty sections, got %d", strings.Count(result, "0 tasks"))
	}
}

func TestGenerateWeeklyMarkdown_WithData(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)

	data := &weeklyData{
		waiting: []*types.Task{
			{ID: 1, Title: "Waiting for response", ProjectID: 1},
		},
		next: []*types.Task{
			{ID: 2, Title: "High priority task", ProjectID: 1, DueDate: &dueDate},
		},
		active: []*types.Task{
			{ID: 3, Title: "Medium priority task", ProjectID: 2},
		},
		someday: []*types.Task{
			{ID: 4, Title: "Low priority task", ProjectID: 1},
		},
		projectMap: map[int]string{
			1: "Project A",
			2: "Project B",
		},
	}

	result := generateWeeklyMarkdown(data)

	// Check tasks are listed
	if !strings.Contains(result, "#1 Waiting for response") {
		t.Error("Expected waiting task to be listed")
	}
	if !strings.Contains(result, "#2 High priority task") {
		t.Error("Expected next task to be listed")
	}
	if !strings.Contains(result, "#3 Medium priority task") {
		t.Error("Expected active task to be listed")
	}
	if !strings.Contains(result, "#4 Low priority task") {
		t.Error("Expected someday task to be listed")
	}

	// Check project names are included
	if !strings.Contains(result, "(Project A)") {
		t.Error("Expected project name to be included")
	}
	if !strings.Contains(result, "(Project B)") {
		t.Error("Expected project name to be included")
	}

	// Check task counts - each section has 1 task
	if strings.Count(result, "1 task\n") != 4 {
		t.Errorf("Expected 4 '1 task' entries, got %d", strings.Count(result, "1 task\n"))
	}
}

func TestDefaultWeeklyReportPath(t *testing.T) {
	result := DefaultWeeklyReportPath("/home/user/reports")
	expected := "/home/user/reports/weekly-review.md"
	if result != expected {
		t.Errorf("DefaultWeeklyReportPath() = %q, want %q", result, expected)
	}
}
