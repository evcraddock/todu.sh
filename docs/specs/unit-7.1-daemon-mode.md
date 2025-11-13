# Unit 7.1: Daemon Mode Implementation

**Status**: ✅ COMPLETE

**Goal**: Implement background daemon for periodic sync

**Prerequisites**: Unit 6.1 complete

**Estimated time**: 45 minutes

---

## Requirements

### 1. Daemon Package

Create `internal/daemon/daemon.go` with:

```go
type Daemon struct {
    engine   *sync.Engine
    config   *config.Config
    stopChan chan struct{}
}

func New(engine *sync.Engine, config *config.Config) *Daemon
```

### 2. Daemon Methods

Implement:

- `Start(ctx context.Context) error`
  - Starts daemon main loop
  - Runs sync on configured interval
  - Respects context cancellation
  - Logs sync results

- `Stop() error`
  - Gracefully stops daemon
  - Waits for current sync to complete
  - Cleans up resources

### 3. Periodic Sync Logic

In daemon main loop:

1. Parse interval from config (e.g., "5m")
2. Wait for interval duration
3. Run sync with configured projects (or all if empty)
4. Log results (successes and errors)
5. Handle errors gracefully (don't crash)
6. Exponential backoff on repeated failures

### 4. Signal Handling

Implement graceful shutdown:

- Catch SIGINT and SIGTERM
- Call Stop() on daemon
- Wait for current sync to finish
- Exit cleanly

### 5. Logging

Implement logging in `internal/daemon/logger.go`:

- Log to stdout/stderr
- Include timestamps
- Log levels: INFO, WARN, ERROR
- Log each sync start/end
- Log sync results
- Log errors with full context

### 6. Error Recovery

Handle errors:

- API connection failures
- Plugin errors
- Configuration errors

On error:

- Log error with details
- Increment failure count
- Use exponential backoff (up to max interval)
- Reset backoff on success

### 7. Configuration

Use config values from `config.yaml`:

```yaml
daemon:
  interval: 5m
  projects: []  # empty = all projects
```

### 8. Health Status

Implement status tracking:

- Last sync time
- Last sync result
- Current status (running/stopped)
- Error count
- Write status to file: `~/.todu/daemon.status`

### 9. Testing

Create `internal/daemon/daemon_test.go`:

- Test daemon start/stop
- Test periodic sync execution
- Test signal handling
- Test error recovery
- Test exponential backoff
- Mock sync engine

---

## Success Criteria

- ✅ Daemon starts and runs sync on interval
- ✅ Daemon stops gracefully on signal
- ✅ Errors don't crash daemon
- ✅ Exponential backoff works
- ✅ Logs are written correctly
- ✅ Status file is maintained
- ✅ Tests pass: `go test ./internal/daemon`

---

## Verification

- `go test ./internal/daemon -v` - all tests pass
- Daemon runs in foreground
- Can be stopped with Ctrl+C
- Sync runs on schedule

---

## Commit Message

```text
feat: implement daemon mode
```
