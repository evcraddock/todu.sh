# Unit 3.1: GitHub Plugin Structure

**Status**: 🔲 TODO

**Goal**: Set up GitHub plugin as separate module

**Prerequisites**: Unit 2.3 complete

**Estimated time**: 15 minutes

---

## Requirements

### 1. Plugin Module Structure

Create directory structure:

```
plugins/github/
├── go.mod           # Separate Go module
├── plugin.go        # Plugin implementation
├── client.go        # GitHub API client wrapper
├── mapper.go        # Type conversions
└── plugin_test.go   # Tests
```

### 2. Go Module Initialization

Initialize plugin module:
- `cd plugins/github && go mod init github.com/evcraddock/todu.sh/plugins/github`
- Add dependency on main module: `require github.com/evcraddock/todu.sh v0.1.0`
- Add GitHub API client: `github.com/google/go-github/v56/github`
- Add OAuth2 for auth: `golang.org/x/oauth2`

### 3. Plugin Struct

Create in `plugins/github/plugin.go`:

```go
type Plugin struct {
    client *client
    config map[string]string
}
```

### 4. Plugin Registration

In `plugin.go` init() function:
- Register with global registry
- Use name "github"

```go
func init() {
    registry.Register("github", func() plugin.Plugin {
        return &Plugin{}
    })
}
```

### 5. Metadata Methods

Implement:
- `Name() string` - Returns "github"
- `Version() string` - Returns plugin version (e.g., "1.0.0")

### 6. Configuration

Required configuration keys:
- `token` - GitHub personal access token
- `url` - GitHub API URL (optional, defaults to "https://api.github.com")

Implement:
- `Configure(config map[string]string) error`
  - Store config
  - Validate required fields
  - Create GitHub API client
- `ValidateConfig() error`
  - Check token is present
  - Validate URL format if provided

### 7. Import in Main Module

Update `cmd/todu/main.go` to import GitHub plugin:

```go
import _ "github.com/evcraddock/todu.sh/plugins/github"
```

This triggers plugin registration via init().

---

## Success Criteria

- ✅ Plugin module initialized
- ✅ Plugin struct defined
- ✅ Registered with global registry
- ✅ Metadata methods implemented
- ✅ Configuration methods work
- ✅ Plugin compiles: `go build ./plugins/github`
- ✅ Main binary includes plugin

---

## Verification

- `cd plugins/github && go build` - must compile
- `todu system config github` - shows configuration requirements
- Plugin appears in registry

---

## Commit Message

```text
feat: add GitHub plugin structure
```
