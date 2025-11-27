package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/evcraddock/todu.sh/internal/config"
)

const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.todu.daemon</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.ExecutablePath}}</string>
		<string>daemon</string>
		<string>start</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>{{.LogPath}}</string>
	<key>StandardErrorPath</key>
	<string>{{.LogPath}}</string>
	<key>WorkingDirectory</key>
	<string>{{.HomeDir}}</string>
	<key>EnvironmentVariables</key>
	<dict>
		<key>HOME</key>
		<string>{{.HomeDir}}</string>{{range $key, $value := .EnvVars}}
		<key>{{$key}}</key>
		<string>{{$value}}</string>{{end}}
	</dict>
</dict>
</plist>
`

type darwinService struct {
	plistPath string
	label     string
}

func newPlatformService() (Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "com.todu.daemon.plist")

	return &darwinService{
		plistPath: plistPath,
		label:     "com.todu.daemon",
	}, nil
}

func (s *darwinService) Install(cfg *config.Config) error {
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

	// Ensure LaunchAgents directory exists
	launchAgentsDir := filepath.Dir(s.plistPath)
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Prepare template data
	logPath := filepath.Join(homeDir, ".config", "todu", "daemon.log")

	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Collect all TODU_PLUGIN_* environment variables
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		// Split into key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// Only include TODU_PLUGIN_* environment variables
		if strings.HasPrefix(key, "TODU_PLUGIN_") {
			envVars[key] = value
		}
	}

	data := struct {
		ExecutablePath string
		LogPath        string
		HomeDir        string
		EnvVars        map[string]string
	}{
		ExecutablePath: execPath,
		LogPath:        logPath,
		HomeDir:        homeDir,
		EnvVars:        envVars,
	}

	// Parse and execute template
	tmpl, err := template.New("plist").Parse(launchdPlistTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create plist file
	file, err := os.Create(s.plistPath)
	if err != nil {
		return fmt.Errorf("failed to create plist file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Load the service
	cmd := exec.Command("launchctl", "load", s.plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to load service: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *darwinService) Uninstall() error {
	// Check if plist exists
	if _, err := os.Stat(s.plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service not installed")
	}

	// Unload the service
	cmd := exec.Command("launchctl", "unload", s.plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if service not loaded
		fmt.Printf("Note: %s\n", string(output))
	}

	// Remove plist file
	if err := os.Remove(s.plistPath); err != nil {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	return nil
}

func (s *darwinService) Start() error {
	// Use 'load' to start the service (works with KeepAlive)
	// If already loaded, this will fail, so check first
	if s.isLoaded() {
		return nil // Already running
	}

	cmd := exec.Command("launchctl", "load", s.plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (s *darwinService) Stop() error {
	// Use 'unload' instead of 'stop' to actually stop the service
	// This works correctly with KeepAlive: true
	if !s.isLoaded() {
		return nil // Already stopped
	}

	cmd := exec.Command("launchctl", "unload", s.plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// isLoaded checks if the service is currently loaded in launchd
func (s *darwinService) isLoaded() bool {
	cmd := exec.Command("launchctl", "list", s.label)
	err := cmd.Run()
	return err == nil // If no error, service is loaded
}

func (s *darwinService) IsInstalled() bool {
	_, err := os.Stat(s.plistPath)
	return err == nil
}
