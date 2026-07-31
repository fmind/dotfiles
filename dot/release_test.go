package dot

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestReleaseCommandAlias(t *testing.T) {
	state := newTestState(&FakeRunner{})
	cmd := NewReleaseCmd(state)

	hasAlias := slices.Contains(cmd.Aliases, "r")
	if !hasAlias {
		t.Errorf("expected release command to have 'r' alias, got: %v", cmd.Aliases)
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

func TestConfirmRelease(t *testing.T) {
	for _, tt := range []struct {
		answer string
		want   bool
	}{
		{answer: "yes\n", want: true},
		{answer: "Y\n", want: true},
		{answer: "no\n", want: false},
		{answer: "", want: false},
	} {
		var output strings.Builder
		if got := confirmRelease(strings.NewReader(tt.answer), &output, "Proceed? "); got != tt.want {
			t.Fatalf("confirmRelease(%q) = %t, want %t", tt.answer, got, tt.want)
		}
		if output.String() != "Proceed? " {
			t.Fatalf("prompt = %q", output.String())
		}
	}
}

func TestReleaseCommandDispatches(t *testing.T) {
	state := newTestState(&FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, _ string, _ ...string) (string, error) {
			return "", errors.New("not a repository")
		},
	})
	cmd := NewReleaseCmd(state)
	if err := cmd.Action(context.Background(), cmd); err == nil {
		t.Fatal("expected release precondition failure")
	}
}

func TestRunReleaseRejectsDirtyWorktree(t *testing.T) {
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse --is-inside-work-tree":
				return "true", nil
			case "git status --porcelain":
				return " M local-change", nil
			default:
				return "", errors.New("unexpected command")
			}
		},
	}
	err := RunRelease(context.Background(), newTestState(runner), true)
	if err == nil || !strings.Contains(err.Error(), "uncommitted or staged changes") {
		t.Fatalf("expected dirty-worktree rejection, got %v", err)
	}
}

func TestValidateReleaseStatusMalformedRecords(t *testing.T) {
	for _, status := range []string{"x\x00", "M? path\x00", "R  renamed\x00", "C  copied\x00"} {
		if err := validateReleaseStatus(status); err == nil {
			t.Fatalf("expected malformed or unsupported status %q to fail", status)
		}
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

// TestRunReleaseCommandFailures walks the release sequence and fails one command at
// a time, asserting each step reports its own context instead of a bare exec error.
func TestRunReleaseCommandFailures(t *testing.T) {
	failRun := func(match string) func(*FakeRunner) {
		return func(runner *FakeRunner) {
			base := runner.RunFunc
			runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name+" "+strings.Join(args, " ") == match {
					return "", errors.New("boom")
				}
				return base(ctx, dir, stdin, name, args...)
			}
		}
	}
	// Prefix matching, because the release-notes commands carry a generated temp path
	// and extra flags that a test should not have to spell out in full.
	failRunPrefix := func(prefix string) func(*FakeRunner) {
		return func(runner *FakeRunner) {
			base := runner.RunFunc
			runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if strings.HasPrefix(name+" "+strings.Join(args, " "), prefix) {
					return "", errors.New("boom")
				}
				return base(ctx, dir, stdin, name, args...)
			}
		}
	}
	failInteractive := func(prefix string) func(*FakeRunner) {
		return func(runner *FakeRunner) {
			runner.RunInteractiveFunc = func(_ context.Context, _, name string, args ...string) error {
				if strings.HasPrefix(name+" "+strings.Join(args, " "), prefix) {
					return errors.New("boom")
				}
				return nil
			}
		}
	}

	tests := []struct {
		name    string
		setup   func(*FakeRunner)
		wantMsg string
	}{
		{name: "git status", setup: failRun("git status --porcelain"), wantMsg: "failed to check git status"},
		{name: "gh auth", setup: failRun("gh auth status"), wantMsg: "github CLI is not authenticated"},
		{
			name: "git-cliff missing",
			setup: func(runner *FakeRunner) {
				runner.LookPathFunc = func(name string) (string, error) {
					if name == "git-cliff" {
						return "", errors.New("not found")
					}
					return "/bin/" + name, nil
				}
			},
			wantMsg: "git-cliff is not installed",
		},
		{
			name:    "bumped version",
			setup:   failRun("git-cliff --config dot_config/git-cliff/cliff.toml --bumped-version"),
			wantMsg: "failed to calculate next version",
		},
		{name: "current branch", setup: failRun("git branch --show-current"), wantMsg: "failed to get current branch"},
		{
			name:    "upstream ref",
			setup:   failRun("git config --get branch.main.merge"),
			wantMsg: "failed to resolve upstream ref",
		},
		{
			name: "invalid upstream ref",
			setup: func(runner *FakeRunner) {
				base := runner.RunFunc
				runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
					if name+" "+strings.Join(args, " ") == "git config --get branch.main.merge" {
						return "refs/tags/v1.2.0", nil
					}
					return base(ctx, dir, stdin, name, args...)
				}
			},
			wantMsg: "invalid upstream ref",
		},
		{
			name:    "changelog generation",
			setup:   failRun("git-cliff --config dot_config/git-cliff/cliff.toml --bump -o CHANGELOG.md"),
			wantMsg: "failed to generate CHANGELOG.md",
		},
		{name: "checks", setup: failInteractive("mise run check"), wantMsg: "project checks failed"},
		{
			name:    "post-validation status",
			setup:   failRun("git status --porcelain=v1 -z --untracked-files=all"),
			wantMsg: "failed to inspect release changes after validation",
		},
		{
			name:    "stage",
			setup:   failRun("git add CHANGELOG.md dot/version.go"),
			wantMsg: "failed to stage release files",
		},
		{name: "commit", setup: failRun("git commit -m chore(release): v1.2.0"), wantMsg: "git commit failed"},
		{name: "tag", setup: failRun("git tag -a v1.2.0 -m v1.2.0"), wantMsg: "git tag failed"},
		{
			name:    "push",
			setup:   failInteractive("git push --atomic origin HEAD:refs/heads/main v1.2.0"),
			wantMsg: "atomic push of HEAD:refs/heads/main and tag v1.2.0 to origin failed",
		},
		{
			name:    "release notes generation",
			setup:   failRunPrefix("git-cliff --config dot_config/git-cliff/cliff.toml --latest"),
			wantMsg: "failed to generate latest changelog",
		},
		{
			name:    "github release creation",
			setup:   failInteractive("gh release create v1.2.0 --title v1.2.0 --notes-file"),
			wantMsg: "failed to create github release",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
			runner := releaseTestRunner("v1.2.0")
			tc.setup(runner)

			err := RunRelease(context.Background(), newTestState(runner), true)
			if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("expected an error containing %q, got %v", tc.wantMsg, err)
			}
		})
	}
}

