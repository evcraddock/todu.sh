# Task Management Application - Brainstorming Session

## Project Overview

**Type**: CLI Application
**Primary User**: LLM (Language Model)
**Core Functionality**: Aggregate and sync tasks from multiple 3rd party systems

## Key Requirements

- Plugin system for 3rd party task management systems
- Sync tasks to aggregation system via API
- Generic plugin interface to support multiple task management systems

## Brainstorming Questions & Answers

### Question 1: Core Use Case

**Asked**: What is the primary use case for this application? Is the LLM:

- Creating/updating tasks across multiple systems from a single interface?
- Querying/searching tasks across all systems in one place?
- Generating reports/summaries from aggregated task data?
- All of the above?

**Answer**: All of the above (D). All queries should be against the aggregation API.

**Implications**:

- Aggregation system is the central source of truth for all LLM queries
- Plugins sync data TO aggregation system
- Need bidirectional sync (changes in aggregation must propagate back to 3rd party systems)
- LLM never directly interacts with 3rd party systems

### Question 2: Synchronization Pattern

**Asked**: How should synchronization work between the 3rd party systems and the aggregation API?

**Options**:

- **A)** Real-time webhooks (3rd party systems push changes immediately to aggregation)
- **B)** Polling (CLI periodically fetches from 3rd party systems and updates aggregation)
- **C)** Manual/on-demand sync (user/LLM triggers sync explicitly)
- **D)** Hybrid approach (combination of the above)

**Answer**: Hybrid approach - Background service + CLI trigger capability

**Implications**:

- Need a background service/daemon that handles periodic syncing
- CLI can trigger on-demand sync when needed (e.g., before critical queries)
- Architecture now consists of:
  - **CLI Application** (LLM interface)
  - **Background Service** (sync daemon)
  - **Aggregation API** (central data store)
  - **Plugins** (3rd party system adapters)

### Question 3: Architecture Boundaries

**Asked**: Where do the plugins live and how are they shared between components?

**Options**:

- **A)** Plugins are part of the CLI, and CLI communicates with background service to trigger syncs
- **B)** Plugins are part of the background service only, CLI just sends commands via API
- **C)** Plugins are shared library/modules used by both CLI and background service
- **D)** Plugins are separate processes that both CLI and service communicate with

**Context**: We need to decide:

- Does the CLI need direct access to 3rd party systems (e.g., for immediate create operations)?
- Or does CLI only talk to aggregation API, and background service handles all 3rd party interactions?
- How do we avoid code duplication if both need plugin functionality?

**Answer**: C - Plugins are shared library/modules used by both CLI and background service

**Implications**:

- Plugins are reusable modules/libraries
- Both CLI and background service import and use the same plugin code
- Need a clean plugin interface/contract that both components can depend on
- Programming language should support good module/library systems
- Plugin configuration can be shared between CLI and service

### Question 4: Programming Language

**Asked**: What programming language should we use for this application?

**Considerations**:

- Needs good CLI library support (argument parsing, output formatting)
- Needs good API client libraries for common task systems
- Needs ability to run as a background service/daemon
- Needs good module/plugin system
- Needs to be LLM-friendly (easy for LLM to work with)
- Should be maintainable and have good tooling

**Common Options**:

- **A)** Python (rich ecosystem, great for APIs, easy scripting)
- **B)** Go (great for CLI tools, compiled binaries, good concurrency)
- **C)** TypeScript/Node.js (good ecosystem, familiar to many)
- **D)** Rust (performance, safety, compiled binaries)
- **E)** Other (specify)

**Answer**: B - Go (Golang)

**Implications**:

- Single compiled binary for easy distribution (CLI and service)
- Excellent concurrency support with goroutines (important for background service)
- Strong standard library for HTTP clients, JSON handling
- Good CLI libraries (cobra, cli, etc.)
- Interfaces provide natural plugin boundaries
- Fast compilation and execution
- Static typing helps with maintainability

**Potential Challenges**:

- Less "batteries included" than Python for some API integrations
- May need to write more boilerplate

### Question 5: Plugin Interface Design

**Asked**: What should the generic plugin interface look like?

**Context**: We need a contract that every task management plugin must implement, regardless of whether it's GitHub Issues, Jira, Todoist, Linear, etc.

**Core operations likely needed**:

- Fetch tasks/issues from 3rd party system
- Create new task in 3rd party system
- Update existing task in 3rd party system
- Delete/close task in 3rd party system
- Map 3rd party task format to unified/canonical format
- Map unified format back to 3rd party format

**Question**: What fields should our unified/canonical task model have?

**Common fields across systems**:

- ID, Title, Description, Status, Priority, Assignee, Labels/Tags, Due Date, Created/Updated timestamps, Comments

**What additional fields are important for your use case?**

**Answer**: The aggregation API already defines the unified task model!

**Existing Data Model** (from API):

**Task Model**:

