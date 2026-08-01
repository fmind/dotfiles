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

func isolateGitEnvironment(t *testing.T) {
	t.Helper()
	command := exec.Command("git", "rev-parse", "--local-env-vars")
	output, err := command.Output()
	if err != nil {
		t.Fatalf("list repository-local Git environment: %v", err)
	}
	for _, name := range strings.Fields(string(output)) {
		value, existed := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("unset %s: %v", name, err)
		}
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(name, value)
			} else {
				_ = os.Unsetenv(name)
			}
		})
	}
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
	// Git exports GIT_DIR and related variables to hooks. A fixture that initializes
	// a foreign repository must not let those variables redirect mutations home.
	isolateGitEnvironment(t)

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

// TestRunCommitPreconditionFailures drives the pre-AI stages with a pure fake so
// each git failure reports its own context instead of a bare exec error.
func TestRunCommitPreconditionFailures(t *testing.T) {
	tests := []struct {
		name    string
		run     func(args []string) (string, error, bool)
		wantMsg string
	}{
		{
			name: "unfiltered diff failure",
			run: func(args []string) (string, error, bool) {
				if args[0] == "diff" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to get unfiltered git diff",
		},
		{
			name: "status failure",
			run: func(args []string) (string, error, bool) {
				if args[0] == "status" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to inspect working tree",
		},
		{
			name: "staging failure",
			run: func(args []string) (string, error, bool) {
				switch args[0] {
				case "status":
					return "?? new.txt\n", nil, true
				case "add":
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to stage working tree changes",
		},
		{
			name: "staging produces no diff",
			run: func(args []string) (string, error, bool) {
				if args[0] == "status" {
					return "?? new.txt\n", nil, true
				}
				return "", nil, false
			},
			wantMsg: "git add -A completed without producing a staged diff",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := newTestState(&FakeRunner{
				RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
					if name != "git" {
						return "", nil
					}
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if out, err, handled := tc.run(args); handled {
						return out, err
					}
					return "", nil
				},
			})

			err := RunCommit(context.Background(), state, "", "")
			if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("expected an error containing %q, got %v", tc.wantMsg, err)
			}
		})
	}
}

func TestRunCommitReportsNothingToCommit(t *testing.T) {
	state := newTestState(&FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name == "git" && args[0] == "rev-parse" {
				return "true", nil
			}
			return "", nil
		},
	})
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunCommit(context.Background(), state, "", ""); err != nil {
		t.Fatalf("RunCommit: %v", err)
	}
	if !strings.Contains(stdout.String(), "No changes to commit.") {
		t.Errorf("expected a no-changes message, got %q", stdout.String())
	}
}

// The commit type and scope hints are appended to the AI prompt; a dropped hint
// silently produces a message that ignores what the user asked for.
func TestRunCommitPromptHints(t *testing.T) {
	tests := []struct{ name, commitType, commitScope, want string }{
		{name: "type only", commitType: "fix", want: "Suggest a scope and use 'fix' as the type."},
		{name: "scope only", commitScope: "dot", want: "Use scope 'dot' and suggest an appropriate type."},
		{name: "both", commitType: "feat", commitScope: "dot", want: "Use type 'feat' and scope 'dot'."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var prompt string
			state := newTestState(&FakeRunner{
				RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
					switch {
					case name == "git" && args[0] == "rev-parse":
						return "true", nil
					case name == "git" && args[0] == "diff":
						return "some diff", nil
					case strings.HasSuffix(name, "agy"):
						for i, arg := range args {
							if arg == "--prompt" && i+1 < len(args) {
								prompt = args[i+1]
							}
						}
						return "fix(dot): message", nil
					}
					return "", nil
				},
			})

			if err := RunCommit(context.Background(), state, tc.commitType, tc.commitScope); err != nil {
				t.Fatalf("RunCommit: %v", err)
			}
			if !strings.Contains(prompt, tc.want) {
				t.Errorf("expected the prompt to contain %q, got %q", tc.want, prompt)
			}
		})
	}
}
