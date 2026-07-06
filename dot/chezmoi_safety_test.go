package dot

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chezmoiCleanEnv wires a fake chezmoi/git environment for RunChezmoiClean where a
// single relPath was deleted from the source history and still exists in home.
func chezmoiCleanEnv(t *testing.T, relPath, managed, targetPath string) (home, source string, state *GlobalState, out *bytes.Buffer) {
	t.Helper()
	tempDir := t.TempDir()
	home = filepath.Join(tempDir, "home")
	source = filepath.Join(tempDir, "source")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	t.Setenv("HOME", home)

	// The candidate exists in home so, absent a guard, it would be flagged as an orphan.
	if err := os.WriteFile(filepath.Join(home, relPath), []byte("content"), 0o644); err != nil {
		t.Fatalf("write home file: %v", err)
	}

	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) { return "/usr/bin/" + name, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name == "chezmoi" {
				switch args[0] {
				case "source-path":
					return source, nil
				case "managed":
					return managed + "\n", nil
				case "target-path":
					return targetPath, nil
				}
			}
			if name == "git" && args[0] == "log" {
				return relPath + "\n", nil
			}
			return "", nil
		},
	}
	state = newTestState(runner)
	out = &bytes.Buffer{}
	state.Stdout = out
	return home, source, state, out
}

// A deleted source whose target is still managed by chezmoi must NOT be treated as an
// orphan, even with --yes: backing it up would remove a file chezmoi actively manages.
func TestRunChezmoiClean_SkipsStillManaged(t *testing.T) {
	home, _, state, out := chezmoiCleanEnv(t, ".managed_file", ".managed_file", ".managed_file")

	if err := RunChezmoiClean(context.Background(), state, true, false); err != nil {
		t.Fatalf("RunChezmoiClean: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".managed_file")); err != nil {
		t.Errorf("expected still-managed file to be left in place, stat err: %v", err)
	}
	if !strings.Contains(out.String(), "No orphaned files found") {
		t.Errorf("expected no orphans reported, got %q", out.String())
	}
}

// A file re-added to the source tree (its source path exists again) must be skipped
// even though it was recorded as deleted in history.
func TestRunChezmoiClean_SkipsReAddedSource(t *testing.T) {
	home, source, state, out := chezmoiCleanEnv(t, ".readded_file", ".other", ".readded_file")

	// Re-create the source file so the source-existence guard triggers.
	if err := os.WriteFile(filepath.Join(source, ".readded_file"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if err := RunChezmoiClean(context.Background(), state, true, false); err != nil {
		t.Fatalf("RunChezmoiClean: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".readded_file")); err != nil {
		t.Errorf("expected re-added file to be left in place, stat err: %v", err)
	}
	if !strings.Contains(out.String(), "No orphaned files found") {
		t.Errorf("expected no orphans reported, got %q", out.String())
	}
}

// uniqueBackupPath must never return a path that already exists, so two orphans that
// collapse to the same destination can't overwrite one another's backup.
func TestUniqueBackupPath(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")

	if got := uniqueBackupPath(p); got != p {
		t.Errorf("expected free path returned unchanged, got %q", got)
	}

	if err := os.WriteFile(p, []byte("1"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := uniqueBackupPath(p)
	if want := filepath.Join(dir, "a.1.txt"); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}

	if err := os.WriteFile(got, []byte("2"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got2, want := uniqueBackupPath(p), filepath.Join(dir, "a.2.txt"); got2 != want {
		t.Errorf("expected %q, got %q", want, got2)
	}
}

// getChezmoiTargetPath touches a probe file inside the real source repo to resolve a
// target path. It must create only the missing intermediates and remove exactly those,
// never a pre-existing directory or its contents.
func TestGetChezmoiTargetPath_NestedCleanup(t *testing.T) {
	source := t.TempDir()
	ctx := context.Background()
	runner := &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if name == "chezmoi" && args[0] == "target-path" {
				return "~/.config/newtool/deep/conf\n", nil
			}
			return "", nil
		},
	}
	state := newTestState(runner)

	t.Run("creates and removes missing intermediates", func(t *testing.T) {
		got, err := getChezmoiTargetPath(ctx, state, source, filepath.Join("dot_config", "newtool", "deep", "dot_conf"))
		if err != nil {
			t.Fatalf("getChezmoiTargetPath: %v", err)
		}
		if got != "~/.config/newtool/deep/conf" {
			t.Errorf("expected resolved target path, got %q", got)
		}
		// Every intermediate it created must be gone afterwards.
		if _, err := os.Stat(filepath.Join(source, "dot_config")); !os.IsNotExist(err) {
			t.Errorf("expected created intermediates removed, stat err: %v", err)
		}
	})

	t.Run("preserves pre-existing directories and files", func(t *testing.T) {
		existing := filepath.Join(source, "existing")
		if err := os.MkdirAll(existing, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		keep := filepath.Join(existing, "keep.txt")
		if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		if _, err := getChezmoiTargetPath(ctx, state, source, filepath.Join("existing", "deep", "conf")); err != nil {
			t.Fatalf("getChezmoiTargetPath: %v", err)
		}
		if _, err := os.Stat(keep); err != nil {
			t.Errorf("expected pre-existing file to survive, stat err: %v", err)
		}
		if _, err := os.Stat(filepath.Join(existing, "deep")); !os.IsNotExist(err) {
			t.Errorf("expected only the created intermediate to be removed, stat err: %v", err)
		}
	})
}
