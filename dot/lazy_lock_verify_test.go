package dot

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLazyLockVerifier(t *testing.T) {
	repo := repositoryRoot(t)
	verifier := filepath.Join(repo, "verify-lazy-lock.sh")

	t.Run("accepts locked checkout with generated dirt", func(t *testing.T) {
		lazyRoot, commit := lazyPluginFixture(t)
		if err := os.WriteFile(filepath.Join(lazyRoot, "example.nvim", "plugin.lua"), []byte("return { generated = true }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(lazyRoot, "example.nvim", ".generated"), []byte("cache\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		lock := writeLazyLock(t, commit)
		output, err := runLazyVerifier(t, verifier, lock, lazyRoot)
		if err != nil {
			t.Fatalf("verifier rejected usable dirty checkout: %v\n%s", err, output)
		}
	})

	t.Run("rejects wrong head", func(t *testing.T) {
		lazyRoot, _ := lazyPluginFixture(t)
		lock := writeLazyLock(t, strings.Repeat("0", 40))
		output, err := runLazyVerifier(t, verifier, lock, lazyRoot)
		if err == nil || !strings.Contains(output, "locked commit") {
			t.Fatalf("verifier should reject a checkout at the wrong commit: %v\n%s", err, output)
		}
	})

	t.Run("rejects missing repository", func(t *testing.T) {
		lazyRoot := t.TempDir()
		lock := writeLazyLock(t, strings.Repeat("0", 40))
		output, err := runLazyVerifier(t, verifier, lock, lazyRoot)
		if err == nil || !strings.Contains(output, "missing Git repository") {
			t.Fatalf("verifier should reject a missing checkout: %v\n%s", err, output)
		}
	})

	t.Run("rejects dot path components", func(t *testing.T) {
		for _, plugin := range []string{".", ".."} {
			t.Run(plugin, func(t *testing.T) {
				lock := writeLazyLockForPlugin(t, plugin, strings.Repeat("0", 40))
				output, err := runLazyVerifier(t, verifier, lock, t.TempDir())
				if err == nil || !strings.Contains(output, "invalid plugin name") {
					t.Fatalf("verifier should reject plugin path component %q: %v\n%s", plugin, err, output)
				}
			})
		}
	})

	t.Run("rejects empty worktree", func(t *testing.T) {
		lazyRoot, commit := lazyPluginFixture(t)
		for _, name := range []string{"plugin.lua", "README.md"} {
			if err := os.Remove(filepath.Join(lazyRoot, "example.nvim", name)); err != nil {
				t.Fatal(err)
			}
		}
		lock := writeLazyLock(t, commit)
		output, err := runLazyVerifier(t, verifier, lock, lazyRoot)
		if err == nil || !strings.Contains(output, "empty worktree") {
			t.Fatalf("verifier should reject an empty checkout: %v\n%s", err, output)
		}
	})

	t.Run("rejects partial worktree", func(t *testing.T) {
		lazyRoot, commit := lazyPluginFixture(t)
		if err := os.Remove(filepath.Join(lazyRoot, "example.nvim", "README.md")); err != nil {
			t.Fatal(err)
		}
		runLazyGit(t, filepath.Join(lazyRoot, "example.nvim"), "add", "--update")
		lock := writeLazyLock(t, commit)
		output, err := runLazyVerifier(t, verifier, lock, lazyRoot)
		if err == nil || !strings.Contains(output, "missing tracked path README.md") {
			t.Fatalf("verifier should reject a partial checkout: %v\n%s", err, output)
		}
	})
}

func lazyPluginFixture(t *testing.T) (string, string) {
	t.Helper()
	// A fixture must not inherit repository routing, signing, or hooks from the
	// developer running the suite. Command-scope config wins over any local values
	// a Git template may seed during init.
	isolateGitEnvironment(t)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_COUNT", "2")
	t.Setenv("GIT_CONFIG_KEY_0", "commit.gpgsign")
	t.Setenv("GIT_CONFIG_VALUE_0", "false")
	t.Setenv("GIT_CONFIG_KEY_1", "core.hooksPath")
	t.Setenv("GIT_CONFIG_VALUE_1", os.DevNull)

	lazyRoot := t.TempDir()
	plugin := filepath.Join(lazyRoot, "example.nvim")
	if err := os.Mkdir(plugin, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugin, "plugin.lua"), []byte("return {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugin, "README.md"), []byte("# Fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runLazyGit(t, plugin, "init", "--quiet")
	runLazyGit(t, plugin, "add", "plugin.lua", "README.md")
	runLazyGit(t, plugin, "-c", "user.name=Dot Test", "-c", "user.email=dot@example.invalid", "commit", "--quiet", "-m", "fixture")
	return lazyRoot, strings.TrimSpace(runLazyGit(t, plugin, "rev-parse", "HEAD"))
}

func writeLazyLock(t *testing.T, commit string) string {
	t.Helper()
	return writeLazyLockForPlugin(t, "example.nvim", commit)
}

func writeLazyLockForPlugin(t *testing.T, plugin, commit string) string {
	t.Helper()
	data, err := json.Marshal(map[string]map[string]string{
		plugin: {"branch": "main", "commit": commit},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "lazy-lock.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runLazyVerifier(t *testing.T, verifier, lock, lazyRoot string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", verifier, lock, lazyRoot)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func runLazyGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", append([]string{"-C", directory}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
