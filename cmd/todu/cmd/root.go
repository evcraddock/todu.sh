package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "todu",
	Short: "Task management across multiple systems",
	Long: `todu.sh - A unified task management CLI

todu provides a single interface to manage tasks and issues across
multiple external task management systems via a plugin architecture.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path (default: ./config.yaml, ~/.config/todu/config.yaml, or ~/.todu/config.yaml)")
}

// GetConfigFile returns the config file path from the --config flag
func GetConfigFile() string {
	return configFile
}
