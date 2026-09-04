package dot

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/urfave/cli/v3"
)

const clusterTestServer = "https://127.0.0.1:6443"

func writeClusterTargetFixture(t *testing.T, state *GlobalState, contextName, server string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), state.Config.Cluster.Name+".yaml")
	content := clusterTestKubeconfig(state.Config.Cluster.Name, contextName, server)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	state.Config.Cluster.KubeconfigPath = path
	return path
}

func newClusterTargetRunner(name string, apiErr error) *FakeRunner {
	authoritative := clusterTestKubeconfig(name, expectedClusterContext(name), clusterTestServer)
	return &FakeRunner{
		LookPathFunc: func(command string) (string, error) { return "/bin/" + command, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, command string, args ...string) (string, error) {
			switch {
			case command == "docker" && slices.Equal(args, []string{"info"}):
				return "ok", nil
			case command == "k3d" && slices.Equal(args, []string{"kubeconfig", "get", name}):
				return authoritative, nil
			case command == "kubectl" && slices.Contains(args, "--raw=/version"):
				if apiErr != nil {
					return "", apiErr
				}
				return `{"gitVersion":"v1"}`, nil
			default:
				return "", nil
			}
		},
	}
}

func TestClusterTargetUsesPerClusterOwnerOnlyKubeconfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defaultKubeconfig := filepath.Join(home, ".kube", "config")
	if err := os.MkdirAll(filepath.Dir(defaultKubeconfig), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(defaultKubeconfig, []byte("user-owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", defaultKubeconfig)

	for _, name := range []string{"local", "other"} {
		state := newTestState(newClusterTargetRunner(name, nil))
		state.Config.Cluster.Name = name
		path, err := refreshClusterKubeconfig(context.Background(), state, ClusterTargetOptions{})
		if err != nil {
			t.Fatalf("refresh %s: %v", name, err)
		}
		want := filepath.Join(home, ".kube", "dot", name+".yaml")
		if path != want {
			t.Fatalf("kubeconfig path = %s, want %s", path, want)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("kubeconfig permissions = %o, want 600", info.Mode().Perm())
		}
	}
	content, err := os.ReadFile(defaultKubeconfig)
	if err != nil || string(content) != "user-owned" {
		t.Fatalf("default kubeconfig changed: %q %v", content, err)
	}
}

func TestWriteOwnerOnlyFilePreservesExistingParentPermissions(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "diagnostic.json")

	if err := writeOwnerOnlyFile(path, []byte("private"), "diagnostic manifest"); err != nil {
		t.Fatal(err)
	}
	parentInfo, err := os.Stat(parent)
	if err != nil {
		t.Fatal(err)
	}
	if got := parentInfo.Mode().Perm(); got != 0o755 {
		t.Fatalf("existing parent permissions = %o, want 755", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("owner-only file permissions = %o, want 600", got)
	}
}

func TestResolveClusterTargetContextOverrideAndMismatch(t *testing.T) {
	const server = clusterTestServer

	t.Run("explicit renamed context", func(t *testing.T) {
		state := newTestState(newClusterTargetRunner("local", nil))
		writeClusterTargetFixture(t, state, "renamed-local", server)
		target, err := resolveClusterTarget(context.Background(), state, ClusterTargetOptions{Context: "renamed-local"}, true)
		if err != nil {
			t.Fatalf("resolve explicit context: %v", err)
		}
		if target.Context != "renamed-local" || target.Fingerprint != clusterOwnershipFingerprint("local") {
			t.Fatalf("unexpected resolved target: %+v", target)
		}
	})

	t.Run("renamed context without override", func(t *testing.T) {
		state := newTestState(newClusterTargetRunner("local", nil))
		writeClusterTargetFixture(t, state, "renamed-local", server)
		_, err := resolveClusterTarget(context.Background(), state, ClusterTargetOptions{}, false)
		if err == nil || !strings.Contains(err.Error(), `expected context "k3d-local", actual current context "renamed-local"`) {
			t.Fatalf("expected explicit mismatch, got %v", err)
		}
	})

	t.Run("server mismatch", func(t *testing.T) {
		state := newTestState(newClusterTargetRunner("local", nil))
		writeClusterTargetFixture(t, state, expectedClusterContext("local"), "https://user:private@remote.example.invalid:6443?token=secret")
		_, err := resolveClusterTarget(context.Background(), state, ClusterTargetOptions{}, false)
		if err == nil || !strings.Contains(err.Error(), `expected context "k3d-local" server`) || !strings.Contains(err.Error(), `actual context "k3d-local" server`) {
			t.Fatalf("expected server mismatch, got %v", err)
		}
		if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "secret") {
			t.Fatalf("context mismatch exposed credentials: %v", err)
		}
	})
}

