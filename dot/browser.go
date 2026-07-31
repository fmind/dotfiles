package dot

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// Browser defines the interface for opening web pages.
type Browser interface {
	Open(url string) error
	HasSupport() bool
}

// OSBrowser is the default Browser implementation that opens URLs using the OS default browser.
type OSBrowser struct{}

// platformOpener maps goos to the command that hands a URL to the default
// browser. Split out of Open so tests can stub the right binary name instead of
// hardcoding the Linux one and failing everywhere else, and so every platform
// branch stays reachable from a single host.
func platformOpener(goos, url string) (cmd string, args []string, err error) {
	switch goos {
	case "linux":
		return "xdg-open", []string{url}, nil
	case "windows":
		return "cmd", []string{"/c", "start", url}, nil
	case "darwin": // macOS
		return "open", []string{url}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", goos)
	}
}

// Open opens the specified URL using the OS default browser.
func (b OSBrowser) Open(url string) error {
	if !b.HasSupport() {
		return errors.New("no browser support")
	}

	cmd, args, err := platformOpener(runtime.GOOS, url)
	if err != nil {
		return err
	}

	c := exec.Command(cmd, args...) //nolint:gosec // G204: command and args are constructed safely inside Open
	return c.Start()
}

// HasSupport checks if the current environment supports opening a browser.
func (b OSBrowser) HasSupport() bool { return browserSupported(runtime.GOOS) }

// browserSupported reports whether goos can hand a URL to a browser. On Linux a
// display server must be reachable, otherwise the opener would fail at exec time.
// Parameterized on goos for the same reason as platformOpener.
func browserSupported(goos string) bool {
	if goos == "linux" {
		return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	}
	return goos == "darwin" || goos == "windows"
}

// urlOpener intercepts writes to search for URLs and opens them with the provided Browser.
type urlOpener struct {
	browser Browser
	buf     strings.Builder
	mu      sync.Mutex
	opened  bool
}

func (u *urlOpener) intercept(w io.Writer, p []byte) (int, error) {
	n, err := w.Write(p)
	if err != nil {
		return n, err
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.opened {
		return n, nil
	}

	u.buf.Write(p)
	content := u.buf.String()

	for _, prefix := range []string{"https://", "http://"} {
		if idx := strings.Index(content, prefix); idx != -1 {
			urlPart := content[idx:]
			if endIdx := strings.IndexAny(urlPart, " \t\r\n\"'"); endIdx != -1 {
				url := urlPart[:endIdx]
				u.opened = true
				_ = u.browser.Open(url)
				break
			}
		}
	}

	return n, nil
}

// urlOpenerWriter wraps an io.Writer and intercepts output to open URLs automatically.
type urlOpenerWriter struct {
	w      io.Writer
	opener *urlOpener
}

// Write intercepts the output to open URLs.
func (u *urlOpenerWriter) Write(p []byte) (n int, err error) {
	return u.opener.intercept(u.w, p)
}
