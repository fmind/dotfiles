package dot

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func rawSessionTestPath(home, source, sessionID string) string {
	switch source {
	case sessionStoreClaude:
		return filepath.Join(home, ".claude", "projects", "project", sessionID+".jsonl")
	case sessionStoreCodex:
		return filepath.Join(home, ".codex", "sessions", "2026", "07", "31", "rollout-2026-07-31T12-00-00-"+sessionID+".jsonl")
	case sessionStoreAgy:
		return filepath.Join(home, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript.jsonl")
	case sessionStoreGrok:
		return filepath.Join(home, ".grok", "sessions", sessionID, grokTranscriptName)
	default:
		return ""
	}
}

func rawSessionTestRoot(home, source string) string {
	switch source {
	case sessionStoreClaude:
		return filepath.Join(home, ".claude", "projects")
	case sessionStoreCodex:
		return filepath.Join(home, ".codex", "sessions")
	case sessionStoreAgy:
		return filepath.Join(home, ".gemini", "antigravity-cli", "brain")
	case sessionStoreGrok:
		return filepath.Join(home, ".grok", "sessions")
	default:
		return ""
	}
}

func copySessionGeneration(t *testing.T, source, target string) {
	t.Helper()
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatalf("failed to create duplicate generation: %v", err)
	}
	for _, name := range []string{"manifest.json", "transcript.jsonl"} {
		content, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatalf("failed to read generation file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(target, name), content, 0o600); err != nil {
			t.Fatalf("failed to copy generation file: %v", err)
		}
	}
}

func TestPruneRawSessionsRequiresVerifiedSuccessor(t *testing.T) {
	for _, source := range []string{sessionStoreClaude, sessionStoreCodex, sessionStoreAgy, sessionStoreGrok} {
		t.Run(source, func(t *testing.T) {
			state, out, home := newPruneTestState(t, &FakeRunner{})
			sessionID := source + "-verified"
			raw := rawSessionTestPath(home, source, sessionID)
			writeAged(t, raw, "raw transcript", 30*24*time.Hour)
			result := writeCompleteSessionSuccessor(t, source, sessionID, raw)

			generation := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, source, result.LineageID, result.GenerationID)
			manifestInfo, err := os.Stat(filepath.Join(generation, "manifest.json"))
			if err != nil {
				t.Fatalf("failed to inspect successor manifest: %v", err)
			}
			state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: rawSessionTestRoot(home, source), Source: source, KeepDays: 7}}
			if err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7, DryRun: true}); err != nil {
				t.Fatalf("dry-run returned an error: %v", err)
			}
			requireExists(t, raw)
			report := out.String()
			for _, field := range []string{"decision=delete", "source=" + source, "lineage=", "age=", "size=", "reason=verified-successor", "successor=generation="} {
				if !strings.Contains(report, field) {
					t.Fatalf("expected %q in dry-run report: %s", field, report)
				}
			}
			out.Reset()
			if err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7}); err != nil {
				t.Fatalf("RunPrune returned an error: %v", err)
			}

			requireGone(t, raw)
			requireExists(t, generation)
			if manifestInfo.Mode().Perm() != 0o600 {
				t.Fatalf("successor manifest permissions changed: %o", manifestInfo.Mode().Perm())
			}
		})
	}
}

func TestPruneRawSessionsRetainsUnverifiedSources(t *testing.T) {
	tests := []struct {
		setup  func(t *testing.T, home, raw, sessionID string)
		name   string
		reason string
	}{
		{name: "unnormalized", reason: "unnormalized"},
		{
			name:   "stale fingerprint",
			reason: "stale-successor",
			setup: func(t *testing.T, _, raw, sessionID string) {
				writeCompleteSessionSuccessor(t, sessionStoreClaude, sessionID, raw)
				writeAged(t, raw, "changed raw transcript", 30*24*time.Hour)
			},
		},
		{
			name:   "partial successor",
			reason: "partial-successor",
			setup: func(t *testing.T, _, raw, sessionID string) {
				fingerprint, err := fingerprintFile(raw)
				if err != nil {
					t.Fatal(err)
				}
				logs := []SessionLogLine{{Agent: sessionStoreClaude, SID: sessionID, Role: "user", Content: "partial"}}
				if _, err := ingestSession(context.Background(), sessionStoreClaude, sessionID, logs, sessionSource{Type: "claude-jsonl", Fingerprint: fingerprint, Completeness: sessionPartial}); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "interrupted ingestion",
			reason: "interrupted-ingestion",
			setup: func(t *testing.T, home, _, sessionID string) {
				lineage := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, sessionStoreClaude, sessionLineageID(sessionStoreClaude, sessionID), ".ingest-abandoned")
				if err := os.MkdirAll(lineage, 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "unreadable successor",
			reason: "unreadable-successor",
			setup: func(t *testing.T, home, _, sessionID string) {
				generation := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, sessionStoreClaude, sessionLineageID(sessionStoreClaude, sessionID), "corrupt")
				if err := os.MkdirAll(generation, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(generation, "manifest.json"), []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "ambiguous successor",
			reason: "ambiguous-successor",
			setup: func(t *testing.T, home, raw, sessionID string) {
				result := writeCompleteSessionSuccessor(t, sessionStoreClaude, sessionID, raw)
				lineage := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, sessionStoreClaude, result.LineageID)
				copySessionGeneration(t, filepath.Join(lineage, result.GenerationID), filepath.Join(lineage, "duplicate-generation"))
			},
		},
		{
			name:   "mismatched high water",
			reason: "unreadable-successor",
			setup: func(t *testing.T, home, raw, sessionID string) {
				result := writeCompleteSessionSuccessor(t, sessionStoreClaude, sessionID, raw)
				generation := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, sessionStoreClaude, result.LineageID, result.GenerationID)
				manifest, err := readSessionManifest(generation)
				if err != nil {
					t.Fatal(err)
				}
				manifest.HighWaterMark = "2099-01-01T00:00:00Z"
				content, err := json.Marshal(manifest)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(generation, "manifest.json"), content, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, out, home := newPruneTestState(t, &FakeRunner{})
			sessionID := strings.ReplaceAll(test.name, " ", "-")
			raw := rawSessionTestPath(home, sessionStoreClaude, sessionID)
			writeAged(t, raw, "raw transcript", 30*24*time.Hour)
			if test.setup != nil {
				test.setup(t, home, raw, sessionID)
			}
			state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/.claude/projects", Source: sessionStoreClaude, KeepDays: 7}}
			if err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7, DryRun: true}); err != nil {
				t.Fatalf("RunPrune returned an error: %v", err)
			}

			requireExists(t, raw)
			report := out.String()
			for _, field := range []string{"decision=retain", "source=claude", "lineage=", "age=", "size=", "reason=" + test.reason, "successor="} {
				if !strings.Contains(report, field) {
					t.Fatalf("expected %q in dry-run report: %s", field, report)
				}
			}
			out.Reset()
			if err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7}); err != nil {
				t.Fatalf("normal prune returned an error: %v", err)
			}
			requireExists(t, raw)
			if !strings.Contains(out.String(), "reason="+test.reason) {
				t.Fatalf("expected retained source warning: %s", out.String())
			}
		})
	}
}

