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

type failingPRDescriptionFile struct {
	writeErr error
	closeErr error
}

func (f failingPRDescriptionFile) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(p), nil
}

func (f failingPRDescriptionFile) Close() error {
	return f.closeErr
}

func TestWritePRDescription(t *testing.T) {
	t.Run("close failure", func(t *testing.T) {
		closeErr := errors.New("close failed")
		err := writePRDescription(failingPRDescriptionFile{closeErr: closeErr}, "description")
		if !errors.Is(err, closeErr) {
			t.Fatalf("expected close error, got %v", err)
		}
	})

	t.Run("write and close failures", func(t *testing.T) {
		writeErr := errors.New("write failed")
		closeErr := errors.New("close failed")
		err := writePRDescription(failingPRDescriptionFile{writeErr: writeErr, closeErr: closeErr}, "description")
		if !errors.Is(err, writeErr) || !errors.Is(err, closeErr) {
			t.Fatalf("expected joined write and close errors, got %v", err)
		}
	})
}

func TestFindPRTemplateSkipsDirectories(t *testing.T) {
	// GitHub also supports a PULL_REQUEST_TEMPLATE/ directory of templates; a directory
	// candidate must be skipped rather than aborting on the EISDIR read.
	templateDir := filepath.Join(t.TempDir(), "PULL_REQUEST_TEMPLATE")
	if err := os.Mkdir(templateDir, 0o700); err != nil {
		t.Fatal(err)
	}

	content, found, err := findPRTemplate([]string{templateDir})
	if err != nil || found || content != "" {
		t.Fatalf("expected directory to be skipped, got content=%q found=%v err=%v", content, found, err)
	}
}

func TestFindPRTemplateReportsReadErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-denied read cannot be simulated as root")
	}
	templatePath := filepath.Join(t.TempDir(), "pull_request_template.md")
	if err := os.WriteFile(templatePath, []byte("body"), 0o000); err != nil {
		t.Fatal(err)
	}

	_, found, err := findPRTemplate([]string{templatePath})
	if err == nil || found || !strings.Contains(err.Error(), "failed to read PR template") {
		t.Fatalf("expected template read error, got found=%v err=%v", found, err)
	}
}

func TestWithPRDescriptionFileReportsCleanupFailure(t *testing.T) {
	var tempPath string
	err := withPRDescriptionFile("description", func(path string) error {
		tempPath = path
		if removeErr := os.Remove(path); removeErr != nil {
			return removeErr
		}
		if mkdirErr := os.Mkdir(path, 0o700); mkdirErr != nil {
			return mkdirErr
		}
		return os.WriteFile(filepath.Join(path, "leftover"), []byte("data"), 0o600)
	})
	if tempPath != "" {
		t.Cleanup(func() { _ = os.RemoveAll(tempPath) })
	}
	if err == nil || !strings.Contains(err.Error(), "failed to remove temporary PR description") {
		t.Fatalf("expected cleanup failure, got %v", err)
	}
}

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

	t.Run("excluded-only changes are not reported clean", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) { return "/usr/bin/" + name, nil },
			RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "rev-parse" {
					return "true", nil
				}
				if name == "git" && args[0] == "diff" {
					if len(args) == 4 {
						return "unfiltered lockfile diff", nil
					}
					return "", nil
				}
				return "", nil
			},
		}
		err := RunPr(ctx, newTestState(runner), nil, "main")
		if err == nil || !strings.Contains(err.Error(), "every changed path is excluded") {
			t.Fatalf("expected excluded-diff error, got %v", err)
		}
	})

	t.Run("secret scan failure prevents AI and gh", func(t *testing.T) {
		aiCalled := false
		ghCalled := false
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) { return "/usr/bin/" + name, nil },
			RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "rev-parse" {
					return "true", nil
				}
				if name == "git" && args[0] == "diff" {
					return "diff containing a secret", nil
				}
				if name == "/usr/bin/gitleaks" {
					return "", errors.New("secret detected")
				}
				if name == "/usr/bin/agy" {
					aiCalled = true
				}
				return "", nil
			},
			RunInteractiveFunc: func(_ context.Context, _, _ string, _ ...string) error {
				ghCalled = true
				return nil
			},
		}
		err := RunPr(ctx, newTestState(runner), nil, "main")
		if err == nil || !strings.Contains(err.Error(), "outgoing diff secret scan failed") {
			t.Fatalf("expected secret scan error, got %v", err)
		}
		if aiCalled || ghCalled {
			t.Fatalf("AI or gh ran after failed scan: ai=%v gh=%v", aiCalled, ghCalled)
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
