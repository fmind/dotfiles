package dot

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestOSBrowserWithoutDisplay(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	browser := OSBrowser{}
	if browser.HasSupport() {
		t.Skip("headless browser assertion is Linux-specific")
	}
	if err := browser.Open("https://example.com"); err == nil || !strings.Contains(err.Error(), "no browser support") {
		t.Fatalf("expected no-browser-support error, got %v", err)
	}
}

func TestOSBrowserStartsPlatformOpener(t *testing.T) {
	t.Setenv("DISPLAY", ":test")
	// Stub whichever opener this platform actually shells out to; PATH is
	// replaced wholesale below, so a hardcoded xdg-open fails on macOS.
	openerName, _, err := platformOpener(runtime.GOOS, "https://example.com")
	if err != nil {
		t.Skipf("no platform opener: %v", err)
	}
	binDir := t.TempDir()
	opener := filepath.Join(binDir, openerName)
	if err := os.WriteFile(opener, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := (OSBrowser{}).Open("https://example.com"); err != nil {
		t.Fatalf("Open: %v", err)
	}
	// Open deliberately starts the browser asynchronously; give the tiny test
	// process time to exit so it cannot leak into the rest of the suite.
	time.Sleep(10 * time.Millisecond)
}

func TestPlatformOpener(t *testing.T) {
	tests := []struct {
		goos    string
		wantCmd string
		wantArg []string
		wantErr bool
	}{
		{goos: "linux", wantCmd: "xdg-open", wantArg: []string{"https://example.com"}},
		{goos: "darwin", wantCmd: "open", wantArg: []string{"https://example.com"}},
		{goos: "windows", wantCmd: "cmd", wantArg: []string{"/c", "start", "https://example.com"}},
		{goos: "plan9", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.goos, func(t *testing.T) {
			cmd, args, err := platformOpener(tc.goos, "https://example.com")
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), tc.goos) {
					t.Fatalf("expected unsupported-platform error naming %s, got %v", tc.goos, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("platformOpener: %v", err)
			}
			if cmd != tc.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tc.wantCmd)
			}
			if strings.Join(args, " ") != strings.Join(tc.wantArg, " ") {
				t.Errorf("args = %v, want %v", args, tc.wantArg)
			}
		})
	}
}

func TestBrowserSupported(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	if browserSupported("linux") {
		t.Error("linux without a display server must not report browser support")
	}
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	if !browserSupported("linux") {
		t.Error("linux with WAYLAND_DISPLAY must report browser support")
	}
	for _, goos := range []string{"darwin", "windows"} {
		if !browserSupported(goos) {
			t.Errorf("%s must report browser support", goos)
		}
	}
	if browserSupported("plan9") {
		t.Error("unsupported platform must not report browser support")
	}
}

// failingWriter reports an error on every write so error paths of writer
// decorators (urlOpener, ansiStripper) stay reachable.
type failingWriter struct{ err error }

func (f failingWriter) Write([]byte) (int, error) { return 0, f.err }

func TestURLOpenerPropagatesWriteError(t *testing.T) {
	wantErr := errors.New("boom")
	opener := &urlOpener{browser: &FakeBrowser{OpenFunc: func(string) error {
		t.Error("browser must not be opened when the underlying write fails")
		return nil
	}}}
	w := &urlOpenerWriter{w: failingWriter{err: wantErr}, opener: opener}

	if _, err := w.Write([]byte("https://example.com\n")); !errors.Is(err, wantErr) {
		t.Fatalf("expected the underlying write error, got %v", err)
	}
}

func TestURLOpenerOpensFirstURLOnly(t *testing.T) {
	var opened []string
	opener := &urlOpener{browser: &FakeBrowser{OpenFunc: func(url string) error {
		opened = append(opened, url)
		return nil
	}}}
	w := &urlOpenerWriter{w: io.Discard, opener: opener}

	for _, chunk := range []string{"visit http://first.example ok\n", "then https://second.example ok\n"} {
		if _, err := w.Write([]byte(chunk)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	if len(opened) != 1 || opened[0] != "http://first.example" {
		t.Fatalf("expected only the first URL to be opened, got %v", opened)
	}
}
