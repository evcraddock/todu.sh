package review

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/pkg/types"
	"golang.org/x/sync/errgroup"
)

// weeklyReviewData holds all data needed for the weekly review
type weeklyReviewData struct {
	startDate      time.Time
	endDate        time.Time
	completedTasks []*types.Task
	habits         []*types.RecurringTaskTemplate
	projectMap     map[int]string
	habitTasks     map[int]map[string]*weeklyHabitTaskInfo // templateID -> date -> taskInfo
}

// weeklyHabitTaskInfo holds the task ID and completion status for a habit on a specific day
type weeklyHabitTaskInfo struct {
	taskID    int
	completed bool
}

// weeklyAPIResults holds raw results from all API calls for weekly review
type weeklyAPIResults struct {
	completedTasks []*types.Task
	scheduledTasks []*types.Task
	habits         []*types.RecurringTaskTemplate
	projects       []*types.Project
}

// WeeklyReport generates a weekly review report and returns the markdown content
// startDate is the first day of the 7-day period
func WeeklyReport(ctx context.Context, client *api.Client, startDate time.Time) (string, error) {
	start, end := getWeekBoundaries(startDate)

	// Fetch all data in parallel
	results, err := fetchWeeklyReviewData(ctx, client, start, end)
	if err != nil {
		return "", err
	}

	// Build project map
	projectMap := buildProjectMap(results.projects)

	// Build habit template set
	habitTemplateIDs := buildHabitTemplateSet(results.habits)

	// Build habit task map: templateID -> date -> taskInfo
	habitTasks := buildWeeklyHabitTaskMap(results.scheduledTasks, habitTemplateIDs)

	// Filter completed tasks to exclude habit tasks
	completedTasks := filterNonHabitTasks(results.completedTasks, habitTemplateIDs)

	data := &weeklyReviewData{
		startDate:      start,
		endDate:        end,
		completedTasks: completedTasks,
		habits:         results.habits,
		projectMap:     projectMap,
		habitTasks:     habitTasks,
	}

	return generateWeeklyReviewMarkdown(data), nil
}

// DefaultWeeklyReportPath returns the default path for the weekly review file
func DefaultWeeklyReportPath(localReportsPath string) string {
	return buildWeeklyExportPath(expandPath(localReportsPath))
}

// BuildWeeklyReportPath returns a dated path for saving weekly reviews
func BuildWeeklyReportPath(localReportsPath string, startDate time.Time) string {
	start, end := getWeekBoundaries(startDate)
	return buildDatedWeeklyExportPath(expandPath(localReportsPath), start, end)
}

// buildWeeklyExportPath constructs the output file path for the weekly review (simple path)
func buildWeeklyExportPath(localReports string) string {
	return localReports + "/weekly-review.md"
}

// buildDatedWeeklyExportPath constructs the dated output file path
// Format: {local_reports}/reviews/YYYY/MM-MonthName/MM-DD-YYYY-weekly-review.md
// Uses the end date of the week for the filename
func buildDatedWeeklyExportPath(localReports string, start, end time.Time) string {
	year := end.Format("2006")
	monthDir := end.Format("01-January")
	fileName := fmt.Sprintf("%s-weekly-review.md", end.Format("01-02-2006"))
	return filepath.Join(localReports, "reviews", year, monthDir, fileName)
}

// getWeekBoundaries calculates the start and end dates for a 7-day period
// Week ends on the specified date and covers 7 days (end - 6 days)
func getWeekBoundaries(endDate time.Time) (start, end time.Time) {
	end = truncateToDay(endDate)
	start = end.AddDate(0, 0, -6) // 7 days total (end - 6)
	return start, end
}

// truncateToDay truncates a time to the start of the day in local timezone
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// fetchWeeklyReviewData fetches all data needed for the weekly review in parallel
func fetchWeeklyReviewData(ctx context.Context, client *api.Client, start, end time.Time) (*weeklyAPIResults, error) {
	ctx, cancel := context.WithTimeout(ctx, exportAPITimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	results := &weeklyAPIResults{}

	startStr := start.Format("2006-01-02")
	// Add 1 day to end for inclusive range (API uses exclusive end)
	endPlusOne := end.AddDate(0, 0, 1).Format("2006-01-02")

	// 1. Completed tasks in date range
	g.Go(func() error {
		var err error
		results.completedTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:        "done",
			UpdatedAfter:  startStr,
			UpdatedBefore: endPlusOne,
			Limit:         maxTaskLimit,
		})
		return err
	})

	// 2. Scheduled tasks in date range (for habit tracking)
	g.Go(func() error {
		var err error
		results.scheduledTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			ScheduledAfter:  startStr,
			ScheduledBefore: endPlusOne,
			Limit:           maxTaskLimit,
		})
		return err
	})

	// 3. Habit templates
	g.Go(func() error {
		var err error
		results.habits, err = client.ListTemplates(ctx, &api.TemplateListOptions{
			TemplateType: "habit",
			Limit:        maxHabitLimit,
		})
		return err
	})

	// 4. Projects
	g.Go(func() error {
		var err error
		results.projects, err = client.ListProjects(ctx, nil)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch weekly review data: %w", err)
	}

	return results, nil
}

