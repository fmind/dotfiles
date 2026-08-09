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
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

// HookInput represents the normalized JSON context passed by agent hooks on stdin.
type HookInput struct {
	SessionID      string `json:"session_id"`
	CWD            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path"`
	StopHookActive bool   `json:"stop_hook_active"`
	FullyIdle      bool   `json:"fullyIdle"`
}

// UnmarshalJSON normalizes the snake_case Claude/Codex hook payload and the
// camelCase Antigravity payload into one canonical HookInput.
func (h *HookInput) UnmarshalJSON(data []byte) error {
	var raw struct {
		SessionID                 string   `json:"session_id"`
		ConversationID            string   `json:"conversationId"`
		CWD                       string   `json:"cwd"`
		TranscriptPath            string   `json:"transcript_path"`
		AntigravityTranscriptPath string   `json:"transcriptPath"`
		WorkspacePaths            []string `json:"workspacePaths"`
		StopHookActive            bool     `json:"stop_hook_active"`
		FullyIdle                 bool     `json:"fullyIdle"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	h.SessionID = raw.SessionID
	if h.SessionID == "" {
		h.SessionID = raw.ConversationID
	}
	h.CWD = raw.CWD
	if h.CWD == "" {
		for _, path := range raw.WorkspacePaths {
			if path != "" {
				h.CWD = path
				break
			}
		}
	}
	h.TranscriptPath = raw.TranscriptPath
	if h.TranscriptPath == "" {
		h.TranscriptPath = raw.AntigravityTranscriptPath
	}
	h.StopHookActive = raw.StopHookActive
	h.FullyIdle = raw.FullyIdle
	return nil
}

// SessionLogLine is the unified format for logging prompt turns.
type SessionLogLine struct {
	TS      string `json:"ts"`
	Agent   string `json:"agent"`
	SID     string `json:"sid"`
	Role    string `json:"role"`
	Content string `json:"content"`
	CWD     string `json:"cwd,omitempty"`
	Model   string `json:"model,omitempty"`
}

// NewAgentCmd constructs the agent command group.
func NewAgentCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "agent",
		Aliases: []string{"a"},
		Usage:   "Manage AI agent integrations and sessions",
		Commands: []*cli.Command{
			NewAgentDoctorCmd(state),
			NewAgentHookCmd(state),
			NewAgentSessionCmd(state),
		},
	}
}

// agentSessionLogger is the per-agent ingestion entrypoint shared by the `agent
// session <agent>` subcommands and the `agent hook session <agent>` wrapper.
type agentSessionLogger func(ctx context.Context, state *GlobalState, sessionID, cwd string) error

// agentSessionEntry is one agent's ingestion entrypoint and the one-line usage its
// subcommand advertises.
type agentSessionEntry struct {
	Log   agentSessionLogger
	Usage string
}

// agentSessionLoggers maps each agent to its ingestion entry. Keeping the five
// agents in a table rather than five near-identical command literals means a new
// agent is one entry, not a copy.
func agentSessionLoggers() map[string]agentSessionEntry {
	return map[string]agentSessionEntry{
		sessionStoreAgy:      {RunAgentSessionLogAgy, "Process session end hook for Antigravity (agy)"},
		sessionStoreClaude:   {RunAgentSessionLogClaude, "Process session end/stop hook for Claude Code"},
		sessionStoreCodex:    {RunAgentSessionLogCodex, "Process session hook for OpenAI Codex"},
		sessionStoreOpenCode: {RunAgentSessionLogOpencode, "Process session hook for OpenCode"},
		sessionStoreCopilot:  {RunAgentSessionLogCopilot, "Process a GitHub Copilot CLI session from its session store"},
	}
}

// NewAgentSessionCmd constructs the agent session command group.
func NewAgentSessionCmd(state *GlobalState) *cli.Command {
	commands := []*cli.Command{
		NewAgentSessionListCmd(state),
		NewAgentSessionShowCmd(state),
		NewAgentSessionExportCmd(state),
	}
	// Table order follows the canonical agent table so `--help` lists the agents in
	// the same order everywhere dot reports them.
	loggers := agentSessionLoggers()
	for _, definition := range agentDefinitions() {
		entry, ok := loggers[definition.Agent]
		if !ok {
			continue
		}
		commands = append(commands, &cli.Command{
			Name:      definition.Agent,
			Usage:     entry.Usage,
			ArgsUsage: "[SESSION-ID] [CWD]",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return entry.Log(ctx, state, cmd.Args().Get(0), cmd.Args().Get(1))
			},
		})
	}
	return &cli.Command{
		Name:    "session",
		Aliases: []string{"s"},
		Usage:   "Manage agent session logs",
		Commands: append(commands,
			NewAgentSessionSyncCmd(state),
			&cli.Command{
				Name:  "migrate",
				Usage: "Select the most complete legacy transcript per lineage without deleting evidence",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "apply", Usage: "Write selected transcripts to the versioned store (default is dry-run)"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return RunAgentSessionMigrate(ctx, state, cmd.Bool("apply"))
				},
			},
		),
	}
}

