package dot

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chezmoiRunner builds a FakeRunner whose chezmoi/git calls all succeed, so a test
// only has to describe the single command it wants to fail.
func chezmoiRunner(sourceDir string, override func(name string, args []string) (string, error, bool)) *FakeRunner {
	return &FakeRunner{
		RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, args ...string) (string, error) {
			if override != nil {
				if out, err, handled := override(name, args); handled {
					return out, err
				}
			}
			if name == "chezmoi" && args[0] == "source-path" {
				return sourceDir, nil
			}
			return "", nil
		},
	}
}

func TestRunChezmoiCleanPrerequisiteFailures(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()

	tests := []struct {
		name     string
		lookPath func(name string) (string, error)
		override func(name string, args []string) (string, error, bool)
		wantErr  error
		wantMsg  string
	}{
		{
			name:     "git is missing",
			lookPath: func(name string) (string, error) { return "", errors.New("not found") },
			wantErr:  ErrGitNotInstalled,
		},
		{
			name: "chezmoi is missing",
			lookPath: func(name string) (string, error) {
				if name == "chezmoi" {
					return "", errors.New("not found")
				}
				return "/usr/bin/" + name, nil
			},
			wantErr: ErrChezmoiNotInstalled,
		},
		{
			name: "source-path fails",
			override: func(name string, args []string) (string, error, bool) {
				if name == "chezmoi" && args[0] == "source-path" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to get chezmoi source path",
		},
		{
			name: "source-path is empty",
			override: func(name string, args []string) (string, error, bool) {
				if name == "chezmoi" && args[0] == "source-path" {
					return "  \n", nil, true
				}
				return "", nil, false
			},
			wantMsg: "chezmoi source path is empty",
		},
		{
			name: "managed fails",
			override: func(name string, args []string) (string, error, bool) {
				if name == "chezmoi" && args[0] == "managed" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to run chezmoi managed",
		},
		{
			name: "git log fails",
			override: func(name string, args []string) (string, error, bool) {
				if name == "git" && args[0] == "log" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to run git log",
		},
		{
			name: "git diff fails",
			override: func(name string, args []string) (string, error, bool) {
				if name == "git" && args[0] == "diff" && args[1] != "--cached" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to run git diff",
		},
		{
			name: "git diff --cached fails",
			override: func(name string, args []string) (string, error, bool) {
				if name == "git" && args[0] == "diff" && args[1] == "--cached" {
					return "", errors.New("boom"), true
				}
				return "", nil, false
			},
			wantMsg: "failed to run git diff --cached",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			runner := chezmoiRunner(sourceDir, tc.override)
			runner.LookPathFunc = tc.lookPath
			state := newTestState(runner)

			err := RunChezmoiClean(ctx, state, false, false)
			switch {
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
			case err == nil || !strings.Contains(err.Error(), tc.wantMsg):
				t.Fatalf("expected an error containing %q, got %v", tc.wantMsg, err)
			}
		})
	}
}

func TestRunChezmoiCleanUnresolvableHome(t *testing.T) {
	t.Setenv("HOME", "")
	state := newTestState(chezmoiRunner(t.TempDir(), nil))

	err := RunChezmoiClean(context.Background(), state, false, false)
	if err == nil || !strings.Contains(err.Error(), "failed to get user home directory") {
		t.Fatalf("expected a home directory error, got %v", err)
	}
}

// TestRunChezmoiCleanSkipsNonOrphans covers the three skip guards in the candidate
// loop: ignored paths, unmappable source paths, and targets absent from $HOME.
func TestRunChezmoiCleanSkipsNonOrphans(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sourceDir := t.TempDir()

	runner := chezmoiRunner(sourceDir, func(name string, args []string) (string, error, bool) {
		switch {
		case name == "git" && args[0] == "log":
			// .git/config is ignored by prefix; the other two reach the mapping step.
			return ".git/config\ndot_unmappable\ndot_absent\n", nil, true
		case name == "chezmoi" && args[0] == "target-path":
			if strings.HasSuffix(args[1], "dot_unmappable") {
				return "", errors.New("cannot map"), true
			}
			return ".absent_from_home", nil, true
		}
		return "", nil, false
	})
	state := newTestState(runner)
	var stdout bytes.Buffer
	state.Stdout = &stdout

	if err := RunChezmoiClean(context.Background(), state, false, false); err != nil {
		t.Fatalf("RunChezmoiClean: %v", err)
	}
	if !strings.Contains(stdout.String(), "No orphaned files found") {
		t.Fatalf("expected every candidate to be skipped, got:\n%s", stdout.String())
	}
}

func TestBackupOrphans(t *testing.T) {
	t.Run("an unresolvable home surfaces the error", func(t *testing.T) {
		t.Setenv("HOME", "")
		state := newTestState(&FakeRunner{})

		err := backupOrphans(state, []string{"/tmp/orphan"})
		if err == nil || !strings.Contains(err.Error(), "failed to resolve home directory") {
			t.Fatalf("expected a home directory error, got %v", err)
		}
	})

	t.Run("orphans outside home keep their full path", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		outside := filepath.Join(t.TempDir(), "outside.conf")
		if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		state := newTestState(&FakeRunner{})
		var stdout bytes.Buffer
		state.Stdout = &stdout

		if err := backupOrphans(state, []string{outside}); err != nil {
			t.Fatalf("backupOrphans: %v", err)
		}
		if _, err := os.Stat(outside); !os.IsNotExist(err) {
			t.Errorf("expected the orphan to be moved out of the way, stat err = %v", err)
		}
		// The backup mirrors the absolute path so same-named orphans cannot collide.
		matches, err := filepath.Glob(filepath.Join(home, ".cache", "dot", "chezmoi-clean", "*", strings.TrimPrefix(outside, "/")))
		if err != nil || len(matches) != 1 {
			t.Fatalf("expected exactly one backup mirroring %s, got %v (%v)", outside, matches, err)
		}
	})

	t.Run("failures are reported instead of a false success", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		state := newTestState(&FakeRunner{})
		var stderr bytes.Buffer
		state.Stderr = &stderr

		// The orphan does not exist, so the rename fails and must be surfaced.
		err := backupOrphans(state, []string{filepath.Join(home, ".missing")})
		if err == nil || !strings.Contains(err.Error(), "failed to back up 1 of 1") {
			t.Fatalf("expected a backup failure error, got %v", err)
		}
		if !strings.Contains(stderr.String(), "Error backing up") {
			t.Errorf("expected the failure to be logged, got %q", stderr.String())
		}
	})
}

func TestResolveHome(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{name: "absolute paths are only cleaned", in: "/etc/../etc/hosts", want: "/etc/hosts"},
		{name: "relative paths resolve against home", in: ".config/dot.yaml", want: "/home/u/.config/dot.yaml"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveHome("/home/u", tc.in); got != tc.want {
				t.Errorf("resolveHome(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	cfg := ChezmoiCleanConfig{
		IgnoredPrefixes: []string{".git", ".github"},
		IgnoredFiles:    []string{"README.md"},
	}

	tests := []struct {
		path string
		want bool
	}{
		{path: ".git/config", want: true},
		{path: ".github/workflows/ci.yml", want: true},
		{path: "docs/README.md", want: true},
		{path: "run_once_install.sh", want: true},
		{path: "run_onchange_setup.sh", want: true},
		{path: "run_before_bootstrap.sh", want: true},
		{path: "dot_gitconfig", want: false},
	}
	for _, tc := range tests {
		if got := shouldIgnore(cfg, tc.path); got != tc.want {
			t.Errorf("shouldIgnore(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestDefaultChezmoiCleanConfigIgnoresRepositoryOnlyFiles(t *testing.T) {
	config := defaultChezmoiCleanConfig()
	for _, path := range []string{"go.work.sum", "verify-lazy-lock.sh"} {
		if !shouldIgnore(config, path) {
			t.Errorf("default clean config must ignore repository-only file %q", path)
		}
	}
}
