package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

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

// getVersionInfo returns version information, preferring ldflags values
// but falling back to Go's embedded build info for go install builds.
func getVersionInfo() (version, commit, buildDate string) {
	version = Version
	commit = Commit
	buildDate = BuildDate

	// If ldflags weren't set, try Go's build info
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			// Module version (works with go install @version)
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}

			// VCS info from build settings
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if commit == "none" {
						commit = setting.Value
						if len(commit) > 7 {
							commit = commit[:7]
						}
					}
				case "vcs.time":
					if buildDate == "unknown" {
						buildDate = setting.Value
					}
				case "vcs.modified":
					// Only add -dirty if not already indicated in version string
					if setting.Value == "true" && version != "dev" &&
						!strings.Contains(version, "dirty") {
						version += "-dirty"
					}
				}
			}
		}
	}

	return version, commit, buildDate
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of todu",
	Long:  `All software has versions. This is todu's`,
	Run: func(cmd *cobra.Command, args []string) {
		version, commit, buildDate := getVersionInfo()
		fmt.Printf("todu version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", buildDate)
		fmt.Printf("  go:     %s\n", runtime.Version())
		fmt.Printf("  os:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
