package dot

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewAgentCmd(t *testing.T) {
	state := newTestState(&FakeRunner{})
	cmd := NewAgentCmd(state)

	if cmd.Name != "agent" {
		t.Errorf("expected command name 'agent', got %q", cmd.Name)
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "a" {
		t.Errorf("expected command alias 'a', got %v", cmd.Aliases)
	}

	if len(cmd.Commands) != 1 || cmd.Commands[0].Name != "session" {
		t.Fatalf("expected sub-command 'session'")
	}
	sessionCmd := cmd.Commands[0]
	if len(sessionCmd.Aliases) != 1 || sessionCmd.Aliases[0] != "s" {
		t.Errorf("expected command alias 's', got %v", sessionCmd.Aliases)
	}

	subCommands := map[string]struct {
		alias string
		found bool
	}{
		"agy":      {alias: "", found: false},
		"claude":   {alias: "", found: false},
		"codex":    {alias: "", found: false},
		"opencode": {alias: "", found: false},
		"copilot":  {alias: "", found: false},
		"sync":     {alias: "s", found: false},
		"clean":    {alias: "c", found: false},
	}

	for _, sub := range sessionCmd.Commands {
		if entry, ok := subCommands[sub.Name]; ok {
			entry.found = true
			if entry.alias != "" {
				if len(sub.Aliases) != 1 || sub.Aliases[0] != entry.alias {
					t.Errorf("expected subcommand %q alias %q, got %v", sub.Name, entry.alias, sub.Aliases)
				}
			} else {
				if len(sub.Aliases) != 0 {
					t.Errorf("expected subcommand %q to have no alias, got %v", sub.Name, sub.Aliases)
				}
			}
			subCommands[sub.Name] = entry
		}
	}

	for name, entry := range subCommands {
		if !entry.found {
			t.Errorf("expected subcommand %q under 'session' to be registered", name)
		}
	}
}

func TestRunAgentSessionLogAgy(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	sessionID := "test-session-123"
	cwd := "/workspace/test"

	// Create a mock transcript.jsonl
	transcriptDir := filepath.Join(tempDir, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs")
	if err := os.MkdirAll(transcriptDir, 0o755); err != nil {
		t.Fatalf("failed to create transcript dir: %v", err)
	}

	transcriptLines := []string{
		`{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-07-09T14:09:12Z","content":"hello agy"}`,
		`{"step_index":1,"source":"SYSTEM","type":"CONVERSATION_HISTORY","status":"DONE","created_at":"2026-07-09T14:09:12Z"}`,
		`{"step_index":2,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-07-09T14:09:12Z","content":"hello user"}`,
		`{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-07-09T14:09:12Z","content":"truncated turn","is_truncated":true}`,
	}
	transcriptContent := strings.Join(transcriptLines, "\n")
	transcriptPath := filepath.Join(transcriptDir, "transcript.jsonl")
	if err := os.WriteFile(transcriptPath, []byte(transcriptContent), 0o644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}

	state := newTestState(&FakeRunner{})
	var stdout strings.Builder
	state.Stdout = &stdout
	ctx := context.Background()

	err := RunAgentSessionLogAgy(ctx, state, sessionID, cwd)
	if err != nil {
		t.Fatalf("RunAgentSessionLogAgy failed: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("manual invocation unexpectedly wrote hook output %q", stdout.String())
	}

	// Verify that log was written to ~/.agents/sessions/YYYY-MM-DD/HHMMSS_agy_sessionID.jsonl
	sessionsDir := filepath.Join(tempDir, ".agents", "sessions")
	var logFiles []string
	err = filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			logFiles = append(logFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk sessions dir: %v", err)
	}

	if len(logFiles) != 1 {
		t.Fatalf("expected exactly 1 log file, found %d: %v", len(logFiles), logFiles)
	}

	logContent, err := os.ReadFile(logFiles[0])
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(logContent)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 log lines, got %d", len(lines))
	}

	var firstLine, secondLine SessionLogLine
	if err := json.Unmarshal([]byte(lines[0]), &firstLine); err != nil {
		t.Fatalf("failed to unmarshal first line: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &secondLine); err != nil {
		t.Fatalf("failed to unmarshal second line: %v", err)
	}

	if firstLine.Role != "user" || firstLine.Content != "hello agy" || firstLine.CWD != cwd {
		t.Errorf("unexpected first line: %+v", firstLine)
	}
	if secondLine.Role != "assistant" || secondLine.Content != "hello user" || secondLine.CWD != cwd {
		t.Errorf("unexpected second line: %+v", secondLine)
	}
}

func TestRunAgentSessionLogAgyAntigravityHookPayload(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sessionID := "conversation-123"
	cwd := filepath.Join(home, "workspace")
	transcriptPath := filepath.Join(home, ".antigravitycli", "transcripts", sessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(transcriptPath), 0o700); err != nil {
		t.Fatalf("create transcript directory: %v", err)
	}
	transcript := strings.Join([]string{
		`{"source":"USER_EXPLICIT","type":"USER_INPUT","created_at":"2026-07-09T14:09:12Z","content":"hook prompt"}`,
		`{"source":"MODEL","type":"PLANNER_RESPONSE","created_at":"2026-07-09T14:09:13Z","content":"hook response"}`,
	}, "\n")
	if err := os.WriteFile(transcriptPath, []byte(transcript), 0o600); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	payload, err := json.Marshal(map[string]any{
		"conversationId": sessionID,
		"fullyIdle":      true,
		"hookEventName":  "Stop",
		"transcriptPath": transcriptPath,
		"workspacePaths": []string{cwd},
	})
	if err != nil {
		t.Fatalf("marshal hook payload: %v", err)
	}
	state := newTestState(&FakeRunner{})
	state.Stdin = strings.NewReader(string(payload))
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunAgentSessionLogAgy(context.Background(), state, "", ""); err != nil {
		t.Fatalf("RunAgentSessionLogAgy from hook payload failed: %v", err)
	}
	if stdout.String() != "{\"decision\":\"\"}\n" {
		t.Fatalf("unexpected Antigravity Stop response %q", stdout.String())
	}

	logs := readAgentSessionLogs(t, home, "agy")[sessionID]
	if len(logs) != 2 {
		t.Fatalf("expected 2 normalized Antigravity messages, got %d: %+v", len(logs), logs)
	}
	if logs[0].Content != "hook prompt" || logs[1].Content != "hook response" {
		t.Errorf("unexpected normalized Antigravity content: %+v", logs)
	}
	for _, log := range logs {
		if log.CWD != cwd {
			t.Errorf("expected workspace path %q, got %+v", cwd, log)
		}
	}
}

func TestRunAgentSessionLogAgySkipsNonIdleHook(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	payload := `{"conversationId":"active-session","fullyIdle":false,"transcriptPath":"/missing/transcript.jsonl"}`
	state := newTestState(&FakeRunner{})
	state.Stdin = strings.NewReader(payload)
	var stdout strings.Builder
	state.Stdout = &stdout

	if err := RunAgentSessionLogAgy(context.Background(), state, "", ""); err != nil {
		t.Fatalf("non-idle Antigravity hook failed: %v", err)
	}
	if stdout.String() != "{\"decision\":\"\"}\n" {
		t.Fatalf("unexpected Antigravity Stop response %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "sessions")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("non-idle hook persisted an incomplete transcript: %v", err)
	}
}

func TestHookInputPreservesSnakeCasePayload(t *testing.T) {
	var input HookInput
	payload := []byte(`{"session_id":"session-123","cwd":"/workspace","transcript_path":"/tmp/transcript.jsonl","stop_hook_active":true,"fullyIdle":true}`)
	if err := json.Unmarshal(payload, &input); err != nil {
		t.Fatalf("unmarshal snake_case hook payload: %v", err)
	}
	if input.SessionID != "session-123" || input.CWD != "/workspace" || input.TranscriptPath != "/tmp/transcript.jsonl" || !input.StopHookActive || !input.FullyIdle {
		t.Errorf("snake_case hook payload changed during normalization: %+v", input)
	}
}

func TestRunAgentSessionLogClaude(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	sessionID := "test-session-claude"
	cwd := "/workspace/test"

	// Create a mock claude session jsonl file
	projectsDir := filepath.Join(tempDir, ".claude", "projects", "-workspace-test")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("failed to create projects dir: %v", err)
	}

	claudeLines := []string{
		`{"type":"user","timestamp":"2026-07-09T14:09:12Z","message":{"content":"hello claude"},"cwd":"/workspace/test"}`,
		`{"type":"assistant","timestamp":"2026-07-09T14:09:14Z","message":{"content":[{"type":"text","text":"hello back"}],"model":"claude-3-5-sonnet"},"cwd":"/workspace/test"}`,
	}
	claudeContent := strings.Join(claudeLines, "\n")
	claudePath := filepath.Join(projectsDir, sessionID+".jsonl")
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0o644); err != nil {
		t.Fatalf("failed to write claude session file: %v", err)
	}

	state := newTestState(&FakeRunner{})
	ctx := context.Background()

	err := RunAgentSessionLogClaude(ctx, state, sessionID, cwd)
	if err != nil {
		t.Fatalf("RunAgentSessionLogClaude failed: %v", err)
	}

	sessionsDir := filepath.Join(tempDir, ".agents", "sessions")
	var logFiles []string
	_ = filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			logFiles = append(logFiles, path)
		}
		return nil
	})

	if len(logFiles) != 1 {
		t.Fatalf("expected exactly 1 log file, found %d", len(logFiles))
	}

	logContent, _ := os.ReadFile(logFiles[0])
	lines := strings.Split(strings.TrimSpace(string(logContent)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 log lines, got %d", len(lines))
	}

	var firstLine, secondLine SessionLogLine
	_ = json.Unmarshal([]byte(lines[0]), &firstLine)
	_ = json.Unmarshal([]byte(lines[1]), &secondLine)

	if firstLine.Role != "user" || firstLine.Content != "hello claude" || firstLine.CWD != cwd || firstLine.Model != "claude-3-5-sonnet" {
		t.Errorf("unexpected first line: %+v", firstLine)
	}
	if secondLine.Role != "assistant" || secondLine.Content != "hello back" || secondLine.CWD != cwd || secondLine.Model != "claude-3-5-sonnet" {
		t.Errorf("unexpected second line: %+v", secondLine)
	}
}

