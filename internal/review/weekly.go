package review

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/pkg/types"
	"golang.org/x/sync/errgroup"
)

// weeklyData holds all data needed for the weekly review
type weeklyData struct {
	waiting    []*types.Task
	next       []*types.Task
	active     []*types.Task
	someday    []*types.Task
	projectMap map[int]string
}

// weeklyAPIResults holds raw results from all API calls
type weeklyAPIResults struct {
	waitingTasks      []*types.Task
	highPriorityTasks []*types.Task
	activeTasks       []*types.Task
	mediumPriority    []*types.Task
	lowPriorityTasks  []*types.Task
	projects          []*types.Project
}

// WeeklyReport generates a weekly review report and returns the markdown content
func WeeklyReport(ctx context.Context, client *api.Client) (string, error) {
	today := time.Now().Format("2006-01-02")

	// Fetch all data in parallel
	results, err := fetchWeeklyData(ctx, client)
	if err != nil {
		return "", err
	}

	// Build project map
	projectMap := buildProjectMap(results.projects)

	// Build sections
	waiting := results.waitingTasks

	// Next: high priority OR overdue (deduplicated)
	next := buildNextWeeklySection(results.highPriorityTasks, results.activeTasks, today)

	// Build set of task IDs in Next section for exclusion
	nextIDs := make(map[int]struct{})
	for _, t := range next {
		nextIDs[t.ID] = struct{}{}
	}

	// Active: medium priority OR no priority (deduplicated, excluding Next tasks)
	active := buildActiveSection(results.mediumPriority, results.activeTasks, nextIDs)

	// Someday: low priority
	someday := results.lowPriorityTasks

	data := &weeklyData{
		waiting:    waiting,
		next:       next,
		active:     active,
		someday:    someday,
		projectMap: projectMap,
	}

	return generateWeeklyMarkdown(data), nil
}

// DefaultWeeklyReportPath returns the default path for the weekly review file
func DefaultWeeklyReportPath(localReportsPath string) string {
	return buildWeeklyExportPath(expandPath(localReportsPath))
}

// buildWeeklyExportPath constructs the output file path for the weekly review
func buildWeeklyExportPath(localReports string) string {
	return localReports + "/weekly-review.md"
}

// fetchWeeklyData fetches all data needed for the weekly review in parallel
func fetchWeeklyData(ctx context.Context, client *api.Client) (*weeklyAPIResults, error) {
	ctx, cancel := context.WithTimeout(ctx, exportAPITimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	results := &weeklyAPIResults{}

	// 1. Waiting tasks
	g.Go(func() error {
		var err error
		results.waitingTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status: "waiting",
			Limit:  maxTaskLimit,
		})
		return err
	})

	// 2. High priority active tasks
	g.Go(func() error {
		var err error
		results.highPriorityTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:   "active",
			Priority: "high",
			Limit:    maxTaskLimit,
		})
		return err
	})

	// 3. All active tasks (for filtering overdue and no-priority)
	g.Go(func() error {
		var err error
		results.activeTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status: "active",
			Limit:  maxTaskLimit,
		})
		return err
	})

	// 4. Medium priority active tasks
	g.Go(func() error {
		var err error
		results.mediumPriority, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:   "active",
			Priority: "medium",
			Limit:    maxTaskLimit,
		})
		return err
	})

	// 5. Low priority active tasks
	g.Go(func() error {
		var err error
		results.lowPriorityTasks, err = client.ListTasks(ctx, &api.TaskListOptions{
			Status:   "active",
			Priority: "low",
			Limit:    maxTaskLimit,
		})
		return err
	})

	// 6. Projects
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

// buildNextWeeklySection builds the Next section from high priority and overdue tasks
func buildNextWeeklySection(highPriority, activeTasks []*types.Task, todayStr string) []*types.Task {
	today, _ := time.ParseInLocation("2006-01-02", todayStr, time.Local)
	endOfToday := today.AddDate(0, 0, 1).Add(-time.Nanosecond)

	seen := make(map[int]struct{})
	var next []*types.Task

	// Add high priority tasks
	for _, t := range highPriority {
		seen[t.ID] = struct{}{}
		next = append(next, t)
	}

	// Add overdue tasks (due before or on today)
	for _, t := range activeTasks {
		if _, exists := seen[t.ID]; exists {
			continue
		}
		if t.DueDate != nil && !t.DueDate.After(endOfToday) {
			seen[t.ID] = struct{}{}
			next = append(next, t)
		}
	}

	// Sort by due date (earliest first, no due date last)
	sortTasksByDueDate(next)

	return next
}

