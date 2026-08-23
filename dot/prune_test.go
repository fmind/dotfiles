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
	"time"

	"github.com/urfave/cli/v3"
)

// recordedRunner captures every external command a prune target issues.
type recordedRunner struct {
	FakeRunner
	calls []string
}

func newRecordedRunner(fail map[string]error) *recordedRunner {
	runner := &recordedRunner{}
	runner.RunFunc = func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
		call := strings.TrimSpace(name + " " + strings.Join(args, " "))
		runner.calls = append(runner.calls, call)
		return "", fail[call]
	}
	return runner
}

func (r *recordedRunner) ran(call string) bool {
	for _, got := range r.calls {
		if got == call {
			return true
		}
	}
	return false
}

// newPruneTestState wires a state whose output is captured and whose home and cache
// directories are redirected into t.TempDir().
func newPruneTestState(t *testing.T, runner Runner) (*GlobalState, *bytes.Buffer, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))

	out := &bytes.Buffer{}
	state := newTestState(runner)
	state.Stdout = out
	return state, out, home
}

// writeAged creates a file with the given contents and modification time.
func writeAged(t *testing.T, path, contents string, age time.Duration) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	stamp := time.Now().Add(-age)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatalf("failed to age %s: %v", path, err)
	}
}

func writeCompleteSessionSuccessor(t *testing.T, agent, sessionID, sourcePath string) sessionIngestionResult {
	t.Helper()
	fingerprint, err := fingerprintFile(sourcePath)
	if err != nil {
		t.Fatalf("failed to fingerprint raw session source: %v", err)
	}
	logs := []SessionLogLine{{Agent: agent, SID: sessionID, Role: "user", Content: "normalized"}}
	result, err := ingestSession(context.Background(), agent, sessionID, logs, sessionSource{
		Type:        agent + "-test",
		Fingerprint: fingerprint,
	})
	if err != nil {
		t.Fatalf("failed to create normalized successor: %v", err)
	}
	if result.Status != sessionIngested {
		t.Fatalf("expected successor ingestion, got %s", result.Status)
	}
	return result
}

func requireExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func requireGone(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be removed, got %v", path, err)
	}
}

func TestPruneAgentSessions(t *testing.T) {
	t.Run("deletes only expired session logs", func(t *testing.T) {
		state, out, home := newPruneTestState(t, &FakeRunner{})
		projects := filepath.Join(home, ".claude", "projects")
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.claude/projects", KeepDays: 7}}

		expired := filepath.Join(projects, "stale", "old.jsonl")
		fresh := filepath.Join(projects, "live", "new.jsonl")
		writeAged(t, expired, "0123456789", 30*24*time.Hour)
		writeAged(t, fresh, "still relevant", time.Hour)
		writeCompleteSessionSuccessor(t, sessionStoreClaude, "old", expired)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    7,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}

		requireGone(t, expired)
		requireExists(t, fresh)
		// The directory left empty by the deletion goes with it; the populated one stays.
		requireGone(t, filepath.Join(projects, "stale"))
		requireExists(t, filepath.Join(projects, "live"))

		if !strings.Contains(out.String(), "deleted 1 file(s) older than 7 days in ~/.claude/projects (10 B)") {
			t.Fatalf("unexpected report: %s", out.String())
		}
		if !strings.Contains(out.String(), "Reclaimed 10 B") {
			t.Fatalf("expected the reclaimed total in the summary: %s", out.String())
		}
	})

	t.Run("preserves agent memory", func(t *testing.T) {
		state, _, home := newPruneTestState(t, &FakeRunner{})
		projects := filepath.Join(home, ".claude", "projects")
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.claude/projects", KeepDays: 7}}

		memoryNote := filepath.Join(projects, "proj", "memory", "preference.md")
		memoryIndex := filepath.Join(projects, "proj", "memory", "MEMORY.md")
		memoryLog := filepath.Join(projects, "proj", "memory.jsonl")
		sessionLog := filepath.Join(projects, "proj", "session.jsonl")
		for _, path := range []string{memoryNote, memoryIndex, memoryLog, sessionLog} {
			writeAged(t, path, "aged out", 90*24*time.Hour)
		}
		writeCompleteSessionSuccessor(t, sessionStoreClaude, "session", sessionLog)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    7,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}

		requireGone(t, sessionLog)
		for _, path := range []string{memoryNote, memoryIndex, memoryLog} {
			requireExists(t, path)
		}
	})

	t.Run("missing session stores are not an error", func(t *testing.T) {
		state, out, _ := newPruneTestState(t, &FakeRunner{})
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.codex/sessions", KeepDays: 7}}

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    7,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}
		if !strings.Contains(out.String(), "no session stores found") {
			t.Fatalf("unexpected report: %s", out.String())
		}
	})

	t.Run("dry run reports without deleting", func(t *testing.T) {
		state, out, home := newPruneTestState(t, &FakeRunner{})
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.agents/sessions", KeepDays: 7}}
		expired := filepath.Join(home, ".agents", "sessions", "2020-01-01", "old.jsonl")
		writeAged(t, expired, "0123456789", 30*24*time.Hour)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    7,
			DryRun:  true,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}

		requireExists(t, expired)
		report := out.String()
		if !strings.Contains(report, "Prune (dry run)") ||
			!strings.Contains(report, "would delete 1 file(s) older than 7 days in ~/.agents/sessions") ||
			!strings.Contains(report, "Would reclaim 10 B") {
			t.Fatalf("unexpected dry run report: %s", report)
		}
	})
}

