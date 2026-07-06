package dot

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestGetCompletionCommand(t *testing.T) {
	tests := []struct {
		tool           string
		expectedBinary string
		expectedArgs   []string
	}{
		{"git-lfs", "git", []string{"lfs", "completion", "fish"}},
		{"gh", "gh", []string{"completion", "-s", "fish"}},
		{"kubectl", "kubectl", []string{"completion", "fish"}},
		{"uv", "uv", []string{"generate-shell-completion", "fish"}},
		{"ast-grep", "ast-grep", []string{"completions", "fish"}},
		{"atlas", "atlas", []string{"completion", "fish"}},
		{"atuin", "atuin", []string{"gen-completions", "--shell", "fish"}},
		{"bat", "bat", []string{"--completion", "fish"}},
		{"carapace", "carapace", []string{"_carapace", "fish"}},
		{"delta", "delta", []string{"--generate-completion", "fish"}},
		{"just", "just", []string{"--completions", "fish"}},
		{"rg", "rg", []string{"--generate", "complete-fish"}},
		{"ruff", "ruff", []string{"generate-shell-completion", "fish"}},
		{"starship", "starship", []string{"completions", "fish"}},
		{"stern", "stern", []string{"--completion", "fish"}},
		{"watchexec", "watchexec", []string{"--completions", "fish"}},
		{"cosign", "cosign", []string{"completion", "fish"}}, // default command, no custom entry
		{"xh", "xh", []string{"--generate", "complete-fish"}},
		{"yq", "yq", []string{"shell-completion", "fish"}},
		{"zellij", "zellij", []string{"setup", "--generate-completion", "fish"}},
		{"unknown-tool", "unknown-tool", []string{"completion", "fish"}}, // default fallback
	}

	for _, tc := range tests {
		state := newTestState(&FakeRunner{})
		binary, args := GetCompletionCommand(state, tc.tool)
		if binary != tc.expectedBinary {
			t.Errorf("GetCompletionCommand(%q) returned binary %q, expected %q", tc.tool, binary, tc.expectedBinary)
		}
		if len(args) != len(tc.expectedArgs) {
			t.Errorf("GetCompletionCommand(%q) returned args of length %d, expected %d", tc.tool, len(args), len(tc.expectedArgs))
			continue
		}
		for i := range args {
			if args[i] != tc.expectedArgs[i] {
				t.Errorf("GetCompletionCommand(%q) arg[%d] = %q, expected %q", tc.tool, i, args[i], tc.expectedArgs[i])
			}
		}
	}
}

func TestCompletionCommand(t *testing.T) {
	// Create a temp home directory so it writes completions there
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Mock runner where LookPath always finds the tool, and Run always succeeds
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			return "# completion output for " + name, nil
		},
	}

	state := newTestState(runner)
	state.Config.Completions.Tools = []string{"gh", "uv"}

	app := &cli.Command{
		Name: "dot",
		Commands: []*cli.Command{
			NewCompletionCmd(state),
		},
	}

	err := app.Run(context.Background(), []string{"dot", "completion"})
	if err != nil {
		t.Fatalf("Expected no error running completion, got %v", err)
	}

	// Verify files were written to ~/.config/fish/completions/
	compDir := filepath.Join(tempDir, ".config", "fish", "completions")
	for _, tool := range []string{"gh", "uv", "dot"} {
		filePath := filepath.Join(compDir, tool+".fish")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected completions file to exist: %s", filePath)
		} else {
			content, _ := os.ReadFile(filePath)
			if tool != "dot" && !strings.Contains(string(content), "completion output for") {
				t.Errorf("Completions file content incorrect for %s: %s", tool, string(content))
			}
		}
	}
}

