package dot

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// requireLinux guards the tests that exercise the runtime.GOOS-dependent
// dispatch path; the per-platform argument builders are covered directly.
func requireLinux(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skipf("dispatch path is linux-specific, running on %s", runtime.GOOS)
	}
}

// envMap returns a getenv function backed by a map, so tests never mutate the
// real process environment.
func envMap(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestBuildNotificationIncludesProjectAndPane(t *testing.T) {
	home := filepath.Join("/home", "fmind")
	cwd := filepath.Join(home, "internals", "fgentic")

	notification, err := buildNotification("claude", "stop", cwd, home, envMap(map[string]string{
		"ZELLIJ_SESSION_NAME": "main",
		"ZELLIJ_PANE_ID":      "3",
	}))
	if err != nil {
		t.Fatalf("buildNotification returned error: %v", err)
	}

	if want := "✅ Claude Code · fgentic"; notification.Summary != want {
		t.Errorf("summary = %q, want %q", notification.Summary, want)
	}
	if !strings.Contains(notification.Headline, "Turn finished") {
		t.Errorf("headline = %q, want it to mention the finished turn", notification.Headline)
	}
	wantDetails := []string{
		"~" + string(os.PathSeparator) + filepath.Join("internals", "fgentic"),
		"zellij main · pane 3",
	}
	if len(notification.Details) != len(wantDetails) {
		t.Fatalf("details = %v, want %v", notification.Details, wantDetails)
	}
	for i, want := range wantDetails {
		if notification.Details[i] != want {
			t.Errorf("details[%d] = %q, want %q", i, notification.Details[i], want)
		}
	}
}

func TestBuildNotificationWithoutZellijOrCWD(t *testing.T) {
	notification, err := buildNotification("codex", "session-end", "", "/home/fmind", envMap(nil))
	if err != nil {
		t.Fatalf("buildNotification returned error: %v", err)
	}
	if want := "🏁 Codex"; notification.Summary != want {
		t.Errorf("summary = %q, want %q", notification.Summary, want)
	}
	if len(notification.Details) != 0 {
		t.Errorf("details = %v, want none", notification.Details)
	}
}

func TestBuildNotificationNeedsInput(t *testing.T) {
	notification, err := buildNotification("claude", "needs-input", "/home/fmind/externals/demo", "/home/fmind", envMap(nil))
	if err != nil {
		t.Fatalf("buildNotification returned error: %v", err)
	}
	if want := "⏳ Claude Code · demo"; notification.Summary != want {
		t.Errorf("summary = %q, want %q", notification.Summary, want)
	}
	if want := "Waiting for your input"; notification.Headline != want {
		t.Errorf("headline = %q, want %q", notification.Headline, want)
	}
}

func TestBuildNotificationUnknownAgentFallsBackToSlug(t *testing.T) {
	notification, err := buildNotification("aider", "stop", "", "/home/fmind", envMap(nil))
	if err != nil {
		t.Fatalf("buildNotification returned error: %v", err)
	}
	if !strings.HasSuffix(notification.Summary, "aider") {
		t.Errorf("summary = %q, want it to end with the raw agent slug", notification.Summary)
	}
}

func TestBuildNotificationRejectsBadInput(t *testing.T) {
	if _, err := buildNotification("", "stop", "", "/home/fmind", envMap(nil)); err == nil {
		t.Error("expected an error for an empty agent name")
	}
	_, err := buildNotification("claude", "compacted", "", "/home/fmind", envMap(nil))
	if err == nil {
		t.Fatal("expected an error for an unknown event")
	}
	if !strings.Contains(err.Error(), "session-end") || !strings.Contains(err.Error(), "stop") {
		t.Errorf("error = %q, want it to list the supported events", err)
	}
}

func TestZellijLocationOmitsPaneWhenUnset(t *testing.T) {
	got := zellijLocation(envMap(map[string]string{"ZELLIJ_SESSION_NAME": "work"}))
	if want := "zellij work"; got != want {
		t.Errorf("zellijLocation = %q, want %q", got, want)
	}
}

func TestDisplayPath(t *testing.T) {
	home := filepath.Join("/home", "fmind")
	cases := map[string]struct{ path, want string }{
		"home itself":  {home, "~"},
		"under home":   {filepath.Join(home, "externals"), "~" + string(os.PathSeparator) + "externals"},
		"outside home": {filepath.Join("/opt", "tools"), filepath.Join("/opt", "tools")},
		"sibling dir":  {home + "-backup", home + "-backup"},
	}
	for name, tc := range cases {
		if got := displayPath(home, tc.path); got != tc.want {
			t.Errorf("%s: displayPath(%q) = %q, want %q", name, tc.path, got, tc.want)
		}
	}
	if got := displayPath("", "/anywhere"); got != "/anywhere" {
		t.Errorf("displayPath with empty home = %q, want %q", got, "/anywhere")
	}
}

func TestNotifySupported(t *testing.T) {
	if !notifySupported("darwin", envMap(nil)) {
		t.Error("darwin should always support notifications")
	}
	if notifySupported("linux", envMap(nil)) {
		t.Error("linux without a session bus should be unsupported")
	}
	if !notifySupported("linux", envMap(map[string]string{"DBUS_SESSION_BUS_ADDRESS": "unix:path=/run/user/1000/bus"})) {
		t.Error("linux with a session bus should be supported")
	}
	if notifySupported("windows", envMap(nil)) {
		t.Error("windows should be unsupported")
	}
}

// notFound stands in for a binary missing from PATH.
func notFound(string) (string, error) { return "", errors.New("not found") }

func TestNotifyCommandLinuxPrefersNotifySend(t *testing.T) {
	notification := Notification{Summary: "✅ Codex · dot", Headline: "Turn finished", Details: []string{"~/dot"}}

	name, args, err := notifyCommand("linux", func(string) (string, error) { return "/usr/bin/notify-send", nil }, notification)
	if err != nil {
		t.Fatalf("notifyCommand returned error: %v", err)
	}
	if name != "notify-send" {
		t.Errorf("name = %q, want notify-send", name)
	}
	if args[len(args)-2] != notification.Summary {
		t.Errorf("summary argument = %q, want %q", args[len(args)-2], notification.Summary)
	}
	if body := args[len(args)-1]; !strings.Contains(body, "Turn finished\n~/dot") {
		t.Errorf("body = %q, want headline and details on separate lines", body)
	}
}

func TestNotifyCommandLinuxFallsBackToGdbus(t *testing.T) {
	lookPath := func(name string) (string, error) {
		if name == "gdbus" {
			return "/usr/bin/gdbus", nil
		}
		return notFound(name)
	}

	name, args, err := notifyCommand("linux", lookPath, Notification{Summary: "s", Headline: "h"})
	if err != nil {
		t.Fatalf("notifyCommand returned error: %v", err)
	}
	if name != "gdbus" {
		t.Errorf("name = %q, want gdbus", name)
	}
	// gdbus rejects untyped numerics, so the annotations are load-bearing.
	joined := strings.Join(args, " ")
	for _, want := range []string{"uint32 0", "@as []", "@a{sv} {}", "int32 " + notifyExpireMillis} {
		if !strings.Contains(joined, want) {
			t.Errorf("args %v missing GVariant annotation %q", args, want)
		}
	}
}

func TestNotifyCommandErrors(t *testing.T) {
	if _, _, err := notifyCommand("linux", notFound, Notification{}); !errors.Is(err, ErrToolNotInstalled) {
		t.Errorf("error = %v, want ErrToolNotInstalled", err)
	}
	if _, _, err := notifyCommand("plan9", notFound, Notification{}); err == nil {
		t.Error("expected an error on an unsupported platform")
	}
}

func TestOsascriptArgs(t *testing.T) {
	name, args, err := notifyCommand("darwin", notFound, Notification{
		Summary:  "✅ Claude Code · dot",
		Headline: "Turn finished",
		Details:  []string{"~/dot", "zellij main · pane 3"},
	})
	if err != nil {
		t.Fatalf("notifyCommand returned error: %v", err)
	}
	if name != "osascript" || args[0] != "-e" {
		t.Fatalf("command = %q %v, want osascript -e ...", name, args)
	}
	script := args[1]
	for _, want := range []string{`"~/dot · zellij main · pane 3"`, `with title "✅ Claude Code · dot"`, `subtitle "Turn finished"`} {
		if !strings.Contains(script, want) {
			t.Errorf("script = %q, want it to contain %q", script, want)
		}
	}
}

func TestOsascriptArgsPromotesHeadlineWithoutDetails(t *testing.T) {
	args := osascriptArgs(Notification{Summary: "🏁 Codex", Headline: "Session ended"})
	if !strings.Contains(args[1], `display notification "Session ended"`) {
		t.Errorf("script = %q, want the headline promoted to the body", args[1])
	}
	if strings.Contains(args[1], "subtitle") {
		t.Errorf("script = %q, want no subtitle when there are no details", args[1])
	}
}

func TestQuoteAppleScriptEscapes(t *testing.T) {
	got := quoteAppleScript(`a"b\c`)
	if want := `"a\"b\\c"`; got != want {
		t.Errorf("quoteAppleScript = %q, want %q", got, want)
	}
}

func TestRunAgentNotifySendsNotification(t *testing.T) {
	requireLinux(t)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/run/user/1000/bus")
	t.Setenv("ZELLIJ_SESSION_NAME", "main")
	t.Setenv("ZELLIJ_PANE_ID", "7")

	var gotName string
	var gotArgs []string
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			if name == "notify-send" {
				return "/usr/bin/notify-send", nil
			}
			return notFound(name)
		},
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			gotName, gotArgs = name, args
			return "", nil
		},
	}

	state := newTestState(runner)
	state.Stdin = strings.NewReader(`{"session_id":"abc","cwd":"/home/fmind/externals/dot"}`)

	if err := RunAgentNotify(context.Background(), state, "claude", "stop"); err != nil {
		t.Fatalf("RunAgentNotify returned error: %v", err)
	}
	if gotName != "notify-send" {
		t.Fatalf("ran %q, want notify-send", gotName)
	}
	joined := strings.Join(gotArgs, " ")
	for _, want := range []string{"dot", "zellij main · pane 7"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args %v missing %q", gotArgs, want)
		}
	}
}