func TestPruneDocker(t *testing.T) {
	t.Run("build level keeps containers", func(t *testing.T) {
		runner := newRecordedRunner(nil)
		state, _, _ := newPruneTestState(t, runner)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"docker": levelBuild},
			Days:    7,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}

		if !runner.ran("docker builder prune -af") {
			t.Fatalf("expected the build cache to be pruned, got %v", runner.calls)
		}
		if runner.ran("docker system prune -f") {
			t.Fatalf("system prune must stay behind the deeper level, got %v", runner.calls)
		}
	})

	t.Run("system level prunes daemon resources", func(t *testing.T) {
		runner := newRecordedRunner(nil)
		state, _, _ := newPruneTestState(t, runner)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"docker": levelSystem},
			Days:    7,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}
		if !runner.ran("docker system prune -f") {
			t.Fatalf("expected a system prune, got %v", runner.calls)
		}
	})

	t.Run("skips when the daemon is down", func(t *testing.T) {
		runner := newRecordedRunner(map[string]error{"docker info": errors.New("cannot connect")})
		state, out, _ := newPruneTestState(t, runner)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"docker": levelSystem},
			Days:    7,
		}); err != nil {
			t.Fatalf("a stopped daemon must not fail the prune: %v", err)
		}
		if runner.ran("docker builder prune -af") {
			t.Fatalf("nothing should run without a daemon, got %v", runner.calls)
		}
		if !strings.Contains(out.String(), "docker daemon is not running") {
			t.Fatalf("unexpected report: %s", out.String())
		}
	})

	t.Run("skips when docker is missing", func(t *testing.T) {
		runner := newRecordedRunner(nil)
		runner.LookPathFunc = func(name string) (string, error) {
			if name == "docker" {
				return "", errors.New("not found")
			}
			return "/usr/bin/" + name, nil
		}
		state, out, _ := newPruneTestState(t, runner)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"docker": levelBuild},
			Days:    7,
		}); err != nil {
			t.Fatalf("a missing tool must not fail the prune: %v", err)
		}
		if !strings.Contains(out.String(), "docker is not installed") {
			t.Fatalf("unexpected report: %s", out.String())
		}
	})
}

func TestPruneLanguageTargets(t *testing.T) {
	tests := []struct {
		name   string
		target string
		level  string
		want   []string
		unwant []string
	}{
		{
			name:   "go build level keeps the module cache",
			target: "go",
			level:  levelBuild,
			want:   []string{"go clean -cache -testcache"},
			unwant: []string{"go clean -modcache"},
		},
		{
			name:   "go module level clears everything",
			target: "go",
			level:  levelModule,
			want:   []string{"go clean -cache -testcache", "go clean -modcache"},
		},
		{
			name:   "python cache level only prunes uv",
			target: "python",
			level:  levelCache,
			want:   []string{"uv cache prune"},
			unwant: []string{"uv cache clean", "pip cache purge"},
		},
		{
			name:   "python all level wipes uv and pip",
			target: "python",
			level:  levelAll,
			want:   []string{"uv cache clean", "pip cache purge"},
			unwant: []string{"uv cache prune"},
		},
		{
			name:   "node cache level only removes npx",
			target: "node",
			level:  levelCache,
			unwant: []string{"npm cache clean --force"},
		},
		{
			name:   "node all level cleans npm",
			target: "node",
			level:  levelAll,
			want:   []string{"npm cache clean --force"},
		},
		{
			name:   "mise cache level keeps config links",
			target: "mise",
			level:  levelCache,
			want:   []string{"mise prune -y", "mise cache clear"},
			unwant: []string{"mise prune --configs -y"},
		},
		{
			name:   "mise configs level prunes links too",
			target: "mise",
			level:  levelConfigs,
			want:   []string{"mise prune --configs -y"},
		},
		{
			name:   "tools clears linter and scanner caches",
			target: "tools",
			level:  levelCache,
			want:   []string{"dprint clear-cache", "golangci-lint cache clean"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := newRecordedRunner(nil)
			state, _, _ := newPruneTestState(t, runner)

			if err := RunPrune(context.Background(), state, PruneOptions{
				Targets: map[string]string{test.target: test.level},
				Days:    7,
			}); err != nil {
				t.Fatalf("RunPrune returned an error: %v", err)
			}

			for _, call := range test.want {
				if !runner.ran(call) {
					t.Fatalf("expected %q to run, got %v", call, runner.calls)
				}
			}
			for _, call := range test.unwant {
				if runner.ran(call) {
					t.Fatalf("did not expect %q to run, got %v", call, runner.calls)
				}
			}
		})
	}
}

