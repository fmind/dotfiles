package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestRenderStatus(t *testing.T) {
	state := newTestState(&FakeRunner{})
	state.Config.Cluster.Name = "local"

	tests := []struct {
		name     string
		status   *SystemStatus
		contains []string
	}{
		{
			name:     "tools missing and no repositories",
			status:   &SystemStatus{},
			contains: []string{"Not installed.", "k3d not installed.", "No repositories found"},
		},
		{
			name: "tools installed but down",
			status: &SystemStatus{
				Docker: DockerStatus{Installed: true},
				K3d:    K3dStatus{Installed: true},
			},
			contains: []string{"Stopped or unreachable.", "Cluster 'local' does not exist or is stopped."},
		},
		{
			name: "everything up with a dirty and a broken repository",
			status: &SystemStatus{
				Docker: DockerStatus{Installed: true, Running: true, Details: "28.0.0"},
				K3d:    K3dStatus{Installed: true, Running: true, Details: "local 1/1 1/1 true"},
				Repositories: []RepoStatus{
					{Name: "dotfiles", ParentBase: "externals", Branch: "main", Dirty: true},
					{Name: "broken", ParentBase: "internals", Err: errors.New("boom")},
				},
			},
			contains: []string{
				"Running: 28.0.0",
				"Cluster 'local': local 1/1 1/1 true",
				"externals/dotfiles [main] [dirty]",
				// A repository that failed to probe must render as "error", never as a blank branch.
				"internals/broken [error]",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			state.Stdout = &stdout
			RenderStatus(tc.status, state)

			// Styling is applied unconditionally and stripped downstream, so compare
			// against the plain text a piped consumer would see.
			got := ansiPattern.ReplaceAllString(stdout.String(), "")
			for _, want := range tc.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, got)
				}
			}
		})
	}
}

func TestGatherK3dStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		listErr       error
		lookPath      func(string) (string, error)
		name          string
		list          string
		wantInstalled bool
		wantRunning   bool
	}{
		{
			name:     "k3d missing",
			lookPath: func(string) (string, error) { return "", errors.New("not found") },
		},
		{
			name:          "list fails",
			listErr:       errors.New("boom"),
			wantInstalled: true,
		},
		{
			name:          "other clusters are ignored",
			list:          "other 1/1 1/1 true\n",
			wantInstalled: true,
		},
		{
			name:          "running cluster",
			list:          "local 1/1 1/1 true\n",
			wantInstalled: true,
			wantRunning:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := newTestState(&FakeRunner{
				LookPathFunc: tc.lookPath,
				RunFunc: func(context.Context, string, io.Reader, string, ...string) (string, error) {
					return tc.list, tc.listErr
				},
			})
			state.Config.Cluster.Name = "local"

			got := gatherK3dStatus(ctx, state)
			if got.Installed != tc.wantInstalled || got.Running != tc.wantRunning {
				t.Fatalf("gatherK3dStatus = %+v, want installed=%v running=%v", got, tc.wantInstalled, tc.wantRunning)
			}
		})
	}
}
