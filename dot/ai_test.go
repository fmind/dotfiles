package dot

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetAIBinary(t *testing.T) {
	t.Run("binary configured", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		state.Config.AI.Binary = "custom-ai"
		if binary := GetAIBinary(state); binary != "custom-ai" {
			t.Errorf("Expected custom-ai, got %q", binary)
		}
	})

	t.Run("binary default", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		state.Config.AI.Binary = ""
		if binary := GetAIBinary(state); binary != "agy" {
			t.Errorf("Expected agy, got %q", binary)
		}
	})
}

func TestScanDiffForSecrets(t *testing.T) {
	t.Run("passes exact diff to gitleaks", func(t *testing.T) {
		const diff = "diff --git a/file b/file\n+safe change\n"
		var scanned string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name != "gitleaks" {
					t.Fatalf("unexpected LookPath(%q)", name)
				}
				return "/bin/gitleaks", nil
			},
			RunFunc: func(_ context.Context, _ string, stdin io.Reader, name string, args ...string) (string, error) {
				if name != "/bin/gitleaks" || strings.Join(args, " ") != "stdin --no-banner --redact" {
					t.Fatalf("unexpected scanner command: %s %v", name, args)
				}
				data, err := io.ReadAll(stdin)
				if err != nil {
					t.Fatal(err)
				}
				scanned = string(data)
				return "", nil
			},
		}

		if err := ScanDiffForSecrets(context.Background(), newTestState(runner), diff); err != nil {
			t.Fatalf("ScanDiffForSecrets failed: %v", err)
		}
		if scanned != diff {
			t.Errorf("scanner received %q, want exact diff %q", scanned, diff)
		}
	})

	t.Run("scanner failure is fatal", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(_ context.Context, _ string, _ io.Reader, _ string, _ ...string) (string, error) {
				return "", errors.New("secret detected")
			},
		}
		err := ScanDiffForSecrets(context.Background(), newTestState(runner), "diff")
		if err == nil || !strings.Contains(err.Error(), "outgoing diff secret scan failed") {
			t.Fatalf("expected secret scan failure, got %v", err)
		}
	})
}

func TestGenerateText(t *testing.T) {
	t.Run("agy uses sandbox in an isolated workspace", func(t *testing.T) {
		var workDir string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(_ context.Context, dir string, _ io.Reader, name string, args ...string) (string, error) {
				if name != "/bin/agy" || strings.Join(args, " ") != "--sandbox --prompt prompt" {
					t.Fatalf("unexpected AI command: %s %v", name, args)
				}
				if info, err := os.Stat(dir); err != nil || !info.IsDir() {
					t.Fatalf("AI workspace is unavailable: %v", err)
				}
				workDir = dir
				if err := os.WriteFile(filepath.Join(dir, "provider-artifact"), []byte("temporary"), 0o600); err != nil {
					t.Fatalf("create provider artifact: %v", err)
				}
				return "generated message", nil
			},
		}

		if _, err := GenerateText(context.Background(), newTestState(runner), "prompt", "input", 100); err != nil {
			t.Fatalf("GenerateText failed: %v", err)
		}
		if workDir == "" {
			t.Fatal("AI command did not receive an isolated workspace")
		}
		if _, err := os.Stat(workDir); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("isolated AI workspace was not removed: %v", err)
		}
	})

	t.Run("custom binary does not receive agy flags", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(_ context.Context, dir string, _ io.Reader, name string, args ...string) (string, error) {
				if dir == "" {
					t.Fatal("custom provider did not receive an isolated workspace")
				}
				if name != "/bin/custom-ai" || strings.Join(args, " ") != "--prompt prompt" {
					t.Fatalf("unexpected custom AI command: %s %v", name, args)
				}
				return "generated message", nil
			},
		}
		state := newTestState(runner)
		state.Config.AI.Binary = "custom-ai"

		if _, err := GenerateText(context.Background(), state, "prompt", "input", 100); err != nil {
			t.Fatalf("GenerateText failed: %v", err)
		}
	})

	t.Run("generate success", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "agy" {
					return "/bin/agy", nil
				}
				return "", errors.New("not found")
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "/bin/agy" {
					return "generated message", nil
				}
				return "", errors.New("unexpected command")
			},
		}
		state := newTestState(runner)

		out, err := GenerateText(context.Background(), state, "prompt", "input", 100)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if out != "generated message" {
			t.Errorf("Expected 'generated message', got %q", out)
		}
	})

	t.Run("generate trims output", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/agy", nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "  trimmed message  \n", nil
			},
		}
		state := newTestState(runner)

		out, err := GenerateText(context.Background(), state, "prompt", "input", 100)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if out != "trimmed message" {
			t.Errorf("Expected 'trimmed message', got %q", out)
		}
	})

	t.Run("generate limits input size", func(t *testing.T) {
		var inputRead string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/agy", nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				data, _ := io.ReadAll(stdin)
				inputRead = string(data)
				return "ok", nil
			},
		}
		state := newTestState(runner)

		_, err := GenerateText(context.Background(), state, "prompt", "abcdefghij", 5)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if inputRead != "abcde" {
			t.Errorf("Expected input to be limited to 5 chars, got %q", inputRead)
		}
	})

	t.Run("binary not in path error", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("not found")
			},
		}
		state := newTestState(runner)

		_, err := GenerateText(context.Background(), state, "prompt", "input", 100)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !errors.Is(err, ErrToolNotInstalled) {
			t.Errorf("Expected error to wrap ErrToolNotInstalled, got %q", err.Error())
		}
	})

	t.Run("maxSize <= 0 uses default limit", func(t *testing.T) {
		var inputRead string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/agy", nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				data, _ := io.ReadAll(stdin)
				inputRead = string(data)
				return "ok", nil
			},
		}
		state := newTestState(runner)
		// Send input slightly longer than default 200000 limit, say 200005 characters
		longInput := strings.Repeat("a", 200005)
		_, err := GenerateText(context.Background(), state, "prompt", longInput, 0)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(inputRead) != 200000 {
			t.Errorf("Expected input to be limited to 200000, got %d", len(inputRead))
		}
	})

	t.Run("empty output returns error", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/agy", nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "   \n  ", nil
			},
		}
		state := newTestState(runner)

		_, err := GenerateText(context.Background(), state, "prompt", "input", 100)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "AI returned empty output") {
			t.Errorf("Expected 'AI returned empty output', got %q", err.Error())
		}
	})
}
