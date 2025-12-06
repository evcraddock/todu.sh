package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/spf13/cobra"
	"github.com/teambition/rrule-go"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage recurring task templates",
	Long: `Manage recurring task templates in todu.

Recurring task templates define patterns for automatically creating tasks
based on RRULE recurrence rules (RFC 5545). When a task linked to a template
is marked as done, a new task is automatically generated for the next occurrence.

Template types:
  - task: Standard recurring tasks
  - habit: Habit tracking tasks`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recurring task templates",
	Long: `List recurring task templates with optional filtering.

Examples:
  todu template list
  todu template list --active
  todu template list --type habit
  todu template list --project myproject`,
	RunE: runTemplateList,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show template details",
	Long:  `Display detailed information about a specific recurring task template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new recurring task template",
	Long: `Create a new recurring task template.

The recurrence rule must be in RRULE format (RFC 5545).

Common RRULE examples:
  Daily:                FREQ=DAILY;INTERVAL=1
  Every 3 days:         FREQ=DAILY;INTERVAL=3
  Weekly on Monday:     FREQ=WEEKLY;BYDAY=MO
  Weekly Mon/Wed/Fri:   FREQ=WEEKLY;BYDAY=MO,WE,FR
  Every weekday:        FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR
  Monthly on 1st:       FREQ=MONTHLY;BYMONTHDAY=1
  Monthly on 15th:      FREQ=MONTHLY;BYMONTHDAY=15
  Yearly on Jan 1:      FREQ=YEARLY;BYMONTH=1;BYMONTHDAY=1

Examples:
  todu template create --title "Daily standup" --recurrence "FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR" --start-date 2024-01-01 --project myproject
  todu template create --title "Weekly review" --recurrence "FREQ=WEEKLY;BYDAY=FR" --start-date 2024-01-05 --type habit`,
	RunE: runTemplateCreate,
}

var templateUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a recurring task template",
	Long: `Update fields of an existing recurring task template.

Only the specified fields will be updated. Other fields will remain unchanged.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateUpdate,
}

var templateActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a recurring task template",
	Long:  `Set a template's is_active status to true, enabling automatic task generation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateActivate,
}

var templateDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a recurring task template",
	Long:  `Set a template's is_active status to false, stopping automatic task generation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateDeactivate,
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a recurring task template",
	Long:  `Delete a recurring task template. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateDelete,
}

var (
	// List flags
	templateListActive  string
	templateListType    string
	templateListProject string
	templateListSkip    int
	templateListLimit   int

	// Create flags
	templateCreateTitle       string
	templateCreateProject     string
	templateCreateDescription string
	templateCreatePriority    string
	templateCreateRecurrence  string
	templateCreateStartDate   string
	templateCreateEndDate     string
	templateCreateTimezone    string
	templateCreateType        string
	templateCreateLabels      []string
	templateCreateAssignees   []string

	// Update flags
	templateUpdateTitle       string
	templateUpdateDescription string
	templateUpdatePriority    string
	templateUpdateRecurrence  string
	templateUpdateEndDate     string
	templateUpdateTimezone    string
	templateUpdateLabels      []string
	templateUpdateAssignees   []string

	// Delete flags
	templateDeleteForce bool
)

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateUpdateCmd)
	templateCmd.AddCommand(templateActivateCmd)
	templateCmd.AddCommand(templateDeactivateCmd)
	templateCmd.AddCommand(templateDeleteCmd)

	// List flags
	templateListCmd.Flags().StringVar(&templateListActive, "active", "", "Filter by active status (true/false)")
	templateListCmd.Flags().StringVar(&templateListType, "type", "", "Filter by template type (task/habit)")
	templateListCmd.Flags().StringVarP(&templateListProject, "project", "p", "", "Filter by project ID or name")
	templateListCmd.Flags().IntVar(&templateListSkip, "skip", 0, "Number of items to skip (pagination)")
	templateListCmd.Flags().IntVar(&templateListLimit, "limit", 50, "Limit number of results")

	// Create flags
	templateCreateCmd.Flags().StringVar(&templateCreateTitle, "title", "", "Template title (required)")
	templateCreateCmd.Flags().StringVarP(&templateCreateProject, "project", "p", "", "Project ID or name (required)")
	templateCreateCmd.Flags().StringVar(&templateCreateDescription, "description", "", "Template description")
	templateCreateCmd.Flags().StringVar(&templateCreatePriority, "priority", "", "Task priority (low/medium/high)")
	templateCreateCmd.Flags().StringVar(&templateCreateRecurrence, "recurrence", "", "Recurrence rule in RRULE format (required)")
	templateCreateCmd.Flags().StringVar(&templateCreateStartDate, "start-date", "", "Start date (YYYY-MM-DD) (required)")
	templateCreateCmd.Flags().StringVar(&templateCreateEndDate, "end-date", "", "End date (YYYY-MM-DD)")
	templateCreateCmd.Flags().StringVar(&templateCreateTimezone, "timezone", "UTC", "IANA timezone (e.g., America/New_York, Europe/London)")
	templateCreateCmd.Flags().StringVar(&templateCreateType, "type", "task", "Template type (task/habit)")
	templateCreateCmd.Flags().StringSliceVar(&templateCreateLabels, "label", []string{}, "Template label (repeatable)")
	templateCreateCmd.Flags().StringSliceVar(&templateCreateAssignees, "assignee", []string{}, "Template assignee (repeatable)")

	// Update flags
	templateUpdateCmd.Flags().StringVar(&templateUpdateTitle, "title", "", "Update template title")
	templateUpdateCmd.Flags().StringVar(&templateUpdateDescription, "description", "", "Update template description")
	templateUpdateCmd.Flags().StringVar(&templateUpdatePriority, "priority", "", "Update task priority")
	templateUpdateCmd.Flags().StringVar(&templateUpdateRecurrence, "recurrence", "", "Update recurrence rule")
	templateUpdateCmd.Flags().StringVar(&templateUpdateEndDate, "end-date", "", "Update end date (YYYY-MM-DD)")
	templateUpdateCmd.Flags().StringVar(&templateUpdateTimezone, "timezone", "", "Update IANA timezone")
	templateUpdateCmd.Flags().StringSliceVar(&templateUpdateLabels, "label", []string{}, "Replace labels (repeatable)")
	templateUpdateCmd.Flags().StringSliceVar(&templateUpdateAssignees, "assignee", []string{}, "Replace assignees (repeatable)")

	// Delete flags
	templateDeleteCmd.Flags().BoolVarP(&templateDeleteForce, "force", "f", false, "Skip confirmation")
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Build API options with filters
	opts := &api.TemplateListOptions{
		Skip:  templateListSkip,
		Limit: templateListLimit,
	}

	// Parse active filter
	if templateListActive != "" {
		active, err := strconv.ParseBool(templateListActive)
		if err != nil {
			return fmt.Errorf("invalid --active value (use true/false): %w", err)
		}
		opts.Active = &active
	}

	// Parse type filter
	if templateListType != "" {
		if templateListType != "task" && templateListType != "habit" {
			return fmt.Errorf("invalid --type value: must be 'task' or 'habit'")
		}
		opts.TemplateType = templateListType
	}

	// Resolve project ID from name or ID if provided
	if templateListProject != "" {
		projectID, err := resolveProjectID(ctx, apiClient, templateListProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}
		opts.ProjectID = &projectID
	}

	templates, err := apiClient.ListTemplates(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	// Display results
	if GetOutputFormat() == "json" {
		return displayTemplatesJSON(templates)
	}

	return displayTemplatesTable(ctx, apiClient, templates)
}

func displayTemplatesTable(ctx context.Context, apiClient *api.Client, templates []*types.RecurringTaskTemplate) error {
	if len(templates) == 0 {
		fmt.Println("No templates found")
		return nil
	}

	// Fetch projects for name lookup
	projectNames := make(map[int]string)
	projects, err := apiClient.ListProjects(ctx, nil)
	if err == nil {
		for _, p := range projects {
			projectNames[p.ID] = p.Name
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tRECURRENCE\tTYPE\tACTIVE\tPROJECT")
	fmt.Fprintln(w, "--\t-----\t----------\t----\t------\t-------")

	for _, tmpl := range templates {
		// Use project name if available
		projectDisplay := projectNames[tmpl.ProjectID]
		if projectDisplay == "" {
			projectDisplay = fmt.Sprintf("%d", tmpl.ProjectID)
		}

		activeDisplay := "no"
		if tmpl.IsActive {
			activeDisplay = "yes"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			tmpl.ID,
			truncate(tmpl.Title, 30),
			truncate(rruleToHuman(tmpl.RecurrenceRule), 25),
			tmpl.TemplateType,
			activeDisplay,
			projectDisplay,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d templates\n", len(templates))
	return nil
}

func displayTemplatesJSON(templates []*types.RecurringTaskTemplate) error {
	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	templateID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	template, err := apiClient.GetTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Get tasks associated with this template
	var associatedTasks []*types.Task
	tasks, err := apiClient.ListTasks(ctx, &api.TaskListOptions{
		ProjectID: &template.ProjectID,
		Limit:     100,
	})
	if err == nil {
		for _, task := range tasks {
			if task.TemplateID != nil && *task.TemplateID == templateID {
				associatedTasks = append(associatedTasks, task)
			}
		}
	}

	// Display results
	if GetOutputFormat() == "json" {
		return displayTemplateWithTasksJSON(template, associatedTasks)
	}

	displayTemplate(template)
	displayAssociatedTasks(associatedTasks)
	return nil
}

func displayTemplateJSON(template *types.RecurringTaskTemplate) error {
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func displayTemplateWithTasksJSON(template *types.RecurringTaskTemplate, tasks []*types.Task) error {
	output := map[string]interface{}{
		"template": template,
		"tasks":    tasks,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func displayAssociatedTasks(tasks []*types.Task) {
	if len(tasks) == 0 {
		return
	}

	fmt.Println()
	fmt.Printf("Associated Tasks (%d):\n", len(tasks))
	fmt.Println(strings.Repeat("-", 40))
	for _, task := range tasks {
		scheduled := ""
		if task.ScheduledDate != nil {
			// Use UTC for date-only fields to preserve the stored date
			scheduled = task.ScheduledDate.UTC().Format("2006-01-02")
		}
		fmt.Printf("  #%d: %s [%s] %s\n", task.ID, truncate(task.Title, 30), task.Status, scheduled)
	}
}

func displayTemplate(tmpl *types.RecurringTaskTemplate) {
	fmt.Printf("Template #%d: %s\n", tmpl.ID, tmpl.Title)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	activeStr := "inactive"
	if tmpl.IsActive {
		activeStr = "active"
	}
	fmt.Printf("Status:       %s\n", activeStr)
	fmt.Printf("Type:         %s\n", tmpl.TemplateType)
	fmt.Printf("Project ID:   %d\n", tmpl.ProjectID)

	if tmpl.Priority != nil {
		fmt.Printf("Priority:     %s\n", *tmpl.Priority)
	}

	fmt.Println()
	fmt.Printf("Recurrence:   %s\n", tmpl.RecurrenceRule)
	fmt.Printf("              (%s)\n", rruleToHuman(tmpl.RecurrenceRule))
	fmt.Printf("Timezone:     %s\n", tmpl.Timezone)
	// Use UTC for date-only fields to preserve the stored date
	fmt.Printf("Start Date:   %s\n", tmpl.StartDate.UTC().Format("2006-01-02"))
	if tmpl.EndDate != nil {
		fmt.Printf("End Date:     %s\n", tmpl.EndDate.UTC().Format("2006-01-02"))
	}

	fmt.Println()
	fmt.Printf("Created:      %s\n", tmpl.CreatedAt.Local().Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:      %s\n", tmpl.UpdatedAt.Local().Format("2006-01-02 15:04:05"))

	if tmpl.Description != nil && *tmpl.Description != "" {
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println(*tmpl.Description)
	}

	if len(tmpl.Labels) > 0 {
		fmt.Println()
		fmt.Print("Labels: ")
		labelNames := make([]string, len(tmpl.Labels))
		for i, label := range tmpl.Labels {
			labelNames[i] = label.Name
		}
		fmt.Println(strings.Join(labelNames, ", "))
	}

	if len(tmpl.Assignees) > 0 {
		fmt.Println()
		fmt.Print("Assignees: ")
		assigneeNames := make([]string, len(tmpl.Assignees))
		for i, assignee := range tmpl.Assignees {
			assigneeNames[i] = assignee.Name
		}
		fmt.Println(strings.Join(assigneeNames, ", "))
	}

	// Show next due datetime
	occurrences := getNextOccurrences(tmpl, 1)
	if len(occurrences) > 0 {
		fmt.Println()
		// Load the template's timezone for display
		loc, err := time.LoadLocation(tmpl.Timezone)
		if err != nil {
			loc = time.UTC
		}
		nextDue := occurrences[0].In(loc)
		fmt.Printf("Next Due:     %s\n", nextDue.Format("Mon, Jan 2, 2006 15:04:05 MST"))
	}
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	// Validate required flags
	if templateCreateTitle == "" {
		return fmt.Errorf("--title is required")
	}
	if templateCreateRecurrence == "" {
		return fmt.Errorf("--recurrence is required")
	}
	if templateCreateStartDate == "" {
		return fmt.Errorf("--start-date is required")
	}

	// Validate RRULE format
	if err := validateRRule(templateCreateRecurrence); err != nil {
		return fmt.Errorf("invalid recurrence rule: %w", err)
	}

	// Validate start date format
	if _, err := time.Parse("2006-01-02", templateCreateStartDate); err != nil {
		return fmt.Errorf("invalid start date format (use YYYY-MM-DD): %w", err)
	}

	// Validate template type
	if templateCreateType != "task" && templateCreateType != "habit" {
		return fmt.Errorf("invalid --type value: must be 'task' or 'habit'")
	}

	// Validate timezone
	if err := validateTimezone(templateCreateTimezone); err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Resolve project ID from flag, config default, or error
	var projectID int
	if templateCreateProject != "" {
		projectID, err = resolveProjectID(ctx, apiClient, templateCreateProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}
	} else if cfg.Defaults.Project != "" {
		projectID, err = ensureDefaultProject(ctx, apiClient, cfg.Defaults.Project)
		if err != nil {
			return fmt.Errorf("failed to ensure default project: %w", err)
		}
	} else {
		return fmt.Errorf("--project is required (or configure defaults.project in config)")
	}

	// Build template create request
	templateCreate := &types.RecurringTaskTemplateCreate{
		ProjectID:      projectID,
		Title:          templateCreateTitle,
		RecurrenceRule: templateCreateRecurrence,
		StartDate:      templateCreateStartDate,
		Timezone:       templateCreateTimezone,
		TemplateType:   templateCreateType,
		IsActive:       true,
	}

	if templateCreateDescription != "" {
		templateCreate.Description = &templateCreateDescription
	}

	if templateCreatePriority != "" {
		templateCreate.Priority = &templateCreatePriority
	}

	if templateCreateEndDate != "" {
		// Validate end date format
		if _, err := time.Parse("2006-01-02", templateCreateEndDate); err != nil {
			return fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
		}
		templateCreate.EndDate = &templateCreateEndDate
	}

	if len(templateCreateLabels) > 0 {
		templateCreate.Labels = templateCreateLabels
	}

	if len(templateCreateAssignees) > 0 {
		templateCreate.Assignees = templateCreateAssignees
	}

	template, err := apiClient.CreateTemplate(ctx, templateCreate)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Println("Template created successfully:")
	displayTemplate(template)
	return nil
}

func runTemplateUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	templateID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	// Build update request
	templateUpdate := &types.RecurringTaskTemplateUpdate{}

	if templateUpdateTitle != "" {
		templateUpdate.Title = &templateUpdateTitle
	}

	if templateUpdateDescription != "" {
		templateUpdate.Description = &templateUpdateDescription
	}

	if templateUpdatePriority != "" {
		templateUpdate.Priority = &templateUpdatePriority
	}

	if templateUpdateRecurrence != "" {
		if err := validateRRule(templateUpdateRecurrence); err != nil {
			return fmt.Errorf("invalid recurrence rule: %w", err)
		}
		templateUpdate.RecurrenceRule = &templateUpdateRecurrence
	}

	if templateUpdateEndDate != "" {
		// Validate end date format
		if _, err := time.Parse("2006-01-02", templateUpdateEndDate); err != nil {
			return fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
		}
		templateUpdate.EndDate = &templateUpdateEndDate
	}

	if templateUpdateTimezone != "" {
		if err := validateTimezone(templateUpdateTimezone); err != nil {
			return fmt.Errorf("invalid timezone: %w", err)
		}
		templateUpdate.Timezone = &templateUpdateTimezone
	}

	if len(templateUpdateLabels) > 0 {
		templateUpdate.Labels = templateUpdateLabels
	}

	if len(templateUpdateAssignees) > 0 {
		templateUpdate.Assignees = templateUpdateAssignees
	}

	template, err := apiClient.UpdateTemplate(ctx, templateID, templateUpdate)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	fmt.Printf("Template #%d updated successfully\n", template.ID)
	return nil
}

func runTemplateActivate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	templateID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	active := true
	templateUpdate := &types.RecurringTaskTemplateUpdate{
		IsActive: &active,
	}

	template, err := apiClient.UpdateTemplate(ctx, templateID, templateUpdate)
	if err != nil {
		return fmt.Errorf("failed to activate template: %w", err)
	}

	fmt.Printf("Template #%d activated successfully\n", template.ID)
	return nil
}

func runTemplateDeactivate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	templateID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	active := false
	templateUpdate := &types.RecurringTaskTemplateUpdate{
		IsActive: &active,
	}

	template, err := apiClient.UpdateTemplate(ctx, templateID, templateUpdate)
	if err != nil {
		return fmt.Errorf("failed to deactivate template: %w", err)
	}

	fmt.Printf("Template #%d deactivated successfully\n", template.ID)
	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	templateID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %s", args[0])
	}

	// Confirm deletion unless --force
	if !templateDeleteForce {
		fmt.Printf("Are you sure you want to delete template #%d? (y/N): ", templateID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	apiClient := api.NewClient(cfg.APIURL)
	ctx := context.Background()

	err = apiClient.DeleteTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("Template #%d deleted successfully\n", templateID)
	return nil
}

// validateRRule performs basic validation of an RRULE string
func validateRRule(rrule string) error {
	// Basic RRULE validation - must start with FREQ=
	if !strings.HasPrefix(strings.ToUpper(rrule), "FREQ=") {
		return fmt.Errorf("RRULE must start with FREQ=")
	}

	// Check for valid frequency
	validFreqs := []string{"DAILY", "WEEKLY", "MONTHLY", "YEARLY"}
	hasValidFreq := false
	upperRRule := strings.ToUpper(rrule)
	for _, freq := range validFreqs {
		if strings.Contains(upperRRule, "FREQ="+freq) {
			hasValidFreq = true
			break
		}
	}
	if !hasValidFreq {
		return fmt.Errorf("FREQ must be one of: DAILY, WEEKLY, MONTHLY, YEARLY")
	}

	// Basic format check - should contain only valid RRULE parts
	validParts := regexp.MustCompile(`^(FREQ|INTERVAL|BYDAY|BYMONTH|BYMONTHDAY|BYYEARDAY|BYWEEKNO|UNTIL|COUNT|WKST)=`)
	parts := strings.Split(rrule, ";")
	for _, part := range parts {
		if !validParts.MatchString(strings.ToUpper(part)) {
			return fmt.Errorf("invalid RRULE part: %s", part)
		}
	}

	return nil
}

// validateTimezone validates an IANA timezone string
func validateTimezone(tz string) error {
	_, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("%q is not a valid IANA timezone", tz)
	}
	return nil
}

// rruleToHuman converts an RRULE to a human-readable string
func rruleToHuman(rrule string) string {
	parts := make(map[string]string)
	for _, part := range strings.Split(rrule, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			parts[strings.ToUpper(kv[0])] = kv[1]
		}
	}

	freq := parts["FREQ"]
	interval := parts["INTERVAL"]
	byday := parts["BYDAY"]
	bymonthday := parts["BYMONTHDAY"]

	// Parse interval
	intervalNum := 1
	if interval != "" {
		if n, err := strconv.Atoi(interval); err == nil {
			intervalNum = n
		}
	}

	var result string

	switch strings.ToUpper(freq) {
	case "DAILY":
		if intervalNum == 1 {
			result = "Daily"
		} else {
			result = fmt.Sprintf("Every %d days", intervalNum)
		}
		if byday != "" {
			result += fmt.Sprintf(" on %s", formatDays(byday))
		}

	case "WEEKLY":
		if intervalNum == 1 {
			result = "Weekly"
		} else {
			result = fmt.Sprintf("Every %d weeks", intervalNum)
		}
		if byday != "" {
			result += fmt.Sprintf(" on %s", formatDays(byday))
		}

	case "MONTHLY":
		if intervalNum == 1 {
			result = "Monthly"
		} else {
			result = fmt.Sprintf("Every %d months", intervalNum)
		}
		if bymonthday != "" {
			result += fmt.Sprintf(" on day %s", bymonthday)
		}

	case "YEARLY":
		if intervalNum == 1 {
			result = "Yearly"
		} else {
			result = fmt.Sprintf("Every %d years", intervalNum)
		}

	default:
		result = rrule
	}

	return result
}

// formatDays converts BYDAY codes to human-readable day names
func formatDays(byday string) string {
	dayMap := map[string]string{
		"MO": "Mon",
		"TU": "Tue",
		"WE": "Wed",
		"TH": "Thu",
		"FR": "Fri",
		"SA": "Sat",
		"SU": "Sun",
	}

	days := strings.Split(byday, ",")
	result := make([]string, 0, len(days))
	for _, d := range days {
		if name, ok := dayMap[strings.ToUpper(d)]; ok {
			result = append(result, name)
		} else {
			result = append(result, d)
		}
	}

	return strings.Join(result, ", ")
}

// getNextOccurrences calculates the next n occurrences for a template
func getNextOccurrences(tmpl *types.RecurringTaskTemplate, count int) []time.Time {
	// Load timezone
	loc, err := time.LoadLocation(tmpl.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Parse the RRULE
	ruleStr := "DTSTART:" + tmpl.StartDate.In(loc).Format("20060102T150405Z") + "\nRRULE:" + tmpl.RecurrenceRule
	rule, err := rrule.StrToRRule(ruleStr)
	if err != nil {
		return nil
	}

	// Get occurrences starting from now
	now := time.Now().In(loc)

	// If there's an end date, use it
	var endDate time.Time
	if tmpl.EndDate != nil {
		endDate = tmpl.EndDate.In(loc)
	} else {
		// Default to 1 year from now if no end date
		endDate = now.AddDate(1, 0, 0)
	}

	// Get occurrences between now and end date
	occurrences := rule.Between(now, endDate, true)

	// Limit to requested count
	if len(occurrences) > count {
		occurrences = occurrences[:count]
	}

	return occurrences
}
