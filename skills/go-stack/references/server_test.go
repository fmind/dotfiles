package <slug>

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"<import_path>/config"
)

func TestNewAppHandler(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	handler := NewAppHandler(logger, config.Development)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}
}