// NewAgentSessionSyncCmd handles checking and syncing all untracked sessions.
func NewAgentSessionSyncCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "sync",
		Aliases: []string{"s"},
		Usage:   "Scan for new sessions across all agents and log them",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return RunAgentSessionSync(ctx, state)
		},
	}
}

func sessionStateWithTranscript(state *GlobalState, sessionID, transcriptPath string) (*GlobalState, error) {
	payload, err := json.Marshal(HookInput{SessionID: sessionID, TranscriptPath: transcriptPath, FullyIdle: true})
	if err != nil {
		return nil, err
	}
	sourced := *state
	sourced.Stdin = strings.NewReader(string(payload))
	return &sourced, nil
}

// hookIdentity is the session identity one ingestion call resolved, merged from the
// command-line operands and an optional hook payload on stdin.
type hookIdentity struct {
	Input   *HookInput
	Session string
	CWD     string
}

// FromHook reports whether this invocation was driven by a hook payload rather than
// by `agent session sync` or a manual command line.
func (h hookIdentity) FromHook() bool { return h.Input != nil }

// resolveHookIdentity merges the operands with any hook payload on stdin and
// validates the result. It returns halt=true when the payload says this invocation
// must do nothing: a re-entrant Stop hook, or — when requireIdle is set for an agent
// whose Stop event also fires mid-turn — a turn that is not finished yet.
func resolveHookIdentity(state *GlobalState, sessionID, cwd string, requireIdle bool) (identity hookIdentity, halt bool, err error) {
	input, err := parseStdin(state.Stdin)
	if err != nil {
		return hookIdentity{}, false, err
	}
	if input != nil {
		if input.StopHookActive || (requireIdle && !input.FullyIdle) {
			return hookIdentity{Input: input}, true, nil
		}
		if sessionID == "" {
			sessionID = input.SessionID
		}
		if cwd == "" {
			cwd = input.CWD
		}
	}
	if sessionID == "" {
		return hookIdentity{}, false, errors.New("missing session_id")
	}
	if !isValidSessionID(sessionID) {
		return hookIdentity{}, false, fmt.Errorf("invalid session_id format: %q", sessionID)
	}
	return hookIdentity{Input: input, Session: sessionID, CWD: resolveCWD(cwd)}, false, nil
}

// hookTranscriptPath returns the transcript the hook payload pointed at, after
// verifying it is a readable regular file. An empty result means the payload named
// no transcript and the caller must discover it from the agent's own store layout.
func hookTranscriptPath(identity hookIdentity, agent string) (string, error) {
	if identity.Input == nil || identity.Input.TranscriptPath == "" {
		return "", nil
	}
	path := ExpandPath(identity.Input.TranscriptPath)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s transcript from hook payload is unavailable at %s: %w", agent, path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s transcript from hook payload is not a file: %s", agent, path)
	}
	return path, nil
}

// parseStdin reads stdin to extract HookInput when data is piped by an agent hook.
func parseStdin(stdin io.Reader) (*HookInput, error) {
	if stdin == nil {
		return nil, nil
	}
	if file, ok := stdin.(*os.File); ok {
		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, nil
		}
	}

	// Hook producers close stdin after writing one JSON payload, so ReadAll cannot
	// block in normal hook execution and also handles payloads without a trailing newline.
	data, err := io.ReadAll(stdin)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse agent hook input: %w", err)
	}
	return &input, nil
}

// resolveCWD converts a relative or empty CWD to an absolute path.
func resolveCWD(cwd string) string {
	if cwd == "" {
		return ""
	}
	if cwd == "." {
		if pwd, err := os.Getwd(); err == nil {
			return pwd
		}
		return "."
	}
	if abs, err := filepath.Abs(cwd); err == nil {
		return abs
	}
	return cwd
}

// isValidSessionRune checks if a rune is allowed in a session ID.
func isValidSessionRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
}

func isValidSessionID(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	for _, r := range sessionID {
		if !isValidSessionRune(r) {
			return false
		}
	}
	return true
}

