# Unit 1.2: Configuration Loading (Basic)

**Status**: ✅ COMPLETE

**Goal**: Load configuration from YAML file and environment variables

**Prerequisites**: Unit 1.1 complete

**Estimated time**: 20 minutes

---

## Requirements

### 1. Dependencies

- Add Viper dependency for configuration management
- Add YAML v3 dependency for YAML parsing

### 2. Configuration Structure

Create configuration types in `internal/config/config.go`:

- `Config` - Main configuration struct
- `DaemonConfig` - Daemon-specific settings
- `OutputConfig` - Output formatting settings

### 3. Configuration Fields

`Config` must include:

- `APIURL` - Base URL for Todu API
- `Daemon` - Daemon configuration
- `Output` - Output configuration

`DaemonConfig` must include:

- `Interval` - Sync interval (e.g., "5m")
- `Projects` - List of project IDs to sync (empty = all)

`OutputConfig` must include:

- `Format` - Output format ("text" or "json")
- `Color` - Whether to use colored output (bool)

### 4. Configuration Loading

Implement `Load()` function that:

- Sets sensible defaults:
  - API URL: `http://localhost:8000`
  - Daemon interval: `5m`
  - Daemon projects: empty list
  - Output format: `text`
  - Output color: `true`
- Searches for config files in:
  - `~/.todu/config.yaml`
  - `~/.config/todu/config.yaml`
  - `./config.yaml` (current directory)
- Supports environment variables with `TODU_` prefix
- Gracefully handles missing config file (use defaults)
- Returns error only for parsing errors, not missing file

### 5. Configuration Precedence

Priority (highest to lowest):

1. Environment variables (`TODU_*`)
2. Config file
3. Defaults

### 6. Testing

Create `internal/config/config_test.go` with:

- Test loading defaults when no config file exists
- Test that defaults are correct
- Verify all configuration fields are properly set

---

## Success Criteria

- ✅ Dependencies added to `go.mod`
- ✅ Config package compiles successfully
- ✅ Can load defaults when no config file exists
- ✅ No errors for missing config file
- ✅ Tests pass: `go test ./internal/config`
- ✅ Returns error only for actual parsing issues

---

## Verification

- `go test ./internal/config -v` - all tests must pass
- Config loads successfully with no config file present

---

## Commit Message

```text
feat: add configuration loading with viper
```
