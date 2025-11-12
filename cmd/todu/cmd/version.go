package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build time
	Version = "dev"
	// Commit is the git commit hash, set during build time
	Commit = "none"
	// BuildDate is the date the binary was built, set during build time
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of todu",
	Long:  `All software has versions. This is todu's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("todu version %s\n", Version)
		fmt.Printf("  commit: %s\n", Commit)
		fmt.Printf("  built:  %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
