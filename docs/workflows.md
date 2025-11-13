# Common Workflows

This guide demonstrates common workflows for using Todu CLI.

## Table of Contents

- [Initial Setup](#initial-setup)
- [Syncing Tasks](#syncing-tasks)
- [Managing Tasks](#managing-tasks)
- [Working with Projects](#working-with-projects)
- [Setting Up the Daemon](#setting-up-the-daemon)
- [Troubleshooting](#troubleshooting)

## Initial Setup

### First-Time Configuration

1. **Create configuration file**:

```bash
mkdir -p ~/.config/todu
cat > ~/.config/todu/config.yaml <<EOF
api_url: "http://localhost:8000"
daemon:
  interval: "5m"
  projects: []
output:
  format: "text"
  color: true
EOF
```

2. **Verify configuration**:

```bash
todu config show
```

### Adding your first system (GitHub)

1. **Create a GitHub Personal Access Token**:
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Generate new token with `repo` scope
   - Copy the token (starts with `ghp_`)

2. **Configure the GitHub token**:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export TODU_GITHUB_TOKEN="ghp_your_token_here"

# Or set it temporarily
export TODU_GITHUB_TOKEN="ghp_your_token_here"
```

3. **Register GitHub system**:

```bash
todu system add github
```

4. **Verify system is registered**:

```bash
todu system list
```

### Linking Your First Project

1. **Discover available repositories**:

```bash
todu project discover --system github
```

2. **Link a specific repository**:

```bash
todu project add \
  --system github \
  --external-id "octocat/Hello-World" \
  --name "Hello World"
```

3. **Verify project is linked**:

```bash
todu project list
```

### Running Your First Sync

1. **Preview what would sync** (dry run):

```bash
todu sync --all --dry-run
```

2. **Perform the actual sync**:

```bash
todu sync --all
```

3. **View synced tasks**:

```bash
todu task list
```

## Syncing Tasks

### Sync All Projects

```bash
# Full sync of all projects
todu sync --all

# Preview changes without applying
todu sync --all --dry-run
```

### Sync Specific Project

```bash
# By project name
todu sync --project "My Project"

# By project ID
todu sync --project 1
```

### Sync by System

Sync all projects for a specific system:

```bash
todu sync --system github
```

### Override Sync Strategy

```bash
# Pull only (external → Todu)
todu sync --all --strategy pull

# Push only (Todu → external)
todu sync --all --strategy push

# Bidirectional (default)
todu sync --all --strategy bidirectional
```

### Check Sync Status

```bash
todu sync status

# Filter by system
todu sync status --system github
```

## Managing Tasks

### Listing Tasks

```bash
# List all tasks
todu task list

# Filter by status
todu task list --status active
todu task list --status done

# Filter by priority
todu task list --priority high

# Filter by project
todu task list --project "My Project"

# Search by keyword
todu task list --search "bug"

# Combine filters
todu task list --status active --priority high --project "Backend"

# Limit results
todu task list --limit 10
```

### Viewing Task Details

```bash
# Show task with comments
todu task show 123

# Output as JSON
todu task show 123 --format json
```

### Creating Tasks

```bash
# Create a basic task
todu task create \
  --title "Fix login bug" \
  --project "Backend"

# Create a detailed task
todu task create \
  --title "Implement OAuth" \
  --description "Add OAuth 2.0 authentication" \
  --project "Backend" \
  --priority high \
  --label "feature" \
  --label "security" \
  --assignee "john" \
  --due "2025-12-31"
```

### Updating Tasks

```bash
# Update title
todu task update 123 --title "Fix critical login bug"

# Change status
todu task update 123 --status done

# Update priority
todu task update 123 --priority high

# Add labels
todu task update 123 --add-label "urgent" --add-label "security"

# Remove labels
todu task update 123 --remove-label "low-priority"

# Add assignees
todu task update 123 --add-assignee "jane"

# Update due date
todu task update 123 --due "2025-12-25"
```

### Closing Tasks

```bash
# Mark task as done
todu task close 123

# Equivalent to:
todu task update 123 --status done
```

### Adding Comments

```bash
# Add a comment
todu task comment 123 "Fixed in PR #456"

# Add comment with flag
todu task comment 123 --message "Deployed to production"

# Specify author
todu task comment 123 "Reviewed and approved" --author "reviewer"
```

### Deleting Tasks

```bash
# Delete a task (with confirmation)
todu task delete 123

# Delete without confirmation
todu task delete 123 --force
```

## Working with Projects

### Listing Projects

```bash
# List all projects
todu project list

# Filter by system
todu project list --system github

# Output as JSON
todu project list --format json
```

### Discovering Projects

Find available projects from external systems:

```bash
# Discover GitHub repositories
todu project discover --system github

# This shows all repos you have access to
# Use the output to decide which to link
```

### Adding Projects

```bash
# Add a GitHub repository
todu project add \
  --system github \
  --external-id "owner/repository" \
  --name "My Repository"

# Add with sync strategy
todu project add \
  --system github \
  --external-id "owner/repo" \
  --name "Backend API" \
  --strategy bidirectional
```

### Viewing Project Details

```bash
# Show project info
todu project show 1

# Show by name
todu project show "My Project"
```

### Updating Projects

```bash
# Update project name
todu project update 1 --name "New Name"

# Change sync strategy
todu project update 1 --strategy pull
```

### Removing Projects

```bash
# Remove a project
todu project remove 1

# Remove by name
todu project remove "Old Project"
```

## Setting Up the Daemon

### Running Daemon in Foreground

For testing or development:

```bash
# Start daemon (runs until Ctrl+C)
todu daemon start
```

### Installing as a Service

For production use:

```bash
# Install with default settings (5m interval)
todu daemon install

# Install with custom interval
todu daemon install --interval 2m

# Install to sync specific projects only
todu daemon install --projects 1,2,3
```

### Managing the Daemon Service

```bash
# Check daemon status
todu daemon status

# Start the daemon
todu daemon start

# Stop the daemon
todu daemon stop

# Restart the daemon
todu daemon restart
```

### Viewing Daemon Logs

```bash
# Show recent logs
todu daemon logs

# Show more lines
todu daemon logs --lines 100

# Follow logs in real-time
todu daemon logs --follow

# Stop following with Ctrl+C
```

### Uninstalling the Daemon

```bash
# Uninstall the service
todu daemon uninstall
```

## Troubleshooting

### Sync Issues

**Problem**: Sync fails with "plugin not configured"

```bash
# Check which plugins are available
todu system list

# Verify environment variables are set
echo $TODU_GITHUB_TOKEN

# Re-export if needed
export TODU_GITHUB_TOKEN="ghp_your_token_here"
```

**Problem**: Sync reports errors but doesn't show details

```bash
# Run sync manually to see full output
todu sync --all

# Check daemon logs
todu daemon logs
```

**Problem**: Tasks aren't syncing

```bash
# Check project configuration
todu project list
todu project show 1

# Verify project has correct external_id
# Run a test sync with dry-run
todu sync --project 1 --dry-run
```

### Authentication Issues

**Problem**: "401 Unauthorized" errors

```bash
# Verify token is set correctly
echo $TODU_GITHUB_TOKEN

# Test with system config
todu system config github

# Check token hasn't expired (regenerate if needed)
```

### Task Issues

**Problem**: Can't find task by name

```bash
# Tasks are identified by ID, not name
# First list to find the ID
todu task list --search "task name"

# Then use the ID
todu task show 123
```

**Problem**: Task updates aren't reflected in external system

```bash
# Check sync strategy
todu project show 1

# If strategy is "pull", changes won't push
# Change to "bidirectional" or "push"
todu project update 1 --strategy bidirectional

# Then sync
todu sync --project 1
```

### Daemon Issues

**Problem**: Daemon not syncing

```bash
# Check daemon status
todu daemon status

# View logs for errors
todu daemon logs

# Restart daemon
todu daemon restart
```

**Problem**: Daemon using old configuration

```bash
# Configuration changes require restart
todu daemon restart
```

### API Connection Issues

**Problem**: "Connection refused" errors

```bash
# Verify API is running
curl http://localhost:8000/health

# Check API URL in config
todu config show

# Update if needed
todu config set api_url "http://correct-url:8000"
```

## Advanced Workflows

### Multi-System Setup

Managing multiple external systems:

```bash
# Add GitHub
export TODU_GITHUB_TOKEN="ghp_..."
todu system add github

# Add Forgejo
export TODU_FORGEJO_TOKEN="token_..."
export TODU_FORGEJO_URL="https://git.example.com"
todu system add forgejo

# Link projects from both
todu project add --system github --external-id "org/repo" --name "GitHub Repo"
todu project add --system forgejo --external-id "user/repo" --name "Forgejo Repo"

# Sync both
todu sync --all
```

### Selective Syncing

Sync only specific projects automatically:

```bash
# Configure daemon to sync only certain projects
cat > ~/.config/todu/config.yaml <<EOF
api_url: "http://localhost:8000"
daemon:
  interval: "5m"
  projects: [1, 3, 5]  # Only sync these project IDs
EOF

# Restart daemon to apply
todu daemon restart
```

### Bulk Operations

```bash
# Close all active tasks for a project
todu task list --project "Old Project" --status active --format json | \
  jq -r '.[].id' | \
  xargs -I {} todu task close {}

# Add label to multiple tasks
todu task list --search "bug" --format json | \
  jq -r '.[].id' | \
  xargs -I {} todu task update {} --add-label "needs-review"
```

### Integration with Scripts

```bash
#!/bin/bash
# Daily sync and report script

# Sync all projects
todu sync --all

# Generate report of active tasks
echo "Active Tasks Report - $(date)"
echo "================================"
todu task list --status active --format json | \
  jq -r '.[] | "\(.id): \(.title) [\(.project_id)]"'

# Check for high priority tasks
HIGH_COUNT=$(todu task list --priority high --format json | jq 'length')
echo ""
echo "High priority tasks: $HIGH_COUNT"
```

## Tips and Best Practices

1. **Use Dry Run First**: Always test sync with `--dry-run` before applying changes
2. **Start with Pull**: Begin with pull-only strategy to avoid accidental changes
3. **Regular Syncs**: Set daemon interval based on your needs (2-10 minutes typical)
4. **Monitor Logs**: Occasionally check daemon logs for issues
5. **Consistent Naming**: Use clear project names that match your workflow
6. **Environment Variables**: Set tokens in shell profile for persistence
7. **Backup Config**: Keep a copy of your config file
8. **JSON Output**: Use `--format json` for scripting and automation
