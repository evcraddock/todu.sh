# Configuration Reference

Complete reference for configuring Todu CLI.

## Configuration File

### Location

Todu looks for configuration in the following locations (in order):

1. `~/.todu/config.yaml`
2. `~/.config/todu/config.yaml`
3. `./config.yaml` (current directory)

### Format

Configuration file is in YAML format:

```yaml
# Todu API endpoint (required)
api_url: "http://localhost:8000"

# Daemon configuration
daemon:
  interval: "5m"      # Sync interval (duration format)
  projects: []        # Project IDs to sync (empty = all)

# Output configuration
output:
  format: "text"      # Output format: text or json
  color: true         # Enable color output
```

## Configuration Options

### api_url

**Type**: String (URL)
**Required**: Yes
**Default**: `http://localhost:8000`

The base URL of the Todu API server.

```yaml
api_url: "https://todu-api.example.com"
```

### daemon.interval

**Type**: String (duration)
**Required**: No
**Default**: `5m`

How often the daemon should run sync operations. Format is Go duration string:

- `30s` - 30 seconds
- `1m` - 1 minute
- `5m` - 5 minutes
- `1h` - 1 hour
- `30m` - 30 minutes

```yaml
daemon:
  interval: "2m"  # Sync every 2 minutes
```

### daemon.projects

**Type**: Array of integers
**Required**: No
**Default**: `[]` (empty, meaning all projects)

List of specific project IDs to sync. If empty, all projects are synced.

```yaml
daemon:
  projects: [1, 3, 5]  # Only sync projects 1, 3, and 5
```

### output.format

**Type**: String
**Required**: No
**Default**: `text`
**Options**: `text`, `json`

Default output format for commands.

```yaml
output:
  format: "json"  # Output JSON by default
```

Can be overridden per-command with `--format` flag.

### output.color

**Type**: Boolean
**Required**: No
**Default**: `true`

Enable or disable color output in terminal.

```yaml
output:
  color: false  # Disable colors
```

## Environment Variables

Environment variables override configuration file values.

### General Configuration

| Variable | Config Equivalent | Description |
|----------|------------------|-------------|
| `TODU_API_URL` | `api_url` | Todu API endpoint |
| `TODU_DAEMON_INTERVAL` | `daemon.interval` | Daemon sync interval |
| `TODU_OUTPUT_FORMAT` | `output.format` | Output format |
| `TODU_OUTPUT_COLOR` | `output.color` | Enable color output |

Example:

```bash
export TODU_API_URL="https://api.example.com"
export TODU_DAEMON_INTERVAL="10m"
```

### Plugin Configuration

Each plugin uses environment variables for authentication and configuration:

#### GitHub Plugin

| Variable | Description | Required |
|----------|-------------|----------|
| `TODU_GITHUB_TOKEN` | GitHub Personal Access Token | Yes |
| `TODU_GITHUB_URL` | GitHub API URL | No (defaults to `https://api.github.com`) |

```bash
export TODU_GITHUB_TOKEN="ghp_your_token_here"
export TODU_GITHUB_URL="https://api.github.com"
```

#### Forgejo Plugin

| Variable | Description | Required |
|----------|-------------|----------|
| `TODU_FORGEJO_TOKEN` | Forgejo API token | Yes |
| `TODU_FORGEJO_URL` | Forgejo instance URL | Yes |

```bash
export TODU_FORGEJO_TOKEN="your_token_here"
export TODU_FORGEJO_URL="https://git.example.com"
```

## Configuration Precedence

Configuration values are resolved in this order (highest to lowest priority):

1. **Command-line flags** (e.g., `--format json`)
2. **Environment variables** (e.g., `TODU_API_URL`)
3. **Configuration file** (`config.yaml`)
4. **Default values**

Example:

```bash
# config.yaml has: api_url: "http://localhost:8000"
export TODU_API_URL="http://staging.example.com"
# API URL will be "http://staging.example.com" (env var wins)
```

## Managing Configuration

### View Current Configuration

```bash
todu config show
```

Output:

```
api_url: http://localhost:8000
daemon.interval: 5m
daemon.projects: []
output.format: text
output.color: true
```

### Get Specific Value

```bash
todu config get api_url
```

### Set Configuration Value

```bash
# Set API URL
todu config set api_url "https://api.example.com"

# Set daemon interval
todu config set daemon.interval "10m"

# Set output format
todu config set output.format "json"
```

### Validate Configuration

```bash
todu config validate
```

Checks:

- Configuration file syntax
- Required values are present
- URLs are valid
- Duration formats are correct

## Example Configurations

### Development Setup

```yaml
# ~/.config/todu/config.yaml
api_url: "http://localhost:8000"
daemon:
  interval: "1m"  # Faster for development
  projects: []
output:
  format: "text"
  color: true
```

### Production Setup

```yaml
# ~/.config/todu/config.yaml
api_url: "https://todu-api.example.com"
daemon:
  interval: "5m"
  projects: []  # Sync all
output:
  format: "json"  # For log parsing
  color: false
```

### Selective Sync

```yaml
# ~/.config/todu/config.yaml
api_url: "http://localhost:8000"
daemon:
  interval: "5m"
  projects: [1, 2, 3]  # Only critical projects
output:
  format: "text"
  color: true
```

### Multiple Environments

Use environment variables to switch between environments:

```bash
# Development
export TODU_API_URL="http://localhost:8000"
todu sync --all

# Staging
export TODU_API_URL="https://staging-api.example.com"
todu sync --all

# Production
export TODU_API_URL="https://api.example.com"
todu sync --all
```

## Platform-Specific Notes

### macOS

Configuration is typically stored in:

- `~/.config/todu/config.yaml`
- Daemon logs: `~/.todu/daemon.log`
- Daemon service: `~/Library/LaunchAgents/com.todu.daemon.plist`

### Linux

Configuration is typically stored in:

- `~/.config/todu/config.yaml`
- Daemon logs: `~/.todu/daemon.log`
- Daemon service: `~/.config/systemd/user/todu.service`

### Windows

Configuration is typically stored in:

- `%USERPROFILE%\.config\todu\config.yaml`
- Daemon service management not yet implemented

## Security Considerations

### Token Storage

**Never** store API tokens in the configuration file. Always use environment variables:

```bash
# Good: Environment variable
export TODU_GITHUB_TOKEN="ghp_..."

# Bad: Don't do this
# api_tokens:
#   github: "ghp_..."  # DON'T STORE TOKENS IN CONFIG!
```

### File Permissions

Ensure configuration file has appropriate permissions:

```bash
chmod 600 ~/.config/todu/config.yaml
```

### Token Rotation

When rotating tokens:

1. Generate new token in external system
2. Update environment variable
3. Test with `todu sync --dry-run`
4. Revoke old token

## Troubleshooting

### Configuration Not Loading

```bash
# Check which config file is being used
todu config show

# Verify file exists and is readable
ls -la ~/.config/todu/config.yaml
cat ~/.config/todu/config.yaml

# Check for YAML syntax errors
todu config validate
```

### Environment Variables Not Working

```bash
# Verify variable is set
env | grep TODU

# Check spelling (case-sensitive)
echo $TODU_API_URL

# Ensure it's exported
export TODU_API_URL="http://localhost:8000"
```

### Plugin Configuration Issues

```bash
# Check plugin is registered
todu system list

# Verify token environment variable
echo $TODU_GITHUB_TOKEN

# Test plugin configuration
todu system config github
```
