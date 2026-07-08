package dot

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestReleaseCommandAlias(t *testing.T) {
	state := newTestState(&FakeRunner{})
	cmd := NewReleaseCmd(state)

	hasAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "r" {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		t.Errorf("expected release command to have 'rl' alias, got: %v", cmd.Aliases)
	}
}

func TestRunReleaseNoBump(t *testing.T) {
	fake := &FakeRunner{
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			cmdStr := name + " " + strings.Join(args, " ")
			switch {
			case strings.Contains(cmdStr, "git rev-parse --is-inside-work-tree"):
				return "true", nil
			case strings.Contains(cmdStr, "git status --porcelain"):
				return "", nil // clean status
			case strings.Contains(cmdStr, "gh auth status"):
				return "Logged in to github.com", nil
			case strings.Contains(cmdStr, "git-cliff --config dot_config/git-cliff/cliff.toml --bumped-version"):
				return "v1.0.0", nil
			case strings.Contains(cmdStr, "git describe --tags --abbrev=0"):
				return "v1.0.0", nil
			default:
				return "", nil
			}
		},
	}
	state := newTestState(fake)
	err := RunRelease(context.Background(), state, true)
	if err != nil {
		t.Fatalf("RunRelease failed: %v", err)
	}
}
