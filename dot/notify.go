package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

// notifyEvent renders one agent hook event as desktop-notification copy. Hook
// vocabulary ("Stop", "SessionEnd") means nothing at a glance, so every event
// carries the icon and headline that answer "why am I being pinged?".
type notifyEvent struct {
	Icon     string
	Headline string
}

// notifyEvents enumerates the hook events wired to a notification; adding an
// entry here is the only code change needed to cover a new hook.
var notifyEvents = map[string]notifyEvent{
	"stop":        {Icon: "✅", Headline: "Turn finished — waiting for you"},
	"session-end": {Icon: "🏁", Headline: "Session ended"},
}

// notifyAgents spells agent slugs the way their vendors do.
var notifyAgents = map[string]string{
	"agy":         "Antigravity",
	"antigravity": "Antigravity",
	"claude":      "Claude Code",
	"codex":       "Codex",
	"copilot":     "Copilot",
	"opencode":    "OpenCode",
}

// notifyExpireMillis keeps the alert on screen long enough to read the pane
// locator without pinning it to the tray forever.
const notifyExpireMillis = "10000"

// Notification is the resolved, backend-agnostic content of a desktop alert.
// Details stay split into lines because Linux renders a multi-line body while
// macOS collapses everything below the subtitle into a single string.
type Notification struct {
	Summary  string
	Headline string
	Details  []string
}

// NewNotifyCmd constructs the top-level notify command.
func NewNotifyCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:      "notify",
		Aliases:   []string{"n"},
		Usage:     "Send an OS-independent desktop notification",
		ArgsUsage: "<agent> <event> | <summary> [headline] [details...]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunNotify(ctx, state, cmd.Args().Slice()...)
		},
	}
}

// RunNotify sends desktop notifications for agent hooks (<agent> <event>) or custom payloads (<summary> [headline] [details...]).
func RunNotify(ctx context.Context, state *GlobalState, args ...string) error {
	if len(args) == 0 {
		return errors.New("agent name or notification summary is required")
	}

	if len(args) >= 2 {
		agent := args[0]
		event := args[1]
		if _, knownAgent := notifyAgents[agent]; knownAgent {
			return RunAgentNotify(ctx, state, agent, event)
		}
		if _, knownEvent := notifyEvents[event]; knownEvent {
			return RunAgentNotify(ctx, state, agent, event)
		}
	}

	summary := args[0]
	headline := ""
	if len(args) > 1 {
		headline = args[1]
	}
	var details []string
	if len(args) > 2 {
		details = args[2:]
	}

	return sendNotification(ctx, state, Notification{
		Summary:  summary,
		Headline: headline,
		Details:  details,
	})
}

