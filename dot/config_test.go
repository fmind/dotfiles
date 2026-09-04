package dot

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestRunConfigShow(t *testing.T) {
	state := newTestState(&FakeRunner{})
	var buf bytes.Buffer
	state.Stdout = &buf

	if err := RunConfigShow(state); err != nil {
		t.Fatalf("RunConfigShow: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "cluster:") || !strings.Contains(out, "name: local") {
		t.Errorf("expected YAML with cluster config, got %q", out)
	}
}

func TestRunConfigInitAndValidate(t *testing.T) {
	// Nested path exercises the MkdirAll branch of RunConfigInit.
	path := filepath.Join(t.TempDir(), "sub", "dot.yaml")
	state := newTestState(&FakeRunner{})
	state.ConfigPath = path

	if err := RunConfigInit(state, false); err != nil {
		t.Fatalf("RunConfigInit: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file at %s: %v", path, err)
	}

	// The scaffolded file must round-trip cleanly under strict decoding.
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("scaffolded config failed to load: %v", err)
	}
	if cfg.Cluster.Name != "local" {
		t.Errorf("expected round-tripped cluster name 'local', got %q", cfg.Cluster.Name)
	}

	// Refuse to overwrite without --force, allow it with --force.
	if err := RunConfigInit(state, false); err == nil {
		t.Error("expected error when config already exists and force is false")
	}
	if err := RunConfigInit(state, true); err != nil {
		t.Errorf("expected force init to succeed, got %v", err)
	}

	var buf bytes.Buffer
	state.Stdout = &buf
	if err := RunConfigValidate(state); err != nil {
		t.Fatalf("RunConfigValidate: %v", err)
	}
	if !strings.Contains(buf.String(), "is valid") {
		t.Errorf("expected valid message, got %q", buf.String())
	}
}

func TestRunConfigValidate_Missing(t *testing.T) {
	state := newTestState(&FakeRunner{})
	state.ConfigPath = filepath.Join(t.TempDir(), "nope.yaml")
	var buf bytes.Buffer
	state.Stdout = &buf

	if err := RunConfigValidate(state); err != nil {
		t.Fatalf("expected no error for a missing config, got %v", err)
	}
	if !strings.Contains(buf.String(), "built-in defaults") {
		t.Errorf("expected defaults message, got %q", buf.String())
	}
}

func TestRunConfigValidate_Invalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot.yaml")
	// 'naem' is an unknown key: strict decoding must reject it.
	if err := os.WriteFile(path, []byte("cluster:\n  naem: typo\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	state := newTestState(&FakeRunner{})
	state.ConfigPath = path
	state.Stdout = &bytes.Buffer{}

	if err := RunConfigValidate(state); err == nil {
		t.Error("expected error for a config with an unknown key")
	}
}

func TestConfigEdit_ScaffoldsAndOpensEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot.yaml")
	var editorCall []string
	runner := &FakeRunner{
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			editorCall = append([]string{name}, args...)
			return nil
		},
	}
	state := newTestState(runner)
	state.ConfigPath = path
	// EDITOR carries a flag to exercise the Fields split.
	t.Setenv("EDITOR", "myeditor --wait")

	if err := RunConfigEdit(context.Background(), state); err != nil {
		t.Fatalf("RunConfigEdit: %v", err)
	}
	// A missing file is scaffolded before the editor opens.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config scaffolded before edit: %v", err)
	}
	want := []string{"myeditor", "--wait", path}
	if len(editorCall) != len(want) {
		t.Fatalf("expected editor call %v, got %v", want, editorCall)
	}
	for i := range want {
		if editorCall[i] != want[i] {
			t.Errorf("editor arg[%d] = %q, want %q", i, editorCall[i], want[i])
		}
	}
}

