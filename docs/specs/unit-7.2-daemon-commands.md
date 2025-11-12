# Unit 7.2: Daemon Commands and Service Management

**Status**: 🔲 TODO

**Goal**: Implement CLI commands for daemon management

**Prerequisites**: Unit 7.1 complete

**Estimated time**: 40 minutes

---

## Requirements

### 1. Daemon Command Group

Create `cmd/todu/cmd/daemon.go` with:

- `daemon` parent command
- Appropriate help text
- Short and long descriptions

### 2. Service Management Package

Create `internal/daemon/service.go`:

- Use `github.com/kardianos/service` package
- Implement service.Interface
- Support macOS (launchd) and Linux (systemd)
- Service name: "todu"
- Service display name: "Todu Task Sync Daemon"

### 3. Daemon Start Command (Foreground)

Implement `daemon start` subcommand that:

- Runs daemon in foreground
- Uses daemon.Start() from Unit 7.1
- Logs to stdout
- Can be stopped with Ctrl+C
- Shows "Daemon started, syncing every X minutes"

### 4. Daemon Install Command

Implement `daemon install` subcommand that:

- Installs daemon as system service
- Optional `--interval <duration>` flag (default from config)
- Optional `--projects <ids>` flag (comma-separated)
- Creates service configuration file:
  - macOS: `~/Library/LaunchAgents/com.todu.daemon.plist`
  - Linux: `/etc/systemd/system/todu.service`
- Sets service to start at login/boot
- Shows installation instructions
- Requires appropriate permissions

### 5. Daemon Uninstall Command

Implement `daemon uninstall` subcommand that:

- Stops daemon if running
- Removes service configuration
- Shows success message

### 6. Daemon Status Command

Implement `daemon status` subcommand that:

- Shows daemon status:
  - Running / Stopped
  - Last sync time
  - Next sync time
  - Sync interval
  - Projects being synced
- Reads from status file: `~/.todu/daemon.status`
- Shows if service is installed
- Shows service status (if installed)

### 7. Daemon Stop Command

Implement `daemon stop` subcommand that:

- Stops running daemon service
- Sends SIGTERM to daemon process
- Waits for graceful shutdown
- Shows success message
- Works for both foreground and service modes

### 8. Daemon Restart Command

Implement `daemon restart` subcommand that:

- Stops then starts daemon service
- Shows restart progress

### 9. Daemon Logs Command

Implement `daemon logs` subcommand that:

- Shows recent daemon logs
- Optional `--follow` flag to tail logs
- Optional `--lines <n>` flag (default: 50)
- Reads from daemon log file
- Works for both foreground and service modes

### 10. Platform-Specific Implementation

**macOS (launchd):**
- Create plist file
- Use `launchctl load/unload`
- User-level service (LaunchAgents)

**Linux (systemd):**
- Create unit file
- Use `systemctl enable/disable/start/stop`
- User-level service or system-level

### 11. Permission Handling

- Check for appropriate permissions
- Guide user on using sudo if needed
- Handle permission errors gracefully

---

## Success Criteria

- ✅ `todu daemon start` runs daemon in foreground
- ✅ `todu daemon install` installs as service
- ✅ `todu daemon status` shows daemon state
- ✅ `todu daemon stop` stops daemon
- ✅ `todu daemon restart` restarts daemon
- ✅ `todu daemon uninstall` removes service
- ✅ `todu daemon logs` shows logs
- ✅ Works on macOS and Linux
- ✅ Help text is clear

---

## Verification

Commands to test:
- `todu daemon --help`
- `todu daemon start` (foreground mode)
- `todu daemon install`
- `todu daemon status`
- `todu daemon logs`
- `todu daemon stop`
- `todu daemon restart`
- `todu daemon uninstall`

Platform-specific:
- macOS: Check plist file creation
- Linux: Check systemd unit file
- Verify service starts at login/boot

---

## Commit Message

```text
feat: add daemon commands and service management
```
