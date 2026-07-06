package dot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLoginWorkspace_GwsMissing(t *testing.T) {
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			if name == "gws" {
				return "", errors.New("gws not found")
			}
			return "/usr/bin/" + name, nil
		},
	}
	state := newTestState(runner)
	err := RunLoginWorkspace(context.Background(), state)
	if !errors.Is(err, ErrGwsNotInstalled) {
		t.Errorf("Expected ErrGwsNotInstalled, got %v", err)
	}
}

func TestRunLoginGcp_Failures(t *testing.T) {
	ctx := context.Background()

	t.Run("gcloud login fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gcloud" && args[0] == "auth" && args[1] == "login" {
					return errors.New("login failed")
				}
				return nil
			},
		}
		state := newTestState(runner)
		err := RunLoginGcp(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "gcloud user login failed") {
			t.Errorf("Expected user login failure error, got %v", err)
		}
	})

	t.Run("gcloud ADC login fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gcloud" && args[0] == "auth" {
					if args[1] == "login" {
						return nil
					}
					if args[1] == "application-default" && args[2] == "login" {
						return errors.New("ADC failed")
					}
				}
				return nil
			},
		}
		state := newTestState(runner)
		err := RunLoginGcp(ctx, state)
		if err == nil || !strings.Contains(err.Error(), "gcloud application default login failed") {
			t.Errorf("Expected application default login failure error, got %v", err)
		}
	})
}

func TestRunLoginClasp(t *testing.T) {
	ctx := context.Background()

	t.Run("clasp missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "clasp" {
					return "", errors.New("clasp not found")
				}
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		err := RunLoginClasp(ctx, state)
		if !errors.Is(err, ErrClaspNotInstalled) {
			t.Errorf("Expected ErrClaspNotInstalled, got %v", err)
		}
	})

	t.Run("clasp installed - not authenticated", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)

		var runCalled bool
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "clasp" && len(args) == 1 && args[0] == "login" {
					runCalled = true
					return nil
				}
				return nil
			},
		}
		state := newTestState(runner)
		err := RunLoginClasp(ctx, state)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !runCalled {
			t.Error("Expected clasp login to be called")
		}
	})

	t.Run("clasp installed - already authenticated - cancel", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)

		claspJSON := filepath.Join(tmp, ".clasprc.json")
		if err := os.WriteFile(claspJSON, []byte("{}"), 0o600); err != nil {
			t.Fatalf("failed to write mock .clasprc.json: %v", err)
		}

		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		state.Stdin = strings.NewReader("n\n")

		err := RunLoginClasp(ctx, state)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("clasp installed - already authenticated - accept", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)

		claspJSON := filepath.Join(tmp, ".clasprc.json")
		if err := os.WriteFile(claspJSON, []byte("{}"), 0o600); err != nil {
			t.Fatalf("failed to write mock .clasprc.json: %v", err)
		}

		var runCalled bool
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "clasp" && len(args) == 1 && args[0] == "login" {
					runCalled = true
					return nil
				}
				return nil
			},
		}
		state := newTestState(runner)
		state.Stdin = strings.NewReader("y\n")

		err := RunLoginClasp(ctx, state)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !runCalled {
			t.Error("Expected clasp login to be called")
		}
	})
}

func TestUrlOpenerWriter(t *testing.T) {
	t.Run("detects https URL", func(t *testing.T) {
		var openedURL string
		var openedCount int
		browser := &FakeBrowser{
			OpenFunc: func(url string) error {
				openedURL = url
				openedCount++
				return nil
			},
		}
		opener := &urlOpener{browser: browser}
		var out strings.Builder
		w := &urlOpenerWriter{w: &out, opener: opener}

		_, err := w.Write([]byte("Open this URL: https://example.com/auth?code=123 \n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if openedCount != 1 {
			t.Errorf("expected openBrowser to be called 1 time, got %d", openedCount)
		}
		if openedURL != "https://example.com/auth?code=123" {
			t.Errorf("expected URL 'https://example.com/auth?code=123', got '%s'", openedURL)
		}
	})

	t.Run("detects URL split across writes", func(t *testing.T) {
		var openedURL string
		var openedCount int
		browser := &FakeBrowser{
			OpenFunc: func(url string) error {
				openedURL = url
				openedCount++
				return nil
			},
		}
		opener := &urlOpener{browser: browser}
		var out strings.Builder
		w := &urlOpenerWriter{w: &out, opener: opener}

		_, _ = w.Write([]byte("URL: https://example"))
		if openedCount != 0 {
			t.Errorf("unexpected call to openBrowser")
		}
		_, _ = w.Write([]byte(".com/login\n"))
		if openedCount != 1 {
			t.Errorf("expected openBrowser to be called 1 time, got %d", openedCount)
		}
		if openedURL != "https://example.com/login" {
			t.Errorf("expected URL 'https://example.com/login', got '%s'", openedURL)
		}
	})
}
