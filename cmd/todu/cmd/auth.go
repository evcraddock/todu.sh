package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/config"
	"github.com/spf13/cobra"
)

var noBrowser bool

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with the todu API via browser",
	Long: `Authenticate with the todu API using a browser-based flow.

This command opens the todu web UI in your browser where you can log in
and generate an API key. After copying the key, paste it back into the CLI.

Use --no-browser to print the URL instead of opening it automatically.`,
	RunE: runAuth,
}

func init() {
	authCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "print the URL instead of opening the browser")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	loginURL := buildLoginURL(cfg.APIURL)

	if noBrowser {
		fmt.Println("Open this URL in your browser to authenticate:")
		fmt.Println()
		fmt.Printf("  %s\n", loginURL)
		fmt.Println()
	} else {
		fmt.Println("Opening browser for authentication...")
		if err := openBrowser(loginURL); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open browser: %v\n", err)
			fmt.Println("Open this URL manually:")
			fmt.Println()
			fmt.Printf("  %s\n", loginURL)
			fmt.Println()
		}
	}

	fmt.Println("After logging in, copy the API key and paste it below.")
	fmt.Println()

	apiKey, err := promptAPIKey(cmd)
	if err != nil {
		return err
	}

	if !isValidKeyFormat(apiKey) {
		fmt.Fprintln(os.Stderr, "⚠ Warning: API key does not start with 'sk_'. It may be invalid.")
	}

	// Verify the key against the API
	fmt.Println("Verifying API key...")
	verifyErr := verifyAPIKey(cfg.APIURL, apiKey)

	if verifyErr != nil {
		fmt.Fprintf(os.Stderr, "⚠ Warning: API key verification failed: %v\n", verifyErr)
		fmt.Println()

		save, err := promptYesNo(cmd, "Save the key anyway?")
		if err != nil {
			return err
		}
		if !save {
			fmt.Println("Aborted. API key was not saved.")
			return nil
		}
	} else {
		fmt.Println("✓ API key verified successfully.")
	}

	// Save the key
	configPath := GetConfigFile()
	if err := config.SetAPIKey(configPath, apiKey); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	fmt.Println("✓ API key saved. You're ready to use todu.")
	return nil
}

// openBrowser opens the given URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// buildLoginURL constructs the browser login URL from the API base URL.
func buildLoginURL(apiURL string) string {
	return apiURL + "/ui/auth/login?next=/ui/auth/cli-setup"
}

// isValidKeyFormat checks whether an API key has the expected sk_ prefix.
func isValidKeyFormat(key string) bool {
	return strings.HasPrefix(key, "sk_")
}

// verifyAPIKey checks that the given API key works against the API.
func verifyAPIKey(apiURL, apiKey string) error {
	client := api.NewClient(apiURL, apiKey)
	_, err := client.ListProjects(context.Background(), nil)
	return err
}

// promptAPIKey reads the API key from stdin.
func promptAPIKey(cmd *cobra.Command) (string, error) {
	fmt.Print("API Key: ")

	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read API key: %w", err)
	}

	key := strings.TrimSpace(line)
	if key == "" {
		return "", fmt.Errorf("API key cannot be empty")
	}

	return key, nil
}

// promptYesNo asks a yes/no question and returns the answer.
func promptYesNo(cmd *cobra.Command, question string) (bool, error) {
	fmt.Printf("%s [y/N]: ", question)

	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes", nil
}
