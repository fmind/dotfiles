package dot

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	sessionSchemaVersion = 1
	// Bump whenever parser behavior can change normalized output for one source.
	sessionParserVersion = "1"
	sessionStoreVersion  = "v1"
)

type sessionCompleteness string

const (
	sessionComplete sessionCompleteness = "complete"
	sessionPartial  sessionCompleteness = "partial"
)

type sessionSource struct {
	Type         string
	Fingerprint  string
	HighWater    string
	Completeness sessionCompleteness
	Malformed    int
	Skipped      int
}

type sessionManifest struct {
	ParserVersion     string              `json:"parser_version"`
	Agent             string              `json:"agent"`
	SessionID         string              `json:"session_id"`
	LineageID         string              `json:"lineage_id"`
	SourceType        string              `json:"source_type"`
	SourceFingerprint string              `json:"source_fingerprint"`
	HighWaterMark     string              `json:"high_water_mark,omitempty"`
	IngestedAt        string              `json:"ingested_at"`
	Completeness      sessionCompleteness `json:"completeness"`
	TranscriptSHA256  string              `json:"transcript_sha256"`
	SchemaVersion     int                 `json:"schema_version"`
	RecordCount       int                 `json:"record_count"`
	MalformedRecords  int                 `json:"malformed_records"`
	SkippedRecords    int                 `json:"skipped_records"`
}

type sessionIngestionStatus string

const (
	sessionIngested  sessionIngestionStatus = "ingested"
	sessionDuplicate sessionIngestionStatus = "duplicate"
	sessionSkipped   sessionIngestionStatus = "skipped"
)

type sessionIngestionResult struct {
	Status       sessionIngestionStatus
	LineageID    string
	GenerationID string
	Manifest     sessionManifest
}

