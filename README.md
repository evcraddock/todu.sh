# Todu CLI

A command-line tool for synchronizing tasks across multiple systems
(GitHub Issues, Jira, Todoist, etc.) with a central Todu API.

## Features

- **Multi-System Sync**: Sync tasks between GitHub, Forgejo, Todoist, and more
- **Local-Only Projects**: Create projects that exist only in Todu without
  external sync
- **Bidirectional Sync**: Push and pull changes between systems
- **Plugin Architecture**: Easy to add new task management systems
- **Background Daemon**: Automatic sync with configurable intervals
- **Task Management**: Create, update, and manage tasks via CLI
- **Recurring Templates**: Define recurring tasks and habits with RRULE patterns
- **Comment Sync**: Synchronize comments across systems
- **Conflict Resolution**: Last-write-wins conflict resolution
- **LLM-Friendly**: Designed for use by AI assistants

## Quick Start

### Installation

#### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/evcraddock/todu.sh/main/install.sh | sh
```

This automatically detects your platform and uses `go install` if Go is
available (no sudo required), otherwise downloads the pre-built binary.

#### Using Go

```bash
go install github.com/evcraddock/todu.sh/cmd/todu@latest
```

#### Manual Download

Download from [Releases](https://github.com/evcraddock/todu.sh/releases/latest):

| Platform | File |
|----------|------|
| Linux (amd64) | `todu_VERSION_linux_amd64.tar.gz` |
| Linux (arm64) | `todu_VERSION_linux_arm64.tar.gz` |
| macOS (Intel) | `todu_VERSION_darwin_amd64.tar.gz` |
| macOS (Apple Silicon) | `todu_VERSION_darwin_arm64.tar.gz` |

```bash
tar -xzf todu_*.tar.gz
sudo mv todu /usr/local/bin/
```

#### Build from Source

```bash
git clone https://github.com/evcraddock/todu.sh
cd todu.sh
make build
sudo mv .build/todu /usr/local/bin/
```

#### Verify Installation

```bash
todu version
```

### Initial Setup

1. **Configure the API URL**:

```bash
# Set the Todu API URL in config
mkdir -p ~/.config/todu
cat > ~/.config/todu/config.yaml <<EOF
api_url: "http://localhost:8000"
EOF
```

1. **Register a System** (e.g., GitHub):

```bash
# Add GitHub as a system
todu system add github

# Configure GitHub token
export TODU_GITHUB_TOKEN="ghp_your_token_here"
```

1. **Link a Project**:

```bash
# Link a GitHub repository
todu project add --system github --external-id "owner/repo" --name "My Project"
```

1. **Run Your First Sync**:

```bash
# Preview what would sync
todu sync --all --dry-run

# Perform the actual sync
todu sync --all
```

1. **Set Up Background Sync** (Optional):

```bash
# Install daemon to sync automatically every 5 minutes
todu daemon install

# Check daemon status
todu daemon status
```

## Usage Examples

### Managing Systems

```bash
# List registered systems
todu system list

# Show system configuration requirements
todu system config github

# Remove a system
todu system remove github
```

### Managing Projects

```bash
# List all projects
todu project list

# Create a local-only project (no external sync)
todu project add --name "My Local Tasks"

# Create a local project with explicit system
todu project add --system local --name "Personal Project"

# Discover available repositories from GitHub
todu project discover --system github

# Add a specific project from external system
todu project add --system github --external-id "octocat/Hello-World" --name "Hello World"

# Show project details
todu project show 1

# Remove a project
todu project remove 1
```

### Syncing Tasks

```bash
# Sync all projects
todu sync --all

# Sync a specific project
todu sync --project "My Project"

# Sync all projects for a system
todu sync --system github

# Preview changes without syncing
todu sync --all --dry-run

# Override sync strategy
todu sync --all --strategy pull    # Pull only
todu sync --all --strategy push    # Push only
```

### Managing Tasks

```bash
# List tasks
todu task list

# Filter tasks
todu task list --status active
todu task list --priority high
todu task list --project "My Project"
todu task list --search "bug"

# Show task details with comments
todu task show 123

# Create a new task
todu task create --title "Fix bug" --project "My Project" --priority high

# Update a task
todu task update 123 --status done
todu task update 123 --add-label "bug" --add-label "urgent"

# Close a task
todu task close 123

# Add a comment
todu task comment 123 "This is fixed in PR #456"

# Delete a task
todu task delete 123
```

### Recurring Task Templates

Recurring task templates allow you to define tasks that repeat on a schedule
using [RRULE][rrule] recurrence patterns.

[rrule]: https://icalendar.org/iCalendar-RFC-5545/3-8-5-3-recurrence-rule.html

```bash
# List all templates
todu template list

# List only active templates
todu template list --active

# Filter by type (task or habit)
todu template list --type habit

# Show template details with upcoming occurrences
todu template show 1

# Create a daily task template
todu template create --project "My Project" --title "Daily standup" \
  --recurrence "FREQ=DAILY" --start-date "2024-01-01" --timezone "America/Chicago"

# Create a weekly habit on specific days
todu template create --project "Personal" --title "Exercise" \
  --type habit --recurrence "FREQ=WEEKLY;BYDAY=MO,WE,FR" \
  --start-date "2024-01-01" --timezone "America/New_York"

