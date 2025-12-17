//go:build !release
// +build !release

package cmd

import (
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
)

// Register test plugins for development
func init() {
	// Register a test plugin
	_ = registry.Register("test", func() plugin.Plugin {
		return plugin.NewMockPlugin("test")
	})
}
