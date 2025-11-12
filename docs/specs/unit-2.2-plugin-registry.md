# Unit 2.2: Plugin Registry

**Status**: ✅ COMPLETE

**Goal**: Create plugin registry system for managing plugin instances

**Prerequisites**: Unit 2.1 complete

**Estimated time**: 25 minutes

---

## Requirements

### 1. Registry Structure

Create `internal/registry/registry.go` with:

- `Registry` struct that stores registered plugins
- `New() *Registry` constructor
- Thread-safe plugin storage (use sync.RWMutex)

### 2. Plugin Registration

Implement methods:

- `Register(name string, factory PluginFactory) error`
  - Registers a plugin factory function by name
  - Returns error if name already registered
  - Plugin names must be lowercase, alphanumeric with hyphens

- `List() []string`
  - Returns list of registered plugin names
  - Names sorted alphabetically

### 3. Plugin Creation

Implement method:

- `Create(name string, config map[string]string) (plugin.Plugin, error)`
  - Creates plugin instance using registered factory
  - Passes configuration to plugin
  - Calls Configure() on plugin
  - Returns error if plugin not registered
  - Returns error if configuration fails

### 4. Plugin Factory Type

Define in `internal/registry/registry.go`:

```go
type PluginFactory func() plugin.Plugin
```

### 5. Global Registry

Provide:

- `var Default = New()` - Global default registry
- `Register(name string, factory PluginFactory) error` - Package-level convenience function
- `Create(name string, config map[string]string) (plugin.Plugin, error)` - Package-level convenience function

This allows plugins to register themselves in init() functions.

### 6. Configuration Loading

Implement `internal/registry/config.go`:

- `LoadPluginConfig(pluginName string) (map[string]string, error)`
  - Loads plugin configuration from environment variables
  - Uses pattern: `TODU_PLUGIN_{PLUGINNAME}_{KEY}`
  - Example: `TODU_PLUGIN_GITHUB_TOKEN` for github plugin's token
  - Returns map of key-value pairs
  - All keys lowercase

### 7. Testing

Create `internal/registry/registry_test.go` with:

- Test plugin registration
- Test duplicate name handling
- Test plugin creation
- Test plugin listing
- Test configuration loading from environment

---

## Success Criteria

- ✅ Registry can register and retrieve plugins
- ✅ Thread-safe operations
- ✅ Plugin factories work correctly
- ✅ Configuration loading from environment works
- ✅ Tests pass: `go test ./internal/registry`
- ✅ Global registry available for use

---

## Verification

- `go test ./internal/registry -v` - all tests pass
- Can register multiple plugins
- Can create plugin instances with configuration

---

## Commit Message

```text
feat: implement plugin registry
```
