package dot

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// FakeRunner implements Runner for unit testing.
type FakeRunner struct {
	LookPathFunc       func(name string) (string, error)
	RunFunc            func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error)
	RunInteractiveFunc func(ctx context.Context, dir, name string, args ...string) error
}

// LookPath mocks finding path.
func (f *FakeRunner) LookPath(name string) (string, error) {
	if f.LookPathFunc != nil {
		return f.LookPathFunc(name)
	}
	return "/usr/bin/" + name, nil
}

// Run mocks executing a command.
func (f *FakeRunner) Run(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
	if f.RunFunc != nil {
		return f.RunFunc(ctx, dir, stdin, name, args...)
	}
	return "", nil
}

// RunInteractive mocks running command interactively.
func (f *FakeRunner) RunInteractive(ctx context.Context, dir, name string, args ...string) error {
	if f.RunInteractiveFunc != nil {
		return f.RunInteractiveFunc(ctx, dir, name, args...)
	}
	return nil
}

// FakeBrowser is a mock implementation of the Browser interface.
type FakeBrowser struct {
	OpenFunc       func(url string) error
	HasSupportFunc func() bool
}

// Open implements the Browser interface.
func (f *FakeBrowser) Open(url string) error {
	if f.OpenFunc != nil {
		return f.OpenFunc(url)
	}
	return nil
}

// HasSupport implements the Browser interface.
func (f *FakeBrowser) HasSupport() bool {
	if f.HasSupportFunc != nil {
		return f.HasSupportFunc()
	}
	return true
}

// newTestState returns a GlobalState preconfigured for testing.
func newTestState(runner Runner) *GlobalState {
	return &GlobalState{
		Config:  DefaultConfig(),
		Logger:  slog.New(slog.DiscardHandler),
		Runner:  runner,
		Browser: &FakeBrowser{},
		Stdin:   strings.NewReader(""),
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}
}