func TestClusterCommandAcceptsExplicitContextOverride(t *testing.T) {
	runner := newClusterTargetRunner("local", nil)
	runner.RunInteractiveFunc = func(_ context.Context, _, command string, args ...string) error {
		if command == "k3d" && slices.Equal(args, []string{"cluster", "list", "local"}) {
			return nil
		}
		if command == "kubectl" && len(args) > 1 && args[0] == "get" && args[1] == "nodes" && slices.Contains(args, "renamed-local") {
			return nil
		}
		return errors.New("unexpected interactive command")
	}
	state := newTestState(runner)
	writeClusterTargetFixture(t, state, "renamed-local", clusterTestServer)
	app := &cli.Command{Commands: []*cli.Command{NewClusterCmd(state)}}
	if err := app.Run(context.Background(), []string{"dot", "cluster", "--context", "renamed-local", "status"}); err != nil {
		t.Fatalf("explicit context command failed: %v", err)
	}
}

func TestClusterMutationFailsClosedWhenAPIUnavailable(t *testing.T) {
	runner := newClusterTargetRunner("local", errors.New("connection refused"))
	var stopCalls atomic.Int32
	runner.RunInteractiveFunc = func(_ context.Context, _, command string, args ...string) error {
		if command == "k3d" && slices.Contains(args, "stop") {
			stopCalls.Add(1)
		}
		return nil
	}
	state := newTestState(runner)
	writeClusterTargetFixture(t, state, expectedClusterContext("local"), "https://127.0.0.1:6443")
	err := RunClusterStop(context.Background(), state)
	if err == nil || !strings.Contains(err.Error(), `expected context "k3d-local", actual context "k3d-local"`) {
		t.Fatalf("expected unavailable API error, got %v", err)
	}
	if stopCalls.Load() != 0 {
		t.Fatal("stop mutation ran after API verification failed")
	}
}

func TestClusterNamespaceReverifiesImmediatelyBeforeMutation(t *testing.T) {
	const server = "https://127.0.0.1:6443"
	authoritative := clusterTestKubeconfig("local", expectedClusterContext("local"), server)
	var apiChecks, creates atomic.Int32
	runner := &FakeRunner{
		LookPathFunc: func(command string) (string, error) { return "/bin/" + command, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, command string, args ...string) (string, error) {
			switch {
			case command == "docker":
				return "ok", nil
			case command == "k3d" && slices.Equal(args, []string{"kubeconfig", "get", "local"}):
				return authoritative, nil
			case command == "kubectl" && slices.Contains(args, "--raw=/version"):
				if apiChecks.Add(1) == 2 {
					return "", errors.New("target changed")
				}
				return "version", nil
			case command == "kubectl" && len(args) > 1 && args[0] == "get" && args[1] == "namespace":
				return "", nil
			case command == "kubectl" && len(args) > 1 && args[0] == "create":
				creates.Add(1)
				return "created", nil
			default:
				return "", nil
			}
		},
	}
	state := newTestState(runner)
	writeClusterTargetFixture(t, state, expectedClusterContext("local"), server)
	err := RunClusterNamespace(context.Background(), state, "safe")
	if err == nil || !strings.Contains(err.Error(), "target changed") {
		t.Fatalf("expected immediate re-verification failure, got %v", err)
	}
	if creates.Load() != 0 {
		t.Fatal("namespace mutation ran after the target changed")
	}
}

func TestClusterTargetRejectsLinkedOrBroadKubeconfig(t *testing.T) {
	state := newTestState(newClusterTargetRunner("local", nil))
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte(clusterTestKubeconfig("local", expectedClusterContext("local"), "https://127.0.0.1:6443")), 0o600); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(t.TempDir(), "linked.yaml")
	if err := os.Symlink(outside, linked); err != nil {
		t.Fatal(err)
	}
	state.Config.Cluster.KubeconfigPath = linked
	if _, err := resolveClusterTarget(context.Background(), state, ClusterTargetOptions{}, false); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("expected linked kubeconfig rejection, got %v", err)
	}

	broad := writeClusterTargetFixture(t, state, expectedClusterContext("local"), "https://127.0.0.1:6443")
	if err := os.Chmod(broad, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveClusterTarget(context.Background(), state, ClusterTargetOptions{}, false); err == nil || !strings.Contains(err.Error(), "expected owner-only") {
		t.Fatalf("expected broad-permission rejection, got %v", err)
	}
}
