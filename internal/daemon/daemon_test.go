package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/evcraddock/todu.sh/internal/sync"
)

// mockEngine is a mock sync engine for testing
type mockEngine struct {
	syncCount  int
	shouldFail bool
	failCount  int
	syncDelay  time.Duration
}

func (m *mockEngine) Sync(ctx context.Context, options sync.Options) (*sync.Result, error) {
	m.syncCount++

	// Simulate sync delay
	if m.syncDelay > 0 {
		time.Sleep(m.syncDelay)
	}

	// Fail first N times if configured
	if m.shouldFail && m.syncCount <= m.failCount {
		return nil, context.DeadlineExceeded
	}

	return &sync.Result{
		ProjectResults: []sync.ProjectResult{},
		Duration:       100 * time.Millisecond,
	}, nil
}

func TestNew(t *testing.T) {
	engine := &mockEngine{}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "1s",
		},
	}

	daemon := New(engine, cfg)

	if daemon == nil {
		t.Fatal("New returned nil")
	}

	if daemon.config != cfg {
		t.Error("Config not set correctly")
	}

	if daemon.status.Running {
		t.Error("Daemon should not be running initially")
	}
}

func TestDaemonStartStop(t *testing.T) {
	engine := &mockEngine{}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "100ms",
		},
	}

	daemon := New(engine, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start daemon in goroutine
	errChan := make(chan error)
	go func() {
		errChan <- daemon.Start(ctx)
	}()

	// Wait for first sync
	time.Sleep(150 * time.Millisecond)

	// Stop daemon
	err := daemon.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Wait for Start to return
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Start did not return after Stop")
	}

	if engine.syncCount < 1 {
		t.Errorf("Expected at least 1 sync, got %d", engine.syncCount)
	}
}

func TestDaemonPeriodicSync(t *testing.T) {
	engine := &mockEngine{}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "100ms",
		},
	}

	daemon := New(engine, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	// Start daemon
	err := daemon.Start(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected context.DeadlineExceeded, got: %v", err)
	}

	// Should have run sync at least 3 times (0ms, 100ms, 200ms, 300ms)
	if engine.syncCount < 3 {
		t.Errorf("Expected at least 3 syncs, got %d", engine.syncCount)
	}
}

func TestDaemonExponentialBackoff(t *testing.T) {
	engine := &mockEngine{
		shouldFail: true,
		failCount:  10, // Fail all attempts during the test
	}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "100ms",
		},
	}

	daemon := New(engine, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 450*time.Millisecond)
	defer cancel()

	start := time.Now()
	daemon.Start(ctx)
	duration := time.Since(start)

	// First sync happens immediately (fails)
	// Second sync waits 100ms (base interval) (fails)
	// Third sync should wait 200ms (100ms * 2^1) due to backoff (fails)
	// Total should be at least 300ms
	if duration < 300*time.Millisecond {
		t.Errorf("Expected backoff to take at least 300ms, took %v", duration)
	}

	// Should have attempted sync at least 3 times
	if engine.syncCount < 3 {
		t.Errorf("Expected at least 3 sync attempts, got %d", engine.syncCount)
	}

	// Error count should reflect failures since all syncs failed
	if daemon.status.ErrorCount < 1 {
		t.Errorf("Expected error count >= 1 after failures, got %d", daemon.status.ErrorCount)
	}
}

func TestDaemonStatusFile(t *testing.T) {
	// Create temp home directory for test
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	engine := &mockEngine{}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "100ms",
		},
	}

	daemon := New(engine, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// Start daemon
	daemon.Start(ctx)

	// Check status file was created
	statusPath := filepath.Join(tempHome, ".todu", "daemon.status")
	if _, err := os.Stat(statusPath); os.IsNotExist(err) {
		t.Error("Status file was not created")
	}

	// Read status file
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("Failed to read status file: %v", err)
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		t.Fatalf("Failed to parse status file: %v", err)
	}

	// Verify status
	if status.LastSyncTime.IsZero() {
		t.Error("LastSyncTime should be set")
	}
}

func TestReadStatus(t *testing.T) {
	// Create temp home directory for test
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test reading non-existent status file
	status, err := ReadStatus()
	if err != nil {
		t.Fatalf("ReadStatus failed: %v", err)
	}
	if status.Running {
		t.Error("Status should show not running when file doesn't exist")
	}

	// Create status file
	statusPath := filepath.Join(tempHome, ".todu", "daemon.status")
	os.MkdirAll(filepath.Dir(statusPath), 0755)

	testStatus := Status{
		Running:      true,
		LastSyncTime: time.Now(),
		ErrorCount:   5,
	}

	data, _ := json.Marshal(testStatus)
	os.WriteFile(statusPath, data, 0644)

	// Read status file
	status, err = ReadStatus()
	if err != nil {
		t.Fatalf("ReadStatus failed: %v", err)
	}

	if !status.Running {
		t.Error("Status should show running")
	}

	if status.ErrorCount != 5 {
		t.Errorf("Expected ErrorCount 5, got %d", status.ErrorCount)
	}
}

func TestDaemonInvalidInterval(t *testing.T) {
	engine := &mockEngine{}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "invalid",
		},
	}

	daemon := New(engine, cfg)
	ctx := context.Background()

	err := daemon.Start(ctx)
	if err == nil {
		t.Error("Expected error for invalid interval")
	}
}

func TestDaemonContextCancellation(t *testing.T) {
	engine := &mockEngine{
		syncDelay: 500 * time.Millisecond, // Long sync to test cancellation
	}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "1s",
		},
	}

	daemon := New(engine, cfg)
	ctx, cancel := context.WithCancel(context.Background())

	// Start daemon in goroutine
	errChan := make(chan error)
	go func() {
		errChan <- daemon.Start(ctx)
	}()

	// Cancel context after first sync starts
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for Start to return
	select {
	case err := <-errChan:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestDaemonProjectFiltering(t *testing.T) {
	engine := &mockEngine{}
	cfg := &config.Config{
		Daemon: config.DaemonConfig{
			Interval: "100ms",
			Projects: []int{1, 2, 3},
		},
	}

	daemon := New(engine, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// Start daemon
	daemon.Start(ctx)

	// Verify sync was called at least once
	if engine.syncCount < 1 {
		t.Error("Expected at least one sync call")
	}
}