func TestPruneRawSessionsRetainsSharedDatabases(t *testing.T) {
	for _, test := range []struct {
		source string
		path   string
	}{
		{source: sessionStoreOpenCode, path: "~/.local/share/opencode/opencode.db"},
		{source: sessionStoreCopilot, path: "~/.copilot/session-store.db"},
	} {
		t.Run(test.source, func(t *testing.T) {
			state, out, home := newPruneTestState(t, &FakeRunner{})
			path := ExpandPath(test.path)
			if !strings.HasPrefix(path, home) {
				t.Fatalf("expected test path under home, got %s", path)
			}
			writeAged(t, path, "shared database", 90*24*time.Hour)
			state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: test.path, Source: test.source, KeepDays: 7}}
			if err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7, DryRun: true}); err != nil {
				t.Fatalf("RunPrune returned an error: %v", err)
			}
			requireExists(t, path)
			if !strings.Contains(out.String(), "reason=ambiguous-shared-store") {
				t.Fatalf("expected shared-store retention evidence: %s", out.String())
			}
		})
	}
}

func TestPruneRawSessionsRefusesLinksAndInvalidSources(t *testing.T) {
	t.Run("linked root", func(t *testing.T) {
		state, _, home := newPruneTestState(t, &FakeRunner{})
		outside := filepath.Join(home, "outside")
		raw := filepath.Join(outside, "session.jsonl")
		writeAged(t, raw, "outside", 90*24*time.Hour)
		linked := filepath.Join(home, "linked-sessions")
		if err := os.Symlink(outside, linked); err != nil {
			t.Fatal(err)
		}
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: linked, Source: sessionStoreClaude, KeepDays: 7}}
		err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7})
		if err == nil || !strings.Contains(err.Error(), "refusing to traverse linked session store") {
			t.Fatalf("expected linked root to be rejected, got %v", err)
		}
		requireExists(t, raw)
	})

	t.Run("nested link", func(t *testing.T) {
		state, out, home := newPruneTestState(t, &FakeRunner{})
		root := rawSessionTestRoot(home, sessionStoreClaude)
		outside := filepath.Join(home, "outside.jsonl")
		writeAged(t, outside, "outside", 90*24*time.Hour)
		link := filepath.Join(root, "project", "linked.jsonl")
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, link); err != nil {
			t.Fatal(err)
		}
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: root, Source: sessionStoreClaude, KeepDays: 7}}
		if err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7, DryRun: true}); err != nil {
			t.Fatal(err)
		}
		requireExists(t, link)
		requireExists(t, outside)
		if !strings.Contains(out.String(), "reason=link-or-special-file") {
			t.Fatalf("expected linked-file retention evidence: %s", out.String())
		}
	})

	t.Run("unknown explicit source", func(t *testing.T) {
		state, _, home := newPruneTestState(t, &FakeRunner{})
		raw := filepath.Join(home, "custom", "old.jsonl")
		writeAged(t, raw, "old", 90*24*time.Hour)
		state.Config.Prune.Agents.Sessions = []PruneSessionStore{{Path: "~/custom", Source: "claud", KeepDays: 7}}
		err := RunPrune(context.Background(), state, PruneOptions{Targets: map[string]string{"agents": levelSessions}, Days: 7})
		if err == nil || !strings.Contains(err.Error(), `invalid session source "claud"`) {
			t.Fatalf("expected invalid source error, got %v", err)
		}
		requireExists(t, raw)
	})
}
