package dot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Runner defines the interface for executing external command-line tools.
// Implementations of this interface allow commands to be run either headlessly
// (capturing stdout/stderr) or interactively (binding to os.Stdin/Stdout/Stderr).
type Runner interface {
	Run(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error)
	RunInteractive(ctx context.Context, dir, name string, args ...string) error
	LookPath(name string) (string, error)
}

// StandardRunner implements Runner using the standard os/exec package.
type StandardRunner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewStandardRunner creates a new instance of StandardRunner with specified standard streams.
func NewStandardRunner(stdin io.Reader, stdout, stderr io.Writer) *StandardRunner {
	return &StandardRunner{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
}

// Run executes a command with optional stdin and returns the raw stdout. If the command fails,
// it returns an error containing the command, exit status, and the stderr output.
func (r *StandardRunner) Run(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
	// Trusted call site and single command choke point: name is a constant tool name
	// and args are built by dot, never shell-interpolated from untrusted input.
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: args are dot-built, not user shell input
	if dir != "" {
		cmd.Dir = dir
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// On interrupt/timeout the child is SIGKILLed and Run reports "signal: killed";
		// surface the context cause instead so callers (and main's exit-130 path) see a
		// clean context.Canceled/DeadlineExceeded rather than noisy kill output.
		if ce := ctx.Err(); ce != nil {
			return "", ce
		}
		return "", fmt.Errorf("command %s %v failed: %w\nstderr: %s", name, args, err, stderr.String())
	}

	return stdout.String(), nil
}

// RunInteractive runs a command with stdin, stdout, and stderr bound to the runner's streams.
func (r *StandardRunner) RunInteractive(ctx context.Context, dir, name string, args ...string) error {
	// Trusted call site and single command choke point: name is a constant tool name
	// and args are built by dot, never shell-interpolated from untrusted input.
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: args are dot-built, not user shell input
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr

	if err := cmd.Run(); err != nil {
		// Prefer the context cause on interrupt/timeout so cancellation propagates as
		// context.Canceled (feeding main's clean exit 130) rather than "signal: killed".
		if ce := ctx.Err(); ce != nil {
			return ce
		}
		return err
	}
	return nil
}

// LookPath checks if a command is available in the system PATH.
func (r *StandardRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
