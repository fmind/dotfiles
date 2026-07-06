package dot

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestAnsiStripper(t *testing.T) {
	var buf bytes.Buffer
	s := ansiStripper{&buf}
	input := "\033[32m✓ ok\033[0m"

	n, err := s.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	// The Write contract must report the ORIGINAL length, not the stripped length,
	// or every fmt.Fprint through the stripper would surface io.ErrShortWrite.
	if n != len(input) {
		t.Errorf("expected n=%d (original length), got %d", len(input), n)
	}
	if got := buf.String(); got != "✓ ok" {
		t.Errorf("expected stripped %q, got %q", "✓ ok", got)
	}
}

func TestColorEnabled_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if colorEnabled(os.Stdout) {
		t.Error("expected colorEnabled to be false when NO_COLOR is set")
	}
}

func TestColorWriter_NonTerminalStripsAnsi(t *testing.T) {
	var buf bytes.Buffer
	// A *bytes.Buffer is not a terminal, so colorWriter must strip escapes.
	w := colorWriter(&buf)
	if _, err := w.Write([]byte(green("hello"))); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "\033[") {
		t.Errorf("expected no escape codes in non-terminal output, got %q", out)
	}
	if out != "hello" {
		t.Errorf("expected %q, got %q", "hello", out)
	}
}

func TestStyleHelpers(t *testing.T) {
	// Each helper wraps its argument in an SGR code and a reset, keeping the visible
	// text contiguous so downstream substring assertions stay valid.
	for _, styled := range []string{green("a"), red("b"), yellow("c"), dim("d"), bold("e")} {
		if !strings.HasPrefix(styled, "\033[") {
			t.Errorf("expected an SGR prefix on %q", styled)
		}
		if !strings.HasSuffix(styled, "\033[0m") {
			t.Errorf("expected a reset suffix on %q", styled)
		}
	}
	if got := green("x"); got != "\033[32mx\033[0m" {
		t.Errorf("green(\"x\") = %q", got)
	}
}

func TestSection(t *testing.T) {
	var buf bytes.Buffer
	section(&buf, "Title")
	if out := buf.String(); !strings.Contains(out, "=> Title") {
		t.Errorf("expected section header to contain '=> Title', got %q", out)
	}
}
