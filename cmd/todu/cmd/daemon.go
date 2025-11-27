package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/daemon"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/internal/sync"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the todu sync daemon",
	Long: `Manage the background daemon that periodically syncs tasks.

The daemon can run in foreground mode or be installed as a system service
that starts automatically at login/boot.`,
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start daemon in foreground",
	Long: `Start the sync daemon in foreground mode.

The daemon will run until interrupted (Ctrl+C) and log to stdout.
For background operation, use 'daemon install' to install as a service.`,
	RunE: runDaemonStart,
}

var daemonInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install daemon as system service",
	Long: `Install the sync daemon as a system service.

The service will start automatically at login/boot and run in the background.
Logs are written to the system log.`,
	RunE: runDaemonInstall,
}

var daemonUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall daemon service",
	Long: `Remove the daemon service from the system.

This will stop the daemon if running and remove the service configuration.`,
	RunE: runDaemonUninstall,
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	Long: `Display the current status of the daemon.

Shows whether the daemon is running, when it last synced, and when
the next sync is scheduled.`,
	RunE: runDaemonStatus,
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop running daemon",
	Long: `Stop the running daemon service.

This sends a graceful shutdown signal and waits for the daemon to finish
its current sync operation.`,
	RunE: runDaemonStop,
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart daemon service",
	Long: `Restart the daemon service.

This stops and then starts the daemon, applying any configuration changes.`,
	RunE: runDaemonRestart,
}

var daemonLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show daemon logs",
	Long: `Display recent daemon logs.

Use --follow to tail logs in real-time.`,
	RunE: runDaemonLogs,
}

var (
	daemonInstallInterval string
	daemonInstallProjects []int
	daemonLogsFollow      bool
	daemonLogsLines       int
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonInstallCmd)
	daemonCmd.AddCommand(daemonUninstallCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonRestartCmd)
	daemonCmd.AddCommand(daemonLogsCmd)

	// Install flags
	daemonInstallCmd.Flags().StringVar(&daemonInstallInterval, "interval", "", "Sync interval (e.g., 5m, 1h)")
	daemonInstallCmd.Flags().IntSliceVar(&daemonInstallProjects, "projects", []int{}, "Project IDs to sync (empty = all)")

	// Logs flags
	daemonLogsCmd.Flags().BoolVarP(&daemonLogsFollow, "follow", "f", false, "Follow log output")
	daemonLogsCmd.Flags().IntVarP(&daemonLogsLines, "lines", "n", 50, "Number of lines to show")
}

func runDaemonStart(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("API URL not configured")
	}

	// Create API client
	apiClient := api.NewClient(cfg.APIURL)

	// Use the global plugin registry (with plugins already registered)
	pluginRegistry := registry.Default

	// Create sync engine
	syncEngine := sync.NewEngine(apiClient, pluginRegistry)

	// Create daemon
	d := daemon.New(syncEngine, cfg)

	// Show startup message
	fmt.Printf("Daemon started, syncing every %s\n", cfg.Daemon.Interval)
	if len(cfg.Daemon.Projects) > 0 {
		fmt.Printf("Syncing projects: %v\n", cfg.Daemon.Projects)
	} else {
		fmt.Println("Syncing all projects")
	}
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Start daemon (blocks until stopped)
	ctx := context.Background()
	if err := d.Start(ctx); err != nil {
		return fmt.Errorf("daemon failed: %w", err)
	}

	return nil
}

