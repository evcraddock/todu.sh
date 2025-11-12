package api

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8000"
	client := NewClient(baseURL)

	if client == nil {
		t.Fatal("Expected NewClient to return a non-nil client")
	}
}

func TestNewClientStoresBaseURL(t *testing.T) {
	baseURL := "http://localhost:8000"
	client := NewClient(baseURL)

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL to be '%s', got '%s'", baseURL, client.baseURL)
	}
}

func TestNewClientInitializesHTTPClient(t *testing.T) {
	client := NewClient("http://localhost:8000")

	if client.httpClient == nil {
		t.Fatal("Expected httpClient to be initialized")
	}
}

func TestHTTPClientTimeout(t *testing.T) {
	client := NewClient("http://localhost:8000")

	expectedTimeout := 30 * time.Second
	if client.httpClient.Timeout != expectedTimeout {
		t.Errorf("Expected HTTP client timeout to be %v, got %v", expectedTimeout, client.httpClient.Timeout)
	}
}

func TestNewClientWithDifferentBaseURLs(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "localhost with port",
			baseURL: "http://localhost:8000",
		},
		{
			name:    "localhost without port",
			baseURL: "http://localhost",
		},
		{
			name:    "remote server",
			baseURL: "https://api.example.com",
		},
		{
			name:    "remote server with path",
			baseURL: "https://api.example.com/v1",
		},
		{
			name:    "IP address",
			baseURL: "http://192.168.1.100:8080",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(tc.baseURL)

			if client == nil {
				t.Fatal("Expected NewClient to return a non-nil client")
			}

			if client.baseURL != tc.baseURL {
				t.Errorf("Expected baseURL to be '%s', got '%s'", tc.baseURL, client.baseURL)
			}

			if client.httpClient == nil {
				t.Error("Expected httpClient to be initialized")
			}
		})
	}
}

func TestClientFieldsArePrivate(t *testing.T) {
	// This test verifies that the Client struct has the expected private fields
	// by creating a client and checking that it works as expected
	client := NewClient("http://localhost:8000")

	// Verify that we can create a client (fields are accessible internally)
	if client.baseURL == "" {
		t.Error("Expected baseURL to be set")
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be set")
	}
}
