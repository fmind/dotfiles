package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type copilotSessionEndReason string

const (
	copilotSessionComplete copilotSessionEndReason = "complete"
	copilotSessionError    copilotSessionEndReason = "error"
	copilotSessionAbort    copilotSessionEndReason = "abort"
	copilotSessionTimeout  copilotSessionEndReason = "timeout"
	copilotSessionUserExit copilotSessionEndReason = "user_exit"
)

// copilotSessionEndInput intentionally mirrors only GitHub's documented
// camelCase sessionEnd payload; hook metadata is never transcript content.
type copilotSessionEndInput struct {
	SessionID string                  `json:"sessionId"`
	CWD       string                  `json:"cwd"`
	Reason    copilotSessionEndReason `json:"reason"`
	Timestamp int64                   `json:"timestamp"`
}

func decodeCopilotSessionEnd(reader io.Reader) (copilotSessionEndInput, error) {
	if reader == nil {
		return copilotSessionEndInput{}, errors.New("missing Copilot sessionEnd payload")
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var input copilotSessionEndInput
	if err := decoder.Decode(&input); err != nil {
		return copilotSessionEndInput{}, fmt.Errorf("invalid Copilot sessionEnd payload: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return copilotSessionEndInput{}, errors.New("invalid Copilot sessionEnd payload: expected one JSON object")
	}
	if !isValidSessionID(input.SessionID) {
		return copilotSessionEndInput{}, fmt.Errorf("invalid Copilot sessionEnd sessionId: %q", input.SessionID)
	}
	if input.CWD == "" {
		return copilotSessionEndInput{}, errors.New("invalid Copilot sessionEnd payload: missing cwd")
	}
	if input.Timestamp <= 0 {
		return copilotSessionEndInput{}, errors.New("invalid Copilot sessionEnd payload: timestamp must be Unix milliseconds")
	}
	switch input.Reason {
	case copilotSessionComplete, copilotSessionError, copilotSessionAbort, copilotSessionTimeout, copilotSessionUserExit:
	default:
		return copilotSessionEndInput{}, fmt.Errorf("unsupported Copilot sessionEnd reason %q", input.Reason)
	}
	return input, nil
}

// RunCopilotSessionEndHook treats sessionEnd as a wake-up signal only. It
// always returns success to Copilot; actionable failures are retained in the
// owner-only hook spool without exposing the raw session identity.
func RunCopilotSessionEndHook(ctx context.Context, state *GlobalState) error {
	input, decodeErr := decodeCopilotSessionEnd(state.Stdin)
	if decodeErr != nil {
		recordNonBlockingCopilotHookFailure(state, "", decodeErr)
		_, _ = fmt.Fprintln(state.Stdout, "{}")
		return nil
	}
	syncErr := RunAgentSessionLogCopilot(ctx, state, input.SessionID, input.CWD)
	if syncErr != nil {
		recordNonBlockingCopilotHookFailure(state, input.SessionID, syncErr)
	}
	_, _ = fmt.Fprintln(state.Stdout, "{}")
	return nil
}

func recordNonBlockingCopilotHookFailure(state *GlobalState, sessionID string, hookErr error) {
	cfg := state.Config.Agent
	spooledErr := spoolHookFailure(cfg, sessionStoreCopilot, "sessionEnd", sessionID, hookErr)
	if spooledErr != nil {
		_, _ = fmt.Fprintf(state.Stderr, "copilot sessionEnd sync failed without blocking the session: %s\n", boundedHookFailureDetail(spooledErr, sessionID, cfg.hookFailureDetailLimit()))
	}
}