func TestGenerateToolCompletion(t *testing.T) {
	t.Run("tool not installed", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", os.ErrNotExist
			},
		}
		_, err := generateToolCompletion(context.Background(), newTestState(runner), "missing")
		if !errors.Is(err, ErrToolNotInstalled) {
			t.Errorf("Expected ErrToolNotInstalled, got %v", err)
		}
	})

	t.Run("generation success on primary command", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && len(args) == 3 && args[0] == "lfs" && args[1] == "completion" && args[2] == "fish" {
					return "git-lfs-fish-completions", nil
				}
				return "", errors.New("unexpected command")
			},
		}
		out, err := generateToolCompletion(context.Background(), newTestState(runner), "git-lfs")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if out != "git-lfs-fish-completions" {
			t.Errorf("Expected 'git-lfs-fish-completions', got %q", out)
		}
	})

	t.Run("generation fallback success", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "ast-grep" {
					if len(args) == 2 && args[0] == "completions" && args[1] == "fish" {
						return "", errors.New("unsupported command")
					}
					if len(args) == 2 && args[0] == "completion" && args[1] == "fish" {
						return "ast-grep-fallback-completions", nil
					}
				}
				return "", errors.New("unexpected command")
			},
		}
		out, err := generateToolCompletion(context.Background(), newTestState(runner), "ast-grep")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if out != "ast-grep-fallback-completions" {
			t.Errorf("Expected 'ast-grep-fallback-completions', got %q", out)
		}
	})

	t.Run("generation fails completely", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "", errors.New("command failed")
			},
		}
		_, err := generateToolCompletion(context.Background(), newTestState(runner), "ast-grep")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to generate completions") {
			t.Errorf("Expected error to contain 'failed to generate completions', got %v", err)
		}
	})
}

func TestGetCompletionCommand_CustomAndConfigs(t *testing.T) {
	state := newTestState(&FakeRunner{})
	state.Config.Completions.CustomCommands = map[string]ToolConfig{
		"my-custom-tool": {
			Args: []string{"gen", "fish"},
		},
		"other-custom": {
			Binary: "custom-binary",
			Args:   []string{"completions"},
		},
	}

	t.Run("custom tool default binary name", func(t *testing.T) {
		bin, args := GetCompletionCommand(state, "my-custom-tool")
		if bin != "my-custom-tool" {
			t.Errorf("Expected binary 'my-custom-tool', got %q", bin)
		}
		if len(args) != 2 || args[0] != "gen" || args[1] != "fish" {
			t.Errorf("Expected args ['gen', 'fish'], got %v", args)
		}
	})

	t.Run("custom tool explicit binary name", func(t *testing.T) {
		bin, args := GetCompletionCommand(state, "other-custom")
		if bin != "custom-binary" {
			t.Errorf("Expected binary 'custom-binary', got %q", bin)
		}
		if len(args) != 1 || args[0] != "completions" {
			t.Errorf("Expected args ['completions'], got %v", args)
		}
	})
}

func TestCompletionCommand_Errors(t *testing.T) {
	t.Run("failed to create completions directory", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		// Use a path that is a file to cause MkdirAll to fail
		state.Config.Completions.Path = "/dev/null"
		err := RunCompletionGenerate(context.Background(), state)
		if err == nil || !strings.Contains(err.Error(), "failed to create completions directory") {
			t.Errorf("Expected directory creation failure, got %v", err)
		}
	})

	t.Run("failed to write fish file skips but continues", func(t *testing.T) {
		tempDir := t.TempDir()
		state := newTestState(&FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "completion-data", nil
			},
		})
		state.Config.Completions.Path = tempDir
		state.Config.Completions.Tools = []string{"gh"}

		// Make gh.fish a directory to make writing gh.fish fail
		err := os.MkdirAll(filepath.Join(tempDir, "gh.fish"), 0o755)
		if err != nil {
			t.Fatalf("Failed to create sub-dir: %v", err)
		}

		var buf bytes.Buffer
		state.Stdout = &buf

		err = RunCompletionGenerate(context.Background(), state)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to write completions for gh") {
			t.Errorf("Expected error to mention gh write failure, got %v", err)
		}
	})

	t.Run("canceled context returns error without reporting success", func(t *testing.T) {
		tempDir := t.TempDir()
		state := newTestState(&FakeRunner{
			LookPathFunc: func(name string) (string, error) { return "/bin/" + name, nil },
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "completion-data", nil
			},
		})
		state.Config.Completions.Path = tempDir
		state.Config.Completions.Tools = []string{"gh", "uv"}

		var buf bytes.Buffer
		state.Stdout = &buf

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // interrupt before any work runs

		err := RunCompletionGenerate(ctx, state)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Expected context.Canceled, got %v", err)
		}
		if strings.Contains(buf.String(), "Completions updated") {
			t.Errorf("Should not print success on cancellation, got: %s", buf.String())
		}
		// dot.fish must not be written once the run is canceled.
		if _, statErr := os.Stat(filepath.Join(tempDir, "dot.fish")); !os.IsNotExist(statErr) {
			t.Errorf("dot.fish should not be written on cancellation")
		}
	})
}
