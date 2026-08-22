package dot

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type sessionSuccessorEvidence struct {
	Reason   string
	Evidence string
	Verified bool
}

func sessionStoreSource(store PruneSessionStore) string {
	if store.Source != "" {
		return store.Source
	}
	path := filepath.ToSlash(store.Path)
	switch {
	case strings.HasSuffix(path, "/.agents/sessions"):
		return sessionStoreArchive
	case strings.HasSuffix(path, "/.claude/projects"):
		return sessionStoreClaude
	case strings.HasSuffix(path, "/.codex/sessions"):
		return sessionStoreCodex
	case strings.HasSuffix(path, "/.gemini/antigravity-cli/brain"):
		return sessionStoreAgy
	case strings.HasSuffix(path, "/.grok/sessions"):
		return sessionStoreGrok
	case strings.HasSuffix(path, "/.local/share/opencode/opencode.db"):
		return sessionStoreOpenCode
	case strings.HasSuffix(path, "/.copilot/session-store.db"):
		return sessionStoreCopilot
	default:
		return ""
	}
}

func isKnownSessionStoreSource(source string) bool {
	switch source {
	case sessionStoreArchive, sessionStoreAgy, sessionStoreClaude, sessionStoreCodex, sessionStoreGrok, sessionStoreOpenCode, sessionStoreCopilot:
		return true
	default:
		return false
	}
}

func rawSessionIdentity(root, path, source string) (string, bool) {
	name := filepath.Base(path)
	switch source {
	case sessionStoreClaude:
		if !strings.HasSuffix(name, ".jsonl") || name == "memory.jsonl" {
			return "", false
		}
		sessionID := strings.TrimSuffix(name, ".jsonl")
		return sessionID, isValidSessionID(sessionID)
	case sessionStoreCodex:
		if !strings.HasSuffix(name, ".jsonl") {
			return "", false
		}
		sessionID := extractCodexSessionID(strings.TrimSuffix(name, ".jsonl"))
		return sessionID, isValidSessionID(sessionID)
	case sessionStoreGrok:
		// Grok names the session directory rather than the transcript, so identity
		// comes from the parent and every sibling stream is rejected by name.
		if name != grokTranscriptName {
			return "", false
		}
		sessionID := filepath.Base(filepath.Dir(path))
		return sessionID, isValidSessionID(sessionID)
	case sessionStoreAgy:
		if name != "transcript.jsonl" && name != "transcript_full.jsonl" {
			return "", false
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", false
		}
		parts := strings.Split(relative, string(filepath.Separator))
		if len(parts) < 2 || !isValidSessionID(parts[0]) {
			return "", false
		}
		return parts[0], true
	default:
		return "", false
	}
}

