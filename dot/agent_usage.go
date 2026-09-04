package dot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v3"
)

// UsageRecord represents the normalized token usage and cost metrics for one agent session.
type UsageRecord struct {
	Timestamp        string  `json:"timestamp"`
	Harness          string  `json:"harness"`
	Agent            string  `json:"agent"`
	SessionID        string  `json:"session_id"`
	Model            string  `json:"model,omitempty"`
	CWD              string  `json:"cwd,omitempty"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CachedTokens     int64   `json:"cached_tokens,omitempty"`
	CacheWriteTokens int64   `json:"cache_write_tokens,omitempty"`
	ReasoningTokens  int64   `json:"reasoning_tokens,omitempty"`
	TotalTokens      int64   `json:"total_tokens"`
	CostUSD          float64 `json:"cost_usd,omitempty"`
	TurnCount        int     `json:"turn_count,omitempty"`
}

// UsageStatsRow represents aggregated usage stats across a harness or model.
type UsageStatsRow struct {
	Harness          string  `json:"harness"`
	Model            string  `json:"model,omitempty"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CachedTokens     int64   `json:"cached_tokens"`
	CacheWriteTokens int64   `json:"cache_write_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	Sessions         int     `json:"sessions"`
	Turns            int     `json:"turns"`
}

// UsageStatsOptions controls filtering and presentation of usage stats.
type UsageStatsOptions struct {
	Harness string
	Since   string
	Until   string
	ByModel bool
	JSON    bool
}

// UsageListOptions controls listing of individual session usage records.
type UsageListOptions struct {
	Harness string
	Limit   int
	JSON    bool
}

// UsageRoot returns the base directory for token usage records (~/.agents/usages).
func UsageRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agents", "usages"), nil
}

// HarnessUsageDir returns the directory for a specific harness's usage records.
func HarnessUsageDir(harness string) (string, error) {
	root, err := UsageRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, harness), nil
}

// WriteUsageRecord writes a usage record atomically to ~/.agents/usages/<harness>/<session_id>.json.
func WriteUsageRecord(record UsageRecord) error {
	if record.Harness == "" {
		return errors.New("missing harness in usage record")
	}
	if record.SessionID == "" {
		return errors.New("missing session_id in usage record")
	}
	if record.Agent == "" {
		record.Agent = record.Harness
	}
	if record.Timestamp == "" {
		record.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if record.TotalTokens == 0 && (record.InputTokens > 0 || record.OutputTokens > 0) {
		record.TotalTokens = record.InputTokens + record.OutputTokens + record.CachedTokens + record.CacheWriteTokens
	}

	dir, dirErr := HarnessUsageDir(record.Harness)
	if dirErr != nil {
		return dirErr
	}
	if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
		return fmt.Errorf("failed to create usage directory %s: %w", dir, mkErr)
	}

	data, marshalErr := json.MarshalIndent(record, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal usage record: %w", marshalErr)
	}
	data = append(data, '\n')

	safeSession := sanitizeFilename(record.SessionID)
	target := filepath.Join(dir, safeSession+".json")
	// Stop and SessionEnd hooks can fire close enough together to overlap on one
	// session, so the publish must tolerate a concurrent writer.
	if writeErr := publishOwnerOnly(target, data); writeErr != nil {
		return fmt.Errorf("failed to write usage record %s: %w", target, writeErr)
	}
	return nil
}

// recordUsageBestEffort refreshes the usage record for a session that has just
// been logged.
//
// It is deliberately non-fatal: `dot agent hook usage <agent>` is the owner of
// usage records and every deployed hook config runs it alongside the session
// hook, so a session must still be ingested when a token count cannot be
// derived. The failure is reported rather than discarded -- the six copies of
// this block that preceded it dropped both the extraction and the write error
// with no explanation, which hid the fixed-temp-file collision for months.
func recordUsageBestEffort(state *GlobalState, agent string, extract func() (*UsageRecord, error)) {
	record, err := extract()
	if err == nil && record != nil {
		err = WriteUsageRecord(*record)
	}
	if err != nil {
		_, _ = fmt.Fprintf(state.Stderr, "%s: usage not recorded for this session: %v\n", agent, err)
	}
}

func sanitizeFilename(s string) string {
	var buf strings.Builder
	for _, r := range s {
		if isValidSessionRune(r) {
			buf.WriteRune(r)
		} else {
			buf.WriteByte('_')
		}
	}
	return buf.String()
}

// ExtractUsageClaude extracts precise token usage from Claude Code's project transcript.
func ExtractUsageClaude(cfg AgentConfig, sessionID, cwd, transcriptPath string) (*UsageRecord, error) {
	if transcriptPath == "" {
		projectsDir, rootErr := cfg.SourceRoot(sessionStoreClaude)
		if rootErr != nil {
			return nil, rootErr
		}
		if cwd != "" {
			candidate := filepath.Join(projectsDir, claudeProjectDirectory(cwd), sessionID+".jsonl")
			if _, statErr := os.Stat(candidate); statErr == nil {
				transcriptPath = candidate
			}
		}
		if transcriptPath == "" {
			found, findErr := findSessionFile(projectsDir, func(_ string, entry fs.DirEntry) bool {
				return entry.Name() == sessionID+".jsonl"
			})
			if findErr != nil {
				return nil, fmt.Errorf("failed to search Claude transcripts in %s: %w", projectsDir, findErr)
			}
			if found == "" {
				return nil, fmt.Errorf("session file not found for claude session %s", sessionID)
			}
			transcriptPath = found
		}
	}
	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	rec := &UsageRecord{
		Harness:   sessionStoreClaude,
		Agent:     sessionStoreClaude,
		SessionID: sessionID,
		CWD:       cwd,
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		if ts, ok := raw["timestamp"].(string); ok && ts != "" {
			rec.Timestamp = ts
		}
		if lineCWD, ok := raw["cwd"].(string); ok && lineCWD != "" && rec.CWD == "" {
			rec.CWD = resolveCWD(lineCWD)
		}
		typ, _ := raw["type"].(string)
		if typ == "cost-state" {
			if cost, ok := raw["totalCostUSD"].(float64); ok && cost > 0 {
				rec.CostUSD = cost
			}
		}
		if typ == "assistant" {
			rec.TurnCount++
			if msg, ok := raw["message"].(map[string]any); ok {
				if model, ok := msg["model"].(string); ok && model != "" {
					rec.Model = model
				}
				if usage, ok := msg["usage"].(map[string]any); ok {
					if in, ok := usage["input_tokens"].(float64); ok {
						rec.InputTokens += int64(in)
					}
					if out, ok := usage["output_tokens"].(float64); ok {
						rec.OutputTokens += int64(out)
					}
					if cr, ok := usage["cache_read_input_tokens"].(float64); ok {
						rec.CachedTokens += int64(cr)
					}
					if cw, ok := usage["cache_creation_input_tokens"].(float64); ok {
						rec.CacheWriteTokens += int64(cw)
					}
				}
			}
		}
	}
	rec.TotalTokens = rec.InputTokens + rec.OutputTokens + rec.CachedTokens + rec.CacheWriteTokens
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	return rec, scanner.Err()
}

// ExtractUsageCodex extracts token usage from OpenAI Codex's rollout file.
func ExtractUsageCodex(cfg AgentConfig, sessionID, cwd, transcriptPath string) (*UsageRecord, error) {
	if transcriptPath == "" {
		sessionsDir, rootErr := cfg.SourceRoot(sessionStoreCodex)
		if rootErr != nil {
			return nil, rootErr
		}
		found, findErr := findSessionFile(sessionsDir, func(_ string, entry fs.DirEntry) bool {
			if !strings.HasSuffix(entry.Name(), ".jsonl") {
				return false
			}
			return extractCodexSessionID(strings.TrimSuffix(entry.Name(), ".jsonl")) == sessionID
		})
		if findErr != nil {
			return nil, fmt.Errorf("failed to search Codex transcripts in %s: %w", sessionsDir, findErr)
		}
		if found == "" {
			return nil, fmt.Errorf("session file not found for codex session %s", sessionID)
		}
		transcriptPath = found
	}
	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	rec := &UsageRecord{
		Harness:   sessionStoreCodex,
		Agent:     sessionStoreCodex,
		SessionID: sessionID,
		CWD:       cwd,
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		if ts, ok := raw["timestamp"].(string); ok && ts != "" {
			rec.Timestamp = ts
		}
		typ, _ := raw["type"].(string)
		if typ == "turn_context" {
			if p, ok := raw["payload"].(map[string]any); ok {
				if m, ok := p["model"].(string); ok && m != "" {
					rec.Model = m
				}
				if c, ok := p["cwd"].(string); ok && c != "" && rec.CWD == "" {
					rec.CWD = resolveCWD(c)
				}
			}
		}
		if typ == "session_meta" {
			if p, ok := raw["payload"].(map[string]any); ok {
				if c, ok := p["cwd"].(string); ok && c != "" && rec.CWD == "" {
					rec.CWD = resolveCWD(c)
				}
			}
		}
		if typ == "response_item" {
			if p, ok := raw["payload"].(map[string]any); ok {
				if role, ok := p["role"].(string); ok && role == "assistant" {
					rec.TurnCount++
				}
			}
		}
		if typ == "event_msg" {
			if p, ok := raw["payload"].(map[string]any); ok {
				if eventType, ok := p["type"].(string); ok && eventType == "token_count" {
					if info, ok := p["info"].(map[string]any); ok {
						if total, ok := info["total_token_usage"].(map[string]any); ok {
							if in, ok := total["input_tokens"].(float64); ok {
								rec.InputTokens = int64(in)
							}
							if out, ok := total["output_tokens"].(float64); ok {
								rec.OutputTokens = int64(out)
							}
							if ci, ok := total["cached_input_tokens"].(float64); ok {
								rec.CachedTokens = int64(ci)
							}
							if cw, ok := total["cache_write_input_tokens"].(float64); ok {
								rec.CacheWriteTokens = int64(cw)
							}
							if ro, ok := total["reasoning_output_tokens"].(float64); ok {
								rec.ReasoningTokens = int64(ro)
							}
							if tt, ok := total["total_tokens"].(float64); ok {
								rec.TotalTokens = int64(tt)
							}
						}
					}
				}
			}
		}
	}
	if rec.TotalTokens == 0 && (rec.InputTokens > 0 || rec.OutputTokens > 0) {
		rec.TotalTokens = rec.InputTokens + rec.OutputTokens + rec.CachedTokens + rec.CacheWriteTokens
	}
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	return rec, scanner.Err()
}

// ExtractUsageOpencode extracts token usage from OpenCode's SQLite database.
func ExtractUsageOpencode(ctx context.Context, state *GlobalState, sessionID, cwd string) (*UsageRecord, error) {
	dbPath, err := state.Config.Agent.SourceRoot(sessionStoreOpenCode)
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("opencode database not found at %s", dbPath)
	}
	query := fmt.Sprintf("SELECT model, tokens_input, tokens_output, tokens_reasoning, tokens_cache_read, tokens_cache_write, cost, directory, time_created FROM session WHERE id = '%s';", sessionID)
	out, err := runSQLiteJSON(ctx, state, dbPath, query)
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		return nil, fmt.Errorf("session %s not found in opencode database", sessionID)
	}
	var rows []struct {
		Model            any     `json:"model"`
		Directory        string  `json:"directory"`
		TokensInput      int64   `json:"tokens_input"`
		TokensOutput     int64   `json:"tokens_output"`
		TokensReasoning  int64   `json:"tokens_reasoning"`
		TokensCacheRead  int64   `json:"tokens_cache_read"`
		TokensCacheWrite int64   `json:"tokens_cache_write"`
		Cost             float64 `json:"cost"`
		TimeCreated      int64   `json:"time_created"`
	}
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, fmt.Errorf("failed to parse opencode session row: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("session %s not found in opencode database", sessionID)
	}
	row := rows[0]
	modelStr := ""
	switch m := row.Model.(type) {
	case string:
		var modelObj struct {
			ID         string `json:"id"`
			ModelID    string `json:"modelID"`
			ProviderID string `json:"providerID"`
		}
		if json.Unmarshal([]byte(m), &modelObj) == nil && (modelObj.ID != "" || modelObj.ModelID != "") {
			if modelObj.ProviderID != "" && modelObj.ModelID != "" {
				modelStr = modelObj.ProviderID + "/" + modelObj.ModelID
			} else if modelObj.ID != "" {
				modelStr = modelObj.ID
			}
		} else {
			modelStr = m
		}
	case map[string]any:
		prov, _ := m["providerID"].(string)
		mod, _ := m["modelID"].(string)
		id, _ := m["id"].(string)
		if prov != "" && mod != "" {
			modelStr = prov + "/" + mod
		} else if id != "" {
			modelStr = id
		}
	}
	sessionCWD := cwd
	if sessionCWD == "" {
		sessionCWD = row.Directory
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	if row.TimeCreated > 0 {
		ts = time.UnixMilli(row.TimeCreated).UTC().Format(time.RFC3339)
	}
	total := row.TokensInput + row.TokensOutput + row.TokensCacheRead + row.TokensCacheWrite
	return &UsageRecord{
		Timestamp:        ts,
		Harness:          sessionStoreOpenCode,
		Agent:            sessionStoreOpenCode,
		SessionID:        sessionID,
		Model:            modelStr,
		InputTokens:      row.TokensInput,
		OutputTokens:     row.TokensOutput,
		CachedTokens:     row.TokensCacheRead,
		CacheWriteTokens: row.TokensCacheWrite,
		ReasoningTokens:  row.TokensReasoning,
		TotalTokens:      total,
		CostUSD:          row.Cost,
		CWD:              resolveCWD(sessionCWD),
	}, nil
}

// ExtractUsageCopilot extracts token usage from GitHub Copilot CLI's assistant_usage_events.
func ExtractUsageCopilot(ctx context.Context, state *GlobalState, sessionID, cwd string) (*UsageRecord, error) {
	// The session id is interpolated into the SQL below, and the SessionEnd hook
	// takes it straight from the host's JSON payload. Reject anything outside the
	// id alphabet here, at the boundary every caller shares.
	if !isValidSessionID(sessionID) {
		return nil, fmt.Errorf("invalid copilot session id %q", sessionID)
	}
	dbPath, err := state.Config.Agent.SourceRoot(sessionStoreCopilot)
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("copilot database not found at %s", dbPath)
	}

	query := fmt.Sprintf("SELECT model, COALESCE(SUM(input_tokens), 0) AS input_tokens, COALESCE(SUM(output_tokens), 0) AS output_tokens, COALESCE(SUM(cache_read_tokens), 0) AS cache_read_tokens, COALESCE(SUM(cache_write_tokens), 0) AS cache_write_tokens, COALESCE(SUM(reasoning_tokens), 0) AS reasoning_tokens, COUNT(*) AS turns FROM assistant_usage_events WHERE session_id = '%s' GROUP BY model;", sessionID)
	out, err := runSQLiteJSON(ctx, state, dbPath, query)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		Model            string `json:"model"`
		InputTokens      int64  `json:"input_tokens"`
		OutputTokens     int64  `json:"output_tokens"`
		CacheReadTokens  int64  `json:"cache_read_tokens"`
		CacheWriteTokens int64  `json:"cache_write_tokens"`
		ReasoningTokens  int64  `json:"reasoning_tokens"`
		Turns            int    `json:"turns"`
	}
	_ = json.Unmarshal([]byte(strings.TrimSpace(out)), &rows)

	sessionQuery := fmt.Sprintf("SELECT cwd, created_at FROM sessions WHERE id = '%s';", sessionID)
	sessionOut, _ := runSQLiteJSON(ctx, state, dbPath, sessionQuery)
	var sessionRows []struct {
		CWD       string `json:"cwd"`
		CreatedAt string `json:"created_at"`
	}
	_ = json.Unmarshal([]byte(strings.TrimSpace(sessionOut)), &sessionRows)

	sessionCWD := cwd
	ts := time.Now().UTC().Format(time.RFC3339)
	if len(sessionRows) > 0 {
		if sessionCWD == "" {
			sessionCWD = sessionRows[0].CWD
		}
		if sessionRows[0].CreatedAt != "" {
			ts = sessionRows[0].CreatedAt
		}
	}

	rec := &UsageRecord{
		Timestamp: ts,
		Harness:   sessionStoreCopilot,
		Agent:     sessionStoreCopilot,
		SessionID: sessionID,
		CWD:       resolveCWD(sessionCWD),
	}
	for _, row := range rows {
		if rec.Model == "" && row.Model != "" {
			rec.Model = row.Model
		}
		rec.InputTokens += row.InputTokens
		rec.OutputTokens += row.OutputTokens
		rec.CachedTokens += row.CacheReadTokens
		rec.CacheWriteTokens += row.CacheWriteTokens
		rec.ReasoningTokens += row.ReasoningTokens
		rec.TurnCount += row.Turns
	}
	rec.TotalTokens = rec.InputTokens + rec.OutputTokens + rec.CachedTokens + rec.CacheWriteTokens
	return rec, nil
}

// ExtractUsageGrok extracts token usage from Grok Build's signals.json.
func ExtractUsageGrok(state *GlobalState, sessionID, cwd string) (*UsageRecord, error) {
	root, err := state.Config.Agent.SourceRoot(sessionStoreGrok)
	if err != nil {
		return nil, err
	}
	var sessionDir string
	if cwd != "" {
		candidate := filepath.Join(root, grokSessionDirectory(cwd), sessionID)
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			sessionDir = candidate
		}
	}
	if sessionDir == "" {
		entries, readErr := os.ReadDir(root)
		if readErr == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					candidate := filepath.Join(root, entry.Name(), sessionID)
					if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
						sessionDir = candidate
						break
					}
				}
			}
		}
	}
	if sessionDir == "" {
		return nil, fmt.Errorf("grok session directory not found for %s", sessionID)
	}

	rec := &UsageRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Harness:   sessionStoreGrok,
		Agent:     sessionStoreGrok,
		SessionID: sessionID,
		CWD:       cwd,
	}

	signalsPath := filepath.Join(sessionDir, "signals.json")
	if data, readErr := os.ReadFile(signalsPath); readErr == nil {
		var sig struct {
			PrimaryModelID      string `json:"primaryModelId"`
			ContextTokensUsed   int64  `json:"contextTokensUsed"`
			ContextWindowTokens int64  `json:"contextWindowTokens"`
			TurnCount           int    `json:"turnCount"`
		}
		if json.Unmarshal(data, &sig) == nil {
			// Grok records no cumulative token counts anywhere in a session
			// directory -- signals.json only reports how full the context window
			// ended up. That is an input-side measurement, so it goes in
			// InputTokens and TotalTokens keeps the sum-of-parts definition every
			// other harness uses; assigning occupancy straight to TotalTokens made
			// Grok rows self-contradictory and the cross-harness total meaningless.
			// The consequence is a documented undercount: Grok output tokens are
			// not observable.
			rec.InputTokens = sig.ContextTokensUsed
			rec.TotalTokens = rec.InputTokens + rec.OutputTokens + rec.CachedTokens + rec.CacheWriteTokens
			rec.Model = sig.PrimaryModelID
			rec.TurnCount = sig.TurnCount
		}
	}
	return rec, nil
}

// ExtractUsageAgy extracts token usage from Antigravity's transcript steps.
func ExtractUsageAgy(state *GlobalState, sessionID, cwd, transcriptPath string) (*UsageRecord, error) {
	if transcriptPath == "" {
		candidates, err := agyTranscriptCandidates(state.Config.Agent, sessionID)
		if err != nil {
			return nil, err
		}
		for _, c := range candidates {
			if _, statErr := os.Stat(c); statErr == nil {
				transcriptPath = c
				break
			}
		}
		if transcriptPath == "" {
			return nil, fmt.Errorf("transcript file not found for agy session %s", sessionID)
		}
	}
	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	rec := &UsageRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Harness:   sessionStoreAgy,
		Agent:     sessionStoreAgy,
		SessionID: sessionID,
		Model:     "gemini",
		CWD:       cwd,
	}

	var inputBytes int64
	var outputBytes int64

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		if ts, ok := raw["created_at"].(string); ok && ts != "" {
			rec.Timestamp = ts
		}
		source, _ := raw["source"].(string)
		typ, _ := raw["type"].(string)
		content, _ := raw["content"].(string)

		if source == "USER_EXPLICIT" && typ == "USER_INPUT" {
			rec.TurnCount++
			inputBytes += int64(len(content))
		} else if source == "MODEL" && typ == "PLANNER_RESPONSE" {
			outputBytes += int64(len(content))
			if thinking, ok := raw["thinking"].(string); ok {
				outputBytes += int64(len(thinking))
			}
		} else if typ == "RUN_COMMAND" || typ == "SYSTEM_MESSAGE" {
			inputBytes += int64(len(content))
		}
	}

	rec.InputTokens = (inputBytes + 3) / 4
	rec.OutputTokens = (outputBytes + 3) / 4
	rec.TotalTokens = rec.InputTokens + rec.OutputTokens
	return rec, scanner.Err()
}

// RunAgentHookUsage routes an agent usage hook call to the appropriate extractor and saves the record.
func RunAgentHookUsage(ctx context.Context, state *GlobalState, agent, sessionID, cwd string) error {
	var err error
	switch agent {
	case sessionStoreAgy:
		err = handleAgyUsageHook(state, sessionID, cwd)
	case sessionStoreClaude:
		err = handleClaudeUsageHook(state, sessionID, cwd)
	case sessionStoreCodex:
		err = handleCodexUsageHook(state, sessionID, cwd)
	case sessionStoreGrok:
		err = handleGrokUsageHook(state, sessionID, cwd)
	case sessionStoreOpenCode:
		err = handleOpenCodeUsageHook(ctx, state, sessionID, cwd)
	case sessionStoreCopilot:
		err = handleCopilotUsageHook(ctx, state, sessionID, cwd)
	default:
		err = fmt.Errorf("unknown usage hook agent %q", agent)
	}
	return spoolHookFailure(state.Config.Agent, agent, "usage", sessionID, err)
}

func handleAgyUsageHook(state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, true)
	if err != nil {
		return err
	}
	if halt {
		return writeAntigravityStopDecision(state.Stdout)
	}
	sessionID, cwd = identity.Session, identity.CWD
	transcriptPath, err := hookTranscriptPath(identity, "antigravity")
	if err != nil {
		return err
	}
	rec, err := ExtractUsageAgy(state, sessionID, cwd, transcriptPath)
	if err != nil {
		if identity.FromHook() {
			_ = writeAntigravityStopDecision(state.Stdout)
		}
		return err
	}
	if err := WriteUsageRecord(*rec); err != nil {
		if identity.FromHook() {
			_ = writeAntigravityStopDecision(state.Stdout)
		}
		return err
	}
	if identity.FromHook() {
		return writeAntigravityStopDecision(state.Stdout)
	}
	return nil
}

func handleClaudeUsageHook(state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, false)
	if err != nil {
		return err
	}
	if halt {
		return nil
	}
	sessionID, cwd = identity.Session, identity.CWD
	transcriptPath, err := hookTranscriptPath(identity, "claude")
	if err != nil {
		return err
	}
	rec, err := ExtractUsageClaude(state.Config.Agent, sessionID, cwd, transcriptPath)
	if err != nil {
		return err
	}
	return WriteUsageRecord(*rec)
}

func handleCodexUsageHook(state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, false)
	if err != nil {
		return err
	}
	if halt {
		return nil
	}
	sessionID, cwd = identity.Session, identity.CWD
	transcriptPath, err := hookTranscriptPath(identity, "codex")
	if err != nil {
		return err
	}
	rec, err := ExtractUsageCodex(state.Config.Agent, sessionID, cwd, transcriptPath)
	if err != nil {
		return err
	}
	return WriteUsageRecord(*rec)
}

func handleGrokUsageHook(state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, false)
	if err != nil {
		return err
	}
	if halt {
		return nil
	}
	sessionID, cwd = identity.Session, identity.CWD
	rec, err := ExtractUsageGrok(state, sessionID, cwd)
	if err != nil {
		return err
	}
	return WriteUsageRecord(*rec)
}

func handleOpenCodeUsageHook(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, false)
	if err != nil {
		return err
	}
	if halt {
		return nil
	}
	sessionID, cwd = identity.Session, identity.CWD
	rec, err := ExtractUsageOpencode(ctx, state, sessionID, cwd)
	if err != nil {
		return err
	}
	return WriteUsageRecord(*rec)
}

func handleCopilotUsageHook(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	defer func() {
		_, _ = fmt.Fprintln(state.Stdout, "{}")
	}()
	if sessionID == "" && state.Stdin != nil {
		input, err := decodeCopilotSessionEnd(state.Stdin)
		if err == nil {
			sessionID = input.SessionID
			cwd = input.CWD
		}
	}
	if sessionID == "" {
		return errors.New("missing copilot session id")
	}
	rec, err := ExtractUsageCopilot(ctx, state, sessionID, cwd)
	if err != nil {
		return err
	}
	return WriteUsageRecord(*rec)
}

// LoadAllUsageRecords reads all usage records from ~/.agents/usages.
func LoadAllUsageRecords() ([]UsageRecord, error) {
	root, err := UsageRoot()
	if err != nil {
		return nil, err
	}
	var records []UsageRecord
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, os.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || strings.HasSuffix(entry.Name(), ".tmp") {
			return nil
		}
		//nolint:gosec // G122: the walked root is the owner-only usages directory, not an attacker-writable path
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		var rec UsageRecord
		if json.Unmarshal(data, &rec) == nil {
			records = append(records, rec)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// RunAgentUsageStats computes aggregated usage stats per harness or model.
func RunAgentUsageStats(ctx context.Context, state *GlobalState, opts UsageStatsOptions) error {
	records, err := LoadAllUsageRecords()
	if err != nil {
		return fmt.Errorf("failed to load usage records: %w", err)
	}

	var sinceTime, untilTime time.Time
	if opts.Since != "" {
		if sinceTime, err = parseFlexibleTime(opts.Since); err != nil {
			return fmt.Errorf("--since: %w", err)
		}
	}
	if opts.Until != "" {
		if untilTime, err = parseFlexibleTime(opts.Until); err != nil {
			return fmt.Errorf("--until: %w", err)
		}
	}

	grouped := make(map[string]*UsageStatsRow)
	for _, rec := range records {
		if opts.Harness != "" && rec.Harness != opts.Harness && rec.Agent != opts.Harness {
			continue
		}
		if !sinceTime.IsZero() {
			t, parseErr := time.Parse(time.RFC3339, rec.Timestamp)
			if parseErr == nil && t.Before(sinceTime) {
				continue
			}
		}
		if !untilTime.IsZero() {
			t, parseErr := time.Parse(time.RFC3339, rec.Timestamp)
			if parseErr == nil && t.After(untilTime) {
				continue
			}
		}

		key := rec.Harness
		if opts.ByModel {
			model := rec.Model
			if model == "" {
				model = "unknown"
			}
			key = rec.Harness + "/" + model
		}

		row, ok := grouped[key]
		if !ok {
			row = &UsageStatsRow{
				Harness: rec.Harness,
			}
			if opts.ByModel {
				row.Model = rec.Model
			}
			grouped[key] = row
		}
		row.Sessions++
		row.Turns += rec.TurnCount
		row.InputTokens += rec.InputTokens
		row.OutputTokens += rec.OutputTokens
		row.CachedTokens += rec.CachedTokens
		row.CacheWriteTokens += rec.CacheWriteTokens
		row.ReasoningTokens += rec.ReasoningTokens
		row.TotalTokens += rec.TotalTokens
		row.CostUSD += rec.CostUSD
	}

	var rows []UsageStatsRow
	for _, r := range grouped {
		rows = append(rows, *r)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Harness != rows[j].Harness {
			return rows[i].Harness < rows[j].Harness
		}
		return rows[i].Model < rows[j].Model
	})

	if opts.JSON {
		encoder := json.NewEncoder(state.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(rows)
	}

	if len(rows) == 0 {
		_, _ = fmt.Fprintln(state.Stdout, "No usage records found in ~/.agents/usages. Run 'dot agent usage sync' to backfill existing sessions.")
		return nil
	}

	w := tabwriter.NewWriter(state.Stdout, 0, 0, 3, ' ', 0)
	if opts.ByModel {
		_, _ = fmt.Fprintln(w, "HARNESS\tMODEL\tSESSIONS\tTURNS\tINPUT TOKENS\tOUTPUT TOKENS\tCACHED TOKENS\tREASONING\tTOTAL TOKENS\tCOST (USD)")
	} else {
		_, _ = fmt.Fprintln(w, "HARNESS\tSESSIONS\tTURNS\tINPUT TOKENS\tOUTPUT TOKENS\tCACHED TOKENS\tREASONING\tTOTAL TOKENS\tCOST (USD)")
	}

	var totalRow UsageStatsRow
	for _, r := range rows {
		totalRow.Sessions += r.Sessions
		totalRow.Turns += r.Turns
		totalRow.InputTokens += r.InputTokens
		totalRow.OutputTokens += r.OutputTokens
		totalRow.CachedTokens += r.CachedTokens
		totalRow.CacheWriteTokens += r.CacheWriteTokens
		totalRow.ReasoningTokens += r.ReasoningTokens
		totalRow.TotalTokens += r.TotalTokens
		totalRow.CostUSD += r.CostUSD

		if opts.ByModel {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t$%0.4f\n",
				r.Harness, r.Model, r.Sessions, r.Turns,
				formatTokens(r.InputTokens), formatTokens(r.OutputTokens),
				formatTokens(r.CachedTokens), formatTokens(r.ReasoningTokens),
				formatTokens(r.TotalTokens), r.CostUSD)
		} else {
			_, _ = fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t$%0.4f\n",
				r.Harness, r.Sessions, r.Turns,
				formatTokens(r.InputTokens), formatTokens(r.OutputTokens),
				formatTokens(r.CachedTokens), formatTokens(r.ReasoningTokens),
				formatTokens(r.TotalTokens), r.CostUSD)
		}
	}

	if opts.ByModel {
		_, _ = fmt.Fprintf(w, "TOTAL\t-\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t$%0.4f\n",
			totalRow.Sessions, totalRow.Turns,
			formatTokens(totalRow.InputTokens), formatTokens(totalRow.OutputTokens),
			formatTokens(totalRow.CachedTokens), formatTokens(totalRow.ReasoningTokens),
			formatTokens(totalRow.TotalTokens), totalRow.CostUSD)
	} else {
		_, _ = fmt.Fprintf(w, "TOTAL\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t$%0.4f\n",
			totalRow.Sessions, totalRow.Turns,
			formatTokens(totalRow.InputTokens), formatTokens(totalRow.OutputTokens),
			formatTokens(totalRow.CachedTokens), formatTokens(totalRow.ReasoningTokens),
			formatTokens(totalRow.TotalTokens), totalRow.CostUSD)
	}
	return w.Flush()
}

func formatTokens(n int64) string {
	in := strconv.FormatInt(n, 10)
	if len(in) <= 3 {
		return in
	}
	var out []byte
	lead := len(in) % 3
	if lead > 0 {
		out = append(out, in[:lead]...)
		if len(in) > lead {
			out = append(out, ',')
		}
	}
	for i := lead; i < len(in); i += 3 {
		out = append(out, in[i:i+3]...)
		if i+3 < len(in) {
			out = append(out, ',')
		}
	}
	return string(out)
}

// parseFlexibleTime accepts a Go duration ("24h"), a day count ("7d"), or an
// absolute date. It reports an error rather than a zero time so an unparseable
// window fails the command instead of silently widening it to all time.
func parseFlexibleTime(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	if d, err := time.ParseDuration(val); err == nil {
		return time.Now().UTC().Add(-d), nil
	}
	if strings.HasSuffix(val, "d") {
		daysStr := strings.TrimSuffix(val, "d")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err == nil && days > 0 {
			return time.Now().UTC().AddDate(0, 0, -days), nil
		}
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, val); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q; use a duration (24h), a day count (7d), or a date (2006-01-02)", val)
}

// RunAgentUsageList lists individual usage records.
func RunAgentUsageList(ctx context.Context, state *GlobalState, opts UsageListOptions) error {
	records, err := LoadAllUsageRecords()
	if err != nil {
		return fmt.Errorf("failed to load usage records: %w", err)
	}

	var filtered []UsageRecord
	for _, r := range records {
		if opts.Harness != "" && r.Harness != opts.Harness && r.Agent != opts.Harness {
			continue
		}
		filtered = append(filtered, r)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	if opts.JSON {
		encoder := json.NewEncoder(state.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(filtered)
	}

	if len(filtered) == 0 {
		_, _ = fmt.Fprintln(state.Stdout, "No usage records found.")
		return nil
	}

	w := tabwriter.NewWriter(state.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "TIMESTAMP\tHARNESS\tSESSION ID\tMODEL\tTOTAL TOKENS\tCOST (USD)")
	for _, r := range filtered {
		ts := r.Timestamp
		if len(ts) > 19 {
			ts = ts[:19]
		}
		model := r.Model
		if model == "" {
			model = "-"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t$%0.4f\n",
			ts, r.Harness, r.SessionID, model, formatTokens(r.TotalTokens), r.CostUSD)
	}
	return w.Flush()
}

// RunAgentUsageShow displays detailed usage for a single session.
func RunAgentUsageShow(ctx context.Context, state *GlobalState, harness, sessionID string) error {
	if harness == "" || sessionID == "" {
		return errors.New("usage: dot agent usage show <harness> <session-id>")
	}
	dir, err := HarnessUsageDir(harness)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, sanitizeFilename(sessionID)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("usage record not found for %s session %s: %w", harness, sessionID, err)
	}
	_, _ = state.Stdout.Write(data)
	return nil
}

// usageSyncCandidate is one session a backfill sweep found in a raw harness store.
type usageSyncCandidate struct {
	SessionID string
	CWD       string
	// Path is the transcript to read. It stays empty for stores whose extractor
	// resolves its own source (databases, and Grok's signals.json).
	Path string
}

// usageSyncSource backfills one harness: enumerate the raw store, then turn every
// candidate it yields into a usage record. Splitting the two halves keeps the six
// stores from repeating the extract-write-count block they used to copy verbatim.
type usageSyncSource struct {
	enumerate func(ctx context.Context, state *GlobalState, root string) ([]usageSyncCandidate, error)
	extract   func(ctx context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error)
	agent     string
	// database marks a store held in a single SQLite file rather than a directory,
	// which decides how its existence is probed.
	database bool
}

// sync backfills one store and returns how many records it wrote. A missing store
// is not an error -- that harness simply is not installed -- but an unreadable one
// is: reporting "Synced 0 usage records" over a permission error is how a broken
// backfill used to look exactly like an empty one.
func (source usageSyncSource) sync(ctx context.Context, state *GlobalState) (int, error) {
	root, err := state.Config.Agent.SourceRoot(source.agent)
	if err != nil {
		return 0, err
	}
	if source.database {
		info, statErr := os.Stat(root)
		if errors.Is(statErr, os.ErrNotExist) {
			return 0, nil
		}
		if statErr != nil {
			return 0, fmt.Errorf("failed to inspect %s database %s: %w", source.agent, root, statErr)
		}
		if info.IsDir() {
			return 0, fmt.Errorf("%s database path is a directory: %s", source.agent, root)
		}
	} else {
		present, existsErr := sourceDirectoryExists(root, source.agent)
		if existsErr != nil {
			return 0, existsErr
		}
		if !present {
			return 0, nil
		}
	}

	candidates, err := source.enumerate(ctx, state, root)
	if err != nil {
		return 0, fmt.Errorf("failed to scan %s usage sources: %w", source.agent, err)
	}
	written := 0
	for _, candidate := range candidates {
		record, extractErr := source.extract(ctx, state, candidate)
		if extractErr != nil {
			return written, fmt.Errorf("failed to extract %s usage for session %s: %w", source.agent, candidate.SessionID, extractErr)
		}
		if record == nil {
			continue
		}
		if writeErr := WriteUsageRecord(*record); writeErr != nil {
			return written, fmt.Errorf("failed to write %s usage for session %s: %w", source.agent, candidate.SessionID, writeErr)
		}
		written++
	}
	return written, nil
}

// transcriptCandidates walks a directory store and yields one candidate per JSONL
// transcript whose name maps to a session id.
func transcriptCandidates(root string, sessionIDOf func(name string) string) ([]usageSyncCandidate, error) {
	var candidates []usageSyncCandidate
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
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
		candidates = append(candidates, usageSyncCandidate{SessionID: sessionID, Path: path})
		return nil
	})
	return candidates, err
}

// sqliteCandidates reads session identities out of a SQLite-backed store.
func sqliteCandidates(ctx context.Context, state *GlobalState, dbPath, query string) ([]usageSyncCandidate, error) {
	out, err := runSQLiteJSON(ctx, state, dbPath, query)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" || trimmed == "[]" {
		return nil, nil
	}
	var rows []struct {
		ID  string `json:"id"`
		CWD string `json:"cwd"`
	}
	if err := json.Unmarshal([]byte(trimmed), &rows); err != nil {
		return nil, fmt.Errorf("failed to parse session query result: %w", err)
	}
	candidates := make([]usageSyncCandidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, usageSyncCandidate{SessionID: row.ID, CWD: row.CWD})
	}
	return candidates, nil
}

// usageSyncSources lists every raw store the backfill sweeps, in reporting order.
func usageSyncSources() []usageSyncSource {
	return []usageSyncSource{
		{
			agent: sessionStoreClaude,
			enumerate: func(_ context.Context, _ *GlobalState, root string) ([]usageSyncCandidate, error) {
				return transcriptCandidates(root, func(name string) string {
					// memory.jsonl is hand-curated long-term memory, not a session
					// transcript, and its stem parses as a valid session id.
					if name == "memory.jsonl" {
						return ""
					}
					sessionID := strings.TrimSuffix(name, ".jsonl")
					if !isValidSessionID(sessionID) {
						return ""
					}
					return sessionID
				})
			},
			extract: func(_ context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error) {
				return ExtractUsageClaude(state.Config.Agent, candidate.SessionID, candidate.CWD, candidate.Path)
			},
		},
		{
			agent: sessionStoreCodex,
			enumerate: func(_ context.Context, _ *GlobalState, root string) ([]usageSyncCandidate, error) {
				return transcriptCandidates(root, func(name string) string {
					return extractCodexSessionID(strings.TrimSuffix(name, ".jsonl"))
				})
			},
			extract: func(_ context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error) {
				return ExtractUsageCodex(state.Config.Agent, candidate.SessionID, candidate.CWD, candidate.Path)
			},
		},
		{
			agent:    sessionStoreOpenCode,
			database: true,
			enumerate: func(ctx context.Context, state *GlobalState, root string) ([]usageSyncCandidate, error) {
				return sqliteCandidates(ctx, state, root, "SELECT id, directory AS cwd FROM session;")
			},
			extract: func(ctx context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error) {
				return ExtractUsageOpencode(ctx, state, candidate.SessionID, candidate.CWD)
			},
		},
		{
			agent:    sessionStoreCopilot,
			database: true,
			enumerate: func(ctx context.Context, state *GlobalState, root string) ([]usageSyncCandidate, error) {
				return sqliteCandidates(ctx, state, root, "SELECT id, cwd FROM sessions;")
			},
			extract: func(ctx context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error) {
				return ExtractUsageCopilot(ctx, state, candidate.SessionID, candidate.CWD)
			},
		},
		{
			agent: sessionStoreGrok,
			enumerate: func(_ context.Context, _ *GlobalState, root string) ([]usageSyncCandidate, error) {
				// Grok nests sessions one level below a per-cwd directory whose name
				// is an encoded path, so the extractor re-locates the session itself.
				cwdDirs, err := os.ReadDir(root)
				if err != nil {
					return nil, err
				}
				var candidates []usageSyncCandidate
				for _, cwdDir := range cwdDirs {
					if !cwdDir.IsDir() {
						continue
					}
					sessions, readErr := os.ReadDir(filepath.Join(root, cwdDir.Name()))
					if readErr != nil {
						return nil, readErr
					}
					for _, session := range sessions {
						if session.IsDir() {
							candidates = append(candidates, usageSyncCandidate{SessionID: session.Name()})
						}
					}
				}
				return candidates, nil
			},
			extract: func(_ context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error) {
				return ExtractUsageGrok(state, candidate.SessionID, candidate.CWD)
			},
		},
		{
			agent: sessionStoreAgy,
			enumerate: func(_ context.Context, state *GlobalState, root string) ([]usageSyncCandidate, error) {
				entries, err := os.ReadDir(root)
				if err != nil {
					return nil, err
				}
				var candidates []usageSyncCandidate
				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}
					// Resolve through the shared candidate list so the backfill reads
					// the same transcript the hook does. Hard-coding transcript.jsonl
					// skipped sessions carrying only transcript_full.jsonl, and read
					// the shorter file when both existed -- which then overwrote the
					// accurate hook-written record with a byte-count undercount.
					paths, pathErr := agyTranscriptCandidates(state.Config.Agent, entry.Name())
					if pathErr != nil {
						return nil, pathErr
					}
					for _, path := range paths {
						if _, statErr := os.Stat(path); statErr == nil {
							candidates = append(candidates, usageSyncCandidate{SessionID: entry.Name(), Path: path})
							break
						}
					}
				}
				return candidates, nil
			},
			extract: func(_ context.Context, state *GlobalState, candidate usageSyncCandidate) (*UsageRecord, error) {
				return ExtractUsageAgy(state, candidate.SessionID, candidate.CWD, candidate.Path)
			},
		},
	}
}

// RunAgentUsageSync scans existing raw session stores and backfills/refreshes usage records.
func RunAgentUsageSync(ctx context.Context, state *GlobalState) error {
	syncedCount := 0
	harnessesSynced := 0
	for _, source := range usageSyncSources() {
		written, err := source.sync(ctx, state)
		if err != nil {
			return err
		}
		if written > 0 {
			reportSyncOutcome(state, source.agent, syncOutcome{verb: "recorded", count: written})
			harnessesSynced++
			syncedCount += written
		}
	}

	_, _ = fmt.Fprintf(state.Stdout, "Synced %d usage records across %d harnesses into ~/.agents/usages\n", syncedCount, harnessesSynced)
	return nil
}

// NewAgentUsageCmd constructs the `dot agent usage` command group.
func NewAgentUsageCmd(state *GlobalState) *cli.Command {
	statsAction := func(ctx context.Context, cmd *cli.Command) error {
		return RunAgentUsageStats(ctx, state, UsageStatsOptions{
			Harness: cmd.String("harness"),
			Since:   cmd.String("since"),
			Until:   cmd.String("until"),
			ByModel: cmd.Bool("by-model"),
			JSON:    cmd.Bool("json"),
		})
	}

	return &cli.Command{
		Name:    "usage",
		Aliases: []string{"u"},
		Usage:   "Manage and compute stats on agent token usage in ~/.agents/usages",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "harness", Aliases: []string{"a"}, Usage: "Filter by harness name (claude, codex, grok, etc.)"},
			&cli.StringFlag{Name: "since", Usage: "Filter by duration (24h, 7d) or RFC3339 timestamp"},
			&cli.StringFlag{Name: "until", Usage: "Filter up to duration or timestamp"},
			&cli.BoolFlag{Name: "by-model", Aliases: []string{"m"}, Usage: "Break down token usage by harness and model"},
			&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "Output stats as JSON"},
		},
		Action: statsAction,
		Commands: []*cli.Command{
			{
				Name:    "stats",
				Aliases: []string{"s"},
				Usage:   "Compute summary token usage statistics across harnesses",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "harness", Aliases: []string{"a"}, Usage: "Filter by harness name"},
					&cli.StringFlag{Name: "since", Usage: "Filter by start duration/timestamp"},
					&cli.StringFlag{Name: "until", Usage: "Filter by end duration/timestamp"},
					&cli.BoolFlag{Name: "by-model", Aliases: []string{"m"}, Usage: "Break down by harness and model"},
					&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "Output stats as JSON"},
				},
				Action: statsAction,
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List session usage records",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "harness", Aliases: []string{"a"}, Usage: "Filter by harness name"},
					&cli.IntFlag{Name: "limit", Aliases: []string{"n"}, Value: 50, Usage: "Maximum sessions to list"},
					&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "Output list as JSON"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return RunAgentUsageList(ctx, state, UsageListOptions{
						Harness: cmd.String("harness"),
						Limit:   int(cmd.Int("limit")),
						JSON:    cmd.Bool("json"),
					})
				},
			},
			{
				Name:      "show",
				Aliases:   []string{"w"},
				Usage:     "Show detailed usage record for a session",
				ArgsUsage: "<harness> <session-id>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return RunAgentUsageShow(ctx, state, cmd.Args().Get(0), cmd.Args().Get(1))
				},
			},
			{
				Name:    "sync",
				Aliases: []string{"y"},
				Usage:   "Backfill and synchronize usage records from raw harness stores",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return RunAgentUsageSync(ctx, state)
				},
			},
		},
	}
}
