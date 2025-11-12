# Unit 4.1: Project Management Commands

**Status**: ✅ DONE

**Goal**: Implement CLI commands for managing projects

**Prerequisites**: Unit 3.2 complete

**Estimated time**: 40 minutes

---

## Requirements

### 1. Project Command Group

Create `cmd/todu/cmd/project.go` with:

- `project` parent command
- Appropriate help text
- Short and long descriptions

### 2. List Projects Command

Implement `project list` subcommand that:

- Lists all projects from Todu API
- Optional `--system <id>` flag to filter by system
- Displays: ID, Name, System, External ID, Status
- Uses table format for text output
- Supports `--format json` flag
- Shows system name (not just ID)

### 3. Add Project Command

Implement `project add` subcommand that:

- Creates a new project in Todu API
- Requires `--system <id>` flag (system ID from `todu system list`)
- Requires `--external-id <id>` flag (e.g., "owner/repo" for GitHub)
- Requires `--name <name>` flag
- Optional `--description <text>` flag
- Optional `--status <status>` flag (default: "active")
- Optional `--sync-strategy <strategy>` flag (default: "bidirectional")
  - Valid values: "pull", "push", "bidirectional"
  - "pull": External → Todu only (read-only sync from external)
  - "push": Todu → External only (push changes to external)
  - "bidirectional": Two-way sync with conflict resolution
- Validates system exists
- Validates sync strategy is valid
- Shows created project details

### 4. Show Project Command

Implement `project show <id>` subcommand that:

- Displays detailed project information
- Shows all fields: ID, Name, Description, System, External ID, Status, Sync Strategy
- Shows when project was created/updated
- Displays associated system details
- Uses human-readable format

### 5. Update Project Command

Implement `project update <id>` subcommand that:

- Updates project fields
- Optional `--name <name>` flag
- Optional `--description <text>` flag
- Optional `--status <status>` flag
- Optional `--sync-strategy <strategy>` flag
  - Valid values: "pull", "push", "bidirectional"
  - Validates strategy is valid
- Only updates fields that are provided
- Shows updated project details

### 6. Remove Project Command

Implement `project remove <id>` subcommand that:

- Deletes project from Todu API
- Requires confirmation (unless `--force` flag)
- Shows error if project has associated tasks
- Provides helpful message about what needs to be done first

### 7. Discover Projects Command

Implement `project discover` subcommand that:

- Lists available projects from external system
- Requires `--system <id>` flag
- Uses plugin's FetchProjects() method
- Shows external_id, name, description
- Allows user to select which to import
- Option: `--auto-import` to import all
- Shows which projects are already imported

### 8. Output Formatting

For text output:

- Use tables for lists
- Clear column headers
- Aligned columns
- Truncate long descriptions

For JSON output:

- Pretty-printed JSON
- Include all fields
- Arrays for lists

---

## Success Criteria

- ✅ All project commands implemented
- ✅ `todu project list` shows projects
- ✅ `todu project add` creates projects
- ✅ `todu project show` displays details
- ✅ `todu project update` modifies projects
- ✅ `todu project remove` deletes projects
- ✅ `todu project discover` lists external projects
- ✅ Output formats work (text and JSON)
- ✅ Help text is clear for all commands

---

## Verification

Commands to test:

- `todu project --help`
- `todu project list`
- `todu project list --system 1`
- `todu project add --system 1 --external-id "owner/repo" --name "My Project"`
- `todu project add --system 1 --external-id "owner/repo2" \`
  `--name "My Project 2" --sync-strategy pull`
- `todu project show 1`
- `todu project update 1 --name "Updated Name"`
- `todu project discover --system 1`
- `todu project remove 1`

---

## Commit Message

```text
feat: add project management commands
```
