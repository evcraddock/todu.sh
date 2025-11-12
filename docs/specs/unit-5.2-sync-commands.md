# Unit 5.2: Sync Commands

**Status**: 🔲 TODO

**Goal**: Implement CLI commands for running sync operations

**Prerequisites**: Unit 5.1 complete

**Estimated time**: 30 minutes

---

## Requirements

### 1. Sync Command Group

Create `cmd/todu/cmd/sync.go` with:

- `sync` parent command
- Appropriate help text
- Short and long descriptions

### 2. Sync Run Command

Implement `sync` or `sync run` subcommand that:

- Runs synchronization
- Optional `--project <id>` flag to sync specific project
- Optional `--system <id>` flag to sync all projects for system
- Optional `--all` flag to sync all projects (default if no flags)
- Optional `--strategy <strategy>` flag (pull/push/bidirectional, default: bidirectional)
- Optional `--dry-run` flag to preview changes
- Shows progress during sync
- Shows summary results after sync

### 3. Sync Status Command

Implement `sync status` subcommand that:

- Lists all projects with last sync times
- Shows when each project was last synced
- Shows time since last sync (e.g., "2 hours ago")
- Highlights projects never synced
- Optional `--system <id>` flag to filter by system

### 4. Progress Display

During sync:
- Show current project being synced
- Show tasks processed
- Use progress indicators
- Show errors as they occur

After sync:
- Summary table with results per project
- Total tasks created/updated/skipped
- Total time taken
- Any errors encountered

### 5. Error Handling

- Continue syncing even if one project fails
- Show all errors at end
- Exit with non-zero code if any errors
- Helpful error messages

### 6. Dry Run Output

When `--dry-run` is used:
- Show "DRY RUN" prominently
- Display what would be created/updated
- Show conflicts that would be resolved
- Don't make any actual changes

---

## Success Criteria

- ✅ `todu sync` syncs all projects
- ✅ `todu sync --project <id>` syncs one project
- ✅ `todu sync --system <id>` syncs system projects
- ✅ `todu sync --dry-run` shows preview
- ✅ `todu sync status` shows last sync times
- ✅ Progress is displayed clearly
- ✅ Results are summarized well
- ✅ Help text is clear

---

## Verification

Commands to test:
- `todu sync --help`
- `todu sync --dry-run`
- `todu sync --project 1`
- `todu sync --system 1`
- `todu sync --all`
- `todu sync --strategy pull`
- `todu sync status`

---

## Commit Message

```text
feat: add sync commands
```