// extractCodexSessionID extracts the session ID from rollout filename (without .jsonl).
func extractCodexSessionID(name string) string {
	if !strings.HasPrefix(name, "rollout-") {
		return ""
	}
	parts := strings.Split(name, "-")
	if len(parts) >= 7 {
		return strings.Join(parts[6:], "-")
	}
	return ""
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func mapValue(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func textFromCodexContent(value any) string {
	switch content := value.(type) {
	case string:
		return content
	case []any:
		textParts := make([]string, 0, len(content))
		for _, part := range content {
			switch partValue := part.(type) {
			case string:
				textParts = append(textParts, partValue)
			case map[string]any:
				if text := stringValue(partValue["text"]); text != "" {
					textParts = append(textParts, text)
				} else if text := stringValue(partValue["content"]); text != "" {
					textParts = append(textParts, text)
				}
			}
		}
		return strings.Join(textParts, "\n")
	default:
		return ""
	}
}

func codexRole(raw map[string]any) string {
	if role := stringValue(raw["role"]); role != "" {
		return role
	}

	payload := mapValue(raw["payload"])
	if payload != nil {
		if role := stringValue(payload["role"]); role != "" {
			return role
		}
	}

	switch stringValue(raw["type"]) {
	case "user", "user_message":
		return "user"
	case "assistant", "assistant_message", "agent_message":
		return "assistant"
	default:
		return ""
	}
}

func codexContent(raw map[string]any) string {
	if content := textFromCodexContent(raw["content"]); content != "" {
		return content
	}

	payload := mapValue(raw["payload"])
	if payload != nil {
		if content := textFromCodexContent(payload["content"]); content != "" {
			return content
		}
		if content := stringValue(payload["message"]); content != "" {
			return content
		}
		if content := stringValue(payload["text"]); content != "" {
			return content
		}
	}

	if content := stringValue(raw["message"]); content != "" {
		return content
	}
	return stringValue(raw["text"])
}

func codexModel(raw map[string]any) string {
	if model := stringValue(raw["model"]); model != "" {
		return model
	}
	if payload := mapValue(raw["payload"]); payload != nil {
		return stringValue(payload["model"])
	}
	return ""
}

func codexCWD(raw map[string]any) string {
	if cwd := stringValue(raw["cwd"]); cwd != "" {
		return cwd
	}
	if payload := mapValue(raw["payload"]); payload != nil {
		return stringValue(payload["cwd"])
	}
	return ""
}

// agyTranscriptNames are the transcript files an agy session may carry, most
// complete first.
var agyTranscriptNames = []string{"transcript_full.jsonl", "transcript.jsonl"}

// agyTranscriptCandidates returns the transcript paths for one agy session, in
// preference order, beneath the configured source root.
func agyTranscriptCandidates(cfg AgentConfig, sessionID string) ([]string, error) {
	root, err := cfg.SourceRoot(sessionStoreAgy)
	if err != nil {
		return nil, err
	}
	candidates := make([]string, 0, len(agyTranscriptNames))
	for _, name := range agyTranscriptNames {
		candidates = append(candidates, filepath.Join(root, sessionID, ".system_generated", "logs", name))
	}
	return candidates, nil
}

// RunAgentSessionLogAgy reads the agy transcript files and processes the session.
func RunAgentSessionLogAgy(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	// agy fires Stop on every turn boundary, so a busy turn must be acknowledged
	// without ingesting a half-written transcript.
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
	if transcriptPath == "" {
		candidates, candidateErr := agyTranscriptCandidates(state.Config.Agent, sessionID)
		if candidateErr != nil {
			return candidateErr
		}
		for _, candidate := range candidates {
			if _, statErr := os.Stat(candidate); statErr == nil {
				transcriptPath = candidate
				break
			}
		}
		if transcriptPath == "" {
			return fmt.Errorf("transcript file not found for agy session %s", sessionID)
		}
	}

	fingerprint, stored, isStored, err := fingerprintStoredTranscript("agy", sessionID, transcriptPath)
	if err != nil {
		return err
	}
	if isStored {
		reportStoredGeneration(state.Stderr, stored)
		if identity.FromHook() {
			return writeAntigravityStopDecision(state.Stdout)
		}
		return nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	decodeStats, decodeErr := decodeJSONLWithStats(state.Stderr, transcriptPath, file, func(raw map[string]any) error {
		if trunc, ok := raw["is_truncated"].(bool); ok && trunc {
			return nil
		}

		source, _ := raw["source"].(string)
		typ, _ := raw["type"].(string)
		createdAt, _ := raw["created_at"].(string)
		content, _ := raw["content"].(string)

		var role string
		if source == "USER_EXPLICIT" && typ == "USER_INPUT" {
			role = "user"
		} else if source == "MODEL" && typ == "PLANNER_RESPONSE" {
			role = "assistant"
		} else {
			return nil
		}

		if strings.TrimSpace(content) == "" {
			return nil
		}

		logs = append(logs, SessionLogLine{
			TS:      createdAt,
			Agent:   "agy",
			SID:     sessionID,
			Role:    role,
			Content: content,
			CWD:     cwd,
		})
		return nil
	})
	if decodeErr != nil {
		return decodeErr
	}

	if _, err := writeSessionLogs(ctx, state, "agy", sessionID, logs, sessionSource{Type: "antigravity-jsonl", Fingerprint: fingerprint, Malformed: decodeStats.Malformed, Skipped: decodeStats.Decoded - len(logs)}); err != nil {
		return err
	}
	if identity.FromHook() {
		return writeAntigravityStopDecision(state.Stdout)
	}
	return nil
}

func writeAntigravityStopDecision(stdout io.Writer) error {
	response := struct {
		Decision string `json:"decision"`
	}{}
	if err := json.NewEncoder(stdout).Encode(response); err != nil {
		return fmt.Errorf("failed to write Antigravity Stop response: %w", err)
	}
	return nil
}

func sourceDirectoryExists(path, source string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to inspect %s session directory %s: %w", source, path, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s session path is not a directory: %s", source, path)
	}
	return true, nil
}

func findSessionFile(root string, matches func(path string, entry fs.DirEntry) bool) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && matches(path, entry) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return found, nil
}

