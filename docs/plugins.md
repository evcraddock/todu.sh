# Plugin Documentation

Guide to available plugins and their configuration.

## Available Plugins

### GitHub Plugin

Sync tasks with GitHub Issues.

#### Configuration

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `TODU_GITHUB_TOKEN` | Personal Access Token | Yes | - |
| `TODU_GITHUB_URL` | API endpoint | No | `https://api.github.com` |

#### Setup

1. **Create Personal Access Token**:
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Click "Generate new token (classic)"
   - Select scopes: `repo` (full control of private repositories)
   - Copy the token (starts with `ghp_`)

2. **Configure token**:

```bash
export TODU_GITHUB_TOKEN="ghp_your_token_here"
```

3. **Register system**:

```bash
todu system add github
```

4. **Link repository**:

```bash
todu project add --system github --external-id "owner/repo" --name "My Repo"
```

#### Type Mappings

**Repository → Project:**

- `external_id`: "owner/repo" (e.g., "octocat/Hello-World")
- `name`: Repository name
- `description`: Repository description

**Issue → Task:**

- `external_id`: Issue number as string (e.g., "123")
- `title`: Issue title
- `description`: Issue body
- `status`: "open" → "active", "closed" → "done"
- `priority`: From labels (priority:high, priority:medium, priority:low)
- `labels`: Issue labels (excluding priority labels)
- `assignees`: Issue assignees
- `source_url`: Issue HTML URL
- `due_date`: Milestone due date (if set)

**Comment → Comment:**

- `content`: Comment body
- `author`: Comment author login
- `created_at`/`updated_at`: Comment timestamps

#### Supported Operations

- ✅ Fetch repositories
- ✅ Fetch issues
- ✅ Create issues
- ✅ Update issues (title, body, state, labels, assignees)
- ✅ Fetch comments
- ✅ Create comments
- ❌ Delete issues (not supported by GitHub API)
- ❌ Update comments (GitHub comments are immutable)

#### Notes

- **Rate Limiting**: GitHub has rate limits (5000 requests/hour for authenticated users)
- **Draft PRs**: Not synced as tasks
- **Pull Requests**: Not currently synced (issues only)
- **Projects**: GitHub Projects are not synced (only Issues)
- **Milestones**: Milestone due dates map to task due dates

### Forgejo Plugin

Sync tasks with Forgejo/Gitea Issues.

#### Configuration

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `TODU_FORGEJO_TOKEN` | API Token | Yes | - |
| `TODU_FORGEJO_URL` | Forgejo instance URL | Yes | - |

#### Setup

1. **Create API Token**:
   - Go to Forgejo Settings → Applications → Generate New Token
   - Select scopes: `repo` or `write:issue`
   - Copy the token

2. **Configure plugin**:

```bash
export TODU_FORGEJO_TOKEN="your_token_here"
export TODU_FORGEJO_URL="https://git.example.com"
```

3. **Register system**:

```bash
todu system add forgejo
```

4. **Link repository**:

```bash
todu project add --system forgejo --external-id "owner/repo" --name "My Repo"
```

#### Type Mappings

Same as GitHub plugin (Forgejo/Gitea use GitHub-compatible API).

#### Supported Operations

- ✅ Fetch repositories
- ✅ Fetch issues
- ✅ Create issues
- ✅ Update issues
- ✅ Fetch comments
- ✅ Create comments

#### Notes

- **Compatible**: Works with both Forgejo and Gitea
- **Self-Hosted**: Requires `TODU_FORGEJO_URL` to point to your instance
- **API Compatibility**: Uses Gitea-compatible API

## Plugin System Architecture

### Plugin Interface

