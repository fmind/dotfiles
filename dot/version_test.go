package dot

import (
	"bytes"
	"context"
	"runtime/debug"
	"slices"
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

	// Verify that "i" is in the aliases
	hasAlias := slices.Contains(cmd.Aliases, "i")
	if !hasAlias {
		t.Errorf("expected version command to have 'i' alias, got: %v", cmd.Aliases)
	}
}

func TestFormatVersion(t *testing.T) {
	setting := func(k, v string) debug.BuildSetting { return debug.BuildSetting{Key: k, Value: v} }

	tests := []struct {
		name     string
		want     string
		settings []debug.BuildSetting
	}{
		{
			name:     "no vcs metadata falls back to the bare version",
			settings: []debug.BuildSetting{setting("-race", "true")},
			want:     "dot 9.9.9",
		},
		{
			name:     "long revisions are truncated to 12 characters",
			settings: []debug.BuildSetting{setting("vcs.revision", "0123456789abcdef0123"), setting("vcs.modified", "false")},
			want:     "dot 9.9.9 (0123456789ab)",
		},
		{
			name:     "short revisions are kept intact",
			settings: []debug.BuildSetting{setting("vcs.revision", "abc123")},
			want:     "dot 9.9.9 (abc123)",
		},
		{
			name:     "a modified worktree is flagged as dirty",
			settings: []debug.BuildSetting{setting("vcs.revision", "abc123"), setting("vcs.modified", "true")},
			want:     "dot 9.9.9 (abc123, dirty)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatVersion("9.9.9", tc.settings); got != tc.want {
				t.Errorf("formatVersion = %q, want %q", got, tc.want)
			}
		})
	}
}
