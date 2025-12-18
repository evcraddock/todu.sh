package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Home directory expansion", "~/Documents", filepath.Join(home, "Documents")},
		{"Home directory nested", "~/foo/bar/baz", filepath.Join(home, "foo/bar/baz")},
		{"Absolute path unchanged", "/usr/local/bin", "/usr/local/bin"},
		{"Relative path unchanged", "relative/path", "relative/path"},
		{"Just tilde", "~", "~"}, // Only expands ~/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildProjectMap(t *testing.T) {
	projects := []*types.Project{
		{ID: 1, Name: "Project A"},
		{ID: 2, Name: "Project B"},
		{ID: 10, Name: "Project C"},
	}

	result := buildProjectMap(projects)

	if len(result) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(result))
	}

	expectedMappings := map[int]string{
		1:  "Project A",
		2:  "Project B",
		10: "Project C",
	}

	for id, name := range expectedMappings {
		if result[id] != name {
			t.Errorf("Expected projectMap[%d] = %q, got %q", id, name, result[id])
		}
	}
}

func TestBuildProjectMap_Empty(t *testing.T) {
	result := buildProjectMap(nil)
	if len(result) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(result))
	}
}

func TestBuildHabitTemplateSet(t *testing.T) {
	habits := []*types.RecurringTaskTemplate{
		{ID: 1, Title: "Exercise"},
		{ID: 5, Title: "Read"},
		{ID: 10, Title: "Meditate"},
	}

	result := buildHabitTemplateSet(habits)

	if len(result) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(result))
	}

	for _, h := range habits {
		if _, exists := result[h.ID]; !exists {
			t.Errorf("Expected habit ID %d to be in set", h.ID)
		}
	}
}

func TestBuildHabitTemplateSet_Empty(t *testing.T) {
	result := buildHabitTemplateSet(nil)
	if len(result) != 0 {
		t.Errorf("Expected empty set, got %d entries", len(result))
	}
}

func TestBuildHabitTaskMap(t *testing.T) {
	templateID1 := 1
	templateID2 := 2

	scheduledTasks := []*types.Task{
		{ID: 100, TemplateID: &templateID1, Status: "done"},
		{ID: 101, TemplateID: &templateID2, Status: "active"},
		{ID: 102, TemplateID: nil, Status: "done"}, // Not a habit task
	}

	habitTemplateIDs := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	result := buildHabitTaskMap(scheduledTasks, habitTemplateIDs)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}

	if result[1] == nil || !result[1].completed {
		t.Error("Expected habit 1 to be completed (true)")
	}
	if result[1] == nil || result[1].taskID != 100 {
		t.Error("Expected habit 1 to have task ID 100")
	}
	if result[2] == nil || result[2].completed {
		t.Error("Expected habit 2 to be not completed (false)")
	}
	if result[2] == nil || result[2].taskID != 101 {
		t.Error("Expected habit 2 to have task ID 101")
	}
	if _, exists := result[3]; exists {
		t.Error("Expected habit 3 to not be in map (no scheduled task)")
	}
}

func TestBuildDailyGoals(t *testing.T) {
	habits := []*types.RecurringTaskTemplate{
		{ID: 1, Title: "Exercise"},
		{ID: 2, Title: "Read"},
		{ID: 3, Title: "Meditate"},
	}

	habitTasks := map[int]*habitTaskInfo{
		1: {taskID: 100, completed: true},
		2: {taskID: 101, completed: false},
		// 3 not in map, should default to taskID=0 and completed=false
	}

	result := buildDailyGoals(habits, habitTasks)

	if len(result) != 3 {
		t.Errorf("Expected 3 goals, got %d", len(result))
	}

	expected := map[string]struct {
		taskID    int
		completed bool
	}{
		"Exercise": {taskID: 100, completed: true},
		"Read":     {taskID: 101, completed: false},
		"Meditate": {taskID: 0, completed: false},
	}

	for _, g := range result {
		exp := expected[g.name]
		if exp.completed != g.completed {
			t.Errorf("Goal %q: expected completed=%t, got %t", g.name, exp.completed, g.completed)
		}
		if exp.taskID != g.taskID {
			t.Errorf("Goal %q: expected taskID=%d, got %d", g.name, exp.taskID, g.taskID)
		}
	}
}

