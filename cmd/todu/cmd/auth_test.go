package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAuthCmd_Exists(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("auth command not registered on root command")
	}
}

func TestAuthCmd_HasNoBrowserFlag(t *testing.T) {
	flag := authCmd.Flags().Lookup("no-browser")
	if flag == nil {
		t.Error("auth command missing --no-browser flag")
	}
}

func TestBuildLoginURL(t *testing.T) {
	tests := []struct {
		apiURL string
		want   string
	}{
		{"http://localhost:8000", "http://localhost:8000/ui/auth/login?next=/ui/auth/cli-setup"},
		{"https://api.example.com", "https://api.example.com/ui/auth/login?next=/ui/auth/cli-setup"},
		{"http://10.10.1.197:8000", "http://10.10.1.197:8000/ui/auth/login?next=/ui/auth/cli-setup"},
	}

	for _, tt := range tests {
		got := buildLoginURL(tt.apiURL)
		if got != tt.want {
			t.Errorf("buildLoginURL(%q) = %q, want %q", tt.apiURL, got, tt.want)
		}
	}
}

func TestIsValidKeyFormat(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"sk_test_abc123", true},
		{"sk_live_xyz", true},
		{"sk_", true},
		{"invalid_key", false},
		{"abc123", false},
		{"", false},
		{"SK_test", false}, // case sensitive
	}

	for _, tt := range tests {
		got := isValidKeyFormat(tt.key)
		if got != tt.want {
			t.Errorf("isValidKeyFormat(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestVerifyAPIKey_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/projects/" {
			auth := r.Header.Get("Authorization")
			if auth == "Bearer sk_test_valid" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	err := verifyAPIKey(server.URL, "sk_test_valid")
	if err != nil {
		t.Errorf("verifyAPIKey() with valid key error = %v, want nil", err)
	}
}

func TestVerifyAPIKey_InvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/projects/" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	err := verifyAPIKey(server.URL, "sk_test_bad")
	if err == nil {
		t.Error("verifyAPIKey() with invalid key expected error, got nil")
	}
}

func TestVerifyAPIKey_ServerDown(t *testing.T) {
	err := verifyAPIKey("http://127.0.0.1:1", "sk_test_key")
	if err == nil {
		t.Error("verifyAPIKey() with unreachable server expected error, got nil")
	}
}

func TestPromptAPIKey_ValidKey(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("sk_test_abc123\n"))

	key, err := promptAPIKey(cmd)
	if err != nil {
		t.Fatalf("promptAPIKey() error = %v", err)
	}
	if key != "sk_test_abc123" {
		t.Errorf("promptAPIKey() = %q, want %q", key, "sk_test_abc123")
	}
}

func TestPromptAPIKey_EmptyKey(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("\n"))

	_, err := promptAPIKey(cmd)
	if err == nil {
		t.Error("promptAPIKey() expected error for empty key")
	}
}

func TestPromptAPIKey_WhitespaceOnly(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("   \n"))

	_, err := promptAPIKey(cmd)
	if err == nil {
		t.Error("promptAPIKey() expected error for whitespace-only key")
	}
}

func TestPromptAPIKey_TrimsWhitespace(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("  sk_test_abc123  \n"))

	key, err := promptAPIKey(cmd)
	if err != nil {
		t.Fatalf("promptAPIKey() error = %v", err)
	}
	if key != "sk_test_abc123" {
		t.Errorf("promptAPIKey() = %q, want %q", key, "sk_test_abc123")
	}
}

func TestPromptYesNo(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"Y\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"\n", false},
		{"anything\n", false},
	}

	for _, tt := range tests {
		cmd := &cobra.Command{}
		cmd.SetIn(strings.NewReader(tt.input))

		got, err := promptYesNo(cmd, "test?")
		if err != nil {
			t.Fatalf("promptYesNo(%q) error = %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("promptYesNo(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
