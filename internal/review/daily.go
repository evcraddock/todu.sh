package review

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/pkg/types"
	"golang.org/x/sync/errgroup"
)

// Constants for API limits
const (
	maxTaskLimit     = 500
	maxHabitLimit    = 100
	exportAPITimeout = 30 * time.Second
)

// dailyData holds all data needed for the daily review
type dailyData struct {
	targetDate   time.Time
	inProgress   []*types.Task
	dailyGoals   []*habitStatus
	comingUpSoon []*types.Task
	next         []*types.Task
	waiting      []*types.Task
	doneToday    []*types.Task
	projectMap   map[int]string
}

// habitStatus represents a habit and its completion status for the day
type habitStatus struct {
	name      string
	completed bool
}

// apiResults holds raw results from all API calls
type apiResults struct {
	inProgressTasks []*types.Task
	scheduledTasks  []*types.Task
	activeTasks     []*types.Task
	waitingTasks    []*types.Task
	doneTasks       []*types.Task
	highPriority    []*types.Task
	defaultProject  []*types.Task
	habits          []*types.RecurringTaskTemplate
	projects        []*types.Project
}

// DailyReport generates a daily review report and returns the markdown content
func DailyReport(ctx context.Context, client *api.Client, targetDate time.Time, defaultProject string) (string, error) {
	dateStr := targetDate.Format("2006-01-02")
	soonDate := targetDate.AddDate(0, 0, 3).Format("2006-01-02")

	// Resolve default project ID if configured
	var defaultProjectID *int
	if defaultProject != "" {
		projects, err := client.ListProjects(ctx, nil)
		if err == nil {
			lowerName := strings.ToLower(defaultProject)
			for _, p := range projects {
				if strings.ToLower(p.Name) == lowerName {
					defaultProjectID = &p.ID
					break
				}
			}
		}
	}

	// Fetch all data in parallel
	results, err := fetchDailyData(ctx, client, dateStr, soonDate, defaultProjectID)
	if err != nil {
		return "", err
	}

	// Build project map
	projectMap := buildProjectMap(results.projects)

	// Build habit template set and task map
	habitTemplateIDs := buildHabitTemplateSet(results.habits)
	habitTasks := buildHabitTaskMap(results.scheduledTasks, habitTemplateIDs)

	// Build daily goals from habits
	dailyGoals := buildDailyGoals(results.habits, habitTasks)

	// Filter done tasks to only those updated today and exclude habit tasks
	doneToday := filterDoneToday(results.doneTasks, targetDate, habitTemplateIDs)

	// Build "Next" section: high priority + scheduled today + default project (deduplicated)
	next := buildNextSection(results.highPriority, results.scheduledTasks, results.defaultProject, habitTemplateIDs)

	// Filter coming up soon to exclude habit tasks
	comingUpSoon := filterComingUpSoon(results.activeTasks, targetDate, soonDate, habitTemplateIDs)

	data := &dailyData{
		targetDate:   targetDate,
		inProgress:   results.inProgressTasks,
		dailyGoals:   dailyGoals,
		comingUpSoon: comingUpSoon,
		next:         next,
		waiting:      results.waitingTasks,
		doneToday:    doneToday,
		projectMap:   projectMap,
	}

	return generateDailyMarkdown(data), nil
}

// SaveDailyReport saves the daily review markdown to a file
func SaveDailyReport(markdown, outputPath string) error {
	outputPath = expandPath(outputPath)
	outputDir := filepath.Dir(outputPath)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	return nil
}

// DefaultDailyReportPath returns the default path for the daily review file
func DefaultDailyReportPath(localReportsPath string) string {
	return buildDailyExportPath(expandPath(localReportsPath))
}

