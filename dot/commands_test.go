package dot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestClusterCommands(t *testing.T) {
	t.Run("start existing cluster", func(t *testing.T) {
		var startCalled, mergeCalled, waitCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				switch name {
				case "docker":
					if args[0] == "info" {
						return "docker ok", nil
					}
				case "k3d":
					if args[0] == "cluster" && args[1] == "list" {
						return "local", nil // cluster exists
					}
					if args[0] == "cluster" && args[1] == "start" {
						atomic.AddInt32(&startCalled, 1)
						return "started", nil
					}
					if args[0] == "kubeconfig" && args[1] == "merge" {
						atomic.AddInt32(&mergeCalled, 1)
						return "merged", nil
					}
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "kubectl" && args[1] == "wait" {
					atomic.AddInt32(&waitCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected interactive command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "start"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&startCalled) != 1 {
			t.Error("Expected k3d cluster start to be called once")
		}
		if atomic.LoadInt32(&mergeCalled) != 1 {
			t.Error("Expected k3d kubeconfig merge to be called once")
		}
		if atomic.LoadInt32(&waitCalled) != 1 {
			t.Error("Expected kubectl wait to be called once")
		}
	})

	t.Run("stop cluster", func(t *testing.T) {
		var stopCalled int32
		runner := &FakeRunner{
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "stop" {
					atomic.AddInt32(&stopCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "stop"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if atomic.LoadInt32(&stopCalled) != 1 {
			t.Error("Expected k3d cluster stop to be called")
		}
	})

	t.Run("stop cluster missing k3d", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("k3d not found")
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "stop"})
		if err == nil {
			t.Error("Expected error because k3d is not installed")
		}
	})

	t.Run("start non-existing cluster config not found", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "ok", nil
				}
				if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
					return "", nil // cluster does not exist
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		state.Config.Cluster.ConfigPath = "/tmp/non-existent-config-file-path.yaml"

		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "start"})
		if err == nil {
			t.Fatal("Expected error because cluster config file does not exist")
		}
		if !strings.Contains(err.Error(), "config file not found") {
			t.Errorf("Expected 'config file not found' error, got: %v", err)
		}
	})

	t.Run("start non-existing cluster success", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "k3d-config-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp config: %v", err)
		}
		defer func() { _ = os.Remove(tempFile.Name()) }()
		_ = tempFile.Close()

		var createCalled, mergeCalled, waitCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "ok", nil
				}
				if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
					return "", nil // cluster does not exist
				}
				if name == "k3d" && args[0] == "kubeconfig" && args[1] == "merge" {
					atomic.AddInt32(&mergeCalled, 1)
					return "merged", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "create" && args[3] == tempFile.Name() {
					atomic.AddInt32(&createCalled, 1)
					return nil
				}
				if name == "kubectl" && args[1] == "wait" {
					atomic.AddInt32(&waitCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected interactive command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		state.Config.Cluster.ConfigPath = tempFile.Name()

		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err = app.Run(context.Background(), []string{"dot", "cluster", "start"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&createCalled) != 1 {
			t.Error("Expected k3d cluster create to be called once")
		}
		if atomic.LoadInt32(&mergeCalled) != 1 {
			t.Error("Expected k3d kubeconfig merge to be called once")
		}
		if atomic.LoadInt32(&waitCalled) != 1 {
			t.Error("Expected kubectl wait to be called once")
		}
	})

	t.Run("delete cluster", func(t *testing.T) {
		var deleteCalled int32
		runner := &FakeRunner{
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "delete" {
					atomic.AddInt32(&deleteCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "delete", "--yes"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if atomic.LoadInt32(&deleteCalled) != 1 {
			t.Error("Expected k3d cluster delete to be called")
		}
	})

	t.Run("delete cluster declined at prompt does not delete", func(t *testing.T) {
		var deleteCalled int32
		runner := &FakeRunner{
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "delete" {
					atomic.AddInt32(&deleteCalled, 1)
				}
				return nil
			},
		}
		state := newTestState(runner)
		state.Stdin = strings.NewReader("n\n")
		var out strings.Builder
		state.Stdout = &out
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}
		// Without --yes, answering 'n' must abort teardown of the SHARED local cluster
		// entirely. This guards the confirmation prompt so it cannot silently regress.
		err := app.Run(context.Background(), []string{"dot", "cluster", "delete"})
		if err != nil {
			t.Fatalf("Expected no error when declining, got %v", err)
		}
		if atomic.LoadInt32(&deleteCalled) != 0 {
			t.Error("Expected k3d cluster delete NOT to be called when the confirmation is declined")
		}
		if !strings.Contains(out.String(), "Canceled") {
			t.Errorf("Expected 'Canceled.' in output, got %q", out.String())
		}
	})

	t.Run("delete cluster missing k3d", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("k3d not found")
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "delete"})
		if err == nil {
			t.Error("Expected error because k3d is not installed")
		}
	})

	t.Run("status cluster", func(t *testing.T) {
		var listCalled, getCalled int32
		runner := &FakeRunner{
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
					atomic.AddInt32(&listCalled, 1)
					return nil
				}
				if name == "kubectl" && args[0] == "get" && args[1] == "nodes" {
					atomic.AddInt32(&getCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "status"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if atomic.LoadInt32(&listCalled) != 1 {
			t.Error("Expected k3d cluster list to be called")
		}
		if atomic.LoadInt32(&getCalled) != 1 {
			t.Error("Expected kubectl get nodes to be called")
		}
	})

	t.Run("status cluster missing kubectl", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "kubectl" {
					return "", errors.New("kubectl not found")
				}
				return "/bin/" + name, nil
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "status"})
		if err == nil {
			t.Error("Expected error because kubectl is not installed")
		}
	})

	t.Run("namespace cluster create new and set context", func(t *testing.T) {
		var getCalled, createCalled, setCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "kubectl" {
					if args[0] == "get" && args[1] == "namespace" && args[2] == "test-ns" {
						atomic.AddInt32(&getCalled, 1)
						return "", nil // --ignore-not-found: empty output, no error = absent
					}
					if args[0] == "create" && args[1] == "namespace" && args[2] == "test-ns" {
						atomic.AddInt32(&createCalled, 1)
						return "created", nil
					}
					if args[0] == "config" && args[1] == "set-context" && args[4] == "test-ns" {
						atomic.AddInt32(&setCalled, 1)
						return "context set", nil
					}
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewClusterCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "cluster", "namespace", "test-ns"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if atomic.LoadInt32(&getCalled) != 1 {
			t.Error("Expected kubectl get namespace to be called")
		}
		if atomic.LoadInt32(&createCalled) != 1 {
			t.Error("Expected kubectl create namespace to be called")
		}
		if atomic.LoadInt32(&setCalled) != 1 {
			t.Error("Expected kubectl config set-context to be called")
		}
	})
}

