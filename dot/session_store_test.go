package dot

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentHookAndSyncIngestionConvergesForEveryAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	agents := []string{"agy", "claude", "codex", "opencode", "copilot"}
	type outcome struct {
		err    error
		agent  string
		status sessionIngestionStatus
	}
	outcomes := make(chan outcome, len(agents)*2)
	var wait sync.WaitGroup
	for _, agent := range agents {
		for range 2 {
			wait.Add(1)
			go func(agent string) {
				defer wait.Done()
				sessionID := "shared-session"
				logs := []SessionLogLine{{TS: "2026-08-01T12:00:00Z", Agent: agent, SID: sessionID, Role: "user", Content: "private"}}
				result, err := ingestSession(context.Background(), agent, sessionID, logs, sessionSource{Type: agent + "-test", Fingerprint: strings.Repeat("a", 64)})
				outcomes <- outcome{agent: agent, status: result.Status, err: err}
			}(agent)
		}
	}
	wait.Wait()
	close(outcomes)

	statuses := make(map[string]map[sessionIngestionStatus]int)
	for item := range outcomes {
		if item.err != nil {
			t.Fatalf("%s concurrent ingestion failed: %v", item.agent, item.err)
		}
		if statuses[item.agent] == nil {
			statuses[item.agent] = make(map[sessionIngestionStatus]int)
		}
		statuses[item.agent][item.status]++
	}
	for _, agent := range agents {
		if statuses[agent][sessionIngested] != 1 || statuses[agent][sessionDuplicate] != 1 {
			t.Fatalf("%s race did not converge: %+v", agent, statuses[agent])
		}
		lineageDir := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, agent, sessionLineageID(agent, "shared-session"))
		entries, err := os.ReadDir(lineageDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 || strings.HasPrefix(entries[0].Name(), ".ingest-") {
			t.Fatalf("%s left a partial or duplicate generation: %+v", agent, entries)
		}
	}
}

func TestSessionGenerationsPreserveChangedSourcesAndManifest(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	state := newTestState(&FakeRunner{})
	state.Stderr = nil
	logs := []SessionLogLine{{TS: "2026-08-01T12:00:00Z", Agent: "codex", SID: "generation-session", Role: "user", Content: "first"}}

	first, err := writeSessionLogs(context.Background(), state, "codex", "generation-session", logs, sessionSource{Type: "codex-jsonl", Fingerprint: strings.Repeat("a", 64), Malformed: 1})
	if err != nil {
		t.Fatal(err)
	}
	logs[0].Content = "second"
	second, err := writeSessionLogs(context.Background(), state, "codex", "generation-session", logs, sessionSource{Type: "codex-jsonl", Fingerprint: strings.Repeat("b", 64)})
	if err != nil {
		t.Fatal(err)
	}
	if first.GenerationID == second.GenerationID || first.LineageID != second.LineageID {
		t.Fatalf("changed source did not preserve one lineage with distinct generations: first=%+v second=%+v", first, second)
	}
	if first.Manifest.SchemaVersion != sessionSchemaVersion || first.Manifest.ParserVersion != sessionParserVersion || first.Manifest.Completeness != sessionPartial || first.Manifest.SourceFingerprint != strings.Repeat("a", 64) {
		t.Fatalf("manifest omitted lineage metadata: %+v", first.Manifest)
	}
	lineageDir := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, "codex", first.LineageID)
	entries, err := os.ReadDir(lineageDir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("expected two preserved generations, got %+v (%v)", entries, err)
	}
	for _, result := range []sessionIngestionResult{first, second} {
		path := filepath.Join(lineageDir, result.GenerationID)
		if err := validateSessionGeneration(path, result.Manifest); err != nil {
			t.Fatalf("generation %s is invalid: %v", result.GenerationID, err)
		}
	}
}

