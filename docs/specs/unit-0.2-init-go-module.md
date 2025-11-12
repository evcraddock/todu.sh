# Unit 0.2: Initialize Go Module

**Goal**: Create main Go module with basic structure

**Prerequisites**: Unit 0.1 complete

**Estimated time**: 10 minutes

---

## Requirements

### 1. Go Module Initialization

- Initialize Go module with name `github.com/yourorg/todu.sh`
- Module must be properly configured with `go.mod` file

### 2. Directory Structure

Create the following directory structure:

- `cmd/todu/` - Main CLI entry point
- `pkg/types/` - Shared type definitions
- `pkg/plugin/` - Public plugin interface
- `internal/api/` - Todu API client (private)
- `internal/config/` - Configuration management (private)
- `internal/registry/` - Plugin registry (private)
- `internal/sync/` - Sync engine (private)
- `plugins/` - Plugin implementations

### 3. Minimal Executable

- Create a minimal `main.go` in `cmd/todu/` that compiles
- Program must print a simple message identifying itself
- Must be executable

---

## Success Criteria

- ✅ `go.mod` file exists with correct module name
- ✅ All required directories are created
- ✅ `go build ./cmd/todu` compiles without errors
- ✅ Running the built binary produces expected output
- ✅ Code follows Go project structure conventions

---

## Verification

- Build command succeeds: `go build -o todu ./cmd/todu`
- Running `./todu` displays program identification
- All directories exist as specified

---

## Commit Message

```text
feat: initialize Go module and basic structure
```