func TestPruneRemovesCacheDirectories(t *testing.T) {
	runner := newRecordedRunner(nil)
	state, out, home := newPruneTestState(t, runner)

	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("failed to resolve the cache directory: %v", err)
	}
	npx := filepath.Join(home, ".npm", "_npx", "abcdef", "package.json")
	trivy := filepath.Join(cache, "trivy", "db", "trivy.db")
	helm := filepath.Join(cache, "helm", "repository", "index.yaml")
	downloads := filepath.Join(home, ".local", "share", "mise", "downloads", "go", "go.tar.gz")
	for _, path := range []string{npx, trivy, helm, downloads} {
		writeAged(t, path, "cached", time.Hour)
	}

	if err := RunPrune(context.Background(), state, PruneOptions{
		Targets: map[string]string{"node": levelCache, "mise": levelCache, "tools": levelCache},
		Days:    7,
	}); err != nil {
		t.Fatalf("RunPrune returned an error: %v", err)
	}

	requireGone(t, filepath.Join(home, ".npm", "_npx"))
	requireGone(t, filepath.Join(cache, "trivy"))
	requireGone(t, filepath.Join(cache, "helm"))
	requireGone(t, filepath.Dir(downloads))
	// mise recreates entries inside its download directory but expects the directory
	// itself to exist, so only the contents go.
	requireExists(t, filepath.Join(home, ".local", "share", "mise", "downloads"))

	if !strings.Contains(out.String(), "Reclaimed 24 B") {
		t.Fatalf("expected the reclaimed total of four 6-byte files: %s", out.String())
	}
}

func TestPruneConfiguredLevels(t *testing.T) {
	t.Run("a bare flag uses the configured depth", func(t *testing.T) {
		runner := newRecordedRunner(nil)
		state, _, _ := newPruneTestState(t, runner)
		// A machine that never runs a local k3d cluster can make the deep prune the norm.
		state.Config.Prune.Docker.Level = levelSystem

		app := &cli.Command{Writer: io.Discard, Commands: []*cli.Command{NewPruneCmd(state)}}
		if err := app.Run(context.Background(), []string{"dot", "prune", "--docker"}); err != nil {
			t.Fatalf("prune failed: %v", err)
		}
		if !runner.ran("docker system prune -f") {
			t.Fatalf("expected the configured level to apply, got %v", runner.calls)
		}
	})

	t.Run("an explicit level still wins", func(t *testing.T) {
		runner := newRecordedRunner(nil)
		state, _, _ := newPruneTestState(t, runner)
		state.Config.Prune.Docker.Level = levelSystem

		app := &cli.Command{Writer: io.Discard, Commands: []*cli.Command{NewPruneCmd(state)}}
		if err := app.Run(context.Background(), []string{"dot", "prune", "--docker=build"}); err != nil {
			t.Fatalf("prune failed: %v", err)
		}
		if runner.ran("docker system prune -f") {
			t.Fatalf("expected --docker=build to override the config, got %v", runner.calls)
		}
	})

	t.Run("an unknown configured level is rejected", func(t *testing.T) {
		runner := newRecordedRunner(nil)
		state, _, _ := newPruneTestState(t, runner)
		state.Config.Prune.Go.Level = "everything"

		app := &cli.Command{Writer: io.Discard, Commands: []*cli.Command{NewPruneCmd(state)}}
		err := app.Run(context.Background(), []string{"dot", "prune", "--go"})
		if err == nil || !strings.Contains(err.Error(), "invalid prune.go.level") {
			t.Fatalf("expected the misconfigured level to be reported, got %v", err)
		}
		if len(runner.calls) != 0 {
			t.Fatalf("nothing should run on a bad config, got %v", runner.calls)
		}
	})
}

