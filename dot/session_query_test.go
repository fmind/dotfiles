package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func ingestQueryFixture(t *testing.T, agent, sessionID, cwd, fingerprint string, completeness sessionCompleteness) string {
	t.Helper()
	logs := []SessionLogLine{{TS: "2026-07-31T10:00:00Z", Agent: agent, SID: sessionID, Role: "user", Content: "private prompt", CWD: cwd}}
	result, err := ingestSession(context.Background(), agent, sessionID, logs, sessionSource{Type: "fixture", Fingerprint: fingerprint, Completeness: completeness})
	if err != nil {
		t.Fatalf("ingestSession failed: %v", err)
	}
	root, err := sessionStoreRoot()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, agent, result.LineageID, result.GenerationID)
}

func rewriteQueryManifest(t *testing.T, generation, ingestedAt string, schemaVersion, malformed int) {
	t.Helper()
	manifest, err := readSessionManifest(generation)
	if err != nil {
		t.Fatal(err)
	}
	manifest.IngestedAt = ingestedAt
	manifest.SchemaVersion = schemaVersion
	manifest.MalformedRecords = malformed
	if malformed > 0 {
		manifest.Completeness = sessionPartial
	}
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(filepath.Join(generation, "manifest.json"), content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func duplicateQueryGeneration(t *testing.T, generation string) {
	t.Helper()
	duplicate := filepath.Join(filepath.Dir(generation), "duplicate-generation")
	if err := os.Mkdir(duplicate, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"manifest.json", "transcript.jsonl"} {
		content, err := os.ReadFile(filepath.Join(generation, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(duplicate, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSessionQueryFiltersAndVisibleLineageStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	oldGeneration := ingestQueryFixture(t, "codex", "session-1", "/work/project-a", strings.Repeat("a", 64), sessionComplete)
	newGeneration := ingestQueryFixture(t, "codex", "session-1", "/work/project-a", strings.Repeat("b", 64), sessionComplete)
	partialGeneration := ingestQueryFixture(t, "claude", "session-2", "/work/project-b", strings.Repeat("c", 64), sessionPartial)
	rewriteQueryManifest(t, oldGeneration, "2026-07-30T10:00:00Z", sessionSchemaVersion, 0)
	rewriteQueryManifest(t, newGeneration, "2026-07-31T10:00:00Z", sessionSchemaVersion, 0)
	rewriteQueryManifest(t, partialGeneration, "2026-08-01T10:00:00Z", sessionSchemaVersion, 2)
	duplicateQueryGeneration(t, oldGeneration)

	summaries, err := querySessionSummaries(SessionQuery{Agent: "codex", CWD: "/work/project-a", Since: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 3 {
		t.Fatalf("expected three codex generations, got %+v", summaries)
	}
	allStatuses := make([]string, 0)
	for _, summary := range summaries {
		allStatuses = append(allStatuses, summary.Status...)
	}
	statuses := strings.Join(allStatuses, ",")
	if !strings.Contains(statuses, "stale") || !strings.Contains(statuses, "duplicate") {
		t.Fatalf("expected duplicate and stale lineage status, got %+v", summaries)
	}
	partial, err := querySessionSummaries(SessionQuery{Identity: "session-2"}, false)
	if err != nil || len(partial) != 1 || !strings.Contains(strings.Join(partial[0].Status, ","), "partial") {
		t.Fatalf("partial ingestion was not visible: %+v, %v", partial, err)
	}
}

func TestSessionShowRequiresExplicitContent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ingestQueryFixture(t, "codex", "session-private", "/work/private", strings.Repeat("d", 64), sessionComplete)
	state := newTestState(&FakeRunner{})
	var output bytes.Buffer
	state.Stdout = &output
	query := SessionQuery{Identity: "session-private"}
	if err := RunAgentSessionShow(context.Background(), state, query, false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private prompt") || strings.Contains(output.String(), `"records":`) {
		t.Fatalf("metadata-only show leaked content: %s", output.String())
	}
	output.Reset()
	if err := RunAgentSessionShow(context.Background(), state, query, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "private prompt") {
		t.Fatalf("explicit content show omitted records: %s", output.String())
	}
}

func TestSessionExportSchemasAndRedactionAreStable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ingestQueryFixture(t, "claude", "session-export", "/work/export", strings.Repeat("e", 64), sessionComplete)
	state := newTestState(&FakeRunner{})
	var output bytes.Buffer
	state.Stdout = &output
	if err := RunAgentSessionExport(context.Background(), state, SessionQuery{}, "json", false, true); err != nil {
		t.Fatal(err)
	}
	var exported SessionExport
	if err := json.Unmarshal(output.Bytes(), &exported); err != nil {
		t.Fatal(err)
	}
	if exported.Schema != sessionExportSchema || len(exported.Sessions) != 1 || exported.Sessions[0].Records[0].Content != "[redacted]" {
		t.Fatalf("unexpected JSON export: %+v", exported)
	}
	output.Reset()
	if err := RunAgentSessionExport(context.Background(), state, SessionQuery{}, "ndjson", false, false); err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(strings.TrimSpace(output.String()), "\n") + 1; lines != 1 || !strings.Contains(output.String(), sessionExportSchema) {
		t.Fatalf("unexpected NDJSON export: %s", output.String())
	}
}

func TestSessionQuerySurfacesUnsupportedAndRejectsUnsafePermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	generation := ingestQueryFixture(t, "agy", "session-unsupported", "/work/a", strings.Repeat("f", 64), sessionComplete)
	rewriteQueryManifest(t, generation, "2026-08-01T10:00:00Z", sessionSchemaVersion+1, 0)
	summaries, err := querySessionSummaries(SessionQuery{}, false)
	if err != nil || len(summaries) != 1 || !strings.Contains(strings.Join(summaries[0].Status, ","), "unsupported") {
		t.Fatalf("unsupported ingestion was not visible: %+v, %v", summaries, err)
	}
	if err := os.Chmod(filepath.Join(generation, "transcript.jsonl"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := querySessionSummaries(SessionQuery{}, false); err == nil || !strings.Contains(err.Error(), "owner-only") {
		t.Fatalf("expected unsafe permission rejection, got %v", err)
	}
}
