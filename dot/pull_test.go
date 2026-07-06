package dot

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// pushRunner returns a FakeRunner for a single repo that is `ahead` commits ahead of
// its upstream, counting how many times `git push` is invoked.
func pushRunner(dirty bool, ahead string, pushed *int32) *FakeRunner {
	return &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name != "git" {
				return "", nil
			}
			switch args[0] {
			case "branch":
				return "main\n", nil
			case "status":
				if dirty {
					return "M file.go\n", nil
				}
				return "", nil
			case "rev-list":
				if len(args) > 2 && args[2] == "@{u}..HEAD" {
					return ahead + "\n", nil // commits ahead of upstream
				}
				return "0\n", nil // commits behind upstream
			case "push":
				atomic.AddInt32(pushed, 1)
				return "", nil
			}
			return "", nil
		},
	}
}

func TestRunPull_Push(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempDir, "repo", ".git"), 0o755)

	var pushed int32
	state := newTestState(pushRunner(false, "2", &pushed))
	state.Config.Pull.Directories = []string{tempDir}
	var buf bytes.Buffer
	state.Stdout = &buf

	if err := RunPull(context.Background(), state, true); err != nil {
		t.Fatalf("RunPull push: %v", err)
	}
	if atomic.LoadInt32(&pushed) != 1 {
		t.Errorf("expected git push once for a clean repo that is ahead, got %d", pushed)
	}
	if !strings.Contains(buf.String(), "pushed 2 commit(s)") {
		t.Errorf("expected push confirmation, got %q", buf.String())
	}
}

func TestRunPull_PushSkipsDirty(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempDir, "repo", ".git"), 0o755)

	var pushed int32
	state := newTestState(pushRunner(true, "2", &pushed))
	state.Config.Pull.Directories = []string{tempDir}
	var buf bytes.Buffer
	state.Stdout = &buf

	if err := RunPull(context.Background(), state, true); err != nil {
		t.Fatalf("RunPull push: %v", err)
	}
	if atomic.LoadInt32(&pushed) != 0 {
		t.Errorf("expected no push for a dirty repo, got %d", pushed)
	}
	if !strings.Contains(buf.String(), "2 unpushed") {
		t.Errorf("expected an unpushed notice for the dirty repo, got %q", buf.String())
	}
}

func TestRunPull_PushFailure(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempDir, "repo", ".git"), 0o755)

	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name != "git" {
				return "", nil
			}
			switch args[0] {
			case "branch":
				return "main\n", nil
			case "rev-list":
				if len(args) > 2 && args[2] == "@{u}..HEAD" {
					return "1\n", nil
				}
				return "0\n", nil
			case "push":
				return "", errors.New("remote rejected")
			}
			return "", nil
		},
	}
	state := newTestState(runner)
	state.Config.Pull.Directories = []string{tempDir}
	var buf bytes.Buffer
	state.Stdout = &buf

	// A failed push must surface as a run failure, not be swallowed after a good pull.
	if err := RunPull(context.Background(), state, true); err == nil {
		t.Fatal("expected push failure to make RunPull return an error")
	}
	if !strings.Contains(buf.String(), "✗ Push failed:") {
		t.Errorf("expected push failure message, got %q", buf.String())
	}
}

// A canceled context (e.g. the per-repo timeout firing on a wedged remote) must be
// reported as a real failure, not silently downgraded to "no upstream".
func TestPullRepo_TimeoutNotNoUpstream(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	_ = os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the deadline has effectively already passed

	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name == "git" {
				switch args[0] {
				case "branch":
					return "main\n", nil
				case "rev-list":
					return "", errors.New("context canceled")
				}
			}
			return "", nil
		},
	}
	state := newTestState(runner)

	res := pullRepo(ctx, state, repoDir, false)
	if res.NoUpstream {
		t.Error("expected a canceled context to be a failure, not 'no upstream'")
	}
	if res.Err == nil {
		t.Error("expected an error to be recorded for the timed-out repo")
	}
}