func TestRunAgentSessionLogClaudeUsesHookTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	transcriptPath := filepath.Join(home, "hook-transcripts", "claude.jsonl")
	if err := os.MkdirAll(filepath.Dir(transcriptPath), 0o700); err != nil {
		t.Fatal(err)
	}
	transcript := `{"type":"user","timestamp":"2026-07-09T14:09:12Z","message":{"content":"hook transcript"},"cwd":"/workspace/hook"}`
	if err := os.WriteFile(transcriptPath, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"session_id":      "claude-hook-session",
		"cwd":             "/wrong/cwd",
		"transcript_path": transcriptPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	state := newTestState(&FakeRunner{})
	state.Stdin = strings.NewReader(string(payload))

	if err := RunAgentSessionLogClaude(context.Background(), state, "", ""); err != nil {
		t.Fatalf("RunAgentSessionLogClaude failed: %v", err)
	}
	logs := readAgentSessionLogs(t, home, "claude")["claude-hook-session"]
	if len(logs) != 1 || logs[0].Content != "hook transcript" || logs[0].CWD != "/workspace/hook" {
		t.Fatalf("unexpected Claude hook transcript logs: %+v", logs)
	}
}

const opencodeFixtureSQL = `
CREATE TABLE session (
    id TEXT PRIMARY KEY,
    directory TEXT NOT NULL
);
CREATE TABLE message (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    time_created INTEGER NOT NULL,
    data TEXT NOT NULL
);
CREATE TABLE part (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    time_created INTEGER NOT NULL,
    data TEXT NOT NULL
);

INSERT INTO session VALUES ('ses_direct', '/workspace/direct');
INSERT INTO session VALUES ('ses_sync', '/workspace/sync');

INSERT INTO message VALUES ('msg_direct_user', 'ses_direct', 1780220445000, '{"role":"user","model":{"providerID":"google","modelID":"gemini-3.5-flash"}}');
INSERT INTO message VALUES ('msg_direct_assistant', 'ses_direct', 1780220446000, '{"role":"assistant"}');
INSERT INTO part VALUES ('part_direct_user', 'msg_direct_user', 'ses_direct', 1780220445001, '{"type":"text","text":"hello direct"}');
INSERT INTO part VALUES ('part_direct_reasoning', 'msg_direct_assistant', 'ses_direct', 1780220446001, '{"type":"reasoning","text":"internal reasoning"}');
INSERT INTO part VALUES ('part_direct_text_1', 'msg_direct_assistant', 'ses_direct', 1780220446002, '{"type":"text","text":"first response"}');
INSERT INTO part VALUES ('part_direct_text_2', 'msg_direct_assistant', 'ses_direct', 1780220446003, '{"type":"text","text":"second response"}');

INSERT INTO message VALUES ('msg_sync_user', 'ses_sync', 1780220455000, '{"role":"user","model":{"providerID":"anthropic","modelID":"claude-sonnet"}}');
INSERT INTO message VALUES ('msg_sync_assistant', 'ses_sync', 1780220456000, '{"role":"assistant"}');
INSERT INTO part VALUES ('part_sync_user', 'msg_sync_user', 'ses_sync', 1780220455001, '{"type":"text","text":"hello sync"}');
INSERT INTO part VALUES ('part_sync_tool', 'msg_sync_assistant', 'ses_sync', 1780220456001, '{"type":"tool","text":"ignored tool payload"}');
INSERT INTO part VALUES ('part_sync_text', 'msg_sync_assistant', 'ses_sync', 1780220456002, '{"type":"text","text":"synced response"}');
`

