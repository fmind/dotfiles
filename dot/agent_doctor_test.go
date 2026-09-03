package dot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeDoctorFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeDoctorLink(t *testing.T, path, target string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}

func setupHealthyAgentDoctor(t *testing.T) (*GlobalState, *strings.Builder, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	persona := filepath.Join(home, ".agents", "AGENTS.md")
	skills := filepath.Join(home, ".agents", "skills")
	writeDoctorFile(t, persona, "# Persona\n")
	if err := os.MkdirAll(skills, 0o700); err != nil {
		t.Fatal(err)
	}
	writeDoctorLink(t, filepath.Join(home, ".gemini", "GEMINI.md"), persona)
	writeDoctorLink(t, filepath.Join(home, ".gemini", "config", "skills"), skills)
	writeDoctorLink(t, filepath.Join(home, ".claude", "CLAUDE.md"), persona)
	writeDoctorLink(t, filepath.Join(home, ".claude", "skills"), skills)
	writeDoctorLink(t, filepath.Join(home, ".codex", "AGENTS.md"), persona)
	writeDoctorLink(t, filepath.Join(home, ".grok", "AGENTS.md"), persona)
	writeDoctorLink(t, filepath.Join(home, ".grok", "skills"), skills)
	writeDoctorLink(t, filepath.Join(home, ".copilot", "copilot-instructions.md"), persona)

	writeDoctorFile(t, filepath.Join(home, ".gemini", "config", "hooks.json"), `{"hooks":["dot agent hook session agy","dot agent hook notify agy stop"]}`)
	writeDoctorFile(t, filepath.Join(home, ".claude", "settings.json"), `{"hooks":["dot agent hook session claude","dot agent hook notify claude stop"]}`)
	writeDoctorFile(t, filepath.Join(home, ".codex", "config.toml"), "command = \"dot agent hook session codex\"\ncommand = \"dot agent hook notify codex stop\"\n")
	writeDoctorFile(t, filepath.Join(home, ".grok", "hooks", "hooks.json"), `{"hooks":["dot agent hook session grok","dot agent hook notify grok stop"]}`)
	writeDoctorFile(t, filepath.Join(home, ".config", "opencode", "opencode.json"), `{"instructions":["~/.agents/AGENTS.md"]}`)
	writeDoctorFile(t, filepath.Join(home, ".config", "opencode", "plugins", "session-log.ts"), "dot agent hook session opencode")
	writeDoctorFile(t, filepath.Join(home, ".copilot", "hooks", "session-log.json"), `{"version":1,"hooks":{"sessionEnd":[{"bash":"dot agent hook copilot-session-end"}]}}`)
	for _, path := range []string{
		filepath.Join(home, ".gemini", "antigravity-cli", "brain", ".keep"),
		filepath.Join(home, ".claude", "projects", ".keep"),
		filepath.Join(home, ".codex", "sessions", ".keep"),
		filepath.Join(home, ".grok", "sessions", ".keep"),
		filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
		filepath.Join(home, ".copilot", "session-store.db"),
	} {
		writeDoctorFile(t, path, "")
	}

	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) { return "/bin/" + name, nil },
		RunFunc:      func(context.Context, string, io.Reader, string, ...string) (string, error) { return "ok", nil },
	}
	output := &strings.Builder{}
	state := newTestState(runner)
	state.Stdout = output
	return state, output, home
}

func doctorResultFor(t *testing.T, results []agentDoctorResult, agent string) agentDoctorResult {
	t.Helper()
	for _, result := range results {
		if result.Agent == agent {
			return result
		}
	}
	t.Fatalf("missing doctor result for %s", agent)
	return agentDoctorResult{}
}

func TestAgentDoctorHealthyReadOnlyReport(t *testing.T) {
	state, output, _ := setupHealthyAgentDoctor(t)
	if err := RunAgentDoctor(context.Background(), state, AgentDoctorOptions{}); err != nil {
		t.Fatalf("RunAgentDoctor returned an error: %v\n%s", err, output.String())
	}
	report := output.String()
	for _, agent := range []string{sessionStoreAgy, sessionStoreClaude, sessionStoreCodex, sessionStoreGrok, sessionStoreOpenCode, sessionStoreCopilot} {
		if !strings.Contains(report, passIcon+" "+agent+":") {
			t.Fatalf("missing healthy %s report: %s", agent, report)
		}
	}
	if strings.Contains(report, "Persona") || strings.Contains(report, "transcript") {
		t.Fatalf("doctor exposed file content: %s", report)
	}
}