// fetchDailyData fetches all data needed for the daily review in parallel
func fetchDailyData(ctx context.Context, client *api.Client, dateStr, soonDate string, defaultProjectID *int) (*apiResults, error) {
	ctx, cancel := context.WithTimeout(ctx, exportAPITimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	results := &apiResults{}

	// 1. In Progress tasks
	g.Go(func() error {
		var err error
		results.inProgressTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status: "inprogress",
			Limit:  maxTaskLimit,
		})
		return err
	})

	// 2. Scheduled tasks for today (for habits)
	g.Go(func() error {
		var err error
		results.scheduledTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			ScheduledDate: dateStr,
			Limit:         maxTaskLimit,
		})
		return err
	})

	// 3. Active tasks (for coming up soon - filter by due date client-side)
	g.Go(func() error {
		var err error
		results.activeTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status: "active",
			Limit:  maxTaskLimit,
		})
		return err
	})

	// 4. Waiting tasks
	g.Go(func() error {
		var err error
		results.waitingTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status: "waiting",
			Limit:  maxTaskLimit,
		})
		return err
	})

	// 5. Done tasks (filter by updated date client-side)
	g.Go(func() error {
		var err error
		results.doneTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:       "done",
			UpdatedAfter: dateStr,
			Limit:        maxTaskLimit,
		})
		return err
	})

	// 6. High priority active tasks
	g.Go(func() error {
		var err error
		results.highPriority, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:   "active",
			Priority: "high",
			Limit:    maxTaskLimit,
		})
		return err
	})

	// 7. Default project tasks (if configured)
	g.Go(func() error {
		if defaultProjectID == nil {
			return nil
		}
		var err error
		results.defaultProject, err = client.ListTasks(ctx, &api.TaskListOptions{
			ProjectID: defaultProjectID,
			Status:    "active",
			Limit:     maxTaskLimit,
		})
		return err
	})

	// 8. Habit templates
	g.Go(func() error {
		var err error
		results.habits, err = client.ListTemplates(ctx, &api.TemplateListOptions{
			TemplateType: "habit",
			Limit:        maxHabitLimit,
		})
		return err
	})

	// 9. Projects
	g.Go(func() error {
		var err error
		results.projects, err = client.ListProjects(ctx, nil)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch daily review data: %w", err)
	}

	return results, nil
}

// buildProjectMap creates a map from project ID to project name
func buildProjectMap(projects []*types.Project) map[int]string {
	projectMap := make(map[int]string)
	for _, p := range projects {
		projectMap[p.ID] = p.Name
	}
	return projectMap
}

// buildHabitTemplateSet creates a set of habit template IDs
func buildHabitTemplateSet(habits []*types.RecurringTaskTemplate) map[int]struct{} {
	habitTemplateIDs := make(map[int]struct{})
	for _, h := range habits {
		habitTemplateIDs[h.ID] = struct{}{}
	}
	return habitTemplateIDs
}

// buildHabitTaskMap creates a map from template ID to completion status
func buildHabitTaskMap(scheduledTasks []*types.Task, habitTemplateIDs map[int]struct{}) map[int]bool {
	habitTasks := make(map[int]bool)
	for _, t := range scheduledTasks {
		if t.TemplateID != nil {
			if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
				habitTasks[*t.TemplateID] = t.Status == "done"
			}
		}
	}
	return habitTasks
}

// buildDailyGoals builds the daily goals section from habits
func buildDailyGoals(habits []*types.RecurringTaskTemplate, habitTasks map[int]bool) []*habitStatus {
	var goals []*habitStatus
	for _, h := range habits {
		completed := habitTasks[h.ID] // defaults to false if not in map
		goals = append(goals, &habitStatus{
			name:      h.Title,
			completed: completed,
		})
	}
	return goals
}

// filterDoneToday filters done tasks to only those updated today, excluding habit tasks
func filterDoneToday(tasks []*types.Task, targetDate time.Time, habitTemplateIDs map[int]struct{}) []*types.Task {
	var filtered []*types.Task
	for _, t := range tasks {
		// Skip habit tasks
		if t.TemplateID != nil {
			if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
				continue
			}
		}
		// Check if updated on target date
		if isSameDay(t.UpdatedAt.Local(), targetDate) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// filterComingUpSoon filters active tasks due within the next few days
func filterComingUpSoon(tasks []*types.Task, targetDate time.Time, soonDateStr string, habitTemplateIDs map[int]struct{}) []*types.Task {
	soonDate, _ := time.ParseInLocation("2006-01-02", soonDateStr, time.Local)
	endOfSoon := soonDate.AddDate(0, 0, 1).Add(-time.Nanosecond)

	var filtered []*types.Task
	for _, t := range tasks {
		// Skip habit tasks
		if t.TemplateID != nil {
			if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
				continue
			}
		}
		// Check if has due date within range
		if t.DueDate != nil && !t.DueDate.After(endOfSoon) {
			filtered = append(filtered, t)
		}
	}

	// Sort by due date
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].DueDate == nil {
			return false
		}
		if filtered[j].DueDate == nil {
			return true
		}
		return filtered[i].DueDate.Before(*filtered[j].DueDate)
	})

	return filtered
}

