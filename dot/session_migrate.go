package dot

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var legacySessionNamePattern = regexp.MustCompile(`^[0-9]{6}_(agy|claude|codex|opencode|copilot)_(.+)\.jsonl$`)

type legacySessionCandidate struct {
	Agent       string
	SessionID   string
	Path        string
	Fingerprint string
	Logs        []SessionLogLine
	Malformed   int
	Bytes       int64
}

func readLegacySession(path, agent, sessionID string) (legacySessionCandidate, error) {
	file, err := os.Open(path) //nolint:gosec // path is discovered under the private legacy archive
	if err != nil {
		return legacySessionCandidate{}, err
	}
	defer func() { _ = file.Close() }()
	var logs []SessionLogLine
	malformed := 0
	reader := bufio.NewReader(file)
	for {
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return legacySessionCandidate{}, readErr
		}
		if len(line) == 0 && errors.Is(readErr, io.EOF) {
			break
		}
		var log SessionLogLine
		if decodeErr := json.Unmarshal(line, &log); decodeErr != nil {
			malformed++
		} else if log.Agent != agent || log.SID != sessionID {
			malformed++
		} else {
			logs = append(logs, log)
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
	}
	info, err := file.Stat()
	if err != nil {
		return legacySessionCandidate{}, err
	}
	fingerprint, err := fingerprintFile(path)
	if err != nil {
		return legacySessionCandidate{}, err
	}
	return legacySessionCandidate{Agent: agent, SessionID: sessionID, Path: path, Fingerprint: fingerprint, Logs: logs, Malformed: malformed, Bytes: info.Size()}, nil
}

func betterLegacySession(left, right legacySessionCandidate) bool {
	if len(left.Logs) != len(right.Logs) {
		return len(left.Logs) > len(right.Logs)
	}
	if left.Malformed != right.Malformed {
		return left.Malformed < right.Malformed
	}
	if left.Bytes != right.Bytes {
		return left.Bytes > right.Bytes
	}
	return left.Path < right.Path
}

func collectLegacySessions(root string) (map[string][]legacySessionCandidate, int, error) {
	lineages := make(map[string][]legacySessionCandidate)
	malformed := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && entry.Name() == sessionStoreVersion {
				return filepath.SkipDir
			}
			return nil
		}
		matches := legacySessionNamePattern.FindStringSubmatch(entry.Name())
		if len(matches) != 3 || !isValidSessionID(matches[2]) {
			if strings.HasSuffix(entry.Name(), ".jsonl") {
				malformed++
			}
			return nil
		}
		candidate, err := readLegacySession(path, matches[1], matches[2])
		if err != nil {
			return err
		}
		key := sessionLineageID(candidate.Agent, candidate.SessionID)
		lineages[key] = append(lineages[key], candidate)
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return lineages, malformed, nil
	}
	return lineages, malformed, err
}

func RunAgentSessionMigrate(ctx context.Context, state *GlobalState, apply bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	root := filepath.Join(home, ".agents", "sessions")
	lineages, malformedFiles, err := collectLegacySessions(root)
	if err != nil {
		return fmt.Errorf("failed to scan legacy session archive: %w", err)
	}
	keys := make([]string, 0, len(lineages))
	for key := range lineages {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	selected, duplicates, partial, skipped := 0, 0, 0, 0
	for _, key := range keys {
		candidates := lineages[key]
		slices.SortFunc(candidates, func(left, right legacySessionCandidate) int {
			if betterLegacySession(left, right) {
				return -1
			}
			if betterLegacySession(right, left) {
				return 1
			}
			return 0
		})
		best := candidates[0]
		selected++
		duplicates += len(candidates) - 1
		if best.Malformed > 0 {
			partial++
		}
		if len(best.Logs) == 0 {
			skipped++
		}
		lineage := key[:12]
		_, _ = fmt.Fprintf(state.Stdout, "migration: select lineage=%s records=%d candidates=%d malformed=%d\n", lineage, len(best.Logs), len(candidates), best.Malformed)
		if !apply {
			continue
		}
		result, ingestErr := ingestSession(ctx, best.Agent, best.SessionID, best.Logs, sessionSource{Type: "legacy-jsonl", Fingerprint: best.Fingerprint, Malformed: best.Malformed})
		if ingestErr != nil {
			return fmt.Errorf("failed to migrate lineage %s: %w", lineage, ingestErr)
		}
		reportSessionIngestion(state.Stdout, result)
	}
	mode := "dry-run"
	if apply {
		mode = "apply"
	}
	_, _ = fmt.Fprintf(state.Stdout, "migration: %s selected=%d duplicate=%d partial=%d skipped=%d malformed_files=%d legacy_preserved=true\n", mode, selected, duplicates, partial, skipped, malformedFiles)
	return nil
}
