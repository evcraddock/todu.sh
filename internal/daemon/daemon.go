package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/internal/journal"
	"github.com/evcraddock/todu.sh/internal/sync"
)

// SyncEngine is an interface for sync operations (allows mocking in tests)
type SyncEngine interface {
	Sync(ctx context.Context, options sync.Options) (*sync.Result, error)
}

// Status represents the current daemon status
type Status struct {
	Running       bool      `json:"running"`
	PID           int       `json:"pid,omitempty"`
	LastSyncTime  time.Time `json:"last_sync_time,omitempty"`
	LastSyncError string    `json:"last_sync_error,omitempty"`
	ErrorCount    int       `json:"error_count"`
	NextSyncTime  time.Time `json:"next_sync_time,omitempty"`
}

// APIClient defines the interface for API operations needed by the daemon
type APIClient interface {
	ProcessDueTemplates(ctx context.Context) (*api.ProcessDueTemplatesResponse, error)
}

// Daemon manages background synchronization
type Daemon struct {
	engine        SyncEngine
	apiClient     APIClient
	fullAPIClient *api.Client
	config        *config.Config
	logger        zerolog.Logger
	stopChan      chan struct{}
	doneChan      chan struct{}
	status        Status
}

// New creates a new Daemon instance
func New(engine SyncEngine, apiClient APIClient, config *config.Config) *Daemon {
	// Setup logger with rotation
	logger, err := setupLogger(config)
	if err != nil {
		// Fallback to a basic logger if setup fails
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
		logger.Error().Err(err).Msg("Failed to setup logger, using stderr")
	}

	// If the engine supports WithLogger, set the logger on it
	if e, ok := engine.(interface {
		WithLogger(zerolog.Logger) *sync.Engine
	}); ok {
		e.WithLogger(logger)
	}

	d := &Daemon{
		engine:    engine,
		apiClient: apiClient,
		config:    config,
		logger:    logger,
		stopChan:  make(chan struct{}),
		doneChan:  make(chan struct{}),
		status: Status{
			Running: false,
		},
	}

	// If the apiClient is a full API client, store it for journal export
	if fullClient, ok := apiClient.(*api.Client); ok {
		d.fullAPIClient = fullClient
	}

	return d
}

// Start begins the daemon main loop
func (d *Daemon) Start(ctx context.Context) error {
	d.status.Running = true
	d.status.PID = os.Getpid()
	d.writeStatus()

	d.logger.Info().Msg("Daemon starting...")

	// Parse interval
	interval, err := time.ParseDuration(d.config.Daemon.Interval)
	if err != nil {
		return fmt.Errorf("invalid daemon interval: %w", err)
	}

	d.logger.Info().Dur("interval", interval).Msg("Sync interval configured")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Variables for exponential backoff
	baseInterval := interval
	failureCount := 0
	maxBackoff := 5 * time.Minute

	// Run first sync immediately
	d.runSync(ctx)

	// Main daemon loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info().Msg("Context cancelled, stopping daemon...")
			d.stop()
			return ctx.Err()

		case <-sigChan:
			d.logger.Info().Msg("Received shutdown signal, stopping daemon...")
			d.stop()
			return nil

		case <-d.stopChan:
			d.logger.Info().Msg("Stop signal received, shutting down...")
			d.stop()
			return nil

		case <-ticker.C:
			// Run sync
			if err := d.runSync(ctx); err != nil {
				failureCount++
				d.logger.Error().Err(err).Int("failure_count", failureCount).Msg("Sync failed")

				// Calculate backoff duration using exponential backoff
				backoffMultiplier := math.Pow(2, float64(failureCount-1))
				backoffDuration := time.Duration(float64(baseInterval) * backoffMultiplier)
				if backoffDuration > maxBackoff {
					backoffDuration = maxBackoff
				}

				d.logger.Info().Dur("next_sync_in", backoffDuration).Msg("Next sync scheduled with backoff")

				// Reset ticker with backoff
				ticker.Reset(backoffDuration)
				d.status.NextSyncTime = time.Now().Add(backoffDuration)
			} else {
				// Success - reset failure count and backoff
				if failureCount > 0 {
					d.logger.Info().Msg("Sync succeeded, resetting backoff")
					failureCount = 0
					ticker.Reset(baseInterval)
				}
				d.status.NextSyncTime = time.Now().Add(baseInterval)
			}

			d.writeStatus()
		}
	}
}

// Stop gracefully stops the daemon
func (d *Daemon) Stop() error {
	d.logger.Info().Msg("Stopping daemon...")
	close(d.stopChan)
	<-d.doneChan
	return nil
}

// stop is the internal stop implementation
func (d *Daemon) stop() {
	d.status.Running = false
	d.status.PID = 0
	d.writeStatus()
	close(d.doneChan)
	d.logger.Info().Msg("Daemon stopped")
}

