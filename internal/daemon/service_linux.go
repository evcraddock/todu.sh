package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/evcraddock/todu.sh/internal/config"
)

const systemdUnitTemplate = `[Unit]
Description=Todu Task Sync Daemon
After=network.target

[Service]
Type=simple
ExecStart={{.ExecutablePath}} daemon start
Restart=on-failure
RestartSec=10
StandardOutput=append:{{.LogPath}}
StandardError=append:{{.LogPath}}
WorkingDirectory={{.HomeDir}}
Environment="HOME={{.HomeDir}}"

[Install]
WantedBy=default.target
`

type linuxService struct {
	unitPath    string
	serviceName string
}

func newPlatformService() (Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	unitPath := filepath.Join(homeDir, ".config", "systemd", "user", "todu.service")

	return &linuxService{
		unitPath:    unitPath,
		serviceName: "todu.service",
	}, nil
}

func (s *linuxService) Install(cfg *config.Config) error {
	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Ensure systemd user directory exists
	systemdDir := filepath.Dir(s.unitPath)
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Prepare template data
	logPath := filepath.Join(homeDir, ".config", "todu", "daemon.log")

	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	data := struct {
		ExecutablePath string
		LogPath        string
		HomeDir        string
	}{
		ExecutablePath: execPath,
		LogPath:        logPath,
		HomeDir:        homeDir,
	}

	// Parse and execute template
	tmpl, err := template.New("unit").Parse(systemdUnitTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create unit file
	file, err := os.Create(s.unitPath)
	if err != nil {
		return fmt.Errorf("failed to create unit file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write unit file: %w", err)
	}

	// Reload systemd to recognize the new unit
	cmd := exec.Command("systemctl", "--user", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w\nOutput: %s", err, string(output))
	}

	// Enable the service
	cmd = exec.Command("systemctl", "--user", "enable", s.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable service: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *linuxService) Uninstall() error {
	// Check if unit file exists
	if _, err := os.Stat(s.unitPath); os.IsNotExist(err) {
		return fmt.Errorf("service not installed")
	}

	// Disable the service
	cmd := exec.Command("systemctl", "--user", "disable", s.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log but don't fail
		fmt.Printf("Note: %s\n", string(output))
	}

	// Stop the service
	cmd = exec.Command("systemctl", "--user", "stop", s.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log but don't fail
		fmt.Printf("Note: %s\n", string(output))
	}

	// Remove unit file
	if err := os.Remove(s.unitPath); err != nil {
		return fmt.Errorf("failed to remove unit file: %w", err)
	}

	// Reload systemd
	cmd = exec.Command("systemctl", "--user", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *linuxService) Start() error {
	cmd := exec.Command("systemctl", "--user", "start", s.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (s *linuxService) Stop() error {
	cmd := exec.Command("systemctl", "--user", "stop", s.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (s *linuxService) IsInstalled() bool {
	_, err := os.Stat(s.unitPath)
	return err == nil
}
