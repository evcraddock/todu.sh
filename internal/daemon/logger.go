package daemon

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/evcraddock/todu.sh/internal/config"
)

// setupLogger creates a zerolog logger with rotation based on config
func setupLogger(cfg *config.Config) (zerolog.Logger, error) {
	// Get home directory for log path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return zerolog.Logger{}, err
	}

	logPath := filepath.Join(homeDir, ".config", "todu", "daemon.log")

	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return zerolog.Logger{}, err
	}

	// Set up log rotation
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.Daemon.LogMaxSizeMB,   // megabytes
		MaxBackups: cfg.Daemon.LogMaxBackups,  // number of old files to keep
		MaxAge:     cfg.Daemon.LogMaxAgeDays,  // days
		Compress:   false,                     // don't compress old logs
	}

	// Parse log level
	level := parseLogLevel(cfg.Daemon.LogLevel)

	// Create multi-writer (file with rotation)
	var writers []io.Writer
	writers = append(writers, fileWriter)

	// If running interactively (not via launchd), also write to console
	if isInteractive() {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})
	}

	multi := io.MultiWriter(writers...)

	// Create logger
	logger := zerolog.New(multi).
		Level(level).
		With().
		Timestamp().
		Logger()

	return logger, nil
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// isInteractive checks if the process is running in an interactive terminal
func isInteractive() bool {
	// Check if stdout is a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
