# Unit 6.1: Task Management Commands

**Status**: 🔲 TODO

**Goal**: Implement CLI commands for managing tasks

**Prerequisites**: Unit 5.2 complete

**Estimated time**: 50 minutes

---

## Requirements

### 1. Task Command Group

Create `cmd/todu/cmd/task.go` with:

- `task` parent command
- Appropriate help text
- Short and long descriptions

### 2. List Tasks Command

Implement `task list` subcommand that:

- Lists tasks from Todu API
- Filter options:
  - `--status <status>` - Filter by status
  - `--priority <priority>` - Filter by priority
  - `--project <id>` - Filter by project
  - `--assignee <name>` - Filter by assignee
  - `--label <name>` - Filter by label (repeatable)
  - `--search <text>` - Full-text search
  - `--due-before <date>` - Due before date
  - `--due-after <date>` - Due after date
- Displays: ID, Title, Status, Priority, Project, Due Date
- Uses table format for text output
- Supports `--format json` flag
- Limit results with `--limit <n>` (default: 50)

### 3. Show Task Command

Implement `task show <id>` subcommand that:

- Displays detailed task information
- Shows all fields
- Shows labels and assignees
- Shows all comments with timestamps and authors
- Shows source URL if available
- Uses human-readable format

### 4. Create Task Command

Implement `task create` subcommand that:

- Creates a new task in Todu API
- Required `--title <text>` flag
- Required `--project <id>` flag
- Optional `--description <text>` flag
- Optional `--status <status>` flag (default: "open")
- Optional `--priority <priority>` flag
- Optional `--due <date>` flag (format: YYYY-MM-DD)
- Optional `--label <name>` flags (repeatable)
- Optional `--assignee <name>` flags (repeatable)
- Shows created task details
- Option: Interactive mode if no flags provided

### 5. Update Task Command

Implement `task update <id>` subcommand that:

- Updates task fields
- Optional `--title <text>` flag
- Optional `--description <text>` flag
- Optional `--status <status>` flag
- Optional `--priority <priority>` flag
- Optional `--due <date>` flag
- Optional `--add-label <name>` flags (repeatable)
- Optional `--remove-label <name>` flags (repeatable)
- Optional `--add-assignee <name>` flags (repeatable)
- Optional `--remove-assignee <name>` flags (repeatable)
- Only updates fields that are provided
- Shows updated task details

### 6. Close Task Command

Implement `task close <id>` subcommand that:

- Marks task as done/closed
- Shortcut for `task update <id> --status done`
- Shows updated task

### 7. Comment Command

Implement `task comment <id>` subcommand that:

- Adds comment to task
- Required positional argument: comment text
- Or `--message <text>` flag for longer comments
- Sets author to current user (from config)
- Shows added comment

### 8. Delete Task Command

Implement `task delete <id>` subcommand that:

- Deletes task from Todu API
- Requires confirmation (unless `--force` flag)
- Shows success message

### 9. Output Formatting

For text output:

- Tables for lists
- Clear field labels for details
- Readable date formats
- Color coding for status/priority (if output.color is true)

For JSON output:

- Pretty-printed JSON
- Include all fields including nested objects

### 10. Interactive Create Mode

When `task create` is run without flags:

- Prompt for each field
- Allow user to skip optional fields
- Validate input
- Show preview before creating

---

## Success Criteria

- ✅ All task commands implemented
- ✅ `todu task list` shows tasks with filters
- ✅ `todu task show` displays full details
- ✅ `todu task create` creates tasks
- ✅ `todu task update` modifies tasks
- ✅ `todu task close` marks tasks done
- ✅ `todu task comment` adds comments
- ✅ `todu task delete` removes tasks
- ✅ Filtering works correctly
- ✅ Output is LLM-friendly
- ✅ Help text is clear

---

## Verification

Commands to test:

- `todu task --help`
- `todu task list`
- `todu task list --status open --project 1`
- `todu task show 1`
- `todu task create --title "New Task" --project 1`
- `todu task update 1 --status done`
- `todu task close 1`
- `todu task comment 1 "This is a comment"`
- `todu task delete 1`

---

## Commit Message

```text
feat: add task management commands
```
