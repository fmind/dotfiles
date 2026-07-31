package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	// Local package
	"dot"
)

type fakeApp struct {
	err error
}

func (a fakeApp) Run(context.Context, []string) error {
	return a.err
}

func TestAppRuns(t *testing.T) {
	app := dot.NewApp()

	// Test help command works and runs cleanly
	err := app.Run(context.Background(), []string{"dot", "help"})
	if err != nil {
		t.Fatalf("Expected no error running dot help, got %v", err)
	}
}

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		err        error
		name       string
		wantStderr string
		wantCode   int
	}{
		{name: "success", wantCode: 0},
		{name: "interrupted", err: context.Canceled, wantCode: 130},
		{name: "failure", err: errors.New("boom"), wantCode: 1, wantStderr: "dot: boom\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			code := run(context.Background(), fakeApp{err: tt.err}, []string{"dot"}, &stderr)
			if code != tt.wantCode {
				t.Fatalf("run returned %d, want %d", code, tt.wantCode)
			}
			if stderr.String() != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}
