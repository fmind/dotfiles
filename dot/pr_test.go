package dot

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestRunPr_ErrorsAndBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("not inside work tree", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "rev-parse" {
					return "", errors.New("not git repository")
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		err := RunPr(ctx, state, nil, "main")
		if !errors.Is(err, ErrNotGitRepository) {
			t.Errorf("Expected ErrNotGitRepository, got %v", err)
		}
	})

	t.Run("gh not installed", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "gh" {
					return "", errors.New("not found")
				}
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "", nil // is git worktree, etc
			},
		}
		state := newTestState(runner)
		err := RunPr(ctx, state, nil, "main")
		if !errors.Is(err, ErrGhNotInstalled) {
			t.Errorf("Expected ErrGhNotInstalled, got %v", err)
		}
	})

	t.Run("get base diff fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "diff" {
					return "", errors.New("git diff error")
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		err := RunPr(ctx, state, nil, "main")
		if err == nil || !strings.Contains(err.Error(), "failed to get git diff against main") {
			t.Errorf("Expected git diff error, got %v", err)
		}
	})

	t.Run("no changes detected", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "", nil
			},
		}
		var buf bytes.Buffer
		state := newTestState(runner)
		state.Stdout = &buf
		err := RunPr(ctx, state, nil, "main")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !strings.Contains(buf.String(), "No changes detected against base branch 'main'") {
			t.Errorf("Expected message about no changes, got %q", buf.String())
		}
	})

	t.Run("AI generation fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "diff" {
					return "some changes", nil
				}
				if name == "/usr/bin/agy" {
					return "", errors.New("agy failed")
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		err := RunPr(ctx, state, nil, "main")
		if err == nil || !strings.Contains(err.Error(), "AI invocation failed") {
			t.Errorf("Expected AI invocation failed error, got %v", err)
		}
	})

	t.Run("gh pr create fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "diff" {
					return "some changes", nil
				}
				if name == "/usr/bin/agy" {
					return "AI description output", nil
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "/usr/bin/gh" {
					return errors.New("gh fail")
				}
				return nil
			},
		}
		state := newTestState(runner)
		err := RunPr(ctx, state, nil, "main")
		if err == nil || !strings.Contains(err.Error(), "gh pr create failed") {
			t.Errorf("Expected gh pr create failed error, got %v", err)
		}
	})
}

func TestNewPrCmd_Flags(t *testing.T) {
	ctx := context.Background()

	t.Run("explicit base flag", func(t *testing.T) {
		var passedBaseBranch string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "diff" {
					// Extract base branch being diffed against
					// args are usually like: "diff", "branch...", "."
					passedBaseBranch = strings.TrimSuffix(args[1], "...")
					return "diff content", nil
				}
				if strings.Contains(name, "agy") {
					return "AI description output", nil
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				return nil
			},
		}
		state := newTestState(runner)
		cmd := NewPrCmd(state)
		app := &cli.Command{
			Commands: []*cli.Command{cmd},
		}
		err := app.Run(ctx, []string{"dot", "pr", "--base", "feature-branch"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if passedBaseBranch != "feature-branch" {
			t.Errorf("Expected base branch to be 'feature-branch', got %q", passedBaseBranch)
		}
	})

	t.Run("base from config", func(t *testing.T) {
		var passedBaseBranch string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "diff" {
					passedBaseBranch = strings.TrimSuffix(args[1], "...")
					return "diff content", nil
				}
				if strings.Contains(name, "agy") {
					return "AI description output", nil
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				return nil
			},
		}
		state := newTestState(runner)
		state.Config.PR.BaseBranch = "develop"
		cmd := NewPrCmd(state)
		// We have to clear/unset the base flag string value, or since it has a default Value: "main",
		// cCtx.IsSet("base") checks if the user provided it. In NewPrCmd:
		// baseBranch := cCtx.String("base")
		// if !cCtx.IsSet("base") && state.Config.PR.BaseBranch != "" {
		//     baseBranch = state.Config.PR.BaseBranch
		// }
		// So if user does NOT set --base, cCtx.IsSet("base") is false, and it will fall back to state.Config.PR.BaseBranch.
		app := &cli.Command{
			Commands: []*cli.Command{cmd},
		}
		err := app.Run(ctx, []string{"dot", "pr"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if passedBaseBranch != "develop" {
			t.Errorf("Expected base branch to be 'develop', got %q", passedBaseBranch)
		}
	})
}

func TestRunPr_CustomTemplates(t *testing.T) {
	ctx := context.Background()

	t.Run("custom templates from config", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "custom_template_*.md")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		expectedContent := "This is a custom template content!"
		if _, writeErr := tmpFile.WriteString(expectedContent); writeErr != nil {
			t.Fatalf("failed to write to temp file: %v", writeErr)
		}
		_ = tmpFile.Close()

		var passedPrompt string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "diff" {
					return "diff content", nil
				}
				if strings.Contains(name, "agy") {
					for i, arg := range args {
						if arg == "--prompt" && i+1 < len(args) {
							passedPrompt = args[i+1]
						}
					}
					return "AI description output", nil
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				return nil
			},
		}

		state := newTestState(runner)
		state.Config.PR.Templates = []string{tmpFile.Name()}

		if err := RunPr(ctx, state, nil, "main"); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !strings.Contains(passedPrompt, expectedContent) {
			t.Errorf("Expected prompt to contain custom template content %q, but got:\n%s", expectedContent, passedPrompt)
		}
	})
}
