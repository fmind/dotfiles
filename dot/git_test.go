package dot

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsInsideWorkTree(t *testing.T) {
	ctx := context.Background()

	t.Run("inside work tree", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "true", nil
			},
		}
		state := newTestState(runner)
		err := IsInsideWorkTree(ctx, state)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("outside work tree", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "", errors.New("exit status 128")
			},
		}
		state := newTestState(runner)
		err := IsInsideWorkTree(ctx, state)
		if !errors.Is(err, ErrNotGitRepository) {
			t.Errorf("Expected ErrNotGitRepository, got %v", err)
		}
	})
}

func TestGetCachedDiff(t *testing.T) {
	ctx := context.Background()
	runner := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if len(args) > 1 && args[0] == "diff" && args[1] == "--cached" {
				return "cached diff content", nil
			}
			return "", errors.New("unexpected command")
		},
	}
	state := newTestState(runner)
	diff, err := GetCachedDiff(ctx, state)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if diff != "cached diff content" {
		t.Errorf("Expected 'cached diff content', got %q", diff)
	}
}

func TestGetUnstagedDiff(t *testing.T) {
	ctx := context.Background()
	runner := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if len(args) > 0 && args[0] == "diff" && args[1] != "--cached" {
				return "unstaged diff content", nil
			}
			return "", errors.New("unexpected command")
		},
	}
	state := newTestState(runner)
	diff, err := GetUnstagedDiff(ctx, state)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if diff != "unstaged diff content" {
		t.Errorf("Expected 'unstaged diff content', got %q", diff)
	}
}

// TestGitDiffIntegration exercises the real pathspec/merge-base logic against an actual
// git repo (the lightest real resource) rather than canned FakeRunner output, so a broken
// :(exclude) pathspec or merge-base invocation would be caught.
func TestGitDiffIntegration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	ctx := context.Background()
	repo := t.TempDir()
	// The diff helpers run git in the process working directory (dir=""), so point it at the repo.
	t.Chdir(repo)
	runner := NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard)
	state := newTestState(runner)

	git := func(args ...string) {
		t.Helper()
		if _, err := runner.Run(ctx, repo, nil, "git", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(repo, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	git("init", "-b", "main")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	git("config", "commit.gpgsign", "false")
	write("base.txt", "base\n")
	git("add", ".")
	git("commit", "-m", "base")

	// On a feature branch, add a normal file and an excluded (go.sum) file.
	git("checkout", "-b", "feature")
	write("feature.txt", "feature\n")
	write("go.sum", "should be excluded\n")
	git("add", ".")
	git("commit", "-m", "feature")

	// GetBaseDiff must see the feature file via the merge-base (...) path,
	// and must exclude go.sum via the default ExcludeDiff pathspec.
	diff, err := GetBaseDiff(ctx, state, "main")
	if err != nil {
		t.Fatalf("GetBaseDiff: %v", err)
	}
	if !strings.Contains(diff, "feature.txt") {
		t.Errorf("expected feature.txt in base diff, got:\n%s", diff)
	}
	if strings.Contains(diff, "go.sum") {
		t.Errorf("expected go.sum to be excluded from base diff, got:\n%s", diff)
	}

	// GetCachedDiff must pick up a staged change.
	write("base.txt", "changed\n")
	git("add", "base.txt")
	cached, err := GetCachedDiff(ctx, state)
	if err != nil {
		t.Fatalf("GetCachedDiff: %v", err)
	}
	if !strings.Contains(cached, "base.txt") {
		t.Errorf("expected base.txt in cached diff, got:\n%s", cached)
	}
}

// TestGitDiffFromSubdirectory guards the ":/" (repo-root) pathspec: the diff helpers must
// cover the whole repo even when invoked from a subdirectory, matching the scope of the
// `git commit`/`gh pr create` that consume the generated message. A cwd-relative "." would
// miss the root-level change here.
func TestGitDiffFromSubdirectory(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	ctx := context.Background()
	repo := t.TempDir()
	runner := NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard)
	state := newTestState(runner)

	git := func(args ...string) {
		t.Helper()
		if _, err := runner.Run(ctx, repo, nil, "git", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(repo, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "root.txt"), []byte("root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("init", "-b", "main")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	git("config", "commit.gpgsign", "false")
	git("add", ".")
	git("commit", "-m", "base")

	// Stage a change to the ROOT file, then run the helper from the SUBDIRECTORY.
	if err := os.WriteFile(filepath.Join(repo, "root.txt"), []byte("root changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "root.txt")
	t.Chdir(filepath.Join(repo, "sub"))

	cached, err := GetCachedDiff(ctx, state)
	if err != nil {
		t.Fatalf("GetCachedDiff: %v", err)
	}
	if !strings.Contains(cached, "root.txt") {
		t.Errorf("expected root.txt in cached diff run from a subdirectory, got:\n%s", cached)
	}
}

func TestGetBaseDiff(t *testing.T) {
	ctx := context.Background()

	t.Run("merge base succeeds", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if len(args) > 1 && args[0] == "diff" && args[1] == "main..." {
					return "merge base diff", nil
				}
				return "", errors.New("unexpected command")
			},
		}
		state := newTestState(runner)
		diff, err := GetBaseDiff(ctx, state, "main")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if diff != "merge base diff" {
			t.Errorf("Expected 'merge base diff', got %q", diff)
		}
	})

	t.Run("fallback to direct diff succeeds", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if len(args) > 1 && args[0] == "diff" {
					if args[1] == "main..." {
						return "", errors.New("merge base diff failed")
					}
					if args[1] == "main" {
						return "direct diff", nil
					}
				}
				return "", errors.New("unexpected command")
			},
		}
		state := newTestState(runner)
		diff, err := GetBaseDiff(ctx, state, "main")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if diff != "direct diff" {
			t.Errorf("Expected 'direct diff', got %q", diff)
		}
	})

	t.Run("all diff attempts fail", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "", errors.New("diff failed")
			},
		}
		state := newTestState(runner)
		_, err := GetBaseDiff(ctx, state, "main")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to get git diff against main") {
			t.Errorf("Expected custom error message, got %v", err)
		}
	})
}