// claudeProjectDirectory encodes a working directory the way Claude Code names its
// per-project transcript folder: path separators and dots become dashes.
func claudeProjectDirectory(cwd string) string {
	encoded := strings.ReplaceAll(cwd, "/", "-")
	encoded = strings.ReplaceAll(encoded, ".", "-")
	return "-" + strings.TrimPrefix(encoded, "-")
}

// RunAgentSessionLogClaude reads the Claude JSONL files and processes the session.
func RunAgentSessionLogClaude(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, false)
	if err != nil {
		return err
	}
	if halt {
		return nil
	}
	sessionID, cwd = identity.Session, identity.CWD

	sessionFile, err := hookTranscriptPath(identity, "claude")
	if err != nil {
		return err
	}
	if sessionFile == "" {
		projectsDir, rootErr := state.Config.Agent.SourceRoot(sessionStoreClaude)
		if rootErr != nil {
			return rootErr
		}
		// The CWD-derived path is exact when it exists, so the directory scan below is
		// only a fallback for a session logged from a different working directory.
		sessionFile = filepath.Join(projectsDir, claudeProjectDirectory(cwd), sessionID+".jsonl")
		_, statErr := os.Stat(sessionFile)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("failed to inspect expected Claude transcript %s: %w", sessionFile, statErr)
		}
		if statErr != nil {
			found, findErr := findSessionFile(projectsDir, func(_ string, entry fs.DirEntry) bool {
				return entry.Name() == sessionID+".jsonl"
			})
			if findErr != nil {
				return fmt.Errorf("failed to search Claude transcripts in %s: %w", projectsDir, findErr)
			}
			if found == "" {
				return fmt.Errorf("session file not found for claude session %s", sessionID)
			}
			sessionFile = found
		}
	}

	fingerprint, stored, isStored, err := fingerprintStoredTranscript("claude", sessionID, sessionFile)
	if err != nil {
		return err
	}
	if isStored {
		reportStoredGeneration(state.Stderr, stored)
		return nil
	}

	file, err := os.Open(sessionFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	decodeStats, decodeErr := decodeJSONLWithStats(state.Stderr, sessionFile, file, func(raw map[string]any) error {
		typ, _ := raw["type"].(string)
		if typ != "user" && typ != "assistant" {
			return nil
		}

		ts, _ := raw["timestamp"].(string)
		msgVal, ok := raw["message"].(map[string]any)
		if !ok {
			return nil
		}

		var content string
		switch typ {
		case "user":
			content, _ = msgVal["content"].(string)
		case "assistant":
			if contentsList, ok := msgVal["content"].([]any); ok {
				var textParts []string
				for _, part := range contentsList {
					if partMap, ok := part.(map[string]any); ok {
						if ptype, _ := partMap["type"].(string); ptype == "text" {
							if text, _ := partMap["text"].(string); text != "" {
								textParts = append(textParts, text)
							}
						}
					}
				}
				content = strings.Join(textParts, "\n")
			}
		}

		logCWD, _ := raw["cwd"].(string)
		if logCWD == "" {
			logCWD = cwd
		}
		logCWD = resolveCWD(logCWD)

		var model string
		if m, ok := msgVal["model"].(string); ok {
			model = m
		}

		if strings.TrimSpace(content) == "" {
			return nil
		}

		logs = append(logs, SessionLogLine{
			TS:      ts,
			Agent:   "claude",
			SID:     sessionID,
			Role:    typ,
			Content: content,
			CWD:     logCWD,
			Model:   model,
		})
		return nil
	})
	if decodeErr != nil {
		return decodeErr
	}

	_, err = writeSessionLogs(ctx, state, "claude", sessionID, logs, sessionSource{Type: "claude-jsonl", Fingerprint: fingerprint, Malformed: decodeStats.Malformed, Skipped: decodeStats.Decoded - len(logs)})
	return err
}