func TestPruneConfiguredPaths(t *testing.T) {
	runner := newRecordedRunner(nil)
	state, _, home := newPruneTestState(t, runner)

	npx := filepath.Join(home, "custom-npx")
	staging := filepath.Join(home, "custom-mise")
	toolCache := filepath.Join(home, "custom-tools")
	state.Config.Prune.Node.Paths = []string{npx}
	state.Config.Prune.Mise.Paths = []string{staging}
	state.Config.Prune.Tools.Paths = []string{toolCache}
	for _, path := range []string{npx, staging, toolCache} {
		writeAged(t, filepath.Join(path, "entry", "file.bin"), "cached", time.Hour)
	}

	if err := RunPrune(context.Background(), state, PruneOptions{
		Targets: map[string]string{"node": levelCache, "mise": levelCache, "tools": levelCache},
		Days:    pruneDaysFromConfig,
	}); err != nil {
		t.Fatalf("RunPrune returned an error: %v", err)
	}

	requireGone(t, npx)
	requireGone(t, toolCache)
	// mise keeps its staging directory and loses only the contents.
	requireExists(t, staging)
	requireGone(t, filepath.Join(staging, "entry"))
}

func TestRunPruneReportsFailuresWithoutStopping(t *testing.T) {
	runner := newRecordedRunner(map[string]error{"go clean -cache -testcache": errors.New("disk error")})
	state, out, _ := newPruneTestState(t, runner)

	err := RunPrune(context.Background(), state, PruneOptions{
		Targets: map[string]string{"go": levelBuild, "python": levelCache},
		Days:    7,
	})
	if err == nil {
		t.Fatal("expected the failing target to surface an error")
	}
	if !strings.Contains(err.Error(), "disk error") {
		t.Fatalf("expected the underlying failure to be wrapped, got %v", err)
	}
	// A failing target must not cancel the ones after it.
	if !runner.ran("uv cache prune") {
		t.Fatalf("expected later targets to still run, got %v", runner.calls)
	}
	if !strings.Contains(out.String(), failIcon+" go:") {
		t.Fatalf("expected the failure to be reported inline: %s", out.String())
	}
}

func TestRunPruneWithoutTargets(t *testing.T) {
	state, out, home := newPruneTestState(t, &FakeRunner{})
	state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.agents/sessions", KeepDays: 7}}
	expired := filepath.Join(home, ".agents", "sessions", "old.jsonl")
	writeAged(t, expired, "keep me", 90*24*time.Hour)

	if err := RunPrune(context.Background(), state, PruneOptions{Days: 7}); err != nil {
		t.Fatalf("an empty selection must be a no-op, got %v", err)
	}

	requireExists(t, expired)
	report := out.String()
	if !strings.Contains(report, "No target selected") || !strings.Contains(report, "--agents") {
		t.Fatalf("expected the available targets to be listed: %s", report)
	}
}

func TestPruneRetention(t *testing.T) {
	// Each store keeps its own history, so the same file age expires in one and not the
	// other. Both files are aged 10 days: past the 7 day store, inside the 30 day one.
	setup := func(t *testing.T) (*GlobalState, *bytes.Buffer, string, string) {
		t.Helper()
		state, out, home := newPruneTestState(t, &FakeRunner{})
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{
			{Path: "~/.codex/sessions", KeepDays: 7},
			{Path: "~/.agents/sessions", KeepDays: 30},
		}
		raw := filepath.Join(home, ".codex", "sessions", "rollout-2026-07-09T12-00-00-retention-session.jsonl")
		archived := filepath.Join(home, ".agents", "sessions", "session.jsonl")
		writeAged(t, raw, "raw", 10*24*time.Hour)
		writeAged(t, archived, "archived", 10*24*time.Hour)
		writeCompleteSessionSuccessor(t, sessionStoreCodex, "retention-session", raw)
		return state, out, raw, archived
	}

	t.Run("per directory keep_days applies", func(t *testing.T) {
		state, out, raw, archived := setup(t)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    pruneDaysFromConfig,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}

		requireGone(t, raw)
		requireExists(t, archived)
		report := out.String()
		if !strings.Contains(report, "older than 7 days in ~/.codex/sessions") ||
			!strings.Contains(report, "older than 30 days in ~/.agents/sessions") {
			t.Fatalf("expected a line per store with its own retention: %s", report)
		}
	})

	t.Run("days overrides every directory", func(t *testing.T) {
		state, _, raw, archived := setup(t)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    60,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}
		requireExists(t, raw)
		requireExists(t, archived)
	})

	t.Run("zero days deletes everything", func(t *testing.T) {
		state, _, raw, archived := setup(t)

		if err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    0,
		}); err != nil {
			t.Fatalf("RunPrune returned an error: %v", err)
		}
		requireGone(t, raw)
		requireGone(t, archived)
	})

	t.Run("negative days are rejected", func(t *testing.T) {
		state, _, _, _ := setup(t)

		err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    -2,
		})
		if err == nil || !strings.Contains(err.Error(), "cannot be negative") {
			t.Fatalf("expected a retention validation error, got %v", err)
		}
	})

	t.Run("a negative keep_days is rejected", func(t *testing.T) {
		state, _, _, _ := setup(t)
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.codex/sessions", KeepDays: -1}}

		err := RunPrune(context.Background(), state, PruneOptions{
			Targets: map[string]string{"agents": levelSessions},
			Days:    pruneDaysFromConfig,
		})
		if err == nil || !strings.Contains(err.Error(), "keep_days cannot be negative") {
			t.Fatalf("expected the misconfigured directory to be reported, got %v", err)
		}
	})
}

