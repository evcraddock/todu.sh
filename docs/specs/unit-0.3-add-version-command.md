# Unit 0.3: Add Version Command with Cobra

**Goal**: Set up Cobra CLI framework with version command

**Prerequisites**: Unit 0.2 complete

**Estimated time**: 15 minutes

---

## Requirements

### 1. CLI Framework

- Add Cobra dependency to the project
- Set up Cobra command structure in `cmd/todu/cmd/` package
- Create a root command that serves as the main CLI entry point
- Root command must provide help text and usage information

### 2. Root Command

The root command must:

- Use the command name `todu`
- Provide a short description of the application
- Provide a long description explaining its purpose
- Be accessible via `Execute()` function
- Handle errors appropriately

### 3. Version Command

Implement a `version` subcommand that:

- Displays the version number
- Displays the git commit hash
- Displays the build date
- Has appropriate help text
- Uses variables that can be set at build time: `Version`, `Commit`, `BuildDate`
- Defaults: Version="dev", Commit="none", BuildDate="unknown"

### 4. Integration

- Update `main.go` to call the Cobra `Execute()` function
- Handle errors from `Execute()` with appropriate exit codes
- Ensure clean error propagation

---

## Success Criteria

- ✅ Cobra dependency added to `go.mod`
- ✅ `go build ./cmd/todu` compiles successfully
- ✅ `./todu` displays help text
- ✅ `./todu version` displays version information
- ✅ `./todu --help` shows comprehensive help
- ✅ `./todu version --help` shows version command help
- ✅ Error handling works correctly

---

## Verification

Commands to verify:

- `go build -o todu ./cmd/todu` - must succeed
- `./todu` - must display help
- `./todu version` - must show version info
- `./todu --help` - must show help
- `./todu version --help` - must show version command help

---

## Commit Message

```text
feat: add cobra framework and version command
```