// RunAgentSessionLogCodex reads Codex rollout session files and processes the session.
func RunAgentSessionLogCodex(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
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
	if transcriptPath == "" {
		sessionsDir, rootErr := state.Config.Agent.SourceRoot(sessionStoreCodex)
		if rootErr != nil {
			return rootErr
		}
		found, findErr := findSessionFile(sessionsDir, func(_ string, entry fs.DirEntry) bool {
			if !strings.HasSuffix(entry.Name(), ".jsonl") {
				return false
			}
			return extractCodexSessionID(strings.TrimSuffix(entry.Name(), ".jsonl")) == sessionID
		})
		if findErr != nil {
			return fmt.Errorf("failed to search Codex transcripts in %s: %w", sessionsDir, findErr)
		}
		if found == "" {
			return fmt.Errorf("session file not found for codex session %s", sessionID)
		}
		transcriptPath = found
	}

	fingerprint, stored, isStored, err := fingerprintStoredTranscript("codex", sessionID, transcriptPath)
	if err != nil {
		return err
	}
	if isStored {
		reportStoredGeneration(state.Stderr, stored)
		return nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	activeModel := ""
	activeCWD := cwd
	decodeStats, decodeErr := decodeJSONLWithStats(state.Stderr, transcriptPath, file, func(raw map[string]any) error {
		if model := codexModel(raw); model != "" {
			activeModel = model
		}
		if logCWD := codexCWD(raw); logCWD != "" {
			activeCWD = resolveCWD(logCWD)
		}

		role := codexRole(raw)
		if role != "user" && role != "assistant" {
			return nil
		}

		content := codexContent(raw)
		if strings.TrimSpace(content) == "" {
			return nil
		}

		ts, _ := raw["timestamp"].(string)
		if ts == "" {
			ts, _ = raw["created_at"].(string)
		}
		if ts == "" {
			ts, _ = raw["ts"].(string)
		}
		// A missing source time remains empty so normalized output does not change
		// merely because a hook and sync run at different times.

		model := codexModel(raw)
		if model == "" {
			model = activeModel
		}

		logCWD := codexCWD(raw)
		if logCWD != "" {
			logCWD = resolveCWD(logCWD)
		} else {
			logCWD = activeCWD
		}

		logs = append(logs, SessionLogLine{
			TS:      ts,
			Agent:   "codex",
			SID:     sessionID,
			Role:    role,
			Content: content,
			CWD:     logCWD,
			Model:   model,
		})
		return nil
	})
	if decodeErr != nil {
		return decodeErr
	}

	_, err = writeSessionLogs(ctx, state, "codex", sessionID, logs, sessionSource{Type: "codex-jsonl", Fingerprint: fingerprint, Malformed: decodeStats.Malformed, Skipped: decodeStats.Decoded - len(logs)})
	return err
}

// OpencodeData represents the nested structure inside message.data for OpenCode.
type OpencodeData struct {
	Role  string `json:"role"`
	Model struct {
		ProviderID string `json:"providerID"`
		ModelID    string `json:"modelID"`
	} `json:"model"`
}

// OpencodePart represents the content stored in OpenCode's part table.
type OpencodePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// OpencodeRow represents the joined message and part rows returned by sqlite3 -json.
type OpencodeRow struct {
	SessionID   string `json:"session_id"`
	MessageID   string `json:"message_id"`
	PartID      string `json:"part_id"`
	Data        string `json:"data"`
	PartData    string `json:"part_data"`
	Directory   string `json:"directory"`
	TimeCreated int64  `json:"time_created"`
}

const opencodeMessagesQuery = `SELECT
    m.session_id AS session_id,
    m.id AS message_id,
    m.time_created,
    m.data,
    s.directory,
    p.id AS part_id,
    p.data AS part_data
FROM message m
JOIN session s ON m.session_id = s.id
LEFT JOIN part p ON p.message_id = m.id
WHERE %s
ORDER BY m.session_id, m.time_created, m.id, p.time_created, p.id`

const opencodeSessionsQuery = "SELECT id, directory FROM session"

// parseOpencodeRows converts OpenCode's normalized message/part rows into session logs.
func parseOpencodeRows(sessionID, fallbackCWD string, rows []OpencodeRow) ([]SessionLogLine, error) {
	type message struct {
		data        OpencodeData
		id          string
		directory   string
		textParts   []string
		timeCreated int64
	}

	logs := make([]SessionLogLine, 0, len(rows))
	var current message
	flush := func() {
		content := strings.Join(current.textParts, "\n")
		if (current.data.Role != "user" && current.data.Role != "assistant") || strings.TrimSpace(content) == "" {
			return
		}

		logCWD := current.directory
		if logCWD == "" {
			logCWD = fallbackCWD
		}
		logCWD = resolveCWD(logCWD)

		model := current.data.Model.ModelID
		if current.data.Model.ProviderID != "" && model != "" {
			model = current.data.Model.ProviderID + "/" + model
		}

		logs = append(logs, SessionLogLine{
			TS:      time.UnixMilli(current.timeCreated).UTC().Format(time.RFC3339),
			Agent:   "opencode",
			SID:     sessionID,
			Role:    current.data.Role,
			Content: content,
			CWD:     logCWD,
			Model:   model,
		})
	}

	for _, row := range rows {
		if row.MessageID != current.id {
			if current.id != "" {
				flush()
			}
			current = message{
				id:          row.MessageID,
				directory:   row.Directory,
				timeCreated: row.TimeCreated,
			}
			if err := json.Unmarshal([]byte(row.Data), &current.data); err != nil {
				return nil, fmt.Errorf("failed to parse OpenCode message %s: %w", row.MessageID, err)
			}
		}

		if row.PartData == "" {
			continue
		}
		var part OpencodePart
		if err := json.Unmarshal([]byte(row.PartData), &part); err != nil {
			return nil, fmt.Errorf("failed to parse OpenCode part %s: %w", row.PartID, err)
		}
		if part.Type == "text" && part.Text != "" {
			current.textParts = append(current.textParts, part.Text)
		}
	}
	if current.id != "" {
		flush()
	}

	return logs, nil
}

// decodeSessionIDs parses a sqlite3 -json `SELECT id ...` result. Idempotence is
// decided later from the agent-scoped fingerprint, never from a bare session ID.
func decodeSessionIDs(label, output string) ([]string, error) {
	output = strings.TrimSpace(output)
	if output == "" || output == "[]" {
		return nil, nil
	}

	var rows []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		return nil, fmt.Errorf("failed to decode %s session query result: %w", label, err)
	}
	if rows == nil {
		return nil, fmt.Errorf("failed to decode %s session query result: expected a JSON array", label)
	}

	sessionIDs := make([]string, 0, len(rows))
	seen := make(map[string]bool)
	for rowNumber, row := range rows {
		sessionID := row.ID
		if !isValidSessionID(sessionID) {
			return nil, fmt.Errorf("malformed %s session row %d: invalid session ID %q", label, rowNumber+1, sessionID)
		}
		if seen[sessionID] {
			return nil, fmt.Errorf("malformed %s session rows: duplicate session ID %q", label, sessionID)
		}
		seen[sessionID] = true
		sessionIDs = append(sessionIDs, sessionID)
	}
	return sessionIDs, nil
}

