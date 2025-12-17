package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/review"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Generate review reports",
	Long: `Generate review reports for tasks and habits.

Available reports:
  daily   Generate a daily review with tasks organized by status
  weekly  Generate a weekly review with tasks organized by priority`,
}

var reviewDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Generate daily review report",
	Long: `Generate a daily review report with tasks and habits.

The report includes:
  - In Progress: Tasks currently being worked on
  - Daily Goals: Habit completion status for the day
  - Coming up Soon: Tasks due within the next 3 days
  - Next: High priority, scheduled, and default project tasks
  - Waiting: Tasks in waiting status
  - Done Today: Tasks completed today

Example:
  todu review daily                        # Display review to stdout
  todu review daily --save                 # Save to default location
  todu review daily --save=./review.md     # Save to specific path (use = for path)
  todu review daily --date 2025-12-15      # Generate review for specific date`,
	RunE: runReviewDaily,
}

var reviewWeeklyCmd = &cobra.Command{
	Use:   "weekly",
	Short: "Generate weekly review report",
	Long: `Generate a weekly review report with tasks organized by priority.

The report includes:
  - Waiting: Tasks in waiting status
  - Next: High priority tasks and overdue tasks
  - Active: Medium priority tasks and tasks with no priority
  - Someday: Low priority tasks

Example:
  todu review weekly                        # Display review to stdout
  todu review weekly --save                 # Save to default location
  todu review weekly --save=./review.md     # Save to specific path (use = for path)`,
	RunE: runReviewWeekly,
}

var (
	reviewDailyDate  string
	reviewDailySave  string
	reviewWeeklySave string
)

func init() {
	rootCmd.AddCommand(reviewCmd)
	reviewCmd.AddCommand(reviewDailyCmd)
	reviewCmd.AddCommand(reviewWeeklyCmd)

	reviewDailyCmd.Flags().StringVar(&reviewDailyDate, "date", "", "Target date (YYYY-MM-DD, defaults to today)")
	reviewDailyCmd.Flags().StringVar(&reviewDailySave, "save", "", "Save to file (optional path, defaults to {local_reports}/daily-review.md)")
	reviewDailyCmd.Flags().Lookup("save").NoOptDefVal = "default"

	reviewWeeklyCmd.Flags().StringVar(&reviewWeeklySave, "save", "", "Save to file (optional path, defaults to {local_reports}/weekly-review.md)")
	reviewWeeklyCmd.Flags().Lookup("save").NoOptDefVal = "default"
}

func runReviewDaily(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	// Parse target date
	targetDate := time.Now()
	if reviewDailyDate != "" {
		parsed, err := time.ParseInLocation("2006-01-02", reviewDailyDate, time.Local)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", reviewDailyDate)
		}
		targetDate = parsed
	}

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	// Generate the report
	markdown, err := review.DailyReport(ctx, apiClient, targetDate, cfg.Defaults.Project)
	if err != nil {
		return fmt.Errorf("failed to generate daily review: %w", err)
	}

	// If --save flag is provided, save to file
	if reviewDailySave != "" {
		var outputPath string
		if reviewDailySave == "default" {
			// Use default location
			if cfg.LocalReports == "" {
				return fmt.Errorf("local_reports path not configured. Set it in your config file or specify a path: --save ./review.md")
			}
			outputPath = review.DefaultDailyReportPath(cfg.LocalReports)
		} else {
			// Use provided path
			outputPath = reviewDailySave
		}

		if err := review.SaveDailyReport(markdown, outputPath); err != nil {
			return fmt.Errorf("failed to save daily review: %w", err)
		}
		fmt.Printf("Daily review saved to: %s\n", outputPath)
		return nil
	}

	// Default: print to stdout
	fmt.Print(markdown)
	return nil
}

func runReviewWeekly(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()

	// Generate the report
	markdown, err := review.WeeklyReport(ctx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to generate weekly review: %w", err)
	}

	// If --save flag is provided, save to file
	if reviewWeeklySave != "" {
		var outputPath string
		if reviewWeeklySave == "default" {
			// Use default location
			if cfg.LocalReports == "" {
				return fmt.Errorf("local_reports path not configured. Set it in your config file or specify a path: --save ./review.md")
			}
			outputPath = review.DefaultWeeklyReportPath(cfg.LocalReports)
		} else {
			// Use provided path
			outputPath = reviewWeeklySave
		}

		if err := review.SaveReport(markdown, outputPath); err != nil {
			return fmt.Errorf("failed to save weekly review: %w", err)
		}
		fmt.Printf("Weekly review saved to: %s\n", outputPath)
		return nil
	}

	// Default: print to stdout
	fmt.Print(markdown)
	return nil
}
