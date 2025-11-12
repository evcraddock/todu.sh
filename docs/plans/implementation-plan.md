# Todu CLI Implementation Plan

## Overview

Build a Go-based CLI application that syncs tasks from multiple 3rd party task management systems to a central Todu aggregation API. The CLI will be used primarily by LLMs to manage tasks across different systems.

## Architecture Summary

- **Language**: Go
- **Components**: Single binary with CLI + optional daemon mode
- **Plugin System**: Shared Go packages implementing common interface
- **Configuration**: YAML config file + environment variables for secrets
- **Output**: Text format (default) with JSON option via `--format` flag

## Implementation Stages

### Stage 1: Project Setup and Core Foundation

**Goal**: Set up project structure, core types, and basic CLI framework

**Tasks**:

1. Initialize Go module and project structure

   ```
   todu/                   # Monorepo root
   ├── go.mod              # Main module: github.com/yourorg/todu
   ├── cmd/todu/           # Main CLI entry point
   ├── internal/
   │   ├── api/            # Todu API client
   │   ├── config/         # Configuration management
   │   ├── registry/       # Plugin registry
   │   └── sync/           # Sync engine
   ├── pkg/
   │   ├── plugin/         # Public plugin interface
   │   └── types/          # Shared types
   ├── plugins/            # Each plugin is separate Go module
   │   ├── github/
   │   │   ├── go.mod      # github.com/yourorg/todu/plugins/github
   │   │   └── github.go
   │   ├── jira/           # Future
   │   │   └── go.mod
   │   └── todoist/        # Future
   │       └── go.mod
   └── docs/
   ```

2. Set up Go modules:
   - Main module: `go mod init github.com/yourorg/todu`
   - Note: Plugin modules will be created in Stage 3

3. Define core types in `pkg/types/`:
   - Task, Project, System, Label, Assignee, Comment
   - Match Todu API schema exactly

4. Set up CLI framework using `cobra`:
   - `todu` - Root command
   - `todu version` - Show version info
   - Basic flag parsing and help text

5. Implement configuration management (`internal/config/`):
   - YAML config file loading from `~/.todu/config.yaml`
   - Environment variable support with `TODU_*` prefix
   - Precedence: CLI flags > env vars > config file > defaults
   - Commands: `todu config show/get/set/validate`

6. Create Todu API client (`internal/api/`):
   - HTTP client wrapper
   - Methods for all API endpoints (systems, projects, tasks, comments, etc.)
   - Error handling and response parsing

**Success Criteria**:

- ✅ Project compiles successfully
- ✅ `todu version` shows version info
- ✅ `todu config` commands work (set/get/show)
- ✅ Can load config from file and env vars
- ✅ API client can make requests to Todu API (basic test)

**Tests**:

- Unit tests for config loading and precedence
- Unit tests for API client methods
- Integration test: API client against running Todu API

**Status**: Not Started

---

### Stage 2: Plugin Interface and Registry

**Goal**: Define plugin interface and create plugin registry system

**Tasks**:

1. Define plugin interface in `internal/plugin/interface.go`:

   ```go
   type Plugin interface {
       Name() string
       Version() string
       Configure(config map[string]string) error
       ValidateConfig() error
       FetchProjects() ([]*types.Project, error)
       FetchProject(externalID string) (*types.Project, error)
       FetchTasks(projectExternalID *string, since *time.Time) ([]*types.Task, error)
       FetchTask(projectExternalID *string, taskExternalID string) (*types.Task, error)
       CreateTask(projectExternalID *string, task *types.Task) (*types.Task, error)
       UpdateTask(projectExternalID *string, taskExternalID string, task *types.Task) (*types.Task, error)
       FetchComments(projectExternalID *string, taskExternalID string) ([]*types.Comment, error)
       CreateComment(projectExternalID *string, taskExternalID string, comment *types.Comment) (*types.Comment, error)
   }
   ```

2. Create plugin registry (`internal/plugin/registry.go`):
   - Register plugins by name
   - Load plugin configuration from env vars
   - Factory function to create plugin instances

3. Add standard errors:
   - `ErrNotSupported` for optional operations
   - `ErrNotConfigured` for missing configuration