func TestRunAgentNotifyRejectsUnknownEvent(t *testing.T) {
	state := newTestState(&FakeRunner{})
	if err := RunAgentNotify(context.Background(), state, "claude", "exploded"); err == nil {
		t.Error("expected an error for an unknown event")
	}
}

func TestRunAgentNotifySkipsWithoutDesktopSession(t *testing.T) {
	requireLinux(t)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "")

	ran := false
	runner := &FakeRunner{RunFunc: func(_ context.Context, _ string, _ io.Reader, _ string, _ ...string) (string, error) {
		ran = true
		return "", nil
	}}

	state := newTestState(runner)
	if err := RunAgentNotify(context.Background(), state, "claude", "stop"); err != nil {
		t.Fatalf("RunAgentNotify returned error: %v", err)
	}
	if ran {
		t.Error("no notifier should run without a desktop session")
	}
}

func TestRunNotifyCustomAndAgent(t *testing.T) {
	requireLinux(t)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/run/user/1000/bus")

	t.Run("empty args returns error", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		if err := RunNotify(context.Background(), state); err == nil {
			t.Error("expected error for empty args")
		}
	})

	t.Run("custom summary and headline", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "notify-send" {
					return "/usr/bin/notify-send", nil
				}
				return notFound(name)
			},
			RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
				gotName, gotArgs = name, args
				return "", nil
			},
		}

		state := newTestState(runner)
		if err := RunNotify(context.Background(), state, "Custom Title", "Sub Headline", "Detail 1"); err != nil {
			t.Fatalf("RunNotify returned error: %v", err)
		}
		if gotName != "notify-send" {
			t.Fatalf("ran %q, want notify-send", gotName)
		}
		joined := strings.Join(gotArgs, " ")
		if !strings.Contains(joined, "Custom Title") || !strings.Contains(joined, "Sub Headline") {
			t.Errorf("args %v missing custom notification payload", gotArgs)
		}
	})

	t.Run("agent event dispatch", func(t *testing.T) {
		var gotArgs []string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "notify-send" {
					return "/usr/bin/notify-send", nil
				}
				return notFound(name)
			},
			RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
				gotArgs = args
				return "", nil
			},
		}

		state := newTestState(runner)
		if err := RunNotify(context.Background(), state, "claude", "stop"); err != nil {
			t.Fatalf("RunNotify returned error: %v", err)
		}
		joined := strings.Join(gotArgs, " ")
		if !strings.Contains(joined, "Claude Code") {
			t.Errorf("args %v missing Claude Code agent headline", gotArgs)
		}
	})
}