// sqliteSweep describes how one SQLite-backed agent store is swept end to end.
// OpenCode and Copilot differ only in their queries, row shape, and per-row
// validation, so the sweep itself is written once and parameterized here.
type sqliteSweep[R any] struct {
	// sessionOf reads the owning session ID from a row.
	sessionOf func(R) string
	// validate rejects a malformed row; rowNumber is 1-based for the message.
	validate func(row R, rowNumber int) error
	// parse turns one session's rows into normalized log lines.
	parse func(sessionID string, rows []R) ([]SessionLogLine, error)
	// agent is the canonical agent name; sourceType labels the ingested generation.
	agent      string
	sourceType string
	// label names the store in error messages; rowNoun and rowNounPlural name its
	// detail rows, singular for one bad row and plural for a failed batch query.
	label         string
	rowNoun       string
	rowNounPlural string
	// sessionsQuery enumerates the session IDs; rowsQuery takes one %s filter clause.
	sessionsQuery string
	rowsQuery     string
	// filterColumn is the qualified column the batched IN filter targets.
	filterColumn string
}

// syncSQLiteSessions enumerates every session in a SQLite-backed store, fetches all
// detail rows in one batched query, groups them per session, and ingests each
// lineage. Sessions whose rows produce no records are still recorded, so an empty
// session is proven empty rather than left looking unprocessed. It returns the
// number of sessions newly ingested.
func syncSQLiteSessions[R any](ctx context.Context, state *GlobalState, dbPath string, sweep sqliteSweep[R]) (int, error) {
	output, err := runSQLiteJSON(ctx, state, dbPath, sweep.sessionsQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query %s sessions: %w", sweep.label, err)
	}
	sessionIDs, err := decodeSessionIDs(sweep.label, output)
	if err != nil {
		return 0, err
	}
	if len(sessionIDs) == 0 {
		return 0, nil
	}

	quoted := make([]string, len(sessionIDs))
	requested := make(map[string]bool, len(sessionIDs))
	for index, sessionID := range sessionIDs {
		// Session IDs were validated above, so the exact value is safe to use as a
		// SQLite string literal without silently changing its identity.
		quoted[index] = "'" + sessionID + "'"
		requested[sessionID] = true
	}

	recordEmpty := func(sessionID string) error {
		_, writeErr := writeSessionLogs(ctx, state, sweep.agent, sessionID, nil, sessionSource{Type: sweep.sourceType})
		return writeErr
	}

	filter := sweep.filterColumn + " IN (" + strings.Join(quoted, ",") + ")"
	rowOutput, err := runSQLiteJSON(ctx, state, dbPath, fmt.Sprintf(sweep.rowsQuery, filter))
	if err != nil {
		return 0, fmt.Errorf("failed to query %s %s: %w", sweep.label, sweep.rowNounPlural, err)
	}
	rowOutput = strings.TrimSpace(rowOutput)
	if rowOutput == "" || rowOutput == "[]" {
		for _, sessionID := range sessionIDs {
			if writeErr := recordEmpty(sessionID); writeErr != nil {
				return 0, writeErr
			}
		}
		return 0, nil
	}

	var rows []R
	if err := json.Unmarshal([]byte(rowOutput), &rows); err != nil {
		return 0, fmt.Errorf("failed to decode %s %s query result: %w", sweep.label, sweep.rowNoun, err)
	}
	if rows == nil {
		return 0, fmt.Errorf("failed to decode %s %s query result: expected a JSON array", sweep.label, sweep.rowNoun)
	}

	sessionRows := make(map[string][]R, len(sessionIDs))
	for index, row := range rows {
		rowNumber := index + 1
		sessionID := sweep.sessionOf(row)
		if !isValidSessionID(sessionID) {
			return 0, fmt.Errorf("malformed %s %s row %d: invalid session ID %q", sweep.label, sweep.rowNoun, rowNumber, sessionID)
		}
		if !requested[sessionID] {
			return 0, fmt.Errorf("malformed %s %s row %d: unexpected session ID %q", sweep.label, sweep.rowNoun, rowNumber, sessionID)
		}
		if sweep.validate != nil {
			if err := sweep.validate(row, rowNumber); err != nil {
				return 0, err
			}
		}
		sessionRows[sessionID] = append(sessionRows[sessionID], row)
	}

	count := 0
	for _, sessionID := range sessionIDs {
		rows := sessionRows[sessionID]
		logs, parseErr := sweep.parse(sessionID, rows)
		if parseErr != nil {
			return 0, fmt.Errorf("failed to parse %s session %q: %w", sweep.label, sessionID, parseErr)
		}
		if len(logs) == 0 {
			if writeErr := recordEmpty(sessionID); writeErr != nil {
				return 0, writeErr
			}
			continue
		}
		fingerprint, fingerprintErr := fingerprintJSON(rows)
		if fingerprintErr != nil {
			return 0, fingerprintErr
		}
		result, writeErr := writeSessionLogs(ctx, state, sweep.agent, sessionID, logs, sessionSource{Type: sweep.sourceType, Fingerprint: fingerprint})
		if writeErr != nil {
			return 0, fmt.Errorf("failed to write %s session %q: %w", sweep.label, sessionID, writeErr)
		}
		if result.Status == sessionIngested {
			count++
		}
	}
	return count, nil
}

