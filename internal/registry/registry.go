package registry

import (
	"fmt"
	"regexp"
	"sort"
	"sync"

	"github.com/evcraddock/todu.sh/pkg/plugin"
)

// PluginFactory is a function that creates a new plugin instance.
//
// Plugins should register a factory function that creates a fresh instance
// of the plugin. The registry will call this factory when Create is called.
//
// Example:
//
//	func init() {
//	    registry.Register("github", func() plugin.Plugin {
//	        return &GitHubPlugin{}
//	    })
//	}
type PluginFactory func() plugin.Plugin

// Registry manages registered plugins and their factory functions.
//
// The registry is thread-safe and can be used concurrently.
// Plugins register themselves by providing a factory function that
// creates new instances of the plugin.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]PluginFactory
}

// pluginNamePattern validates plugin names.
// Plugin names must be lowercase alphanumeric with hyphens.
var pluginNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// New creates a new plugin registry.
func New() *Registry {
	return &Registry{
		factories: make(map[string]PluginFactory),
	}
}

// Register registers a plugin factory function with the given name.
//
// The name must be lowercase, alphanumeric, and may contain hyphens.
// Returns an error if the name is invalid or already registered.
//
// Example:
//
//	reg := registry.New()
//	err := reg.Register("github", func() plugin.Plugin {
//	    return &GitHubPlugin{}
//	})
func (r *Registry) Register(name string, factory PluginFactory) error {
	// Validate plugin name
	if !pluginNamePattern.MatchString(name) {
		return fmt.Errorf("invalid plugin name %q: must be lowercase alphanumeric with hyphens", name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// List returns a sorted list of registered plugin names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Create creates a new plugin instance using the registered factory.
//
// The plugin is created, configured with the provided config map,
// and returned. Returns an error if:
//   - The plugin is not registered
//   - The plugin configuration fails
//
// Example:
//
//	config := map[string]string{
//	    "token": "ghp_abc123",
//	    "url": "https://api.github.com",
//	}
//	p, err := reg.Create("github", config)
func (r *Registry) Create(name string, config map[string]string) (plugin.Plugin, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %q not registered", name)
	}

	// Create new plugin instance
	p := factory()

	// Configure the plugin
	if err := p.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure plugin %q: %w", name, err)
	}

	return p, nil
}

// Default is the global default registry.
//
// Plugins can register themselves in init() functions using the
// package-level Register function, which uses this default registry.
var Default = New()

// Register registers a plugin with the default registry.
//
// This is a convenience function that calls Default.Register.
// Plugins typically call this in their init() function.
//
// Example:
//
//	func init() {
//	    registry.Register("github", func() plugin.Plugin {
//	        return &GitHubPlugin{}
//	    })
//	}
func Register(name string, factory PluginFactory) error {
	return Default.Register(name, factory)
}

// List returns registered plugin names from the default registry.
//
// This is a convenience function that calls Default.List.
func List() []string {
	return Default.List()
}

// Create creates a plugin instance from the default registry.
//
// This is a convenience function that calls Default.Create.
func Create(name string, config map[string]string) (plugin.Plugin, error) {
	return Default.Create(name, config)
}
