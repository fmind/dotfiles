package main

import (
	"context"
	"testing"

	// Local package
	"dot"
)

func TestAppRuns(t *testing.T) {
	app := dot.NewApp()

	// Test help command works and runs cleanly
	err := app.Run(context.Background(), []string{"dot", "help"})
	if err != nil {
		t.Fatalf("Expected no error running dot help, got %v", err)
	}
}