- `id` (int) - Unique task ID (Todu internal)
- `external_id` (string, max 255 chars, optional) - External system task identifier (e.g., "123" for GitHub issue #123)
- `source_url` (string, max 500 chars, optional) - URL to external task source
- `title` (string, 1-500 chars) - Task title
- `description` (string, max 5000 chars, optional) - Task description
- `project_id` (int, optional) - Associated project
- `status` (string) - active, inprogress, waiting, done, cancelled
- `priority` (string, optional) - high, medium, low
- `due_date` (date, optional) - Due date
- `created_at`, `updated_at` (datetime) - Timestamps
- `labels` (array) - Associated labels
- `assignees` (array) - Associated assignees

**Project Model**:

- `id` (int) - Unique project ID
- `name` (string, 1-200 chars) - Project name
- `description` (string, max 2000 chars, optional)
- `system_id` (int, optional) - Links to external system
- `external_id` (string, optional) - External system identifier (e.g., "owner/repo")
- `status` (string) - active, done, cancelled
- `created_at`, `updated_at` (datetime)

**System Model** (represents external task systems):

- `id` (int) - Unique system ID
- `identifier` (string, unique) - Lowercase slug (e.g., "github", "jira")
- `name` (string) - Display name
- `url` (string) - Base URL of the system
- `metadata` (key-value pairs) - Flexible config storage
- `created_at`, `updated_at` (datetime)

**Supporting Models**:

- **Label**: id, name
- **Assignee**: id, name
- **Comment**: id, task_id, author, content, timestamps
- **ActivityLog**: id, task_id, action, changes, user, timestamp (auto-generated)

**Key Design Features**:

- Generic "System" model allows any external task system to be registered
- Flexible metadata on systems for plugin-specific configuration
- External_id on projects links to external system identifiers
- No hardcoded system types - fully extensible

### Question 6: Plugin Responsibilities and Sync Strategy

**Asked**: What should each plugin be responsible for, and how should syncing work?

**Context**: Now that we know the API structure, we need to define what the plugins do.

**Plugin responsibilities could include**:

1. **Mapping** - Convert between external system format and Todu API format
2. **Authentication** - Handle auth tokens/credentials for external system
3. **API calls** - Make HTTP requests to external system
4. **Conflict resolution** - Handle cases where both systems have changes
5. **Metadata management** - Store plugin-specific config in system.metadata

**Sync strategy questions**:

- Should plugins track "last synced" timestamps to only fetch changes?
- Should we use external_id to match tasks, or maintain a separate mapping table?
- How do we handle conflicts (e.g., task updated in both systems)?
- Should plugins support incremental sync or always full sync?

**Answer**: With `external_id` now available, here's the proposed approach:

**Plugin Responsibilities**:

1. **Mapping** - Convert between external system format and Todu API format
2. **Authentication** - Handle auth tokens/credentials for external system
3. **API calls** - Fetch/create/update tasks in external system
4. **ID Management** - Use `external_id` + `project_id` to uniquely identify tasks
5. **Metadata** - Store plugin config in `system.metadata` (e.g., API tokens, base URLs)

**Sync Strategy**:

- Use `external_id` to match tasks (e.g., GitHub issue number "123")
- Query: `GET /api/v1/tasks?project_id=X&external_id=Y` to find existing task
- If exists: UPDATE, if not: CREATE
- Store `source_url` for easy linking back to external system
- Track `updated_at` timestamps to detect conflicts
- For incremental sync: plugins can use metadata to store "last_synced_at"

**Conflict Resolution** (simple approach to start):

- Last-write-wins based on `updated_at` timestamps
- Future: Can add conflict detection and manual resolution

### Question 7: CLI Commands and Workflow

**Asked**: What CLI commands should the application support?

**Context**: The LLM will be the primary user. What operations should be available?

**Possible commands**:

- **System Management**: Register/configure external systems (GitHub, Jira, etc.)
- **Project Management**: Link external projects to Todu (e.g., "owner/repo" → Todu project)
- **Sync Operations**: Trigger sync for specific project or all projects
- **Task Operations**: Create/update/query tasks (or just use API directly?)
- **Configuration**: Manage auth tokens, sync intervals, etc.

**Example workflow**:

```bash
# Register a system
todu system add github --url https://api.github.com --token xxx

# Link a project
todu project add --system github --external-id "owner/repo" --name "My Repo"

# Sync
todu sync --project "My Repo"
# or
todu sync --all

# Query (maybe not needed if LLM can query API directly?)
todu task list --status active
```

**Question**: What commands are most important for your workflow?

**Answer**: The LLM will use the CLI exclusively (not the API directly). CLI responses should help the LLM decide what to do next.

**Key Design Implications**:

1. **LLM-Friendly Output**: CLI should output structured, informative responses (JSON or rich text)
2. **Complete Coverage**: CLI must expose ALL operations (tasks, projects, systems, sync, comments, etc.)
3. **Contextual Guidance**: Responses should include helpful context for next steps
4. **Error Messages**: Should guide the LLM on how to fix issues
5. **Status Feedback**: Sync operations should provide detailed progress/results

**Required Command Groups**:

1. **System Management**:
   - `todu system list` - Show registered systems
   - `todu system add <identifier>` - Register new system
   - `todu system config <identifier>` - Configure system (tokens, etc.)
   - `todu system remove <identifier>` - Remove system

2. **Project Management**:
   - `todu project list [--system <id>]` - List projects
   - `todu project add --system <id> --external-id <id> --name <name>` - Link project
   - `todu project show <id>` - Show project details
   - `todu project remove <id>` - Remove project

3. **Sync Operations**:
   - `todu sync --project <id>` - Sync specific project
   - `todu sync --system <id>` - Sync all projects for a system
   - `todu sync --all` - Sync everything
   - `todu sync status` - Show last sync times, status

4. **Task Operations**:
   - `todu task list [filters]` - Search/list tasks
   - `todu task show <id>` - Show task details with comments
   - `todu task create` - Create new task
   - `todu task update <id>` - Update task
   - `todu task comment <id>` - Add comment
   - `todu task close <id>` - Close task

5. **Configuration**:
   - `todu config set <key> <value>` - Set config
   - `todu config get <key>` - Get config
   - `todu config list` - Show all config

### Question 8: CLI Output Format

**Asked**: What output format should the CLI use for LLM consumption?

**Options**:

- **A)** JSON only (structured, easy to parse programmatically)
- **B)** Human-readable text only (easier for LLMs to understand contextually)
- **C)** Both with a flag (e.g., `--format json` or `--format text`)
- **D)** Rich structured text (tables, colors, but still parseable)