// buildNextSection builds the Next section from high priority, scheduled, and default project tasks
func buildNextSection(highPriority, scheduledTasks, defaultProject []*types.Task, habitTemplateIDs map[int]struct{}) []*types.Task {
	seen := make(map[int]struct{})
	var next []*types.Task

	addTasks := func(tasks []*types.Task) {
		for _, t := range tasks {
			// Skip habit tasks
			if t.TemplateID != nil {
				if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
					continue
				}
			}
			// Skip if already added
			if _, exists := seen[t.ID]; exists {
				continue
			}
			// Only include active tasks
			if t.Status != "active" {
				continue
			}
			seen[t.ID] = struct{}{}
			next = append(next, t)
		}
	}

	addTasks(highPriority)
	addTasks(scheduledTasks)
	addTasks(defaultProject)

	// Sort by due date (earliest first, no due date last)
	sort.Slice(next, func(i, j int) bool {
		if next[i].DueDate == nil && next[j].DueDate == nil {
			return next[i].ID < next[j].ID
		}
		if next[i].DueDate == nil {
			return false
		}
		if next[j].DueDate == nil {
			return true
		}
		return next[i].DueDate.Before(*next[j].DueDate)
	})

	return next
}

// isSameDay checks if two times are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// buildDailyExportPath constructs the output file path for the daily review
func buildDailyExportPath(localReports string) string {
	return filepath.Join(localReports, "daily-review.md")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// generateDailyMarkdown generates the markdown content for the daily review
func generateDailyMarkdown(data *dailyData) string {
	var sb strings.Builder
	now := time.Now()

	sb.WriteString("# Daily Review\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", now.Format("2006-01-02 15:04")))

	// In Progress section
	sb.WriteString("## In Progress\n\n")
	if len(data.inProgress) == 0 {
		sb.WriteString("0 tasks\n\n")
	} else {
		for _, t := range data.inProgress {
			projectName := data.projectMap[t.ProjectID]
			sb.WriteString(fmt.Sprintf("- #%d %s (%s)\n", t.ID, t.Title, projectName))
		}
		sb.WriteString(fmt.Sprintf("\n%d task", len(data.inProgress)))
		if len(data.inProgress) != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n\n")
	}

	// Daily Goals section
	sb.WriteString("## Daily Goals\n\n")
	if len(data.dailyGoals) == 0 {
		sb.WriteString("0 tasks\n\n")
	} else {
		for _, h := range data.dailyGoals {
			sb.WriteString(fmt.Sprintf("- %s : %t\n", h.name, h.completed))
		}
		sb.WriteString(fmt.Sprintf("\n%d task", len(data.dailyGoals)))
		if len(data.dailyGoals) != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n\n")
	}

	// Coming up Soon section
	sb.WriteString("## Coming up Soon\n\n")
	if len(data.comingUpSoon) == 0 {
		sb.WriteString("0 tasks\n\n")
	} else {
		for _, t := range data.comingUpSoon {
			projectName := data.projectMap[t.ProjectID]
			dueStr := ""
			if t.DueDate != nil {
				dueStr = fmt.Sprintf(" - Due: %s", t.DueDate.Local().Format("2006-01-02"))
			}
			sb.WriteString(fmt.Sprintf("- #%d %s (%s)%s\n", t.ID, t.Title, projectName, dueStr))
		}
		sb.WriteString(fmt.Sprintf("\n%d task", len(data.comingUpSoon)))
		if len(data.comingUpSoon) != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n\n")
	}

	// Next section
	sb.WriteString("## Next\n\n")
	if len(data.next) == 0 {
		sb.WriteString("0 tasks\n\n")
	} else {
		for _, t := range data.next {
			projectName := data.projectMap[t.ProjectID]
			dueStr := ""
			if t.DueDate != nil {
				dueStr = fmt.Sprintf(" - Due: %s", t.DueDate.Local().Format("2006-01-02"))
			}
			sb.WriteString(fmt.Sprintf("- #%d %s (%s)%s\n", t.ID, t.Title, projectName, dueStr))
		}
		sb.WriteString(fmt.Sprintf("\n%d task", len(data.next)))
		if len(data.next) != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n\n")
	}

	// Waiting section
	sb.WriteString("## Waiting\n\n")
	if len(data.waiting) == 0 {
		sb.WriteString("0 tasks\n\n")
	} else {
		for _, t := range data.waiting {
			projectName := data.projectMap[t.ProjectID]
			sb.WriteString(fmt.Sprintf("- #%d %s (%s)\n", t.ID, t.Title, projectName))
		}
		sb.WriteString(fmt.Sprintf("\n%d task", len(data.waiting)))
		if len(data.waiting) != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n\n")
	}

	// Done Today section
	sb.WriteString("## Done Today\n\n")
	if len(data.doneToday) == 0 {
		sb.WriteString("0 tasks\n\n")
	} else {
		for _, t := range data.doneToday {
			projectName := data.projectMap[t.ProjectID]
			sb.WriteString(fmt.Sprintf("- #%d %s (%s)\n", t.ID, t.Title, projectName))
		}
		sb.WriteString(fmt.Sprintf("\n%d task", len(data.doneToday)))
		if len(data.doneToday) != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
