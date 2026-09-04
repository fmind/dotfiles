package dot

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// A duration must survive `config init` -> `config validate`. yaml.v3 marshals a bare
// time.Duration as a nanosecond integer that it then refuses to decode back, so the
// wrapper is what keeps the scaffolded file loadable.
func TestDurationRoundTripsThroughGeneratedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot.yaml")
	state := newTestState(&FakeRunner{})
	state.ConfigPath = path
	state.Stdout = io.Discard

	if err := RunConfigInit(state, false); err != nil {
		t.Fatalf("config init: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "stale_lag: 24h0m0s") {
		t.Fatalf("durations were not written in readable form:\n%s", content)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("scaffolded config did not reload: %v", err)
	}
	if got := loaded.Agent.staleLag(); got != defaultDoctorStaleLag {
		t.Fatalf("stale_lag round-tripped to %s, want %s", got, defaultDoctorStaleLag)
	}
	if got := loaded.Pull.Timeout.Duration(); got != defaultPullTimeout {
		t.Fatalf("pull timeout round-tripped to %s, want %s", got, defaultPullTimeout)
	}
}

func TestDurationRejectsUnusableValues(t *testing.T) {
	for name, input := range map[string]string{
		"bare integer":     "3000",
		"unparseable text": "\"soon\"",
		"zero":             "\"0s\"",
		"negative":         "\"-5s\"",
	} {
		t.Run(name, func(t *testing.T) {
			var value Duration
			if err := yaml.Unmarshal([]byte(input), &value); err == nil {
				t.Fatalf("%s was accepted as a duration (%s)", input, value.Duration())
			}
		})
	}
}

// A non-positive knob must fall back to the built-in default rather than disabling a
// timeout or serializing a worker pool.
func TestConfigKnobsFallBackWhenNonPositive(t *testing.T) {
	cfg := AgentConfig{}
	if got := cfg.scanLimit(); got != defaultDoctorScanLimit {
		t.Errorf("scanLimit() = %d, want %d", got, defaultDoctorScanLimit)
	}
	if got := cfg.staleLag(); got != defaultDoctorStaleLag {
		t.Errorf("staleLag() = %s, want %s", got, defaultDoctorStaleLag)
	}
	if got := cfg.hookFailureLimit(); got != defaultHookFailureLimit {
		t.Errorf("hookFailureLimit() = %d, want %d", got, defaultHookFailureLimit)
	}
	if got := positiveOr(0, defaultPullConcurrency); got != defaultPullConcurrency {
		t.Errorf("positiveOr(0, %d) = %d", defaultPullConcurrency, got)
	}
	if got := positiveOr(3, defaultPullConcurrency); got != 3 {
		t.Errorf("positiveOr kept %d instead of the configured 3", got)
	}
}

func TestAgentSourceRootPrefersConfiguredOverride(t *testing.T) {
	cfg := defaultAgentConfig()
	cfg.Sources[sessionStoreClaude] = "~/elsewhere/projects"

	root, err := cfg.SourceRoot(sessionStoreClaude)
	if err != nil {
		t.Fatal(err)
	}
	if root != ExpandPath("~/elsewhere/projects") {
		t.Fatalf("SourceRoot ignored the override: %s", root)
	}
	if _, unknownErr := cfg.SourceRoot("nonexistent"); unknownErr == nil {
		t.Fatal("an unknown agent must not resolve to a path")
	}

	// A blank override is treated as absent so a stray key cannot silently point
	// ingestion at the working directory.
	cfg.Sources[sessionStoreCodex] = "   "
	root, err = cfg.SourceRoot(sessionStoreCodex)
	if err != nil {
		t.Fatal(err)
	}
	if root != ExpandPath("~/.codex/sessions") {
		t.Fatalf("blank override was honored: %s", root)
	}
}

// Retention must cover exactly the stores ingestion reads, plus the archive. Deriving
// both from one table is what prevents prune from expiring an unread directory.
func TestRetentionDefaultsCoverEveryIngestedSource(t *testing.T) {
	stores := defaultRawSessionStores()
	bySource := make(map[string]PruneSessionStore, len(stores))
	for _, store := range stores {
		bySource[store.Source] = store
	}
	for _, definition := range agentDefinitions() {
		store, ok := bySource[definition.Agent]
		if !ok {
			t.Fatalf("agent %q has no retention entry", definition.Agent)
		}
		if store.Path != definition.Source {
			t.Errorf("agent %q prunes %s but ingests %s", definition.Agent, store.Path, definition.Source)
		}
	}
	if archive, ok := bySource[sessionStoreArchive]; !ok {
		t.Fatal("the normalized archive has no retention entry")
	} else if archive.KeepDays <= defaultRawSessionKeepDays {
		t.Errorf("archive retention %d must outlast raw sources (%d)", archive.KeepDays, defaultRawSessionKeepDays)
	}
}

