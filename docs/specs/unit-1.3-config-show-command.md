# Unit 1.3: Config Show Command

**Status**: ✅ COMPLETE

**Goal**: Add `todu config show` command

**Prerequisites**: Unit 1.2 complete

**Estimated time**: 15 minutes

---

## Requirements

### 1. Config Command Group

Create `cmd/todu/cmd/config.go` with:

- `config` parent command
- Appropriate help text
- Short and long descriptions

### 2. Show Subcommand

Implement `config show` subcommand that:

- Displays the current configuration
- Shows API URL
- Shows daemon settings (interval, projects list)
- Shows output settings (format, color)
- Uses human-readable formatting
- Loads configuration using `config.Load()`
- Returns error if configuration fails to load

### 3. Output Format

The output must:

- Be clearly labeled and organized
- Group related settings together
- Be easy to read for both humans and LLMs
- Display nested configuration hierarchically

### 4. Command Integration

- Register `config` command with root command
- Register `show` subcommand with `config` command
- Ensure proper command hierarchy

### 5. Help Text

Must provide:

- Help for `todu config` command
- Help for `todu config show` command
- Both short and long descriptions

---

## Success Criteria

- ✅ Code compiles: `go build ./cmd/todu`
- ✅ `todu config` displays help text
- ✅ `todu config show` displays current configuration
- ✅ Shows defaults when no config file exists
- ✅ Output is clear and well-formatted
- ✅ `todu config --help` works
- ✅ `todu config show --help` works

---

## Verification

Build and test:

- `go build -o todu ./cmd/todu`
- `./todu config` - shows config command help
- `./todu config show` - displays configuration
- `./todu config --help` - shows help
- Output must include all configuration values

---

## Commit Message

```text
feat: add config show command
```