func truncatedDigest(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func findCompleteSessionSuccessor(agent, sessionID, sourceFingerprint string) (sessionSuccessorEvidence, error) {
	root, err := sessionStoreRoot()
	if err != nil {
		return sessionSuccessorEvidence{}, err
	}
	lineageID := sessionLineageID(agent, sessionID)
	lineageDir := filepath.Join(root, agent, lineageID)
	entries, err := os.ReadDir(lineageDir)
	if errors.Is(err, os.ErrNotExist) {
		return sessionSuccessorEvidence{Reason: "unnormalized", Evidence: "none"}, nil
	}
	if err != nil {
		return sessionSuccessorEvidence{Reason: "unreadable-successor", Evidence: "lineage=" + truncatedDigest(lineageID)}, nil
	}

	var (
		matches     []string
		interrupted bool
		partial     bool
		stale       bool
		unreadable  bool
	)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".ingest-") {
			interrupted = true
			continue
		}
		if !entry.IsDir() {
			unreadable = true
			continue
		}
		generationDir := filepath.Join(lineageDir, entry.Name())
		manifest, readErr := readSessionManifest(generationDir)
		if readErr != nil {
			unreadable = true
			continue
		}
		if manifest.Agent != agent || manifest.SessionID != sessionID || manifest.LineageID != lineageID {
			unreadable = true
			continue
		}
		if manifest.SourceFingerprint != sourceFingerprint || manifest.SchemaVersion != sessionSchemaVersion || manifest.ParserVersion != sessionParserVersion {
			stale = true
			continue
		}
		if manifest.Completeness != sessionComplete {
			partial = true
			continue
		}
		if validateErr := validateSessionGeneration(generationDir, manifest); validateErr != nil {
			unreadable = true
			continue
		}
		matches = append(matches, entry.Name())
	}
	if len(matches) == 1 {
		return sessionSuccessorEvidence{Reason: "verified-successor", Evidence: fmt.Sprintf("generation=%s fingerprint=%s completeness=complete", truncatedDigest(matches[0]), truncatedDigest(sourceFingerprint)), Verified: true}, nil
	}
	if len(matches) > 1 {
		return sessionSuccessorEvidence{Reason: "ambiguous-successor", Evidence: fmt.Sprintf("matches=%d fingerprint=%s", len(matches), truncatedDigest(sourceFingerprint))}, nil
	}
	switch {
	case partial:
		return sessionSuccessorEvidence{Reason: "partial-successor", Evidence: "fingerprint=" + truncatedDigest(sourceFingerprint)}, nil
	case unreadable:
		return sessionSuccessorEvidence{Reason: "unreadable-successor", Evidence: "lineage=" + truncatedDigest(lineageID)}, nil
	case interrupted:
		return sessionSuccessorEvidence{Reason: "interrupted-ingestion", Evidence: "lineage=" + truncatedDigest(lineageID)}, nil
	case stale:
		return sessionSuccessorEvidence{Reason: "stale-successor", Evidence: "fingerprint=" + truncatedDigest(sourceFingerprint)}, nil
	default:
		return sessionSuccessorEvidence{Reason: "unnormalized", Evidence: "none"}, nil
	}
}

// grokChatHistorySibling reports the sibling grokChatHistoryName file next to a
// verified Grok updates.jsonl, so it can be measured and removed alongside it.
func grokChatHistorySibling(root *os.Root, source, relative string) (string, int64, bool) {
	if source != sessionStoreGrok {
		return "", 0, false
	}
	siblingRelative := filepath.Join(filepath.Dir(relative), grokChatHistoryName)
	info, err := root.Lstat(siblingRelative)
	if err != nil || info.Mode()&os.ModeType != 0 {
		return "", 0, false
	}
	return siblingRelative, info.Size(), true
}

func sessionFileAge(now time.Time, info fs.FileInfo) time.Duration {
	age := now.Sub(info.ModTime())
	if age < 0 {
		return 0
	}
	return age.Truncate(time.Second)
}

func (r *pruneRun) reportSessionDecision(decision, source, lineage string, age time.Duration, size int64, reason, successor string) {
	// A normal prune stays quiet for expected retention, but surfaces anything that
	// prevented an aged source from being proved safe to delete. Dry-run inventories
	// every file decision, including successful proof and ordinary retention.
	if !r.dryRun && (decision == "delete" || reason == "within-retention" || reason == "protected-tree" || reason == "protected-name") {
		return
	}
	r.skipf("agents", "decision=%s source=%s lineage=%s age=%s size=%s reason=%s successor=%s", decision, source, lineage, age, humanBytes(size), reason, successor)
}

