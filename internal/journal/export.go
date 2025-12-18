package journal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/pkg/types"
	"golang.org/x/sync/errgroup"
)

// Export constants
const (
	maxJournalLimit  = 100
	maxTaskLimit     = 500
	maxHabitLimit    = 100
	exportAPITimeout = 30 * time.Second
)

// habitTaskInfo holds task information for a habit on a specific day
type habitTaskInfo struct {
	taskID    int
	completed bool
}

// exportData holds all data needed for export
type exportData struct {
	targetDate     time.Time
	journals       []*types.Comment
	completedTasks []*types.Task
	habits         []*types.RecurringTaskTemplate
	projectMap     map[int]string
	habitTasks     map[int]habitTaskInfo
}

// apiResults holds the raw results from all API calls
type apiResults struct {
	journals       []*types.Comment
	doneTasks      []*types.Task
	habits         []*types.RecurringTaskTemplate
	projects       []*types.Project
	scheduledTasks []*types.Task
}

// Export exports the journal for a specific date to a markdown file
func Export(ctx context.Context, client *api.Client, targetDate time.Time, localReportsPath string) (string, error) {
	if localReportsPath == "" {
		return "", fmt.Errorf("local_reports path not configured")
	}

	// Fetch all data from API in parallel
	dateStr := targetDate.Format("2006-01-02")
	results, err := fetchData(ctx, client, dateStr)
	if err != nil {
		return "", err
	}

	// Process fetched data
	projectMap := buildProjectMap(results.projects)
	habitTemplateIDs := buildHabitTemplateSet(results.habits)
	habitTasks := buildHabitTaskMap(results.scheduledTasks, habitTemplateIDs)

	data := &exportData{
		targetDate:     targetDate,
		journals:       filterJournalsByTargetDate(results.journals, targetDate),
		completedTasks: filterTasksByTargetDate(results.doneTasks, targetDate),
		habits:         results.habits,
		projectMap:     projectMap,
		habitTasks:     habitTasks,
	}

	// Generate markdown
	markdown := generateMarkdown(data)

	// Write to file
	outputPath := buildExportPath(expandPath(localReportsPath), targetDate)
	outputDir := filepath.Dir(outputPath)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	return outputPath, nil
}

