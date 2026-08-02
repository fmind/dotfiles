package dot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func clusterTestKubeconfig(name, contextName, server string) string {
	return fmt.Sprintf("apiVersion: v1\nclusters:\n  - name: k3d-%[1]s\n    cluster:\n      server: %[3]s\ncontexts:\n  - name: %[2]s\n    context:\n      cluster: k3d-%[1]s\n      namespace: default\ncurrent-context: %[2]s\n", name, contextName, server)
}

func configureClusterTestTarget(t *testing.T, state *GlobalState, runner *FakeRunner) {
	t.Helper()
	name := state.Config.Cluster.Name
	contextName := expectedClusterContext(name)
	content := clusterTestKubeconfig(name, contextName, "https://127.0.0.1:6443")
	path := filepath.Join(t.TempDir(), name+".yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	state.Config.Cluster.KubeconfigPath = path
	originalRun := runner.RunFunc
	runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, command string, args ...string) (string, error) {
		if command == "docker" && slices.Equal(args, []string{"info"}) {
			return "docker ok", nil
		}
		if command == "k3d" && slices.Equal(args, []string{"kubeconfig", "get", name}) {
			return content, nil
		}
		if command == "kubectl" && slices.Contains(args, "--raw=/version") {
			return `{"gitVersion":"v1"}`, nil
		}
		if originalRun != nil {
			return originalRun(ctx, dir, stdin, command, args...)
		}
		return "", nil
	}
}

func TestRunClusterNamespace_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("empty name", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		err := RunClusterNamespace(ctx, state, "")
		if err == nil || err.Error() != "namespace name is required" {
			t.Errorf("Expected 'namespace name is required', got %v", err)
		}
	})

	t.Run("kubectl not installed", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "kubectl" {
					return "", errors.New("not found")
				}
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterNamespace(ctx, state, "test-ns")
		if !errors.Is(err, ErrToolNotInstalled) {
			t.Errorf("Expected ErrToolNotInstalled, got %v", err)
		}
	})

	t.Run("kubectl create namespace fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "kubectl" {
					if args[0] == "get" && args[1] == "namespace" {
						return "", nil // --ignore-not-found: empty output, no error = absent
					}
					if args[0] == "create" && args[1] == "namespace" {
						return "", errors.New("create namespace error")
					}
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterNamespace(ctx, state, "test-ns")
		if err == nil || !strings.Contains(err.Error(), "failed to create namespace 'test-ns'") {
			t.Errorf("Expected create namespace error, got %v", err)
		}
	})

	t.Run("kubectl get namespace query fails (unreachable/unauthorized)", func(t *testing.T) {
		var createCalled bool
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "kubectl" {
					if args[0] == "get" && args[1] == "namespace" {
						// A real failure (not a missing namespace) must NOT be treated as "not found".
						return "", errors.New("Unable to connect to the server")
					}
					if args[0] == "create" {
						createCalled = true
					}
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterNamespace(ctx, state, "test-ns")
		if err == nil || !strings.Contains(err.Error(), "failed to query namespace 'test-ns'") {
			t.Errorf("Expected query namespace error, got %v", err)
		}
		if createCalled {
			t.Error("Expected create NOT to be called when the query itself fails")
		}
	})

	t.Run("kubectl config set-context fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "kubectl" {
					if args[0] == "get" && args[1] == "namespace" {
						return "test-ns", nil
					}
					if args[0] == "config" && args[1] == "set-context" {
						return "", errors.New("set-context error")
					}
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterNamespace(ctx, state, "test-ns")
		if err == nil || !strings.Contains(err.Error(), "failed to set context namespace") {
			t.Errorf("Expected set-context error, got %v", err)
		}
	})
}

