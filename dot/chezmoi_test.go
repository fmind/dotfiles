package dot

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestRunChezmoiClean(t *testing.T) {
	ctx := context.Background()

	setupEnv := func(t *testing.T) (string, string, *GlobalState, *bytes.Buffer) {
		tempDir := t.TempDir()
		homeDir := filepath.Join(tempDir, "home")
		sourceDir := filepath.Join(tempDir, "source")

		_ = os.MkdirAll(homeDir, 0o755)
		_ = os.MkdirAll(sourceDir, 0o755)

		t.Setenv("HOME", homeDir)

		// Create the orphan file in home directory
		orphanPath := filepath.Join(homeDir, ".orphaned_file")
		_ = os.WriteFile(orphanPath, []byte("orphan content"), 0o644)

		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "chezmoi" {
					if args[0] == "source-path" {
						return sourceDir, nil
					}
					if args[0] == "managed" {
						return ".other_managed_file\n", nil
					}
					if args[0] == "target-path" {
						return ".orphaned_file", nil
					}
				}
				if name == "git" {
					if args[0] == "log" {
						return ".orphaned_file\n", nil
					}
					if args[0] == "diff" {
						return "", nil
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		var stdout bytes.Buffer
		state.Stdout = &stdout

		return homeDir, orphanPath, state, &stdout
	}

	t.Run("default non-interactive mode only prints candidates and does not delete", func(t *testing.T) {
		homeDir, orphanPath, state, stdout := setupEnv(t)

		err := RunChezmoiClean(ctx, state, false, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify that the orphan file was NOT deleted
		if _, err := os.Stat(orphanPath); os.IsNotExist(err) {
			t.Error("Expected orphaned file to still exist, but it was deleted")
		}

		// Verify that candidates were printed
		out := stdout.String()
		if !strings.Contains(out, "Detected 1 orphaned file(s)") {
			t.Errorf("Expected 'Detected 1 orphaned file(s)' in stdout, got %q", out)
		}
		if !strings.Contains(out, filepath.Join(homeDir, ".orphaned_file")) {
			t.Errorf("Expected orphaned file path in stdout, got %q", out)
		}
	})

	t.Run("yes mode backs up orphans (moved, not deleted) without prompting", func(t *testing.T) {
		homeDir, orphanPath, state, stdout := setupEnv(t)

		err := RunChezmoiClean(ctx, state, true, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// The original path must be gone...
		if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
			t.Error("Expected orphaned file to be moved out of its original path, but it still exists")
		}

		// ...but recoverable: backupOrphans MOVES the file into ~/.cache/dot/chezmoi-clean
		// rather than deleting it. Pin that contract so a future refactor to os.Remove cannot
		// silently turn this into a destructive delete while the suite stays green.
		matches, _ := filepath.Glob(filepath.Join(homeDir, ".cache", "dot", "chezmoi-clean", "*", ".orphaned_file"))
		if len(matches) == 0 {
			t.Error("Expected the orphan to be backed up under ~/.cache/dot/chezmoi-clean, but no backup was found")
		} else if b, _ := os.ReadFile(matches[0]); string(b) != "orphan content" {
			t.Errorf("Expected the backup to preserve the original contents, got %q", string(b))
		}

		// Verify cleanup message
		out := stdout.String()
		if !strings.Contains(out, "Clean up complete.") {
			t.Errorf("Expected 'Clean up complete.' in stdout, got %q", out)
		}
	})

	t.Run("interactive mode with user saying yes deletes orphans", func(t *testing.T) {
		_, orphanPath, state, stdout := setupEnv(t)

		// Set stdin to say yes
		state.Stdin = strings.NewReader("y\n")

		err := RunChezmoiClean(ctx, state, false, true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify that the orphan file was deleted
		if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
			t.Error("Expected orphaned file to be deleted, but it still exists")
		}

		// Verify interactive prompt and cleanup message
		out := stdout.String()
		if !strings.Contains(out, "Do you want to move all orphaned files to a backup directory? [y/N]:") {
			t.Errorf("Expected confirmation prompt in stdout, got %q", out)
		}
		if !strings.Contains(out, "Clean up complete.") {
			t.Errorf("Expected 'Clean up complete.' in stdout, got %q", out)
		}
	})

	t.Run("interactive mode with user saying no cancels cleanup", func(t *testing.T) {
		_, orphanPath, state, stdout := setupEnv(t)

		// Set stdin to say no
		state.Stdin = strings.NewReader("n\n")

		err := RunChezmoiClean(ctx, state, false, true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify that the orphan file was NOT deleted
		if _, err := os.Stat(orphanPath); os.IsNotExist(err) {
			t.Error("Expected orphaned file to still exist, but it was deleted")
		}

		// Verify cancel message
		out := stdout.String()
		if !strings.Contains(out, "Clean up canceled. No files were modified.") {
			t.Errorf("Expected 'Clean up canceled. No files were modified.' in stdout, got %q", out)
		}
	})
}

func TestChezmoiCommand(t *testing.T) {
	ctx := context.Background()

	setupApp := func(t *testing.T) (*cli.Command, *bytes.Buffer) {
		tempDir := t.TempDir()
		homeDir := filepath.Join(tempDir, "home")
		sourceDir := filepath.Join(tempDir, "source")

		_ = os.MkdirAll(homeDir, 0o755)
		_ = os.MkdirAll(sourceDir, 0o755)

		t.Setenv("HOME", homeDir)

		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "chezmoi" {
					if args[0] == "source-path" {
						return sourceDir, nil
					}
					if args[0] == "managed" {
						return "", nil
					}
				}
				if name == "git" {
					if args[0] == "log" {
						return "", nil
					}
					if args[0] == "diff" {
						return "", nil
					}
				}
				return "", nil
			},
		}

		state := newTestState(runner)
		var stdout bytes.Buffer
		state.Stdout = &stdout

		app := &cli.Command{
			Commands: []*cli.Command{
				NewChezmoiCmd(state),
			},
		}
		return app, &stdout
	}

	t.Run("dot chezmoi clean works", func(t *testing.T) {
		app, stdout := setupApp(t)
		err := app.Run(ctx, []string{"dot", "chezmoi", "clean"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !strings.Contains(stdout.String(), "No orphaned files found") {
			t.Errorf("Expected no orphaned files message, got: %q", stdout.String())
		}
	})

	t.Run("dot m clean works", func(t *testing.T) {
		app, stdout := setupApp(t)
		err := app.Run(ctx, []string{"dot", "m", "clean"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !strings.Contains(stdout.String(), "No orphaned files found") {
			t.Errorf("Expected no orphaned files message, got: %q", stdout.String())
		}
	})

	t.Run("dot m c works", func(t *testing.T) {
		app, stdout := setupApp(t)
		err := app.Run(ctx, []string{"dot", "m", "c"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !strings.Contains(stdout.String(), "No orphaned files found") {
			t.Errorf("Expected no orphaned files message, got: %q", stdout.String())
		}
	})
}