func TestConfigEdit_WhitespaceEditorFallsBackToVi(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot.yaml")
	if err := os.WriteFile(path, []byte("cluster:\n  name: local\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	var editorCall []string
	runner := &FakeRunner{
		RunInteractiveFunc: func(_ context.Context, _, name string, args ...string) error {
			editorCall = append([]string{name}, args...)
			return nil
		},
	}
	state := newTestState(runner)
	state.ConfigPath = path
	// A whitespace-only EDITOR must not panic and must fall back to vi.
	t.Setenv("EDITOR", "   ")

	if err := RunConfigEdit(context.Background(), state); err != nil {
		t.Fatalf("RunConfigEdit: %v", err)
	}
	if len(editorCall) != 2 || editorCall[0] != "vi" || editorCall[1] != path {
		t.Errorf("expected [vi %s], got %v", path, editorCall)
	}
}

func TestConfigPathCommand(t *testing.T) {
	state := newTestState(&FakeRunner{})
	state.ConfigPath = "/some/where/dot.yaml"
	var buf bytes.Buffer
	state.Stdout = &buf

	cmd := NewConfigPathCmd(state)
	if err := cmd.Action(context.Background(), cmd); err != nil {
		t.Fatalf("path action: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "/some/where/dot.yaml" {
		t.Errorf("expected resolved path output, got %q", buf.String())
	}
}

func TestAppConfigFatality(t *testing.T) {
	ctx := context.Background()
	// An unknown key fails strict decoding: a parse error (not a missing file).
	malformed := filepath.Join(t.TempDir(), "dot.yaml")
	if err := os.WriteFile(malformed, []byte("cluster:\n  naem: typo\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Run("malformed config is fatal for a non-config command", func(t *testing.T) {
		app := NewApp()
		err := app.Run(ctx, []string{"dot", "--config", malformed, "version"})
		if err == nil || !strings.Contains(err.Error(), "parse config file") {
			t.Errorf("expected a fatal parse error before the command runs, got %v", err)
		}
	})

	t.Run("explicitly requested missing config is fatal for a non-config command", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.yaml")
		app := NewApp()
		err := app.Run(ctx, []string{"dot", "--config", missing, "version"})
		if err == nil || !strings.Contains(err.Error(), "failed to read config file") {
			t.Errorf("expected an explicitly requested missing config to fail, got %v", err)
		}
	})

	t.Run("explicitly requested missing config is fatal for config validate", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.yaml")
		app := NewApp()
		err := app.Run(ctx, []string{"dot", "--config", missing, "config", "validate"})
		if err == nil || !strings.Contains(err.Error(), "failed to read config file") {
			t.Errorf("expected config validate to reject an explicitly requested missing config, got %v", err)
		}
	})

	t.Run("config path stays reachable despite a malformed config", func(t *testing.T) {
		app := NewApp()
		// `config path` only prints the resolved path; if the Before hook wrongly treated
		// the malformed config as fatal, this would return the parse error instead of nil.
		err := app.Run(ctx, []string{"dot", "--config", malformed, "config", "path"})
		if err != nil {
			t.Errorf("expected config path to remain usable for a bad file, got %v", err)
		}
	})

	t.Run("config init stays reachable despite a missing config", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.yaml")
		app := NewApp()
		if err := app.Run(ctx, []string{"dot", "--config", missing, "config", "init"}); err != nil {
			t.Fatalf("expected config init to repair a missing explicit config: %v", err)
		}
		if _, err := LoadConfig(missing); err != nil {
			t.Fatalf("config init did not write a valid file: %v", err)
		}
	})

	t.Run("config edit stays reachable despite a malformed config", func(t *testing.T) {
		t.Setenv("EDITOR", "true")
		app := NewApp()
		if err := app.Run(ctx, []string{"dot", "--config", malformed, "config", "edit"}); err != nil {
			t.Fatalf("expected config edit to repair a malformed explicit config: %v", err)
		}
	})
}

func TestConfigCommandAliases(t *testing.T) {
	state := newTestState(&FakeRunner{})
	cmd := NewConfigCmd(state)
	hasCfg := false
	hasF := false
	for _, alias := range cmd.Aliases {
		if alias == "cfg" {
			hasCfg = true
		}
		if alias == "f" {
			hasF = true
		}
	}
	if hasCfg || !hasF {
		t.Errorf("expected config command to have 'f' alias but NOT 'cfg', got: %v", cmd.Aliases)
	}
}