All plugins implement the same interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Configure(config map[string]string) error
    ValidateConfig() error
    FetchProjects() ([]*types.Project, error)
    FetchTasks(projectExternalID *string, since *time.Time) ([]*types.Task, error)
    FetchTask(projectExternalID, taskExternalID string) (*types.Task, error)
    CreateTask(projectExternalID *string, task *types.Task) (*types.Task, error)
    UpdateTask(projectExternalID, taskExternalID string, task *types.Task) (*types.Task, error)
    FetchComments(projectExternalID, taskExternalID string) ([]*types.Comment, error)
    CreateComment(projectExternalID, taskExternalID string, comment *types.Comment) (*types.Comment, error)
}
```

### Configuration Loading

Plugins are configured via environment variables with the pattern:
`TODU_<PLUGIN>_<SETTING>`

Examples:

- `TODU_GITHUB_TOKEN`
- `TODU_FORGEJO_URL`
- `TODU_JIRA_USERNAME`

### Error Handling

Plugins return standard errors:

- `plugin.ErrNotSupported` - Operation not supported by this plugin
- `plugin.ErrNotConfigured` - Plugin not properly configured
- Standard Go errors for other failures

## Coming Soon

### Jira Plugin (Planned)

Sync tasks with Atlassian Jira.

**Configuration**:

- `TODU_JIRA_URL` - Jira instance URL
- `TODU_JIRA_USERNAME` - Username
- `TODU_JIRA_API_TOKEN` - API token

**Mappings**:

- Project → Jira Project
- Task → Jira Issue
- Custom fields supported

### Todoist Plugin (Planned)

Sync tasks with Todoist.

**Configuration**:

- `TODU_TODOIST_TOKEN` - API token

**Mappings**:

- Project → Todoist Project
- Task → Todoist Task
- Priority levels mapped
- Labels and due dates supported

### Linear Plugin (Planned)

Sync tasks with Linear.

**Configuration**:

- `TODU_LINEAR_API_KEY` - API key
- `TODU_LINEAR_TEAM_ID` - Team ID

**Mappings**:

- Project → Linear Project
- Task → Linear Issue
- Statuses and priorities mapped

## Developing New Plugins

See [Plugin Development Guide](plugin-development.md) for creating custom plugins.

## Troubleshooting

### Plugin Not Found

```bash
# Check registered systems
todu system list

# Verify plugin is available
todu system config <plugin-name>
```

### Authentication Errors

```bash
# Verify token is set
echo $TODU_GITHUB_TOKEN

# Check token hasn't expired
# Regenerate token if needed
```

### Rate Limiting

Most APIs have rate limits:

**GitHub**: 5000 requests/hour (authenticated)
**Solution**: Reduce sync frequency or use multiple tokens

### API Version Issues

Some self-hosted instances may have different API versions:

**Problem**: API calls fail with "not found"
**Solution**: Check API documentation for your version

### SSL/TLS Issues

For self-hosted instances with self-signed certificates:

```bash
# Temporary (not recommended for production)
export TODU_SKIP_VERIFY="true"

# Better: Add certificate to system trust store
```

## Best Practices

### Token Management

1. **Use separate tokens per application**
2. **Rotate tokens regularly**
3. **Never commit tokens to version control**
4. **Use minimum required permissions**

### Sync Strategies

1. **Start with pull-only** to avoid accidental changes
2. **Test with dry-run** before actual sync
3. **Use bidirectional** only when you understand implications
4. **Monitor sync logs** for issues

### Performance

1. **Don't sync too frequently** (respect API rate limits)
2. **Sync specific projects** if you have many
3. **Use filters** to reduce data transfer
4. **Monitor API usage** in external systems

### Multiple Instances

If using multiple instances of the same system:

```bash
# Not currently supported
# Workaround: Use different config profiles

# Create separate config for each instance
mkdir -p ~/.config/todu-prod
mkdir -p ~/.config/todu-staging

# Use different configs
TODU_CONFIG=~/.config/todu-prod/config.yaml todu sync --all
TODU_CONFIG=~/.config/todu-staging/config.yaml todu sync --all
```