func syncOpencodeSessions(ctx context.Context, state *GlobalState, dbPath string) (int, error) {
	return syncSQLiteSessions(ctx, state, dbPath, sqliteSweep[OpencodeRow]{
		agent:         sessionStoreOpenCode,
		sourceType:    "opencode-db",
		label:         "OpenCode",
		rowNoun:       "message",
		rowNounPlural: "messages",
		sessionsQuery: opencodeSessionsQuery,
		rowsQuery:     opencodeMessagesQuery,
		filterColumn:  "m.session_id",
		sessionOf:     func(row OpencodeRow) string { return row.SessionID },
		validate: func(row OpencodeRow, rowNumber int) error {
			if row.MessageID == "" {
				return fmt.Errorf("malformed OpenCode message row %d: missing message ID", rowNumber)
			}
			if row.PartData != "" && row.PartID == "" {
				return fmt.Errorf("malformed OpenCode message row %d: missing part ID", rowNumber)
			}
			return nil
		},
		parse: func(sessionID string, rows []OpencodeRow) ([]SessionLogLine, error) {
			return parseOpencodeRows(sessionID, "", rows)
		},
	})
}

// RunAgentSessionLogOpencode reads OpenCode session records and writes them to the sessions directory.
func RunAgentSessionLogOpencode(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	identity, halt, err := resolveHookIdentity(state, sessionID, cwd, false)
	if err != nil {
		return err
	}
	if halt {
		return nil
	}
	sessionID, cwd = identity.Session, identity.CWD

	dbPath, err := state.Config.Agent.SourceRoot(sessionStoreOpenCode)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return fmt.Errorf("opencode database not found at %s", dbPath)
	}

	sqlQuery := fmt.Sprintf(opencodeMessagesQuery, "m.session_id = '"+sessionID+"'")
	out, err := runSQLiteJSON(ctx, state, dbPath, sqlQuery)
	if err != nil {
		return err
	}

	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		_, err = writeSessionLogs(ctx, state, "opencode", sessionID, nil, sessionSource{Type: "opencode-db"})
		return err
	}

	var rows []OpencodeRow
	if parseErr := json.Unmarshal([]byte(out), &rows); parseErr != nil {
		return fmt.Errorf("failed to parse OpenCode query result: %w", parseErr)
	}

	logs, err := parseOpencodeRows(sessionID, cwd, rows)
	if err != nil {
		return err
	}
	fingerprint, err := fingerprintJSON(rows)
	if err != nil {
		return err
	}
	_, err = writeSessionLogs(ctx, state, "opencode", sessionID, logs, sessionSource{Type: "opencode-db", Fingerprint: fingerprint})
	return err
}

// CopilotRow is a joined session/turn row returned by sqlite3 -json from Copilot's session-store.db.
type CopilotRow struct {
	SessionID         string `json:"session_id"`
	UserMessage       string `json:"user_message"`
	AssistantResponse string `json:"assistant_response"`
	Timestamp         string `json:"timestamp"`
	CWD               string `json:"cwd"`
	TurnIndex         int    `json:"turn_index"`
}

const copilotTurnsQuery = `SELECT
    t.session_id AS session_id,
    t.turn_index AS turn_index,
    t.user_message AS user_message,
    t.assistant_response AS assistant_response,
    t.timestamp AS timestamp,
    s.cwd AS cwd
FROM turns t
JOIN sessions s ON t.session_id = s.id
WHERE %s
ORDER BY t.session_id, t.turn_index, t.id`

