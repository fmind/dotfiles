package dot

import (
	"context"
	"errors"
	"io"
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

func TestGenerateText(t *testing.T) {
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