func TestCommitCommand(t *testing.T) {
	t.Run("successful commit with cached changes", func(t *testing.T) {
		var gitCommitCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if args[0] == "diff" && args[1] == "--cached" {
						return "some git diff content", nil
					}
				}
				if name == "/bin/agy" {
					diffBytes, _ := io.ReadAll(stdin)
					if string(diffBytes) != "some git diff content" {
						return "", fmt.Errorf("expected diff input, got %q", string(diffBytes))
					}
					return "feat(ui): add button", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "git" && args[0] == "commit" && args[3] == "feat(ui): add button" {
					atomic.AddInt32(&gitCommitCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&gitCommitCalled) != 1 {
			t.Error("Expected git commit to be called once")
		}
	})

	t.Run("successful commit with unstaged changes when no cached changes", func(t *testing.T) {
		var gitCommitCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if args[0] == "diff" && args[1] == "--cached" {
						return "", nil // no cached changes
					}
					if args[0] == "diff" && args[1] == "--" {
						return "unstaged changes content", nil
					}
				}
				if name == "/bin/agy" {
					return "fix(api): handle empty input", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				// Should include -a for commit
				if name == "git" && args[0] == "commit" && args[1] == "-a" && args[4] == "fix(api): handle empty input" {
					atomic.AddInt32(&gitCommitCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&gitCommitCalled) != 1 {
			t.Error("Expected git commit -a to be called once")
		}
	})

	t.Run("successful commit with both type and scope, verifying prompt", func(t *testing.T) {
		var gitCommitCalled int32
		var promptUsed string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if args[0] == "diff" && args[1] == "--cached" {
						return "some git diff content", nil
					}
				}
				if name == "/bin/agy" {
					for i, arg := range args {
						if arg == "--prompt" && i+1 < len(args) {
							promptUsed = args[i+1]
						}
					}
					return "feat(ui): add button", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "git" && args[0] == "commit" && args[3] == "feat(ui): add button" {
					atomic.AddInt32(&gitCommitCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit", "--type", "feat", "--scope", "ui"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&gitCommitCalled) != 1 {
			t.Error("Expected git commit to be called once")
		}
		if !strings.Contains(promptUsed, "Use type 'feat' and scope 'ui'") {
			t.Errorf("Expected prompt to contain type and scope instruction, got %q", promptUsed)
		}
	})

	t.Run("successful commit with scope only, verifying prompt", func(t *testing.T) {
		var gitCommitCalled int32
		var promptUsed string
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if args[0] == "diff" && args[1] == "--cached" {
						return "some git diff content", nil
					}
				}
				if name == "/bin/agy" {
					for i, arg := range args {
						if arg == "--prompt" && i+1 < len(args) {
							promptUsed = args[i+1]
						}
					}
					return "feat(ui): add button", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "git" && args[0] == "commit" && args[3] == "feat(ui): add button" {
					atomic.AddInt32(&gitCommitCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit", "--scope", "ui"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&gitCommitCalled) != 1 {
			t.Error("Expected git commit to be called once")
		}
		if !strings.Contains(promptUsed, "Use scope 'ui' and suggest an appropriate type") {
			t.Errorf("Expected prompt to contain scope only instruction, got %q", promptUsed)
		}
	})

	t.Run("fails when not in a git repository", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" && args[0] == "rev-parse" {
					return "", errors.New("fatal: not a git repository")
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit"})
		if !errors.Is(err, ErrNotGitRepository) {
			t.Fatalf("Expected ErrNotGitRepository, got %v", err)
		}
	})

	t.Run("fails when AI client is not found in PATH", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("file not found")
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if args[0] == "diff" {
						return "some diff content", nil
					}
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit"})
		if err == nil || !strings.Contains(err.Error(), "CLI is not installed or not in PATH") {
			t.Fatalf("Expected installation error, got %v", err)
		}
	})

	t.Run("fails when AI returns empty output", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "rev-parse" {
						return "true", nil
					}
					if args[0] == "diff" && args[1] == "--cached" {
						return "some git diff content", nil
					}
				}
				if name == "/bin/agy" {
					return "", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewCommitCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "commit"})
		if err == nil || !strings.Contains(err.Error(), "returned empty output") {
			t.Fatalf("Expected empty output error, got %v", err)
		}
	})
}

