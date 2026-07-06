package dot

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestStandardRunner_LookPath(t *testing.T) {
	runner := NewStandardRunner(os.Stdin, os.Stdout, os.Stderr)
	path, err := runner.LookPath("go")
	if err != nil {
		t.Skip("go not in PATH, skipping LookPath test")
	}
	if path == "" {
		t.Error("Expected non-empty path for go")
	}
}

func TestStandardRunner_Run(t *testing.T) {
	runner := NewStandardRunner(os.Stdin, os.Stdout, os.Stderr)
	out, err := runner.Run(context.Background(), "", nil, "echo", "hello-world")
	if err != nil {
		t.Fatalf("Expected no error running echo, got %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "hello-world" {
		t.Errorf("Expected output 'hello-world', got %q", trimmed)
	}

	// Test command failure returns error
	_, err = runner.Run(context.Background(), "", nil, "false")
	if err == nil {
		t.Error("Expected error when running false, got nil")
	}

	// Test non-empty dir
	tempDir := t.TempDir()
	out, err = runner.Run(context.Background(), tempDir, nil, "pwd")
	if err != nil {
		t.Fatalf("Expected no error running pwd in tempDir, got %v", err)
	}
	if !strings.Contains(out, "tmp") && !strings.Contains(out, "temp") && !strings.Contains(out, "/") {
		t.Errorf("Expected pwd output to contain directory path structure, got %q", out)
	}
}

func TestStandardRunner_RunWithStdin(t *testing.T) {
	runner := NewStandardRunner(os.Stdin, os.Stdout, os.Stderr)
	stdin := strings.NewReader("piped-input-data")
	out, err := runner.Run(context.Background(), "", stdin, "cat")
	if err != nil {
		t.Fatalf("Expected no error running cat with stdin, got %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "piped-input-data" {
		t.Errorf("Expected output 'piped-input-data', got %q", trimmed)
	}
}

func TestStandardRunner_RunInteractive(t *testing.T) {
	rIn, wIn, _ := os.Pipe()
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	runner := &StandardRunner{
		Stdin:  rIn,
		Stdout: wOut,
		Stderr: wErr,
	}

	// Close write end of stdin immediately so it returns EOF
	_ = wIn.Close()

	err := runner.RunInteractive(context.Background(), t.TempDir(), "true")
	_ = wOut.Close()
	_ = wErr.Close()

	if err != nil {
		t.Fatalf("Expected no error running true interactively, got %v", err)
	}

	// Clean up pipes
	_ = rIn.Close()
	_ = rOut.Close()
	_ = rErr.Close()
}
