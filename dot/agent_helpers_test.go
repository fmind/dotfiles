package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// errReader fails on the first read so the io.ReadAll error path of parseStdin
// stays reachable without a real broken pipe.
type errReader struct{ err error }

func (e errReader) Read([]byte) (int, error) { return 0, e.err }

func TestParseStdin(t *testing.T) {
	t.Run("nil reader yields no hook input", func(t *testing.T) {
		got, err := parseStdin(nil)
		if err != nil || got != nil {
			t.Fatalf("parseStdin(nil) = %v, %v; want nil, nil", got, err)
		}
	})

	t.Run("a terminal stdin yields no hook input", func(t *testing.T) {
		// /dev/null is a character device, which is how parseStdin recognizes an
		// interactive invocation with nothing piped in.
		file, err := os.Open(os.DevNull)
		if err != nil {
			t.Skipf("cannot open %s: %v", os.DevNull, err)
		}
		defer func() { _ = file.Close() }()

		got, err := parseStdin(file)
		if err != nil || got != nil {
			t.Fatalf("parseStdin(devnull) = %v, %v; want nil, nil", got, err)
		}
	})

	t.Run("an unstattable stdin surfaces the error", func(t *testing.T) {
		file, err := os.Open(os.DevNull)
		if err != nil {
			t.Skipf("cannot open %s: %v", os.DevNull, err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}

		if _, err := parseStdin(file); err == nil {
			t.Fatal("expected a stat error for a closed stdin")
		}
	})

	t.Run("a read failure surfaces the error", func(t *testing.T) {
		wantErr := errors.New("broken pipe")
		if _, err := parseStdin(errReader{err: wantErr}); !errors.Is(err, wantErr) {
			t.Fatalf("expected the read error, got %v", err)
		}
	})

	t.Run("empty input yields no hook input", func(t *testing.T) {
		got, err := parseStdin(strings.NewReader(""))
		if err != nil || got != nil {
			t.Fatalf("parseStdin(empty) = %v, %v; want nil, nil", got, err)
		}
	})

	t.Run("malformed JSON is rejected", func(t *testing.T) {
		_, err := parseStdin(strings.NewReader("{not json"))
		if err == nil || !strings.Contains(err.Error(), "failed to parse agent hook input") {
			t.Fatalf("expected a parse error, got %v", err)
		}
	})

	t.Run("a payload without a trailing newline is parsed", func(t *testing.T) {
		got, err := parseStdin(strings.NewReader(`{"session_id":"abc","cwd":"/tmp"}`))
		if err != nil {
			t.Fatalf("parseStdin: %v", err)
		}
		if got == nil || got.SessionID != "abc" || got.CWD != "/tmp" {
			t.Fatalf("parseStdin = %+v; want session abc in /tmp", got)
		}
	})
}

func TestHookInputRejectsMalformedJSON(t *testing.T) {
	var input HookInput
	if err := input.UnmarshalJSON([]byte("[]")); err == nil {
		t.Fatal("expected an error when the hook payload is not an object")
	}
}

func TestSessionLineageIdentityIsStableAndAgentScoped(t *testing.T) {
	first := sessionLineageID("claude", "sess-1")
	if first != sessionLineageID("claude", "sess-1") {
		t.Fatal("lineage identity is not stable")
	}
	if first == sessionLineageID("codex", "sess-1") {
		t.Fatal("the same bare session ID collided across agents")
	}
	if strings.Contains(first, "sess-1") || len(first) != 64 {
		t.Fatalf("lineage identity leaks its raw session ID: %q", first)
	}
}

func TestWriteSessionLogsIgnoresEmptyInput(t *testing.T) {
	// No HOME override: an empty batch must return before touching the filesystem.
	t.Setenv("HOME", "")
	result, err := writeSessionLogs(context.Background(), newTestState(&FakeRunner{}), "claude", "sess-1", nil, sessionSource{Type: "test"})
	if err != nil {
		t.Fatalf("writeSessionLogs(nil) = %v, want nil", err)
	}
	if result.Status != sessionSkipped {
		t.Fatalf("empty input status = %q, want skipped", result.Status)
	}
}

func TestResolveCWD(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct{ name, in, want string }{
		{name: "empty stays empty", in: "", want: ""},
		{name: "dot resolves to the working directory", in: ".", want: pwd},
		{name: "relative resolves against the working directory", in: "sub", want: filepath.Join(pwd, "sub")},
		{name: "absolute is kept", in: "/var/tmp", want: "/var/tmp"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCWD(tc.in); got != tc.want {
				t.Errorf("resolveCWD(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsValidSessionID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{in: "", want: false},
		{in: "abc-123_XYZ", want: true},
		{in: "../escape", want: false},
		{in: "with/slash", want: false},
		{in: "with space", want: false},
	}
	for _, tc := range tests {
		if got := isValidSessionID(tc.in); got != tc.want {
			t.Errorf("isValidSessionID(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestExtractCodexSessionID(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{name: "non-rollout names are ignored", in: "session-abc", want: ""},
		{name: "too few segments yields nothing", in: "rollout-2026-07-31", want: ""},
		{name: "the trailing segments form the id", in: "rollout-2026-07-31T09-08-07-abc-def", want: "abc-def"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractCodexSessionID(tc.in); got != tc.want {
				t.Errorf("extractCodexSessionID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMapValue(t *testing.T) {
	if got := mapValue(map[string]any{"a": 1}); got == nil {
		t.Error("expected a map to be returned unchanged")
	}
	if got := mapValue("not a map"); got != nil {
		t.Errorf("expected nil for a non-map value, got %v", got)
	}
}

func TestTextFromCodexContent(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "plain string", in: "hello", want: "hello"},
		{name: "string parts are joined", in: []any{"a", "b"}, want: "a\nb"},
		{name: "text keys are preferred", in: []any{map[string]any{"text": "t", "content": "c"}}, want: "t"},
		{name: "content is the fallback key", in: []any{map[string]any{"content": "c"}}, want: "c"},
		{name: "unknown part shapes are skipped", in: []any{42, map[string]any{}}, want: ""},
		{name: "unsupported types yield nothing", in: 42, want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := textFromCodexContent(tc.in); got != tc.want {
				t.Errorf("textFromCodexContent(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCodexFieldExtraction(t *testing.T) {
	t.Run("role", func(t *testing.T) {
		tests := []struct {
			name string
			raw  map[string]any
			want string
		}{
			{name: "top level", raw: map[string]any{"role": "user"}, want: "user"},
			{name: "payload", raw: map[string]any{"payload": map[string]any{"role": "assistant"}}, want: "assistant"},
			{name: "user event type", raw: map[string]any{"type": "user_message"}, want: "user"},
			{name: "agent event type", raw: map[string]any{"type": "agent_message"}, want: "assistant"},
			{name: "unknown", raw: map[string]any{"type": "tool_call"}, want: ""},
		}
		for _, tc := range tests {
			if got := codexRole(tc.raw); got != tc.want {
				t.Errorf("codexRole(%s) = %q, want %q", tc.name, got, tc.want)
			}
		}
	})

	t.Run("content", func(t *testing.T) {
		tests := []struct {
			name string
			raw  map[string]any
			want string
		}{
			{name: "top level content", raw: map[string]any{"content": "top"}, want: "top"},
			{name: "payload content", raw: map[string]any{"payload": map[string]any{"content": "pc"}}, want: "pc"},
			{name: "payload message", raw: map[string]any{"payload": map[string]any{"message": "pm"}}, want: "pm"},
			{name: "payload text", raw: map[string]any{"payload": map[string]any{"text": "pt"}}, want: "pt"},
			{name: "top level message", raw: map[string]any{"message": "m"}, want: "m"},
			{name: "top level text", raw: map[string]any{"text": "x"}, want: "x"},
			{name: "nothing", raw: map[string]any{}, want: ""},
		}
		for _, tc := range tests {
			if got := codexContent(tc.raw); got != tc.want {
				t.Errorf("codexContent(%s) = %q, want %q", tc.name, got, tc.want)
			}
		}
	})

	t.Run("model", func(t *testing.T) {
		tests := []struct {
			name string
			raw  map[string]any
			want string
		}{
			{name: "top level", raw: map[string]any{"model": "m1"}, want: "m1"},
			{name: "payload", raw: map[string]any{"payload": map[string]any{"model": "m2"}}, want: "m2"},
			{name: "nothing", raw: map[string]any{}, want: ""},
		}
		for _, tc := range tests {
			if got := codexModel(tc.raw); got != tc.want {
				t.Errorf("codexModel(%s) = %q, want %q", tc.name, got, tc.want)
			}
		}
	})

	t.Run("cwd", func(t *testing.T) {
		tests := []struct {
			name string
			raw  map[string]any
			want string
		}{
			{name: "top level", raw: map[string]any{"cwd": "/a"}, want: "/a"},
			{name: "payload", raw: map[string]any{"payload": map[string]any{"cwd": "/b"}}, want: "/b"},
			{name: "nothing", raw: map[string]any{}, want: ""},
		}
		for _, tc := range tests {
			if got := codexCWD(tc.raw); got != tc.want {
				t.Errorf("codexCWD(%s) = %q, want %q", tc.name, got, tc.want)
			}
		}
	})
}

func TestDecodeJSONL(t *testing.T) {
	openFixture := func(t *testing.T, content string) *os.File {
		t.Helper()
		path := filepath.Join(t.TempDir(), "transcript.jsonl")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = file.Close() })
		return file
	}

	t.Run("malformed lines are warned about and skipped", func(t *testing.T) {
		file := openFixture(t, "{\"a\":1}\nnot json\n{\"a\":2}\n")
		var warnings bytes.Buffer
		var seen int

		if _, err := decodeJSONLWithStats(&warnings, file.Name(), file, func(map[string]any) error {
			seen++
			return nil
		}); err != nil {
			t.Fatalf("decodeJSONLWithStats: %v", err)
		}
		if seen != 2 {
			t.Errorf("callback ran %d times, want 2", seen)
		}
		if !strings.Contains(warnings.String(), "failed to decode JSON line") {
			t.Errorf("expected a decode warning, got %q", warnings.String())
		}
	})

	t.Run("a callback error aborts the scan", func(t *testing.T) {
		file := openFixture(t, "{\"a\":1}\n{\"a\":2}\n")
		wantErr := errors.New("stop")
		var seen int

		_, err := decodeJSONLWithStats(io.Discard, file.Name(), file, func(map[string]any) error {
			seen++
			return wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected the callback error, got %v", err)
		}
		if seen != 1 {
			t.Errorf("callback ran %d times, want 1 before aborting", seen)
		}
	})

	t.Run("a read failure is wrapped with the file path", func(t *testing.T) {
		file := openFixture(t, "{\"a\":1}\n")
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}

		_, err := decodeJSONLWithStats(io.Discard, file.Name(), file, func(map[string]any) error { return nil })
		if err == nil || !strings.Contains(err.Error(), file.Name()) {
			t.Fatalf("expected a read error naming %s, got %v", file.Name(), err)
		}
	})
}

// sessionLogCommands are the per-agent entry points that share the same session_id
// validation and hook-payload contract.
var sessionLogCommands = map[string]func(context.Context, *GlobalState, string, string) error{
	"agy":      RunAgentSessionLogAgy,
	"claude":   RunAgentSessionLogClaude,
	"codex":    RunAgentSessionLogCodex,
	"opencode": RunAgentSessionLogOpencode,
	"copilot":  RunAgentSessionLogCopilot,
}

func TestSessionLogCommandsRejectUnsafeSessionIDs(t *testing.T) {
	ctx := context.Background()
	// A session ID is interpolated into file paths and SQL, so traversal and quoting
	// characters must be rejected before any of that happens.
	for agent, run := range sessionLogCommands {
		for _, sessionID := range []string{"../../etc/passwd", "abc'; DROP TABLE turns;--"} {
			t.Run(agent+"/"+sessionID, func(t *testing.T) {
				state := newTestState(&FakeRunner{
					RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
						t.Error("no command must run for an invalid session ID")
						return "", nil
					},
				})
				err := run(ctx, state, sessionID, "")
				if err == nil || !strings.Contains(err.Error(), "invalid session_id format") {
					t.Fatalf("expected an invalid session_id error, got %v", err)
				}
			})
		}
	}
}

func TestSessionLogCommandsRequireSessionID(t *testing.T) {
	ctx := context.Background()
	for agent, run := range sessionLogCommands {
		t.Run(agent, func(t *testing.T) {
			err := run(ctx, newTestState(&FakeRunner{}), "", "")
			if err == nil || !strings.Contains(err.Error(), "missing session_id") {
				t.Fatalf("expected a missing session_id error, got %v", err)
			}
		})
	}
}

func TestSessionLogCommandsHonorStopHookActive(t *testing.T) {
	ctx := context.Background()
	// A Stop hook that re-entered must return immediately, or the agent would loop.
	for _, agent := range []string{"claude", "codex", "opencode"} {
		t.Run(agent, func(t *testing.T) {
			state := newTestState(&FakeRunner{})
			state.Stdin = strings.NewReader(`{"session_id":"abc","stop_hook_active":true}`)

			if err := sessionLogCommands[agent](ctx, state, "", ""); err != nil {
				t.Fatalf("expected an early return, got %v", err)
			}
		})
	}
}

func TestSessionLogCommandsSurfaceStdinErrors(t *testing.T) {
	ctx := context.Background()
	for _, agent := range []string{"agy", "claude", "codex", "opencode"} {
		t.Run(agent, func(t *testing.T) {
			state := newTestState(&FakeRunner{})
			state.Stdin = strings.NewReader("{not json")

			err := sessionLogCommands[agent](ctx, state, "abc", "")
			if err == nil || !strings.Contains(err.Error(), "failed to parse agent hook input") {
				t.Fatalf("expected a hook parse error, got %v", err)
			}
		})
	}
}

func TestSQLiteBackedCommandsReportMissingDatabase(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		agent   string
		wantMsg string
	}{
		{agent: "opencode", wantMsg: "opencode database not found at"},
		{agent: "copilot", wantMsg: "copilot database not found at"},
	}
	for _, tc := range tests {
		t.Run(tc.agent, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			err := sessionLogCommands[tc.agent](ctx, newTestState(&FakeRunner{}), "abc", "")
			if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("expected %q, got %v", tc.wantMsg, err)
			}
		})
	}
}

func TestSQLiteBackedCommandsHandleQueryResults(t *testing.T) {
	ctx := context.Background()
	// Both stores live under HOME; seeding an empty file is enough since the query
	// itself is faked through the runner.
	seedDB := func(t *testing.T, rel ...string) string {
		t.Helper()
		home := t.TempDir()
		t.Setenv("HOME", home)
		path := filepath.Join(append([]string{home}, rel...)...)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		return home
	}
	dbFor := map[string][]string{
		"opencode": {".local", "share", "opencode", "opencode.db"},
		"copilot":  nil, // resolved below via copilotDBPath
	}

	for _, agent := range []string{"opencode", "copilot"} {
		t.Run(agent, func(t *testing.T) {
			rel := dbFor[agent]
			if rel == nil {
				home := t.TempDir()
				rel = strings.Split(strings.TrimPrefix(filepath.Join(home, ".copilot", "session-store.db"), home+string(filepath.Separator)), string(filepath.Separator))
			}

			t.Run("query failure is surfaced", func(t *testing.T) {
				seedDB(t, rel...)
				state := newTestState(&FakeRunner{
					RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
						return "", errors.New("sqlite3 exploded")
					},
				})
				if err := sessionLogCommands[agent](ctx, state, "abc", ""); err == nil {
					t.Fatal("expected the sqlite3 failure to be surfaced")
				}
			})

			t.Run("an empty result set writes nothing", func(t *testing.T) {
				home := seedDB(t, rel...)
				state := newTestState(&FakeRunner{
					RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
						return "[]\n", nil
					},
				})
				if err := sessionLogCommands[agent](ctx, state, "abc", ""); err != nil {
					t.Fatalf("expected no error for an empty result, got %v", err)
				}
				if _, err := os.Stat(filepath.Join(home, ".agents", "sessions")); !os.IsNotExist(err) {
					t.Errorf("expected no session log to be written, stat err = %v", err)
				}
			})

			t.Run("malformed JSON is rejected", func(t *testing.T) {
				seedDB(t, rel...)
				state := newTestState(&FakeRunner{
					RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
						return "{not json", nil
					},
				})
				err := sessionLogCommands[agent](ctx, state, "abc", "")
				if err == nil || !strings.Contains(err.Error(), "query result") {
					t.Fatalf("expected a query result parse error, got %v", err)
				}
			})
		})
	}
}

func TestDecodeSessionIDs(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr string
		want    []string
	}{
		{name: "null decodes to an explicit error", output: "null", wantErr: "expected a JSON array"},
		{name: "malformed JSON is rejected", output: "{", wantErr: "failed to decode"},
		{
			name:    "duplicate session IDs are rejected",
			output:  `[{"id":"a"},{"id":"a"}]`,
			wantErr: "duplicate session ID",
		},
		{name: "valid sessions preserve source order", output: `[{"id":"a"},{"id":"b"}]`, want: []string{"a", "b"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeSessionIDs("OpenCode", tc.output)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected an error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeSessionIDs: %v", err)
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("decodeSessionIDs = %v, want %v", got, tc.want)
			}
		})
	}
}

func marshalCopilotRows(t *testing.T, rows []CopilotRow) string {
	t.Helper()
	data, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestSyncCopilotSessionsFailures mirrors the OpenCode sync failure matrix: every
// malformed or unexpected row must abort the sync rather than write a partial log.
func TestSyncCopilotSessionsFailures(t *testing.T) {
	validSessions := `[{"id":"ses-valid_1"}]`
	validRows := marshalCopilotRows(t, []CopilotRow{{
		SessionID:         "ses-valid_1",
		TurnIndex:         0,
		UserMessage:       "hello",
		AssistantResponse: "hi",
		Timestamp:         "2026-07-31T09:00:00Z",
		CWD:               "/workspace",
	}})

	tests := []struct {
		sessionErr    error
		turnErr       error
		name          string
		sessionOutput string
		turnOutput    string
		wantErr       string
		breakWriter   bool
	}{
		{name: "session query", sessionErr: errors.New("boom"), wantErr: "failed to query Copilot sessions"},
		{name: "session decode", sessionOutput: "{", wantErr: "failed to decode Copilot session query result"},
		{name: "turn query", sessionOutput: validSessions, turnErr: errors.New("boom"), wantErr: "failed to query Copilot turns"},
		{name: "turn decode", sessionOutput: validSessions, turnOutput: "{", wantErr: "failed to decode Copilot turn query result"},
		{name: "turn null", sessionOutput: validSessions, turnOutput: "null", wantErr: "expected a JSON array"},
		{
			name:          "invalid turn session ID",
			sessionOutput: validSessions,
			turnOutput:    marshalCopilotRows(t, []CopilotRow{{SessionID: "bad!id", UserMessage: "x"}}),
			wantErr:       "invalid session ID",
		},
		{
			name:          "unexpected turn session ID",
			sessionOutput: validSessions,
			turnOutput:    marshalCopilotRows(t, []CopilotRow{{SessionID: "ses-other_1", UserMessage: "x"}}),
			wantErr:       "unexpected session ID",
		},
		{
			name:          "session write",
			sessionOutput: validSessions,
			turnOutput:    validRows,
			breakWriter:   true,
			wantErr:       "failed to write Copilot session",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			if tc.breakWriter {
				// A file where the sessions directory belongs makes every write fail.
				if err := os.MkdirAll(filepath.Join(home, ".agents"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(home, ".agents", "sessions"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}

			state := newTestState(&FakeRunner{
				RunFunc: func(_ context.Context, _ string, _ io.Reader, _ string, args ...string) (string, error) {
					// The sessions query carries no WHERE clause; the turns query does.
					if strings.Contains(args[len(args)-1], "session_id IN (") {
						return tc.turnOutput, tc.turnErr
					}
					return tc.sessionOutput, tc.sessionErr
				},
			})

			_, err := syncCopilotSessions(context.Background(), state, filepath.Join(home, ".copilot", "session-store.db"))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected an error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestSyncCopilotSessionsWritesLogs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	state := newTestState(&FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, _ string, args ...string) (string, error) {
			if strings.Contains(args[len(args)-1], "session_id IN (") {
				return marshalCopilotRows(t, []CopilotRow{{
					SessionID:         "ses-valid_1",
					UserMessage:       "hello",
					AssistantResponse: "hi",
					Timestamp:         "2026-07-31T09:00:00Z",
					CWD:               "/workspace",
				}}), nil
			}
			// One source has no turns and must be reported as skipped.
			return `[{"id":"ses-valid_1"},{"id":"ses-skipped"}]`, nil
		},
	})

	count, err := syncCopilotSessions(context.Background(), state, filepath.Join(home, ".copilot", "session-store.db"))
	if err != nil {
		t.Fatalf("syncCopilotSessions: %v", err)
	}
	if count != 1 {
		t.Fatalf("synced %d sessions, want 1", count)
	}
	logs := readAgentSessionLogs(t, home, "copilot")
	if len(logs["ses-valid_1"]) != 2 {
		t.Fatalf("expected one normalized copilot session, got %+v", logs)
	}
}

func TestRunAgentSessionSyncRejectsDirectoryDatabases(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  func(home string) string
		wantMsg string
	}{
		{
			name:    "opencode",
			dbPath:  func(home string) string { return filepath.Join(home, ".local", "share", "opencode", "opencode.db") },
			wantMsg: "OpenCode database path is a directory",
		},
		{
			name:    "copilot",
			dbPath:  func(home string) string { return filepath.Join(home, ".copilot", "session-store.db") },
			wantMsg: "Copilot database path is a directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			if err := os.MkdirAll(tc.dbPath(home), 0o755); err != nil {
				t.Fatal(err)
			}

			err := RunAgentSessionSync(context.Background(), newTestState(&FakeRunner{}))
			if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("expected %q, got %v", tc.wantMsg, err)
			}
		})
	}
}
