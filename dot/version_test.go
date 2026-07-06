package dot

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestVersionString(t *testing.T) {
	got := VersionString()
	if !strings.HasPrefix(got, "dot "+Version) {
		t.Errorf("expected version string to start with %q, got %q", "dot "+Version, got)
	}
}

func TestVersionCommand(t *testing.T) {
	state := newTestState(&FakeRunner{})
	var buf bytes.Buffer
	state.Stdout = &buf

	cmd := NewVersionCmd(state)
	if err := cmd.Action(context.Background(), cmd); err != nil {
		t.Fatalf("version action: %v", err)
	}
	if !strings.Contains(buf.String(), "dot "+Version) {
		t.Errorf("expected version output to contain %q, got %q", "dot "+Version, buf.String())
	}

	// Verify that "n" is in the aliases
	hasAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "n" {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		t.Errorf("expected version command to have 'n' alias, got: %v", cmd.Aliases)
	}
}