// runPruneCommand parses args through the real command so flag composition is covered
// end to end, and reports the resulting target selection.
func runPruneCommand(t *testing.T, args ...string) (map[string]string, int, bool, error) {
	t.Helper()
	var (
		targets map[string]string
		days    int
		dryRun  bool
	)
	app := &cli.Command{
		Writer: io.Discard,
		Commands: []*cli.Command{{
			Name:  "prune",
			Flags: pruneFlags(),
			Action: func(_ context.Context, cmd *cli.Command) error {
				resolved, err := resolvePruneTargets(cmd, &DefaultConfig().Prune)
				if err != nil {
					return err
				}
				targets, days, dryRun = resolved, int(cmd.Int("days")), cmd.Bool("dry-run")
				return nil
			},
		}},
	}
	err := app.Run(context.Background(), append([]string{"dot", "prune"}, args...))
	return targets, days, dryRun, err
}

func TestPruneFlagComposition(t *testing.T) {
	tests := []struct {
		want map[string]string
		name string
		args []string
	}{
		{
			name: "no flags selects nothing",
			args: nil,
			want: map[string]string{},
		},
		{
			name: "bare flags select the shallowest level",
			args: []string{"--agents", "--docker"},
			want: map[string]string{"agents": levelSessions, "docker": levelBuild},
		},
		{
			name: "short flags compose",
			args: []string{"-g", "-p", "-t"},
			want: map[string]string{"go": levelBuild, "python": levelCache, "tools": levelCache},
		},
		{
			name: "explicit levels are honored",
			args: []string{"--docker=system", "--go=module"},
			want: map[string]string{"docker": levelSystem, "go": levelModule},
		},
		{
			name: "all selects every target at its default level",
			args: []string{"--all"},
			want: map[string]string{
				"agents": levelSessions, "docker": levelBuild, "go": levelBuild,
				"python": levelCache, "node": levelCache, "mise": levelCache, "tools": levelCache,
			},
		},
		{
			name: "all=deep selects every target at its deepest level",
			args: []string{"--all=deep"},
			want: map[string]string{
				"agents": levelSessions, "docker": levelSystem, "go": levelModule,
				"python": levelAll, "node": levelAll, "mise": levelConfigs, "tools": levelCache,
			},
		},
		{
			name: "an explicit target overrides all",
			args: []string{"--all=deep", "--docker=build"},
			want: map[string]string{
				"agents": levelSessions, "docker": levelBuild, "go": levelModule,
				"python": levelAll, "node": levelAll, "mise": levelConfigs, "tools": levelCache,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			targets, _, _, err := runPruneCommand(t, test.args...)
			if err != nil {
				t.Fatalf("failed to parse %v: %v", test.args, err)
			}
			if len(targets) != len(test.want) {
				t.Fatalf("expected %v, got %v", test.want, targets)
			}
			for name, level := range test.want {
				if targets[name] != level {
					t.Fatalf("expected %s at %q, got %q", name, level, targets[name])
				}
			}
		})
	}
}

func TestPruneFlagModifiers(t *testing.T) {
	t.Run("days and dry-run are parsed", func(t *testing.T) {
		_, days, dryRun, err := runPruneCommand(t, "--agents", "--days=21", "--dry-run")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if days != 21 || !dryRun {
			t.Fatalf("expected 21 days and a dry run, got %d and %v", days, dryRun)
		}
	})

	t.Run("an unknown level is rejected", func(t *testing.T) {
		_, _, _, err := runPruneCommand(t, "--go=everything")
		if err == nil || !strings.Contains(err.Error(), "expected one of: build, module") {
			t.Fatalf("expected the accepted levels to be listed, got %v", err)
		}
	})
}

