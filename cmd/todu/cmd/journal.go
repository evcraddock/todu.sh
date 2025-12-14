package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/internal/journal"
	"github.com/evcraddock/todu.sh/pkg/types"
	"github.com/spf13/cobra"
)

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Manage personal journal entries",
	Long: `Manage personal journal entries in todu.

Journal entries are like comments but without being attached to a specific task.
Use them for daily logs, notes, reflections, or any personal note-taking.`,
}

var journalAddCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Add a journal entry",
	Long: `Add a journal entry.

If text is provided, creates the entry directly.
If no text is provided, opens your default editor ($VISUAL or $EDITOR).`,
	RunE: runJournalAdd,
}

var journalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List journal entries and comments",
	Long: `List journal entries and comments with optional filtering.

Use --type to filter by entry type:
  journal: Only journal entries (default)
  comment: Only task comments
  all: Both journal entries and task comments`,
	RunE: runJournalList,
}

var journalShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show journal entry details",
	Long:  `Display detailed information about a specific journal entry.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runJournalShow,
}

var journalEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a journal entry",
	Long:  `Edit an existing journal entry in your default editor.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runJournalEdit,
}

var journalDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a journal entry",
	Long:  `Delete a journal entry. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runJournalDelete,
}

var journalSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search journal entries",
	Long:  `Search journal entries by content.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runJournalSearch,
}

var journalExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export daily journal to markdown file",
	Long: `Export journal entries, completed tasks, and habits for a day to a markdown file.

The file is saved to {local_reports}/YYYY/MM-Monthname/MM-DD-YYYY-journal.md

Example:
  todu journal export              # Export today's journal
  todu journal export --date 2025-12-11  # Export specific date`,
	RunE: runJournalExport,
}

var (
	// Add flags
	journalAddAuthor string

	// List flags
	journalListToday bool
	journalListLast  int
	journalListSince string
	journalListLimit int
	journalListType  string

	// Delete flags
	journalDeleteForce bool

	// Export flags
	journalExportDate string
)

func init() {
	rootCmd.AddCommand(journalCmd)
	journalCmd.AddCommand(journalAddCmd)
	journalCmd.AddCommand(journalListCmd)
	journalCmd.AddCommand(journalShowCmd)
	journalCmd.AddCommand(journalEditCmd)
	journalCmd.AddCommand(journalDeleteCmd)
	journalCmd.AddCommand(journalSearchCmd)
	journalCmd.AddCommand(journalExportCmd)

	// Add flags
	journalAddCmd.Flags().StringVar(&journalAddAuthor, "author", "", "Entry author (defaults to config/git user)")

	// List flags
	journalListCmd.Flags().BoolVar(&journalListToday, "today", false, "Show only today's entries")
	journalListCmd.Flags().IntVar(&journalListLast, "last", 0, "Show last N days of entries")
	journalListCmd.Flags().StringVar(&journalListSince, "since", "", "Show entries since date (YYYY-MM-DD)")
	journalListCmd.Flags().IntVar(&journalListLimit, "limit", 50, "Maximum number of entries to show")
	journalListCmd.Flags().StringVar(&journalListType, "type", "journal", "Filter by type: 'journal' (journal entries), 'comment' (task comments), or 'all'")

	// Delete flags
	journalDeleteCmd.Flags().BoolVarP(&journalDeleteForce, "force", "f", false, "Skip confirmation")

	// Export flags
	journalExportCmd.Flags().StringVar(&journalExportDate, "date", "", "Date to export (YYYY-MM-DD, defaults to today)")
}

func runJournalAdd(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	var content string

	// If text provided as argument, use it; otherwise open editor
	if len(args) > 0 {
		content = strings.Join(args, " ")
	} else {
		// Open editor with blank content
		editedContent, err := openEditor("")
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
		content = editedContent
	}

	// Check if content is empty
	if content == "" {
		fmt.Println("Empty entry. Cancelled.")
		return nil
	}

	// Get author (from flag, config, git, or default)
	author := getAuthor(journalAddAuthor, cfg)

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	// Create journal entry (TaskID is nil)
	entryCreate := &types.CommentCreate{
		TaskID:  nil, // nil for journal entries
		Content: content,
		Author:  author,
	}

	entry, err := apiClient.CreateComment(ctx, entryCreate)
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	fmt.Printf("Journal entry created (ID: %d)\n", entry.ID)
	fmt.Printf("[%s] %s:\n", entry.CreatedAt.Local().Format("2006-01-02 15:04"), entry.Author)
	fmt.Println(entry.Content)
	return nil
}

func runJournalList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	// Validate type parameter
	if journalListType != "all" && journalListType != "journal" && journalListType != "comment" {
		return fmt.Errorf("invalid type: %s (must be 'all', 'journal', or 'comment')", journalListType)
	}

	// Fetch entries based on type
	var entries []*types.Comment
	if journalListType == "journal" {
		entries, err = apiClient.ListJournals(ctx, 0, journalListLimit)
	} else {
		entries, err = apiClient.ListAllComments(ctx, journalListType, 0, journalListLimit)
	}
	if err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}

	// Apply date filters
	entries = filterJournalsByDate(entries)

	if len(entries) == 0 {
		fmt.Println("No entries found")
		return nil
	}

	// Display results
	if GetOutputFormat() == "json" {
		return displayJournalsJSON(entries)
	}

	return displayJournalsTable(entries)
}

