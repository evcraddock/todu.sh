# Unit 8.1: Polish and Documentation

**Status**: ✅ COMPLETE

**Goal**: Finalize CLI polish and create comprehensive documentation

**Prerequisites**: Unit 7.2 complete

**Estimated time**: 60 minutes

---

## Requirements

### 1. Enhanced Help Text

Review and improve all command help text:

- Clear short descriptions (one line)
- Detailed long descriptions with context
- Examples in help output
- Common workflows documented
- Flag descriptions are helpful
- Required vs optional flags are clear

For each command group (config, system, project, task, sync, daemon):

- Add usage examples
- Show common patterns
- Explain when to use each command

### 2. Error Messages

Review and improve error messages:

- Clear and actionable
- Suggest next steps when operations fail
- Provide helpful hints for common issues
- Include relevant context (IDs, names)
- Examples:
  - "System not found (ID: 5). Use 'todu system list' to see available systems."
  - "Plugin 'github' not configured. Run 'todu system config github' for details."
  - "Project has 10 tasks. Delete tasks first or use --force to delete anyway."

### 3. Output Improvements

Ensure consistent formatting:

- Tables: Aligned columns, clear headers
- Details: Grouped related fields
- Lists: Sorted appropriately
- Timestamps: Human-readable (e.g., "2 hours ago")
- Color: Use colors appropriately (if enabled)
  - Green for success
  - Yellow for warnings
  - Red for errors
  - Blue for info

### 4. Progress Indicators

Add progress indicators for long operations:

- Sync operations
- Bulk operations
- API calls with multiple requests
- Use simple text-based progress (no fancy spinners for LLM compatibility)

### 5. User Documentation

Create documentation files:

**`README.md`:**

- Project overview
- Quick start guide
- Installation instructions
- Basic usage examples
- Link to detailed docs

**`docs/installation.md`:**

- Installation from source
- Installation from binary (future)
- Requirements (Go version, etc.)
- Platform-specific notes

**`docs/configuration.md`:**

- Configuration file format
- Environment variables
- Configuration precedence
- All config options documented
- Examples for each plugin

**`docs/workflows.md`:**

- Initial setup workflow
- Linking external systems
- Running first sync
- Managing tasks
- Setting up daemon
- Common troubleshooting

**`docs/plugins.md`:**

- Available plugins
- Plugin configuration
- Plugin-specific notes
- External system requirements
- Authentication setup

**`docs/cli-reference.md`:**

- Complete command reference
- All flags documented
- Examples for each command
- Output format specifications

### 6. Plugin Development Guide

Create `docs/plugin-development.md`:

- Plugin interface overview
- How to implement a plugin
- Type mapping guidelines
- Testing guidelines
- Registration process
- Example plugin walkthrough
- Template for new plugins

### 7. Example Workflows

Document common workflows in `docs/workflows.md`:

**Initial Setup:**

```bash
# 1. Set API URL
todu config show

# 2. Add GitHub system
todu system add --identifier github --name "GitHub"

# 3. Configure GitHub (set environment variable)
export TODU_PLUGIN_GITHUB_TOKEN="ghp_..."

# 4. Discover repositories
todu project discover --system 1

# 5. Link a repository
todu project add --system 1 --external-id "owner/repo" --name "My Project"
```

**Running Sync:**

```bash
# Dry run to preview
todu sync --dry-run

# Sync specific project
todu sync --project 1

# Sync all projects
todu sync --all
```

**Managing Tasks:**

```bash
# List open tasks
todu task list --status open

# Show task details
todu task show 123

# Close a task
todu task close 123

# Add comment
todu task comment 123 "This is fixed"
```

**Setting up Daemon:**

```bash
# Install daemon
todu daemon install --interval 5m

# Check status
todu daemon status

# View logs
todu daemon logs --follow
```

### 8. LLM-Friendly Output

Ensure all output is LLM-friendly:

- Structured and parseable
- Clear field labels
- Consistent formatting
- JSON output available for all commands
- Tables are well-aligned
- No unnecessary decorations
- Progress messages are clear

### 9. Completion Scripts

Generate shell completion scripts:

- Bash completion
- Zsh completion
- Fish completion

Using cobra's built-in completion generation.

### 10. Testing

Manual testing checklist:

- All commands work as documented
- Help text is accurate
- Examples in docs are correct
- Error messages are helpful
- Output is consistent
- No broken links in docs
- LLM can use CLI from help text alone

---

## Success Criteria

- ✅ All commands have enhanced help text
- ✅ Error messages are clear and actionable
- ✅ Output is consistent and well-formatted
- ✅ All documentation is complete
- ✅ Examples work as documented
- ✅ LLM can use CLI effectively
- ✅ Completion scripts available
- ✅ Manual testing passes

---

## Verification

- Review all help text
- Test all examples in documentation
- Verify error messages are helpful
- Test with LLM (provide only help text and docs)
- Check that documentation covers all features

---

## Commit Message

```text
docs: add comprehensive documentation and polish CLI
```