func (r *pruneRun) pruneRawSessions(root, source string, cutoff, now time.Time, keep []string) (int, int64, error) {
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return 0, 0, err
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 {
		return 0, 0, fmt.Errorf("refusing to traverse linked session store %s", root)
	}
	if source == sessionStoreOpenCode || source == sessionStoreCopilot {
		r.reportSessionDecision("retain", source, truncatedDigest(sessionDigest(root)), sessionFileAge(now, rootInfo), rootInfo.Size(), "ambiguous-shared-store", "none")
		return 0, 0, nil
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to confine session store %s: %w", root, err)
	}
	defer func() { _ = rootHandle.Close() }()

	var (
		emptied []string
		files   int
		total   int64
	)
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == root {
				return nil
			}
			if slices.Contains(keep, entry.Name()) {
				if r.dryRun {
					size, measureErr := dirBytes(path)
					if measureErr != nil {
						return measureErr
					}
					info, infoErr := entry.Info()
					if infoErr != nil {
						return infoErr
					}
					r.reportSessionDecision("retain", source, truncatedDigest(sessionDigest(path)), sessionFileAge(now, info), size, "protected-tree", "not-required")
				}
				return filepath.SkipDir
			}
			emptied = append(emptied, path)
			return nil
		}

		relative, relativeErr := filepath.Rel(root, path)
		if relativeErr != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("session source escaped configured root: %s", path)
		}
		if source == sessionStoreGrok && entry.Name() == grokChatHistoryName {
			// Carries no identity of its own; its disposal is decided alongside its
			// sibling updates.jsonl below, not flagged here as unrecognized.
			return nil
		}
		info, infoErr := rootHandle.Lstat(relative)
		if errors.Is(infoErr, os.ErrNotExist) {
			return nil
		}
		if infoErr != nil {
			return infoErr
		}
		lineageLabel := truncatedDigest(sessionDigest(path))
		age := sessionFileAge(now, info)
		if entry.Type()&os.ModeSymlink != 0 || info.Mode()&os.ModeType != 0 {
			r.reportSessionDecision("retain", source, lineageLabel, age, info.Size(), "link-or-special-file", "none")
			return nil
		}
		if slices.Contains(keep, entry.Name()) {
			r.reportSessionDecision("retain", source, lineageLabel, age, info.Size(), "protected-name", "not-required")
			return nil
		}
		if !info.ModTime().Before(cutoff) {
			r.reportSessionDecision("retain", source, lineageLabel, age, info.Size(), "within-retention", "not-required")
			return nil
		}
		sessionID, recognized := rawSessionIdentity(root, path, source)
		if !recognized {
			r.reportSessionDecision("retain", source, lineageLabel, age, info.Size(), "unrecognized-source", "none")
			return nil
		}
		lineageID := sessionLineageID(source, sessionID)
		file, openErr := rootHandle.Open(relative)
		if openErr != nil {
			r.reportSessionDecision("retain", source, truncatedDigest(lineageID), age, info.Size(), "unreadable-source", "none")
			return nil
		}
		fingerprint, fingerprintErr := fingerprintReader(file)
		closeErr := file.Close()
		if fingerprintErr != nil {
			r.reportSessionDecision("retain", source, truncatedDigest(lineageID), age, info.Size(), "unreadable-source", "none")
			return nil
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close raw session source: %w", closeErr)
		}
		evidence, evidenceErr := findCompleteSessionSuccessor(source, sessionID, fingerprint)
		if evidenceErr != nil {
			return evidenceErr
		}
		if !evidence.Verified {
			r.reportSessionDecision("retain", source, truncatedDigest(lineageID), age, info.Size(), evidence.Reason, evidence.Evidence)
			return nil
		}
		r.reportSessionDecision("delete", source, truncatedDigest(lineageID), age, info.Size(), evidence.Reason, evidence.Evidence)
		files++
		total += info.Size()
		siblingRelative, siblingSize, hasSibling := grokChatHistorySibling(rootHandle, source, relative)
		if hasSibling {
			files++
			total += siblingSize
		}
		if r.dryRun {
			return nil
		}
		// WalkDir does not follow directory symlinks, and only a recognized regular
		// source beneath the validated root reaches this mutation.
		if removeErr := rootHandle.Remove(relative); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("failed to remove verified raw session source: %w", removeErr)
		}
		if hasSibling {
			if removeErr := rootHandle.Remove(siblingRelative); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				return fmt.Errorf("failed to remove grok raw wire-format sibling: %w", removeErr)
			}
		}
		return nil
	})
	if walkErr != nil {
		return 0, 0, fmt.Errorf("failed to scan %s: %w", root, walkErr)
	}
	r.freed += total
	if !r.dryRun {
		if err := removeEmptyDirs(emptied); err != nil {
			return files, total, err
		}
	}
	return files, total, nil
}