func TestNewPruneCmdUsesConfiguredRetention(t *testing.T) {
	state, out, home := newPruneTestState(t, &FakeRunner{})
	state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.agents/sessions", KeepDays: 30}}

	// Older than the 7 days the raw stores keep, but within this store's 30.
	recent := filepath.Join(home, ".agents", "sessions", "recent.jsonl")
	writeAged(t, recent, "recent", 10*24*time.Hour)

	app := &cli.Command{Writer: io.Discard, Commands: []*cli.Command{NewPruneCmd(state)}}
	if err := app.Run(context.Background(), []string{"dot", "prune", "--agents"}); err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	requireExists(t, recent)
	if !strings.Contains(out.String(), "older than 30 days") {
		t.Fatalf("expected the configured retention to apply: %s", out.String())
	}

	// An explicit --days beats the configured retention.
	if err := app.Run(context.Background(), []string{"dot", "prune", "--agents", "--days=1"}); err != nil {
		t.Fatalf("prune failed: %v", err)
	}
	requireGone(t, recent)
}

func TestPruneCmdAliasAndFlags(t *testing.T) {
	state, _, _ := newPruneTestState(t, &FakeRunner{})
	cmd := NewPruneCmd(state)

	if !cmd.HasName("x") {
		t.Fatal("expected the prune command to answer to its alias")
	}
	// One flag per target, plus --all, --days, and --dry-run.
	if got, want := len(cmd.Flags), len(pruneTargets)+3; got != want {
		t.Fatalf("expected %d flags, got %d", want, got)
	}
	for _, flag := range cmd.Flags {
		if levelFlag, ok := flag.(*pruneLevelFlag); ok {
			if !strings.Contains(levelFlag.String(), "[=") {
				t.Fatalf("expected %q to advertise its optional levels", levelFlag.String())
			}
		}
	}
}

func TestRemoveEmptyDirs(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	populated := filepath.Join(root, "keep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create %s: %v", nested, err)
	}
	writeAged(t, filepath.Join(populated, "file.txt"), "data", time.Hour)

	dirs := []string{filepath.Join(root, "a"), nested, populated, filepath.Join(root, "gone")}
	if err := removeEmptyDirs(dirs); err != nil {
		t.Fatalf("removeEmptyDirs returned an error: %v", err)
	}

	// The parent is removed in the same pass as the child that emptied it.
	requireGone(t, filepath.Join(root, "a"))
	requireExists(t, populated)
}

func TestDirBytes(t *testing.T) {
	root := t.TempDir()
	writeAged(t, filepath.Join(root, "one.txt"), "12345", time.Hour)
	writeAged(t, filepath.Join(root, "nested", "two.txt"), "678", time.Hour)

	total, err := dirBytes(root)
	if err != nil {
		t.Fatalf("dirBytes returned an error: %v", err)
	}
	if total != 8 {
		t.Fatalf("expected 8 bytes, got %d", total)
	}

	missing, err := dirBytes(filepath.Join(root, "absent"))
	if err != nil || missing != 0 {
		t.Fatalf("expected a missing path to measure 0 bytes, got %d (%v)", missing, err)
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		want  string
		bytes int64
	}{
		{bytes: 0, want: "0 B"},
		{bytes: 512, want: "512 B"},
		{bytes: 1024, want: "1.0 KiB"},
		{bytes: 1536, want: "1.5 KiB"},
		{bytes: 1024 * 1024, want: "1.0 MiB"},
		{bytes: 3 * 1024 * 1024 * 1024, want: "3.0 GiB"},
	}
	for _, test := range tests {
		if got := humanBytes(test.bytes); got != test.want {
			t.Fatalf("humanBytes(%d) = %q, want %q", test.bytes, got, test.want)
		}
	}
}

func TestPruneLevelValue(t *testing.T) {
	var destination string
	value := pruneLevelValue{}.Create("", &destination, cli.NoConfig{})

	levelValue, ok := value.(*pruneLevelValue)
	if !ok || !levelValue.IsBoolFlag() {
		t.Fatal("the level value must be usable without an explicit value")
	}
	if err := value.Set(levelSystem); err != nil {
		t.Fatalf("Set returned an error: %v", err)
	}
	if destination != levelSystem || value.String() != levelSystem || value.Get() != levelSystem {
		t.Fatalf("expected the destination to hold %q, got %q", levelSystem, destination)
	}
	if got := (pruneLevelValue{}).ToString(levelDeep); got != levelDeep {
		t.Fatalf("ToString(%q) = %q", levelDeep, got)
	}
	// An unbound value renders empty instead of panicking on the nil destination.
	if got := (&pruneLevelValue{}).String(); got != "" {
		t.Fatalf("expected an empty string for an unbound value, got %q", got)
	}
}