4. CLI commands for system management:
   - `todu system list` - List registered systems
   - `todu system add <identifier>` - Register system in Todu API
   - `todu system config <identifier>` - Show system config requirements
   - `todu system remove <identifier>` - Remove system

**Success Criteria**:

- ✅ Plugin interface is well-defined and documented
- ✅ Registry can register and retrieve plugins
- ✅ `todu system` commands work
- ✅ Can configure plugins via environment variables

**Tests**:

- Unit tests for plugin registry
- Mock plugin implementation for testing
- Test system CRUD operations against API

**Status**: Not Started

---

### Stage 3: GitHub Plugin Implementation

**Goal**: Create first working plugin for GitHub Issues

**Tasks**:

1. Create GitHub plugin as separate Go module in `plugins/github/`:
   - Initialize `go.mod`: `module github.com/yourorg/todu/plugins/github`
   - Import main module: `require github.com/yourorg/todu v0.1.0`
   - `plugin.go` - Implements Plugin interface
   - `client.go` - GitHub API client wrapper
   - `mapper.go` - Convert between GitHub and Todu types

2. GitHub-specific mappings:
   - Repository → Project (external_id = "owner/repo")
   - Issue → Task (external_id = issue number as string)
   - Issue state: open/closed → active/done
   - Priority from labels (priority:high, etc.)
   - Comments map 1:1