func TestIsSameDay(t *testing.T) {
	loc := time.Local

	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected bool
	}{
		{
			"Same day, same time",
			time.Date(2025, 12, 15, 10, 0, 0, 0, loc),
			time.Date(2025, 12, 15, 10, 0, 0, 0, loc),
			true,
		},
		{
			"Same day, different time",
			time.Date(2025, 12, 15, 8, 0, 0, 0, loc),
			time.Date(2025, 12, 15, 20, 30, 0, 0, loc),
			true,
		},
		{
			"Different day",
			time.Date(2025, 12, 15, 10, 0, 0, 0, loc),
			time.Date(2025, 12, 16, 10, 0, 0, 0, loc),
			false,
		},
		{
			"Different month",
			time.Date(2025, 11, 15, 10, 0, 0, 0, loc),
			time.Date(2025, 12, 15, 10, 0, 0, 0, loc),
			false,
		},
		{
			"Different year",
			time.Date(2024, 12, 15, 10, 0, 0, 0, loc),
			time.Date(2025, 12, 15, 10, 0, 0, 0, loc),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameDay(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("isSameDay() = %t, want %t", result, tt.expected)
			}
		})
	}
}

func TestBuildDailyExportPath(t *testing.T) {
	result := buildDailyExportPath("/home/user/reports")
	expected := "/home/user/reports/daily-review.md"
	if result != expected {
		t.Errorf("buildDailyExportPath() = %q, want %q", result, expected)
	}
}

func TestGenerateDailyMarkdown_Empty(t *testing.T) {
	data := &dailyData{
		targetDate:   time.Date(2025, 12, 16, 0, 0, 0, 0, time.Local),
		inProgress:   nil,
		dailyGoals:   nil,
		comingUpSoon: nil,
		next:         nil,
		waiting:      nil,
		doneToday:    nil,
		projectMap:   make(map[int]string),
	}

	result := generateDailyMarkdown(data)

	// Check that all sections are present
	expectedSections := []string{
		"# Daily Review",
		"## In Progress",
		"## Daily Goals",
		"## Coming up Soon",
		"## Next",
		"## Waiting",
		"## Done Today",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected markdown to contain %q", section)
		}
	}

	// Check that empty sections show "0 tasks"
	if strings.Count(result, "0 tasks") != 6 {
		t.Errorf("Expected 6 '0 tasks' entries for empty sections")
	}
}

func TestGenerateDailyMarkdown_WithData(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)

	data := &dailyData{
		targetDate: now,
		inProgress: []*types.Task{
			{ID: 1, Title: "Working on feature", ProjectID: 1},
		},
		dailyGoals: []*habitStatus{
			{taskID: 10, name: "Exercise", completed: true},
			{taskID: 11, name: "Read", completed: false},
		},
		comingUpSoon: []*types.Task{
			{ID: 2, Title: "Due soon task", ProjectID: 1, DueDate: &dueDate},
		},
		next: []*types.Task{
			{ID: 3, Title: "High priority task", ProjectID: 2},
		},
		waiting: []*types.Task{
			{ID: 4, Title: "Waiting for response", ProjectID: 1},
		},
		doneToday: []*types.Task{
			{ID: 5, Title: "Completed task", ProjectID: 1},
		},
		projectMap: map[int]string{
			1: "Project A",
			2: "Project B",
		},
	}

	result := generateDailyMarkdown(data)

	// Check tasks are listed
	if !strings.Contains(result, "#1 Working on feature") {
		t.Error("Expected in progress task to be listed")
	}
	if !strings.Contains(result, "#10 Exercise : true") {
		t.Error("Expected completed habit to show task ID and true")
	}
	if !strings.Contains(result, "#11 Read : false") {
		t.Error("Expected incomplete habit to show task ID and false")
	}
	if !strings.Contains(result, "#2 Due soon task") {
		t.Error("Expected coming up soon task to be listed")
	}
	if !strings.Contains(result, "#3 High priority task") {
		t.Error("Expected next task to be listed")
	}
	if !strings.Contains(result, "#4 Waiting for response") {
		t.Error("Expected waiting task to be listed")
	}
	if !strings.Contains(result, "#5 Completed task") {
		t.Error("Expected done task to be listed")
	}

	// Check project names are included
	if !strings.Contains(result, "(Project A)") {
		t.Error("Expected project name to be included")
	}

	// Check task counts
	if !strings.Contains(result, "1 task\n") {
		t.Error("Expected '1 task' count for single-task sections")
	}
	if !strings.Contains(result, "2 tasks\n") {
		t.Error("Expected '2 tasks' count for daily goals section")
	}
}

