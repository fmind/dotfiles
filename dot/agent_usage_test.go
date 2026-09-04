package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAndLoadUsageRecord(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	rec := UsageRecord{
		Timestamp:    "2026-09-03T12:00:00Z",
		Harness:      "claude",
		SessionID:    "session-123",
		Model:        "claude-opus-5",
		InputTokens:  1000,
		OutputTokens: 200,
		CachedTokens: 500,
		CostUSD:      0.05,
		TurnCount:    3,
		CWD:          "/home/test/project",
	}

	if err := WriteUsageRecord(rec); err != nil {
		t.Fatalf("WriteUsageRecord failed: %v", err)
	}

	expectedPath := filepath.Join(tempHome, ".agents", "usages", "claude", "session-123.json")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	var loaded UsageRecord
	if unmarshalErr := json.Unmarshal(data, &loaded); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal written file: %v", unmarshalErr)
	}

	if loaded.Harness != "claude" || loaded.Agent != "claude" {
		t.Errorf("expected harness/agent claude, got %s/%s", loaded.Harness, loaded.Agent)
	}
	if loaded.TotalTokens != 1700 {
		t.Errorf("expected TotalTokens 1700, got %d", loaded.TotalTokens)
	}

	all, err := LoadAllUsageRecords()
	if err != nil {
		t.Fatalf("LoadAllUsageRecords failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 record, got %d", len(all))
	}
}

func TestExtractUsageClaude(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "claude-session.jsonl")

	lines := []string{
		`{"type":"system","message":"init"}`,
		`{"type":"assistant","message":{"model":"claude-3-7-sonnet","usage":{"input_tokens":150,"output_tokens":45,"cache_read_input_tokens":300,"cache_creation_input_tokens":50}}}`,
		`{"type":"cost-state","totalCostUSD":0.0123}`,
		`{"type":"assistant","message":{"model":"claude-3-7-sonnet","usage":{"input_tokens":200,"output_tokens":60,"cache_read_input_tokens":400,"cache_creation_input_tokens":0}}}`,
	}
	if err := os.WriteFile(transcriptPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("failed to write test transcript: %v", err)
	}

	cfg := AgentConfig{}
	rec, err := ExtractUsageClaude(cfg, "test-session", "/test/dir", transcriptPath)
	if err != nil {
		t.Fatalf("ExtractUsageClaude failed: %v", err)
	}

	if rec.Harness != "claude" {
		t.Errorf("expected harness claude, got %s", rec.Harness)
	}
	if rec.InputTokens != 350 {
		t.Errorf("expected input tokens 350, got %d", rec.InputTokens)
	}
	if rec.OutputTokens != 105 {
		t.Errorf("expected output tokens 105, got %d", rec.OutputTokens)
	}
	if rec.CachedTokens != 700 {
		t.Errorf("expected cached tokens 700, got %d", rec.CachedTokens)
	}
	if rec.CacheWriteTokens != 50 {
		t.Errorf("expected cache write tokens 50, got %d", rec.CacheWriteTokens)
	}
	if rec.TotalTokens != 1205 {
		t.Errorf("expected total tokens 1205, got %d", rec.TotalTokens)
	}
	if rec.CostUSD != 0.0123 {
		t.Errorf("expected cost 0.0123, got %f", rec.CostUSD)
	}
	if rec.TurnCount != 2 {
		t.Errorf("expected 2 turns, got %d", rec.TurnCount)
	}
}

func TestExtractUsageCodex(t *testing.T) {
	tmpDir := t.TempDir()
	rolloutPath := filepath.Join(tmpDir, "rollout-session.jsonl")

	lines := []string{
		`{"type":"turn_context","payload":{"model":"gpt-4o","cwd":"/test/codex"}}`,
		`{"type":"response_item","payload":{"role":"assistant","content":[{"text":"hello"}]}}`,
		`{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":1200,"cached_input_tokens":800,"cache_write_input_tokens":100,"output_tokens":350,"reasoning_output_tokens":50,"total_tokens":2450}}}}`,
	}
	if err := os.WriteFile(rolloutPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("failed to write test rollout: %v", err)
	}

	cfg := AgentConfig{}
	rec, err := ExtractUsageCodex(cfg, "codex-123", "/test/codex", rolloutPath)
	if err != nil {
		t.Fatalf("ExtractUsageCodex failed: %v", err)
	}

	if rec.Harness != "codex" {
		t.Errorf("expected harness codex, got %s", rec.Harness)
	}
	if rec.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", rec.Model)
	}
	if rec.InputTokens != 1200 {
		t.Errorf("expected input tokens 1200, got %d", rec.InputTokens)
	}
	if rec.OutputTokens != 350 {
		t.Errorf("expected output tokens 350, got %d", rec.OutputTokens)
	}
	if rec.CachedTokens != 800 {
		t.Errorf("expected cached tokens 800, got %d", rec.CachedTokens)
	}
	if rec.ReasoningTokens != 50 {
		t.Errorf("expected reasoning tokens 50, got %d", rec.ReasoningTokens)
	}
	if rec.TotalTokens != 2450 {
		t.Errorf("expected total tokens 2450, got %d", rec.TotalTokens)
	}
}

