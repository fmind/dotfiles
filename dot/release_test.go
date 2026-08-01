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
	if !slices.Contains(cmd.Aliases, "r") || cmd.Usage != "Prepare and dispatch an exact-head GitHub release" {
		t.Fatalf("unexpected release command: %+v", cmd)
	}
	if len(cmd.Commands) != 1 || cmd.Commands[0].Name != "publish" || !cmd.Commands[0].Hidden {
		t.Fatalf("workflow-only publish subcommand is unavailable: %+v", cmd.Commands)
	}
}

func TestRunReleasePreparesPushesAndDispatchesWithoutPublishing(t *testing.T) {
	t.Chdir(releaseFixture(t, "1.1.1"))
	runner := newPrepareRunner(t)
	state := newTestState(runner.fake)
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunRelease(context.Background(), state, true); err != nil {
		t.Fatal(err)
	}
	if !runner.committed || !runner.pushed || runner.dispatches != 1 {
		t.Fatalf("release was not prepared exactly once: %+v", runner)
	}
	if runner.published || runner.tagged {
		t.Fatal("local preparation created a tag or GitHub release")
	}
	if !strings.Contains(stdout.String(), releaseTestCommit) || !strings.Contains(stdout.String(), "https://github.com/fmind/dotfiles/actions/workflows/release.yml") {
		t.Fatalf("prepared commit and follow-up URL were not reported:\n%s", stdout.String())
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
			if runner.committed || runner.pushed || runner.dispatches > 0 {
				t.Fatal("failed precondition allowed release mutation")
			}
		})
	}
}

