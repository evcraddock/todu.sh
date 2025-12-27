package review

import (
	"strings"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestGetWeekBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		endDate       time.Time
		expectedStart string
		expectedEnd   string
	}{
		{
			name:          "regular week",
			endDate:       time.Date(2025, 12, 27, 10, 30, 0, 0, time.Local),
			expectedStart: "2025-12-21",
			expectedEnd:   "2025-12-27",
		},
		{
			name:          "week spanning month boundary",
			endDate:       time.Date(2026, 1, 3, 0, 0, 0, 0, time.Local),
			expectedStart: "2025-12-28",
			expectedEnd:   "2026-01-03",
		},
		{
			name:          "week spanning year boundary",
			endDate:       time.Date(2026, 1, 5, 15, 45, 30, 0, time.Local),
			expectedStart: "2025-12-30",
			expectedEnd:   "2026-01-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getWeekBoundaries(tt.endDate)

			if start.Format("2006-01-02") != tt.expectedStart {
				t.Errorf("start = %s, want %s", start.Format("2006-01-02"), tt.expectedStart)
			}
			if end.Format("2006-01-02") != tt.expectedEnd {
				t.Errorf("end = %s, want %s", end.Format("2006-01-02"), tt.expectedEnd)
			}

			// Verify end is truncated to beginning of day
			if end.Hour() != 0 || end.Minute() != 0 || end.Second() != 0 {
				t.Error("end should be truncated to beginning of day")
			}
		})
	}
}

func TestTruncateToDay(t *testing.T) {
	input := time.Date(2025, 12, 25, 14, 30, 45, 123456789, time.Local)
	result := truncateToDay(input)

	if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
		t.Errorf("truncateToDay did not zero out time components: got %v", result)
	}
	if result.Year() != 2025 || result.Month() != 12 || result.Day() != 25 {
		t.Errorf("truncateToDay changed date components: got %v", result)
	}
}

func TestBuildWeeklyHabitTaskMap(t *testing.T) {
	templateID1 := 1
	templateID2 := 2
	templateID3 := 3 // Not a habit

	scheduledDate1 := time.Date(2025, 12, 21, 0, 0, 0, 0, time.Local)
	scheduledDate2 := time.Date(2025, 12, 22, 0, 0, 0, 0, time.Local)

	scheduledTasks := []*types.Task{
		{ID: 101, TemplateID: &templateID1, ScheduledDate: &scheduledDate1, Status: "done"},
		{ID: 102, TemplateID: &templateID1, ScheduledDate: &scheduledDate2, Status: "active"},
		{ID: 103, TemplateID: &templateID2, ScheduledDate: &scheduledDate1, Status: "done"},
		{ID: 104, TemplateID: &templateID3, ScheduledDate: &scheduledDate1, Status: "done"}, // Not a habit
		{ID: 105, TemplateID: nil, ScheduledDate: &scheduledDate1, Status: "done"},          // No template
	}

	habitTemplateIDs := map[int]struct{}{
		1: {},
		2: {},
	}

	result := buildWeeklyHabitTaskMap(scheduledTasks, habitTemplateIDs)

	// Check template 1 has two days
	if len(result[1]) != 2 {
		t.Errorf("Template 1 should have 2 days, got %d", len(result[1]))
	}

	// Check template 2 has one day
	if len(result[2]) != 1 {
		t.Errorf("Template 2 should have 1 day, got %d", len(result[2]))
	}

	// Check template 3 is not included (not a habit)
	if _, exists := result[3]; exists {
		t.Error("Template 3 should not be included (not a habit)")
	}

	// Check completion status
	if !result[1]["2025-12-21"].completed {
		t.Error("Task 101 should be marked as completed")
	}
	if result[1]["2025-12-22"].completed {
		t.Error("Task 102 should not be marked as completed")
	}
}

