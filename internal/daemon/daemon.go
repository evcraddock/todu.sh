package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/evcraddock/todu.sh/internal/config"
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

// Daemon manages background synchronization
type Daemon struct {
	engine   SyncEngine
	config   *config.Config
	stopChan chan struct{}
	doneChan chan struct{}
	status   Status
}

// New creates a new Daemon instance
func New(engine SyncEngine, config *config.Config) *Daemon {
	return &Daemon{
		engine:   engine,
		config:   config,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		status: Status{
			Running: false,
		},
	}
}

// Start begins the daemon main loop
func (d *Daemon) Start(ctx context.Context) error {
	d.status.Running = true
	d.status.PID = os.Getpid()
	d.writeStatus()

	log.Println("Daemon starting...")

	// Parse interval
	interval, err := time.ParseDuration(d.config.Daemon.Interval)
	if err != nil {
		return fmt.Errorf("invalid daemon interval: %w", err)
	}

	log.Printf("Sync interval: %s", interval)

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
			log.Println("Context cancelled, stopping daemon...")
			d.stop()
			return ctx.Err()

		case <-sigChan:
			log.Println("Received shutdown signal, stopping daemon...")
			d.stop()
			return nil

		case <-d.stopChan:
			log.Println("Stop signal received, shutting down...")
			d.stop()
			return nil

		case <-ticker.C:
			// Run sync
			if err := d.runSync(ctx); err != nil {
				failureCount++
				log.Printf("Sync failed (failure count: %d): %v", failureCount, err)

				// Calculate backoff duration using exponential backoff
				backoffMultiplier := math.Pow(2, float64(failureCount-1))
				backoffDuration := time.Duration(float64(baseInterval) * backoffMultiplier)
				if backoffDuration > maxBackoff {
					backoffDuration = maxBackoff
				}

				log.Printf("Next sync in %s (with backoff)", backoffDuration)

				// Reset ticker with backoff
				ticker.Reset(backoffDuration)
				d.status.NextSyncTime = time.Now().Add(backoffDuration)
			} else {
				// Success - reset failure count and backoff
				if failureCount > 0 {
					log.Printf("Sync succeeded, resetting backoff")
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
	log.Println("Stopping daemon...")
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
	log.Println("Daemon stopped")
}

// runSync executes a single sync operation
func (d *Daemon) runSync(ctx context.Context) error {
	log.Println("Starting sync...")
	d.status.LastSyncTime = time.Now()
	d.status.LastSyncError = ""

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

	// Log results
	log.Printf("Sync completed: %d created, %d updated, %d skipped, %d errors",
		result.TotalCreated, result.TotalUpdated, result.TotalSkipped, result.TotalErrors)

	if result.TotalErrors > 0 {
		d.status.LastSyncError = fmt.Sprintf("%d errors occurred during sync", result.TotalErrors)
		d.status.ErrorCount++
		d.writeStatus()
		return fmt.Errorf("sync completed with %d errors", result.TotalErrors)
	}

	// Reset error count on success
	d.status.ErrorCount = 0
	d.writeStatus()
	return nil
}

// writeStatus writes the current status to the status file
func (d *Daemon) writeStatus() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: failed to get home directory: %v", err)
		return
	}

	statusPath := filepath.Join(homeDir, ".todu", "daemon.status")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(statusPath), 0755); err != nil {
		log.Printf("Warning: failed to create status directory: %v", err)
		return
	}

	// Marshal status to JSON
	data, err := json.MarshalIndent(d.status, "", "  ")
	if err != nil {
		log.Printf("Warning: failed to marshal status: %v", err)
		return
	}

	// Write to file
	if err := os.WriteFile(statusPath, data, 0644); err != nil {
		log.Printf("Warning: failed to write status file: %v", err)
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

	statusPath := filepath.Join(homeDir, ".todu", "daemon.status")

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