**Considerations**:

- LLMs are good at parsing both structured JSON and well-formatted text
- JSON is precise but can be verbose
- Human-readable text can include helpful hints and guidance
- Some commands might benefit from different formats (e.g., lists as tables vs. details as JSON)

**Example outputs**:

**JSON format** for `todu task list`:

```json
{
  "total": 5,
  "tasks": [
    {
      "id": 1,
      "title": "Fix authentication bug",
      "status": "active",
      "priority": "high",
      "project": "my-app",
      "external_id": "123",
      "source_url": "https://github.com/owner/repo/issues/123"
    }
  ]
}
```

**Text format** for `todu task list`:

```
Found 5 tasks:

ID  Title                    Status      Priority  Project
1   Fix authentication bug   active      high      my-app
2   Update documentation     done        medium    my-app
...

Use 'todu task show <id>' to see details.
```

**Question**: Which format would work best for your LLM workflow?

**Answer**: Support `--format` flag with text as the default (C)

**Implementation Details**:

- All commands accept `--format` flag with values: `text` (default), `json`
- Text format: Human-readable, includes helpful hints and next-step suggestions
- JSON format: Structured data for programmatic parsing when needed
- Default to text so it's natural for both LLMs and humans testing the CLI

**Examples**:

```bash
# Uses text format (default)
todu task list --status active

# Explicit text format
todu task list --status active --format text

# JSON format when needed
todu task list --status active --format json
```

### Question 9: Plugin Interface Design

**Asked**: What should the Go plugin interface look like?

**Context**: We need to define a Go interface that all plugins implement. This is the contract between the core system and each external task system plugin.

**Proposed Interface**:

```go
type Plugin interface {
    // Metadata
    Name() string        // e.g., "github", "jira", "todoist"
    Version() string     // Plugin version

    // Authentication
    Configure(config map[string]string) error
    ValidateConfig() error

    // System operations
    RegisterSystem(apiBaseURL string) (*System, error)

    // Project operations
    FetchProjects() ([]*Project, error)
    FetchProject(externalID string) (*Project, error)

    // Task sync operations
    FetchTasks(projectExternalID string, since *time.Time) ([]*Task, error)
    FetchTask(projectExternalID string, taskExternalID string) (*Task, error)
    CreateTask(projectExternalID string, task *Task) (*Task, error)
    UpdateTask(projectExternalID string, taskExternalID string, task *Task) (*Task, error)

    // Comment operations
    FetchComments(projectExternalID string, taskExternalID string) ([]*Comment, error)
    CreateComment(projectExternalID string, taskExternalID string, comment *Comment) (*Comment, error)
}
```

**Questions**:

- Is this interface complete for basic sync operations?
- Should plugins handle bidirectional sync logic, or should core handle it?
- Do we need webhook support in the interface?
- **Why do we need FetchProjects() and FetchProject()?**

**Discussion on FetchProjects()/FetchProject()**:

**Possible use cases**:

- **Discovery**: Help LLM discover available projects (e.g., "show me all my GitHub repos")
- **Validation**: Verify external_id exists before linking
- **Metadata**: Fetch project description, settings, etc.

**Counter-argument**:

- Users already know which projects to track (e.g., "owner/repo")
- Can just manually specify external_id when linking
- Adds complexity to plugin interface
- Not needed for core sync functionality

**Question**: Do you want:

- **A)** Keep FetchProjects() for discovery/listing
- **B)** Remove it - users manually specify external_id
- **C)** Make it optional (some plugins implement, some don't)

**Answer**: _Pending_

### Question 10: Handling 3rd Party Systems Without Projects

**Asked**: How should the plugin interface handle 3rd party task management systems that don't have the concept of a project?

**Context**:

- GitHub: Has repositories (maps to projects) ✓
- Jira: Has projects ✓
- Todoist: Has projects ✓
- Simple todo apps: Might just have a flat list of tasks with no grouping
- Trello: Has boards (could map to projects)
- Some systems: Only workspaces, tags, or no organization at all

**Current Plugin Interface Issue**:

```go
// This assumes projects exist!
FetchTasks(projectExternalID string, since *time.Time) ([]*Task, error)
CreateTask(projectExternalID string, task *Task) (*Task, error)
```

**Problem**: If a 3rd party system doesn't have projects, what does the plugin pass for `projectExternalID`?

**Options for Plugin Interface Design**:

**A) Make projectExternalID optional (nullable)**:

```go
FetchTasks(projectExternalID *string, since *time.Time) ([]*Task, error)
CreateTask(projectExternalID *string, task *Task) (*Task, error)
```

- Pass `nil` for systems without projects
- Plugin handles both cases
- Pro: Explicit, clean interface
- Con: Every plugin must handle nil case

**B) Separate methods for system-level vs project-level**:

```go
// System-level (for project-less systems)
FetchAllTasks(since *time.Time) ([]*Task, error)
CreateTaskInSystem(task *Task) (*Task, error)

// Project-level (for systems with projects)
FetchProjectTasks(projectExternalID string, since *time.Time) ([]*Task, error)
CreateTaskInProject(projectExternalID string, task *Task) (*Task, error)
```

- Plugins implement only relevant methods
- Pro: Clear semantics
- Con: More methods, need to know which to call

**C) Empty string convention**:

```go
FetchTasks(projectExternalID string, since *time.Time) ([]*Task, error)
// Pass "" for systems without projects
```

- Simple, no interface changes
- Pro: Minimal change
- Con: Implicit convention, less type-safe

**D) Configuration-driven**:

```go
// Plugin declares capabilities
type PluginCapabilities struct {
    SupportsProjects bool
    // ...
}

// If SupportsProjects == false, core doesn't pass projectID
FetchTasks(projectExternalID string, since *time.Time) ([]*Task, error)
```

- Core knows plugin capabilities
- Pro: Self-documenting
- Con: Added complexity

**Question**: Which approach for the plugin interface makes the most sense?

**Answer**: A - Make projectExternalID optional (nullable using *string)

**Implementation**:

```go
type Plugin interface {
    // Metadata
    Name() string
    Version() string

    // Authentication
    Configure(config map[string]string) error
    ValidateConfig() error

    // Task sync operations
    // projectExternalID can be nil for systems without projects
    FetchTasks(projectExternalID *string, since *time.Time) ([]*Task, error)
    FetchTask(projectExternalID *string, taskExternalID string) (*Task, error)
    CreateTask(projectExternalID *string, task *Task) (*Task, error)
    UpdateTask(projectExternalID *string, taskExternalID string, task *Task) (*Task, error)

    // Comment operations
    FetchComments(projectExternalID *string, taskExternalID string) ([]*Comment, error)
    CreateComment(projectExternalID *string, taskExternalID string, comment *Comment) (*Comment, error)
}
```

**Implications**:

- Explicit and type-safe
- GitHub plugin passes `&"owner/repo"`
- Simple todo list plugin passes `nil`
- Each plugin must handle both cases appropriately
- Clear in function signature that projects are optional

**Still need to resolve from Question 9**:

- Should we keep FetchProjects()/FetchProject() methods?
- Should plugins handle sync logic or just fetch/push?
- Do we need webhook support?

### Question 11: Background Service Architecture

**Asked**: How should the background sync service be structured?

**Context**: We decided earlier to have a background service that handles periodic syncing. Now we need to design it.

**Options**:

**A) Separate daemon binary**:

- `todu` - CLI tool
- `todud` - Background daemon
- Daemon runs continuously, CLI communicates with it
- Pro: Clean separation, daemon can run as system service
- Con: Two binaries to manage

**B) Single binary with daemon mode**:

- `todu sync --daemon` - Runs as daemon
- `todu task list` - Normal CLI commands
- Same binary, different modes
- Pro: Single binary to distribute
- Con: More complex initialization logic

**C) No persistent daemon**:

- `todu sync --all` - Run sync and exit
- Users/cron schedule it periodically
- No long-running process
- Pro: Simplest architecture
- Con: No real-time syncing, relies on external scheduler

**D) Hybrid - Optional daemon**:

- Works standalone: `todu sync --all` (run and exit)
- Can run as daemon: `todu sync --daemon --interval 5m`
- Best of both worlds
- Pro: Flexible
- Con: Most complex

**Question**: Which architecture fits your needs best?

**Answer**: Leaning towards daemon - need to understand cross-platform implications

**Daemon Implementation on Different OSes**:

**macOS (launchd)**:

- Uses `launchd` as the service manager
- Configuration via plist files in `~/Library/LaunchAgents/` (user) or `/Library/LaunchDaemons/` (system)
- Example plist:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yourname.todu</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/todu</string>
        <string>sync</string>
        <string>--daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

- Commands: `launchctl load/unload/start/stop`
- Logs go to system log (viewable via Console.app or `log` command)

**Linux (systemd)**:

- Most modern Linux distros use `systemd`
- Configuration via unit files in `/etc/systemd/system/` (system) or `~/.config/systemd/user/` (user)
- Example service file:

```ini
[Unit]
Description=Todu Task Sync Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/todu sync --daemon
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

- Commands: `systemctl start/stop/enable/disable todu`
- Logs go to journald: `journalctl -u todu -f`

**Older Linux (init.d/upstart)**:

- Some systems still use init scripts in `/etc/init.d/`
- More complex to write and maintain
- Generally being phased out in favor of systemd

**Implementation Comparison**:

**Option B: Single binary with daemon mode** (`todu sync --daemon`)

- **Pro**: Simple installation - just copy one binary
- **Pro**: Same binary for all operations
- **Pro**: Easy testing without installing service
- **Setup**:
  - Copy binary to `/usr/local/bin/todu`
  - Install plist/service file
  - Service file calls `todu sync --daemon`
- **Development**: Easy to test daemon mode locally

**Option A: Separate daemon binary** (`todu` + `todud`)

- **Pro**: Cleaner separation of concerns
- **Pro**: Daemon can be minimal/optimized
- **Con**: Must distribute two binaries
- **Setup**:
  - Copy both binaries
  - Install plist/service file
  - Service file calls `todud`
- **Development**: Need to build/test two binaries

**Cross-platform Management Options**:

**1) Manual installation**:

- User manually creates plist/service file
- Pro: Simple, no magic
- Con: Error-prone, requires platform knowledge

**2) CLI-assisted setup**:

```bash
# Generates and installs service file
todu daemon install

