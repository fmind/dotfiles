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
	"time"
)

const (
	releaseTestCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	releaseTestParent = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestReleaseCommandContract(t *testing.T) {
	cmd := NewReleaseCmd(newTestState(&FakeRunner{}))
	if !slices.Contains(cmd.Aliases, "r") || cmd.Usage != "Prepare, tag, and push a release commit to trigger CD publication" {
		t.Fatalf("unexpected release command: %+v", cmd)
	}
	if len(cmd.Commands) != 0 {
		t.Fatalf("unexpected subcommands: %+v", cmd.Commands)
	}
}

func TestRunReleasePreparesPushesAndTags(t *testing.T) {
	t.Chdir(releaseFixture(t, "1.1.1"))
	runner := newPrepareRunner(t)
	state := newTestState(runner.fake)
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunRelease(context.Background(), state, true); err != nil {
		t.Fatal(err)
	}
	if !runner.committed || !runner.pushed || !runner.tagPushed {
		t.Fatalf("release was not prepared, pushed, and tagged: %+v", runner)
	}
	if !strings.Contains(stdout.String(), releaseTestCommit) || !strings.Contains(stdout.String(), "Released and tagged v1.2.0") {
		t.Fatalf("release output missing:\n%s", stdout.String())
	}
}

func TestRunReleaseRequiresExactConfiguredBranch(t *testing.T) {
	for _, test := range []struct {
		name      string
		branch    string
		head      string
		upstream  string
		wantError string
	}{
		{name: "wrong branch", branch: "feature", head: releaseTestParent, upstream: releaseTestParent, wantError: "requires branch \"main\""},
		{name: "diverged", branch: "main", head: releaseTestCommit, upstream: releaseTestParent, wantError: "release branch diverged"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(releaseFixture(t, "1.1.1"))
			runner := newPrepareRunner(t)
			runner.branch, runner.head, runner.upstream = test.branch, test.head, test.upstream
			err := RunRelease(context.Background(), newTestState(runner.fake), true)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected %q, got %v", test.wantError, err)
			}
			if runner.committed || runner.pushed || runner.tagPushed {
				t.Fatal("failed precondition allowed release mutation")
			}
		})
	}
}

func TestRunReleaseResumesLocalPreparedCommitAfterPushFailure(t *testing.T) {
	t.Chdir(releaseFixture(t, "1.2.0"))
	runner := newPrepareRunner(t)
	runner.head = releaseTestCommit
	runner.upstream = releaseTestParent
	runner.prepared = true
	runner.pushReachedRemote = false
	state := newTestState(runner.fake)

	if err := RunRelease(context.Background(), state, true); err != nil {
		t.Fatal(err)
	}
	if runner.commits != 0 || !runner.pushed || !runner.tagPushed {
		t.Fatalf("prepared commit was not resumed safely: %+v", runner)
	}
}

func TestPushPreparedCommitRecoveryIsBoundedAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	runner := &FakeRunner{
		RunFunc: func(runCtx context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git fetch origin main":
				assertReleaseRecoveryContext(t, runCtx)
				return "", nil
			case "git rev-parse origin/main":
				assertReleaseRecoveryContext(t, runCtx)
				return releaseTestCommit, nil
			default:
				return "", errors.New("unexpected command: " + command)
			}
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			if name+" "+strings.Join(args, " ") == "git push origin HEAD:refs/heads/main" {
				cancel()
				return context.Canceled
			}
			return errors.New("unexpected interactive command")
		},
	}

	if err := pushPreparedCommit(ctx, newTestState(runner), defaultReleaseConfig(), releaseTestCommit); err != nil {
		t.Fatalf("remote commit at the release commit should recover a canceled push response: %v", err)
	}
}

