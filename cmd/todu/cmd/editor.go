package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// openEditor opens the user's preferred editor with optional initial content
// Returns the edited content or an error
func openEditor(initialContent string) (string, error) {
	editor := getEditor()

	// Create a temporary file with .md extension for syntax highlighting
	tmpFile, err := os.CreateTemp("", "todu-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write initial content if provided
	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			tmpFile.Close()
			return "", fmt.Errorf("failed to write initial content: %w", err)
		}
	}
	tmpFile.Close()

	// Open editor
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the edited content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	// Trim whitespace and return
	result := strings.TrimSpace(string(content))
	return result, nil
}

// getEditor returns the user's preferred editor
// Priority: $VISUAL > $EDITOR > vim > vi > nano
func getEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	// Try to find vim, vi, or nano
	for _, editor := range []string{"vim", "vi", "nano"} {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}
	// Default to vi (POSIX standard)
	return "vi"
}
