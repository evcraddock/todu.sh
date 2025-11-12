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

Create platform-specific service management files:

**`internal/daemon/service.go`** (shared interface):

- Define common service operations interface
- Service name: "todu"
- Service display name: "Todu Task Sync Daemon"

**`internal/daemon/service_darwin.go`** (macOS):

- Generate and write launchd plist files
- Use `launchctl` commands via `exec.Command`
- Target: `~/Library/LaunchAgents/com.todu.daemon.plist`

**`internal/daemon/service_linux.go`** (Linux):

- Generate and write systemd unit files
- Use `systemctl --user` commands via `exec.Command`
- Target: `~/.config/systemd/user/todu.service` (user service)

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
  - Linux: `~/.config/systemd/user/todu.service`
- Sets service to start at login/boot
- Shows installation instructions

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

- Generate plist XML with `KeepAlive`, `RunAtLoad`, and `ProgramArguments`
- Write to `~/Library/LaunchAgents/com.todu.daemon.plist`
- Use `launchctl load/unload/start/stop` via `exec.Command`
- User-level service (no sudo required)

**Linux (systemd):**

- Generate unit file with `[Unit]`, `[Service]`, and `[Install]` sections
- Write to `~/.config/systemd/user/todu.service`
- Use `systemctl --user enable/disable/start/stop` via `exec.Command`
- User-level service (no sudo required)
- Run `systemctl --user daemon-reload` after writing unit file

### 11. Error Handling

- Handle missing directories (create `~/.config/systemd/user/` if needed)
- Provide clear error messages for command failures
- Verify executable path exists before writing service files

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
