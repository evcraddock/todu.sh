# Unit 2.3: System Management Commands

**Status**: ✅ COMPLETE

**Goal**: Implement CLI commands for managing external systems

**Prerequisites**: Unit 2.2 complete

**Estimated time**: 30 minutes

---

## Requirements

### 1. System Command Group

Create `cmd/todu/cmd/system.go` with:

- `system` parent command
- Appropriate help text
- Short and long descriptions

### 2. List Systems Command

Implement `system list` subcommand that:

- Lists all registered systems from Todu API
- Displays: ID, Identifier, Name, URL
- Uses table format for text output
- Supports `--format json` flag
- Shows "(not configured)" if plugin config missing

### 3. Add System Command

Implement `system add` subcommand that:

- Creates a new system in Todu API
- Requires `--identifier <name>` flag (plugin name)
- Requires `--name <display-name>` flag
- Optional `--url <api-url>` flag
- Optional `--metadata key=value` flags (repeatable)
- Validates that plugin exists in registry
- Shows configuration requirements after creation

### 4. Show System Command

Implement `system show <id>` subcommand that:

- Displays detailed system information
- Shows all metadata fields
- Shows when system was created/updated
- Displays plugin configuration requirements
- Shows whether plugin is configured (has env vars)

### 5. Config Requirements Command

Implement `system config <identifier>` subcommand that:

- Shows configuration requirements for a plugin
- Lists required environment variables
- Example: `TODU_PLUGIN_GITHUB_TOKEN`
- Provides description of each variable
- Shows current configuration status (set/unset)

### 6. Remove System Command

Implement `system remove <id>` subcommand that:

- Deletes system from Todu API
- Requires confirmation (unless `--force` flag)
- Shows error if system has associated projects
- Provides helpful message about what needs to be done first

### 7. Error Handling

- Clear error messages for invalid identifiers
- Helpful suggestions when plugin not found
- List available plugins when identifier is invalid
- Handle API errors gracefully

---

## Success Criteria

- ✅ All system commands implemented
- ✅ `todu system list` shows systems
- ✅ `todu system add` creates systems
- ✅ `todu system show` displays details
- ✅ `todu system config` shows requirements
- ✅ `todu system remove` deletes systems
- ✅ Help text is clear for all commands
- ✅ Error messages are helpful

---

## Verification

Commands to test:
- `todu system --help`
- `todu system list`
- `todu system add --identifier test --name "Test System"`
- `todu system show 1`
- `todu system config test`
- `todu system remove 1`

---

## Commit Message

```text
feat: add system management commands
```
