# Todu CLI

A command-line tool for synchronizing tasks across multiple systems (GitHub Issues, Jira, Todoist, etc.) with a central Todu API.

## Features

- **Multi-System Sync**: Sync tasks between GitHub, Forgejo, Todoist, and more
- **Bidirectional Sync**: Push and pull changes between systems
- **Plugin Architecture**: Easy to add new task management systems
- **Background Daemon**: Automatic sync with configurable intervals
- **Task Management**: Create, update, and manage tasks via CLI
- **Comment Sync**: Synchronize comments across systems
- **Conflict Resolution**: Last-write-wins conflict resolution
- **LLM-Friendly**: Designed for use by AI assistants

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/evcraddock/todu.sh
cd todu.sh

# Build the CLI
go build -o todu ./cmd/todu

# (Optional) Install to your PATH
sudo mv todu /usr/local/bin/
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

2. **Register a System** (e.g., GitHub):

```bash
# Add GitHub as a system
todu system add github

# Configure GitHub token
export TODU_GITHUB_TOKEN="ghp_your_token_here"
```

3. **Link a Project**:

```bash
# Link a GitHub repository
todu project add --system github --external-id "owner/repo" --name "My Project"
```

4. **Run Your First Sync**:

```bash
# Preview what would sync
todu sync --all --dry-run

# Perform the actual sync
todu sync --all
```

5. **Set Up Background Sync** (Optional):

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

# Discover available repositories from GitHub
todu project discover --system github

# Add a specific project
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

```
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
├──────────────┬──────────────┬──────────────────────────┤
│    GitHub    │   Forgejo    │   Future Plugins...      │
│    Plugin    │   Plugin     │   (Jira, Todoist, etc.)  │
└──────────────┴──────────────┴──────────────────────────┘
        │              │                   │
        ▼              ▼                   ▼
   GitHub API     Forgejo API        Other APIs
```

## Plugin System

Todu uses a plugin architecture to support multiple task management systems. Each plugin implements a common interface for fetching and updating tasks.

**Currently Available Plugins:**

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