func TestSetupCommand(t *testing.T) {
	t.Run("setup workspace success", func(t *testing.T) {
		var gcloudCalled, gwsCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "gcloud" && args[0] == "config" && args[1] == "get-value" {
					return "test-project-123\n", nil
				}
				return "", fmt.Errorf("unexpected command: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gcloud" && args[0] == "services" && args[1] == "enable" {
					atomic.AddInt32(&gcloudCalled, 1)
					// Verify project argument is passed
					projArgFound := false
					for _, arg := range args {
						if arg == "test-project-123" {
							projArgFound = true
						}
					}
					if !projArgFound {
						return fmt.Errorf("expected project ID to be enabled, args: %v", args)
					}
					return nil
				}
				if name == "gws" && args[0] == "auth" && args[1] == "setup" && args[3] == "test-project-123" {
					atomic.AddInt32(&gwsCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		t.Setenv(EnvGWSProject, "")

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewSetupCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "setup", "workspace", "test-project-123"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&gcloudCalled) != 1 {
			t.Error("Expected gcloud services enable to be called")
		}
		if atomic.LoadInt32(&gwsCalled) != 1 {
			t.Error("Expected gws auth setup to be called")
		}
	})
}

func TestLoginCommand(t *testing.T) {
	t.Run("login workspace success", func(t *testing.T) {
		var loginCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gws" && args[0] == "auth" && args[1] == "login" {
					atomic.AddInt32(&loginCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewLoginCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "login", "workspace"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&loginCalled) != 1 {
			t.Error("Expected gws auth login to be called")
		}
	})

	t.Run("login github already authenticated, cancel", func(t *testing.T) {
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "gh" && args[0] == "auth" && args[1] == "status" {
					return "Logged in to github.com", nil // already authenticated
				}
				return "", fmt.Errorf("unexpected Run: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		state.Stdin = strings.NewReader("n\n")

		app := &cli.Command{
			Commands: []*cli.Command{
				NewLoginCmd(state),
			},
		}
		err := app.Run(context.Background(), []string{"dot", "login", "github"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("login github not authenticated, success", func(t *testing.T) {
		var loginCalled int32
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "gh" && args[0] == "auth" && args[1] == "status" {
					return "", errors.New("not logged in") // not authenticated
				}
				return "", fmt.Errorf("unexpected Run: %s %v", name, args)
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gh" && args[0] == "auth" && args[1] == "login" {
					atomic.AddInt32(&loginCalled, 1)
					return nil
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewLoginCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "login", "github"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&loginCalled) != 1 {
			t.Error("Expected gh auth login to be called")
		}
	})

	t.Run("login github missing gh", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("gh not found")
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewLoginCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "login", "github"})
		if err == nil {
			t.Error("Expected error because gh is not installed")
		}
	})

	t.Run("login gcp success", func(t *testing.T) {
		var loginCalled int32
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gcloud" && args[0] == "auth" && args[1] == "login" {
					// Verify --update-adc is passed
					for _, a := range args {
						if a == "--update-adc" {
							atomic.AddInt32(&loginCalled, 1)
							return nil
						}
					}
					return fmt.Errorf("expected --update-adc flag, got args: %v", args)
				}
				return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewLoginCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "login", "gcp"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if atomic.LoadInt32(&loginCalled) != 1 {
			t.Error("Expected gcloud auth login --update-adc to be called")
		}
	})

	t.Run("login gcp missing gcloud", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("gcloud not found")
			},
		}

		state := newTestState(runner)
		app := &cli.Command{
			Commands: []*cli.Command{
				NewLoginCmd(state),
			},
		}

		err := app.Run(context.Background(), []string{"dot", "login", "gcp"})
		if err == nil {
			t.Error("Expected error because gcloud is not installed")
		} else if !errors.Is(err, ErrGcloudNotInstalled) {
			t.Errorf("Expected error to be ErrGcloudNotInstalled, got %v", err)
		}
	})
}