func TestRunAgentNotifySurfacesFailures(t *testing.T) {
	t.Run("a malformed hook payload is rejected", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		state.Stdin = strings.NewReader("{not json")

		err := RunAgentNotify(context.Background(), state, "claude", "stop")
		if err == nil || !strings.Contains(err.Error(), "failed to parse agent hook input") {
			t.Fatalf("expected a hook parse error, got %v", err)
		}
	})

	t.Run("an unresolvable home is reported", func(t *testing.T) {
		t.Setenv("HOME", "")
		state := newTestState(&FakeRunner{})

		err := RunAgentNotify(context.Background(), state, "claude", "stop")
		if err == nil || !strings.Contains(err.Error(), "failed to resolve home directory") {
			t.Fatalf("expected a home directory error, got %v", err)
		}
	})

	t.Run("a notifier failure is reported", func(t *testing.T) {
		requireLinux(t)
		t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/run/user/1000/bus")

		state := newTestState(&FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "notify-send" {
					return "/usr/bin/notify-send", nil
				}
				return notFound(name)
			},
			RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
				return "", errors.New("dbus unavailable")
			},
		})

		err := RunAgentNotify(context.Background(), state, "claude", "stop")
		if err == nil || !strings.Contains(err.Error(), "failed to send desktop notification with notify-send") {
			t.Fatalf("expected a notifier failure, got %v", err)
		}
	})
}