func TestRollbackStagedReleaseIsBoundedAfterCancellation(t *testing.T) {
	root := releaseFixture(t, "1.1.1")
	snapshot, err := snapshotReleaseFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner := &FakeRunner{RunFunc: func(runCtx context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
		if name+" "+strings.Join(args, " ") != "git reset --mixed HEAD" {
			return "", errors.New("unexpected command")
		}
		assertReleaseRecoveryContext(t, runCtx)
		return "", nil
	}}

	cause := errors.New("release preparation failed")
	if got := rollbackStagedRelease(ctx, newTestState(runner), snapshot, cause); !errors.Is(got, cause) {
		t.Fatalf("rollback lost its original cause: %v", got)
	}
}

func TestPushReleaseTagRecoversAnnotatedTagPushResult(t *testing.T) {
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse refs/tags/v1.2.0^{}":
				return "", errors.New("not found")
			case "git tag -a v1.2.0 -m v1.2.0 " + releaseTestCommit:
				return "", nil
			case "git ls-remote --tags origin refs/tags/v1.2.0 refs/tags/v1.2.0^{}":
				return "cccccccccccccccccccccccccccccccccccccccc\trefs/tags/v1.2.0\n" +
					releaseTestCommit + "\trefs/tags/v1.2.0^{}\n", nil
			default:
				return "", errors.New("unexpected command: " + command)
			}
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			if name+" "+strings.Join(args, " ") == "git push origin refs/tags/v1.2.0" {
				return errors.New("connection lost after remote accepted tag")
			}
			return errors.New("unexpected interactive command")
		},
	}

	err := pushReleaseTag(context.Background(), newTestState(runner), defaultReleaseConfig(), "v1.2.0", releaseTestCommit)
	if err != nil {
		t.Fatalf("remote annotated tag at the release commit should recover a lost push response: %v", err)
	}
}

func TestPushReleaseTagRecoversLightweightTagAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	runner := &FakeRunner{
		RunFunc: func(runCtx context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse refs/tags/v1.2.0^{}":
				return releaseTestCommit, nil
			case "git ls-remote --tags origin refs/tags/v1.2.0 refs/tags/v1.2.0^{}":
				assertReleaseRecoveryContext(t, runCtx)
				return releaseTestCommit + "\trefs/tags/v1.2.0\n", nil
			default:
				return "", errors.New("unexpected command: " + command)
			}
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			if name+" "+strings.Join(args, " ") == "git push origin refs/tags/v1.2.0" {
				cancel()
				return context.Canceled
			}
			return errors.New("unexpected interactive command")
		},
	}

	if err := pushReleaseTag(ctx, newTestState(runner), defaultReleaseConfig(), "v1.2.0", releaseTestCommit); err != nil {
		t.Fatalf("remote lightweight tag at the release commit should recover a canceled push response: %v", err)
	}
}

func assertReleaseRecoveryContext(t *testing.T, ctx context.Context) {
	t.Helper()
	if ctx.Err() != nil {
		t.Fatalf("remote recovery must outlive the canceled push context: %v", ctx.Err())
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("remote recovery must have a deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > 30*time.Second {
		t.Fatalf("remote recovery deadline is outside its bound: %s", remaining)
	}
}

func TestPushReleaseTagDoesNotMistakeLocalTagForRemoteAcceptance(t *testing.T) {
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse refs/tags/v1.2.0^{}":
				return releaseTestCommit, nil
			case "git fetch --tags origin":
				return "", nil
			case "git ls-remote --tags origin refs/tags/v1.2.0 refs/tags/v1.2.0^{}":
				return "", nil
			default:
				return "", errors.New("unexpected command: " + command)
			}
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			if name+" "+strings.Join(args, " ") == "git push origin refs/tags/v1.2.0" {
				return errors.New("remote rejected tag")
			}
			return errors.New("unexpected interactive command")
		},
	}

	err := pushReleaseTag(context.Background(), newTestState(runner), defaultReleaseConfig(), "v1.2.0", releaseTestCommit)
	if err == nil || !strings.Contains(err.Error(), "failed to push tag") {
		t.Fatalf("a matching local tag must not prove remote acceptance, got %v", err)
	}
}