func TestExtractUsageGrok(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	grokDir := filepath.Join(tempHome, ".grok", "sessions", "cwd_encoded", "session-grok-1")
	if err := os.MkdirAll(grokDir, 0o700); err != nil {
		t.Fatalf("failed to create grok dir: %v", err)
	}

	sigJSON := `{"contextTokensUsed": 45000, "contextWindowTokens": 128000, "primaryModelId": "grok-code", "turnCount": 6}`
	if err := os.WriteFile(filepath.Join(grokDir, "signals.json"), []byte(sigJSON), 0o600); err != nil {
		t.Fatalf("failed to write signals.json: %v", err)
	}

	state := &GlobalState{
		Config: DefaultConfig(),
	}
	state.Config.Agent.Sources = map[string]string{
		sessionStoreGrok: filepath.Join(tempHome, ".grok", "sessions"),
	}

	rec, err := ExtractUsageGrok(state, "session-grok-1", "/test/project")
	if err != nil {
		t.Fatalf("ExtractUsageGrok failed: %v", err)
	}

	if rec.Harness != "grok" {
		t.Errorf("expected harness grok, got %s", rec.Harness)
	}
	if rec.Model != "grok-code" {
		t.Errorf("expected model grok-code, got %s", rec.Model)
	}
	if rec.TotalTokens != 45000 {
		t.Errorf("expected total tokens 45000, got %d", rec.TotalTokens)
	}
	if rec.TurnCount != 6 {
		t.Errorf("expected turn count 6, got %d", rec.TurnCount)
	}
}

func TestExtractUsageAgy(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")

	lines := []string{
		`{"created_at":"2026-09-03T10:00:00Z","source":"USER_EXPLICIT","type":"USER_INPUT","content":"Please write a function"}`,
		`{"created_at":"2026-09-03T10:00:05Z","source":"MODEL","type":"PLANNER_RESPONSE","content":"Here is the function","thinking":"Thinking about the solution"}`,
	}
	if err := os.WriteFile(transcriptPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("failed to write agy transcript: %v", err)
	}

	state := &GlobalState{Config: DefaultConfig()}
	rec, err := ExtractUsageAgy(state, "agy-sess", "/test/agy", transcriptPath)
	if err != nil {
		t.Fatalf("ExtractUsageAgy failed: %v", err)
	}

	if rec.Harness != "agy" {
		t.Errorf("expected harness agy, got %s", rec.Harness)
	}
	if rec.TurnCount != 1 {
		t.Errorf("expected 1 turn, got %d", rec.TurnCount)
	}
	if rec.InputTokens <= 0 || rec.OutputTokens <= 0 {
		t.Errorf("expected positive token counts, got input=%d output=%d", rec.InputTokens, rec.OutputTokens)
	}
	if rec.TotalTokens != rec.InputTokens+rec.OutputTokens {
		t.Errorf("expected total tokens = input + output, got total=%d", rec.TotalTokens)
	}
}

