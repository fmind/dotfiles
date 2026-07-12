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

func TestReleaseCommandAlias(t *testing.T) {
	state := newTestState(&FakeRunner{})
	cmd := NewReleaseCmd(state)

	hasAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "r" {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		t.Errorf("expected release command to have 'rl' alias, got: %v", cmd.Aliases)
	}
}

func TestPublishGitHubReleaseReportsCleanupFailure(t *testing.T) {
	var tempPath string
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name == "git-cliff" && strings.Contains(strings.Join(args, " "), "--latest") {
				return "release notes", nil
			}
			return "", errors.New("unexpected command")
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			if name != "gh" {
				return errors.New("unexpected interactive command")
			}
			for i, arg := range args {
				if arg == "--notes-file" && i+1 < len(args) {
					tempPath = args[i+1]
				}
			}
			if tempPath == "" {
				return errors.New("missing notes file")
			}
			if err := os.Remove(tempPath); err != nil {
				return err
			}
			if err := os.Mkdir(tempPath, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(tempPath, "leftover"), []byte("data"), 0o600)
		},
	}

	err := publishGitHubRelease(context.Background(), newTestState(runner), "v1.2.0")
	if tempPath != "" {
		t.Cleanup(func() { _ = os.RemoveAll(tempPath) })
	}
	if err == nil || !strings.Contains(err.Error(), "failed to remove temporary release notes") {
		t.Fatalf("expected release-notes cleanup failure, got %v", err)
	}
}

func TestRunReleaseNoBump(t *testing.T) {
	fake := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			cmdStr := name + " " + strings.Join(args, " ")
			switch {
			case strings.Contains(cmdStr, "git rev-parse --is-inside-work-tree"):
				return "true", nil
			case strings.Contains(cmdStr, "git status --porcelain"):
				return "", nil // clean status
			case strings.Contains(cmdStr, "gh auth status"):
				return "Logged in to github.com", nil
			case strings.Contains(cmdStr, "git-cliff --config dot_config/git-cliff/cliff.toml --bumped-version"):
				return "v1.0.0", nil
			case strings.Contains(cmdStr, "git describe --tags --abbrev=0"):
				return "v1.0.0", nil
			default:
				return "", nil
			}
		},
	}
	state := newTestState(fake)
	err := RunRelease(context.Background(), state, true)
	if err != nil {
		t.Fatalf("RunRelease failed: %v", err)
	}
}

func releaseTestRunner(bumped string) *FakeRunner {
	return &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch {
			case command == "git rev-parse --is-inside-work-tree":
				return "true", nil
			case command == "git status --porcelain":
				return "", nil
			case command == "gh auth status":
				return "authenticated", nil
			case strings.Contains(command, "git-cliff --config dot_config/git-cliff/cliff.toml --bumped-version"):
				return bumped, nil
			case command == "git describe --tags --abbrev=0":
				return "v1.1.1", nil
			case command == "git branch --show-current":
				return "main", nil
			case command == "git config --get branch.main.remote":
				return "origin", nil
			case command == "git config --get branch.main.merge":
				return "refs/heads/main", nil
			case strings.Contains(command, "git-cliff --config dot_config/git-cliff/cliff.toml --latest"):
				return "release notes", nil
			default:
				return "", nil
			}
		},
	}
}