func TestAgentDoctorSanitizedFailureFixtures(t *testing.T) {
	t.Run("missing discovery link", func(t *testing.T) {
		state, _, home := setupHealthyAgentDoctor(t)
		if err := os.Remove(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
			t.Fatal(err)
		}
		result := doctorResultFor(t, gatherAgentDoctor(context.Background(), state, home, time.Now()), sessionStoreClaude)
		if result.Healthy || result.Discovery != "persona-broken" {
			t.Fatalf("unexpected missing-link result: %+v", result)
		}
	})

	t.Run("malformed hooks", func(t *testing.T) {
		state, _, home := setupHealthyAgentDoctor(t)
		writeDoctorFile(t, filepath.Join(home, ".claude", "settings.json"), "{")
		result := doctorResultFor(t, gatherAgentDoctor(context.Background(), state, home, time.Now()), sessionStoreClaude)
		if result.Healthy || result.Hooks != "malformed" {
			t.Fatalf("unexpected malformed-hook result: %+v", result)
		}
	})

	t.Run("unavailable tool", func(t *testing.T) {
		state, _, home := setupHealthyAgentDoctor(t)
		state.Runner = &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "codex" {
					return "", errors.New("missing")
				}
				return "/bin/" + name, nil
			},
			RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) { return "ok", nil },
		}
		result := doctorResultFor(t, gatherAgentDoctor(context.Background(), state, home, time.Now()), sessionStoreCodex)
		if result.Healthy || !strings.Contains(result.Tools, "codex:missing") {
			t.Fatalf("unexpected missing-tool result: %+v", result)
		}
	})

	t.Run("partial lineage", func(t *testing.T) {
		state, _, home := setupHealthyAgentDoctor(t)
		logs := []SessionLogLine{{Agent: sessionStoreCodex, SID: "partial-lineage", Role: "user", Content: "sanitized"}}
		if _, err := ingestSession(context.Background(), sessionStoreCodex, "partial-lineage", logs, sessionSource{Completeness: sessionPartial}); err != nil {
			t.Fatal(err)
		}
		result := doctorResultFor(t, gatherAgentDoctor(context.Background(), state, home, time.Now()), sessionStoreCodex)
		if result.Healthy || result.LastIngestion != "partial-only" {
			t.Fatalf("unexpected partial-lineage result: %+v", result)
		}
	})

	t.Run("stale ingestion and bounded failure metadata", func(t *testing.T) {
		state, _, home := setupHealthyAgentDoctor(t)
		logs := []SessionLogLine{{Agent: sessionStoreClaude, SID: "stale-lineage", Role: "user", Content: "must-not-appear"}}
		result, err := ingestSession(context.Background(), sessionStoreClaude, "stale-lineage", logs, sessionSource{})
		if err != nil {
			t.Fatal(err)
		}
		generation := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, sessionStoreClaude, result.LineageID, result.GenerationID)
		manifest, err := readSessionManifest(generation)
		if err != nil {
			t.Fatal(err)
		}
		manifest.IngestedAt = time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339Nano)
		content, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(generation, "manifest.json"), content, 0o600); err != nil {
			t.Fatal(err)
		}
		writeDoctorFile(t, filepath.Join(home, ".claude", "projects", "project", "stale-lineage.jsonl"), "newer raw source")
		_ = spoolHookFailure(defaultAgentConfig(), sessionStoreClaude, "session", "stale-lineage", errors.New("sanitized failure"))
		doctor := doctorResultFor(t, gatherAgentDoctor(context.Background(), state, home, time.Now()), sessionStoreClaude)
		if doctor.Healthy || doctor.ArchiveLag == "0s" || doctor.LastFailure == "none" {
			t.Fatalf("unexpected stale result: %+v", doctor)
		}
		if strings.Contains(doctor.LastFailure, "sanitized failure") || strings.Contains(doctor.LastFailure, "must-not-appear") {
			t.Fatalf("doctor exposed failure or transcript content: %+v", doctor)
		}
	})
}

func TestAgentDoctorRepairIsExplicitAndPreviewable(t *testing.T) {
	state, _, _ := setupHealthyAgentDoctor(t)
	runner := newRecordedRunner(nil)
	state.Runner = runner
	if err := repairAgentIntegrations(context.Background(), state, true); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 1 || !strings.HasPrefix(runner.calls[0], "chezmoi apply --dry-run --force ") {
		t.Fatalf("unexpected repair preview: %v", runner.calls)
	}
	if err := RunAgentDoctor(context.Background(), state, AgentDoctorOptions{DryRun: true}); err == nil || !strings.Contains(err.Error(), "requires --fix") {
		t.Fatalf("expected explicit --fix requirement, got %v", err)
	}
}

func TestInspectLastHookFailureSkipsEmptySpoolFiles(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion)
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	// A zero-byte record is an interrupted spool write, not corruption: the
	// doctor must keep reading older records instead of reporting "unreadable".
	if err := os.WriteFile(filepath.Join(root, "20260902T000000000000000Z-empty.json"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	record := `{"occurred_at":"2026-09-01T00:00:00Z","agent":"claude","operation":"session","detail":"boom"}` + "\n"
	if err := os.WriteFile(filepath.Join(root, "20260901T000000000000000Z-record.json"), []byte(record), 0o600); err != nil {
		t.Fatal(err)
	}
	status, ok := inspectLastHookFailure(AgentConfig{}, home, sessionStoreClaude)
	if !ok || status != "2026-09-01T00:00:00Z:session" {
		t.Fatalf("inspectLastHookFailure = %q, %v; want the older record and ok", status, ok)
	}
	status, ok = inspectLastHookFailure(AgentConfig{}, home, sessionStoreCodex)
	if !ok || status != "none" {
		t.Fatalf("inspectLastHookFailure for another agent = %q, %v; want none, ok", status, ok)
	}
}
