package journal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestFormatTimezone(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
		expected      string
	}{
		{"Eastern Time", -5 * 3600, "ET"},
		{"Central Time", -6 * 3600, "CT"},
		{"Mountain Time", -7 * 3600, "MT"},
		{"Pacific Time", -8 * 3600, "PT"},
		{"Atlantic Time", -4 * 3600, "AT"},
		{"Positive offset +5", 5 * 3600, "+0500"},
		{"Positive offset +9", 9 * 3600, "+0900"},
		{"Negative offset -9", -9 * 3600, "-0900"},
		{"Negative offset -3", -3 * 3600, "-0300"},
		{"UTC", 0, "+0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimezone(tt.offsetSeconds)
			if result != tt.expected {
				t.Errorf("formatTimezone(%d) = %s, want %s", tt.offsetSeconds, result, tt.expected)
			}
		})
	}
}

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
		{"Tilde in middle", "/home/~/test", "/home/~/test"},
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
		{ID: 1, Title: "Habit 1"},
		{ID: 5, Title: "Habit 2"},
		{ID: 10, Title: "Habit 3"},
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

	// Check that non-existent ID is not in set
	if _, exists := result[999]; exists {
		t.Errorf("Did not expect ID 999 to be in set")
	}
}

func TestBuildHabitTaskMap(t *testing.T) {
	habitTemplateIDs := map[int]struct{}{
		1: {},
		2: {},
	}

	templateID1 := 1
	templateID2 := 2
	templateID3 := 3 // Not a habit

	scheduledTasks := []*types.Task{
		{ID: 100, TemplateID: &templateID1, Status: "done"},
		{ID: 101, TemplateID: &templateID2, Status: "active"},
		{ID: 102, TemplateID: &templateID3, Status: "done"}, // Not a habit
		{ID: 103, TemplateID: nil, Status: "active"},        // No template
	}

	result := buildHabitTaskMap(scheduledTasks, habitTemplateIDs)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}

	// Check habit 1 (completed)
	if info, exists := result[1]; !exists {
		t.Errorf("Expected habit template ID 1 to be in map")
	} else {
		if info.taskID != 100 {
			t.Errorf("Expected taskID 100, got %d", info.taskID)
		}
		if !info.completed {
			t.Errorf("Expected completed to be true")
		}
	}

	// Check habit 2 (not completed)
	if info, exists := result[2]; !exists {
		t.Errorf("Expected habit template ID 2 to be in map")
	} else {
		if info.taskID != 101 {
			t.Errorf("Expected taskID 101, got %d", info.taskID)
		}
		if info.completed {
			t.Errorf("Expected completed to be false")
		}
	}

	// Check that non-habit task is not in map
	if _, exists := result[3]; exists {
		t.Errorf("Did not expect template ID 3 (non-habit) to be in map")
	}
}

func TestFilterJournalsByTargetDate(t *testing.T) {
	targetDate := time.Date(2025, 12, 13, 0, 0, 0, 0, time.Local)

	journals := []*types.Comment{
		{ID: 1, CreatedAt: time.Date(2025, 12, 13, 10, 30, 0, 0, time.Local)}, // Same day
		{ID: 2, CreatedAt: time.Date(2025, 12, 13, 23, 59, 0, 0, time.Local)}, // Same day, late
		{ID: 3, CreatedAt: time.Date(2025, 12, 12, 10, 30, 0, 0, time.Local)}, // Previous day
		{ID: 4, CreatedAt: time.Date(2025, 12, 14, 10, 30, 0, 0, time.Local)}, // Next day
	}

	result := filterJournalsByTargetDate(journals, targetDate)

	if len(result) != 2 {
		t.Errorf("Expected 2 journals, got %d", len(result))
	}

	// Verify the correct ones were kept
	for _, j := range result {
		if j.ID != 1 && j.ID != 2 {
			t.Errorf("Unexpected journal ID %d in result", j.ID)
		}
	}
}

func TestFilterTasksByTargetDate(t *testing.T) {
	targetDate := time.Date(2025, 12, 13, 0, 0, 0, 0, time.Local)

	tasks := []*types.Task{
		{ID: 1, UpdatedAt: time.Date(2025, 12, 13, 10, 30, 0, 0, time.Local)}, // Same day
		{ID: 2, UpdatedAt: time.Date(2025, 12, 12, 10, 30, 0, 0, time.Local)}, // Previous day
		{ID: 3, UpdatedAt: time.Date(2025, 12, 13, 0, 1, 0, 0, time.Local)},   // Same day, early
	}

	result := filterTasksByTargetDate(tasks, targetDate)

	if len(result) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(result))
	}

	for _, task := range result {
		if task.ID != 1 && task.ID != 3 {
			t.Errorf("Unexpected task ID %d in result", task.ID)
		}
	}
}

