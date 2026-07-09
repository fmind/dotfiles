package dot

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	ctx := context.Background()

	err := RunAgentSessionLogAgy(ctx, state, sessionID, cwd)
	if err != nil {
		t.Fatalf("RunAgentSessionLogAgy failed: %v", err)
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

func TestRunAgentSessionClean(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	sessionsDir := filepath.Join(tempDir, ".agents", "sessions")

	// Create directories representing old and new dates
	oldDateDir := filepath.Join(sessionsDir, "2026-05-01")
	newDateDir := filepath.Join(sessionsDir, "2026-07-09")
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
