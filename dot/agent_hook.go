package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/urfave/cli/v3"
)

const (
	hookFailureStoreVersion       = "v1"
	defaultHookFailureLimit       = 100
	defaultHookFailureDetailLimit = 512
)

type hookFailureRecord struct {
	OccurredAt  string `json:"occurred_at"`
	Agent       string `json:"agent"`
	Operation   string `json:"operation"`
	SessionHash string `json:"session_hash,omitempty"`
	Detail      string `json:"detail"`
}

func NewAgentHookCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "hook",
		Aliases: []string{"h"},
		Usage:   "Run an observable agent hook and spool bounded failure metadata",
		Commands: []*cli.Command{
			{
				Name:    "copilot-session-end",
				Aliases: []string{"c"},
				Usage:   "Use Copilot sessionEnd metadata to trigger non-blocking targeted sync",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return RunCopilotSessionEndHook(ctx, state)
				},
			},
			{
				Name:      "session",
				Aliases:   []string{"s"},
				Usage:     "Run a session-ingestion hook",
				ArgsUsage: "<agent> [session-id] [cwd]",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return RunAgentHookSession(ctx, state, cmd.Args().Get(0), cmd.Args().Get(1), cmd.Args().Get(2))
				},
			},
			{
				Name:      "notify",
				Aliases:   []string{"n"},
				Usage:     "Run an agent notification hook",
				ArgsUsage: "<agent> <event>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return RunAgentHookNotify(ctx, state, cmd.Args().Get(0), cmd.Args().Get(1))
				},
			},
		},
	}
}

func RunAgentHookSession(ctx context.Context, state *GlobalState, agent, sessionID, cwd string) error {
	var err error
	switch agent {
	case sessionStoreAgy:
		err = RunAgentSessionLogAgy(ctx, state, sessionID, cwd)
	case sessionStoreClaude:
		err = RunAgentSessionLogClaude(ctx, state, sessionID, cwd)
	case sessionStoreCodex:
		err = RunAgentSessionLogCodex(ctx, state, sessionID, cwd)
	case sessionStoreGrok:
		err = RunAgentSessionLogGrok(ctx, state, sessionID, cwd)
	case sessionStoreOpenCode:
		err = RunAgentSessionLogOpencode(ctx, state, sessionID, cwd)
	case sessionStoreCopilot:
		err = RunAgentSessionLogCopilot(ctx, state, sessionID, cwd)
	default:
		err = fmt.Errorf("unknown session hook agent %q", agent)
	}
	return spoolHookFailure(state.Config.Agent, agent, "session", sessionID, err)
}

func RunAgentHookNotify(ctx context.Context, state *GlobalState, agent, event string) error {
	err := RunAgentNotify(ctx, state, agent, event)
	return spoolHookFailure(state.Config.Agent, agent, "notify:"+event, "", err)
}

func hookFailureRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion), nil
}

func boundedHookFailureDetail(err error, sessionID string, detailLimit int) string {
	detail := err.Error()
	if sessionID != "" {
		detail = strings.ReplaceAll(detail, sessionID, "<session>")
	}
	detail = strings.Join(strings.Fields(detail), " ")
	for len(detail) > detailLimit {
		_, size := utf8.DecodeLastRuneInString(detail)
		detail = detail[:len(detail)-size]
	}
	return detail
}

func spoolHookFailure(cfg AgentConfig, agent, operation, sessionID string, hookErr error) error {
	if hookErr == nil {
		return nil
	}
	root, err := hookFailureRoot()
	if err != nil {
		return errors.Join(hookErr, fmt.Errorf("failed to resolve hook failure spool: %w", err))
	}
	if info, statErr := os.Lstat(root); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.Join(hookErr, fmt.Errorf("refusing linked hook failure spool %s", root))
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return errors.Join(hookErr, fmt.Errorf("failed to inspect hook failure spool: %w", statErr))
	}
	if secureErr := secureSessionDirectories(root); secureErr != nil {
		return errors.Join(hookErr, fmt.Errorf("failed to secure hook failure spool: %w", secureErr))
	}
	record := hookFailureRecord{
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		Agent:      agent,
		Operation:  operation,
		Detail:     boundedHookFailureDetail(hookErr, sessionID, cfg.hookFailureDetailLimit()),
	}
	if sessionID != "" {
		record.SessionHash = truncatedDigest(sessionDigest(agent, sessionID))
	}
	content, err := json.Marshal(record)
	if err != nil {
		return errors.Join(hookErr, fmt.Errorf("failed to encode hook failure: %w", err))
	}
	name := strings.NewReplacer(":", "", ".", "").Replace(time.Now().UTC().Format("20060102T150405.000000000Z")) + "-" + truncatedDigest(sessionDigest(agent, operation, record.Detail)) + ".json"
	// Write beside the target and rename so a reader never sees a half-written
	// record: the doctor and the trimmer only pick up `.json` names.
	target := filepath.Join(root, name)
	if err := writeOwnerOnly(target+".tmp", append(content, '\n')); err != nil {
		return errors.Join(hookErr, fmt.Errorf("failed to write hook failure spool: %w", err))
	}
	if err := os.Rename(target+".tmp", target); err != nil {
		return errors.Join(hookErr, fmt.Errorf("failed to publish hook failure spool: %w", err))
	}
	if err := trimHookFailureSpool(root, cfg.hookFailureLimit()); err != nil {
		return errors.Join(hookErr, err)
	}
	return hookErr
}

func trimHookFailureSpool(path string, limit int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read hook failure spool: %w", err)
	}
	records := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".json") {
			records = append(records, entry)
		}
	}
	if len(records) <= limit {
		return nil
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return fmt.Errorf("failed to confine hook failure spool: %w", err)
	}
	defer func() { _ = root.Close() }()
	for _, entry := range records[:len(records)-limit] {
		if err := root.Remove(entry.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to trim hook failure spool: %w", err)
		}
	}
	return nil
}
