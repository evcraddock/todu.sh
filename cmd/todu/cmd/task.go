package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long: `Manage tasks in todu.

Tasks represent work items that can be tracked across different projects
and synchronized with external systems.`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks from todu with optional filtering.

Displays tasks in a table format with key information. Use filters to
narrow down results to specific projects, statuses, or other criteria.`,
	RunE: runTaskList,
}

var taskShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Long:  `Display detailed information about a specific task including comments.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskShow,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long: `Create a new task in todu.

Tasks must belong to a project. Use --project to specify which project
the task belongs to.`,
	RunE: runTaskCreate,
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a task",
	Long: `Update task fields.

Only the specified fields will be updated. Other fields will remain unchanged.`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskUpdate,
}

var taskCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a task",
	Long:  `Mark a task as done/closed. This is a shortcut for updating the status to "done".`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskClose,
}

var taskCommentCmd = &cobra.Command{
	Use:   "comment <id> <text>",
	Short: "Add a comment to a task",
	Long:  `Add a comment to a task. The comment text can be provided as an argument or via the --message flag.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskComment,
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Long:  `Delete a task from todu. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDelete,
}

var (
	// List flags
	taskListStatus     string
	taskListPriority   string
	taskListProject    int
	taskListAssignee   string
	taskListLabels     []string
	taskListSearch     string
	taskListDueBefore  string
	taskListDueAfter   string
	taskListLimit      int
	taskListFormat     string

	// Create flags
	taskCreateTitle       string
	taskCreateProject     int
	taskCreateDescription string
	taskCreateStatus      string
	taskCreatePriority    string
	taskCreateDue         string
	taskCreateLabels      []string
	taskCreateAssignees   []string
	taskCreateExternalID  string

	// Update flags
	taskUpdateTitle          string
	taskUpdateDescription    string
	taskUpdateStatus         string
	taskUpdatePriority       string
	taskUpdateDue            string
	taskUpdateAddLabels      []string
	taskUpdateRemoveLabels   []string
	taskUpdateAddAssignees   []string
	taskUpdateRemoveAssignees []string

	// Comment flags
	taskCommentMessage string
	taskCommentAuthor  string

	// Delete flags
	taskDeleteForce bool
)

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskCloseCmd)
	taskCmd.AddCommand(taskCommentCmd)
	taskCmd.AddCommand(taskDeleteCmd)

	// List flags
	taskListCmd.Flags().StringVar(&taskListStatus, "status", "", "Filter by status")
	taskListCmd.Flags().StringVar(&taskListPriority, "priority", "", "Filter by priority")
	taskListCmd.Flags().IntVarP(&taskListProject, "project", "p", 0, "Filter by project ID")
	taskListCmd.Flags().StringVar(&taskListAssignee, "assignee", "", "Filter by assignee")
	taskListCmd.Flags().StringSliceVar(&taskListLabels, "label", []string{}, "Filter by label (repeatable)")
	taskListCmd.Flags().StringVar(&taskListSearch, "search", "", "Full-text search")
	taskListCmd.Flags().StringVar(&taskListDueBefore, "due-before", "", "Due before date (YYYY-MM-DD)")
	taskListCmd.Flags().StringVar(&taskListDueAfter, "due-after", "", "Due after date (YYYY-MM-DD)")
	taskListCmd.Flags().IntVar(&taskListLimit, "limit", 50, "Limit number of results")
	taskListCmd.Flags().StringVar(&taskListFormat, "format", "text", "Output format (text|json)")

	// Create flags
	taskCreateCmd.Flags().StringVar(&taskCreateTitle, "title", "", "Task title (required)")
	taskCreateCmd.Flags().IntVarP(&taskCreateProject, "project", "p", 0, "Project ID (required)")
	taskCreateCmd.Flags().StringVar(&taskCreateDescription, "description", "", "Task description")
	taskCreateCmd.Flags().StringVar(&taskCreateStatus, "status", "active", "Task status")
	taskCreateCmd.Flags().StringVar(&taskCreatePriority, "priority", "", "Task priority")
	taskCreateCmd.Flags().StringVar(&taskCreateDue, "due", "", "Due date (YYYY-MM-DD)")
	taskCreateCmd.Flags().StringSliceVar(&taskCreateLabels, "label", []string{}, "Task label (repeatable)")
	taskCreateCmd.Flags().StringSliceVar(&taskCreateAssignees, "assignee", []string{}, "Task assignee (repeatable)")
	taskCreateCmd.Flags().StringVar(&taskCreateExternalID, "external-id", "", "External ID")

	// Update flags
	taskUpdateCmd.Flags().StringVar(&taskUpdateTitle, "title", "", "Update task title")
	taskUpdateCmd.Flags().StringVar(&taskUpdateDescription, "description", "", "Update task description")
	taskUpdateCmd.Flags().StringVar(&taskUpdateStatus, "status", "", "Update task status")
	taskUpdateCmd.Flags().StringVar(&taskUpdatePriority, "priority", "", "Update task priority")
	taskUpdateCmd.Flags().StringVar(&taskUpdateDue, "due", "", "Update due date (YYYY-MM-DD)")
	taskUpdateCmd.Flags().StringSliceVar(&taskUpdateAddLabels, "add-label", []string{}, "Add label (repeatable)")
	taskUpdateCmd.Flags().StringSliceVar(&taskUpdateRemoveLabels, "remove-label", []string{}, "Remove label (repeatable)")
	taskUpdateCmd.Flags().StringSliceVar(&taskUpdateAddAssignees, "add-assignee", []string{}, "Add assignee (repeatable)")
	taskUpdateCmd.Flags().StringSliceVar(&taskUpdateRemoveAssignees, "remove-assignee", []string{}, "Remove assignee (repeatable)")

	// Comment flags
	taskCommentCmd.Flags().StringVarP(&taskCommentMessage, "message", "m", "", "Comment message")
	taskCommentCmd.Flags().StringVar(&taskCommentAuthor, "author", "user", "Comment author")

	// Delete flags
	taskDeleteCmd.Flags().BoolVarP(&taskDeleteForce, "force", "f", false, "Skip confirmation")
}