func setupOpenCodeFixture(t *testing.T, home string) {
	t.Helper()

	sqlitePath, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 is required for the OpenCode integration test")
	}

	dbDir := filepath.Join(home, ".local", "share", "opencode")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		t.Fatalf("create OpenCode data directory: %v", err)
	}
	dbPath := filepath.Join(dbDir, "opencode.db")
	cmd := exec.Command(sqlitePath, "-batch", "-bail", "-init", os.DevNull, dbPath) //nolint:gosec // fixed test binary and temp path
	cmd.Stdin = strings.NewReader(opencodeFixtureSQL)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create OpenCode fixture: %v\n%s", err, output)
	}
}

func readAgentSessionLogs(t *testing.T, home, agent string) map[string][]SessionLogLine {
	t.Helper()

	logs := make(map[string][]SessionLogLine)
	sessionsDir := filepath.Join(home, ".agents", "sessions")
	if err := filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.Contains(entry.Name(), "_"+agent+"_") || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
			var log SessionLogLine
			if err := json.Unmarshal([]byte(line), &log); err != nil {
				return err
			}
			logs[log.SID] = append(logs[log.SID], log)
		}
		return nil
	}); err != nil {
		t.Fatalf("read normalized session logs: %v", err)
	}
	return logs
}

func TestRunAgentSessionLogOpencodeCurrentSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupOpenCodeFixture(t, home)

	state := newTestState(NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard))
	if err := RunAgentSessionLogOpencode(context.Background(), state, "ses_direct", ""); err != nil {
		t.Fatalf("RunAgentSessionLogOpencode failed: %v", err)
	}

	logs := readAgentSessionLogs(t, home, "opencode")["ses_direct"]
	if len(logs) != 2 {
		t.Fatalf("expected 2 normalized OpenCode messages, got %d: %+v", len(logs), logs)
	}
	if logs[0].Role != "user" || logs[0].Content != "hello direct" || logs[0].CWD != "/workspace/direct" {
		t.Errorf("unexpected normalized user message: %+v", logs[0])
	}
	if logs[1].Role != "assistant" || logs[1].Content != "first response\nsecond response" {
		t.Errorf("unexpected normalized assistant message: %+v", logs[1])
	}
	for _, log := range logs {
		if log.Model != "google/gemini-3.5-flash" {
			t.Errorf("expected propagated model, got %+v", log)
		}
	}
}

func TestRunAgentSessionSyncOpencodeCurrentSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupOpenCodeFixture(t, home)

	state := newTestState(NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard))
	if err := RunAgentSessionSync(context.Background(), state); err != nil {
		t.Fatalf("RunAgentSessionSync failed: %v", err)
	}

	logs := readAgentSessionLogs(t, home, "opencode")
	if len(logs) != 2 {
		t.Fatalf("expected 2 synced OpenCode sessions, got %d: %+v", len(logs), logs)
	}
	syncLogs := logs["ses_sync"]
	if len(syncLogs) != 2 {
		t.Fatalf("expected 2 messages for ses_sync, got %d: %+v", len(syncLogs), syncLogs)
	}
	if syncLogs[0].Content != "hello sync" || syncLogs[1].Content != "synced response" {
		t.Errorf("unexpected synced OpenCode content: %+v", syncLogs)
	}
}