func TestBuildExportPath(t *testing.T) {
	tests := []struct {
		name         string
		localReports string
		targetDate   time.Time
		expected     string
	}{
		{
			name:         "Standard path",
			localReports: "/home/user/vault",
			targetDate:   time.Date(2025, 12, 13, 0, 0, 0, 0, time.Local),
			expected:     "/home/user/vault/reviews/2025/12-December/12-13-2025-journal.md",
		},
		{
			name:         "January date",
			localReports: "/reports",
			targetDate:   time.Date(2025, 1, 5, 0, 0, 0, 0, time.Local),
			expected:     "/reports/reviews/2025/01-January/01-05-2025-journal.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildExportPath(tt.localReports, tt.targetDate)
			if result != tt.expected {
				t.Errorf("buildExportPath(%q, %v) = %q, want %q",
					tt.localReports, tt.targetDate, result, tt.expected)
			}
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No special chars", "normal text", "normal text"},
		{"Asterisks", "text with *asterisks*", `text with \*asterisks\*`},
		{"Underscores", "text_with_underscores", `text\_with\_underscores`},
		{"Brackets", "[link](url)", `\[link\](url)`},
		{"Hash", "#heading", `\#heading`},
		{"Backslash", `path\to\file`, `path\\to\\file`},
		{"Backticks", "`code`", "\\`code\\`"},
		{"Pipes", "col1|col2", `col1\|col2`},
		{"Multiple specials", "*bold* and _italic_", `\*bold\* and \_italic\_`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateMarkdown_Empty(t *testing.T) {
	data := &exportData{
		targetDate:     time.Date(2025, 12, 13, 0, 0, 0, 0, time.Local),
		journals:       nil,
		completedTasks: nil,
		habits:         nil,
		projectMap:     make(map[int]string),
		habitTasks:     make(map[int]habitTaskInfo),
	}

	result := generateMarkdown(data)

	// Check header
	if !contains(result, "# 12-13-2025 Journal") {
		t.Errorf("Expected header '# 12-13-2025 Journal' in output")
	}

	// Check empty completed section
	if !contains(result, "No Tasks") {
		t.Errorf("Expected 'No Tasks' in output")
	}

	// Check empty habits section
	if !contains(result, "No Habits") {
		t.Errorf("Expected 'No Habits' in output")
	}
}

func TestGenerateMarkdown_WithData(t *testing.T) {
	priority := "high"
	templateID := 1

	data := &exportData{
		targetDate: time.Date(2025, 12, 13, 12, 0, 0, 0, time.Local),
		journals: []*types.Comment{
			{ID: 1, Content: "Test journal entry", CreatedAt: time.Date(2025, 12, 13, 10, 30, 0, 0, time.Local)},
		},
		completedTasks: []*types.Task{
			{ID: 100, Title: "Complete task", ProjectID: 1, Priority: &priority},
		},
		habits: []*types.RecurringTaskTemplate{
			{ID: 1, Title: "Morning routine", ProjectID: 2},
			{ID: 2, Title: "Evening review", ProjectID: 2},
		},
		projectMap: map[int]string{
			1: "Work",
			2: "Personal",
		},
		habitTasks: map[int]habitTaskInfo{
			1: {taskID: 200, completed: true},
			// Habit 2 has no task scheduled
		},
	}

	// Set templateID for the completed task (not a habit)
	data.completedTasks[0].TemplateID = nil

	result := generateMarkdown(data)

	// Check journal entry
	if !contains(result, "Test journal entry") {
		t.Errorf("Expected journal content in output")
	}

	// Check completed task with checkbox format
	if !contains(result, "- [x] #100 Complete task - Work (priority: high)") {
		t.Errorf("Expected completed task with checkbox format in output, got: %s", result)
	}

	// Check habit with task ID (completed)
	if !contains(result, "- #200 Personal - Morning routine:: true") {
		t.Errorf("Expected completed habit with task ID in output")
	}

	// Check habit without task
	if !contains(result, "- Personal - Evening review:: false") {
		t.Errorf("Expected habit without task in output")
	}

	// Verify habit tasks are excluded from completed section
	data.completedTasks = append(data.completedTasks, &types.Task{
		ID:         201,
		Title:      "Habit task",
		ProjectID:  2,
		TemplateID: &templateID,
		Priority:   &priority,
	})

	result2 := generateMarkdown(data)
	// The habit task should NOT appear in completed section
	if contains(result2, "#201") {
		t.Errorf("Habit task should not appear in completed section")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