# Starts the daemon
todu daemon start

# Stops the daemon
todu daemon stop

# Removes service file
todu daemon uninstall
```

- Pro: User-friendly, cross-platform
- Con: CLI must understand each platform's service system
- Libraries: Go has libraries like `github.com/kardianos/service` that abstract this

**3) Installation script**:

- Separate shell script that sets up service
- Pro: Flexible, can handle edge cases
- Con: Another thing to maintain

**Recommendation**:

Given your needs (LLM-friendly CLI), I'd suggest:

**Option B (Single binary) + CLI-assisted daemon management**:

```bash
# One binary to install
todu sync --daemon --interval 5m    # Run manually for testing

# Or install as service
todu daemon install --interval 5m   # Creates service config
todu daemon start                   # Starts service
todu daemon status                  # Shows status
```

**Why this works well**:

- Single binary is easiest to distribute and install
- Can test daemon behavior without installing service
- CLI commands to manage daemon are LLM-friendly
- Libraries like `kardianos/service` handle platform differences
- Daemon mode and one-off sync use same code path

**Question**: Does this help? Should we go with Option B + CLI daemon management?

**Answer**: Yes, Option B (Single binary with daemon mode) + CLI-assisted daemon management

**Selected Architecture**:

- Single binary: `todu`
- Daemon mode: `todu sync --daemon --interval <duration>`
- Service management: `todu daemon install/start/stop/status/uninstall`
- Use `github.com/kardianos/service` library for cross-platform service management
- Support both manual daemon execution and system service installation

### Question 12: Configuration Management

**Asked**: How should configuration be stored and managed?

**Context**: Need to store:

- API base URL (for Todu API)
- System configurations (tokens, URLs for each plugin)
- Daemon settings (sync interval, enabled projects)
- User preferences

**Options**:

**A) Single config file (YAML/JSON/TOML)**:

```yaml
# ~/.todu/config.yaml
api_url: http://localhost:8000
daemon:
  interval: 5m
  enabled: true

systems:
  github:
    token: ghp_xxxxxxxxxxxx
    url: https://api.github.com
  jira:
    token: xxx
    url: https://mycompany.atlassian.net
```

- Pro: Simple, human-readable, easy to edit
- Con: Sensitive tokens in plaintext

**B) Config file + Secure credential storage**:

- Config file for settings
- System keychain for tokens (macOS Keychain, Linux Secret Service)
- Example: `todu system config github --token xxx` stores in keychain
- Pro: Secure token storage
- Con: More complex, platform-specific

**C) Environment variables + Config file**:

- Sensitive values from env vars: `TODU_GITHUB_TOKEN`
- Settings from config file
- Pro: Follows 12-factor app principles
- Con: Can be cumbersome for multiple systems

**D) Store everything in Todu API**:

- Use API's system.metadata to store plugin configs
- CLI just needs API URL + maybe API key
- Pro: Centralized, can sync config across machines
- Con: Tokens stored in remote API (security concern)

**Question**: Which configuration approach fits your workflow?

**Answer**: C - Environment variables + Config file

**Implementation**:

**Config file** (`~/.todu/config.yaml` or `~/.config/todu/config.yaml`):

```yaml
# Todu API configuration
api_url: http://localhost:8000

# Daemon settings
daemon:
  interval: 5m
  # Which projects to sync (empty = all)
  projects: []

# Output preferences
output:
  format: text  # text or json
  color: true
```

**Environment variables for sensitive data**:

```bash
# Todu API (if authentication is added later)
TODU_API_KEY=xxx

# System tokens
TODU_GITHUB_TOKEN=ghp_xxxxxxxxxxxx
TODU_JIRA_TOKEN=xxx
TODU_TODOIST_TOKEN=xxx

# System URLs (optional, override defaults)
TODU_GITHUB_URL=https://api.github.com
TODU_JIRA_URL=https://mycompany.atlassian.net
```

**Precedence** (highest to lowest):

1. Command-line flags: `--api-url`, `--format`, etc.
2. Environment variables: `TODU_*`
3. Config file: `~/.todu/config.yaml`
4. Built-in defaults

**Benefits**:

- Sensitive tokens never written to disk
- Config file can be version controlled (no secrets)
- Follows 12-factor app principles
- Works well in containerized environments
- LLM can read config file for context, but not tokens

**CLI helper commands**:

```bash
# Show current configuration (redacts tokens)
todu config show

# Set config file values
todu config set daemon.interval 10m
todu config get daemon.interval