func TestDefaultPruneConfig(t *testing.T) {
	state, _, home := newPruneTestState(t, &FakeRunner{})
	cfg := state.Config.Prune

	want := map[string]PruneSessionStore{
		"~/.claude/projects":                  {Source: sessionStoreClaude, KeepDays: 30},
		"~/.codex/sessions":                   {Source: sessionStoreCodex, KeepDays: 30},
		"~/.gemini/antigravity-cli/brain":     {Source: sessionStoreAgy, KeepDays: 30},
		"~/.grok/sessions":                    {Source: sessionStoreGrok, KeepDays: 30},
		"~/.local/share/opencode/opencode.db": {Source: sessionStoreOpenCode, KeepDays: 30},
		"~/.copilot/session-store.db":         {Source: sessionStoreCopilot, KeepDays: 30},
		// The normalized archive outlives the raw session logs it was distilled from,
		// and is the only durable copy once they expire.
		"~/.agents/sessions": {Source: sessionStoreArchive, KeepDays: 365},
	}
	if len(cfg.Agents.Sessions) != len(want) {
		t.Fatalf("expected %d session stores, got %v", len(want), cfg.Agents.Sessions)
	}
	for _, store := range cfg.Agents.Sessions {
		expected, known := want[store.Path]
		if !known {
			t.Fatalf("unexpected session store %s", store.Path)
		}
		if store.KeepDays != expected.KeepDays || store.Source != expected.Source {
			t.Fatalf("expected %s to use source %s and keep %d days, got %+v", store.Path, expected.Source, expected.KeepDays, store)
		}
	}
	for _, keep := range []string{"memory", "memory.jsonl", "MEMORY.md"} {
		if !slicesContains(cfg.Agents.Keep, keep) {
			t.Fatalf("expected %s to be preserved, got %v", keep, cfg.Agents.Keep)
		}
	}

	// Every tool target ships a default depth, and the file-based ones their paths.
	levels := map[string]string{
		"docker": cfg.Docker.Level, "go": cfg.Go.Level, "python": cfg.Python.Level,
		"node": cfg.Node.Level, "mise": cfg.Mise.Level, "tools": cfg.Tools.Level,
	}
	for _, target := range pruneTargets {
		level, configurable := levels[target.name]
		if !configurable {
			continue
		}
		if !slicesContains(target.levels, level) {
			t.Fatalf("prune.%s.level %q is not one of %v", target.name, level, target.levels)
		}
	}
	if !slicesContains(cfg.Node.Paths, "~/.npm/_npx") {
		t.Fatalf("expected the npx cache among the node paths, got %v", cfg.Node.Paths)
	}
	if len(cfg.Mise.Paths) != 2 || !slicesContains(cfg.Mise.Paths, "~/.local/share/mise/downloads") {
		t.Fatalf("expected the mise download staging areas, got %v", cfg.Mise.Paths)
	}
	// Cache paths resolve against this machine's cache directory and stay ~-relative.
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("failed to resolve the cache directory: %v", err)
	}
	for _, path := range cfg.Tools.Paths {
		if !strings.HasPrefix(path, "~/") {
			t.Fatalf("expected a ~-relative tool cache path, got %s", path)
		}
		if !strings.HasPrefix(ExpandPath(path), cache) {
			t.Fatalf("expected %s to resolve under %s (home %s)", path, cache, home)
		}
	}
}

func slicesContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestPruneTargetTableIsConsistent guards the table the flags, help, and execution order
// are all generated from.
func TestPruneTargetTableIsConsistent(t *testing.T) {
	cfg := DefaultConfig().Prune
	seenNames := map[string]bool{}
	seenAliases := map[string]bool{}
	for _, target := range pruneTargets {
		if len(target.levels) == 0 {
			t.Fatalf("target %s declares no level", target.name)
		}
		if target.prune == nil {
			t.Fatalf("target %s has no implementation", target.name)
		}
		if seenNames[target.name] {
			t.Fatalf("duplicate target name %s", target.name)
		}
		if seenAliases[target.alias] {
			t.Fatalf("duplicate target alias %s", target.alias)
		}
		seenNames[target.name], seenAliases[target.alias] = true, true

		configured, err := target.configured(&cfg)
		if err != nil {
			t.Fatalf("the default config must be valid for %s: %v", target.name, err)
		}
		if target.resolve(pruneLevelBare, configured) != configured {
			t.Fatalf("a bare --%s must resolve to its configured level %s", target.name, configured)
		}
		if target.resolve(target.deepest(), configured) != target.deepest() {
			t.Fatalf("an explicit level must survive resolution for --%s", target.name)
		}
	}
	if seenAliases["A"] {
		t.Fatal("the -A alias is reserved for --all")
	}
}