func sessionDigest(values ...string) string {
	hash := sha256.New()
	for _, value := range values {
		_, _ = io.WriteString(hash, value)
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func sessionLineageID(agent, sessionID string) string {
	return sessionDigest(agent, sessionID)
}

func sessionGenerationID(sourceFingerprint string) string {
	return sessionDigest(sessionParserVersion, sourceFingerprint)
}

func fingerprintBytes(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func fingerprintFile(path string) (string, error) {
	file, err := os.Open(path) //nolint:gosec // paths come from the owning agent's configured session store
	if err != nil {
		return "", fmt.Errorf("failed to fingerprint session source %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()
	fingerprint, err := fingerprintReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to fingerprint session source %s: %w", path, err)
	}
	return fingerprint, nil
}

func fingerprintReader(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func fingerprintLogs(logs []SessionLogLine) (string, error) {
	content, err := marshalSessionLogs(logs)
	if err != nil {
		return "", err
	}
	return fingerprintBytes(content), nil
}

func fingerprintJSON(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to fingerprint structured session source: %w", err)
	}
	return fingerprintBytes(content), nil
}

func sessionHighWater(logs []SessionLogLine) string {
	var highWater string
	for _, log := range logs {
		if log.TS > highWater {
			highWater = log.TS
		}
	}
	return highWater
}

func marshalSessionLogs(logs []SessionLogLine) ([]byte, error) {
	var content strings.Builder
	encoder := json.NewEncoder(&content)
	encoder.SetEscapeHTML(false)
	for _, log := range logs {
		if err := encoder.Encode(log); err != nil {
			return nil, fmt.Errorf("failed to encode session transcript: %w", err)
		}
	}
	return []byte(content.String()), nil
}

func normalizeSessionLogs(agent, sessionID string, logs []SessionLogLine) error {
	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}
	for index := range logs {
		if logs[index].Agent != agent || logs[index].SID != sessionID {
			return fmt.Errorf("session record %d does not match its lineage", index+1)
		}
	}
	var activeModel string
	for index := range logs {
		if logs[index].Model != "" {
			activeModel = logs[index].Model
		} else if activeModel != "" {
			logs[index].Model = activeModel
		}
	}
	for index := len(logs) - 1; index >= 0; index-- {
		if logs[index].Model != "" {
			activeModel = logs[index].Model
		} else if activeModel != "" {
			logs[index].Model = activeModel
		}
	}
	return nil
}

func sessionStoreRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agents", "sessions", sessionStoreVersion), nil
}

func secureSessionDirectories(paths ...string) error {
	for _, path := range paths {
		if err := os.MkdirAll(path, 0o700); err != nil {
			return fmt.Errorf("failed to create session directory %s: %w", path, err)
		}
		if err := os.Chmod(path, 0o700); err != nil {
			return fmt.Errorf("failed to secure session directory %s: %w", path, err)
		}
	}
	return nil
}

func writeOwnerOnly(path string, content []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func validateSessionGeneration(path string, expected sessionManifest) error {
	manifest, err := readSessionManifest(path)
	if err != nil {
		return err
	}
	if manifest != expected {
		return errors.New("session manifest did not round-trip")
	}
	transcript, err := os.ReadFile(filepath.Join(path, "transcript.jsonl")) //nolint:gosec // path is the private generation created by this process
	if err != nil {
		return err
	}
	if fingerprintBytes(transcript) != manifest.TranscriptSHA256 {
		return errors.New("session transcript fingerprint mismatch")
	}
	logs := make([]SessionLogLine, 0, manifest.RecordCount)
	decoder := json.NewDecoder(bytes.NewReader(transcript))
	for {
		var log SessionLogLine
		decodeErr := decoder.Decode(&log)
		if errors.Is(decodeErr, io.EOF) {
			break
		}
		if decodeErr != nil {
			return fmt.Errorf("invalid normalized transcript record: %w", decodeErr)
		}
		if log.Agent != manifest.Agent || log.SID != manifest.SessionID {
			return errors.New("normalized transcript record has mismatched lineage")
		}
		logs = append(logs, log)
	}
	if len(logs) != manifest.RecordCount {
		return fmt.Errorf("normalized transcript contains %d records, expected %d", len(logs), manifest.RecordCount)
	}
	if highWater := sessionHighWater(logs); highWater != manifest.HighWaterMark {
		return fmt.Errorf("normalized transcript high-water mark %q does not match manifest %q", highWater, manifest.HighWaterMark)
	}
	return nil
}

func readSessionManifest(path string) (sessionManifest, error) {
	manifestContent, err := os.ReadFile(filepath.Join(path, "manifest.json")) //nolint:gosec // path is a private generation beneath the versioned store
	if err != nil {
		return sessionManifest{}, err
	}
	var manifest sessionManifest
	if decodeErr := json.Unmarshal(manifestContent, &manifest); decodeErr != nil {
		return sessionManifest{}, decodeErr
	}
	return manifest, nil
}

func validateExistingSessionGeneration(path string, expected sessionManifest) error {
	manifest, err := readSessionManifest(path)
	if err != nil {
		return err
	}
	if manifest.SchemaVersion != expected.SchemaVersion || manifest.ParserVersion != expected.ParserVersion || manifest.Agent != expected.Agent || manifest.SessionID != expected.SessionID || manifest.LineageID != expected.LineageID || manifest.SourceFingerprint != expected.SourceFingerprint {
		return errors.New("existing session generation does not match its immutable identity")
	}
	return validateSessionGeneration(path, manifest)
}

func ingestSession(ctx context.Context, agent, sessionID string, logs []SessionLogLine, source sessionSource) (sessionIngestionResult, error) {
	lineageID := sessionLineageID(agent, sessionID)
	result := sessionIngestionResult{LineageID: lineageID}
	if len(logs) == 0 {
		result.Status = sessionSkipped
		completeness := source.Completeness
		if completeness == "" {
			completeness = sessionComplete
		}
		if source.Malformed > 0 {
			completeness = sessionPartial
		}
		skipped := source.Skipped
		if skipped == 0 {
			skipped = 1
		}
		result.Manifest = sessionManifest{Agent: agent, SessionID: sessionID, LineageID: lineageID, SourceType: source.Type, Completeness: completeness, MalformedRecords: source.Malformed, SkippedRecords: skipped}
		return result, nil
	}
	if err := normalizeSessionLogs(agent, sessionID, logs); err != nil {
		return result, err
	}
	if source.Fingerprint == "" {
		fingerprint, err := fingerprintLogs(logs)
		if err != nil {
			return result, err
		}
		source.Fingerprint = fingerprint
	}
	if _, err := hex.DecodeString(source.Fingerprint); err != nil || len(source.Fingerprint) != sha256.Size*2 {
		return result, errors.New("session source fingerprint must be a full SHA-256 digest")
	}
	if source.Type == "" {
		source.Type = "normalized"
	}
	if source.HighWater == "" {
		source.HighWater = sessionHighWater(logs)
	}
	if source.Completeness == "" {
		source.Completeness = sessionComplete
	}
	if source.Malformed > 0 {
		source.Completeness = sessionPartial
	}

	transcript, err := marshalSessionLogs(logs)
	if err != nil {
		return result, err
	}
	generationID := sessionGenerationID(source.Fingerprint)
	result.GenerationID = generationID
	manifest := sessionManifest{
		SchemaVersion:     sessionSchemaVersion,
		ParserVersion:     sessionParserVersion,
		Agent:             agent,
		SessionID:         sessionID,
		LineageID:         lineageID,
		SourceType:        source.Type,
		SourceFingerprint: source.Fingerprint,
		HighWaterMark:     source.HighWater,
		IngestedAt:        time.Now().UTC().Format(time.RFC3339Nano),
		Completeness:      source.Completeness,
		RecordCount:       len(logs),
		MalformedRecords:  source.Malformed,
		SkippedRecords:    source.Skipped,
		TranscriptSHA256:  fingerprintBytes(transcript),
	}
	result.Manifest = manifest

	root, err := sessionStoreRoot()
	if err != nil {
		return result, err
	}
	lineageDir := filepath.Join(root, agent, lineageID)
	if secureErr := secureSessionDirectories(root, filepath.Join(root, agent), lineageDir); secureErr != nil {
		return result, secureErr
	}
	finalDir := filepath.Join(lineageDir, generationID)
	if _, statErr := os.Stat(finalDir); statErr == nil {
		if validateErr := validateExistingSessionGeneration(finalDir, manifest); validateErr != nil {
			return result, fmt.Errorf("existing session generation is invalid: %w", validateErr)
		}
		result.Status = sessionDuplicate
		return result, nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return result, fmt.Errorf("failed to inspect session generation: %w", statErr)
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return result, contextErr
	}
	tempDir, err := os.MkdirTemp(lineageDir, ".ingest-")
	if err != nil {
		return result, fmt.Errorf("failed to create temporary session generation: %w", err)
	}
	// The path is created beneath the validated private lineage directory and never
	// follows user-provided links. A failed or losing concurrent writer owns it fully.
	defer func() { _ = os.RemoveAll(tempDir) }()
	if chmodErr := os.Chmod(tempDir, 0o700); chmodErr != nil {
		return result, chmodErr
	}
	if writeErr := writeOwnerOnly(filepath.Join(tempDir, "transcript.jsonl"), transcript); writeErr != nil {
		return result, fmt.Errorf("failed to write temporary session transcript: %w", writeErr)
	}
	manifestContent, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return result, err
	}
	manifestContent = append(manifestContent, '\n')
	if err := writeOwnerOnly(filepath.Join(tempDir, "manifest.json"), manifestContent); err != nil {
		return result, fmt.Errorf("failed to write temporary session manifest: %w", err)
	}
	if err := validateSessionGeneration(tempDir, manifest); err != nil {
		return result, fmt.Errorf("failed to validate temporary session generation: %w", err)
	}
	if err := os.Rename(tempDir, finalDir); err != nil {
		if _, statErr := os.Stat(finalDir); statErr == nil {
			if validateErr := validateExistingSessionGeneration(finalDir, manifest); validateErr != nil {
				return result, fmt.Errorf("concurrent session generation is invalid: %w", validateErr)
			}
			result.Status = sessionDuplicate
			return result, nil
		}
		return result, fmt.Errorf("failed to atomically publish session generation: %w", err)
	}
	result.Status = sessionIngested
	return result, nil
}

func reportSessionIngestion(output io.Writer, result sessionIngestionResult) {
	if output == nil {
		return
	}
	lineage := result.LineageID
	if len(lineage) > 12 {
		lineage = lineage[:12]
	}
	_, _ = fmt.Fprintf(output, "agent-session: %s lineage=%s records=%d malformed=%d skipped=%d completeness=%s\n", result.Status, lineage, result.Manifest.RecordCount, result.Manifest.MalformedRecords, result.Manifest.SkippedRecords, result.Manifest.Completeness)
}

func writeSessionLogs(ctx context.Context, state *GlobalState, agent, sessionID string, logs []SessionLogLine, source sessionSource) (sessionIngestionResult, error) {
	result, err := ingestSession(ctx, agent, sessionID, logs, source)
	if err != nil {
		return result, err
	}
	reportSessionIngestion(state.Stderr, result)
	return result, nil
}
