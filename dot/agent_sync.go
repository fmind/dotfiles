package dot

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Whole-store sweeps behind `dot agent session sync`. The per-agent ingestion
// entrypoints in agent.go are driven one session at a time by hooks; this file is
// the batch counterpart that walks every store and replays them through the same
// loggers, so a machine that missed hook events can still be brought up to date.

// syncOutcome is one agent's contribution to a sync run.
type syncOutcome struct {
	// verb distinguishes a store dot walked file by file ("checked") from one it
	// swept and ingested wholesale ("ingested").
	verb  string
	count int
}

// reportSyncOutcome prints one agent's tally. A blank line precedes any non-empty
// tally so per-session output stays visually separated from the summary.
func reportSyncOutcome(state *GlobalState, agent string, outcome syncOutcome) {
	if outcome.count > 0 {
		_, _ = fmt.Fprintln(state.Stderr)
	}
	_, _ = fmt.Fprintf(state.Stderr, "%s: %d %s\n", agent, outcome.count, outcome.verb)
}

// syncAgySessions walks the agy brain directory, ingesting every session that has a
// readable transcript. Sessions without one are skipped, not failed: agy creates the
// directory before the transcript exists.
func syncAgySessions(ctx context.Context, state, sourceState *GlobalState, root string) (int, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0, fmt.Errorf("failed to read agy sessions: %w", err)
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionID := entry.Name()
		candidates, candidateErr := agyTranscriptCandidates(state.Config.Agent, sessionID)
		if candidateErr != nil {
			return 0, candidateErr
		}
		present := false
		for _, candidate := range candidates {
			if _, statErr := os.Stat(candidate); statErr == nil {
				present = true
				break
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return 0, fmt.Errorf("failed to inspect agy transcript %s: %w", candidate, statErr)
			}
		}
		if !present {
			continue
		}
		if logErr := RunAgentSessionLogAgy(ctx, sourceState, sessionID, ""); logErr != nil {
			return 0, fmt.Errorf("failed to sync agy session %s: %w", sessionID, logErr)
		}
		count++
	}
	return count, nil
}

// syncTranscriptSessions walks a JSONL transcript tree, ingesting every file whose
// name yields a session ID. The discovered path is handed to the per-agent logger as
// a synthetic hook payload so it never has to rediscover the file.
func syncTranscriptSessions(ctx context.Context, sourceState *GlobalState, root, label string, sessionIDOf func(name string) string, log agentSessionLogger) (int, error) {
	count := 0
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		sessionID := sessionIDOf(entry.Name())
		if sessionID == "" {
			return nil
		}
		transcriptState, stateErr := sessionStateWithTranscript(sourceState, sessionID, path)
		if stateErr != nil {
			return stateErr
		}
		if logErr := log(ctx, transcriptState, sessionID, ""); logErr != nil {
			return fmt.Errorf("failed to sync %s session %s: %w", label, sessionID, logErr)
		}
		count++
		return nil
	})
	if walkErr != nil {
		return 0, fmt.Errorf("failed to scan %s sessions: %w", label, walkErr)
	}
	return count, nil
}

// syncDatabaseSessions sweeps a SQLite-backed store when it exists. A missing store
// is not an error: that agent simply is not installed.
func syncDatabaseSessions(ctx context.Context, state *GlobalState, dbPath, label string, sweep func(context.Context, *GlobalState, string) (int, error)) (int, bool, error) {
	info, err := os.Stat(dbPath)
	if errors.Is(err, os.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to inspect %s database %s: %w", label, dbPath, err)
	}
	if info.IsDir() {
		return 0, false, fmt.Errorf("%s database path is a directory: %s", label, dbPath)
	}
	count, syncErr := sweep(ctx, state, dbPath)
	return count, true, syncErr
}

// RunAgentSessionSync scans all agent storage and triggers logging for unprocessed sessions.
func RunAgentSessionSync(ctx context.Context, state *GlobalState) error {
	cfg := state.Config.Agent

	// The per-agent log helpers double as hook entrypoints and parse state.Stdin.
	// When invoked from sync there is no hook payload, so detach stdin to avoid
	// consuming (and choking on) whatever the sync process was started with.
	noStdinState := *state
	noStdinState.Stdin = nil

	total := 0
	record := func(agent string, outcome syncOutcome) {
		reportSyncOutcome(state, agent, outcome)
		total += outcome.count
	}

	agyRoot, err := cfg.SourceRoot(sessionStoreAgy)
	if err != nil {
		return err
	}
	present, err := sourceDirectoryExists(agyRoot, sessionStoreAgy)
	if err != nil {
		return err
	}
	if present {
		count, syncErr := syncAgySessions(ctx, state, &noStdinState, agyRoot)
		if syncErr != nil {
			return syncErr
		}
		record(sessionStoreAgy, syncOutcome{verb: "checked", count: count})
	}

	claudeRoot, err := cfg.SourceRoot(sessionStoreClaude)
	if err != nil {
		return err
	}
	present, err = sourceDirectoryExists(claudeRoot, "Claude")
	if err != nil {
		return err
	}
	if present {
		count, syncErr := syncTranscriptSessions(ctx, &noStdinState, claudeRoot, "Claude", func(name string) string {
			// memory.jsonl is hand-curated long-term memory, not a session transcript.
			if name == "memory.jsonl" {
				return ""
			}
			return strings.TrimSuffix(name, ".jsonl")
		}, RunAgentSessionLogClaude)
		if syncErr != nil {
			return syncErr
		}
		record(sessionStoreClaude, syncOutcome{verb: "checked", count: count})
	}

	codexRoot, err := cfg.SourceRoot(sessionStoreCodex)
	if err != nil {
		return err
	}
	present, err = sourceDirectoryExists(codexRoot, "Codex")
	if err != nil {
		return err
	}
	if present {
		count, syncErr := syncTranscriptSessions(ctx, &noStdinState, codexRoot, "Codex", func(name string) string {
			return extractCodexSessionID(strings.TrimSuffix(name, ".jsonl"))
		}, RunAgentSessionLogCodex)
		if syncErr != nil {
			return syncErr
		}
		record(sessionStoreCodex, syncOutcome{verb: "checked", count: count})
	}

	opencodeDB, err := cfg.SourceRoot(sessionStoreOpenCode)
	if err != nil {
		return err
	}
	count, present, err := syncDatabaseSessions(ctx, state, opencodeDB, "OpenCode", syncOpencodeSessions)
	if err != nil {
		return err
	}
	if present {
		record(sessionStoreOpenCode, syncOutcome{verb: "ingested", count: count})
	}

	// A full idempotent backfill complements the targeted sessionEnd sync.
	copilotDB, err := cfg.SourceRoot(sessionStoreCopilot)
	if err != nil {
		return err
	}
	count, present, err = syncDatabaseSessions(ctx, state, copilotDB, "Copilot", syncCopilotSessions)
	if err != nil {
		return err
	}
	if present {
		record(sessionStoreCopilot, syncOutcome{verb: "ingested", count: count})
	}

	_, _ = fmt.Fprintf(state.Stderr, "agent-session-sync: done (%d total processed)\n", total)
	return nil
}