# Validate configuration (checks env vars exist)
todu config validate
```

### Summary: Ready to Start Implementation?

We've now designed:

1. ✅ Architecture: CLI + Background Service + Plugins + Aggregation API
2. ✅ Language: Go
3. ✅ Plugin system: Shared libraries with generic interface
4. ✅ Sync strategy: Using external_id + bidirectional sync
5. ✅ CLI commands: Complete coverage of all operations
6. ✅ Output format: Text (default) with --format json option
7. ✅ Plugin interface: Nullable projectExternalID for flexibility
8. ✅ Daemon: Single binary with daemon mode + CLI management
9. ✅ Configuration: Env vars for secrets + config file for settings

**Still need to decide**:

- ~~Should we keep FetchProjects() in plugin interface?~~ ✅ Yes, but optional
- ~~Should plugins handle sync logic or just fetch/push data?~~ ✅ Just fetch/push, core handles sync logic

**Final Plugin Interface Design**:

```go
type Plugin interface {
    // Metadata
    Name() string        // e.g., "github", "jira", "todoist"
    Version() string     // Plugin version

    // Authentication
    Configure(config map[string]string) error
    ValidateConfig() error

    // Optional: Project discovery (not all plugins need to implement)
    // Returns ErrNotSupported if not implemented
    FetchProjects() ([]*Project, error)
    FetchProject(externalID string) (*Project, error)

    // Task operations - just fetch/push data, no sync logic
    // projectExternalID can be nil for systems without projects
    FetchTasks(projectExternalID *string, since *time.Time) ([]*Task, error)
    FetchTask(projectExternalID *string, taskExternalID string) (*Task, error)
    CreateTask(projectExternalID *string, task *Task) (*Task, error)
    UpdateTask(projectExternalID *string, taskExternalID string, task *Task) (*Task, error)

    // Comment operations
    FetchComments(projectExternalID *string, taskExternalID string) ([]*Comment, error)
    CreateComment(projectExternalID *string, taskExternalID string, comment *Comment) (*Comment, error)
}

// Standard error for unsupported operations
var ErrNotSupported = errors.New("operation not supported by this plugin")
```

**Sync Logic** (handled by core, not plugins):

1. Core calls plugin.FetchTasks(projectID, lastSyncTime)
2. Core compares fetched tasks with Todu API tasks by external_id
3. Core determines what needs to be created/updated based on timestamps
4. Core calls plugin.CreateTask() or plugin.UpdateTask() as needed
5. Core updates Todu API with changes
6. Core tracks last sync time in project metadata

**Benefits of this approach**:

- Plugins are simple - just data adapters
- Core has consistent sync logic across all systems
- Easier to add new plugins
- Easier to change sync strategy without updating plugins

**Next Steps**:
Would you like to:

- **A)** Resolve remaining questions and create full design document
- **B)** Start implementation with current design
- **C)** Create a detailed implementation plan with stages
- **D)** Other?

**Answer**: C - Created detailed implementation plan in `docs/plans/implementation-plan.md`

### Question 13: Plugin Distribution Strategy

**Asked**: Should plugins be in their own separate libraries/repositories?

**Options**:

**A) Monorepo - All plugins in main repository**:

```
todu/
├── plugins/
│   ├── github/
│   ├── jira/
│   └── todoist/
└── pkg/
    └── plugin/     # Public plugin interface
```

- All plugins compiled into single binary
- Versioned together
- Easier development and testing
- Simpler build process
- Pro: Fast iteration, consistent versions
- Con: Main repo grows with every plugin

**B) Separate repositories for each plugin**:

```
github.com/yourorg/todu                    # Main CLI
github.com/yourorg/todu-plugin-github      # Separate repo
github.com/yourorg/todu-plugin-jira        # Separate repo
```

- Each plugin is its own Go module
- Main CLI imports as dependencies
- Pro: Independent versioning, third-party plugins possible
- Con: More complex development, version coordination needed
- Note: Go doesn't support runtime plugin loading well

**C) Hybrid - Core plugins in main repo, extensible for external**:

```
todu/
├── plugins/           # Built-in core plugins
│   └── github/
└── pkg/
    └── plugin/       # Public interface for external plugins
```

- Core plugins (GitHub, Jira, etc.) built into binary
- Plugin interface in `pkg/` is public API
- External plugins could be added later (compile-time only)
- Pro: Simple to start, extensible later
- Con: External plugins require recompilation

**D) Plugin as separate binaries (exec model)**:

- Plugins are separate executables
- CLI calls plugin binaries via subprocess
- Pro: True runtime plugin system, any language
- Con: Much more complex, slower, deployment complexity

**Considerations for this project**:

- Starting with just GitHub plugin
- LLM is primary user (needs simple deployment)
- Single binary is easiest to distribute
- Plugin interface may evolve as we learn
- Go doesn't have good runtime plugin support

**Recommendation**: Start with **Option C (Hybrid)**

**Rationale**:

1. Keep plugins in main repo initially for faster development
2. Put plugin interface in `pkg/plugin/` (public API)
3. Single binary with all built-in plugins
4. Interface is designed to allow external plugins later if needed
5. Similar to how kubectl, terraform, and other Go CLIs work

**Implementation**:

```
todu/
├── cmd/todu/              # Main entry point
├── internal/
│   ├── api/               # Todu API client
│   ├── config/            # Config management
│   └── sync/              # Sync engine
├── pkg/
│   ├── plugin/            # Public plugin interface
│   │   └── interface.go
│   └── types/             # Shared types
└── plugins/               # Built-in plugins
    ├── github/
    ├── jira/             # Future
    └── todoist/          # Future