func TestGenerateDailyMarkdown_HabitWithoutTaskID(t *testing.T) {
	data := &dailyData{
		targetDate: time.Now(),
		inProgress: nil,
		dailyGoals: []*habitStatus{
			{taskID: 0, name: "Unscheduled habit", completed: false},
		},
		comingUpSoon: nil,
		next:         nil,
		waiting:      nil,
		doneToday:    nil,
		projectMap:   make(map[int]string),
	}

	result := generateDailyMarkdown(data)

	// Habits without task ID should not show the # prefix
	if !strings.Contains(result, "- Unscheduled habit : false") {
		t.Error("Expected habit without task ID to be listed without # prefix")
	}
	if strings.Contains(result, "#0 Unscheduled habit") {
		t.Error("Expected habit with taskID=0 to not show #0 prefix")
	}
}

func TestFilterDoneToday(t *testing.T) {
	targetDate := time.Date(2025, 12, 16, 0, 0, 0, 0, time.Local)
	todayTime := time.Date(2025, 12, 16, 14, 30, 0, 0, time.Local)
	yesterdayTime := time.Date(2025, 12, 15, 14, 30, 0, 0, time.Local)

	templateID := 1
	tasks := []*types.Task{
		{ID: 1, Title: "Done today", UpdatedAt: todayTime, TemplateID: nil},
		{ID: 2, Title: "Done yesterday", UpdatedAt: yesterdayTime, TemplateID: nil},
		{ID: 3, Title: "Habit done today", UpdatedAt: todayTime, TemplateID: &templateID},
	}

	habitTemplateIDs := map[int]struct{}{
		1: {},
	}

	result := filterDoneToday(tasks, targetDate, habitTemplateIDs)

	if len(result) != 1 {
		t.Errorf("Expected 1 task, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("Expected task ID 1, got %d", result[0].ID)
	}
}

func TestBuildNextSection_Deduplication(t *testing.T) {
	task1 := &types.Task{ID: 1, Title: "Task 1", Status: "active"}
	task2 := &types.Task{ID: 2, Title: "Task 2", Status: "active"}
	task3 := &types.Task{ID: 3, Title: "Task 3", Status: "active"}

	// task1 appears in both high priority and default project
	highPriority := []*types.Task{task1, task2}
	scheduledTasks := []*types.Task{}
	defaultProject := []*types.Task{task1, task3}

	habitTemplateIDs := make(map[int]struct{})

	result := buildNextSection(highPriority, scheduledTasks, defaultProject, habitTemplateIDs)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique tasks, got %d", len(result))
	}

	// Check all tasks are present
	ids := make(map[int]bool)
	for _, task := range result {
		ids[task.ID] = true
	}
	for i := 1; i <= 3; i++ {
		if !ids[i] {
			t.Errorf("Expected task ID %d to be in result", i)
		}
	}
}

func TestBuildNextSection_ExcludesHabits(t *testing.T) {
	templateID := 1
	task1 := &types.Task{ID: 1, Title: "Regular task", Status: "active"}
	task2 := &types.Task{ID: 2, Title: "Habit task", Status: "active", TemplateID: &templateID}

	highPriority := []*types.Task{task1, task2}

	habitTemplateIDs := map[int]struct{}{
		1: {},
	}

	result := buildNextSection(highPriority, nil, nil, habitTemplateIDs)

	if len(result) != 1 {
		t.Errorf("Expected 1 task (excluding habit), got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("Expected task ID 1, got %d", result[0].ID)
	}
}