func TestExtractUsageOpencode(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dbFile := filepath.Join(tempHome, "opencode.db")
	if err := os.WriteFile(dbFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	runner := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			return `[{"model":{"providerID":"anthropic","modelID":"claude-3-7-sonnet"},"tokens_input":500,"tokens_output":150,"tokens_reasoning":20,"tokens_cache_read":800,"tokens_cache_write":100,"cost":0.025,"directory":"/my/project","time_created":1700000000000}]`, nil
		},
	}

	state := &GlobalState{
		Config: DefaultConfig(),
		Runner: runner,
	}
	state.Config.Agent.Sources = map[string]string{
		sessionStoreOpenCode: dbFile,
	}

	rec, err := ExtractUsageOpencode(context.Background(), state, "sess-oc-1", "/my/project")
	if err != nil {
		t.Fatalf("ExtractUsageOpencode failed: %v", err)
	}

	if rec.Harness != "opencode" {
		t.Errorf("expected harness opencode, got %s", rec.Harness)
	}
	if rec.Model != "anthropic/claude-3-7-sonnet" {
		t.Errorf("expected anthropic/claude-3-7-sonnet, got %s", rec.Model)
	}
	if rec.InputTokens != 500 || rec.OutputTokens != 150 {
		t.Errorf("expected 500 in / 150 out, got %d / %d", rec.InputTokens, rec.OutputTokens)
	}
	if rec.CachedTokens != 800 || rec.CacheWriteTokens != 100 {
		t.Errorf("expected 800 cached / 100 write, got %d / %d", rec.CachedTokens, rec.CacheWriteTokens)
	}
	if rec.ReasoningTokens != 20 {
		t.Errorf("expected 20 reasoning, got %d", rec.ReasoningTokens)
	}
	if rec.CostUSD != 0.025 {
		t.Errorf("expected cost 0.025, got %f", rec.CostUSD)
	}
}

func TestExtractUsageCopilot(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dbFile := filepath.Join(tempHome, "copilot.db")
	if err := os.WriteFile(dbFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	runner := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			query := args[len(args)-1]
			if strings.Contains(query, "assistant_usage_events") {
				return `[{"model":"claude-3.5-sonnet","input_tokens":1200,"output_tokens":300,"cache_read_tokens":600,"cache_write_tokens":50,"reasoning_tokens":10,"turns":4}]`, nil
			}
			return `[{"cwd":"/copilot/project","created_at":"2026-09-03T11:00:00Z"}]`, nil
		},
	}

	state := &GlobalState{
		Config: DefaultConfig(),
		Runner: runner,
	}
	state.Config.Agent.Sources = map[string]string{
		sessionStoreCopilot: dbFile,
	}

	rec, err := ExtractUsageCopilot(context.Background(), state, "copilot-sess-1", "/copilot/project")
	if err != nil {
		t.Fatalf("ExtractUsageCopilot failed: %v", err)
	}

	if rec.Harness != "copilot" {
		t.Errorf("expected harness copilot, got %s", rec.Harness)
	}
	if rec.Model != "claude-3.5-sonnet" {
		t.Errorf("expected model claude-3.5-sonnet, got %s", rec.Model)
	}
	if rec.InputTokens != 1200 || rec.OutputTokens != 300 {
		t.Errorf("expected 1200 in / 300 out, got %d / %d", rec.InputTokens, rec.OutputTokens)
	}
	if rec.CachedTokens != 600 {
		t.Errorf("expected 600 cached, got %d", rec.CachedTokens)
	}
	if rec.TurnCount != 4 {
		t.Errorf("expected 4 turns, got %d", rec.TurnCount)
	}
}