```

**Benefits**:

- Simple development workflow (one repo, one binary)
- Fast iteration on plugin interface
- Easy testing
- Can add external plugin support later if needed
- LLM gets single binary to install

**Future extensibility** (if needed):

- Third parties could create `todu-plugin-custom` packages
- Users would fork todu, add import, recompile
- Or: could add simple exec-based plugin system later

**Question**: Does this approach work for you?

**Answer**: C - Hybrid approach, but with separate Go modules for each plugin

**Clarification**: Plugins are separate Go modules (each with own `go.mod`) but in same repository (monorepo)

**Structure**:

```
todu/                              # Monorepo root
├── go.mod                         # Main module: github.com/yourorg/todu
├── go.sum
├── cmd/todu/                      # Main CLI
├── internal/
│   ├── api/
│   ├── config/
│   └── sync/
├── pkg/
│   ├── plugin/                    # Public plugin interface
│   │   └── interface.go
│   └── types/                     # Shared types
└── plugins/
    ├── github/
    │   ├── go.mod                 # Separate module: github.com/yourorg/todu/plugins/github
    │   ├── go.sum
    │   └── github.go
    ├── jira/
    │   ├── go.mod                 # Separate module
    │   └── jira.go
    └── todoist/
        ├── go.mod                 # Separate module
        └── todoist.go
```

**How it works**:

```go
// Main module go.mod
module github.com/yourorg/todu

// Plugin go.mod (plugins/github/go.mod)
module github.com/yourorg/todu/plugins/github

require github.com/yourorg/todu v0.1.0  // Import main module for interface
```

**Benefits**:

- Each plugin has independent dependencies
- Plugins can be versioned separately
- Clear module boundaries
- Main CLI imports plugins as separate modules
- Still in one repo for easier development

**Main CLI imports plugins**:

```go
// cmd/todu/main.go
import (
    "github.com/yourorg/todu/pkg/plugin"
    ghplugin "github.com/yourorg/todu/plugins/github"
    jiraplugin "github.com/yourorg/todu/plugins/jira"
)

