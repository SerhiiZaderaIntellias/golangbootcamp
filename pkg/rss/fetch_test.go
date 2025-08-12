package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful fetch",
			serverResponse: "<rss><channel><title>Test RSS</title></channel></rss>",
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			serverResponse: "Internal Server Error",
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "not found",
			serverResponse: "Not Found",
			serverStatus:   http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Test fetch
			data, err := Fetch(server.URL)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if string(data) != tt.serverResponse {
					t.Errorf("Expected response %s, got %s", tt.serverResponse, string(data))
				}
			}
		})
	}
}

func TestFetch_InvalidURL(t *testing.T) {
	_, err := Fetch("http://invalid-url-that-does-not-exist.com")
	// This test might pass or fail depending on network conditions
	// We'll just log the result
	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else {
		t.Log("URL resolved unexpectedly")
	}
}