package dot

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	t.Run("config group stays reachable despite a malformed config", func(t *testing.T) {
		app := NewApp()
		// `config path` only prints the resolved path; if the Before hook wrongly treated
		// the malformed config as fatal, this would return the parse error instead of nil.
		err := app.Run(ctx, []string{"dot", "--config", malformed, "config", "path"})
		if err != nil {
			t.Errorf("expected the config group to remain usable to repair a bad file, got %v", err)
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