func TestFilterNonHabitTasks(t *testing.T) {
	templateID1 := 1 // Habit
	templateID2 := 2 // Not a habit

	tasks := []*types.Task{
		{ID: 1, Title: "Regular task", TemplateID: nil},
		{ID: 2, Title: "Habit task", TemplateID: &templateID1},
		{ID: 3, Title: "Recurring non-habit", TemplateID: &templateID2},
	}

	habitTemplateIDs := map[int]struct{}{
		1: {},
	}

	result := filterNonHabitTasks(tasks, habitTemplateIDs)

	if len(result) != 2 {
		t.Errorf("Expected 2 non-habit tasks, got %d", len(result))
	}

	ids := make(map[int]bool)
	for _, task := range result {
		ids[task.ID] = true
	}

	if ids[2] {
		t.Error("Habit task should be filtered out")
	}
	if !ids[1] || !ids[3] {
		t.Error("Non-habit tasks should be included")
	}
}

func TestGenerateWeeklyReviewMarkdown_Empty(t *testing.T) {
	data := &weeklyReviewData{
		startDate:      time.Date(2025, 12, 21, 0, 0, 0, 0, time.Local),
		endDate:        time.Date(2025, 12, 27, 0, 0, 0, 0, time.Local),
		completedTasks: nil,
		habits:         nil,
		projectMap:     make(map[int]string),
		habitTasks:     make(map[int]map[string]*weeklyHabitTaskInfo),
	}

	result := generateWeeklyReviewMarkdown(data)

	// Check header with date range
	if !strings.Contains(result, "# Weekly Review: 12-21-2025 to 12-27-2025") {
		t.Error("Expected header with date range")
	}

	// Check all sections are present
	expectedSections := []string{
		"## Projects Worked On",
		"## Habits Summary",
		"## Weekly Stats",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected markdown to contain %q", section)
		}
	}

	// Check empty messages
	if !strings.Contains(result, "No tasks completed this week.") {
		t.Error("Expected 'No tasks completed' message")
	}
	if !strings.Contains(result, "No habits tracked.") {
		t.Error("Expected 'No habits tracked' message")
	}
}

func TestGenerateWeeklyReviewMarkdown_WithData(t *testing.T) {
	data := &weeklyReviewData{
		startDate: time.Date(2025, 12, 21, 0, 0, 0, 0, time.Local),
		endDate:   time.Date(2025, 12, 27, 0, 0, 0, 0, time.Local),
		completedTasks: []*types.Task{
			{ID: 1, Title: "Fix login bug", ProjectID: 1},
			{ID: 2, Title: "Add tests", ProjectID: 1},
			{ID: 3, Title: "Update docs", ProjectID: 2},
		},
		habits: []*types.RecurringTaskTemplate{
			{ID: 10, Title: "Exercise"},
			{ID: 11, Title: "Reading"},
		},
		projectMap: map[int]string{
			1: "todu.sh",
			2: "Documentation",
		},
		habitTasks: map[int]map[string]*weeklyHabitTaskInfo{
			10: {
				"2025-12-21": {taskID: 100, completed: true},
				"2025-12-22": {taskID: 101, completed: false},
				"2025-12-23": {taskID: 102, completed: true},
			},
			11: {
				"2025-12-21": {taskID: 200, completed: true},
			},
		},
	}

	result := generateWeeklyReviewMarkdown(data)

	// Check project sections
	if !strings.Contains(result, "### todu.sh") {
		t.Error("Expected todu.sh project section")
	}
	if !strings.Contains(result, "### Documentation") {
		t.Error("Expected Documentation project section")
	}

	// Check task count
	if !strings.Contains(result, "Completed 2 task(s)") {
		t.Error("Expected '2 task(s)' for todu.sh project")
	}

	// Check tasks are listed with checkboxes
	if !strings.Contains(result, "- [x] #1 Fix login bug") {
		t.Error("Expected task with checkbox")
	}

	// Check habits summary table header
	if !strings.Contains(result, "| Habit |") {
		t.Error("Expected habits table header")
	}
	if !strings.Contains(result, "| Exercise |") {
		t.Error("Expected Exercise habit row")
	}

	// Check stats
	if !strings.Contains(result, "**Tasks Completed**: 3") {
		t.Error("Expected tasks completed stat")
	}
	if !strings.Contains(result, "**Habits Completed**: 3/4") {
		t.Error("Expected habits completed stat (3 done out of 4 scheduled)")
	}
}