func TestHookCommandArgumentsStopAtOperands(t *testing.T) {
	agents := agentNameSet()
	for command, want := range map[string]string{
		"dot agent hook session claude":      "agent hook session",
		"dot agent hook usage claude":        "agent hook usage",
		"dot agent hook notify agy stop":     "agent hook notify",
		"dot agent hook copilot-session-end": "agent hook copilot-session-end",
		"dot agent hook session opencode":    "agent hook session",
		"dot agent doctor --fix":             "agent doctor",
		"chezmoi apply":                      "",
	} {
		got := strings.Join(hookCommandArguments(command, agents), " ")
		if got != want {
			t.Errorf("hookCommandArguments(%q) = %q, want %q", command, got, want)
		}
	}
}

// Every hook command the doctor expects must actually be wired somewhere in the
// managed dotfiles, and must resolve to a real command in dot's own tree. Without
// this, the table can drift from what chezmoi deploys and the doctor would demand a
// command no config sets — or bless one dot cannot run.
func TestHookCommandsAreWiredAndResolvable(t *testing.T) {
	repo := repositoryRoot(t)
	sources, err := filepath.Glob(filepath.Join(repo, "dot_*"))
	if err != nil {
		t.Fatal(err)
	}
	managed := make([]string, 0, len(sources))
	for _, root := range sources {
		walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() {
				return nil // an unreadable managed file is not this test's subject
			}
			content, readErr := os.ReadFile(path)
			if readErr == nil {
				managed = append(managed, string(content))
			}
			return nil
		})
		if walkErr != nil {
			t.Fatal(walkErr)
		}
	}

	commands := commandPaths(NewApp())
	agents := agentNameSet()
	for _, definition := range agentDefinitions() {
		for _, command := range definition.HookCommands {
			wired := false
			for _, content := range managed {
				if strings.Contains(content, command) {
					wired = true
					break
				}
			}
			if !wired {
				t.Errorf("agent %q expects hook command %q, but no managed dotfile wires it", definition.Agent, command)
			}
			path := strings.Join(hookCommandArguments(command, agents), " ")
			if _, ok := commands[path]; !ok {
				t.Errorf("agent %q expects hook command %q, which resolves to unknown command path %q", definition.Agent, command, path)
			}
		}
	}
}

// Wiring a hook command into a config file proves nothing if the deployed binary
// cannot run it: chezmoi updates hook files independently of the dot binary, so a
// config can reference a subcommand a stale install predates.
func TestAgentDoctorFlagsHookCommandsTheInstalledBinaryCannotRun(t *testing.T) {
	definition := agentDefinition{
		Agent:        sessionStoreClaude,
		HookPath:     "../dot_claude/settings.json.tmpl",
		HookCommands: []string{"dot agent hook session claude"},
	}
	if status, ok := checkAgentHooks(definition, func([]string) bool { return true }); !ok || status != "healthy" {
		t.Fatalf("a runnable hook command was rejected: %s", status)
	}
	status, ok := checkAgentHooks(definition, func([]string) bool { return false })
	if ok || status != "command-unavailable" {
		t.Fatalf("an unrunnable hook command was reported as %q (ok=%v)", status, ok)
	}
}

// The prober must consult the dot on PATH, not the running binary, and must treat a
// missing dot as "nothing can run" rather than silently passing.
func TestDotCommandProberUsesPathBinary(t *testing.T) {
	var probed []string
	runner := &FakeRunner{
		LookPathFunc: func(string) (string, error) { return "/usr/local/bin/dot", nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			probed = append(probed, name+" "+strings.Join(args, " "))
			if strings.Contains(strings.Join(args, " "), "hook") {
				return "", errors.New("No help topic for 'hook'")
			}
			return "", nil
		},
	}
	prober := newDotCommandProber(context.Background(), runner)

	if prober([]string{"agent", "session"}) != true {
		t.Error("an existing command was reported missing")
	}
	if prober([]string{"agent", "hook", "session"}) != false {
		t.Error("a missing command was reported present")
	}
	// The second lookup of the same path must be served from cache.
	prober([]string{"agent", "session"})
	if len(probed) != 2 {
		t.Errorf("prober ran %d commands, want 2 (results must be cached): %v", len(probed), probed)
	}
	if !strings.HasPrefix(probed[0], "/usr/local/bin/dot ") || !strings.HasSuffix(probed[0], "--help") {
		t.Errorf("prober did not probe the PATH binary with --help: %q", probed[0])
	}

	missing := newDotCommandProber(context.Background(), &FakeRunner{
		LookPathFunc: func(string) (string, error) { return "", errors.New("not found") },
	})
	if missing([]string{"agent", "session"}) {
		t.Error("commands must not be reported runnable when dot is absent from PATH")
	}
}