func TestPushReleaseTagRejectsConflictingLocalTag(t *testing.T) {
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			if command == "git rev-parse refs/tags/v1.2.0^{}" {
				return "dddddddddddddddddddddddddddddddddddddddd", nil
			}
			return "", errors.New("unexpected command: " + command)
		},
		RunInteractiveFunc: func(context.Context, string, string, ...string) error {
			t.Fatal("a conflicting local tag must fail before push")
			return nil
		},
	}

	err := pushReleaseTag(context.Background(), newTestState(runner), defaultReleaseConfig(), "v1.2.0", releaseTestCommit)
	if err == nil || !strings.Contains(err.Error(), "local tag v1.2.0 resolves to") {
		t.Fatalf("expected a conflicting local tag error, got %v", err)
	}
}

func TestRunReleaseValidationAndCancellation(t *testing.T) {
	t.Run("dirty tree", func(t *testing.T) {
		runner := newPrepareRunner(t)
		runner.dirty = true
		err := RunRelease(context.Background(), newTestState(runner.fake), true)
		if err == nil || !strings.Contains(err.Error(), "uncommitted or staged changes") {
			t.Fatalf("expected dirty-tree rejection, got %v", err)
		}
	})
	t.Run("validation failure", func(t *testing.T) {
		root := releaseFixture(t, "1.1.1")
		t.Chdir(root)
		runner := newPrepareRunner(t)
		runner.testFailure = true
		err := RunRelease(context.Background(), newTestState(runner.fake), true)
		if err == nil || !strings.Contains(err.Error(), "project test failed") || runner.committed {
			t.Fatalf("validation did not stop the commit: %v", err)
		}
		content, readErr := os.ReadFile(filepath.Join(root, "dot", "version.go"))
		if readErr != nil || !strings.Contains(string(content), `Version = "1.1.1"`) {
			t.Fatalf("failed preparation did not restore version.go: %v %s", readErr, content)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		t.Chdir(releaseFixture(t, "1.1.1"))
		runner := newPrepareRunner(t)
		state := newTestState(runner.fake)
		state.Stdin = strings.NewReader("n\n")
		if err := RunRelease(context.Background(), state, false); err != nil {
			t.Fatal(err)
		}
		if runner.committed || runner.pushed || runner.tagPushed {
			t.Fatal("canceled preparation mutated the release")
		}
	})
	t.Run("commit interruption restores files and index", func(t *testing.T) {
		root := releaseFixture(t, "1.1.1")
		t.Chdir(root)
		runner := newPrepareRunner(t)
		runner.commitFailure = true
		err := RunRelease(context.Background(), newTestState(runner.fake), true)
		if err == nil || !strings.Contains(err.Error(), "git commit failed") || runner.resets != 1 {
			t.Fatalf("commit interruption was not rolled back: %v, resets=%d", err, runner.resets)
		}
		content, readErr := os.ReadFile(filepath.Join(root, "dot", "version.go"))
		if readErr != nil || !strings.Contains(string(content), `Version = "1.1.1"`) {
			t.Fatalf("commit interruption did not restore version.go: %v %s", readErr, content)
		}
	})
}

func TestValidateReleaseStatus(t *testing.T) {
	if err := validateReleaseStatus(" M CHANGELOG.md\x00 M dot/version.go\x00"); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"x\x00", "R  renamed\x00", "?? unrelated\x00"} {
		if err := validateReleaseStatus(status); err == nil {
			t.Fatalf("unsafe status %q was accepted", status)
		}
	}
}

func TestConfirmRelease(t *testing.T) {
	for _, test := range []struct {
		answer string
		want   bool
	}{{"yes\n", true}, {"Y\n", true}, {"no\n", false}, {"", false}} {
		var output strings.Builder
		if got := confirmRelease(strings.NewReader(test.answer), &output, "Proceed? "); got != test.want || output.String() != "Proceed? " {
			t.Fatalf("confirmRelease(%q) = %t, output %q", test.answer, got, output.String())
		}
	}
}

