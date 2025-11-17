package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evcraddock/todu.sh/internal/api"
)

func TestEnsureLocalSystem_CreatesWhenNotExist(t *testing.T) {
	// Setup mock server that returns empty systems list first, then creates
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/systems/":
			// First call: list systems (empty)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/systems/":
			// Second call: create system
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id": 1, "identifier": "local", "name": "Local Tasks", "created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	systemID, err := ensureLocalSystem(client)

	if err != nil {
		t.Errorf("ensureLocalSystem() error = %v, want nil", err)
	}
	if systemID != 1 {
		t.Errorf("ensureLocalSystem() = %d, want 1", systemID)
	}
	if callCount != 1 {
		t.Errorf("CreateSystem called %d times, want 1", callCount)
	}
}

func TestEnsureLocalSystem_ReturnsExistingID(t *testing.T) {
	// Setup mock server that returns existing local system
	createCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/systems/":
			// Return existing local system
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 42, "identifier": "local", "name": "Local Tasks", "created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/systems/":
			createCalled = true
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	systemID, err := ensureLocalSystem(client)

	if err != nil {
		t.Errorf("ensureLocalSystem() error = %v, want nil", err)
	}
	if systemID != 42 {
		t.Errorf("ensureLocalSystem() = %d, want 42", systemID)
	}
	if createCalled {
		t.Error("CreateSystem should not be called when local system exists")
	}
}

func TestResolveSystemID_WithNumericID(t *testing.T) {
	// No server needed for numeric ID
	client := api.NewClient("http://unused")
	id, err := resolveSystemID(client, "123")

	if err != nil {
		t.Errorf("resolveSystemID() error = %v, want nil", err)
	}
	if id != 123 {
		t.Errorf("resolveSystemID() = %d, want 123", id)
	}
}

func TestResolveSystemID_WithIdentifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/systems/" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{"id": 1, "identifier": "github", "name": "GitHub", "created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z"},
				{"id": 2, "identifier": "forgejo", "name": "Forgejo", "created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z"}
			]`))
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	id, err := resolveSystemID(client, "forgejo")

	if err != nil {
		t.Errorf("resolveSystemID() error = %v, want nil", err)
	}
	if id != 2 {
		t.Errorf("resolveSystemID() = %d, want 2", id)
	}
}

func TestResolveSystemID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/systems/" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 1, "identifier": "github", "name": "GitHub", "created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z"}]`))
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	_, err := resolveSystemID(client, "nonexistent")

	if err == nil {
		t.Error("resolveSystemID() expected error for nonexistent system")
	}
}

func TestResolveSystemID_EmptyString(t *testing.T) {
	client := api.NewClient("http://unused")
	id, err := resolveSystemID(client, "")

	if err != nil {
		t.Errorf("resolveSystemID() error = %v, want nil", err)
	}
	if id != 0 {
		t.Errorf("resolveSystemID() = %d, want 0", id)
	}
}

// Ensure context is used (compile-time check)
var _ = context.Background()