// fetchData fetches all data needed for export in parallel
func fetchData(ctx context.Context, client *api.Client, dateStr string) (*apiResults, error) {
	ctx, cancel := context.WithTimeout(ctx, exportAPITimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	results := &apiResults{}

	// 1. Fetch journals
	g.Go(func() error {
		var err error
		results.journals, err = client.ListJournals(ctx, 0, maxJournalLimit)
		return err
	})

	// 2. Fetch completed tasks
	g.Go(func() error {
		var err error
		results.doneTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:       "done",
			UpdatedAfter: dateStr,
			Limit:        maxTaskLimit,
		})
		return err
	})

	// 3. Fetch habit templates
	g.Go(func() error {
		var err error
		results.habits, err = client.ListTemplates(ctx, &api.TemplateListOptions{
			TemplateType: "habit",
			Limit:        maxHabitLimit,
		})
		return err
	})

	// 4. Fetch projects
	g.Go(func() error {
		var err error
		results.projects, err = client.ListProjects(ctx, nil)
		return err
	})

	// 5. Fetch scheduled tasks
	g.Go(func() error {
		var err error
		results.scheduledTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			ScheduledDate: dateStr,
			Limit:         maxTaskLimit,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch export data: %w", err)
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

// buildHabitTaskMap creates a map from template ID to task info for scheduled habit tasks
func buildHabitTaskMap(scheduledTasks []*types.Task, habitTemplateIDs map[int]struct{}) map[int]habitTaskInfo {
	habitTasks := make(map[int]habitTaskInfo)
	for _, t := range scheduledTasks {
		if t.TemplateID != nil {
			if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
				habitTasks[*t.TemplateID] = habitTaskInfo{
					taskID:    t.ID,
					completed: t.Status == "done",
				}
			}
		}
	}
	return habitTasks
}

// filterJournalsByTargetDate filters journals to only those created on the target date
func filterJournalsByTargetDate(journals []*types.Comment, targetDate time.Time) []*types.Comment {
	var filtered []*types.Comment
	for _, j := range journals {
		if isSameDay(j.CreatedAt.Local(), targetDate) {
			filtered = append(filtered, j)
		}
	}
	return filtered
}

// filterTasksByTargetDate filters tasks to only those updated on the target date
func filterTasksByTargetDate(tasks []*types.Task, targetDate time.Time) []*types.Task {
	var filtered []*types.Task
	for _, t := range tasks {
		if isSameDay(t.UpdatedAt.Local(), targetDate) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// isSameDay checks if two times are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// buildExportPath constructs the output file path for the export
func buildExportPath(localReports string, targetDate time.Time) string {
	year := targetDate.Format("2006")
	monthDir := targetDate.Format("01-January")
	fileName := targetDate.Format("01-02-2006") + "-journal.md"
	return filepath.Join(localReports, "reviews", year, monthDir, fileName)
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

// escapeMarkdown escapes special markdown characters in text
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`*`, `\*`,
		`_`, `\_`,
		`[`, `\[`,
		`]`, `\]`,
		`#`, `\#`,
		`|`, `\|`,
		"`", "\\`",
	)
	return replacer.Replace(s)
}

// formatTimezone formats a timezone offset in seconds to a string like "CT" or "+0530"
func formatTimezone(offset int) string {
	// Common US timezone abbreviations (using standard time names)
	switch offset {
	case -5 * 3600:
		return "ET" // Eastern Time
	case -6 * 3600:
		return "CT" // Central Time
	case -7 * 3600:
		return "MT" // Mountain Time
	case -8 * 3600:
		return "PT" // Pacific Time
	case -4 * 3600:
		return "AT" // Atlantic Time
	}

	// Default: format as +/-HHMM
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	if minutes < 0 {
		minutes = -minutes
	}
	return fmt.Sprintf("%+03d%02d", hours, minutes)
}

// generateMarkdown generates the markdown content for the journal export
func generateMarkdown(data *exportData) string {
	var sb strings.Builder
	dateFormatted := data.targetDate.Format("01-02-2006")

	sb.WriteString(fmt.Sprintf("# %s Journal\n\n", dateFormatted))

	// Journal entries section
	for _, j := range data.journals {
		_, offset := j.CreatedAt.Local().Zone()
		tz := formatTimezone(offset)
		timeStr := j.CreatedAt.Local().Format("15:04")
		sb.WriteString(fmt.Sprintf("- #### time: %s %s\n", timeStr, tz))
		sb.WriteString(fmt.Sprintf("%s\n\n", j.Content))
	}

	// Completed Today section (exclude habit tasks)
	sb.WriteString("## Completed Today\n")
	habitTemplateIDs := make(map[int]struct{})
	for _, h := range data.habits {
		habitTemplateIDs[h.ID] = struct{}{}
	}

	hasCompletedTasks := false
	for _, t := range data.completedTasks {
		// Skip tasks that are from habit templates
		if t.TemplateID != nil {
			if _, isHabit := habitTemplateIDs[*t.TemplateID]; isHabit {
				continue
			}
		}
		hasCompletedTasks = true
		projectName := data.projectMap[t.ProjectID]
		priority := "medium"
		if t.Priority != nil {
			priority = *t.Priority
		}
		sb.WriteString(fmt.Sprintf("- [x] #%d %s - %s (priority: %s)\n", t.ID, escapeMarkdown(t.Title), projectName, priority))
	}
	if !hasCompletedTasks {
		sb.WriteString("No Tasks\n")
	}
	sb.WriteString("\n")

	// Habits section - only include habits with instantiated tasks for this day
	sb.WriteString("## Habits\n")
	hasHabits := false
	for _, h := range data.habits {
		if info, hasTask := data.habitTasks[h.ID]; hasTask {
			hasHabits = true
			projectName := data.projectMap[h.ProjectID]
			sb.WriteString(fmt.Sprintf("- #%d %s - %s:: %t\n", info.taskID, projectName, escapeMarkdown(h.Title), info.completed))
		}
	}
	if !hasHabits {
		sb.WriteString("No Habits\n")
	}

	return sb.String()
}