// RunAgentNotify turns an agent hook payload into a desktop notification naming
// the project and pane to go back to.
func RunAgentNotify(ctx context.Context, state *GlobalState, agent, event string) error {
	input, err := parseStdin(state.Stdin)
	if err != nil {
		return err
	}
	cwd := ""
	if input != nil {
		cwd = resolveCWD(input.CWD)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	notification, err := buildNotification(agent, event, cwd, home, os.Getenv)
	if err != nil {
		return err
	}

	return sendNotification(ctx, state, notification)
}

// sendNotification dispatches a Notification struct to the host desktop notification service.
func sendNotification(ctx context.Context, state *GlobalState, notification Notification) error {
	if !notifySupported(runtime.GOOS, os.Getenv) {
		// Headless runs (CI, ssh without a session bus) have nowhere to draw an
		// alert; skipping keeps the hook from failing on every single turn.
		state.Logger.Debug("Skipping desktop notification: no desktop session", "os", runtime.GOOS)
		return nil
	}

	name, args, err := notifyCommand(runtime.GOOS, state.Runner.LookPath, notification)
	if err != nil {
		return err
	}
	if _, err := state.Runner.Run(ctx, "", nil, name, args...); err != nil {
		return fmt.Errorf("failed to send desktop notification with %s: %w", name, err)
	}
	return nil
}

// buildNotification assembles the alert text: what happened, in which project,
// and which pane to return to.
func buildNotification(agent, event, cwd, home string, getenv func(string) string) (Notification, error) {
	if agent == "" {
		return Notification{}, errors.New("agent name is required")
	}
	meta, known := notifyEvents[event]
	if !known {
		return Notification{}, fmt.Errorf("unknown agent notify event %q (want one of: %s)", event, strings.Join(notifyEventNames(), ", "))
	}

	label, named := notifyAgents[agent]
	if !named {
		label = agent
	}

	summary := meta.Icon + " " + label
	details := make([]string, 0, 2)
	if cwd != "" {
		summary += " · " + filepath.Base(cwd)
		details = append(details, displayPath(home, cwd))
	}
	if location := zellijLocation(getenv); location != "" {
		details = append(details, location)
	}

	return Notification{Summary: summary, Headline: meta.Headline, Details: details}, nil
}

// notifyEventNames lists supported events in a stable order for error messages.
func notifyEventNames() []string {
	names := make([]string, 0, len(notifyEvents))
	for name := range notifyEvents {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// zellijLocation names the pane to jump back to. Zellij exports both variables
// into every pane it spawns, so hooks inherit them from the agent process.
func zellijLocation(getenv func(string) string) string {
	session := getenv("ZELLIJ_SESSION_NAME")
	if session == "" {
		return ""
	}
	location := "zellij " + session
	if pane := getenv("ZELLIJ_PANE_ID"); pane != "" {
		location += " · pane " + pane
	}
	return location
}

// displayPath shortens a home-relative path to its "~" form so the notification
// body stays readable at tray width.
func displayPath(home, path string) string {
	if home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if prefix := home + string(os.PathSeparator); strings.HasPrefix(path, prefix) {
		return "~" + string(os.PathSeparator) + path[len(prefix):]
	}
	return path
}

// notifySupported reports whether this host can draw a desktop notification.
func notifySupported(goos string, getenv func(string) string) bool {
	switch goos {
	case "darwin":
		return true
	case "linux":
		// Crostini routes org.freedesktop.Notifications to the ChromeOS tray over
		// the session bus, so D-Bus — not X11/Wayland — is what actually gates this.
		return getenv("DBUS_SESSION_BUS_ADDRESS") != ""
	default:
		return false
	}
}

// notifyCommand resolves the notifier binary and its arguments for this host.
func notifyCommand(goos string, lookPath func(string) (string, error), n Notification) (string, []string, error) {
	switch goos {
	case "darwin":
		return "osascript", osascriptArgs(n), nil
	case "linux":
		if _, err := lookPath("notify-send"); err == nil {
			return "notify-send", notifySendArgs(n), nil
		}
		// Crostini exposes the notification service but ships no libnotify, and
		// gdbus comes with glib, so it is the reliable fallback there.
		if _, err := lookPath("gdbus"); err == nil {
			return "gdbus", gdbusArgs(n), nil
		}
		return "", nil, fmt.Errorf("%w: install libnotify-bin (notify-send) or glib (gdbus)", ErrToolNotInstalled)
	default:
		return "", nil, fmt.Errorf("desktop notifications are unsupported on %s", goos)
	}
}

// linuxBody renders the multi-line body understood by libnotify and D-Bus.
func linuxBody(n Notification) string {
	return strings.Join(append([]string{n.Headline}, n.Details...), "\n")
}

// notifySendArgs builds a libnotify invocation.
func notifySendArgs(n Notification) []string {
	return []string{"--app-name=dot", "--expire-time=" + notifyExpireMillis, n.Summary, linuxBody(n)}
}

// gdbusArgs builds a raw org.freedesktop.Notifications.Notify call. gdbus does
// not infer D-Bus numeric types, so replaces_id, actions, hints and the timeout
// must carry explicit GVariant annotations or the call fails with InvalidArgs.
func gdbusArgs(n Notification) []string {
	return []string{
		"call", "--session",
		"--dest", "org.freedesktop.Notifications",
		"--object-path", "/org/freedesktop/Notifications",
		"--method", "org.freedesktop.Notifications.Notify",
		"dot", "uint32 0", "dialog-information",
		n.Summary, linuxBody(n),
		"@as []", "@a{sv} {}", "int32 " + notifyExpireMillis,
	}
}

// osascriptArgs builds the macOS notification script.
func osascriptArgs(n Notification) []string {
	// macOS renders one collapsed body line and offers a dedicated subtitle, so
	// details join with a separator instead of newlines.
	message := strings.Join(n.Details, " · ")
	if message == "" {
		// `display notification` requires a non-empty body; promote the headline.
		return []string{"-e", "display notification " + quoteAppleScript(n.Headline) + " with title " + quoteAppleScript(n.Summary)}
	}
	script := "display notification " + quoteAppleScript(message) +
		" with title " + quoteAppleScript(n.Summary) +
		" subtitle " + quoteAppleScript(n.Headline)
	return []string{"-e", script}
}

// quoteAppleScript wraps s in AppleScript string quotes. osascript receives the
// whole script as one argument, so quotes and backslashes inside project paths
// must be escaped or they terminate the literal early.
func quoteAppleScript(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