# Create a monthly task
todu template create --project "Work" --title "Monthly report" \
  --recurrence "FREQ=MONTHLY;BYMONTHDAY=1" \
  --start-date "2024-01-01" --timezone "UTC"

# Update a template
todu template update 1 --title "Updated title"
todu template update 1 --recurrence "FREQ=WEEKLY"

# Activate/deactivate a template
todu template activate 1
todu template deactivate 1

# Delete a template
todu template delete 1
```

**Template Types:**

- **task**: Regular recurring tasks with deadlines
  (e.g., weekly reports, monthly reviews)
- **habit**: Streak-based activities for habit tracking
  (e.g., daily exercise, meditation)

**Common RRULE Patterns:**

| Pattern                              | Description                  |
| ------------------------------------ | ---------------------------- |
| `FREQ=DAILY`                         | Every day                    |
| `FREQ=WEEKLY`                        | Every week                   |
| `FREQ=WEEKLY;BYDAY=MO,WE,FR`         | Monday, Wednesday, Friday    |
| `FREQ=MONTHLY;BYMONTHDAY=1`          | First of every month         |
| `FREQ=MONTHLY;BYDAY=1MO`             | First Monday of every month  |
| `FREQ=YEARLY;BYMONTH=1;BYMONTHDAY=1` | January 1st every year       |
| `FREQ=DAILY;INTERVAL=2`              | Every other day              |

**Timezones:**

Templates use [IANA timezone names][tz] (e.g., `America/New_York`,
`Europe/London`, `Asia/Tokyo`). Common shortcuts like `EST`, `CST`, `PST`
are also supported.

[tz]: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

### Daemon Management

```bash
# Run daemon in foreground (for testing)
todu daemon start

# Install as background service
todu daemon install --interval 5m

# Check daemon status
todu daemon status

# View daemon logs
todu daemon logs
todu daemon logs --follow

# Stop the daemon
todu daemon stop

# Restart the daemon
todu daemon restart

# Uninstall the daemon
todu daemon uninstall
```

## Configuration

Configuration file: `~/.config/todu/config.yaml`

```yaml
# Todu API endpoint
api_url: "http://localhost:8000"

# Daemon settings
daemon:
  interval: "5m"      # Sync interval (1m, 5m, 1h, etc.)
  projects: []        # Empty = all projects, or [1, 2, 3] for specific ones

# Output settings
output:
  format: "text"      # text or json
  color: true         # Enable color output
```

### Environment Variables

Plugin configuration is done via environment variables:

```bash
# GitHub plugin
export TODU_GITHUB_TOKEN="ghp_your_token_here"
export TODU_GITHUB_URL="https://api.github.com"  # Optional, defaults to GitHub.com

# Future plugins would follow similar pattern
# export TODU_JIRA_TOKEN="..."
# export TODU_TODOIST_TOKEN="..."
```

## Architecture

```text
┌─────────────────────────────────────────────────────────┐
│                        Todu CLI                         │
├─────────────────────────────────────────────────────────┤
│  Commands: system, project, task, sync, daemon         │
├─────────────────────────────────────────────────────────┤
│                     Sync Engine                         │
│  - Bidirectional sync                                   │
│  - Conflict resolution                                  │
│  - Comment synchronization                              │
├─────────────────────────────────────────────────────────┤
│                   Plugin System                         │
├──────────┬───────────┬───────────┬─────────────────────┤
│  Local   │  GitHub   │  Forgejo  │  Future Plugins...  │
│  Plugin  │  Plugin   │  Plugin   │  (Jira, Todoist)    │
└──────────┴───────────┴───────────┴─────────────────────┘
     │           │            │               │
     ▼           ▼            ▼               ▼
  (no-op)   GitHub API   Forgejo API     Other APIs
```

## Plugin System

Todu uses a plugin architecture to support multiple task management systems.
Each plugin implements a common interface for fetching and updating tasks.

**Currently Available Plugins:**

- **Local**: Local-only projects with no external sync
  (auto-registered on first use)
- **GitHub**: Sync with GitHub Issues
- **Forgejo**: Sync with Forgejo/Gitea Issues

**Coming Soon:**

- Jira
- Todoist
- Linear
- And more...

See [docs/plugin-development.md](docs/plugin-development.md) for creating new plugins.

## Documentation

- [Installation Guide](docs/installation.md)
- [Configuration Reference](docs/configuration.md)
- [Common Workflows](docs/workflows.md)
- [Plugin Documentation](docs/plugins.md)
- [CLI Reference](docs/cli-reference.md)
- [Plugin Development](docs/plugin-development.md)

## Platform Support

- **macOS**: Full support with launchd integration
- **Linux**: Full support with systemd integration
- **Windows**: Command-line support (service management not yet implemented)

## Requirements

- Go 1.21 or later
- Access to a Todu API instance
- API tokens for external systems you want to sync

## Development

```bash
# Run tests
go test ./...

# Run specific package tests
go test ./internal/sync -v

# Build
go build -o todu ./cmd/todu

# Run without building
go run ./cmd/todu [command]
```

## Contributing

Contributions are welcome! Areas where help is needed:

- New plugin implementations
- Documentation improvements
- Bug fixes
- Feature requests

## License

[Add your license here]

## Support

For issues, questions, or feature requests, please open an issue on GitHub.