const copilotDirectSessionID = "11111111-1111-4111-8111-111111111111"

const copilotSyncSessionID = "22222222-2222-4222-8222-222222222222"

const copilotFixtureSQL = `
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    cwd TEXT,
    branch TEXT
);
CREATE TABLE turns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    turn_index INTEGER NOT NULL,
    user_message TEXT,
    assistant_response TEXT,
    timestamp TEXT
);

INSERT INTO sessions VALUES ('11111111-1111-4111-8111-111111111111', '/workspace/direct', 'main');
INSERT INTO sessions VALUES ('22222222-2222-4222-8222-222222222222', '/workspace/sync', 'main');

INSERT INTO turns (session_id, turn_index, user_message, assistant_response, timestamp) VALUES ('11111111-1111-4111-8111-111111111111', 0, 'hello direct', 'first direct response', '2026-04-23T19:00:59.100Z');
INSERT INTO turns (session_id, turn_index, user_message, assistant_response, timestamp) VALUES ('11111111-1111-4111-8111-111111111111', 1, 'follow up direct', NULL, '2026-04-23T19:01:10.200Z');
INSERT INTO turns (session_id, turn_index, user_message, assistant_response, timestamp) VALUES ('22222222-2222-4222-8222-222222222222', 0, 'hello sync', 'synced response', '2026-04-23T19:05:00.000Z');
`

func setupCopilotFixture(t *testing.T, home string) {
	t.Helper()

	sqlitePath, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 is required for the Copilot integration test")
	}

	dbDir := filepath.Join(home, ".copilot")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		t.Fatalf("create Copilot data directory: %v", err)
	}
	dbPath := filepath.Join(dbDir, "session-store.db")
	cmd := exec.Command(sqlitePath, "-batch", "-bail", "-init", os.DevNull, dbPath) //nolint:gosec // fixed test binary and temp path
	cmd.Stdin = strings.NewReader(copilotFixtureSQL)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create Copilot fixture: %v\n%s", err, output)
	}
}

func TestRunAgentSessionLogCopilotCurrentSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupCopilotFixture(t, home)

	state := newTestState(NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard))
	if err := RunAgentSessionLogCopilot(context.Background(), state, copilotDirectSessionID, ""); err != nil {
		t.Fatalf("RunAgentSessionLogCopilot failed: %v", err)
	}

	logs := readAgentSessionLogs(t, home, "copilot")[copilotDirectSessionID]
	if len(logs) != 3 {
		t.Fatalf("expected 3 normalized Copilot lines, got %d: %+v", len(logs), logs)
	}
	if logs[0].Role != "user" || logs[0].Content != "hello direct" || logs[0].CWD != "/workspace/direct" {
		t.Errorf("unexpected first Copilot line: %+v", logs[0])
	}
	if logs[1].Role != "assistant" || logs[1].Content != "first direct response" {
		t.Errorf("unexpected second Copilot line: %+v", logs[1])
	}
	// The second turn has a NULL assistant_response, so only its prompt is logged.
	if logs[2].Role != "user" || logs[2].Content != "follow up direct" {
		t.Errorf("unexpected third Copilot line: %+v", logs[2])
	}
}

func TestRunAgentSessionSyncCopilotCurrentSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupCopilotFixture(t, home)

	state := newTestState(NewStandardRunner(strings.NewReader(""), io.Discard, io.Discard))
	if err := RunAgentSessionSync(context.Background(), state); err != nil {
		t.Fatalf("RunAgentSessionSync failed: %v", err)
	}

	logs := readAgentSessionLogs(t, home, "copilot")
	if len(logs) != 2 {
		t.Fatalf("expected 2 synced Copilot sessions, got %d: %+v", len(logs), logs)
	}
	syncLogs := logs[copilotSyncSessionID]
	if len(syncLogs) != 2 {
		t.Fatalf("expected 2 lines for the synced Copilot session, got %d: %+v", len(syncLogs), syncLogs)
	}
	if syncLogs[0].Content != "hello sync" || syncLogs[1].Content != "synced response" {
		t.Errorf("unexpected synced Copilot content: %+v", syncLogs)
	}
}

type opencodeSyncStub struct {
	messageSQL     string
	messageQueries int
}

func newOpencodeSyncState(
	t *testing.T,
	home, sessionOutput string,
	sessionErr error,
	messageOutput string,
	messageErr error,
) (*GlobalState, *strings.Builder, *opencodeSyncStub) {
	t.Helper()

	dbDir := filepath.Join(home, ".local", "share", "opencode")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		t.Fatalf("create OpenCode directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "opencode.db"), nil, 0o600); err != nil {
		t.Fatalf("create OpenCode database placeholder: %v", err)
	}

	stub := &opencodeSyncStub{}
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name != "sqlite3" || len(args) == 0 {
				return "", fmt.Errorf("unexpected command: %s %s", name, strings.Join(args, " "))
			}
			query := args[len(args)-1]
			if query == opencodeSessionsQuery {
				return sessionOutput, sessionErr
			}
			if strings.Contains(query, "FROM message m") {
				stub.messageQueries++
				stub.messageSQL = query
				return messageOutput, messageErr
			}
			return "", fmt.Errorf("unexpected SQLite query: %s", query)
		},
	}
	state := newTestState(runner)
	stderr := &strings.Builder{}
	state.Stderr = stderr
	return state, stderr, stub
}