func TestConfigCommandActions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot.yaml")
	runner := &FakeRunner{
		RunInteractiveFunc: func(_ context.Context, _, _ string, _ ...string) error {
			return nil
		},
	}
	state := newTestState(runner)
	state.ConfigPath = path
	state.Stdout = &bytes.Buffer{}
	t.Setenv("EDITOR", "editor")

	command := NewConfigCmd(state)
	byName := make(map[string]*cli.Command, len(command.Commands))
	for _, subcommand := range command.Commands {
		byName[subcommand.Name] = subcommand
	}
	for _, name := range []string{"show", "path", "init", "edit", "validate"} {
		if err := byName[name].Action(context.Background(), byName[name]); err != nil {
			t.Fatalf("%s action: %v", name, err)
		}
	}
}

// TestConfigCommandsRequireResolvedPath pins the guard that every config subcommand
// shares: an unresolved path must fail loudly instead of silently targeting "".
func TestConfigCommandsRequireResolvedPath(t *testing.T) {
	state := newTestState(&FakeRunner{})
	state.ConfigPath = ""

	tests := []struct {
		call func() error
		name string
	}{
		{name: "init", call: func() error { return RunConfigInit(state, false) }},
		{name: "edit", call: func() error { return RunConfigEdit(context.Background(), state) }},
		{name: "validate", call: func() error { return RunConfigValidate(state) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil || !strings.Contains(err.Error(), "could not resolve configuration path") {
				t.Fatalf("expected an unresolved path error, got %v", err)
			}
		})
	}
}

func TestRunConfigInitWriteFailures(t *testing.T) {
	t.Run("an unusable parent directory is reported", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		// os.DevNull is a file, so creating a directory beneath it must fail.
		state.ConfigPath = filepath.Join(os.DevNull, "sub", "dot.yaml")

		err := RunConfigInit(state, false)
		if err == nil || !strings.Contains(err.Error(), "failed to create config directory") {
			t.Fatalf("expected a directory creation error, got %v", err)
		}
	})

	t.Run("an unwritable target is reported", func(t *testing.T) {
		dir := t.TempDir()
		// A directory at the config path makes the write fail regardless of ownership.
		path := filepath.Join(dir, "dot.yaml")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		state := newTestState(&FakeRunner{})
		state.ConfigPath = path

		err := RunConfigInit(state, true)
		if err == nil || !strings.Contains(err.Error(), "failed to write config file") {
			t.Fatalf("expected a write error, got %v", err)
		}
	})
}

func TestRunConfigEditScaffoldsAndRequiresEditor(t *testing.T) {
	t.Run("a missing editor is reported", func(t *testing.T) {
		state := newTestState(&FakeRunner{
			LookPathFunc: func(string) (string, error) { return "", errors.New("not found") },
		})
		state.ConfigPath = filepath.Join(t.TempDir(), "dot.yaml")
		var stdout bytes.Buffer
		state.Stdout = &stdout
		t.Setenv("EDITOR", "definitely-not-an-editor")

		err := RunConfigEdit(context.Background(), state)
		if err == nil || !strings.Contains(err.Error(), `editor "definitely-not-an-editor" not found`) {
			t.Fatalf("expected a missing editor error, got %v", err)
		}
		// The file is scaffolded before the editor lookup, so it must exist by now.
		if _, statErr := os.Stat(state.ConfigPath); statErr != nil {
			t.Errorf("expected the config to be scaffolded: %v", statErr)
		}
	})

	t.Run("scaffolding failures abort the edit", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}
		// Readable but not writable: the config looks absent, yet cannot be created.
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		state := newTestState(&FakeRunner{})
		state.ConfigPath = filepath.Join(dir, "sub", "dot.yaml")

		err := RunConfigEdit(context.Background(), state)
		if err == nil || !strings.Contains(err.Error(), "failed to create config directory") {
			t.Fatalf("expected the scaffolding error to propagate, got %v", err)
		}
	})
}
