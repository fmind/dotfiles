package dot

import (
	"os"
	"path/filepath"
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
	openerName, _, err := platformOpener("https://example.com")
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
