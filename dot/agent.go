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
			NewAgentSessionCmd(state),
		},
	}
}

// NewAgentSessionCmd constructs the agent session command group.
func NewAgentSessionCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "session",
		Aliases: []string{"s"},
		Usage:   "Manage agent session logs",
		Commands: []*cli.Command{
			{
				Name:  "agy",
				Usage: "Process session end hook for Antigravity (agy)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sessionID := cmd.Args().Get(0)
					cwd := cmd.Args().Get(1)
					return RunAgentSessionLogAgy(ctx, state, sessionID, cwd)
				},
			},
			{
				Name:  "claude",
				Usage: "Process session end/stop hook for Claude Code",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sessionID := cmd.Args().Get(0)
					cwd := cmd.Args().Get(1)
					return RunAgentSessionLogClaude(ctx, state, sessionID, cwd)
				},
			},
			{
				Name:  "codex",
				Usage: "Process session hook for OpenAI Codex",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sessionID := cmd.Args().Get(0)
					cwd := cmd.Args().Get(1)
					return RunAgentSessionLogCodex(ctx, state, sessionID, cwd)
				},
			},
			{
				Name:  "opencode",
				Usage: "Process session hook for OpenCode",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sessionID := cmd.Args().Get(0)
					cwd := cmd.Args().Get(1)
					return RunAgentSessionLogOpencode(ctx, state, sessionID, cwd)
				},
			},
			{
				Name:  "copilot",
				Usage: "Process a GitHub Copilot CLI session from its session store",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sessionID := cmd.Args().Get(0)
					cwd := cmd.Args().Get(1)
					return RunAgentSessionLogCopilot(ctx, state, sessionID, cwd)
				},
			},
			NewAgentSessionSyncCmd(state),
			NewAgentSessionCleanCmd(state),
		},
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

// NewAgentSessionCleanCmd constructs the agent session clean command.
func NewAgentSessionCleanCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "clean",
		Aliases: []string{"c"},
		Usage:   "Clean up session logs older than 30 days",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "days",
				Aliases: []string{"d"},
				Value:   30,
				Usage:   "Number of days of history to keep",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			days := cmd.Int("days")
			return RunAgentSessionClean(ctx, state, int(days))
		},
	}
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

// getOutputPath resolves target file under ~/.agents/sessions/YYYY-MM-DD/HHMMSS_agent_session.jsonl
func getOutputPath(agent, sessionID string, sessionTime time.Time) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sessionsDir := filepath.Join(home, ".agents", "sessions")
	datePart := sessionTime.Format("2006-01-02")
	timePart := sessionTime.Format("150405")
	dir := filepath.Join(sessionsDir, datePart)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	// Session transcripts can contain source code, credentials, and private prompts.
	// Tighten directories created by older dot versions as they are encountered.
	for _, path := range []string{sessionsDir, dir} {
		if err := os.Chmod(path, 0o700); err != nil {
			return "", fmt.Errorf("failed to secure session directory %s: %w", path, err)
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s_%s_%s.jsonl", timePart, agent, sessionID)), nil
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

func flushSessionLog(writer *bufio.Writer, outPath string) error {
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush session log %s: %w", outPath, err)
	}
	return nil
}

func closeSessionLog(closer io.Closer, outPath string) error {
	if err := closer.Close(); err != nil {
		return fmt.Errorf("failed to close session log %s: %w", outPath, err)
	}
	return nil
}

// writeSessionLogs marshals and writes log lines to the session file.
func writeSessionLogs(agent, sessionID string, logs []SessionLogLine) (resultErr error) {
	if len(logs) == 0 {
		return nil
	}

	// Propagate model name across all lines in the session
	var activeModel string
	for i := 0; i < len(logs); i++ {
		if logs[i].Model != "" {
			activeModel = logs[i].Model
		} else if activeModel != "" {
			logs[i].Model = activeModel
		}
	}
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i].Model != "" {
			activeModel = logs[i].Model
		} else if activeModel != "" {
			logs[i].Model = activeModel
		}
	}

	sessionTime := time.Now().UTC()
	if len(logs) > 0 && logs[0].TS != "" {
		if t, err := time.Parse(time.RFC3339Nano, logs[0].TS); err == nil {
			sessionTime = t.UTC()
		} else if t, err := time.Parse(time.RFC3339, logs[0].TS); err == nil {
			sessionTime = t.UTC()
		}
	}

	outPath, err := getOutputPath(agent, sessionID, sessionTime)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := closeSessionLog(file, outPath); closeErr != nil {
			resultErr = errors.Join(resultErr, closeErr)
		}
	}()
	// OpenFile does not change the mode of an existing file, so explicitly repair
	// transcripts written with the historical 0644 permissions.
	if err := file.Chmod(0o600); err != nil {
		return fmt.Errorf("failed to secure session log %s: %w", outPath, err)
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate session log %s: %w", outPath, err)
	}

	writer := bufio.NewWriter(file)

	for _, log := range logs {
		data, err := json.Marshal(log)
		if err != nil {
			return fmt.Errorf("failed to encode session log %s: %w", outPath, err)
		}
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to buffer session log %s: %w", outPath, err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("failed to buffer session log %s: %w", outPath, err)
		}
	}
	return flushSessionLog(writer, outPath)
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

func stringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func mapValue(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func textFromCodexContent(value interface{}) string {
	switch content := value.(type) {
	case string:
		return content
	case []interface{}:
		textParts := make([]string, 0, len(content))
		for _, part := range content {
			switch partValue := part.(type) {
			case string:
				textParts = append(textParts, partValue)
			case map[string]interface{}:
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

func codexRole(raw map[string]interface{}) string {
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

func codexContent(raw map[string]interface{}) string {
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

func codexModel(raw map[string]interface{}) string {
	if model := stringValue(raw["model"]); model != "" {
		return model
	}
	if payload := mapValue(raw["payload"]); payload != nil {
		return stringValue(payload["model"])
	}
	return ""
}

func codexCWD(raw map[string]interface{}) string {
	if cwd := stringValue(raw["cwd"]); cwd != "" {
		return cwd
	}
	if payload := mapValue(raw["payload"]); payload != nil {
		return stringValue(payload["cwd"])
	}
	return ""
}

// RunAgentSessionLogAgy reads the agy transcript files and processes the session.
func RunAgentSessionLogAgy(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, inputErr := parseStdin(state.Stdin)
	if inputErr != nil {
		return inputErr
	}
	isHookCall := stdinInput != nil
	if stdinInput != nil {
		if stdinInput.StopHookActive || !stdinInput.FullyIdle {
			return writeAntigravityStopDecision(state.Stdout)
		}
		if sessionID == "" {
			sessionID = stdinInput.SessionID
		}
		if cwd == "" {
			cwd = stdinInput.CWD
		}
	}

	if sessionID == "" {
		return errors.New("missing session_id")
	}
	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	transcriptPath := ""
	if stdinInput != nil && stdinInput.TranscriptPath != "" {
		transcriptPath = ExpandPath(stdinInput.TranscriptPath)
		if _, statErr := os.Stat(transcriptPath); statErr != nil {
			return fmt.Errorf("antigravity transcript from hook payload is unavailable at %s: %w", transcriptPath, statErr)
		}
	}
	if transcriptPath == "" {
		transcriptPath = filepath.Join(home, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript_full.jsonl")
		if _, statErr := os.Stat(transcriptPath); os.IsNotExist(statErr) {
			transcriptPath = filepath.Join(home, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript.jsonl")
			if _, statErr2 := os.Stat(transcriptPath); statErr2 != nil {
				return fmt.Errorf("transcript file not found for agy session %s", sessionID)
			}
		}
	}

	cwd = resolveCWD(cwd)

	file, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	decodeErr := decodeJSONL(state.Stderr, transcriptPath, file, func(raw map[string]interface{}) error {
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

	if err := writeSessionLogs("agy", sessionID, logs); err != nil {
		return err
	}
	if isHookCall {
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

// RunHookClaude reads the Claude JSONL files and processes the session.
// RunAgentSessionLogClaude reads the Claude JSONL files and processes the session.
func RunAgentSessionLogClaude(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, inputErr := parseStdin(state.Stdin)
	if inputErr != nil {
		return inputErr
	}
	if stdinInput != nil {
		if stdinInput.StopHookActive {
			return nil
		}
		if sessionID == "" {
			sessionID = stdinInput.SessionID
		}
		if cwd == "" {
			cwd = stdinInput.CWD
		}
	}

	if sessionID == "" {
		return errors.New("missing session_id")
	}
	cwd = resolveCWD(cwd)

	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var sessionFile string
	if stdinInput != nil && stdinInput.TranscriptPath != "" {
		sessionFile = ExpandPath(stdinInput.TranscriptPath)
		info, statErr := os.Stat(sessionFile)
		if statErr != nil {
			return fmt.Errorf("claude transcript from hook payload is unavailable at %s: %w", sessionFile, statErr)
		}
		if info.IsDir() {
			return fmt.Errorf("claude transcript from hook payload is not a file: %s", sessionFile)
		}
	} else {
		encodedCWD := cwd
		encodedCWD = strings.ReplaceAll(encodedCWD, "/", "-")
		encodedCWD = strings.ReplaceAll(encodedCWD, ".", "-")
		encodedCWD = strings.TrimPrefix(encodedCWD, "-")
		encodedCWD = "-" + encodedCWD
		sessionFile = filepath.Join(home, ".claude", "projects", encodedCWD, sessionID+".jsonl")

		_, statErr := os.Stat(sessionFile)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("failed to inspect expected Claude transcript %s: %w", sessionFile, statErr)
		}
		if statErr == nil {
			// The CWD-derived transcript path is exact, so no fallback scan is needed.
		} else {
			projectsDir := filepath.Join(home, ".claude", "projects")
			found, findErr := findSessionFile(projectsDir, func(_ string, entry fs.DirEntry) bool {
				return entry.Name() == sessionID+".jsonl"
			})
			if findErr != nil {
				return fmt.Errorf("failed to search Claude transcripts in %s: %w", projectsDir, findErr)
			}
			if found != "" {
				sessionFile = found
			} else {
				return fmt.Errorf("session file not found for claude session %s", sessionID)
			}
		}
	}

	file, err := os.Open(sessionFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	decodeErr := decodeJSONL(state.Stderr, sessionFile, file, func(raw map[string]interface{}) error {
		typ, _ := raw["type"].(string)
		if typ != "user" && typ != "assistant" {
			return nil
		}

		ts, _ := raw["timestamp"].(string)
		msgVal, ok := raw["message"].(map[string]interface{})
		if !ok {
			return nil
		}

		var content string
		switch typ {
		case "user":
			content, _ = msgVal["content"].(string)
		case "assistant":
			if contentsList, ok := msgVal["content"].([]interface{}); ok {
				var textParts []string
				for _, part := range contentsList {
					if partMap, ok := part.(map[string]interface{}); ok {
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

	return writeSessionLogs("claude", sessionID, logs)
}

// RunAgentSessionLogCodex reads Codex rollout session files and processes the session.
func RunAgentSessionLogCodex(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, inputErr := parseStdin(state.Stdin)
	if inputErr != nil {
		return inputErr
	}
	if stdinInput != nil {
		if stdinInput.StopHookActive {
			return nil
		}
		if sessionID == "" {
			sessionID = stdinInput.SessionID
		}
		if cwd == "" {
			cwd = stdinInput.CWD
		}
	}

	if sessionID == "" {
		return errors.New("missing session_id")
	}
	cwd = resolveCWD(cwd)

	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var transcriptPath string
	if stdinInput != nil && stdinInput.TranscriptPath != "" {
		transcriptPath = ExpandPath(stdinInput.TranscriptPath)
		info, statErr := os.Stat(transcriptPath)
		if statErr != nil {
			return fmt.Errorf("codex transcript from hook payload is unavailable at %s: %w", transcriptPath, statErr)
		}
		if info.IsDir() {
			return fmt.Errorf("codex transcript from hook payload is not a file: %s", transcriptPath)
		}
	}

	if transcriptPath == "" {
		sessionsDir := filepath.Join(home, ".codex", "sessions")
		found, findErr := findSessionFile(sessionsDir, func(_ string, entry fs.DirEntry) bool {
			if !strings.HasSuffix(entry.Name(), ".jsonl") {
				return false
			}
			name := strings.TrimSuffix(entry.Name(), ".jsonl")
			return extractCodexSessionID(name) == sessionID
		})
		if findErr != nil {
			return fmt.Errorf("failed to search Codex transcripts in %s: %w", sessionsDir, findErr)
		}
		if found != "" {
			transcriptPath = found
		} else {
			return fmt.Errorf("session file not found for codex session %s", sessionID)
		}
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	activeModel := ""
	activeCWD := cwd
	decodeErr := decodeJSONL(state.Stderr, transcriptPath, file, func(raw map[string]interface{}) error {
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
		if ts == "" {
			ts = time.Now().UTC().Format(time.RFC3339)
		}

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

	return writeSessionLogs("codex", sessionID, logs)
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

// decodeSessionIDs parses a sqlite3 -json `SELECT id ...` result into the set of
// not-yet-processed session IDs. Shared by the OpenCode and Copilot DB scanners;
// label names the agent in error messages.
func decodeSessionIDs(label, output string, processed map[string]bool) ([]string, error) {
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
		if !processed[sessionID] {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	return sessionIDs, nil
}

func syncOpencodeSessions(ctx context.Context, state *GlobalState, dbPath string, processed map[string]bool) (int, error) {
	output, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, opencodeSessionsQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query OpenCode sessions: %w", err)
	}

	sessionIDs, err := decodeSessionIDs("OpenCode", output, processed)
	if err != nil {
		return 0, err
	}
	if len(sessionIDs) == 0 {
		return 0, nil
	}

	quotedSessionIDs := make([]string, len(sessionIDs))
	requested := make(map[string]bool, len(sessionIDs))
	for i, sessionID := range sessionIDs {
		// Session IDs were validated above, so the exact value is safe to use as a
		// SQLite string literal without silently changing its identity.
		quotedSessionIDs[i] = "'" + sessionID + "'"
		requested[sessionID] = true
	}
	sqlQuery := fmt.Sprintf(opencodeMessagesQuery, "m.session_id IN ("+strings.Join(quotedSessionIDs, ",")+")")
	messageOutput, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, sqlQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query OpenCode messages: %w", err)
	}

	messageOutput = strings.TrimSpace(messageOutput)
	if messageOutput == "" || messageOutput == "[]" {
		return 0, nil
	}

	var rows []OpencodeRow
	if err := json.Unmarshal([]byte(messageOutput), &rows); err != nil {
		return 0, fmt.Errorf("failed to decode OpenCode message query result: %w", err)
	}
	if rows == nil {
		return 0, errors.New("failed to decode OpenCode message query result: expected a JSON array")
	}

	sessionRows := make(map[string][]OpencodeRow, len(sessionIDs))
	for rowNumber, row := range rows {
		if !isValidSessionID(row.SessionID) {
			return 0, fmt.Errorf("malformed OpenCode message row %d: invalid session ID %q", rowNumber+1, row.SessionID)
		}
		if !requested[row.SessionID] {
			return 0, fmt.Errorf("malformed OpenCode message row %d: unexpected session ID %q", rowNumber+1, row.SessionID)
		}
		if row.MessageID == "" {
			return 0, fmt.Errorf("malformed OpenCode message row %d: missing message ID", rowNumber+1)
		}
		if row.PartData != "" && row.PartID == "" {
			return 0, fmt.Errorf("malformed OpenCode message row %d: missing part ID", rowNumber+1)
		}
		sessionRows[row.SessionID] = append(sessionRows[row.SessionID], row)
	}

	count := 0
	for _, sessionID := range sessionIDs {
		logs, parseErr := parseOpencodeRows(sessionID, "", sessionRows[sessionID])
		if parseErr != nil {
			return 0, fmt.Errorf("failed to parse OpenCode session %q: %w", sessionID, parseErr)
		}
		if len(logs) == 0 {
			continue
		}
		if writeErr := writeSessionLogs("opencode", sessionID, logs); writeErr != nil {
			return 0, fmt.Errorf("failed to write OpenCode session %q: %w", sessionID, writeErr)
		}
		count++
		_, _ = fmt.Fprint(state.Stderr, ".")
	}
	return count, nil
}

// RunAgentSessionLogOpencode reads OpenCode session records and writes them to the sessions directory.
func RunAgentSessionLogOpencode(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, inputErr := parseStdin(state.Stdin)
	if inputErr != nil {
		return inputErr
	}
	if stdinInput != nil {
		if stdinInput.StopHookActive {
			return nil
		}
		if sessionID == "" {
			sessionID = stdinInput.SessionID
		}
		if cwd == "" {
			cwd = stdinInput.CWD
		}
	}

	if sessionID == "" {
		return errors.New("missing session_id")
	}
	cwd = resolveCWD(cwd)

	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return fmt.Errorf("opencode database not found at %s", dbPath)
	}

	sqlQuery := fmt.Sprintf(opencodeMessagesQuery, "m.session_id = '"+sessionID+"'")
	out, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, sqlQuery)
	if err != nil {
		return err
	}

	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		return nil
	}

	var rows []OpencodeRow
	if parseErr := json.Unmarshal([]byte(out), &rows); parseErr != nil {
		return fmt.Errorf("failed to parse OpenCode query result: %w", parseErr)
	}

	logs, err := parseOpencodeRows(sessionID, cwd, rows)
	if err != nil {
		return err
	}

	return writeSessionLogs("opencode", sessionID, logs)
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

// copilotDBPath returns the path to the Copilot CLI session store.
func copilotDBPath(home string) string {
	return filepath.Join(home, ".copilot", "session-store.db")
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
// Copilot exposes no session-end hook, so this is invoked with an explicit session ID
// (by the sync scanner or manually) rather than from a piped hook payload.
func RunAgentSessionLogCopilot(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	if sessionID == "" {
		return errors.New("missing session_id")
	}
	if !isValidSessionID(sessionID) {
		return fmt.Errorf("invalid session_id format: %q", sessionID)
	}
	cwd = resolveCWD(cwd)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dbPath := copilotDBPath(home)
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return fmt.Errorf("copilot database not found at %s", dbPath)
	}

	sqlQuery := fmt.Sprintf(copilotTurnsQuery, "t.session_id = '"+sessionID+"'")
	out, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, sqlQuery)
	if err != nil {
		return err
	}

	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		return nil
	}

	var rows []CopilotRow
	if parseErr := json.Unmarshal([]byte(out), &rows); parseErr != nil {
		return fmt.Errorf("failed to parse Copilot query result: %w", parseErr)
	}

	return writeSessionLogs("copilot", sessionID, parseCopilotRows(sessionID, cwd, rows))
}

// syncCopilotSessions scans the Copilot store and logs every untracked session.
func syncCopilotSessions(ctx context.Context, state *GlobalState, dbPath string, processed map[string]bool) (int, error) {
	output, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, copilotSessionsQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query Copilot sessions: %w", err)
	}

	sessionIDs, err := decodeSessionIDs("Copilot", output, processed)
	if err != nil {
		return 0, err
	}
	if len(sessionIDs) == 0 {
		return 0, nil
	}

	quotedSessionIDs := make([]string, len(sessionIDs))
	requested := make(map[string]bool, len(sessionIDs))
	for i, sessionID := range sessionIDs {
		// Session IDs were validated above, so the exact value is safe as a SQLite literal.
		quotedSessionIDs[i] = "'" + sessionID + "'"
		requested[sessionID] = true
	}
	sqlQuery := fmt.Sprintf(copilotTurnsQuery, "t.session_id IN ("+strings.Join(quotedSessionIDs, ",")+")")
	turnOutput, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", os.DevNull, "-json", dbPath, sqlQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query Copilot turns: %w", err)
	}

	turnOutput = strings.TrimSpace(turnOutput)
	if turnOutput == "" || turnOutput == "[]" {
		return 0, nil
	}

	var rows []CopilotRow
	if err := json.Unmarshal([]byte(turnOutput), &rows); err != nil {
		return 0, fmt.Errorf("failed to decode Copilot turn query result: %w", err)
	}
	if rows == nil {
		return 0, errors.New("failed to decode Copilot turn query result: expected a JSON array")
	}

	sessionRows := make(map[string][]CopilotRow, len(sessionIDs))
	for rowNumber, row := range rows {
		if !isValidSessionID(row.SessionID) {
			return 0, fmt.Errorf("malformed Copilot turn row %d: invalid session ID %q", rowNumber+1, row.SessionID)
		}
		if !requested[row.SessionID] {
			return 0, fmt.Errorf("malformed Copilot turn row %d: unexpected session ID %q", rowNumber+1, row.SessionID)
		}
		sessionRows[row.SessionID] = append(sessionRows[row.SessionID], row)
	}

	count := 0
	for _, sessionID := range sessionIDs {
		logs := parseCopilotRows(sessionID, "", sessionRows[sessionID])
		if len(logs) == 0 {
			continue
		}
		if writeErr := writeSessionLogs("copilot", sessionID, logs); writeErr != nil {
			return 0, fmt.Errorf("failed to write Copilot session %q: %w", sessionID, writeErr)
		}
		count++
		_, _ = fmt.Fprint(state.Stderr, ".")
	}
	return count, nil
}

// RunAgentSessionSync scans all agent storage and triggers logging for unprocessed sessions.
func RunAgentSessionSync(ctx context.Context, state *GlobalState) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sessionsDir := filepath.Join(home, ".agents", "sessions")

	// The per-agent log helpers double as hook entrypoints and parse state.Stdin.
	// When invoked from sync there is no hook payload, so detach stdin to avoid
	// consuming (and choking on) whatever the sync process was started with.
	noStdinState := *state
	noStdinState.Stdin = nil

	processed := make(map[string]bool)
	processedWalkErr := filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
			name := strings.TrimSuffix(d.Name(), ".jsonl")
			parts := strings.Split(name, "_")
			if len(parts) >= 3 {
				sid := strings.Join(parts[2:], "_")
				processed[sid] = true
			}
		}
		return nil
	})
	if processedWalkErr != nil && !errors.Is(processedWalkErr, os.ErrNotExist) {
		return fmt.Errorf("failed to scan processed session logs: %w", processedWalkErr)
	}

	total := 0

	// 1. agy
	brainDir := filepath.Join(home, ".gemini", "antigravity-cli", "brain")
	hasAgySessions, err := sourceDirectoryExists(brainDir, "agy")
	if err != nil {
		return err
	}
	if hasAgySessions {
		count := 0
		entries, readErr := os.ReadDir(brainDir)
		if readErr != nil {
			return fmt.Errorf("failed to read agy sessions: %w", readErr)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				sid := entry.Name()
				transPath := filepath.Join(brainDir, sid, ".system_generated", "logs", "transcript_full.jsonl")
				transPath2 := filepath.Join(brainDir, sid, ".system_generated", "logs", "transcript.jsonl")
				if _, statErr := os.Stat(transPath); statErr != nil {
					if !errors.Is(statErr, os.ErrNotExist) {
						return fmt.Errorf("failed to inspect agy transcript %s: %w", transPath, statErr)
					}
					if _, fallbackErr := os.Stat(transPath2); fallbackErr != nil {
						if !errors.Is(fallbackErr, os.ErrNotExist) {
							return fmt.Errorf("failed to inspect agy transcript %s: %w", transPath2, fallbackErr)
						}
						continue
					}
				}
				if !processed[sid] {
					if logErr := RunAgentSessionLogAgy(ctx, &noStdinState, sid, ""); logErr != nil {
						return fmt.Errorf("failed to sync agy session %s: %w", sid, logErr)
					}
					count++
					_, _ = fmt.Fprint(state.Stderr, ".")
				}
			}
		}
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "agy: %d new\n", count)
		total += count
	}

	// 2. claude
	projectsDir := filepath.Join(home, ".claude", "projects")
	hasClaudeSessions, err := sourceDirectoryExists(projectsDir, "Claude")
	if err != nil {
		return err
	}
	if hasClaudeSessions {
		count := 0
		walkErr := filepath.WalkDir(projectsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") && d.Name() != "memory.jsonl" {
				sid := strings.TrimSuffix(d.Name(), ".jsonl")
				if !processed[sid] {
					if logErr := RunAgentSessionLogClaude(ctx, &noStdinState, sid, ""); logErr != nil {
						return fmt.Errorf("failed to sync Claude session %s: %w", sid, logErr)
					}
					count++
					_, _ = fmt.Fprint(state.Stderr, ".")
				}
			}
			return nil
		})
		if walkErr != nil {
			return fmt.Errorf("failed to scan Claude sessions: %w", walkErr)
		}
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "claude: %d new\n", count)
		total += count
	}

	// 3. codex
	codexDir := filepath.Join(home, ".codex", "sessions")
	hasCodexSessions, err := sourceDirectoryExists(codexDir, "Codex")
	if err != nil {
		return err
	}
	if hasCodexSessions {
		count := 0
		walkErr := filepath.WalkDir(codexDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
				name := strings.TrimSuffix(d.Name(), ".jsonl")
				sid := extractCodexSessionID(name)
				if sid != "" && !processed[sid] {
					if logErr := RunAgentSessionLogCodex(ctx, &noStdinState, sid, ""); logErr != nil {
						return fmt.Errorf("failed to sync Codex session %s: %w", sid, logErr)
					}
					count++
					_, _ = fmt.Fprint(state.Stderr, ".")
				}
			}
			return nil
		})
		if walkErr != nil {
			return fmt.Errorf("failed to scan Codex sessions: %w", walkErr)
		}
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "codex: %d new\n", count)
		total += count
	}

	// 4. opencode
	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	info, statErr := os.Stat(dbPath)
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("failed to inspect OpenCode database %s: %w", dbPath, statErr)
	}
	if statErr == nil {
		if info.IsDir() {
			return fmt.Errorf("OpenCode database path is a directory: %s", dbPath)
		}
		count, syncErr := syncOpencodeSessions(ctx, state, dbPath, processed)
		if syncErr != nil {
			return syncErr
		}
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "opencode: %d new\n", count)
		total += count
	}

	// 5. copilot (no live hook API; captured from its session store)
	copilotDB := copilotDBPath(home)
	copilotInfo, copilotStatErr := os.Stat(copilotDB)
	if copilotStatErr != nil && !errors.Is(copilotStatErr, os.ErrNotExist) {
		return fmt.Errorf("failed to inspect Copilot database %s: %w", copilotDB, copilotStatErr)
	}
	if copilotStatErr == nil {
		if copilotInfo.IsDir() {
			return fmt.Errorf("copilot database path is a directory: %s", copilotDB)
		}
		count, syncErr := syncCopilotSessions(ctx, state, copilotDB, processed)
		if syncErr != nil {
			return syncErr
		}
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "copilot: %d new\n", count)
		total += count
	}

	_, _ = fmt.Fprintf(state.Stderr, "agent-session-sync: done (%d total new)\n", total)
	return nil
}

