package dot

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"golang.org/x/term"
)

const (
	statusPass = "pass"
	statusFail = "fail"
	statusWarn = "warn"
	statusSkip = "skip"
)

// ANSI SGR codes. Every styled string is written through colorWriter, which
// strips these sequences when the destination is not an interactive terminal
// (pipes, files, NO_COLOR), so call sites may style unconditionally and never
// have to reason about color detection themselves.
const (
	codeBold   = "1"
	codeDim    = "2"
	codeRed    = "31"
	codeGreen  = "32"
	codeYellow = "33"
)

// style wraps s in the given ANSI SGR code followed by a reset.
func style(code, s string) string { return "\033[" + code + "m" + s + "\033[0m" }

func bold(s string) string   { return style(codeBold, s) }
func dim(s string) string    { return style(codeDim, s) }
func red(s string) string    { return style(codeRed, s) }
func green(s string) string  { return style(codeGreen, s) }
func yellow(s string) string { return style(codeYellow, s) }

// Status icons, styled once here so every command renders them identically.
var (
	passIcon = green("✓")
	failIcon = red("✗")
	warnIcon = yellow("!")
	skipIcon = dim("-")
)

// section prints a bold "=> Title" header, the shared heading style across commands.
func section(w io.Writer, title string) {
	_, _ = fmt.Fprintln(w, bold("=> "+title))
}

// ansiPattern matches ANSI SGR (color/style) escape sequences.
var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// colorEnabled reports whether ANSI styling should be written to w. It honors the
// NO_COLOR convention and only enables color when writing to an interactive terminal.
func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// ansiStripper wraps a writer and removes ANSI style sequences as they pass through.
type ansiStripper struct{ w io.Writer }

func (s ansiStripper) Write(p []byte) (int, error) {
	if _, err := s.w.Write(ansiPattern.ReplaceAll(p, nil)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// colorWriter returns w unchanged when color is enabled for it, otherwise a writer
// that strips ANSI sequences so pipes, files, and NO_COLOR consumers stay clean.
func colorWriter(w io.Writer) io.Writer {
	if colorEnabled(w) {
		return w
	}
	return ansiStripper{w}
}