func TestWriteHabitsSummary_DayHeaders(t *testing.T) {
	data := &weeklyReviewData{
		startDate: time.Date(2025, 12, 21, 0, 0, 0, 0, time.Local), // Sunday
		endDate:   time.Date(2025, 12, 27, 0, 0, 0, 0, time.Local), // Saturday
		habits: []*types.RecurringTaskTemplate{
			{ID: 1, Title: "Test Habit"},
		},
		habitTasks: map[int]map[string]*weeklyHabitTaskInfo{
			1: {
				"2025-12-21": {taskID: 100, completed: true},
				"2025-12-24": {taskID: 101, completed: false},
			},
		},
	}

	var sb strings.Builder
	writeHabitsSummary(&sb, data)
	result := sb.String()

	// Check day headers are present (Sun through Sat)
	expectedDays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for _, day := range expectedDays {
		if !strings.Contains(result, day) {
			t.Errorf("Expected day header %s", day)
		}
	}

	// Check symbols
	if !strings.Contains(result, "✓") {
		t.Error("Expected checkmark for completed habit")
	}
	if !strings.Contains(result, "○") {
		t.Error("Expected circle for incomplete habit")
	}
	if !strings.Contains(result, "-") {
		t.Error("Expected dash for non-scheduled days")
	}
}

func TestDefaultWeeklyReportPath(t *testing.T) {
	result := DefaultWeeklyReportPath("/home/user/reports")
	expected := "/home/user/reports/weekly-review.md"
	if result != expected {
		t.Errorf("DefaultWeeklyReportPath() = %q, want %q", result, expected)
	}
}

func TestBuildWeeklyReportPath(t *testing.T) {
	endDate := time.Date(2025, 12, 27, 0, 0, 0, 0, time.Local)
	result := BuildWeeklyReportPath("/home/user/reports", endDate)
	expected := "/home/user/reports/reviews/2025/12-December/12-27-2025-weekly-review.md"
	if result != expected {
		t.Errorf("BuildWeeklyReportPath() = %q, want %q", result, expected)
	}
}

func TestBuildDatedWeeklyExportPath(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			name:     "regular path",
			start:    time.Date(2025, 12, 21, 0, 0, 0, 0, time.Local),
			end:      time.Date(2025, 12, 27, 0, 0, 0, 0, time.Local),
			expected: "/reports/reviews/2025/12-December/12-27-2025-weekly-review.md",
		},
		{
			name:     "week spanning months",
			start:    time.Date(2025, 12, 28, 0, 0, 0, 0, time.Local),
			end:      time.Date(2026, 1, 3, 0, 0, 0, 0, time.Local),
			expected: "/reports/reviews/2026/01-January/01-03-2026-weekly-review.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDatedWeeklyExportPath("/reports", tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("buildDatedWeeklyExportPath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestWriteProjectSummaries_SortedByProjectID(t *testing.T) {
	tasks := []*types.Task{
		{ID: 3, Title: "Task C", ProjectID: 3},
		{ID: 1, Title: "Task A", ProjectID: 1},
		{ID: 2, Title: "Task B", ProjectID: 2},
	}

	projectMap := map[int]string{
		1: "Alpha",
		2: "Beta",
		3: "Gamma",
	}

	var sb strings.Builder
	writeProjectSummaries(&sb, tasks, projectMap)
	result := sb.String()

	// Check that projects appear in order (by ID)
	alphaIdx := strings.Index(result, "### Alpha")
	betaIdx := strings.Index(result, "### Beta")
	gammaIdx := strings.Index(result, "### Gamma")

	if alphaIdx == -1 || betaIdx == -1 || gammaIdx == -1 {
		t.Fatal("Missing project sections")
	}

	if !(alphaIdx < betaIdx && betaIdx < gammaIdx) {
		t.Error("Projects should be sorted by ID (Alpha < Beta < Gamma)")
	}
}