func TestSessionIngestionRejectsCorruptExistingGeneration(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	logs := []SessionLogLine{{Agent: "claude", SID: "corrupt-session", Role: "user", Content: "original"}}
	result, err := ingestSession(context.Background(), "claude", "corrupt-session", logs, sessionSource{Type: "claude-jsonl", Fingerprint: strings.Repeat("c", 64)})
	if err != nil {
		t.Fatal(err)
	}
	transcript := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, "claude", result.LineageID, result.GenerationID, "transcript.jsonl")
	if writeErr := os.WriteFile(transcript, []byte("corrupt\n"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	_, err = ingestSession(context.Background(), "claude", "corrupt-session", logs, sessionSource{Type: "claude-jsonl", Fingerprint: strings.Repeat("c", 64)})
	if err == nil || !strings.Contains(err.Error(), "existing session generation is invalid") {
		t.Fatalf("corrupt generation was misreported as a duplicate: %v", err)
	}
}

func TestSessionMigrationDryRunAndApplySelectMostComplete(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	legacyDir := filepath.Join(home, ".agents", "sessions", "2026-07-31")
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	sessionID := "migration-session"
	shortPath := filepath.Join(legacyDir, "090000_codex_"+sessionID+".jsonl")
	longPath := filepath.Join(legacyDir, "100000_codex_"+sessionID+".jsonl")
	shortLogs := []SessionLogLine{{Agent: "codex", SID: sessionID, Role: "user", Content: "one"}}
	longLogs := []SessionLogLine{{Agent: "codex", SID: sessionID, Role: "user", Content: "one"}, {Agent: "codex", SID: sessionID, Role: "assistant", Content: "two"}}
	writeLegacyLogs(t, shortPath, shortLogs)
	writeLegacyLogs(t, longPath, longLogs)
	if err := os.WriteFile(filepath.Join(legacyDir, "unrecognized.jsonl"), []byte("evidence"), 0o600); err != nil {
		t.Fatal(err)
	}

	state := newTestState(&FakeRunner{})
	var output strings.Builder
	state.Stdout = &output
	if err := RunAgentSessionMigrate(context.Background(), state, false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), sessionID) {
		t.Fatalf("dry-run exposed a raw session identifier: %s", output.String())
	}
	if !strings.Contains(output.String(), "dry-run selected=1 duplicate=1 partial=0 skipped=0 malformed_files=1 legacy_preserved=true") {
		t.Fatalf("dry-run outcomes were incomplete: %s", output.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "sessions", sessionStoreVersion)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run wrote the versioned store: %v", err)
	}

	output.Reset()
	if err := RunAgentSessionMigrate(context.Background(), state, true); err != nil {
		t.Fatal(err)
	}
	logs := readAgentSessionLogs(t, home, "codex")
	if len(logs[sessionID]) != 2 {
		t.Fatalf("migration did not select the most complete transcript: %+v", logs)
	}
	for _, path := range []string{shortPath, longPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("migration removed legacy evidence %s: %v", path, err)
		}
	}
}

func writeLegacyLogs(t *testing.T, path string, logs []SessionLogLine) {
	t.Helper()
	var content strings.Builder
	encoder := json.NewEncoder(&content)
	for _, log := range logs {
		if err := encoder.Encode(log); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, []byte(content.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestStoredGenerationMatchesIngestedIdentity pins the fast path `dot agent session
// sync` relies on: the generation identity is fully determined by the source
// fingerprint, so an already-ingested source must be recognizable without re-parsing
// its transcript, and anything that is not a byte-identical match must miss.
func TestStoredGenerationMatchesIngestedIdentity(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	const agent = "claude"
	const sessionID = "session-fast-path"
	fingerprint := strings.Repeat("b", 64)
	logs := []SessionLogLine{{TS: "2026-08-01T12:00:00Z", Agent: agent, SID: sessionID, Role: "user", Content: "hello"}}

	if _, ok := storedGeneration(agent, sessionID, fingerprint); ok {
		t.Fatal("reported a stored generation before anything was ingested")
	}

	result, err := ingestSession(context.Background(), agent, sessionID, logs, sessionSource{Type: "claude-jsonl", Fingerprint: fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != sessionIngested {
		t.Fatalf("status = %q, want %q", result.Status, sessionIngested)
	}

	manifest, ok := storedGeneration(agent, sessionID, fingerprint)
	if !ok {
		t.Fatal("ingested generation was not recognized by the fast path")
	}
	if manifest.RecordCount != len(logs) || manifest.LineageID != result.LineageID {
		t.Errorf("manifest = %+v, want record_count=%d lineage=%s", manifest, len(logs), result.LineageID)
	}

	// Each component of the identity must be load-bearing: a different source, session,
	// or agent is a different generation and has to fall through to the full ingest.
	for name, probe := range map[string][3]string{
		"different fingerprint": {agent, sessionID, strings.Repeat("c", 64)},
		"different session":     {agent, "other-session", fingerprint},
		"different agent":       {"codex", sessionID, fingerprint},
		"empty fingerprint":     {agent, sessionID, ""},
	} {
		if _, ok := storedGeneration(probe[0], probe[1], probe[2]); ok {
			t.Errorf("%s: fast path claimed a stored generation", name)
		}
	}
}

// TestStoredGenerationRejectsTamperedManifest ensures the cheap check still refuses a
// generation whose manifest no longer matches its own path, so corruption falls through
// to the full ingest rather than being silently accepted as already done.
func TestStoredGenerationRejectsTamperedManifest(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	const agent = "codex"
	const sessionID = "session-tampered"
	fingerprint := strings.Repeat("d", 64)
	logs := []SessionLogLine{{TS: "2026-08-01T12:00:00Z", Agent: agent, SID: sessionID, Role: "user", Content: "hello"}}
	if _, err := ingestSession(context.Background(), agent, sessionID, logs, sessionSource{Type: "codex-jsonl", Fingerprint: fingerprint}); err != nil {
		t.Fatal(err)
	}

	root, err := sessionStoreRoot()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, agent, sessionLineageID(agent, sessionID), sessionGenerationID(fingerprint))
	manifest, err := readSessionManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	manifest.SourceFingerprint = strings.Repeat("e", 64)
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), append(content, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, ok := storedGeneration(agent, sessionID, fingerprint); ok {
		t.Error("fast path accepted a generation whose manifest contradicts its path")
	}
}