func marshalOpencodeRows(t *testing.T, rows []OpencodeRow) string {
	t.Helper()
	data, err := json.Marshal(rows)
	if err != nil {
		t.Fatalf("marshal OpenCode rows: %v", err)
	}
	return string(data)
}

func TestRunAgentSessionSyncOpencodeFailures(t *testing.T) {
	validSessions := `[{"id":"ses-valid_1","directory":"/workspace"}]`
	validRows := marshalOpencodeRows(t, []OpencodeRow{{
		SessionID:   "ses-valid_1",
		MessageID:   "msg-1",
		PartID:      "part-1",
		Data:        `{"role":"user"}`,
		PartData:    `{"type":"text","text":"hello"}`,
		Directory:   "/workspace",
		TimeCreated: 1780220445000,
	}})

	tests := []struct {
		name          string
		sessionOutput string
		sessionErr    error
		messageOutput string
		messageErr    error
		wantError     string
		breakWriter   bool
	}{
		{
			name:       "session query",
			sessionErr: errors.New("session query failed"),
			wantError:  "failed to query OpenCode sessions",
		},
		{
			name:          "session decode",
			sessionOutput: `{`,
			wantError:     "failed to decode OpenCode session query result",
		},
		{
			name:          "invalid session row",
			sessionOutput: `[{"id":"bad!id","directory":"/workspace"}]`,
			wantError:     "invalid session ID",
		},
		{
			name:          "message query",
			sessionOutput: validSessions,
			messageErr:    errors.New("message query failed"),
			wantError:     "failed to query OpenCode messages",
		},
		{
			name:          "message decode",
			sessionOutput: validSessions,
			messageOutput: `{`,
			wantError:     "failed to decode OpenCode message query result",
		},
		{
			name:          "message parse",
			sessionOutput: validSessions,
			messageOutput: marshalOpencodeRows(t, []OpencodeRow{{
				SessionID: "ses-valid_1",
				MessageID: "msg-1",
				Data:      `{`,
			}}),
			wantError: "failed to parse OpenCode session",
		},
		{
			name:          "session write",
			sessionOutput: validSessions,
			messageOutput: validRows,
			wantError:     "failed to write OpenCode session",
			breakWriter:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			state, _, _ := newOpencodeSyncState(t, home, test.sessionOutput, test.sessionErr, test.messageOutput, test.messageErr)
			if test.breakWriter {
				if err := os.MkdirAll(filepath.Join(home, ".agents"), 0o700); err != nil {
					t.Fatalf("create agents directory: %v", err)
				}
				if err := os.WriteFile(filepath.Join(home, ".agents", "sessions"), []byte("not a directory"), 0o600); err != nil {
					t.Fatalf("create broken sessions path: %v", err)
				}
			}

			err := RunAgentSessionSync(context.Background(), state)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected error containing %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestRunAgentSessionSyncOpencodeEmpty(t *testing.T) {
	tests := []struct {
		name           string
		sessionOutput  string
		messageOutput  string
		messageQueries int
	}{
		{name: "no sessions", sessionOutput: `[]`},
		{
			name:           "session without messages",
			sessionOutput:  `[{"id":"ses-valid_1","directory":"/workspace"}]`,
			messageOutput:  `[]`,
			messageQueries: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			state, stderr, stub := newOpencodeSyncState(t, home, test.sessionOutput, nil, test.messageOutput, nil)
			if err := RunAgentSessionSync(context.Background(), state); err != nil {
				t.Fatalf("RunAgentSessionSync failed: %v", err)
			}
			if !strings.Contains(stderr.String(), "opencode: 0 new\n") {
				t.Fatalf("missing zero-session summary in %q", stderr.String())
			}
			if stub.messageQueries != test.messageQueries {
				t.Fatalf("expected %d message queries, got %d", test.messageQueries, stub.messageQueries)
			}
			if test.messageQueries > 0 && !strings.Contains(stub.messageSQL, "m.session_id IN ('ses-valid_1')") {
				t.Fatalf("session ID was not preserved in query: %s", stub.messageSQL)
			}
		})
	}
}

func TestRunAgentSessionSyncReportsProcessedTreeErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".agents"), []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("create broken agents path: %v", err)
	}

	err := RunAgentSessionSync(context.Background(), newTestState(&FakeRunner{}))
	if err == nil || !strings.Contains(err.Error(), "failed to scan processed session logs") {
		t.Fatalf("expected processed-session traversal error, got %v", err)
	}
}

func TestRunAgentSessionSyncReportsSourceStatErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".gemini"), []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("create broken Antigravity path: %v", err)
	}

	err := RunAgentSessionSync(context.Background(), newTestState(&FakeRunner{}))
	if err == nil || !strings.Contains(err.Error(), "failed to inspect agy session directory") {
		t.Fatalf("expected source stat error, got %v", err)
	}
}

