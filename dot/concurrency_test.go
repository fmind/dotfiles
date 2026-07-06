package dot

import (
	"context"
	"io"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestVerifyConcurrencyPanicAndTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				if name == "docker" {
					panic("simulated docker checker panic")
				}
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				// Simulate a slow command for gcloud print-access-token to trigger timeout
				if name == "gcloud" && len(args) > 1 && args[1] == "print-access-token" {
					select {
					case <-time.After(40 * time.Second):
						return "token", nil
					case <-ctx.Done():
						return "", ctx.Err()
					}
				}
				return "ok", nil
			},
		}

		state := newTestState(runner)
		t.Setenv(EnvJulesAPIKey, "dummy")
		t.Setenv(EnvStitchAccessToken, "dummy")

		// Run sanity checks
		results := RunAllChecks(context.Background(), state, false)

		// 1. Check that Docker checker failed due to panic recovery
		foundDockerPanic := false
		for _, r := range results.Docker {
			if r.Name == "Docker Service" && r.Status == statusFail && strings.Contains(r.Details, "PANIC") {
				foundDockerPanic = true
			}
		}
		if !foundDockerPanic {
			t.Errorf("Expected Docker checker to fail with PANIC details, got: %+v", results.Docker)
		}

		// 2. Check that Auth checker failed due to timeout on gcloud
		foundAuthFail := false
		for _, r := range results.Auth {
			if r.Name == "gcloud" && r.Status == statusFail {
				foundAuthFail = true
			}
		}
		if !foundAuthFail {
			t.Errorf("Expected gcloud check to fail due to timeout, got: %+v", results.Auth)
		}

		// The whole verification should be marked as failed
		if results.Passed {
			t.Error("Expected overall verification to fail")
		}
	})
}
