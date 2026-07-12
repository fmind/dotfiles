package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRunStatus_JSON(t *testing.T) {
	// A repository whose branch lookup fails exercises the RepoStatus.Error serialization.
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "brokenrepo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) { return "/bin/" + name, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			switch name {
			case "docker":
				return "penguin (Containers: 1, Running: 1)", nil
			case "k3d":
				return "local   1/1   1/1   true", nil
			case "git":
				if len(args) > 0 && args[0] == "branch" {
					return "", errors.New("not a git repository")
				}
			}
			return "", nil
		},
	}
	state := newTestState(runner)
	state.Config.Pull.Directories = []string{tempDir}
	var buf bytes.Buffer
	state.Stdout = &buf

	if err := RunStatus(context.Background(), state, true); err != nil {
		t.Fatalf("RunStatus json: %v", err)
	}

	var got SystemStatus
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v (%q)", err, buf.String())
	}
	if !got.Docker.Installed || !got.Docker.Running {
		t.Errorf("expected docker installed+running, got %+v", got.Docker)
	}
	if !got.K3d.Running {
		t.Errorf("expected k3d running, got %+v", got.K3d)
	}
	if len(got.Repositories) != 1 || got.Repositories[0].Error == "" {
		t.Errorf("expected one repository carrying a serialized error, got %+v", got.Repositories)
	}
}

func TestGatherK3dStatus_StoppedNotRunning(t *testing.T) {
	// A stopped-but-existing cluster still lists, with SERVERS "0/1"; it must not
	// be reported as running just because the name matches.
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) { return "/bin/" + name, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, _ ...string) (string, error) {
			if name == "k3d" {
				return "local   0/1   0/1   false", nil
			}
			return "", nil
		},
	}
	got := gatherK3dStatus(context.Background(), newTestState(runner))
	if !got.Installed {
		t.Errorf("expected k3d installed, got %+v", got)
	}
	if got.Running {
		t.Errorf("expected stopped cluster to report not running, got %+v", got)
	}
}

func TestGatherRepoStatus_StatusFailureIsReported(t *testing.T) {
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name == "git" && args[0] == "branch" {
				return "main\n", nil
			}
			if name == "git" && args[0] == "status" {
				return "", errors.New("status unavailable")
			}
			return "", nil
		},
	}

	got := gatherRepoStatus(context.Background(), newTestState(runner), t.TempDir())
	if got.Err == nil || got.Error == "" {
		t.Fatalf("expected repository status error, got %+v", got)
	}
	if got.Dirty {
		t.Fatalf("repository with unknown status must not be reported dirty or clean: %+v", got)
	}
}