func TestWriteSessionLogsRepairsPrivatePermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dateDir := filepath.Join(home, ".agents", "sessions", "2026-07-09")
	if err := os.MkdirAll(dateDir, 0o755); err != nil {
		t.Fatalf("create permissive session directory: %v", err)
	}
	sessionsDir := filepath.Dir(dateDir)
	for _, path := range []string{sessionsDir, dateDir} {
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatalf("set permissive directory mode: %v", err)
		}
	}

	outPath := filepath.Join(dateDir, "140912_codex_existing-session.jsonl")
	if err := os.WriteFile(outPath, []byte("old content\n"), 0o644); err != nil {
		t.Fatalf("create permissive session log: %v", err)
	}
	if err := os.Chmod(outPath, 0o644); err != nil {
		t.Fatalf("set permissive file mode: %v", err)
	}

	logs := []SessionLogLine{{
		TS:      time.Date(2026, time.July, 9, 14, 9, 12, 0, time.UTC).Format(time.RFC3339),
		Agent:   "codex",
		SID:     "existing-session",
		Role:    "user",
		Content: "private prompt",
	}}
	if err := writeSessionLogs("codex", "existing-session", logs); err != nil {
		t.Fatalf("writeSessionLogs failed: %v", err)
	}

	for path, want := range map[string]os.FileMode{
		sessionsDir: 0o700,
		dateDir:     0o700,
		outPath:     0o600,
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("mode for %s: got %04o, want %04o", path, got, want)
		}
	}
}

type failingSessionWriter struct {
	err error
}