func writeReleaseVersionFile(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll("dot", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("dot", "version.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunReleaseSafetyChecks(t *testing.T) {
	t.Run("rejects invalid bumped tag", func(t *testing.T) {
		err := RunRelease(context.Background(), newTestState(releaseTestRunner("release-1.2.0")), true)
		if err == nil || !strings.Contains(err.Error(), "invalid semantic version tag") {
			t.Fatalf("expected invalid tag error, got %v", err)
		}
	})

	t.Run("rejects detached head", func(t *testing.T) {
		runner := releaseTestRunner("v1.2.0")
		baseRun := runner.RunFunc
		runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "git" && strings.Join(args, " ") == "branch --show-current" {
				return "", nil
			}
			return baseRun(ctx, dir, stdin, name, args...)
		}
		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "detached HEAD") {
			t.Fatalf("expected detached HEAD error, got %v", err)
		}
	})

	t.Run("requires upstream", func(t *testing.T) {
		runner := releaseTestRunner("v1.2.0")
		baseRun := runner.RunFunc
		runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "git" && strings.Join(args, " ") == "config --get branch.main.remote" {
				return "", errors.New("no upstream")
			}
			return baseRun(ctx, dir, stdin, name, args...)
		}
		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "upstream remote") {
			t.Fatalf("expected upstream error, got %v", err)
		}
	})

	t.Run("rejects local-dot upstream", func(t *testing.T) {
		runner := releaseTestRunner("v1.2.0")
		baseRun := runner.RunFunc
		runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "git" && strings.Join(args, " ") == "config --get branch.main.remote" {
				return ".", nil
			}
			return baseRun(ctx, dir, stdin, name, args...)
		}
		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "has no upstream") {
			t.Fatalf("expected local-dot upstream rejection, got %v", err)
		}
	})

	t.Run("requires mise", func(t *testing.T) {
		runner := releaseTestRunner("v1.2.0")
		runner.LookPathFunc = func(name string) (string, error) {
			if name == "mise" {
				return "", errors.New("not found")
			}
			return "/bin/" + name, nil
		}
		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "release validation cannot run") {
			t.Fatalf("expected missing mise error, got %v", err)
		}
	})

	t.Run("requires version assignment", func(t *testing.T) {
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nconst Other = \"1.1.1\"\n")
		err := RunRelease(context.Background(), newTestState(releaseTestRunner("v1.2.0")), true)
		if err == nil || !strings.Contains(err.Error(), "expected version assignment") {
			t.Fatalf("expected version assignment error, got %v", err)
		}
	})

	t.Run("rejects multiple version assignments", func(t *testing.T) {
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\nvar Version = \"duplicate\"\n")
		err := RunRelease(context.Background(), newTestState(releaseTestRunner("v1.2.0")), true)
		if err == nil || !strings.Contains(err.Error(), "found 2") {
			t.Fatalf("expected duplicate version assignment error, got %v", err)
		}
	})

	t.Run("formatter failure stops release", func(t *testing.T) {
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
		runner := releaseTestRunner("v1.2.0")
		runner.RunInteractiveFunc = func(_ context.Context, _, name string, args ...string) error {
			if name == "mise" && strings.Join(args, " ") == "run format" {
				return errors.New("format failed")
			}
			return nil
		}
		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "project formatting failed") {
			t.Fatalf("expected formatter error, got %v", err)
		}
	})

	t.Run("test failure stops release before commit", func(t *testing.T) {
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
		runner := releaseTestRunner("v1.2.0")
		baseRun := runner.RunFunc
		committed := false
		runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "git" && len(args) > 0 && args[0] == "commit" {
				committed = true
			}
			return baseRun(ctx, dir, stdin, name, args...)
		}
		runner.RunInteractiveFunc = func(_ context.Context, _, name string, args ...string) error {
			if name == "mise" && strings.Join(args, " ") == "run test" {
				return errors.New("tests failed")
			}
			return nil
		}
		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "project tests failed") {
			t.Fatalf("expected test failure, got %v", err)
		}
		if committed {
			t.Fatal("release commit ran after test failure")
		}
	})

	t.Run("unrelated validation changes stop release before commit", func(t *testing.T) {
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
		runner := releaseTestRunner("v1.2.0")
		baseRun := runner.RunFunc
		committed := false
		runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			if command == "git status --porcelain=v1 -z --untracked-files=all" {
				return " M CHANGELOG.md\x00 M dot/version.go\x00?? unrelated.md\x00", nil
			}
			if name == "git" && len(args) > 0 && args[0] == "commit" {
				committed = true
			}
			return baseRun(ctx, dir, stdin, name, args...)
		}

		err := RunRelease(context.Background(), newTestState(runner), true)
		if err == nil || !strings.Contains(err.Error(), "release validation changed unrelated paths: unrelated.md") {
			t.Fatalf("expected unrelated release change error, got %v", err)
		}
		if committed {
			t.Fatal("release commit ran with unrelated validation changes")
		}
	})

	t.Run("pushes to configured remote and upstream ref atomically", func(t *testing.T) {
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
		runner := releaseTestRunner("v1.2.0")
		baseRun := runner.RunFunc
		runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "git" && strings.Join(args, " ") == "config --get branch.main.remote" {
				return "work", nil
			}
			if name == "git" && strings.Join(args, " ") == "config --get branch.main.merge" {
				return "refs/heads/release", nil
			}
			return baseRun(ctx, dir, stdin, name, args...)
		}
		pushes := make([]string, 0, 1)
		runner.RunInteractiveFunc = func(_ context.Context, _, name string, args ...string) error {
			if name == "git" && len(args) > 0 && args[0] == "push" {
				pushes = append(pushes, strings.Join(args, " "))
			}
			return nil
		}
		if err := RunRelease(context.Background(), newTestState(runner), true); err != nil {
			t.Fatalf("RunRelease failed: %v", err)
		}
		if len(pushes) != 1 || pushes[0] != "push --atomic work HEAD:refs/heads/release v1.2.0" {
			t.Fatalf("expected one atomic push, got %v", pushes)
		}
	})
}