// runSync executes a single sync operation
func (d *Daemon) runSync(ctx context.Context) error {
	d.logger.Info().Msg("Starting sync...")
	previousSyncTime := d.status.LastSyncTime
	d.status.LastSyncTime = time.Now()
	d.status.LastSyncError = ""

	// Check if this is a new day - if so, export yesterday's journal
	if d.isNewDay(previousSyncTime) {
		d.exportYesterdayJournal(ctx)
	}

	// Build sync options
	options := sync.Options{
		ProjectIDs: d.config.Daemon.Projects,
	}

	// Run sync
	result, err := d.engine.Sync(ctx, options)
	if err != nil {
		d.status.LastSyncError = err.Error()
		d.status.ErrorCount++
		d.writeStatus()
		return err
	}

	// Log results summary
	d.logger.Info().
		Int("created", result.TotalCreated).
		Int("updated", result.TotalUpdated).
		Int("skipped", result.TotalSkipped).
		Int("errors", result.TotalErrors).
		Msg("Sync completed")

	if result.TotalErrors > 0 {
		// Log detailed errors for each project (always log errors regardless of level)
		for _, pr := range result.ProjectResults {
			if len(pr.Errors) > 0 {
				for _, err := range pr.Errors {
					d.logger.Error().
						Str("project", pr.ProjectName).
						Err(err).
						Msg("Project sync error")
				}
			}
		}

		d.status.LastSyncError = fmt.Sprintf("%d errors occurred during sync", result.TotalErrors)
		d.status.ErrorCount++
		d.writeStatus()
		return fmt.Errorf("sync completed with %d errors", result.TotalErrors)
	}

	// Reset error count on success
	d.status.ErrorCount = 0

	// Process recurring task templates if enabled
	if d.config.RecurringTasks.Enabled {
		if err := d.processRecurringTasks(ctx); err != nil {
			d.logger.Warn().Err(err).Msg("Failed to process recurring templates")
			// Don't fail the entire sync - just log the error
		}
	}

	d.writeStatus()
	return nil
}

// processRecurringTasks processes due recurring task templates
func (d *Daemon) processRecurringTasks(ctx context.Context) error {
	d.logger.Debug().Msg("Processing recurring task templates...")

	result, err := d.apiClient.ProcessDueTemplates(ctx)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	// Log summary
	if result.TasksCreated > 0 || result.Skipped > 0 || result.Failed > 0 {
		d.logger.Info().
			Int("created", result.TasksCreated).
			Int("skipped", result.Skipped).
			Int("failed", result.Failed).
			Int("processed", result.Processed).
			Msg("Recurring tasks processed")

		// Log details at DEBUG level
		for _, detail := range result.Details {
			if detail.Action == "created" && detail.TaskID != nil {
				d.logger.Debug().
					Int("task_id", *detail.TaskID).
					Int("template_id", detail.TemplateID).
					Msg("Created task from template")
			} else if detail.Action == "skipped" {
				d.logger.Debug().
					Int("template_id", detail.TemplateID).
					Str("reason", detail.Reason).
					Msg("Skipped template")
			} else if detail.Action == "failed" {
				d.logger.Error().
					Int("template_id", detail.TemplateID).
					Str("error", detail.Error).
					Msg("Failed to process template")
			}
		}
	} else {
		d.logger.Debug().Msg("No recurring tasks due")
	}

	return nil
}

// writeStatus writes the current status to the status file
func (d *Daemon) writeStatus() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		d.logger.Warn().Err(err).Msg("Failed to get home directory for status file")
		return
	}

	statusPath := filepath.Join(homeDir, ".config", "todu", "daemon.status")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(statusPath), 0755); err != nil {
		d.logger.Warn().Err(err).Msg("Failed to create status directory")
		return
	}

	// Marshal status to JSON
	data, err := json.MarshalIndent(d.status, "", "  ")
	if err != nil {
		d.logger.Warn().Err(err).Msg("Failed to marshal status")
		return
	}

	// Write to file
	if err := os.WriteFile(statusPath, data, 0644); err != nil {
		d.logger.Warn().Err(err).Msg("Failed to write status file")
	}
}

// GetStatus returns the current daemon status
func (d *Daemon) GetStatus() Status {
	return d.status
}

// ReadStatus reads the daemon status from the status file
func ReadStatus() (*Status, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	statusPath := filepath.Join(homeDir, ".config", "todu", "daemon.status")

	data, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Status{Running: false}, nil
		}
		return nil, fmt.Errorf("failed to read status file: %w", err)
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status file: %w", err)
	}

	// Verify the process is actually running if status says it is
	if status.Running && status.PID > 0 {
		// Check if process exists by sending signal 0 (doesn't actually send a signal)
		process, err := os.FindProcess(status.PID)
		if err != nil {
			// Process doesn't exist
			status.Running = false
			status.PID = 0
		} else {
			// On Unix systems, FindProcess always succeeds, so we need to signal it
			err := process.Signal(syscall.Signal(0))
			if err != nil {
				// Process doesn't exist or we don't have permission
				status.Running = false
				status.PID = 0
			}
		}
	}

	return &status, nil
}

// isNewDay checks if the current sync is on a different day than the previous sync
func (d *Daemon) isNewDay(previousSyncTime time.Time) bool {
	// If this is the first sync ever, consider it a new day
	if previousSyncTime.IsZero() {
		return true
	}

	now := time.Now()
	prevYear, prevMonth, prevDay := previousSyncTime.Local().Date()
	nowYear, nowMonth, nowDay := now.Local().Date()

	return prevYear != nowYear || prevMonth != nowMonth || prevDay != nowDay
}

// exportYesterdayJournal exports the previous day's journal to a markdown file
func (d *Daemon) exportYesterdayJournal(ctx context.Context) {
	if d.fullAPIClient == nil {
		d.logger.Debug().Msg("Skipping journal export: full API client not available")
		return
	}

	if d.config.LocalReports == "" {
		d.logger.Debug().Msg("Skipping journal export: local_reports path not configured")
		return
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	d.logger.Info().Time("date", yesterday).Msg("Exporting previous day's journal")

	outputPath, err := journal.Export(ctx, d.fullAPIClient, yesterday, d.config.LocalReports)
	if err != nil {
		d.logger.Warn().Err(err).Msg("Failed to export journal")
		return
	}

	d.logger.Info().Str("path", outputPath).Msg("Journal exported successfully")
}