func (w failingSessionWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

type failingSessionCloser struct {
	err error
}

func (c failingSessionCloser) Close() error {
	return c.err
}

func TestSessionLogFinalizationErrors(t *testing.T) {
	flushFailure := errors.New("disk full")
	writer := bufio.NewWriter(failingSessionWriter{err: flushFailure})
	if _, err := writer.WriteString("buffered log"); err != nil {
		t.Fatalf("buffer log: %v", err)
	}
	if err := flushSessionLog(writer, "session.jsonl"); !errors.Is(err, flushFailure) {
		t.Fatalf("expected flush failure, got %v", err)
	}

	closeFailure := errors.New("close failed")
	if err := closeSessionLog(failingSessionCloser{err: closeFailure}, "session.jsonl"); !errors.Is(err, closeFailure) {
		t.Fatalf("expected close failure, got %v", err)
	}
}

func TestSessionTranscriptMalformedJSONSkips(t *testing.T) {
	tests := []struct {
		setup func(t *testing.T, home, sessionID string)
		run   func(ctx context.Context, state *GlobalState, sessionID string) error
		name  string
		agent string
	}{
		{
			name:  "agy",
			agent: "agy",
			setup: func(t *testing.T, home, sessionID string) {
				t.Helper()
				path := filepath.Join(home, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript.jsonl")
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(`{"source":"USER_EXPLICIT","type":"USER_INPUT","content":"valid"}`+"\n{"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			run: func(ctx context.Context, state *GlobalState, sessionID string) error {
				return RunAgentSessionLogAgy(ctx, state, sessionID, "")
			},
		},
		{
			name:  "claude",
			agent: "claude",
			setup: func(t *testing.T, home, sessionID string) {
				t.Helper()
				path := filepath.Join(home, ".claude", "projects", "project", sessionID+".jsonl")
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(`{"type":"user","message":{"content":"valid"}}`+"\n{"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			run: func(ctx context.Context, state *GlobalState, sessionID string) error {
				return RunAgentSessionLogClaude(ctx, state, sessionID, "")
			},
		},
		{
			name:  "codex",
			agent: "codex",
			setup: func(t *testing.T, home, sessionID string) {
				t.Helper()
				path := filepath.Join(home, ".codex", "sessions", "rollout-2026-07-09T14-09-12-"+sessionID+".jsonl")
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(`{"type":"user_message","payload":{"message":"valid"}}`+"\n{"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			run: func(ctx context.Context, state *GlobalState, sessionID string) error {
				return RunAgentSessionLogCodex(ctx, state, sessionID, "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			sessionID := "malformed-session"
			tt.setup(t, home, sessionID)

			err := tt.run(context.Background(), newTestState(&FakeRunner{}), sessionID)
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}

			sessionsDir := filepath.Join(home, ".agents", "sessions")
			var files []string
			err = filepath.Walk(sessionsDir, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if !info.IsDir() && strings.HasSuffix(path, ".jsonl") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("failed to walk sessions dir: %v", err)
			}
			if len(files) != 1 {
				t.Fatalf("expected exactly one parsed session log file, got %d files in %s", len(files), sessionsDir)
			}
			content, err := os.ReadFile(files[0])
			if err != nil {
				t.Fatalf("failed to read parsed session log file: %v", err)
			}
			if !strings.Contains(string(content), `"Content":"valid"`) && !strings.Contains(string(content), `"content":"valid"`) {
				t.Fatalf("expected output log to contain the valid entry, got: %s", string(content))
			}
		})
	}
}

func TestRunAgentSessionClean(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	sessionsDir := filepath.Join(tempDir, ".agents", "sessions")

	// Create directories representing old and new dates relative to the test run.
	now := time.Now().UTC()
	oldDateDir := filepath.Join(sessionsDir, now.AddDate(0, 0, -60).Format("2006-01-02"))
	newDateDir := filepath.Join(sessionsDir, now.Format("2006-01-02"))
	if err := os.MkdirAll(oldDateDir, 0o755); err != nil {
		t.Fatalf("failed to create old dir: %v", err)
	}
	if err := os.MkdirAll(newDateDir, 0o755); err != nil {
		t.Fatalf("failed to create new dir: %v", err)
	}

	// Write mock session log files
	oldFile := filepath.Join(oldDateDir, "120000_agy_old-session.jsonl")
	newFile := filepath.Join(newDateDir, "120000_agy_new-session.jsonl")

	if err := os.WriteFile(oldFile, []byte("old content"), 0o644); err != nil {
		t.Fatalf("failed to write old file: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new content"), 0o644); err != nil {
		t.Fatalf("failed to write new file: %v", err)
	}

	state := newTestState(&FakeRunner{})
	ctx := context.Background()

	// Clean logs older than 30 days. Cutoff from 2026-07-09 is 2026-06-09.
	err := RunAgentSessionClean(ctx, state, 30)
	if err != nil {
		t.Fatalf("RunAgentSessionClean failed: %v", err)
	}

	// Verify old file is deleted and its directory is removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("expected old file to be deleted, but it exists")
	}
	if _, err := os.Stat(oldDateDir); !os.IsNotExist(err) {
		t.Errorf("expected empty old date directory to be removed, but it exists")
	}

	// Verify new file still exists along with its directory
	if _, err := os.Stat(newFile); err != nil {
		t.Errorf("expected new file to persist, got error: %v", err)
	}
	if _, err := os.Stat(newDateDir); err != nil {
		t.Errorf("expected new date directory to persist, got error: %v", err)
	}
}

func TestRunAgentSessionCleanRejectsInvalidRetention(t *testing.T) {
	for _, days := range []int{0, -1} {
		if err := RunAgentSessionClean(context.Background(), newTestState(&FakeRunner{}), days); err == nil {
			t.Errorf("expected retention days %d to be rejected", days)
		}
	}
}

func TestRunAgentSessionCleanSurfacesRemovalFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	oldDir := filepath.Join(home, ".agents", "sessions", time.Now().UTC().AddDate(0, 0, -60).Format("2006-01-02"))
	if err := os.MkdirAll(oldDir, 0o700); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(oldDir, "120000_agy_old-session.jsonl")
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(oldDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(oldDir, 0o700) })

	err := RunAgentSessionClean(context.Background(), newTestState(&FakeRunner{}), 30)
	if err == nil || !strings.Contains(err.Error(), "failed to remove expired session log") {
		t.Fatalf("expected removal failure, got %v", err)
	}
}

func TestRunAgentSessionLogCodex(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	sessionID := "test-session-codex-123"
	cwd := "/workspace/test"

	// Create a mock codex session jsonl file
	sessionsDir := filepath.Join(tempDir, ".codex", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	codexLines := []string{
		`{"timestamp":"2026-07-09T14:09:12Z","type":"user_message","payload":{"message":"hello codex"},"cwd":"/workspace/test"}`,
		`{"timestamp":"2026-07-09T14:09:14Z","type":"agent_message","payload":{"content":"hello back","model":"gpt-4o"},"cwd":"/workspace/test"}`,
	}
	codexContent := strings.Join(codexLines, "\n")
	codexPath := filepath.Join(sessionsDir, "rollout-2026-07-09T14-09-12-"+sessionID+".jsonl")
	if err := os.WriteFile(codexPath, []byte(codexContent), 0o644); err != nil {
		t.Fatalf("failed to write codex session file: %v", err)
	}

	state := newTestState(&FakeRunner{})
	ctx := context.Background()

	err := RunAgentSessionLogCodex(ctx, state, sessionID, cwd)
	if err != nil {
		t.Fatalf("RunAgentSessionLogCodex failed: %v", err)
	}

	outputDir := filepath.Join(tempDir, ".agents", "sessions")
	var logFiles []string
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			logFiles = append(logFiles, path)
		}
		return nil
	})

	if len(logFiles) != 1 {
		t.Fatalf("expected exactly 1 log file, found %d", len(logFiles))
	}

	logContent, _ := os.ReadFile(logFiles[0])
	lines := strings.Split(strings.TrimSpace(string(logContent)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 log lines, got %d", len(lines))
	}

	var firstLine, secondLine SessionLogLine
	_ = json.Unmarshal([]byte(lines[0]), &firstLine)
	_ = json.Unmarshal([]byte(lines[1]), &secondLine)

	if firstLine.Role != "user" || firstLine.Content != "hello codex" || firstLine.CWD != cwd || firstLine.Model != "gpt-4o" {
		t.Errorf("unexpected first line: %+v", firstLine)
	}
	if secondLine.Role != "assistant" || secondLine.Content != "hello back" || secondLine.CWD != cwd || secondLine.Model != "gpt-4o" {
		t.Errorf("unexpected second line: %+v", secondLine)
	}
}

func TestRunAgentSessionLogCodexHookPathIsAuthoritative(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sessionID := "codex-hook-session"
	fallbackDir := filepath.Join(home, ".codex", "sessions")
	if err := os.MkdirAll(fallbackDir, 0o700); err != nil {
		t.Fatal(err)
	}
	fallbackPath := filepath.Join(fallbackDir, "rollout-2026-07-09T14-09-12-"+sessionID+".jsonl")
	if err := os.WriteFile(fallbackPath, []byte(`{"type":"user_message","payload":{"message":"wrong fallback"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"session_id":      sessionID,
		"transcript_path": filepath.Join(home, "missing.jsonl"),
	})
	if err != nil {
		t.Fatal(err)
	}
	state := newTestState(&FakeRunner{})
	state.Stdin = strings.NewReader(string(payload))

	err = RunAgentSessionLogCodex(context.Background(), state, "", "")
	if err == nil || !strings.Contains(err.Error(), "codex transcript from hook payload is unavailable") {
		t.Fatalf("expected authoritative hook-path error, got %v", err)
	}
}

func TestRunAgentSessionLogCodexRequiresExactSessionFilename(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sessionsDir := filepath.Join(home, ".codex", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sessionsDir, "rollout-2026-07-09T14-09-12-target-session-extra.jsonl")
	if err := os.WriteFile(path, []byte(`{"type":"user_message","payload":{"message":"wrong session"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := RunAgentSessionLogCodex(context.Background(), newTestState(&FakeRunner{}), "target-session", "")
	if err == nil || !strings.Contains(err.Error(), "session file not found for codex session target-session") {
		t.Fatalf("expected exact session match failure, got %v", err)
	}
}

func TestFindSessionFilePropagatesTraversalErrors(t *testing.T) {
	rootParent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(rootParent, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := findSessionFile(filepath.Join(rootParent, "sessions"), func(_ string, _ os.DirEntry) bool { return false })
	if err == nil {
		t.Fatal("expected session traversal error")
	}
}

func TestRunAgentSessionLogCodexResponseItemFormat(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	sessionID := "019f483e-63e2-7b31-92df-eb28b32a37ec"
	cwd := "/workspace/current"

	sessionsDir := filepath.Join(tempDir, ".codex", "sessions", "2026", "07", "09")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	codexLines := []string{
		`{"timestamp":"2026-07-09T14:09:12Z","type":"session_meta","payload":{"session_id":"019f483e-63e2-7b31-92df-eb28b32a37ec","cwd":"/workspace/current"}}`,
		`{"timestamp":"2026-07-09T14:09:13Z","type":"response_item","payload":{"type":"message","role":"developer","content":[{"type":"input_text","text":"internal instruction"}]}}`,
		`{"timestamp":"2026-07-09T14:09:14Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"hello current codex"}]}}`,
		`{"timestamp":"2026-07-09T14:09:15Z","type":"turn_context","payload":{"cwd":"/workspace/current","model":"gpt-5.5"}}`,
		`{"timestamp":"2026-07-09T14:09:16Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello current user"}]}}`,
	}
	codexPath := filepath.Join(sessionsDir, "rollout-2026-07-09T14-09-12-"+sessionID+".jsonl")
	if err := os.WriteFile(codexPath, []byte(strings.Join(codexLines, "\n")), 0o644); err != nil {
		t.Fatalf("failed to write codex session file: %v", err)
	}

	state := newTestState(&FakeRunner{})
	ctx := context.Background()

	err := RunAgentSessionLogCodex(ctx, state, sessionID, cwd)
	if err != nil {
		t.Fatalf("RunAgentSessionLogCodex failed: %v", err)
	}

	outputDir := filepath.Join(tempDir, ".agents", "sessions")
	var logFiles []string
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			logFiles = append(logFiles, path)
		}
		return nil
	})

	if len(logFiles) != 1 {
		t.Fatalf("expected exactly 1 log file, found %d", len(logFiles))
	}

	logContent, _ := os.ReadFile(logFiles[0])
	lines := strings.Split(strings.TrimSpace(string(logContent)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 log lines, got %d", len(lines))
	}

	var firstLine, secondLine SessionLogLine
	_ = json.Unmarshal([]byte(lines[0]), &firstLine)
	_ = json.Unmarshal([]byte(lines[1]), &secondLine)

	if firstLine.Role != "user" || firstLine.Content != "hello current codex" || firstLine.CWD != cwd || firstLine.Model != "gpt-5.5" {
		t.Errorf("unexpected first line: %+v", firstLine)
	}
	if secondLine.Role != "assistant" || secondLine.Content != "hello current user" || secondLine.CWD != cwd || secondLine.Model != "gpt-5.5" {
		t.Errorf("unexpected second line: %+v", secondLine)
	}
}

func TestRunAgentSessionSync(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// 1. Setup an existing processed session log for codex
	processedDir := filepath.Join(tempDir, ".agents", "sessions", "2026-07-09")
	if err := os.MkdirAll(processedDir, 0o755); err != nil {
		t.Fatalf("failed to create processed sessions dir: %v", err)
	}
	existingLogFile := filepath.Join(processedDir, "120000_codex_existing-session-123.jsonl")
	if err := os.WriteFile(existingLogFile, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to write existing log file: %v", err)
	}

	// 2. Setup codex rollout directory with one existing session and one new session
	sessionsDir := filepath.Join(tempDir, ".codex", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create codex sessions dir: %v", err)
	}

	// Existing/processed rollout file
	existingContent := `{"timestamp":"2026-07-09T12:00:00Z","type":"session_meta","id":"existing-session-123","cwd":"/workspace/test"}`
	existingRolloutPath := filepath.Join(sessionsDir, "rollout-2026-07-09T12-00-00-existing-session-123.jsonl")
	if err := os.WriteFile(existingRolloutPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("failed to write existing rollout: %v", err)
	}

	// New rollout file (unprocessed)
	newContent := `{"timestamp":"2026-07-09T15:00:00Z","type":"user_message","payload":{"message":"hello new sync"},"cwd":"/workspace/test"}`
	newRolloutPath := filepath.Join(sessionsDir, "rollout-2026-07-09T15-00-00-new-session-456.jsonl")
	if err := os.WriteFile(newRolloutPath, []byte(newContent), 0o644); err != nil {
		t.Fatalf("failed to write new rollout: %v", err)
	}

	state := newTestState(&FakeRunner{})
	ctx := context.Background()

	err := RunAgentSessionSync(ctx, state)
	if err != nil {
		t.Fatalf("RunAgentSessionSync failed: %v", err)
	}

	// Verify that ONLY the new session log was written to ~/.agents/sessions/YYYY-MM-DD/HHMMSS_codex_new-session-456.jsonl
	outputDir := filepath.Join(tempDir, ".agents", "sessions")
	var logFiles []string
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			// Skip the mock existing file we created manually
			if !strings.Contains(path, "existing-session-123") {
				logFiles = append(logFiles, path)
			}
		}
		return nil
	})

	if len(logFiles) != 1 {
		t.Fatalf("expected exactly 1 new synced log file, found %d: %v", len(logFiles), logFiles)
	}

	if !strings.Contains(logFiles[0], "new-session-456") {
		t.Errorf("expected synced file to be for new-session-456, got %s", logFiles[0])
	}
}