func filterJournalsByDate(entries []*types.Comment) []*types.Comment {
	var filtered []*types.Comment
	now := time.Now()

	for _, entry := range entries {
		// Today filter
		if journalListToday {
			if !isSameDay(entry.CreatedAt, now) {
				continue
			}
		}

		// Last N days filter
		if journalListLast > 0 {
			cutoff := now.AddDate(0, 0, -journalListLast)
			if entry.CreatedAt.Before(cutoff) {
				continue
			}
		}

		// Since date filter (user input is local timezone, entry.CreatedAt is UTC)
		if journalListSince != "" {
			sinceDate, err := time.ParseInLocation("2006-01-02", journalListSince, time.Local)
			if err == nil && entry.CreatedAt.Before(sinceDate) {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func displayJournalsTable(entries []*types.Comment) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDATE\tAUTHOR\tTASK\tCONTENT")
	fmt.Fprintln(w, "--\t----\t------\t----\t-------")

	for _, entry := range entries {
		content := truncate(entry.Content, 50)
		// Replace newlines with spaces for table display
		content = strings.ReplaceAll(content, "\n", " ")

		// Determine task display
		taskDisplay := "journal"
		if entry.TaskID != nil {
			taskDisplay = fmt.Sprintf("#%d", *entry.TaskID)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			entry.ID,
			entry.CreatedAt.Local().Format("2006-01-02 15:04"),
			entry.Author,
			taskDisplay,
			content,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d entries\n", len(entries))
	return nil
}

func displayJournalsJSON(entries []*types.Comment) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func runJournalShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	entryID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid entry ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	entry, err := apiClient.GetComment(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get journal entry: %w", err)
	}

	// Verify it's a journal entry (not a task comment)
	if entry.TaskID != nil {
		return fmt.Errorf("entry %d is a task comment, not a journal entry", entryID)
	}

	// Display results
	if GetOutputFormat() == "json" {
		data, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	displayJournalEntry(entry)
	return nil
}

func displayJournalEntry(entry *types.Comment) {
	fmt.Printf("Journal Entry #%d\n", entry.ID)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("Author:  %s\n", entry.Author)
	fmt.Printf("Created: %s\n", entry.CreatedAt.Local().Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", entry.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println(entry.Content)
}

func runJournalEdit(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	entryID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid entry ID: %s", args[0])
	}

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	// Get current entry
	entry, err := apiClient.GetComment(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get journal entry: %w", err)
	}

	// Verify it's a journal entry
	if entry.TaskID != nil {
		return fmt.Errorf("entry %d is a task comment, not a journal entry", entryID)
	}

	// Open editor with current content
	editedContent, err := openEditor(entry.Content)
	if err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Check if content changed
	if editedContent == entry.Content {
		fmt.Println("No changes made.")
		return nil
	}

	// Check if empty
	if editedContent == "" {
		fmt.Println("Empty entry. Cancelled.")
		return nil
	}

	// Update entry
	update := &types.CommentUpdate{
		Content: &editedContent,
	}

	updated, err := apiClient.UpdateComment(ctx, entryID, update)
	if err != nil {
		return fmt.Errorf("failed to update journal entry: %w", err)
	}

	fmt.Printf("Journal entry #%d updated successfully\n", updated.ID)
	return nil
}

func runJournalDelete(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	entryID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid entry ID: %s", args[0])
	}

	// Confirm deletion unless --force
	if !journalDeleteForce {
		fmt.Printf("Are you sure you want to delete journal entry #%d? (y/N): ", entryID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	err = apiClient.DeleteComment(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to delete journal entry: %w", err)
	}

	fmt.Printf("Journal entry #%d deleted successfully\n", entryID)
	return nil
}

func runJournalSearch(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	query := strings.ToLower(args[0])

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	// Fetch all journal entries (API max limit is 100)
	entries, err := apiClient.ListJournals(ctx, 0, 100)
	if err != nil {
		return fmt.Errorf("failed to list journal entries: %w", err)
	}

	// Filter by search query
	var matches []*types.Comment
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Content), query) ||
			strings.Contains(strings.ToLower(entry.Author), query) {
			matches = append(matches, entry)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No journal entries found matching '%s'\n", query)
		return nil
	}

	// Display results
	if GetOutputFormat() == "json" {
		return displayJournalsJSON(matches)
	}

	return displayJournalsTable(matches)
}

// getAuthor returns the author name from flag, config, git, or default
func getAuthor(flagValue string, cfg *config.Config) string {
	// 1. From flag
	if flagValue != "" {
		return flagValue
	}

	// 2. From config
	if cfg != nil && cfg.Author != "" {
		return cfg.Author
	}

	// 3. From git config
	if gitAuthor := getGitUser(); gitAuthor != "" {
		return gitAuthor
	}

	// 4. Default
	return "user"
}

// getGitUser returns the git user.name if available
func getGitUser() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func runJournalExport(cmd *cobra.Command, args []string) error {
	// 1. Load and validate config
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	if cfg.LocalReports == "" {
		return fmt.Errorf("local_reports path not configured")
	}

	// 2. Parse target date
	targetDate := time.Now()
	if journalExportDate != "" {
		parsed, err := time.ParseInLocation("2006-01-02", journalExportDate, time.Local)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", journalExportDate)
		}
		targetDate = parsed
	}

	// 3. Export using the shared journal package
	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	outputPath, err := journal.Export(ctx, apiClient, targetDate, cfg.LocalReports)
	if err != nil {
		return err
	}

	fmt.Printf("Journal exported to: %s\n", outputPath)
	return nil
}