func TestRunPull_Scenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("directory not readable/not exist", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		// Use a non-existent path
		state.Config.Pull.Directories = []string{"/nonexistent/dir/path"}
		var buf bytes.Buffer
		state.Stdout = &buf
		err := RunPull(ctx, state, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !strings.Contains(buf.String(), "No git repositories found") {
			t.Errorf("Expected 'No git repositories found', got %q", buf.String())
		}
	})

	t.Run("git branch show-current fails", func(t *testing.T) {
		tempDir := t.TempDir()
		repoDir := filepath.Join(tempDir, "repo")
		_ = os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755)

		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "", errors.New("git branch failed")
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}
		var buf bytes.Buffer
		state.Stdout = &buf

		err := RunPull(ctx, state, false)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to pull 1 repositories") {
			t.Errorf("Expected pull failure error, got %v", err)
		}
	})

	t.Run("git branch empty falls back to rev-parse", func(t *testing.T) {
		tempDir := t.TempDir()
		repoDir := filepath.Join(tempDir, "repo")
		_ = os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755)

		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "\n", nil
					}
					if args[0] == "rev-parse" && args[1] == "--short" {
						return "abc1234\n", nil
					}
					if args[0] == "status" {
						return "", nil
					}
					if args[0] == "fetch" {
						return "", nil
					}
					if args[0] == "rev-list" {
						return "3\n", nil
					}
					if args[0] == "pull" {
						return "pulled", nil
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}
		var buf bytes.Buffer
		state.Stdout = &buf

		err := RunPull(ctx, state, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "abc1234") {
			t.Errorf("Expected branch/rev-parse output to contain 'abc1234', got %q", output)
		}
		if !strings.Contains(output, "pulled 3 commit(s)") {
			t.Errorf("Expected commits count in output, got %q", output)
		}
	})

	t.Run("git branch and rev-parse both fail/empty", func(t *testing.T) {
		tempDir := t.TempDir()
		repoDir := filepath.Join(tempDir, "repo")
		_ = os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755)

		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "", nil
					}
					if args[0] == "rev-parse" && args[1] == "--short" {
						return "", errors.New("rev-parse failed")
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}
		var buf bytes.Buffer
		state.Stdout = &buf

		err := RunPull(ctx, state, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "unknown") {
			t.Errorf("Expected fallback to contain 'unknown', got %q", output)
		}
	})

	t.Run("repo dirty and pull fails", func(t *testing.T) {
		tempDir := t.TempDir()
		repoDir := filepath.Join(tempDir, "repo")
		_ = os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755)

		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "main", nil
					}
					if args[0] == "status" {
						return "M file.go\n", nil
					}
					if args[0] == "pull" {
						return "", errors.New("conflict")
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}
		var buf bytes.Buffer
		state.Stdout = &buf

		err := RunPull(ctx, state, false)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		output := buf.String()
		if !strings.Contains(output, "dirty") {
			t.Errorf("Expected 'dirty' in output, got %q", output)
		}
		if !strings.Contains(output, "✗ Pull failed:") {
			t.Errorf("Expected pull failure message in output, got %q", output)
		}
	})

	t.Run("no upstream branch is skipped not failed", func(t *testing.T) {
		tempDir := t.TempDir()
		repoDir := filepath.Join(tempDir, "repo")
		_ = os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755)

		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "feature\n", nil
					}
					// A branch without an upstream: rev-list against @{u} fails.
					if args[0] == "rev-list" && len(args) > 2 && args[2] == "HEAD..@{u}" {
						return "", errors.New("no upstream configured for branch")
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}
		var buf bytes.Buffer
		state.Stdout = &buf

		err := RunPull(ctx, state, false)
		if err != nil {
			t.Fatalf("Expected no error for a no-upstream repo, got %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "no upstream") {
			t.Errorf("Expected 'no upstream' in output, got %q", output)
		}
		if strings.Contains(output, "Pull failed") {
			t.Errorf("Did not expect 'Pull failed' for a no-upstream repo, got %q", output)
		}
	})
}