func TestRunClusterStart_Scenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("dependency docker missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "docker" {
					return "", errors.New("not found")
				}
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		err := RunClusterStart(ctx, state)
		if !errors.Is(err, ErrToolNotInstalled) {
			t.Errorf("Expected ErrToolNotInstalled, got %v", err)
		}
	})

	t.Run("docker daemon not running", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "", errors.New("docker daemon down")
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		err := RunClusterStart(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "docker daemon is not running") {
			t.Errorf("Expected docker daemon down error, got %v", err)
		}
	})

	t.Run("k3d start failure", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "info", nil
				}
				if name == "k3d" {
					if args[0] == "cluster" && args[1] == "list" {
						return "my-cluster", nil // cluster exists
					}
					if args[0] == "cluster" && args[1] == "start" {
						return "", errors.New("start failed")
					}
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		state.Config.Cluster.Name = "my-cluster"
		configureClusterTestTarget(t, state, runner)
		err := RunClusterStart(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "failed to start cluster") {
			t.Errorf("Expected failed to start cluster error, got %v", err)
		}
	})

	t.Run("k3d create config file missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "info", nil
				}
				if name == "k3d" {
					if args[0] == "cluster" && args[1] == "list" {
						return "", nil // cluster doesn't exist
					}
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		state.Config.Cluster.ConfigPath = "/nonexistent/k3d.yaml"
		err := RunClusterStart(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "config file not found at") {
			t.Errorf("Expected config file missing error, got %v", err)
		}
	})

	t.Run("k3d create interactive run fails", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "k3d.yaml")
		_ = os.WriteFile(configFile, []byte(""), 0o644)

		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "info", nil
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "create" {
					return errors.New("create failed")
				}
				return nil
			},
		}
		state := newTestState(runner)
		state.Config.Cluster.ConfigPath = configFile
		err := RunClusterStart(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "failed to create cluster") {
			t.Errorf("Expected failed to create cluster error, got %v", err)
		}
	})

	t.Run("k3d missing cluster still reaches creation", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "k3d.yaml")
		_ = os.WriteFile(configFile, []byte(""), 0o644)

		clusterName := "regression"
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "info", nil
				}
				if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
					// Real k3d exits non-zero for `cluster list <unknown-name>`, so
					// the existence probe must use the unfiltered list instead.
					if slices.Contains(args, clusterName) {
						return "", errors.New("No nodes found for given cluster")
					}
					return "other 1/1 1/1 true", nil
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "create" {
					return errors.New("create failed")
				}
				return nil
			},
		}
		state := newTestState(runner)
		state.Config.Cluster.Name = clusterName
		state.Config.Cluster.ConfigPath = configFile
		err := RunClusterStart(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "failed to create cluster") {
			t.Errorf("Expected creation attempt for missing cluster, got %v", err)
		}
	})

	t.Run("k3d kubeconfig retrieval fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "info", nil
				}
				if name == "k3d" {
					if args[0] == "cluster" && args[1] == "list" {
						return "my-cluster", nil
					}
					if args[0] == "cluster" && args[1] == "start" {
						return "started", nil
					}
					if args[0] == "kubeconfig" && args[1] == "get" {
						return "", errors.New("kubeconfig failed")
					}
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		state.Config.Cluster.Name = "my-cluster"
		err := RunClusterStart(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "failed to get kubeconfig") {
			t.Errorf("Expected kubeconfig retrieval error, got %v", err)
		}
	})

	t.Run("kubectl wait nodes condition fails writes warning to Stderr", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "info", nil
				}
				if name == "k3d" {
					if args[0] == "cluster" && args[1] == "list" {
						return "my-cluster", nil
					}
					if args[0] == "cluster" && args[1] == "start" {
						return "started", nil
					}
				}
				return "", nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "kubectl" && args[1] == "wait" {
					return errors.New("timeout waiting for condition")
				}
				return nil
			},
		}
		state := newTestState(runner)
		state.Config.Cluster.Name = "my-cluster"
		configureClusterTestTarget(t, state, runner)
		var errBuf bytes.Buffer
		state.Stderr = &errBuf

		err := RunClusterStart(ctx, state)
		if err != nil {
			t.Fatalf("Expected no error from RunClusterStart even if nodes wait fails, got %v", err)
		}
		if !strings.Contains(errBuf.String(), "Warning: Nodes are not fully ready yet") {
			t.Errorf("Expected Warning: Nodes are not fully ready yet in stderr, got %q", errBuf.String())
		}
	})
}

func TestRunClusterStatus_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("k3d dependency missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "k3d" {
					return "", errors.New("not found")
				}
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterStatus(ctx, state)
		if !errors.Is(err, ErrToolNotInstalled) {
			t.Errorf("Expected ErrToolNotInstalled, got %v", err)
		}
	})

	t.Run("k3d cluster list interactive run fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "k3d" && args[0] == "cluster" && args[1] == "list" {
					return errors.New("list failed")
				}
				return nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterStatus(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "failed to list k3d clusters") {
			t.Errorf("Expected list failed error, got %v", err)
		}
	})

	t.Run("kubectl get nodes interactive run fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "kubectl" && args[0] == "get" && args[1] == "nodes" {
					return errors.New("get nodes failed")
				}
				return nil
			},
		}
		state := newTestState(runner)
		configureClusterTestTarget(t, state, runner)
		err := RunClusterStatus(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "failed to get kubectl nodes") {
			t.Errorf("Expected get nodes failed error, got %v", err)
		}
	})
}