func init() {
    registry.Register(ghplugin.New())
    registry.Register(jiraplugin.New())
}
```

**This is better than single module because**:

- Plugin dependencies don't pollute main CLI
- Can test plugins independently
- Can version plugins separately if needed
- Clear separation of concerns

**Future Evolution to True Plugin System**:

This structure supports evolution to a true plugin system post-MVP where:

- Third parties can develop plugins independently
- Plugins can be distributed separately
- Plugins loaded at runtime (not compile-time)

**Possible approaches for true runtime plugins**:

**Option 1: gRPC-based plugins (HashiCorp go-plugin pattern)**:

- Plugins are separate binaries that communicate via gRPC
- Core discovers plugins in `~/.todu/plugins/` directory
- Most robust, language-agnostic, production-ready
- Used by: Terraform, Vault, Packer

**Option 2: Go plugin package**:

- Plugins are `.so` files (shared objects)
- Loaded at runtime via `plugin.Open()`
- Limitations: same Go version, platform-specific, finicky
- Not recommended for production

**Option 3: WebAssembly (WASM)**:

- Plugins compiled to WASM
- Run in sandbox
- Language-agnostic
- Emerging technology, still maturing

**Option 4: Compile-time third-party plugins**:

- Third parties publish Go modules
- Users fork todu, add import, compile
- Simple but requires recompilation
- Good middle ground

**Recommendation for future**: gRPC-based (HashiCorp go-plugin)

**Design considerations for MVP to support future evolution**:

1. Keep plugin interface stable and well-documented
2. Keep interface in `pkg/plugin/` (public API)
3. Design interface with RPC in mind (serializable types, no channels/pointers to shared memory)
4. Version the plugin interface
5. Document plugin lifecycle (init, configure, execute)
6. Keep plugins stateless where possible

**Migration path**:

- **MVP (Phase 1)**: Built-in plugins, separate Go modules, single binary
- **Phase 2**: Plugin discovery system, still compile-time but easier to add
- **Phase 3**: True runtime plugins via gRPC (breaking change acceptable at this point)

### Question 14: Plugin Implementation Strategy - CLI vs Direct API

**Asked**: Should plugins use external CLIs (like `gh`, `jira`) or call APIs directly?

**Option A: Direct API calls** (current plan):

```go
// Plugin directly calls GitHub API
func (p *GitHubPlugin) FetchTasks(projectID *string, since *time.Time) ([]*Task, error) {
    // Use github.com/google/go-github library
    issues, _, err := p.client.Issues.ListByRepo(ctx, owner, repo, opts)
    // Convert issues to Task types
    return tasks, nil
}
```

**Pros**:

- Full API access - all features available
- Structured data (JSON) - no parsing needed
- Programmatic error handling
- No external dependencies (just Go libraries)
- Fast - direct HTTP calls
- Easy to test (mock HTTP responses)
- Consistent across all plugins

**Cons**:

- Need to write/maintain API client code for each system
- Handle authentication ourselves
- Need to keep up with API changes

**Option B: Shell out to external CLIs**:

```go
// Plugin calls gh CLI
func (p *GitHubPlugin) FetchTasks(projectID *string, since *time.Time) ([]*Task, error) {
    // Call: gh issue list --repo owner/repo --json number,title,state
    cmd := exec.Command("gh", "issue", "list", "--repo", *projectID, "--json", "number,title,state")
    output, err := cmd.Output()
    // Parse JSON output
    // Convert to Task types
    return tasks, nil
}
```

**Pros**:

- Leverage existing, well-maintained tools
- Authentication already handled by CLI (users already logged in)
- Don't write API clients
- CLI maintainers handle API changes
- Some users might already have these CLIs installed
- CLIs often have nice features (caching, rate limiting)

**Cons**:

- **Dependency**: Requires external binaries installed (gh, jira, etc.)
- **Fragile**: Parsing CLI output even with JSON (output format changes)
- **Limited**: CLI might not expose all API features we need
- **Slower**: Subprocess overhead for each call
- **Complex**: Different CLIs, different flags, different outputs
- **Testing**: Harder to test (need to mock subprocess calls)
- **Errors**: Error handling more complex (exit codes + stderr parsing)
- **Portability**: Need to find/validate CLI binaries on different platforms
- **Versioning**: Users might have different CLI versions with different behavior

**Comparison**:

| Aspect | Direct API | External CLI |
|--------|-----------|--------------|
| Dependencies | Go libraries only | External binaries required |
| Performance | Fast (direct HTTP) | Slower (subprocess overhead) |
| Data format | Native structs | Parse JSON/text output |
| API coverage | Full access | Limited to CLI features |
| Authentication | We handle | CLI handles |
| Testing | Easy (mock HTTP) | Harder (mock subprocess) |
| Maintenance | We maintain | CLI maintainers do |
| Reliability | Very reliable | More fragile |

**Real-world examples**:

- **kubectl**: Calls Kubernetes API directly, doesn't shell out
- **terraform**: Provider plugins call APIs directly
- **docker CLI**: Calls Docker daemon API directly
- **git GUI tools**: Some shell out to `git`, others use libgit2 (direct)

**User experience considerations**:

_With Direct API_:

```bash
# User setup
export TODU_GITHUB_TOKEN=ghp_xxx
todu project add --system github --external-id "owner/repo"
todu sync --project "owner/repo"  # Just works
```

_With External CLI_:

```bash
# User setup - must install gh first
brew install gh
gh auth login
export TODU_GITHUB_TOKEN=ghp_xxx  # Or rely on gh auth
todu project add --system github --external-id "owner/repo"
todu sync --project "owner/repo"  # Calls: gh issue list ...
# Error: gh not found in PATH (needs troubleshooting)
```

**Hybrid option?**:

- Try CLI first (if available), fallback to direct API
- But adds complexity, not recommended

**Recommendation**: **Option A - Direct API calls**

**Rationale**:

1. More reliable and predictable
2. Full API access
3. Easier to test and maintain
4. Faster performance
5. No external dependencies
6. Standard approach for this type of tool
7. Go has excellent HTTP/API client libraries

**Trade-offs accepted**:

- We write/maintain API client code (but Go libraries help)
- We handle authentication (but env vars are simple)
- We track API changes (but APIs are stable, documented)

**Question**: Do you agree with direct API approach, or do you see benefits in the CLI approach I'm missing?

**Answer**: A - Direct API calls. More reliable, faster, no external dependencies.

---

## Final Design Summary

### All Design Decisions

1. ✅ **Architecture**: CLI + Background Service + Plugins + Aggregation API
2. ✅ **Language**: Go (Golang)
3. ✅ **Plugin system**: Shared libraries with generic interface
4. ✅ **Plugin location**: Monorepo with separate Go modules (MVP), evolving to true runtime plugins (future)
5. ✅ **Plugin implementation**: Direct API calls (not external CLIs)
6. ✅ **Plugin interface design**: RPC-friendly for future gRPC-based runtime plugins
7. ✅ **Sync strategy**: Core handles sync logic, plugins are data adapters using `external_id`
8. ✅ **CLI commands**: Complete coverage of all operations (systems, projects, tasks, sync, daemon, config)
9. ✅ **Output format**: Text (default) with `--format json` option
10. ✅ **Plugin interface**: Nullable `projectExternalID` for systems without projects
11. ✅ **FetchProjects()**: Optional - plugins return `ErrNotSupported` if not applicable
12. ✅ **Daemon**: Single binary with daemon mode + CLI-assisted service management
13. ✅ **Configuration**: Environment variables for secrets + YAML config file for settings

### Project Structure

```
todu/                      # Monorepo root
├── go.mod                 # Main module
├── cmd/todu/              # Main CLI entry point
├── internal/
│   ├── api/               # Todu API client
│   ├── config/            # Configuration management
│   ├── sync/              # Sync engine (core logic)
│   └── registry/          # Plugin registry
├── pkg/
│   ├── plugin/            # Public plugin interface
│   │   └── interface.go
│   └── types/             # Shared types (Task, Project, etc.)
└── plugins/               # Built-in plugins (each is separate Go module)
    ├── github/
    │   ├── go.mod         # Separate module
    │   └── github.go      # Direct GitHub API calls
    ├── jira/              # Future
    │   └── go.mod
    └── todoist/           # Future
        └── go.mod
```

### Ready to Implement

All major design decisions are complete. Implementation plan is in `docs/plans/implementation-plan.md`.

---

## Notes

_This section will be updated as we discuss..._
