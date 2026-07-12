//go:build linux || darwin

package dot

import (
	"context"
	"io"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func TestStandardRunnerIsolatesCapturedCommandProcessGroup(t *testing.T) {
	parentGroup, err := syscall.Getpgid(0)
	if err != nil {
		t.Fatal(err)
	}
	runner := NewStandardRunner(nil, io.Discard, io.Discard)
	output, err := runner.Run(context.Background(), "", nil, "sh", "-c", "ps -o pgid= -p $$")
	if err != nil {
		t.Fatalf("read child process group: %v", err)
	}
	childGroup, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		t.Fatalf("parse child process group %q: %v", output, err)
	}
	if childGroup == parentGroup {
		t.Fatalf("captured command shares parent process group %d", parentGroup)
	}
}