type prepareRunner struct {
	fake              *FakeRunner
	branch            string
	head              string
	upstream          string
	dirty             bool
	prepared          bool
	committed         bool
	pushed            bool
	tagPushed         bool
	pushReturnsError  bool
	pushReachedRemote bool
	testFailure       bool
	commitFailure     bool
	commits           int
	resets            int
}

func newPrepareRunner(t *testing.T) *prepareRunner {
	t.Helper()
	runner := &prepareRunner{branch: "main", head: releaseTestParent, upstream: releaseTestParent}
	runner.fake = &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch {
			case command == "git rev-parse --is-inside-work-tree":
				return "true", nil
			case command == "git rev-parse --show-toplevel":
				root, err := os.Getwd()
				if err != nil {
					return "", err
				}
				return root, nil
			case command == "git status --porcelain":
				if runner.dirty {
					return " M local", nil
				}
				return "", nil
			case command == "gh auth status":
				return "authenticated", nil
			case command == "git branch --show-current":
				return runner.branch, nil
			case command == "git fetch --prune --tags origin", command == "git fetch origin main":
				if runner.pushReachedRemote && runner.committed {
					runner.upstream = releaseTestCommit
				}
				return "", nil
			case command == "git rev-parse HEAD":
				if runner.committed {
					return releaseTestCommit, nil
				}
				return runner.head, nil
			case command == "git rev-parse origin/main":
				return runner.upstream, nil
			case command == "git rev-parse HEAD^":
				return releaseTestParent, nil
			case command == "git log -1 --pretty=%s":
				if runner.prepared || runner.committed {
					return "chore(release): v1.2.0", nil
				}
				return "feat: change", nil
			case command == "git-cliff --config dot_config/git-cliff/cliff.toml --bumped-version":
				return "v1.2.0", nil
			case command == "git describe --tags --abbrev=0":
				return "v1.1.1", nil
			case command == "git-cliff --config dot_config/git-cliff/cliff.toml --bump -o CHANGELOG.md":
				return "", nil
			case command == "git status --porcelain=v1 -z --untracked-files=all":
				return " M CHANGELOG.md\x00 M dot/version.go\x00", nil
			case command == "git add CHANGELOG.md dot/version.go":
				return "", nil
			case command == "git commit -m chore(release): v1.2.0":
				if runner.commitFailure {
					return "", errors.New("interrupted")
				}
				runner.committed = true
				runner.commits++
				return "", nil
			case command == "git reset --mixed HEAD":
				runner.resets++
				return "", nil
			case strings.HasPrefix(command, "git rev-parse refs/tags/"):
				if runner.tagPushed {
					return releaseTestCommit, nil
				}
				return "", errors.New("not found")
			case strings.HasPrefix(command, "git tag -a"):
				return "", nil
			case strings.HasPrefix(command, "mise run --force build"):
				return "", nil
			case strings.HasPrefix(command, "chezmoi apply"):
				return "", nil
			}
			return "", nil
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			command := name + " " + strings.Join(args, " ")
			if command == "mise run test" && runner.testFailure {
				return errors.New("test failed")
			}
			if command == "git push origin HEAD:refs/heads/main" {
				runner.pushed = true
				if runner.pushReachedRemote {
					runner.upstream = releaseTestCommit
				}
				if runner.pushReturnsError {
					return errors.New("connection lost")
				}
				runner.upstream = releaseTestCommit
				return nil
			}
			if strings.HasPrefix(command, "git push origin refs/tags/") {
				runner.tagPushed = true
				return nil
			}
			return nil
		},
	}
	return runner
}

func releaseFixture(t *testing.T, version string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "dot"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "package dot\n\nvar Version = \"" + version + "\"\n"
	if err := os.WriteFile(filepath.Join(root, "dot", "version.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte("# Changelog\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