3. Configuration via environment variables:
   - `TODU_GITHUB_TOKEN` - GitHub personal access token
   - `TODU_GITHUB_URL` - API URL (default: <https://api.github.com>)

4. Implement all Plugin interface methods:
   - FetchProjects() - List user's repositories
   - FetchTasks() - List issues for repo
   - CreateTask() - Create new issue
   - UpdateTask() - Update issue (title, description, state, labels)
   - FetchComments() / CreateComment() - Issue comments

5. Register GitHub plugin in registry

**Success Criteria**:

- ✅ GitHub plugin implements full interface
- ✅ Can list repositories
- ✅ Can fetch issues from a repository
- ✅ Can create/update issues
- ✅ Can fetch/create comments

**Tests**:

- Unit tests for type mappings
- Integration tests against GitHub API (using test repo)
- Test with real GitHub token

**Status**: Not Started

---

### Stage 4: Project Management CLI

**Goal**: CLI commands to manage projects and link to external systems

**Tasks**:

1. Implement project management commands:
   - `todu project list [--system <id>]` - List projects
   - `todu project add --system <identifier> --external-id <id> --name <name>` - Link project
   - `todu project show <id>` - Show project details
   - `todu project remove <id>` - Remove project

2. Project discovery command:
   - `todu project discover --system <identifier>` - List available projects from plugin
   - Uses plugin.FetchProjects() if supported

3. Output formatting:
   - Text format with tables
   - JSON format option
   - Handle `--format` flag

4. Validation:
   - Check system exists before linking project
   - Validate external_id format
   - Handle plugin errors gracefully

**Success Criteria**:

- ✅ Can list projects from Todu API
- ✅ Can link GitHub repo as project
- ✅ Can discover GitHub repos
- ✅ Can show project details
- ✅ Output formats work (text and JSON)

**Tests**:

- Integration tests for project CRUD
- Test project linking with GitHub plugin
- Test output formatting

**Status**: Not Started

---

### Stage 5: Sync Engine (Core Logic)

**Goal**: Implement bidirectional sync logic in core

**Tasks**:

1. Create sync engine in `internal/sync/engine.go`:
   - Sync orchestration logic
   - Conflict detection and resolution (last-write-wins)
   - Progress tracking and reporting

2. Sync algorithm:

   ```
   For each project:
     1. Get last sync time from project metadata
     2. Fetch tasks from plugin (since last sync)
     3. Fetch tasks from Todu API for this project
     4. For each external task:
        - Find in Todu by external_id
        - If not found: Create in Todu
        - If found: Compare updated_at, update if external is newer
     5. For each Todu task:
        - If has external_id: Check if needs push to external system
        - If external is older: Push update to external system
     6. Update last sync time in project metadata
   ```

3. Implement sync strategies:
   - Pull: External → Todu only
   - Push: Todu → External only
   - Bidirectional: Both directions

4. CLI commands:
   - `todu sync --project <id>` - Sync specific project
   - `todu sync --system <identifier>` - Sync all projects for system
   - `todu sync --all` - Sync all projects
   - `todu sync status` - Show last sync times

5. Sync result reporting:
   - Show tasks created/updated/skipped
   - Show any errors
   - Summary statistics

**Success Criteria**:

- ✅ Can sync GitHub issues to Todu API
- ✅ Can push Todu changes back to GitHub
- ✅ Handles new tasks, updates, and conflicts
- ✅ Reports sync progress and results
- ✅ Tracks last sync time

**Tests**:

- Unit tests for sync algorithm
- Integration tests with GitHub plugin and Todu API
- Test conflict resolution (same task updated in both systems)
- Test incremental sync (only changes since last sync)

**Status**: Not Started

---

### Stage 6: Task Management CLI

**Goal**: Full task CRUD operations via CLI

**Tasks**:

1. Implement task commands:
   - `todu task list [filters]` - List/search tasks
   - `todu task show <id>` - Show task details with comments
   - `todu task create` - Create new task (interactive or flags)
   - `todu task update <id>` - Update task fields
   - `todu task comment <id> <text>` - Add comment
   - `todu task close <id>` - Close/complete task

2. Filter options for `task list`:
   - `--status <status>` - Filter by status
   - `--priority <priority>` - Filter by priority
   - `--project <id>` - Filter by project
   - `--assignee <name>` - Filter by assignee
   - `--label <name>` - Filter by label
   - `--search <text>` - Full-text search
   - `--due-before <date>` / `--due-after <date>` - Due date range

3. Task creation:
   - Interactive mode (prompt for fields)
   - Flag-based: `--title`, `--description`, `--project`, `--priority`, etc.
   - Support for labels and assignees

4. Rich text output:
   - Task details with all fields
   - Comments threaded with timestamps
   - Color coding for status/priority
   - Tables for lists

**Success Criteria**:

- ✅ Can list tasks with various filters
- ✅ Can show task details with comments
- ✅ Can create tasks via CLI
- ✅ Can update task fields
- ✅ Can add comments to tasks
- ✅ Output is clear and helpful for LLMs

**Tests**:

- Integration tests for all task operations
- Test filtering and search
- Test output formatting
- Test task creation with/without project

**Status**: Not Started

---

### Stage 7: Daemon Mode and Background Sync

**Goal**: Run sync as a background daemon with periodic execution

**Tasks**:

1. Implement daemon mode:
   - `todu sync --daemon --interval <duration>` - Run as daemon
   - Graceful shutdown on SIGINT/SIGTERM
   - Logging to stdout/stderr
   - Health check endpoint or status file

2. Use `github.com/kardianos/service` for cross-platform service management:
   - Service wrapper implementation
   - Start/stop/restart logic

3. CLI commands for daemon management:
   - `todu daemon install --interval <duration>` - Install as system service
   - `todu daemon start` - Start daemon service
   - `todu daemon stop` - Stop daemon service
   - `todu daemon restart` - Restart daemon
   - `todu daemon status` - Show daemon status
   - `todu daemon uninstall` - Remove service

4. Platform-specific service configuration:
   - macOS: Generate launchd plist
   - Linux: Generate systemd unit file
   - Handle service installation and permissions

5. Configuration in `config.yaml`:

   ```yaml
   daemon:
     interval: 5m
     projects: []  # empty = all
   ```

6. Daemon behavior:
   - Run sync on interval
   - Log sync results
   - Handle errors gracefully (don't crash on API failures)
   - Exponential backoff on repeated failures

**Success Criteria**:

- ✅ Can run `todu sync --daemon` manually
- ✅ Can install as system service
- ✅ Daemon runs on schedule and syncs projects
- ✅ Can start/stop/check status of daemon
- ✅ Works on both macOS and Linux
- ✅ Logs are accessible

**Tests**:

- Test daemon startup and shutdown
- Test periodic sync execution
- Test service installation (manual testing)
- Test error handling and recovery

**Status**: Not Started

---

### Stage 8: Polish and Documentation

**Goal**: Finalize CLI, add comprehensive help, create user documentation

**Tasks**:

1. Enhance CLI help text:
   - Detailed descriptions for all commands
   - Examples in help output
   - Common workflows documented

2. Error messages:
   - Clear, actionable error messages
   - Suggest next steps when operations fail
   - Helpful hints for common issues

3. Output improvements:
   - Consistent formatting across commands
   - Progress indicators for long operations
   - Helpful tips and next-step suggestions

4. Create user documentation:
   - `README.md` - Quick start guide
   - `docs/installation.md` - Installation instructions
   - `docs/configuration.md` - Configuration reference
   - `docs/plugins.md` - Available plugins and configuration
   - `docs/workflows.md` - Common workflows and examples

5. Example workflows in docs:
   - Initial setup
   - Linking GitHub repositories
   - Running sync
   - Querying tasks
   - Setting up daemon

6. Create plugin development guide:
   - `docs/plugin-development.md`
   - Template for new plugins
   - Testing guidelines

**Success Criteria**:

- ✅ All commands have helpful descriptions
- ✅ Error messages guide users to solutions
- ✅ Documentation covers all features
- ✅ Examples demonstrate common workflows
- ✅ LLMs can understand how to use the CLI from help text

**Tests**:

- Manual testing of all help text
- Review error messages for clarity
- Have LLM test the CLI using only help/docs

**Status**: Not Started

---

### Stage 9: Additional Plugins (Future)

**Goal**: Implement additional task management system plugins

**Tasks**:

1. Jira plugin (`plugins/jira/`):
   - Jira API client
   - Project and issue mapping
   - Custom field handling

2. Todoist plugin (`plugins/todoist/`):
   - Todoist API client
   - Project and task mapping
   - Handle Todoist-specific features (sections, etc.)

3. Forgejo plugin (`plugins/forgejo/`):
   - Forgejo API client (similar to GitHub)
   - Repository and issue mapping

**Success Criteria**:

- ✅ Each plugin implements full interface
- ✅ Can sync tasks from respective systems
- ✅ Documented in plugin documentation

**Tests**:

- Integration tests for each plugin
- Test against real instances

**Status**: Not Started

---

## Dependencies and Libraries

### Module Structure

This project uses a monorepo with multiple Go modules:

- **Main module**: `github.com/yourorg/todu` - CLI, core logic, public interfaces
- **Plugin modules**: Each plugin is a separate Go module
  - `github.com/yourorg/todu/plugins/github`
  - `github.com/yourorg/todu/plugins/jira` (future)
  - `github.com/yourorg/todu/plugins/todoist` (future)

Plugins import the main module to access the plugin interface and shared types.

### Core Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/kardianos/service` - Cross-platform service management

### API Clients

- `github.com/google/go-github/v56/github` - GitHub API client
- Standard library `net/http` for Todu API client

### Testing

- Standard library `testing`
- `github.com/stretchr/testify` - Test assertions and mocks

## Development Guidelines

- Follow incremental development: each stage should compile and pass tests
- Write tests before or during implementation
- Commit working code frequently
- Update this plan as implementation progresses
- Mark stages as "In Progress" or "Complete" as you work

## Success Metrics

- CLI can successfully sync tasks between GitHub and Todu API
- LLM can use CLI commands to manage tasks across systems
- Daemon runs reliably in background
- Code is well-tested (>80% coverage for core logic)
- Documentation is clear and complete

## Timeline Estimate

- Stage 1: 2-3 days
- Stage 2: 2 days
- Stage 3: 3-4 days
- Stage 4: 2 days
- Stage 5: 4-5 days (most complex)
- Stage 6: 3 days
- Stage 7: 3-4 days
- Stage 8: 2 days
- Stage 9: Variable (per plugin: 2-3 days each)

**Total for MVP (Stages 1-8)**: ~3-4 weeks

## Notes

- Start with GitHub plugin as first implementation
- Focus on making CLI output LLM-friendly
- Keep plugins simple - they're just data adapters
- Sync logic stays in core for consistency
- Configuration should be easy to understand and debug
