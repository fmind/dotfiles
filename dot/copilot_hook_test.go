package dot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func copilotSessionEndPayload(sessionID, reason string) string {
	return fmt.Sprintf(`{"sessionId":%q,"timestamp":1785600000000,"cwd":"/work/project","reason":%q}`, sessionID, reason)
}

func TestCopilotSessionEndTargetedSyncIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".copilot"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(copilotDBPath(home), []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := &FakeRunner{RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
		if name != "sqlite3" || !strings.Contains(strings.Join(args, " "), "session-live") {
			return "", fmt.Errorf("unexpected targeted query: %s %v", name, args)
		}
		return `[{"session_id":"session-live","turn_index":1,"user_message":"prompt","assistant_response":"response","timestamp":"2026-08-01T10:00:00Z","cwd":"/work/project"}]`, nil
	}}
	state := newTestState(runner)
	var output bytes.Buffer
	state.Stdout = &output
	for range 2 {
		state.Stdin = strings.NewReader(copilotSessionEndPayload("session-live", "complete"))
		if err := RunCopilotSessionEndHook(context.Background(), state); err != nil {
			t.Fatalf("sessionEnd blocked Copilot: %v", err)
		}
	}
	root, err := sessionStoreRoot()
	if err != nil {
		t.Fatal(err)
	}
	manifests, err := filepath.Glob(filepath.Join(root, sessionStoreCopilot, "*", "*", "manifest.json"))
	if err != nil || len(manifests) != 1 {
		t.Fatalf("targeted and full-style ingestion did not converge: %v %v", manifests, err)
	}
	if strings.Count(output.String(), "{}") != 2 {
		t.Fatalf("hook did not emit neutral output: %q", output.String())
	}
}

func TestCopilotSessionEndFailuresNeverBlockAndRemainActionable(t *testing.T) {
	tests := map[string]string{
		"malformed":            `{`,
		"missing store":        copilotSessionEndPayload("session-missing", "complete"),
		"unsupported reason":   copilotSessionEndPayload("session-version", "future_reason"),
		"undocumented payload": `{"sessionId":"session-extra","timestamp":1785600000000,"cwd":"/work","reason":"complete","environment":{"TOKEN":"secret"}}`,
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			state := newTestState(&FakeRunner{})
			state.Stdin = strings.NewReader(payload)
			var stdout, stderr bytes.Buffer
			state.Stdout, state.Stderr = &stdout, &stderr
			if err := RunCopilotSessionEndHook(context.Background(), state); err != nil {
				t.Fatalf("hook failure blocked Copilot: %v", err)
			}
			records := readHookFailureRecords(t, home)
			if len(records) != 1 || records[0].Agent != sessionStoreCopilot || records[0].Operation != "sessionEnd" || records[0].Detail == "" {
				t.Fatalf("failure was not actionable: %+v", records)
			}
			if strings.Contains(records[0].Detail, "TOKEN") || strings.Contains(records[0].Detail, "secret") {
				t.Fatalf("failure spool retained unrelated payload data: %+v", records[0])
			}
			if strings.TrimSpace(stdout.String()) != "{}" || !strings.Contains(stderr.String(), "without blocking") {
				t.Fatalf("unexpected non-blocking hook output: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestCopilotSessionEndAcceptsEveryDocumentedReason(t *testing.T) {
	for _, reason := range []string{"complete", "error", "abort", "timeout", "user_exit"} {
		t.Run(reason, func(t *testing.T) {
			input, err := decodeCopilotSessionEnd(strings.NewReader(copilotSessionEndPayload("session-reason", reason)))
			if err != nil {
				t.Fatalf("documented reason %q was rejected: %v", reason, err)
			}
			if string(input.Reason) != reason {
				t.Fatalf("decoded reason = %q, want %q", input.Reason, reason)
			}
		})
	}
}

func TestCopilotHookConfigAndVersionContract(t *testing.T) {
	content, err := os.ReadFile("../dot_copilot/hooks/session-log.json")
	if err != nil {
		t.Fatal(err)
	}
	integration := agentIntegration{Agent: sessionStoreCopilot, HookPath: "../dot_copilot/hooks/session-log.json", HookCommands: []string{"dot agent hook copilot-session-end"}, HookJSON: true}
	if status, ok := checkAgentHooks(integration); !ok || status != "healthy" {
		t.Fatalf("managed hook config is unhealthy: %s", status)
	}
	unsupported := filepath.Join(t.TempDir(), "hooks.json")
	if err := os.WriteFile(unsupported, bytes.Replace(content, []byte(`"version": 1`), []byte(`"version": 2`), 1), 0o600); err != nil {
		t.Fatal(err)
	}
	integration.HookPath = unsupported
	if status, ok := checkAgentHooks(integration); ok || status != "unsupported-version" {
		t.Fatalf("unsupported hook schema was accepted: %s", status)
	}
}