const copilotSessionsQuery = "SELECT id, cwd FROM sessions"

// runSQLiteJSON runs one read-only query against an agent's SQLite store. `-init
// /dev/null` keeps a user's ~/.sqliterc from injecting output settings that would
// corrupt the JSON this parser depends on.
func runSQLiteJSON(ctx context.Context, state *GlobalState, dbPath, query string) (string, error) {
	return state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, query)
}

// parseCopilotRows expands Copilot's per-turn user/assistant columns into session log lines.
// Copilot stores each turn as one row carrying both the prompt and the response, so a single
// row can yield up to two lines. NULL columns decode to empty strings and are skipped.
func parseCopilotRows(sessionID, fallbackCWD string, rows []CopilotRow) []SessionLogLine {
	logs := make([]SessionLogLine, 0, len(rows)*2)
	for _, row := range rows {
		logCWD := row.CWD
		if logCWD == "" {
			logCWD = fallbackCWD
		}
		logCWD = resolveCWD(logCWD)

		if strings.TrimSpace(row.UserMessage) != "" {
			logs = append(logs, SessionLogLine{
				TS:      row.Timestamp,
				Agent:   "copilot",
				SID:     sessionID,
				Role:    "user",
				Content: row.UserMessage,
				CWD:     logCWD,
			})
		}
		if strings.TrimSpace(row.AssistantResponse) != "" {
			logs = append(logs, SessionLogLine{
				TS:      row.Timestamp,
				Agent:   "copilot",
				SID:     sessionID,
				Role:    "assistant",
				Content: row.AssistantResponse,
				CWD:     logCWD,
			})
		}
	}
	return logs
}

// RunAgentSessionLogCopilot reads a single Copilot session from its store and logs it.
// The Copilot sessionEnd hook supplies only identity/lifecycle metadata, so the
// transcript remains sourced from this store-backed query.
func RunAgentSessionLogCopilot(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	// Copilot's sessionEnd hook carries no transcript payload, so identity comes from
	// the operands alone and stdin is never consumed here.
	if sessionID == "" {
		return errors.New("missing session_id")
	}
	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}
	cwd = resolveCWD(cwd)

	dbPath, err := state.Config.Agent.SourceRoot(sessionStoreCopilot)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return fmt.Errorf("copilot database not found at %s", dbPath)
	}

	sqlQuery := fmt.Sprintf(copilotTurnsQuery, "t.session_id = '"+sessionID+"'")
	out, err := runSQLiteJSON(ctx, state, dbPath, sqlQuery)
	if err != nil {
		return err
	}

	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		_, err = writeSessionLogs(ctx, state, "copilot", sessionID, nil, sessionSource{Type: "copilot-db"})
		return err
	}

	var rows []CopilotRow
	if parseErr := json.Unmarshal([]byte(out), &rows); parseErr != nil {
		return fmt.Errorf("failed to parse Copilot query result: %w", parseErr)
	}

	fingerprint, err := fingerprintJSON(rows)
	if err != nil {
		return err
	}
	_, err = writeSessionLogs(ctx, state, "copilot", sessionID, parseCopilotRows(sessionID, cwd, rows), sessionSource{Type: "copilot-db", Fingerprint: fingerprint})
	return err
}

// syncCopilotSessions scans the Copilot store and logs every untracked session.
func syncCopilotSessions(ctx context.Context, state *GlobalState, dbPath string) (int, error) {
	return syncSQLiteSessions(ctx, state, dbPath, sqliteSweep[CopilotRow]{
		agent:         sessionStoreCopilot,
		sourceType:    "copilot-db",
		label:         "Copilot",
		rowNoun:       "turn",
		rowNounPlural: "turns",
		sessionsQuery: copilotSessionsQuery,
		rowsQuery:     copilotTurnsQuery,
		filterColumn:  "t.session_id",
		sessionOf:     func(row CopilotRow) string { return row.SessionID },
		parse: func(sessionID string, rows []CopilotRow) ([]SessionLogLine, error) {
			return parseCopilotRows(sessionID, "", rows), nil
		},
	})
}

type jsonlDecodeStats struct {
	Decoded   int
	Malformed int
}

// decodeJSONLWithStats reads line-delimited JSON and retains malformed-record
// evidence so the normalized generation is marked partial instead of silently clean.
func decodeJSONLWithStats(warnOut io.Writer, filePath string, file *os.File, callback func(raw map[string]any) error) (jsonlDecodeStats, error) {
	var stats jsonlDecodeStats
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return stats, fmt.Errorf("reading file %s: %w", filePath, err)
		}
		if len(line) > 0 {
			var raw map[string]any
			if decodeErr := json.Unmarshal([]byte(line), &raw); decodeErr != nil {
				stats.Malformed++
				if warnOut != nil {
					_, _ = fmt.Fprintf(warnOut, "warning: failed to decode JSON line in %s: %v\n", filePath, decodeErr)
				}
			} else {
				stats.Decoded++
				if cbErr := callback(raw); cbErr != nil {
					return stats, cbErr
				}
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	return stats, nil
}