func runDaemonInstall(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with command-line flags if provided
	if daemonInstallInterval != "" {
		// Validate interval format
		if _, err := time.ParseDuration(daemonInstallInterval); err != nil {
			return fmt.Errorf("invalid interval format: %w", err)
		}
		cfg.Daemon.Interval = daemonInstallInterval
	}

	if len(daemonInstallProjects) > 0 {
		cfg.Daemon.Projects = daemonInstallProjects
	}

	// Get service manager for platform
	svc, err := daemon.NewService()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Install service
	if err := svc.Install(cfg); err != nil {
		return fmt.Errorf("failed to install service: %w", err)
	}

	fmt.Println("Daemon service installed and started successfully")
	fmt.Printf("Sync interval: %s\n", cfg.Daemon.Interval)
	if len(cfg.Daemon.Projects) > 0 {
		fmt.Printf("Syncing projects: %v\n", cfg.Daemon.Projects)
	} else {
		fmt.Println("Syncing all projects")
	}
	fmt.Println()
	fmt.Println("The daemon is now running in the background and will start automatically at login.")
	fmt.Println()
	fmt.Println("Useful commands:")
	fmt.Println("  todu daemon status   - Check daemon status")
	fmt.Println("  todu daemon stop     - Stop the daemon")
	fmt.Println("  todu daemon restart  - Restart the daemon")
	fmt.Println("  todu daemon logs     - View daemon logs")

	return nil
}

func runDaemonUninstall(cmd *cobra.Command, args []string) error {
	// Get service manager for platform
	svc, err := daemon.NewService()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Stop daemon if running
	if err := svc.Stop(); err != nil {
		// Ignore error if daemon not running
		fmt.Println("Note: daemon was not running")
	}

	// Uninstall service
	if err := svc.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall service: %w", err)
	}

	fmt.Println("Daemon service uninstalled successfully")
	return nil
}

func runDaemonStatus(cmd *cobra.Command, args []string) error {
	// Read status file
	status, err := daemon.ReadStatus()
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}

	// Get service manager for platform
	svc, err := daemon.NewService()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Check if service is installed
	installed := svc.IsInstalled()

	// Display status
	fmt.Println("Daemon Status")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	if installed {
		fmt.Println("Service: Installed")
	} else {
		fmt.Println("Service: Not installed")
	}

	if status.Running {
		fmt.Println("Status: Running")
	} else {
		fmt.Println("Status: Stopped")
	}

	fmt.Println()

	if !status.LastSyncTime.IsZero() {
		fmt.Printf("Last sync: %s (%s ago)\n",
			status.LastSyncTime.Format("2006-01-02 15:04:05"),
			time.Since(status.LastSyncTime).Round(time.Second))
	} else {
		fmt.Println("Last sync: Never")
	}

	if !status.NextSyncTime.IsZero() {
		fmt.Printf("Next sync: %s (in %s)\n",
			status.NextSyncTime.Format("2006-01-02 15:04:05"),
			time.Until(status.NextSyncTime).Round(time.Second))
	}

	if status.LastSyncError != "" {
		fmt.Println()
		fmt.Printf("Last error: %s\n", status.LastSyncError)
		fmt.Printf("Error count: %d\n", status.ErrorCount)
	}

	return nil
}

func runDaemonStop(cmd *cobra.Command, args []string) error {
	// Get service manager for platform
	svc, err := daemon.NewService()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Stop daemon
	if err := svc.Stop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	fmt.Println("Daemon stopped successfully")
	return nil
}

func runDaemonRestart(cmd *cobra.Command, args []string) error {
	// Get service manager for platform
	svc, err := daemon.NewService()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Stop daemon
	fmt.Println("Stopping daemon...")
	if err := svc.Stop(); err != nil {
		// Ignore error if daemon not running
		fmt.Println("Note: daemon was not running")
	}

	// Wait a moment
	time.Sleep(1 * time.Second)

	// Start daemon
	fmt.Println("Starting daemon...")
	if err := svc.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Println("Daemon restarted successfully")
	return nil
}

func runDaemonLogs(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logPath := filepath.Join(homeDir, ".config", "todu", "daemon.log")

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Println("No log file found")
		fmt.Println("The daemon may not have run yet, or logs may be in system log")
		return nil
	}

	if daemonLogsFollow {
		// Use tail -f to follow logs
		tailCmd := exec.Command("tail", "-f", "-n", fmt.Sprintf("%d", daemonLogsLines), logPath)
		tailCmd.Stdout = os.Stdout
		tailCmd.Stderr = os.Stderr
		return tailCmd.Run()
	}

	// Show last N lines
	tailCmd := exec.Command("tail", "-n", fmt.Sprintf("%d", daemonLogsLines), logPath)
	output, err := tailCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to read logs: %w", err)
	}

	fmt.Print(string(output))
	return nil
}