func TestNotifyCommandDispatchesAgentHooks(t *testing.T) {
	requireLinux(t)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "")

	state := newTestState(&FakeRunner{})
	app := &cli.Command{Commands: []*cli.Command{NewNotifyCmd(state)}}

	// "stop" is a known event, so the pair routes to the agent hook path even
	// though "some-agent" is not a known agent.
	if err := app.Run(context.Background(), []string{"dot", "notify", "some-agent", "stop"}); err != nil {
		t.Fatalf("notify command: %v", err)
	}
}

func TestRunAgentNotifySkipsNonIdleOrStopHookActive(t *testing.T) {
	requireLinux(t)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/run/user/1000/bus")

	ran := false
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) { return "/usr/bin/notify-send", nil },
		RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
			ran = true
			return "", nil
		},
	}

	t.Run("skips when agy is not fully idle", func(t *testing.T) {
		ran = false
		state := newTestState(runner)
		state.Stdin = strings.NewReader(`{"conversationId":"abc","fullyIdle":false}`)
		if err := RunAgentNotify(context.Background(), state, "agy", "stop"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ran {
			t.Error("notification should be skipped when agy is not fully idle")
		}
	})

	t.Run("skips when stop hook active", func(t *testing.T) {
		ran = false
		state := newTestState(runner)
		state.Stdin = strings.NewReader(`{"session_id":"abc","stop_hook_active":true}`)
		if err := RunAgentNotify(context.Background(), state, "claude", "stop"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ran {
			t.Error("notification should be skipped when stop hook is active")
		}
	})
}