func runTaskList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// For now, we'll just list all tasks and filter client-side
	// In the future, the API could support query parameters for filtering
	var projectIDPtr *int
	if taskListProject > 0 {
		projectIDPtr = &taskListProject
	}

	tasks, err := apiClient.ListTasks(ctx, projectIDPtr)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Apply filters
	tasks = filterTasks(tasks)

	// Limit results
	if taskListLimit > 0 && len(tasks) > taskListLimit {
		tasks = tasks[:taskListLimit]
	}

	// Display results
	if taskListFormat == "json" {
		return displayTasksJSON(tasks)
	}

	return displayTasksTable(tasks)
}

func filterTasks(tasks []*types.Task) []*types.Task {
	var filtered []*types.Task

	for _, task := range tasks {
		// Status filter
		if taskListStatus != "" && task.Status != taskListStatus {
			continue
		}

		// Priority filter
		if taskListPriority != "" {
			if task.Priority == nil || *task.Priority != taskListPriority {
				continue
			}
		}

		// Assignee filter
		if taskListAssignee != "" {
			found := false
			for _, assignee := range task.Assignees {
				if assignee.Name == taskListAssignee {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Label filter
		if len(taskListLabels) > 0 {
			hasAllLabels := true
			for _, filterLabel := range taskListLabels {
				found := false
				for _, taskLabel := range task.Labels {
					if taskLabel.Name == filterLabel {
						found = true
						break
					}
				}
				if !found {
					hasAllLabels = false
					break
				}
			}
			if !hasAllLabels {
				continue
			}
		}

		// Search filter
		if taskListSearch != "" {
			searchLower := strings.ToLower(taskListSearch)
			titleLower := strings.ToLower(task.Title)
			descLower := ""
			if task.Description != nil {
				descLower = strings.ToLower(*task.Description)
			}

			if !strings.Contains(titleLower, searchLower) && !strings.Contains(descLower, searchLower) {
				continue
			}
		}

		// Due date filters
		if taskListDueBefore != "" && task.DueDate != nil {
			beforeDate, err := time.Parse("2006-01-02", taskListDueBefore)
			if err == nil && task.DueDate.After(beforeDate) {
				continue
			}
		}

		if taskListDueAfter != "" && task.DueDate != nil {
			afterDate, err := time.Parse("2006-01-02", taskListDueAfter)
			if err == nil && task.DueDate.Before(afterDate) {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

func displayTasksTable(tasks []*types.Task) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPRIORITY\tPROJECT\tDUE DATE")
	fmt.Fprintln(w, "--\t-----\t------\t--------\t-------\t--------")

	for _, task := range tasks {
		priority := ""
		if task.Priority != nil {
			priority = *task.Priority
		}

		dueDate := ""
		if task.DueDate != nil {
			dueDate = task.DueDate.Format("2006-01-02")
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%s\n",
			task.ID,
			truncate(task.Title, 40),
			task.Status,
			priority,
			task.ProjectID,
			dueDate,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d tasks\n", len(tasks))
	return nil
}

func displayTasksJSON(tasks []*types.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	task, err := apiClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get comments
	comments, err := apiClient.ListComments(ctx, taskID)
	if err != nil {
		// Don't fail if comments can't be fetched
		comments = []*types.Comment{}
	}

	displayTask(task, comments)
	return nil
}

func displayTask(task *types.Task, comments []*types.Comment) {
	fmt.Printf("Task #%d: %s\n", task.ID, task.Title)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("Status:      %s\n", task.Status)
	if task.Priority != nil {
		fmt.Printf("Priority:    %s\n", *task.Priority)
	}
	fmt.Printf("Project ID:  %d\n", task.ProjectID)
	fmt.Printf("External ID: %s\n", task.ExternalID)

	if task.SourceURL != nil {
		fmt.Printf("Source URL:  %s\n", *task.SourceURL)
	}

	if task.DueDate != nil {
		fmt.Printf("Due Date:    %s\n", task.DueDate.Format("2006-01-02"))
	}

	fmt.Printf("Created:     %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:     %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))

	if task.Description != nil && *task.Description != "" {
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println(*task.Description)
	}

	if len(task.Labels) > 0 {
		fmt.Println()
		fmt.Print("Labels: ")
		labelNames := make([]string, len(task.Labels))
		for i, label := range task.Labels {
			labelNames[i] = label.Name
		}
		fmt.Println(strings.Join(labelNames, ", "))
	}

	if len(task.Assignees) > 0 {
		fmt.Println()
		fmt.Print("Assignees: ")
		assigneeNames := make([]string, len(task.Assignees))
		for i, assignee := range task.Assignees {
			assigneeNames[i] = assignee.Name
		}
		fmt.Println(strings.Join(assigneeNames, ", "))
	}

	if len(comments) > 0 {
		fmt.Println()
		fmt.Printf("Comments (%d):\n", len(comments))
		fmt.Println(strings.Repeat("-", 60))
		for _, comment := range comments {
			fmt.Printf("\n[%s] %s:\n", comment.CreatedAt.Format("2006-01-02 15:04"), comment.Author)
			fmt.Println(comment.Content)
		}
	}
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	// Validate required flags
	if taskCreateTitle == "" {
		return fmt.Errorf("--title is required")
	}
	if taskCreateProject == 0 {
		return fmt.Errorf("--project is required")
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Build task create request
	taskCreate := &types.TaskCreate{
		Title:     taskCreateTitle,
		ProjectID: taskCreateProject,
		Status:    taskCreateStatus,
	}

	if taskCreateExternalID != "" {
		taskCreate.ExternalID = taskCreateExternalID
	}

	if taskCreateDescription != "" {
		taskCreate.Description = &taskCreateDescription
	}

	if taskCreatePriority != "" {
		taskCreate.Priority = &taskCreatePriority
	}

	if taskCreateDue != "" {
		dueDate, err := time.Parse("2006-01-02", taskCreateDue)
		if err != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
		}
		taskCreate.DueDate = &dueDate
	}

	// Add labels
	if len(taskCreateLabels) > 0 {
		labels := make([]types.Label, len(taskCreateLabels))
		for i, name := range taskCreateLabels {
			labels[i] = types.Label{Name: name}
		}
		taskCreate.Labels = labels
	}

	// Add assignees
	if len(taskCreateAssignees) > 0 {
		assignees := make([]types.Assignee, len(taskCreateAssignees))
		for i, name := range taskCreateAssignees {
			assignees[i] = types.Assignee{Name: name}
		}
		taskCreate.Assignees = assignees
	}

	task, err := apiClient.CreateTask(ctx, taskCreate)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Println("Task created successfully:")
	displayTask(task, []*types.Comment{})
	return nil
}

func runTaskUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Get current task to merge labels/assignees
	currentTask, err := apiClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Build update request
	taskUpdate := &types.TaskUpdate{}

	if taskUpdateTitle != "" {
		taskUpdate.Title = &taskUpdateTitle
	}

	if taskUpdateDescription != "" {
		taskUpdate.Description = &taskUpdateDescription
	}

	if taskUpdateStatus != "" {
		taskUpdate.Status = &taskUpdateStatus
	}

	if taskUpdatePriority != "" {
		taskUpdate.Priority = &taskUpdatePriority
	}

	if taskUpdateDue != "" {
		dueDate, err := time.Parse("2006-01-02", taskUpdateDue)
		if err != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
		}
		taskUpdate.DueDate = &dueDate
	}

	// Handle labels
	if len(taskUpdateAddLabels) > 0 || len(taskUpdateRemoveLabels) > 0 {
		labels := currentTask.Labels

		// Add new labels
		for _, name := range taskUpdateAddLabels {
			// Check if already exists
			exists := false
			for _, label := range labels {
				if label.Name == name {
					exists = true
					break
				}
			}
			if !exists {
				labels = append(labels, types.Label{Name: name})
			}
		}

		// Remove labels
		for _, name := range taskUpdateRemoveLabels {
			var newLabels []types.Label
			for _, label := range labels {
				if label.Name != name {
					newLabels = append(newLabels, label)
				}
			}
			labels = newLabels
		}

		taskUpdate.Labels = labels
	}

	// Handle assignees
	if len(taskUpdateAddAssignees) > 0 || len(taskUpdateRemoveAssignees) > 0 {
		assignees := currentTask.Assignees

		// Add new assignees
		for _, name := range taskUpdateAddAssignees {
			exists := false
			for _, assignee := range assignees {
				if assignee.Name == name {
					exists = true
					break
				}
			}
			if !exists {
				assignees = append(assignees, types.Assignee{Name: name})
			}
		}

		// Remove assignees
		for _, name := range taskUpdateRemoveAssignees {
			var newAssignees []types.Assignee
			for _, assignee := range assignees {
				if assignee.Name != name {
					newAssignees = append(newAssignees, assignee)
				}
			}
			assignees = newAssignees
		}

		taskUpdate.Assignees = assignees
	}

	task, err := apiClient.UpdateTask(ctx, taskID, taskUpdate)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Println("Task updated successfully:")
	displayTask(task, []*types.Comment{})
	return nil
}

func runTaskClose(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	status := "done"
	taskUpdate := &types.TaskUpdate{
		Status: &status,
	}

	task, err := apiClient.UpdateTask(ctx, taskID, taskUpdate)
	if err != nil {
		return fmt.Errorf("failed to close task: %w", err)
	}

	fmt.Printf("Task #%d closed successfully\n", task.ID)
	return nil
}

func runTaskComment(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Get comment text from args or flag
	var commentText string
	if taskCommentMessage != "" {
		commentText = taskCommentMessage
	} else if len(args) > 1 {
		commentText = strings.Join(args[1:], " ")
	} else {
		return fmt.Errorf("comment text required (provide as argument or via --message flag)")
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	commentCreate := &types.CommentCreate{
		TaskID:  taskID,
		Content: commentText,
		Author:  taskCommentAuthor,
	}

	comment, err := apiClient.CreateComment(ctx, commentCreate)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	fmt.Printf("Comment added to task #%d:\n", taskID)
	fmt.Printf("[%s] %s:\n", comment.CreatedAt.Format("2006-01-02 15:04"), comment.Author)
	fmt.Println(comment.Content)
	return nil
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Confirm deletion unless --force
	if !taskDeleteForce {
		fmt.Printf("Are you sure you want to delete task #%d? (y/N): ", taskID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	err = apiClient.DeleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Task #%d deleted successfully\n", taskID)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