func TestRunAgentHookUsageAgyStopDecision(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	transcriptDir := filepath.Join(tempHome, ".gemini", "antigravity-cli", "brain", "session-agy-stop", ".system_generated", "logs")
	if err := os.MkdirAll(transcriptDir, 0o700); err != nil {
		t.Fatal(err)
	}
	tPath := filepath.Join(transcriptDir, "transcript.jsonl")
	if err := os.WriteFile(tPath, []byte(`{"created_at":"2026-09-03T10:00:00Z","source":"USER_EXPLICIT","type":"USER_INPUT","content":"hi"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	state := &GlobalState{
		Config: DefaultConfig(),
		Stdin:  strings.NewReader(`{"conversationId":"session-agy-stop","transcriptPath":"` + tPath + `","workspacePaths":["/my/work"],"fullyIdle":true}`),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	state.Config.Agent.Sources = map[string]string{
		sessionStoreAgy: filepath.Join(tempHome, ".gemini", "antigravity-cli", "brain"),
	}

	err := RunAgentHookUsage(context.Background(), state, sessionStoreAgy, "", "")
	if err != nil {
		t.Fatalf("RunAgentHookUsage failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"decision":`) {
		t.Errorf("expected Antigravity stop decision in stdout, got %q", out)
	}

	usageFile := filepath.Join(tempHome, ".agents", "usages", "agy", "session-agy-stop.json")
	if _, statErr := os.Stat(usageFile); statErr != nil {
		t.Errorf("expected usage file at %s: %v", usageFile, statErr)
	}
}

func TestRunAgentHookUsageCopilotOutput(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dbFile := filepath.Join(tempHome, "copilot.db")
	if err := os.WriteFile(dbFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	runner := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			return `[{"model":"gpt-4o","input_tokens":100,"output_tokens":50,"cache_read_tokens":0,"cache_write_tokens":0,"reasoning_tokens":0,"turns":1}]`, nil
		},
	}

	var stdout, stderr bytes.Buffer
	state := &GlobalState{
		Config: DefaultConfig(),
		Runner: runner,
		Stdin:  strings.NewReader(copilotSessionEndPayload("copilot-hook-sess", "complete")),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	state.Config.Agent.Sources = map[string]string{
		sessionStoreCopilot: dbFile,
	}

	err := RunAgentHookUsage(context.Background(), state, sessionStoreCopilot, "", "")
	if err != nil {
		t.Fatalf("RunAgentHookUsage copilot failed: %v", err)
	}

	out := strings.TrimSpace(stdout.String())
	if out != "{}" {
		t.Errorf("expected {} in stdout for copilot hook, got %q", out)
	}

	usageFile := filepath.Join(tempHome, ".agents", "usages", "copilot", "copilot-hook-sess.json")
	if _, statErr := os.Stat(usageFile); statErr != nil {
		t.Errorf("expected usage file at %s: %v", usageFile, statErr)
	}
}

func TestRunAgentUsageStatsAndList(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	records := []UsageRecord{
		{
			Timestamp:    "2026-09-03T10:00:00Z",
			Harness:      "claude",
			SessionID:    "claude-1",
			Model:        "opus",
			InputTokens:  1000,
			OutputTokens: 200,
			CachedTokens: 500,
			TotalTokens:  1700,
			CostUSD:      0.05,
			TurnCount:    2,
		},
		{
			Timestamp:    "2026-09-03T11:00:00Z",
			Harness:      "codex",
			SessionID:    "codex-1",
			Model:        "gpt-4o",
			InputTokens:  2000,
			OutputTokens: 500,
			TotalTokens:  2500,
			CostUSD:      0.02,
			TurnCount:    3,
		},
	}
	for _, r := range records {
		if err := WriteUsageRecord(r); err != nil {
			t.Fatal(err)
		}
	}

	var stdout bytes.Buffer
	state := &GlobalState{
		Stdout: &stdout,
	}

	// 1. Text table
	err := RunAgentUsageStats(context.Background(), state, UsageStatsOptions{})
	if err != nil {
		t.Fatalf("RunAgentUsageStats failed: %v", err)
	}
	tableOutput := stdout.String()
	if !strings.Contains(tableOutput, "claude") || !strings.Contains(tableOutput, "codex") {
		t.Errorf("expected table to mention claude and codex, got:\n%s", tableOutput)
	}
	if !strings.Contains(tableOutput, "TOTAL") {
		t.Errorf("expected table to have TOTAL row, got:\n%s", tableOutput)
	}

	// 2. JSON output
	stdout.Reset()
	err = RunAgentUsageStats(context.Background(), state, UsageStatsOptions{JSON: true})
	if err != nil {
		t.Fatalf("RunAgentUsageStats JSON failed: %v", err)
	}
	var rows []UsageStatsRow
	if jsonErr := json.Unmarshal(stdout.Bytes(), &rows); jsonErr != nil {
		t.Fatalf("failed to decode json stats: %v", jsonErr)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 stats rows, got %d", len(rows))
	}

	// 3. List
	stdout.Reset()
	err = RunAgentUsageList(context.Background(), state, UsageListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("RunAgentUsageList failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "claude-1") {
		t.Errorf("expected list to contain claude-1, got:\n%s", stdout.String())
	}

	// 4. Show
	stdout.Reset()
	err = RunAgentUsageShow(context.Background(), state, "claude", "claude-1")
	if err != nil {
		t.Fatalf("RunAgentUsageShow failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"session_id": "claude-1"`) {
		t.Errorf("expected show to output claude-1 record, got:\n%s", stdout.String())
	}
}

// Stop and SessionEnd hooks overlap on one session, so a second writer must not
// lose its record. The previous implementation staged through a fixed
// "<target>.tmp" opened with O_EXCL, which made concurrent writers collide.
func TestWriteUsageRecordToleratesConcurrentWriters(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	const writers = 8
	errs := make(chan error, writers)
	start := make(chan struct{})
	for i := range writers {
		go func() {
			<-start
			errs <- WriteUsageRecord(UsageRecord{
				Timestamp:   "2026-09-03T12:00:00Z",
				Harness:     "claude",
				SessionID:   "overlapping-session",
				TotalTokens: int64(i),
			})
		}()
	}
	close(start)
	for range writers {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent WriteUsageRecord: %v", err)
		}
	}

	dir, err := HarnessUsageDir("claude")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 1 || names[0] != "overlapping-session.json" {
		t.Fatalf("usage directory = %v, want exactly [overlapping-session.json]; stray temporary files poison later writes", names)
	}
}

// A crash between staging and rename must not leave residue that fails every
// later write for the same session.
func TestWriteUsageRecordSurvivesStaleTemporaryResidue(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dir, err := HarnessUsageDir("codex")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "crashed-session.json")
	if err := os.WriteFile(target+".tmp", []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := WriteUsageRecord(UsageRecord{
		Timestamp: "2026-09-03T12:00:00Z",
		Harness:   "codex",
		SessionID: "crashed-session",
	}); err != nil {
		t.Fatalf("WriteUsageRecord after crash residue: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("record was not published: %v", err)
	}
}

// The Copilot session id reaches a SQL string literal, so ids outside the id
// alphabet must be rejected at the boundary rather than interpolated.
func TestExtractUsageCopilotRejectsUnsafeSessionID(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	state := &GlobalState{Config: DefaultConfig()}

	for _, sessionID := range []string{"", "quote'injection", "a' OR '1'='1", "with space", "../escape"} {
		if _, err := ExtractUsageCopilot(context.Background(), state, sessionID, "/project"); err == nil {
			t.Errorf("ExtractUsageCopilot(%q) = nil error, want rejection", sessionID)
		}
	}
}

// An unparseable window must fail the command rather than silently widening the
// report to all time.
func TestParseFlexibleTimeRejectsGarbage(t *testing.T) {
	for _, valid := range []string{"24h", "7d", "2026-09-03", "2026-09-03T11:00:00Z"} {
		if _, err := parseFlexibleTime(valid); err != nil {
			t.Errorf("parseFlexibleTime(%q) = %v, want success", valid, err)
		}
	}
	for _, invalid := range []string{"", "yesterday", "7days", "-3d", "2026-13-45"} {
		if _, err := parseFlexibleTime(invalid); err == nil {
			t.Errorf("parseFlexibleTime(%q) = nil error, want rejection", invalid)
		}
	}
}

func TestRunAgentUsageStatsRejectsInvalidWindow(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	state := &GlobalState{Config: DefaultConfig(), Stdout: io.Discard, Stderr: io.Discard}

	if err := RunAgentUsageStats(context.Background(), state, UsageStatsOptions{Since: "yesterday"}); err == nil {
		t.Error("RunAgentUsageStats with --since yesterday = nil error, want rejection")
	}
	if err := RunAgentUsageStats(context.Background(), state, UsageStatsOptions{Until: "nonsense"}); err == nil {
		t.Error("RunAgentUsageStats with --until nonsense = nil error, want rejection")
	}
}

// writeAgyTranscript materializes one agy transcript with `turns` model replies, so
// a larger file yields a larger byte-count estimate.
func writeAgyTranscript(t *testing.T, root, sessionID, name string, turns int) {
	t.Helper()
	dir := filepath.Join(root, sessionID, ".system_generated", "logs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	lines := make([]string, 0, turns*2)
	for range turns {
		lines = append(lines,
			`{"created_at":"2026-09-03T10:00:00Z","source":"USER_EXPLICIT","type":"USER_INPUT","content":"Please write a function"}`,
			`{"created_at":"2026-09-03T10:00:05Z","source":"MODEL","type":"PLANNER_RESPONSE","content":"Here is the function","thinking":"Thinking about the solution"}`,
		)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func loadUsageRecord(t *testing.T, harness, sessionID string) UsageRecord {
	t.Helper()
	dir, err := HarnessUsageDir(harness)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, sessionID+".json"))
	if err != nil {
		t.Fatalf("usage record for %s/%s not written: %v", harness, sessionID, err)
	}
	var rec UsageRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	return rec
}

// The backfill must read the same transcript the hook does. Hard-coding
// transcript.jsonl both skipped sessions carrying only transcript_full.jsonl and
// overwrote accurate hook-written records with a short-file undercount.
func TestRunAgentUsageSyncPrefersFullAgyTranscript(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	agyRoot := filepath.Join(tempHome, "agy-sessions")

	writeAgyTranscript(t, agyRoot, "both-transcripts", "transcript.jsonl", 1)
	writeAgyTranscript(t, agyRoot, "both-transcripts", "transcript_full.jsonl", 6)
	writeAgyTranscript(t, agyRoot, "full-only", "transcript_full.jsonl", 3)
	// A session directory agy created before any transcript exists is skipped.
	if err := os.MkdirAll(filepath.Join(agyRoot, "no-transcript"), 0o700); err != nil {
		t.Fatal(err)
	}

	state := &GlobalState{Config: DefaultConfig(), Stdout: io.Discard, Stderr: io.Discard}
	state.Config.Agent.Sources = map[string]string{sessionStoreAgy: agyRoot}

	if err := RunAgentUsageSync(context.Background(), state); err != nil {
		t.Fatalf("RunAgentUsageSync: %v", err)
	}

	// full-only must be backfilled at all; the hard-coded path skipped it entirely.
	fullOnly := loadUsageRecord(t, sessionStoreAgy, "full-only")
	if fullOnly.TotalTokens <= 0 {
		t.Errorf("full-only total_tokens = %d, want > 0", fullOnly.TotalTokens)
	}

	// With both present the longer transcript must win.
	both := loadUsageRecord(t, sessionStoreAgy, "both-transcripts")
	if both.TotalTokens <= fullOnly.TotalTokens {
		t.Errorf("both-transcripts total_tokens = %d, want more than full-only's %d; the short transcript was read",
			both.TotalTokens, fullOnly.TotalTokens)
	}

	dir, err := HarnessUsageDir(sessionStoreAgy)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("wrote %d records, want 2 (the transcript-less directory must be skipped)", len(entries))
	}
}

// memory.jsonl is hand-curated long-term memory whose stem parses as a valid
// session id; every other sweep over this tree already rejects it by name.
func TestRunAgentUsageSyncSkipsClaudeMemoryFile(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	claudeRoot := filepath.Join(tempHome, "claude-projects", "proj")
	if err := os.MkdirAll(claudeRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	line := `{"type":"assistant","message":{"model":"claude-opus-5","usage":{"input_tokens":10,"output_tokens":5}}}` + "\n"
	for _, name := range []string{"memory.jsonl", "real-session.jsonl"} {
		if err := os.WriteFile(filepath.Join(claudeRoot, name), []byte(line), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	state := &GlobalState{Config: DefaultConfig(), Stdout: io.Discard, Stderr: io.Discard}
	state.Config.Agent.Sources = map[string]string{sessionStoreClaude: filepath.Join(tempHome, "claude-projects")}

	if err := RunAgentUsageSync(context.Background(), state); err != nil {
		t.Fatalf("RunAgentUsageSync: %v", err)
	}

	dir, err := HarnessUsageDir(sessionStoreClaude)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 1 || names[0] != "real-session.json" {
		t.Errorf("usage records = %v, want exactly [real-session.json]; memory.jsonl is not a session", names)
	}
}

// An unreadable store must fail the command rather than report "Synced 0".
func TestRunAgentUsageSyncReportsUnreadableStore(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	notADirectory := filepath.Join(tempHome, "claude-store")
	if err := os.WriteFile(notADirectory, []byte("not a directory\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	state := &GlobalState{Config: DefaultConfig(), Stdout: io.Discard, Stderr: io.Discard}
	state.Config.Agent.Sources = map[string]string{sessionStoreClaude: notADirectory}

	if err := RunAgentUsageSync(context.Background(), state); err == nil {
		t.Error("RunAgentUsageSync over an unreadable store = nil error, want failure")
	}
}
