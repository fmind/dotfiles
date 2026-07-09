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

// HookInput represents the JSON context passed by agy or claude on stdin.
type HookInput struct {
	SessionID      string `json:"session_id"`
	CWD            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path"`
	StopHookActive bool   `json:"stop_hook_active"`
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

// parseStdin reads stdin to extract HookInput if piped.
func parseStdin() (*HookInput, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Read all stdin to avoid blocking if the stream is empty or terminated
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			var input HookInput
			if err := json.Unmarshal(data, &input); err == nil {
				return &input, nil
			}
		}
	}
	return nil, nil
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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
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

// writeSessionLogs marshals and writes log lines to the session file.
func writeSessionLogs(agent, sessionID string, logs []SessionLogLine) error {
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
	file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)
	defer func() { _ = writer.Flush() }()

	for _, log := range logs {
		data, err := json.Marshal(log)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	return nil
}

// isValidSessionRune checks if a rune is allowed in a session ID.
func isValidSessionRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
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
	stdinInput, _ := parseStdin()
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
	// Sanitize sessionID to prevent SQL injection or directory traversal
	for _, r := range sessionID {
		if !isValidSessionRune(r) {
			return fmt.Errorf("invalid session_id format: %q", sessionID)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	transcriptPath := filepath.Join(home, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript_full.jsonl")
	if _, statErr := os.Stat(transcriptPath); os.IsNotExist(statErr) {
		transcriptPath = filepath.Join(home, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript.jsonl")
		if _, statErr2 := os.Stat(transcriptPath); statErr2 != nil {
			return fmt.Errorf("transcript file not found for agy session %s", sessionID)
		}
	}

	cwd = resolveCWD(cwd)

	file, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	dec := json.NewDecoder(file)
	for {
		var raw map[string]interface{}
		if err := dec.Decode(&raw); err != nil {
			break
		}

		if trunc, ok := raw["is_truncated"].(bool); ok && trunc {
			continue
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
			continue
		}

		if strings.TrimSpace(content) == "" {
			continue
		}

		logs = append(logs, SessionLogLine{
			TS:      createdAt,
			Agent:   "agy",
			SID:     sessionID,
			Role:    role,
			Content: content,
			CWD:     cwd,
		})
	}

	return writeSessionLogs("agy", sessionID, logs)
}

// RunHookClaude reads the Claude JSONL files and processes the session.
// RunAgentSessionLogClaude reads the Claude JSONL files and processes the session.
func RunAgentSessionLogClaude(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, _ := parseStdin()
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

	// Sanitize sessionID to prevent SQL injection or directory traversal
	for _, r := range sessionID {
		if !isValidSessionRune(r) {
			return fmt.Errorf("invalid session_id format: %q", sessionID)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	encodedCWD := cwd
	encodedCWD = strings.ReplaceAll(encodedCWD, "/", "-")
	encodedCWD = strings.ReplaceAll(encodedCWD, ".", "-")
	encodedCWD = strings.TrimPrefix(encodedCWD, "-")
	encodedCWD = "-" + encodedCWD
	sessionFile := filepath.Join(home, ".claude", "projects", encodedCWD, sessionID+".jsonl")

	if _, statErr := os.Stat(sessionFile); os.IsNotExist(statErr) {
		projectsDir := filepath.Join(home, ".claude", "projects")
		found := ""
		_ = filepath.WalkDir(projectsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && d.Name() == sessionID+".jsonl" {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			sessionFile = found
		} else {
			return fmt.Errorf("session file not found for claude session %s", sessionID)
		}
	}

	file, err := os.Open(sessionFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var logs []SessionLogLine
	dec := json.NewDecoder(file)
	for {
		var raw map[string]interface{}
		if err := dec.Decode(&raw); err != nil {
			break
		}

		typ, _ := raw["type"].(string)
		if typ != "user" && typ != "assistant" {
			continue
		}

		ts, _ := raw["timestamp"].(string)
		msgVal, ok := raw["message"].(map[string]interface{})
		if !ok {
			continue
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
			continue
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
	}

	return writeSessionLogs("claude", sessionID, logs)
}

// RunAgentSessionLogCodex reads Codex rollout session files and processes the session.
func RunAgentSessionLogCodex(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, _ := parseStdin()
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

	// Sanitize sessionID to prevent SQL injection or directory traversal
	for _, r := range sessionID {
		if !isValidSessionRune(r) {
			return fmt.Errorf("invalid session_id format: %q", sessionID)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var transcriptPath string
	if stdinInput != nil && stdinInput.TranscriptPath != "" {
		if _, statErr := os.Stat(stdinInput.TranscriptPath); statErr == nil {
			transcriptPath = stdinInput.TranscriptPath
		}
	}

	if transcriptPath == "" {
		sessionsDir := filepath.Join(home, ".codex", "sessions")
		found := ""
		_ = filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") && strings.Contains(d.Name(), sessionID) {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
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
	dec := json.NewDecoder(file)
	for {
		var raw map[string]interface{}
		if err := dec.Decode(&raw); err != nil {
			break
		}

		if model := codexModel(raw); model != "" {
			activeModel = model
		}
		if logCWD := codexCWD(raw); logCWD != "" {
			activeCWD = resolveCWD(logCWD)
		}

		role := codexRole(raw)
		if role != "user" && role != "assistant" {
			continue
		}

		content := codexContent(raw)
		if strings.TrimSpace(content) == "" {
			continue
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
	Parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"parts"`
}

// OpencodeRow represents the structure returned by sqlite3 -json for OpenCode.
type OpencodeRow struct {
	Data        string  `json:"data"`
	Directory   string  `json:"directory"`
	TimeCreated float64 `json:"time_created"`
}

// RunAgentSessionLogOpencode reads OpenCode session records and writes them to the sessions directory.
func RunAgentSessionLogOpencode(ctx context.Context, state *GlobalState, sessionID, cwd string) error {
	stdinInput, _ := parseStdin()
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

	// Sanitize sessionID to prevent SQL injection or directory traversal
	for _, r := range sessionID {
		if !isValidSessionRune(r) {
			return fmt.Errorf("invalid session_id format: %q", sessionID)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return fmt.Errorf("opencode database not found at %s", dbPath)
	}

	sqlQuery := fmt.Sprintf("SELECT m.time_created, m.data, s.directory FROM message m JOIN session s ON m.session_id = s.id WHERE m.session_id = '%s' ORDER BY m.time_created", sessionID)
	out, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", "/dev/null", "-json", dbPath, sqlQuery)
	if err != nil {
		return err
	}

	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		return nil
	}

	var rows []OpencodeRow
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return err
	}

	var logs []SessionLogLine
	for _, row := range rows {
		var msgData OpencodeData
		if err := json.Unmarshal([]byte(row.Data), &msgData); err != nil {
			continue
		}

		if msgData.Role != "user" && msgData.Role != "assistant" {
			continue
		}

		var textParts []string
		for _, part := range msgData.Parts {
			if part.Type == "text" && part.Text != "" {
				textParts = append(textParts, part.Text)
			}
		}
		content := strings.Join(textParts, "\n")

		// Convert millisecond epoch to RFC3339 string
		t := time.Unix(0, int64(row.TimeCreated)*int64(time.Millisecond)).UTC()
		ts := t.Format(time.RFC3339)

		logCWD := row.Directory
		if logCWD == "" {
			logCWD = cwd
		}
		logCWD = resolveCWD(logCWD)

		var model string
		if msgData.Model.ModelID != "" {
			if msgData.Model.ProviderID != "" {
				model = msgData.Model.ProviderID + "/" + msgData.Model.ModelID
			} else {
				model = msgData.Model.ModelID
			}
		}

		if strings.TrimSpace(content) == "" {
			continue
		}

		logs = append(logs, SessionLogLine{
			TS:      ts,
			Agent:   "opencode",
			SID:     sessionID,
			Role:    msgData.Role,
			Content: content,
			CWD:     logCWD,
			Model:   model,
		})
	}

	return writeSessionLogs("opencode", sessionID, logs)
}

// RunAgentSessionSync scans all agent storage and triggers logging for unprocessed sessions.
func RunAgentSessionSync(ctx context.Context, state *GlobalState) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sessionsDir := filepath.Join(home, ".agents", "sessions")

	processed := make(map[string]bool)
	_ = filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
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

	total := 0

	// 1. agy
	brainDir := filepath.Join(home, ".gemini", "antigravity-cli", "brain")
	if info, err := os.Stat(brainDir); err == nil && info.IsDir() {
		count := 0
		entries, _ := os.ReadDir(brainDir)
		for _, entry := range entries {
			if entry.IsDir() {
				sid := entry.Name()
				transPath := filepath.Join(brainDir, sid, ".system_generated", "logs", "transcript_full.jsonl")
				transPath2 := filepath.Join(brainDir, sid, ".system_generated", "logs", "transcript.jsonl")
				if _, err1 := os.Stat(transPath); err1 != nil {
					if _, err2 := os.Stat(transPath2); err2 != nil {
						continue
					}
				}
				if !processed[sid] {
					if err := RunAgentSessionLogAgy(ctx, state, sid, ""); err == nil {
						count++
						_, _ = fmt.Fprint(state.Stderr, ".")
					}
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
	if info, err := os.Stat(projectsDir); err == nil && info.IsDir() {
		count := 0
		_ = filepath.WalkDir(projectsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") && d.Name() != "memory.jsonl" {
				sid := strings.TrimSuffix(d.Name(), ".jsonl")
				if !processed[sid] {
					cwd := ""
					//nolint:gosec // G122: path is walked from trusted projects directory
					if file, openErr := os.Open(path); openErr == nil {
						dec := json.NewDecoder(file)
						for {
							var line map[string]interface{}
							if err := dec.Decode(&line); err != nil {
								break
							}
							if logCWD, ok := line["cwd"].(string); ok && logCWD != "" {
								cwd = logCWD
								break
							}
						}
						_ = file.Close()
					}
					if err := RunAgentSessionLogClaude(ctx, state, sid, cwd); err == nil {
						count++
						_, _ = fmt.Fprint(state.Stderr, ".")
					}
				}
			}
			return nil
		})
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "claude: %d new\n", count)
		total += count
	}

	// 3. codex
	codexDir := filepath.Join(home, ".codex", "sessions")
	if info, err := os.Stat(codexDir); err == nil && info.IsDir() {
		count := 0
		_ = filepath.WalkDir(codexDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
				name := strings.TrimSuffix(d.Name(), ".jsonl")
				sid := extractCodexSessionID(name)
				if sid != "" && !processed[sid] {
					cwd := ""
					// #nosec G122
					if file, openErr := os.Open(path); openErr == nil {
						dec := json.NewDecoder(file)
						var firstLine map[string]interface{}
						if err := dec.Decode(&firstLine); err == nil {
							if logCWD, ok := firstLine["cwd"].(string); ok && logCWD != "" {
								cwd = logCWD
							}
						}
						_ = file.Close()
					}
					if err := RunAgentSessionLogCodex(ctx, state, sid, cwd); err == nil {
						count++
						_, _ = fmt.Fprint(state.Stderr, ".")
					}
				}
			}
			return nil
		})
		if count > 0 {
			_, _ = fmt.Fprintln(state.Stderr)
		}
		_, _ = fmt.Fprintf(state.Stderr, "codex: %d new\n", count)
		total += count
	}

	// 4. opencode
	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	if info, err := os.Stat(dbPath); err == nil && !info.IsDir() {
		out, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", "/dev/null", "-list", "-noheader", dbPath, "SELECT id, directory FROM session")
		if err == nil {
			var unprocessedSIDs []string
			sidDirs := make(map[string]string)
			lines := strings.Split(out, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, "|", 2)
				if len(parts) >= 1 {
					sid := parts[0]
					if sid == "" {
						continue
					}
					if !processed[sid] {
						unprocessedSIDs = append(unprocessedSIDs, sid)
						if len(parts) >= 2 {
							sidDirs[sid] = parts[1]
						}
					}
				}
			}

			if len(unprocessedSIDs) > 0 {
				escapedSIDs := make([]string, len(unprocessedSIDs))
				for i, sid := range unprocessedSIDs {
					var sb strings.Builder
					for _, r := range sid {
						if isValidSessionRune(r) {
							sb.WriteRune(r)
						}
					}
					escapedSIDs[i] = fmt.Sprintf("'%s'", sb.String())
				}
				sqlQuery := fmt.Sprintf("SELECT m.session_id, m.time_created, m.data, s.directory FROM message m JOIN session s ON m.session_id = s.id WHERE m.session_id IN (%s) ORDER BY m.session_id, m.time_created", strings.Join(escapedSIDs, ","))
				msgOut, err := state.Runner.Run(ctx, "", nil, "sqlite3", "-init", "/dev/null", "-json", dbPath, sqlQuery)
				if err == nil {
					msgOut = strings.TrimSpace(msgOut)
					if msgOut != "" && msgOut != "[]" {
						var rows []struct {
							SessionID   string  `json:"session_id"`
							Data        string  `json:"data"`
							Directory   string  `json:"directory"`
							TimeCreated float64 `json:"time_created"`
						}
						if err := json.Unmarshal([]byte(msgOut), &rows); err == nil {
							sessionLogs := make(map[string][]SessionLogLine)
							for _, row := range rows {
								var msgData OpencodeData
								if err := json.Unmarshal([]byte(row.Data), &msgData); err != nil {
									continue
								}
								if msgData.Role != "user" && msgData.Role != "assistant" {
									continue
								}
								var textParts []string
								for _, part := range msgData.Parts {
									if part.Type == "text" && part.Text != "" {
										textParts = append(textParts, part.Text)
									}
								}
								content := strings.Join(textParts, "\n")
								t := time.Unix(0, int64(row.TimeCreated)*int64(time.Millisecond)).UTC()
								ts := t.Format(time.RFC3339)
								logCWD := row.Directory
								if logCWD == "" {
									logCWD = sidDirs[row.SessionID]
								}
								logCWD = resolveCWD(logCWD)
								var model string
								if msgData.Model.ModelID != "" {
									if msgData.Model.ProviderID != "" {
										model = msgData.Model.ProviderID + "/" + msgData.Model.ModelID
									} else {
										model = msgData.Model.ModelID
									}
								}
								if strings.TrimSpace(content) == "" {
									continue
								}
								sessionLogs[row.SessionID] = append(sessionLogs[row.SessionID], SessionLogLine{
									TS:      ts,
									Agent:   "opencode",
									SID:     row.SessionID,
									Role:    msgData.Role,
									Content: content,
									CWD:     logCWD,
									Model:   model,
								})
							}

							count := 0
							for sid, msgLogs := range sessionLogs {
								if err := writeSessionLogs("opencode", sid, msgLogs); err == nil {
									count++
									_, _ = fmt.Fprint(state.Stderr, ".")
								}
							}
							if count > 0 {
								_, _ = fmt.Fprintln(state.Stderr)
							}
							_, _ = fmt.Fprintf(state.Stderr, "opencode: %d new\n", count)
							total += count
						}
					}
				}
			} else {
				_, _ = fmt.Fprintf(state.Stderr, "opencode: 0 new\n")
			}
		}
	}

	_, _ = fmt.Fprintf(state.Stderr, "agent-session-sync: done (%d total new)\n", total)
	return nil
}

// RunAgentSessionClean deletes session logs older than N days.
func RunAgentSessionClean(ctx context.Context, state *GlobalState, days int) error {
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
				return nil
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
				if err := os.Remove(path); err == nil {
					deletedFiles++
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Remove empty directories in reverse order (deepest first)
	for i := len(dirs) - 1; i >= 0; i-- {
		//nolint:gosec // G122: path is walked from trusted sessions directory
		_ = os.Remove(dirs[i]) // Only succeeds if directory is empty
	}

	_, _ = fmt.Fprintf(state.Stderr, "agent-session-clean: deleted %d file(s) older than %d days\n", deletedFiles, days)
	return nil
}