func TestRunReleaseResumesPartialPushAndDispatch(t *testing.T) {
	t.Chdir(releaseFixture(t, "1.1.1"))
	runner := newPrepareRunner(t)
	runner.pushReturnsError = true
	runner.pushReachedRemote = true
	runner.dispatchFailures = 1
	state := newTestState(runner.fake)

	firstErr := RunRelease(context.Background(), state, true)
	if firstErr == nil || !strings.Contains(firstErr.Error(), "workflow dispatch failed") {
		t.Fatalf("expected first dispatch failure, got %v", firstErr)
	}
	if !runner.committed || !runner.pushed || runner.dispatches != 1 {
		t.Fatalf("partial push was not reconciled: %+v", runner)
	}
	if err := RunRelease(context.Background(), state, true); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if runner.commits != 1 || runner.dispatches != 2 || runner.tagged || runner.published {
		t.Fatalf("retry repeated mutation or published locally: %+v", runner)
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
	if runner.commits != 0 || !runner.pushed || runner.dispatches != 1 {
		t.Fatalf("prepared commit was not resumed safely: %+v", runner)
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
		if runner.committed || runner.pushed || runner.dispatches > 0 {
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

func TestRunReleasePublishGatesTagsAndPublishes(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Chdir(releaseFixture(t, "1.2.0"))
	runner := newPublishRunner(t)
	state := newTestState(runner.fake)
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunReleasePublish(context.Background(), state, releaseTestCommit, &fakeReleaseWaiter{}); err != nil {
		t.Fatal(err)
	}
	if runner.tagCreates != 1 || runner.tagPushes != 1 || runner.releaseCreates != 1 {
		t.Fatalf("publication did not create exactly one tag and release: %+v", runner)
	}
	if !strings.Contains(stdout.String(), runner.ciURL) || !strings.Contains(stdout.String(), runner.releaseURL) {
		t.Fatalf("publication evidence missing:\n%s", stdout.String())
	}

	if err := RunReleasePublish(context.Background(), state, releaseTestCommit, &fakeReleaseWaiter{}); err != nil {
		t.Fatalf("idempotent retry failed: %v", err)
	}
	if runner.tagCreates != 1 || runner.tagPushes != 1 || runner.releaseCreates != 1 {
		t.Fatalf("retry recreated immutable state: %+v", runner)
	}
}

func TestRunReleasePublishRejectsFailedAmbiguousAndTimedOutCI(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Chdir(releaseFixture(t, "1.2.0"))
	for _, test := range []struct {
		name      string
		runs      string
		waiter    *fakeReleaseWaiter
		wantError string
	}{
		{name: "failure", runs: ciRuns("failure"), waiter: &fakeReleaseWaiter{}, wantError: "concluded failure"},
		{name: "canceled", runs: ciRuns("cancel" + "led"), waiter: &fakeReleaseWaiter{}, wantError: "concluded cancel" + "led"},
		{name: "ambiguous", runs: duplicateCIRuns(), waiter: &fakeReleaseWaiter{}, wantError: "ambiguous exact-head CI"},
		{name: "timeout", runs: "[]", waiter: &fakeReleaseWaiter{step: releaseGateTimeout}, wantError: "timed out waiting"},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := newPublishRunner(t)
			runner.runs = test.runs
			err := RunReleasePublish(context.Background(), newTestState(runner.fake), releaseTestCommit, test.waiter)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected %q, got %v", test.wantError, err)
			}
			if runner.tagCreates > 0 || runner.releaseCreates > 0 {
				t.Fatal("failed CI allowed publication")
			}
		})
	}
}

func TestRunReleasePublishRejectsMovedOrLightweightTag(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Chdir(releaseFixture(t, "1.2.0"))
	for _, test := range []struct {
		name      string
		remoteTag string
		wantError string
	}{
		{name: "moved", remoteTag: annotatedTag("v1.2.0", releaseTestParent), wantError: "already targets"},
		{name: "lightweight", remoteTag: releaseTestCommit + "\trefs/tags/v1.2.0\n", wantError: "not annotated"},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := newPublishRunner(t)
			runner.remoteTag = test.remoteTag
			err := RunReleasePublish(context.Background(), newTestState(runner.fake), releaseTestCommit, &fakeReleaseWaiter{})
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestRunReleasePublishRequiresActionsAndExactCheckout(t *testing.T) {
	t.Chdir(releaseFixture(t, "1.2.0"))
	runner := newPublishRunner(t)
	if err := RunReleasePublish(context.Background(), newTestState(runner.fake), releaseTestCommit, &fakeReleaseWaiter{}); err == nil || !strings.Contains(err.Error(), "restricted to GitHub Actions") {
		t.Fatalf("local publication was not rejected: %v", err)
	}
	t.Setenv("GITHUB_ACTIONS", "true")
	runner.head = releaseTestParent
	if err := RunReleasePublish(context.Background(), newTestState(runner.fake), releaseTestCommit, &fakeReleaseWaiter{}); err == nil || !strings.Contains(err.Error(), "does not equal") {
		t.Fatalf("mismatched checkout was not rejected: %v", err)
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
	published         bool
	tagged            bool
	pushReturnsError  bool
	pushReachedRemote bool
	testFailure       bool
	commitFailure     bool
	dispatchFailures  int
	commits           int
	dispatches        int
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
			case strings.HasPrefix(command, "gh workflow run release.yml"):
				runner.dispatches++
				if runner.dispatches <= runner.dispatchFailures {
					return "", errors.New("dispatch failed")
				}
				return "", nil
			case command == "gh repo view --json url --jq .url":
				return "https://github.com/fmind/dotfiles", nil
			case command == "git reset --mixed HEAD":
				runner.resets++
				return "", nil
			case strings.Contains(command, "git tag"):
				runner.tagged = true
			case strings.Contains(command, "gh release"):
				runner.published = true
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
			if strings.HasPrefix(command, "git tag") {
				runner.tagged = true
			}
			if strings.HasPrefix(command, "gh release") {
				runner.published = true
			}
			return nil
		},
	}
	return runner
}

type publishRunner struct {
	fake           *FakeRunner
	head           string
	runs           string
	remoteTag      string
	ciURL          string
	releaseURL     string
	releaseExists  bool
	tagCreates     int
	tagPushes      int
	releaseCreates int
}

func newPublishRunner(t *testing.T) *publishRunner {
	t.Helper()
	runner := &publishRunner{head: releaseTestCommit, ciURL: "https://github.com/fmind/dotfiles/actions/runs/1", releaseURL: "https://github.com/fmind/dotfiles/releases/tag/v1.2.0"}
	runner.runs = ciRuns("success")
	runner.fake = &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch {
			case command == "git rev-parse HEAD":
				return runner.head, nil
			case strings.HasPrefix(command, "gh run list --workflow ci.yml"):
				return runner.runs, nil
			case strings.HasPrefix(command, "git ls-remote --tags origin"):
				return runner.remoteTag, nil
			case command == "git tag -a v1.2.0 -m v1.2.0 "+releaseTestCommit:
				runner.tagCreates++
				return "", nil
			case command == "gh release view v1.2.0 --json url --jq .url":
				if runner.releaseExists {
					return runner.releaseURL, nil
				}
				return "", errors.New("not found")
			default:
				return "", errors.New("unexpected command: " + command)
			}
		},
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git push origin refs/tags/v1.2.0":
				runner.tagPushes++
				runner.remoteTag = annotatedTag("v1.2.0", releaseTestCommit)
				return nil
			case "gh release create v1.2.0 --verify-tag --generate-notes --title v1.2.0":
				runner.releaseCreates++
				runner.releaseExists = true
				return nil
			default:
				return errors.New("unexpected interactive command: " + command)
			}
		},
	}
	return runner
}

type fakeReleaseWaiter struct {
	now  time.Time
	step time.Duration
}

func (waiter *fakeReleaseWaiter) Wait(context.Context, time.Duration) error {
	step := waiter.step
	if step == 0 {
		step = releasePollInterval
	}
	waiter.now = waiter.now.Add(step)
	return nil
}

func (waiter *fakeReleaseWaiter) Now() time.Time { return waiter.now }

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

func ciRuns(conclusion string) string {
	return `[{"databaseId":1,"status":"completed","conclusion":"` + conclusion + `","headSha":"` + releaseTestCommit + `","url":"https://github.com/fmind/dotfiles/actions/runs/1"}]`
}

func duplicateCIRuns() string {
	run := strings.TrimSuffix(strings.TrimPrefix(ciRuns("success"), "["), "]")
	return "[" + run + "," + run + "]"
}

func annotatedTag(tag, commit string) string {
	return strings.Repeat("c", 40) + "\trefs/tags/" + tag + "\n" + commit + "\trefs/tags/" + tag + "^{}\n"
}
