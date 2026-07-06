package dot

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunSetupWorkspace_Failures(t *testing.T) {
	ctx := context.Background()

	t.Run("project ID missing completely", func(t *testing.T) {
		t.Setenv("GWS_PROJECT", "")

		runner := &FakeRunner{}
		state := newTestState(runner)
		err := RunSetupWorkspace(ctx, state, "")
		if err == nil || !strings.Contains(err.Error(), "provide a project ID as an argument") {
			t.Errorf("Expected project ID missing error, got %v", err)
		}
	})

	t.Run("project ID retrieved from GWS_PROJECT", func(t *testing.T) {
		t.Setenv("GWS_PROJECT", "env-project-123")

		var gwsSetupProject string
		runner := &FakeRunner{
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gws" && args[0] == "auth" && args[1] == "setup" {
					for i, arg := range args {
						if arg == "--project" && i+1 < len(args) {
							gwsSetupProject = args[i+1]
						}
					}
				}
				return nil
			},
		}
		state := newTestState(runner)
		err := RunSetupWorkspace(ctx, state, "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if gwsSetupProject != "env-project-123" {
			t.Errorf("Expected project ID 'env-project-123', got %q", gwsSetupProject)
		}
	})

	t.Run("gws CLI missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "gws" {
					return "", errors.New("gws not found")
				}
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		err := RunSetupWorkspace(ctx, state, "proj-1")
		if !errors.Is(err, ErrGwsNotInstalled) {
			t.Errorf("Expected ErrGwsNotInstalled, got %v", err)
		}
	})

	t.Run("gcloud CLI missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "gcloud" {
					return "", errors.New("gcloud not found")
				}
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		err := RunSetupWorkspace(ctx, state, "proj-1")
		if !errors.Is(err, ErrGcloudNotInstalled) {
			t.Errorf("Expected ErrGcloudNotInstalled, got %v", err)
		}
	})

	t.Run("no Workspace APIs configured", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
		}
		state := newTestState(runner)
		state.Config.Setup.WorkspaceAPIs = []string{}
		err := RunSetupWorkspace(ctx, state, "proj-1")
		if err == nil || !strings.Contains(err.Error(), "no Google Workspace APIs configured to enable") {
			t.Errorf("Expected no APIs configured error, got %v", err)
		}
	})

	t.Run("gcloud services enable fails", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/" + name, nil
			},
			RunInteractiveFunc: func(ctx context.Context, dir, name string, args ...string) error {
				if name == "gcloud" && args[0] == "services" && args[1] == "enable" {
					return errors.New("gcloud error")
				}
				return nil
			},
		}
		state := newTestState(runner)
		err := RunSetupWorkspace(ctx, state, "proj-1")
		if err == nil || !strings.Contains(err.Error(), "failed to enable gcloud services") {
			t.Errorf("Expected gcloud services enable failure, got %v", err)
		}
	})
}
