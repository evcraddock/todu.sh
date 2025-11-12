package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "todu",
	Short: "Task management across multiple systems",
	Long: `todu.sh - A unified task management CLI

todu provides a single interface to manage tasks and issues across
multiple systems including GitHub, Forgejo, Todoist, and more.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add global flags here if needed
}