func TestPruneRunReportHelpers(t *testing.T) {
	state, out, _ := newPruneTestState(t, &FakeRunner{})
	run := &pruneRun{state: state, days: 7}

	run.passf("docker", "removed %d layer(s)", 3)
	run.skipf("docker", "%s is not installed", "docker")

	report := out.String()
	if !strings.Contains(report, "docker: removed 3 layer(s)") || !strings.Contains(report, "docker: docker is not installed") {
		t.Fatalf("unexpected report: %s", report)
	}
}

func TestPruneExecSkipsMissingTools(t *testing.T) {
	runner := newRecordedRunner(nil)
	runner.LookPathFunc = func(string) (string, error) { return "", errors.New("not found") }
	state, out, _ := newPruneTestState(t, runner)
	run := &pruneRun{state: state, days: 7}

	if err := run.exec(context.Background(), "tools", "cleared", "dprint", "clear-cache"); err != nil {
		t.Fatalf("a missing tool must not be an error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no command to run, got %v", runner.calls)
	}
	if !strings.Contains(out.String(), "dprint is not installed") {
		t.Fatalf("unexpected report: %s", out.String())
	}
}

func TestPruneRemovePathsReportsUnreadablePaths(t *testing.T) {
	state, _, home := newPruneTestState(t, &FakeRunner{})
	run := &pruneRun{state: state, days: 7}

	// A file where a directory is expected still measures and removes cleanly.
	blocker := filepath.Join(home, "cache")
	writeAged(t, blocker, "0123456789", time.Hour)
	if err := run.removeTree("tools", blocker); err != nil {
		t.Fatalf("removeTree returned an error: %v", err)
	}
	requireGone(t, blocker)
	if run.freed != 10 {
		t.Fatalf("expected 10 reclaimed bytes, got %d", run.freed)
	}

	if err := run.removeContents("tools", filepath.Join(home, "absent")); err != nil {
		t.Fatalf("a missing directory must not be an error: %v", err)
	}
}

func TestPruneSessionsSurfacesWalkErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	state, _, home := newPruneTestState(t, &FakeRunner{})
	run := &pruneRun{state: state, days: 7}

	locked := filepath.Join(home, "locked")
	writeAged(t, filepath.Join(locked, "nested", "old.jsonl"), "old", 90*24*time.Hour)
	if err := os.Chmod(filepath.Join(locked, "nested"), 0o000); err != nil {
		t.Fatalf("failed to lock the directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(locked, "nested"), 0o755) })

	_, _, err := run.pruneSessions(locked, time.Now(), nil)
	if err == nil {
		t.Fatal("expected an unreadable directory to surface an error")
	}
	if !strings.Contains(err.Error(), "failed to scan") {
		t.Fatalf("expected the scan failure to be wrapped, got %v", err)
	}
}

func TestPruneAgentSessionsAggregatesErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	state, out, home := newPruneTestState(t, &FakeRunner{})
	state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/locked", KeepDays: 7}, {Path: "~/open", KeepDays: 7}}

	writeAged(t, filepath.Join(home, "locked", "nested", "old.jsonl"), "old", 90*24*time.Hour)
	if err := os.Chmod(filepath.Join(home, "locked", "nested"), 0o000); err != nil {
		t.Fatalf("failed to lock the directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(home, "locked", "nested"), 0o755) })
	open := filepath.Join(home, "open", "old.jsonl")
	writeAged(t, open, "old", 90*24*time.Hour)

	err := RunPrune(context.Background(), state, PruneOptions{
		Targets: map[string]string{"agents": levelSessions},
		Days:    7,
	})
	if err == nil {
		t.Fatal("expected the unreadable directory to be reported")
	}
	// The readable store is still pruned despite the failure on the other one.
	requireGone(t, open)
	if !strings.Contains(out.String(), failIcon+" agents:") {
		t.Fatalf("expected an inline failure line: %s", out.String())
	}
}
