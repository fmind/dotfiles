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

type commitIntegrationRunner struct {
	standard    *StandardRunner
	scanErr     error
	aiErr       error
	commitErr   error
	rollbackErr error
	scanned     string
}

func (r *commitIntegrationRunner) LookPath(name string) (string, error) {
	switch name {
	case "agy", "gitleaks":
		return "/test/" + name, nil
	default:
		return r.standard.LookPath(name)
	}
}

func (r *commitIntegrationRunner) Run(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
	switch name {
	case "/test/gitleaks":
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", err
		}
		r.scanned = string(data)
		return "", r.scanErr
	case "/test/agy":
		if r.aiErr != nil {
			return "", r.aiErr
		}
		return "feat: add untracked file", nil
	case "git":
		if len(args) == 2 && args[0] == "reset" && args[1] == "--mixed" && r.rollbackErr != nil {
			return "", r.rollbackErr
		}
	}
	return r.standard.Run(ctx, dir, stdin, name, args...)
}

func (r *commitIntegrationRunner) RunInteractive(ctx context.Context, dir, name string, args ...string) error {
	if name == "git" && len(args) > 0 && args[0] == "commit" && r.commitErr != nil {
		return r.commitErr
	}
	return r.standard.RunInteractive(ctx, dir, name, args...)
}

func setupCommitIntegrationRepo(t *testing.T) (*commitIntegrationRunner, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for integration test")
	}

	repo := t.TempDir()
	standard := NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard)
	runner := &commitIntegrationRunner{standard: standard}
	git := func(args ...string) string {
		t.Helper()
		out, err := standard.Run(context.Background(), repo, nil, "git", args...)
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return out
	}

	git("init", "-b", "main")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	git("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "base.txt")
	git("commit", "-m", "base")
	return runner, repo
}

func TestRunCommitStagesUntrackedFiles(t *testing.T) {
	runner, repo := setupCommitIntegrationRepo(t)
	t.Chdir(repo)
	t.Setenv("GIT_EDITOR", "true")
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("new content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RunCommit(context.Background(), newTestState(runner), "", ""); err != nil {
		t.Fatalf("RunCommit failed: %v", err)
	}
	changed, err := runner.standard.Run(context.Background(), repo, nil, "git", "show", "--pretty=format:", "--name-only", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(changed, "untracked.txt") {
		t.Fatalf("untracked file was not committed: %q", changed)
	}
	if !strings.Contains(runner.scanned, "untracked.txt") || !strings.Contains(runner.scanned, "new content") {
		t.Fatalf("gitleaks did not receive the staged untracked-file diff: %q", runner.scanned)
	}
}

func TestRunCommitAutoStageRollback(t *testing.T) {
	tests := []struct {
		scanErr   error
		aiErr     error
		commitErr error
		name      string
	}{
		{name: "scan failure", scanErr: errors.New("secret detected")},
		{name: "AI failure", aiErr: errors.New("provider unavailable")},
		{name: "commit failure", commitErr: errors.New("commit hook rejected changes")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, repo := setupCommitIntegrationRepo(t)
			t.Chdir(repo)
			if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("new content\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			runner.scanErr = tt.scanErr
			runner.aiErr = tt.aiErr
			runner.commitErr = tt.commitErr

			if err := RunCommit(context.Background(), newTestState(runner), "", ""); err == nil {
				t.Fatal("expected pre-commit failure")
			}
			cached, err := runner.standard.Run(context.Background(), repo, nil, "git", "diff", "--cached", "--name-only")
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(cached) != "" {
				t.Fatalf("initially clean index was not restored: %q", cached)
			}
			status, err := runner.standard.Run(context.Background(), repo, nil, "git", "status", "--porcelain")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(status, "?? untracked.txt") {
				t.Fatalf("rollback lost the untracked worktree file: %q", status)
			}
		})
	}
}

func TestRunCommitSurfacesRollbackFailure(t *testing.T) {
	runner, repo := setupCommitIntegrationRepo(t)
	t.Chdir(repo)
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("new content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runner.scanErr = errors.New("secret detected")
	runner.rollbackErr = errors.New("index locked")

	err := RunCommit(context.Background(), newTestState(runner), "", "")
	if err == nil || !strings.Contains(err.Error(), "failed to restore initially clean index") || !strings.Contains(err.Error(), "secret scan failed") {
		t.Fatalf("expected scan and rollback errors, got %v", err)
	}
}