func TestPullCommand(t *testing.T) {
	t.Run("pull success", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "pull-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create mock repo directories
		repo1 := filepath.Join(tempDir, "repo1")
		_ = os.MkdirAll(filepath.Join(repo1, ".git"), 0o755)

		var pullCalled int32
		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "main\n", nil
					}
					if args[0] == "status" && args[1] == "--porcelain" {
						return "", nil // clean repo
					}
					if args[0] == "fetch" {
						return "", nil
					}
					if args[0] == "rev-list" {
						return "0\n", nil
					}
					if args[0] == "pull" {
						atomic.AddInt32(&pullCalled, 1)
						return "Already up to date.", nil
					}
				}
				return "", fmt.Errorf("unexpected Run: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}

		app := &cli.Command{
			Commands: []*cli.Command{
				NewPullCmd(state),
			},
		}

		err = app.Run(context.Background(), []string{"dot", "pull"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if atomic.LoadInt32(&pullCalled) != 1 {
			t.Error("Expected git pull to be called once")
		}
	})

	t.Run("pull sorting success", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "pull-sorting-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create mock repo directories out of order (b_repo and a_repo)
		repoB := filepath.Join(tempDir, "b_repo")
		_ = os.MkdirAll(filepath.Join(repoB, ".git"), 0o755)
		repoA := filepath.Join(tempDir, "a_repo")
		_ = os.MkdirAll(filepath.Join(repoA, ".git"), 0o755)

		runner := &FakeRunner{
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "git" {
					if args[0] == "branch" && args[1] == "--show-current" {
						return "main\n", nil
					}
					if args[0] == "status" && args[1] == "--porcelain" {
						return "", nil
					}
					if args[0] == "fetch" {
						return "", nil
					}
					if args[0] == "rev-list" {
						return "0\n", nil
					}
					if args[0] == "pull" {
						return "Already up to date.", nil
					}
				}
				return "", fmt.Errorf("unexpected Run: %s %v", name, args)
			},
		}

		state := newTestState(runner)
		state.Config.Pull.Directories = []string{tempDir}
		var buf strings.Builder
		state.Stdout = &buf

		app := &cli.Command{
			Commands: []*cli.Command{
				NewPullCmd(state),
			},
		}

		err = app.Run(context.Background(), []string{"dot", "pull"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		output := buf.String()
		// The sorted output should print a_repo before b_repo
		idxA := strings.Index(output, "a_repo")
		idxB := strings.Index(output, "b_repo")
		if idxA == -1 || idxB == -1 {
			t.Fatalf("Expected output to contain a_repo and b_repo, got: %q", output)
		}
		if idxA > idxB {
			t.Errorf("Expected a_repo to be printed before b_repo (sorted), but a_repo index %d is after b_repo index %d", idxA, idxB)
		}
	})
}

