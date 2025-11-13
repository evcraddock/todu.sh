package daemon

import (
	"github.com/evcraddock/todu.sh/internal/config"
)

// Service defines the interface for platform-specific service management
type Service interface {
	// Install installs the daemon as a system service
	Install(cfg *config.Config) error

	// Uninstall removes the daemon service
	Uninstall() error

	// Start starts the daemon service
	Start() error

	// Stop stops the daemon service
	Stop() error

	// IsInstalled returns true if the service is installed
	IsInstalled() bool
}

// NewService creates a new platform-specific service manager
func NewService() (Service, error) {
	return newPlatformService()
}
