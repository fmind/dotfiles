package dot

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readHookFailureRecords(t *testing.T, home string) []hookFailureRecord {
	t.Helper()
	root := filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("failed to read hook failure spool: %v", err)
	}
	records := make([]hookFailureRecord, 0, len(entries))
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var record hookFailureRecord
		if err := json.Unmarshal(content, &record); err != nil {
			t.Fatalf("invalid failure record: %v", err)
		}
		records = append(records, record)
	}
	return records
}

func TestAgentHookSpoolsBoundedOwnerOnlyFailures(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	state := newTestState(&FakeRunner{})
	state.Stdin = strings.NewReader("{")
	err := RunAgentHookSession(context.Background(), state, sessionStoreClaude, "private-session-id", "")
	if err == nil || !strings.Contains(err.Error(), "failed to parse agent hook input") {
		t.Fatalf("expected original hook error, got %v", err)
	}

	root := filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion)
	rootInfo, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if rootInfo.Mode().Perm() != 0o700 {
		t.Fatalf("spool permissions = %o, want 700", rootInfo.Mode().Perm())
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one failure record, got %v: %v", entries, err)
	}
	info, err := entries[0].Info()
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("record permissions = %o, want 600", info.Mode().Perm())
	}
	records := readHookFailureRecords(t, home)
	if records[0].Agent != sessionStoreClaude || records[0].Operation != "session" || len(records[0].Detail) > hookFailureDetailLimit {
		t.Fatalf("unexpected bounded record: %+v", records[0])
	}
	content, err := os.ReadFile(filepath.Join(root, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "private-session-id") {
		t.Fatal("failure record exposed the raw session identity")
	}
}

func TestAgentHookFailureSpoolRetentionAndLinkSafety(t *testing.T) {
	t.Run("retains newest bounded records", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		for index := range hookFailureLimit + 5 {
			err := spoolHookFailure(sessionStoreCodex, "session", "sid", errors.New(strings.Repeat("x", hookFailureDetailLimit+index+1)))
			if err == nil {
				t.Fatal("expected original hook failure")
			}
		}
		records := readHookFailureRecords(t, home)
		if len(records) != hookFailureLimit {
			t.Fatalf("spool retained %d records, want %d", len(records), hookFailureLimit)
		}
	})

	t.Run("refuses linked spool", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		outside := filepath.Join(home, "outside")
		if err := os.MkdirAll(outside, 0o700); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion)
		if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, link); err != nil {
			t.Fatal(err)
		}
		err := spoolHookFailure(sessionStoreCodex, "session", "sid", errors.New("failed"))
		if err == nil || !strings.Contains(err.Error(), "refusing linked hook failure spool") {
			t.Fatalf("expected linked spool rejection, got %v", err)
		}
		entries, readErr := os.ReadDir(outside)
		if readErr != nil || len(entries) != 0 {
			t.Fatalf("linked target changed: %v %v", entries, readErr)
		}
	})
}

func TestAgentHookRejectsUnknownAgentAndPreservesFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := RunAgentHookSession(context.Background(), newTestState(&FakeRunner{}), "unknown", "secret-id", "")
	if err == nil || !strings.Contains(err.Error(), `unknown session hook agent "unknown"`) {
		t.Fatalf("expected unknown-agent error, got %v", err)
	}
	records := readHookFailureRecords(t, home)
	if len(records) != 1 || records[0].SessionHash == "" || records[0].SessionHash == "secret-id" {
		t.Fatalf("expected hashed session evidence, got %+v", records)
	}
	if strings.Contains(records[0].Detail, "secret-id") {
		t.Fatalf("failure detail exposed raw session identity: %+v", records[0])
	}
	if detail := boundedHookFailureDetail(errors.New("session secret-id failed"), "secret-id"); strings.Contains(detail, "secret-id") {
		t.Fatalf("bounded failure detail exposed raw session identity: %s", detail)
	}
}

func TestAgentNotifyHookSpoolsFailures(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := RunAgentHookNotify(context.Background(), newTestState(&FakeRunner{}), sessionStoreClaude, "unknown-event")
	if err == nil || !strings.Contains(err.Error(), "unknown agent notify event") {
		t.Fatalf("expected notification error, got %v", err)
	}
	records := readHookFailureRecords(t, home)
	if len(records) != 1 || records[0].Operation != "notify:unknown-event" {
		t.Fatalf("notification failure was not recorded: %+v", records)
	}
}

func TestAgentHookTemplatesUseObservableBoundary(t *testing.T) {
	tests := map[string][]string{
		"../dot_claude/settings.json.tmpl":                {"dot agent hook session claude", "dot agent hook notify claude"},
		"../dot_codex/modify_private_config.toml":         {"dot agent hook session codex", "dot agent hook notify codex stop"},
		"../dot_gemini/private_config/private_hooks.json": {"dot agent hook session agy", "dot agent hook notify agy stop"},
		"../dot_config/opencode/plugins/session-log.ts":   {"dot agent hook session opencode"},
		"../dot_copilot/hooks/session-log.json":           {"dot agent hook copilot-session-end"},
	}
	for path, commands := range tests {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		text := string(content)
		for _, command := range commands {
			if !strings.Contains(text, command) {
				t.Fatalf("%s omitted %q", path, command)
			}
		}
		if strings.Contains(text, "2>/dev/null") || strings.Contains(text, "|| true") {
			t.Fatalf("%s still silently discards hook failures", path)
		}
	}
}
