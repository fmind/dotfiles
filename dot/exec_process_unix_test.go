//go:build linux || darwin

package dot

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestIsolateProcessGroupBeforeStart(t *testing.T) {
	cmd := exec.Command("true")
	isolateProcessGroup(cmd)

	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.Setpgid {
		t.Fatal("expected an isolated process group")
	}
	if err := cmd.Cancel(); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("Cancel before start = %v, want os.ErrProcessDone", err)
	}

	completed := exec.CommandContext(context.Background(), "true")
	isolateProcessGroup(completed)
	if err := completed.Run(); err != nil {
		t.Fatal(err)
	}
	if err := completed.Cancel(); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("Cancel after exit = %v, want os.ErrProcessDone", err)
	}
}