func TestRunReleaseWithoutPriorTag(t *testing.T) {
	t.Chdir(t.TempDir())
	writeReleaseVersionFile(t, "package dot\n\nvar Version = \"0.0.0\"\n")
	runner := releaseTestRunner("v0.1.0")
	base := runner.RunFunc
	runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
		// A fresh repository has no tags; the release must fall back to v0.0.0.
		if name+" "+strings.Join(args, " ") == "git describe --tags --abbrev=0" {
			return "", errors.New("no names found")
		}
		return base(ctx, dir, stdin, name, args...)
	}
	state := newTestState(runner)
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunRelease(context.Background(), state, true); err != nil {
		t.Fatalf("RunRelease: %v", err)
	}
	if !strings.Contains(stdout.String(), "v0.0.0") {
		t.Errorf("expected the v0.0.0 fallback to be reported, got:\n%s", stdout.String())
	}
}

func TestRunReleaseCanceledAtConfirmation(t *testing.T) {
	t.Chdir(t.TempDir())
	writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
	runner := releaseTestRunner("v1.2.0")
	base := runner.RunFunc
	runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
		if name == "git" && len(args) > 0 && (args[0] == "commit" || args[0] == "tag" || args[0] == "add") {
			t.Errorf("release must not mutate the repository after cancellation (ran git %s)", args[0])
		}
		return base(ctx, dir, stdin, name, args...)
	}
	state := newTestState(runner)
	state.Stdin = strings.NewReader("n\n")
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunRelease(context.Background(), state, false); err != nil {
		t.Fatalf("RunRelease: %v", err)
	}
	if !strings.Contains(stdout.String(), "Release canceled.") {
		t.Errorf("expected the release to be canceled, got:\n%s", stdout.String())
	}
}

func TestRunReleaseVersionFileFailures(t *testing.T) {
	t.Run("a missing version file stops the release", func(t *testing.T) {
		t.Chdir(t.TempDir())
		err := RunRelease(context.Background(), newTestState(releaseTestRunner("v1.2.0")), true)
		if err == nil || !strings.Contains(err.Error(), "failed to read version.go") {
			t.Fatalf("expected a read error, got %v", err)
		}
	})

	t.Run("an unwritable version file stops the release", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root bypasses file permissions")
		}
		t.Chdir(t.TempDir())
		writeReleaseVersionFile(t, "package dot\n\nvar Version = \"1.1.1\"\n")
		if err := os.Chmod(filepath.Join("dot", "version.go"), 0o400); err != nil {
			t.Fatal(err)
		}

		err := RunRelease(context.Background(), newTestState(releaseTestRunner("v1.2.0")), true)
		if err == nil || !strings.Contains(err.Error(), "failed to write version.go") {
			t.Fatalf("expected a write error, got %v", err)
		}
	})
}