// RunAgentSessionClean deletes session logs older than N days.
func RunAgentSessionClean(ctx context.Context, state *GlobalState, days int) error {
	if days <= 0 {
		return errors.New("retention days must be greater than zero")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sessionsDir := filepath.Join(home, ".agents", "sessions")

	if _, statErr := os.Stat(sessionsDir); os.IsNotExist(statErr) {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	deletedFiles := 0
	var dirs []string

	err = filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != sessionsDir {
				dirs = append(dirs, path)
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), ".jsonl") {
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("failed to inspect session log %s: %w", path, err)
			}

			shouldDelete := false
			dirName := filepath.Base(filepath.Dir(path))
			if t, err := time.Parse("2006-01-02", dirName); err == nil {
				if t.Before(cutoff) {
					shouldDelete = true
				}
			} else {
				// Fallback to modtime
				if info.ModTime().Before(cutoff) {
					shouldDelete = true
				}
			}

			if shouldDelete {
				//nolint:gosec // G122: path is walked from trusted sessions directory
				if err := os.Remove(path); err != nil {
					return fmt.Errorf("failed to remove expired session log %s: %w", path, err)
				}
				deletedFiles++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Remove empty directories in reverse order (deepest first)
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, readErr := os.ReadDir(dirs[i])
		if readErr != nil {
			return fmt.Errorf("failed to inspect session directory %s: %w", dirs[i], readErr)
		}
		if len(entries) > 0 {
			continue
		}
		//nolint:gosec // G122: path is walked from trusted sessions directory
		if removeErr := os.Remove(dirs[i]); removeErr != nil {
			return fmt.Errorf("failed to remove empty session directory %s: %w", dirs[i], removeErr)
		}
	}

	_, _ = fmt.Fprintf(state.Stderr, "agent-session-clean: deleted %d file(s) older than %d days\n", deletedFiles, days)
	return nil
}

// decodeJSONL reads a line-delimited JSON file robustly and calls the callback for each line.
// If a line cannot be parsed as JSON, it prints a warning to warnOut and continues.
func decodeJSONL(warnOut io.Writer, filePath string, file *os.File, callback func(raw map[string]interface{}) error) error {
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading file %s: %w", filePath, err)
		}
		if len(line) > 0 {
			var raw map[string]interface{}
			if decodeErr := json.Unmarshal([]byte(line), &raw); decodeErr != nil {
				if warnOut != nil {
					_, _ = fmt.Fprintf(warnOut, "warning: failed to decode JSON line in %s: %v\n", filePath, decodeErr)
				}
			} else {
				if cbErr := callback(raw); cbErr != nil {
					return cbErr
				}
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	return nil
}