// buildActiveSection builds the Active section from medium priority and no-priority tasks
// excludeIDs contains task IDs that should be excluded (e.g., tasks already in Next)
func buildActiveSection(mediumPriority, activeTasks []*types.Task, excludeIDs map[int]struct{}) []*types.Task {
	seen := make(map[int]struct{})
	var active []*types.Task

	// Add medium priority tasks (excluding those in Next)
	for _, t := range mediumPriority {
		if _, excluded := excludeIDs[t.ID]; excluded {
			continue
		}
		seen[t.ID] = struct{}{}
		active = append(active, t)
	}

	// Add tasks with no priority (excluding those in Next)
	for _, t := range activeTasks {
		if _, excluded := excludeIDs[t.ID]; excluded {
			continue
		}
		if _, exists := seen[t.ID]; exists {
			continue
		}
		if t.Priority == nil {
			seen[t.ID] = struct{}{}
			active = append(active, t)
		}
	}

	// Sort by due date (earliest first, no due date last)
	sortTasksByDueDate(active)

	return active
}

// sortTasksByDueDate sorts tasks by due date (earliest first, no due date last)
func sortTasksByDueDate(tasks []*types.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].DueDate == nil && tasks[j].DueDate == nil {
			return tasks[i].ID < tasks[j].ID
		}
		if tasks[i].DueDate == nil {
			return false
		}
		if tasks[j].DueDate == nil {
			return true
		}
		return tasks[i].DueDate.Before(*tasks[j].DueDate)
	})
}

// generateWeeklyMarkdown generates the markdown content for the weekly review
func generateWeeklyMarkdown(data *weeklyData) string {
	var sb strings.Builder
	now := time.Now()

	sb.WriteString("# Weekly Review\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", now.Format("2006-01-02 15:04")))

	// Waiting section
	sb.WriteString("## Waiting\n\n")
	writeTaskSection(&sb, data.waiting, data.projectMap)

	// Next section
	sb.WriteString("## Next\n\n")
	writeTaskSection(&sb, data.next, data.projectMap)

	// Active section
	sb.WriteString("## Active\n\n")
	writeTaskSection(&sb, data.active, data.projectMap)

	// Someday section
	sb.WriteString("## Someday\n\n")
	writeTaskSectionNoTrailingNewline(&sb, data.someday, data.projectMap)

	return sb.String()
}

// writeTaskSection writes a section of tasks to the string builder
func writeTaskSection(sb *strings.Builder, tasks []*types.Task, projectMap map[int]string) {
	if len(tasks) == 0 {
		sb.WriteString("0 tasks\n\n")
		return
	}

	for _, t := range tasks {
		projectName := projectMap[t.ProjectID]
		dueStr := ""
		if t.DueDate != nil {
			dueStr = fmt.Sprintf(" - Due: %s", t.DueDate.Local().Format("2006-01-02"))
		}
		sb.WriteString(fmt.Sprintf("- #%d %s (%s)%s\n", t.ID, t.Title, projectName, dueStr))
	}
	sb.WriteString(fmt.Sprintf("\n%d task", len(tasks)))
	if len(tasks) != 1 {
		sb.WriteString("s")
	}
	sb.WriteString("\n\n")
}

// writeTaskSectionNoTrailingNewline writes a section without trailing double newline
func writeTaskSectionNoTrailingNewline(sb *strings.Builder, tasks []*types.Task, projectMap map[int]string) {
	if len(tasks) == 0 {
		sb.WriteString("0 tasks\n")
		return
	}

	for _, t := range tasks {
		projectName := projectMap[t.ProjectID]
		dueStr := ""
		if t.DueDate != nil {
			dueStr = fmt.Sprintf(" - Due: %s", t.DueDate.Local().Format("2006-01-02"))
		}
		sb.WriteString(fmt.Sprintf("- #%d %s (%s)%s\n", t.ID, t.Title, projectName, dueStr))
	}
	sb.WriteString(fmt.Sprintf("\n%d task", len(tasks)))
	if len(tasks) != 1 {
		sb.WriteString("s")
	}
	sb.WriteString("\n")
}