// buildWeeklyHabitTaskMap creates a map from template ID to date to task info
func buildWeeklyHabitTaskMap(scheduledTasks []*types.Task, habitTemplateIDs map[int]struct{}) map[int]map[string]*weeklyHabitTaskInfo {
	habitTasks := make(map[int]map[string]*weeklyHabitTaskInfo)

	for _, t := range scheduledTasks {
		if t.TemplateID == nil || t.ScheduledDate == nil {
			continue
		}
		if _, isHabit := habitTemplateIDs[*t.TemplateID]; !isHabit {
			continue
		}

		templateID := *t.TemplateID
		dateStr := t.ScheduledDate.Local().Format("2006-01-02")

		if habitTasks[templateID] == nil {
			habitTasks[templateID] = make(map[string]*weeklyHabitTaskInfo)
		}

		habitTasks[templateID][dateStr] = &weeklyHabitTaskInfo{
			taskID:    t.ID,
			completed: t.Status == "done",
		}
	}

	return habitTasks
}

// filterNonHabitTasks filters out habit tasks from a list of tasks
func filterNonHabitTasks(tasks []*types.Task, habitTemplateIDs map[int]struct{}) []*types.Task {
	var filtered []*types.Task
	for _, t := range tasks {
		if t.TemplateID != nil {
			if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
				continue
			}
		}
		filtered = append(filtered, t)
	}
	return filtered
}

// generateWeeklyReviewMarkdown generates the markdown content for the weekly review
func generateWeeklyReviewMarkdown(data *weeklyReviewData) string {
	var sb strings.Builder

	// Header with date range
	sb.WriteString(fmt.Sprintf("# Weekly Review: %s to %s\n\n",
		data.startDate.Format("01-02-2006"),
		data.endDate.Format("01-02-2006")))

	// Projects Worked On section
	writeProjectSummaries(&sb, data.completedTasks, data.projectMap)

	// Habits Summary section
	writeHabitsSummary(&sb, data)

	// Weekly Stats section
	writeWeeklyStats(&sb, data)

	return sb.String()
}

// writeProjectSummaries writes the projects section with completed tasks grouped by project
func writeProjectSummaries(sb *strings.Builder, tasks []*types.Task, projectMap map[int]string) {
	sb.WriteString("## Projects Worked On\n\n")

	if len(tasks) == 0 {
		sb.WriteString("No tasks completed this week.\n\n")
		return
	}

	// Group tasks by project
	tasksByProject := make(map[int][]*types.Task)
	for _, t := range tasks {
		tasksByProject[t.ProjectID] = append(tasksByProject[t.ProjectID], t)
	}

	// Sort project IDs for consistent output
	var projectIDs []int
	for pid := range tasksByProject {
		projectIDs = append(projectIDs, pid)
	}
	sort.Ints(projectIDs)

	for _, pid := range projectIDs {
		projectTasks := tasksByProject[pid]
		projectName := projectMap[pid]
		if projectName == "" {
			projectName = fmt.Sprintf("Project %d", pid)
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", projectName))
		sb.WriteString(fmt.Sprintf("Completed %d task(s)\n\n", len(projectTasks)))

		// Sort tasks by ID for consistent output
		sort.Slice(projectTasks, func(i, j int) bool {
			return projectTasks[i].ID < projectTasks[j].ID
		})

		for _, t := range projectTasks {
			sb.WriteString(fmt.Sprintf("- [x] #%d %s\n", t.ID, t.Title))
		}
		sb.WriteString("\n")
	}
}

// writeHabitsSummary writes the habits summary table
func writeHabitsSummary(sb *strings.Builder, data *weeklyReviewData) {
	sb.WriteString("---\n\n")
	sb.WriteString("## Habits Summary\n\n")

	if len(data.habits) == 0 {
		sb.WriteString("No habits tracked.\n\n")
		return
	}

	// Generate day headers (short day names)
	days := make([]time.Time, 7)
	for i := 0; i < 7; i++ {
		days[i] = data.startDate.AddDate(0, 0, i)
	}

	// Table header
	sb.WriteString("| Habit |")
	for _, day := range days {
		sb.WriteString(fmt.Sprintf(" %s |", day.Format("Mon")))
	}
	sb.WriteString("\n")

	// Table separator
	sb.WriteString("|-------|")
	for range days {
		sb.WriteString("-----|")
	}
	sb.WriteString("\n")

	// Table rows for each habit
	for _, habit := range data.habits {
		sb.WriteString(fmt.Sprintf("| %s |", habit.Title))

		for _, day := range days {
			dateStr := day.Format("2006-01-02")
			symbol := "-"

			if dayTasks, ok := data.habitTasks[habit.ID]; ok {
				if taskInfo, ok := dayTasks[dateStr]; ok {
					if taskInfo.completed {
						symbol = "✓"
					} else {
						symbol = "○" // scheduled but not done
					}
				}
			}

			sb.WriteString(fmt.Sprintf(" %s |", symbol))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

// writeWeeklyStats writes the weekly statistics section
func writeWeeklyStats(sb *strings.Builder, data *weeklyReviewData) {
	sb.WriteString("---\n\n")
	sb.WriteString("## Weekly Stats\n\n")

	// Count completed tasks
	tasksCompleted := len(data.completedTasks)

	// Count habit completions
	habitsCompleted := 0
	habitsPossible := 0

	for _, habit := range data.habits {
		if dayTasks, ok := data.habitTasks[habit.ID]; ok {
			for _, taskInfo := range dayTasks {
				habitsPossible++
				if taskInfo.completed {
					habitsCompleted++
				}
			}
		}
	}

	sb.WriteString(fmt.Sprintf("- **Tasks Completed**: %d\n", tasksCompleted))
	if habitsPossible > 0 {
		sb.WriteString(fmt.Sprintf("- **Habits Completed**: %d/%d\n", habitsCompleted, habitsPossible))
	} else {
		sb.WriteString("- **Habits Completed**: 0\n")
	}
}