func TestPrCommand(t *testing.T) {
	var gitCalled, agyCalled, ghCalled int32
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "git" {
				if args[0] == "rev-parse" && args[1] == "--is-inside-work-tree" {
					return "true", nil
				}
				if args[0] == "diff" && strings.HasPrefix(args[1], "main...") {
					atomic.AddInt32(&gitCalled, 1)
					return "some diff content", nil
				}
			}
			if name == "/bin/agy" && args[0] == "--prompt" {
				atomic.AddInt32(&agyCalled, 1)
				return "pr description text", nil
			}
			return "", fmt.Errorf("unexpected Run: %s %v", name, args)
		},
		RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
			if name == "/bin/gh" && args[0] == "pr" && args[1] == "create" {
				atomic.AddInt32(&ghCalled, 1)
				return nil
			}
			return fmt.Errorf("unexpected RunInteractive: %s %v", name, args)
		},
	}

	state := newTestState(runner)
	app := &cli.Command{
		Commands: []*cli.Command{
			NewPrCmd(state),
		},
	}

	err := app.Run(context.Background(), []string{"dot", "pr"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if atomic.LoadInt32(&gitCalled) != 1 {
		t.Error("Expected git diff to be called once")
	}
	if atomic.LoadInt32(&agyCalled) != 1 {
		t.Error("Expected agy to be called once")
	}
	if atomic.LoadInt32(&ghCalled) != 1 {
		t.Error("Expected gh pr create to be called once")
	}
}

func TestStatusCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "status-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	repo1 := filepath.Join(tempDir, "repo1")
	_ = os.MkdirAll(filepath.Join(repo1, ".git"), 0o755)

	var dockerCalled, k3dCalled, gitBranchCalled, gitStatusCalled int32
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "docker" && args[0] == "info" {
				atomic.AddInt32(&dockerCalled, 1)
				return "my-docker-daemon", nil
			}
			if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
				atomic.AddInt32(&k3dCalled, 1)
				return "local   1/1   1/1   true", nil
			}
			if name == "git" {
				if args[0] == "branch" && args[1] == "--show-current" && dir == repo1 {
					atomic.AddInt32(&gitBranchCalled, 1)
					return "feature-branch\n", nil
				}
				if args[0] == "status" && args[1] == "--porcelain" && dir == repo1 {
					atomic.AddInt32(&gitStatusCalled, 1)
					return "M file.go\n", nil
				}
			}
			return "", fmt.Errorf("unexpected command: %s %v", name, args)
		},
	}

	state := newTestState(runner)
	state.Config.Pull.Directories = []string{tempDir}
	var buf strings.Builder
	state.Stdout = &buf

	app := &cli.Command{
		Commands: []*cli.Command{
			NewStatusCmd(state),
		},
	}

	err = app.Run(context.Background(), []string{"dot", "status"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if atomic.LoadInt32(&dockerCalled) != 1 {
		t.Error("Expected docker info to be called")
	}
	if atomic.LoadInt32(&k3dCalled) != 1 {
		t.Error("Expected k3d cluster list to be called")
	}
	if atomic.LoadInt32(&gitBranchCalled) != 1 {
		t.Error("Expected git branch to be called")
	}
	if atomic.LoadInt32(&gitStatusCalled) != 1 {
		t.Error("Expected git status to be called")
	}

	output := buf.String()
	if !strings.Contains(output, "repo1") {
		t.Errorf("Expected status output to contain 'repo1', got: %s", output)
	}
	if !strings.Contains(output, "feature-branch") {
		t.Errorf("Expected status output to contain 'feature-branch', got: %s", output)
	}
	if !strings.Contains(output, "dirty") {
		t.Errorf("Expected status output to contain 'dirty', got: %s", output)
	}
}

func TestGatherStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gather-status-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	repo1 := filepath.Join(tempDir, "repo1")
	_ = os.MkdirAll(filepath.Join(repo1, ".git"), 0o755)

	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			if name == "docker" && args[0] == "info" {
				return "my-docker-daemon", nil
			}
			if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
				return "local   1/1   1/1   true", nil
			}
			if name == "git" {
				if args[0] == "branch" && args[1] == "--show-current" && dir == repo1 {
					return "feature-branch\n", nil
				}
				if args[0] == "status" && args[1] == "--porcelain" && dir == repo1 {
					return "M file.go\n", nil
				}
			}
			return "", fmt.Errorf("unexpected command: %s %v", name, args)
		},
	}

	state := newTestState(runner)
	state.Config.Pull.Directories = []string{tempDir}

	status, err := GatherStatus(context.Background(), state)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !status.Docker.Installed || !status.Docker.Running || status.Docker.Details != "my-docker-daemon" {
		t.Errorf("Unexpected docker status: %+v", status.Docker)
	}

	if !status.K3d.Installed || !status.K3d.Running || status.K3d.Details != "local   1/1   1/1   true" {
		t.Errorf("Unexpected k3d status: %+v", status.K3d)
	}

	if len(status.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(status.Repositories))
	}

	repo := status.Repositories[0]
	if repo.Name != "repo1" || repo.Branch != "feature-branch" || !repo.Dirty || repo.Err != nil {
		t.Errorf("Unexpected repository status: %+v", repo)
	}
}
