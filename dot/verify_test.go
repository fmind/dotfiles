package dot

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestDefaultVerifyConfigIncludesManagedAgentCLIs(t *testing.T) {
	tools := make(map[string]bool)
	for _, tool := range defaultVerifyConfig().Tools {
		tools[tool] = true
	}
	for _, tool := range []string{"claude", "codex", "copilot"} {
		if !tools[tool] {
			t.Errorf("default verification tools omit managed agent CLI %q", tool)
		}
	}
}

func TestEnvVarsChecker(t *testing.T) {
	checker := &EnvVarsChecker{}
	state := newTestState(&FakeRunner{})

	// Pre-populate environment
	origRequired := state.Config.Verify.EnvVars.Required
	origOptional := state.Config.Verify.EnvVars.Optional

	state.Config.Verify.EnvVars.Required = []string{"TEST_REQ_VAR"}
	state.Config.Verify.EnvVars.Optional = []string{"TEST_OPT_VAR"}

	// 1. Both missing (empty is treated as unset by the checker)
	t.Setenv("TEST_REQ_VAR", "")
	t.Setenv("TEST_OPT_VAR", "")

	res, passed := checker.Check(context.Background(), state, false)
	if passed {
		t.Error("Expected Check to fail when required env var is missing")
	}

	foundReq := false
	foundOpt := false
	for _, r := range res {
		if r.Name == "TEST_REQ_VAR" {
			foundReq = true
			if r.Status != statusFail {
				t.Errorf("Expected TEST_REQ_VAR status to be fail, got %s", r.Status)
			}
		}
		if r.Name == "TEST_OPT_VAR" {
			foundOpt = true
			if r.Status != statusWarn {
				t.Errorf("Expected TEST_OPT_VAR status to be warn, got %s", r.Status)
			}
		}
	}
	if !foundReq || !foundOpt {
		t.Error("Did not find expected env var results")
	}

	// 2. Both present
	t.Setenv("TEST_REQ_VAR", "value1")
	t.Setenv("TEST_OPT_VAR", "value2")

	res, passed = checker.Check(context.Background(), state, false)
	if !passed {
		t.Error("Expected Check to pass when required env var is present")
	}

	for _, r := range res {
		if r.Status != statusPass {
			t.Errorf("Expected status pass for %s, got %s", r.Name, r.Status)
		}
	}

	// Restore config
	state.Config.Verify.EnvVars.Required = origRequired
	state.Config.Verify.EnvVars.Optional = origOptional
}

func TestToolsChecker(t *testing.T) {
	checker := &ToolsChecker{}

	// Setup fake runner that fails to find 'missing-tool' and finds 'existing-tool'
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			if name == "missing-tool" {
				return "", errors.New("not found")
			}
			return "/bin/" + name, nil
		},
	}
	state := newTestState(runner)
	state.Config.Verify.Tools = []string{"existing-tool", "missing-tool"}

	res, passed := checker.Check(context.Background(), state, false)
	if passed {
		t.Error("Expected Check to fail when a tool is missing")
	}

	if len(res) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(res))
	}

	if res[0].Name != "existing-tool" || res[0].Status != statusPass {
		t.Errorf("Expected existing-tool to pass, got %+v", res[0])
	}

	if res[1].Name != "missing-tool" || res[1].Status != statusFail {
		t.Errorf("Expected missing-tool to fail, got %+v", res[1])
	}
}

func TestDockerChecker(t *testing.T) {
	checker := &DockerChecker{}

	t.Run("Docker missing", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "", errors.New("not installed")
			},
		}
		state := newTestState(runner)
		res, passed := checker.Check(context.Background(), state, false)
		if passed {
			t.Error("Expected Docker check to fail when Docker is not installed")
		}
		if len(res) == 0 || res[0].Status != statusFail {
			t.Errorf("Expected fail result, got %+v", res)
		}
	})

	t.Run("Docker installed but daemon stopped", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/docker", nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				if name == "docker" && args[0] == "info" {
					return "", errors.New("cannot connect to docker daemon")
				}
				return "", nil
			},
		}
		state := newTestState(runner)
		res, passed := checker.Check(context.Background(), state, false)
		if passed {
			t.Error("Expected Docker check to fail when daemon is stopped")
		}
		if len(res) == 0 || res[0].Status != statusFail {
			t.Errorf("Expected fail result, got %+v", res)
		}
	})

	t.Run("Docker running", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/usr/bin/docker", nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "Server Version: 24.0.0", nil
			},
		}
		state := newTestState(runner)
		res, passed := checker.Check(context.Background(), state, false)
		if !passed {
			t.Error("Expected Docker check to pass when daemon is running")
		}
		if len(res) == 0 || res[0].Status != statusPass {
			t.Errorf("Expected pass result, got %+v", res)
		}
	})
}

func TestSecretsChecker(t *testing.T) {
	checker := &SecretsChecker{}

	tempDir, err := os.MkdirTemp("", "secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	secretFile := filepath.Join(tempDir, "key.txt")
	err = os.WriteFile(secretFile, []byte("super-secret"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	state := newTestState(&FakeRunner{})
	state.Config.Verify.Secrets = []SecretConfig{
		{
			Path:         secretFile,
			RequiredPerm: 0o600,
		},
	}

	t.Run("Correct permissions", func(t *testing.T) {
		res, passed := checker.Check(context.Background(), state, false)
		if !passed {
			t.Error("Expected secrets check to pass when permissions are correct")
		}
		if len(res) == 0 || res[0].Status != statusPass {
			t.Errorf("Expected pass status, got %+v", res)
		}
	})

	t.Run("Incorrect permissions without fix", func(t *testing.T) {
		err = os.Chmod(secretFile, 0o644)
		if err != nil {
			t.Fatalf("Failed to chmod: %v", err)
		}

		res, passed := checker.Check(context.Background(), state, false)
		if passed {
			t.Error("Expected secrets check to fail when permissions are 0644")
		}
		if len(res) == 0 || res[0].Status != statusFail {
			t.Errorf("Expected fail status, got %+v", res)
		}
	})

	t.Run("Incorrect permissions with fix", func(t *testing.T) {
		res, passed := checker.Check(context.Background(), state, true)
		if !passed {
			t.Error("Expected secrets check to pass when permissions are repaired")
		}
		if len(res) == 0 || res[0].Status != statusPass {
			t.Errorf("Expected pass (repaired) status, got %+v", res)
		}

		// Double check file permissions actually changed back to 0600
		info, err := os.Stat(secretFile)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("Expected permissions to be 0600, got %o", info.Mode().Perm())
		}
	})

	t.Run("Missing file", func(t *testing.T) {
		state.Config.Verify.Secrets[0].Path = filepath.Join(tempDir, "missing.txt")
		res, passed := checker.Check(context.Background(), state, false)
		if !passed {
			t.Error("Missing secret files should be treated as warnings, not fail the whole check")
		}
		if len(res) == 0 || res[0].Status != statusWarn {
			t.Errorf("Expected warn status for missing file, got %+v", res)
		}
	})
}

// TestSecretsCheckerPresenceOnly guards the fix for a config entry that omits
// required_perms (RequiredPerm == 0): the check must pass on presence alone and
// must never chmod the file (a naive exact-match + --fix would set it to 0000).
func TestSecretsCheckerPresenceOnly(t *testing.T) {
	checker := &SecretsChecker{}

	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "key.txt")
	if err := os.WriteFile(secretFile, []byte("super-secret"), 0o600); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	state := newTestState(&FakeRunner{})
	state.Config.Verify.Secrets = []SecretConfig{{Path: secretFile}} // RequiredPerm defaults to 0

	// Even with --fix enabled, a presence-only entry must pass and stay untouched.
	res, passed := checker.Check(context.Background(), state, true)
	if !passed {
		t.Error("Expected presence-only secret (no required_perms) to pass")
	}
	if len(res) == 0 || res[0].Status != statusPass {
		t.Errorf("Expected pass status, got %+v", res)
	}
	info, err := os.Stat(secretFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("Expected permissions untouched at 0600, got %o", info.Mode().Perm())
	}
}

func TestAuthChecker(t *testing.T) {
	checker := &AuthChecker{}

	t.Run("Auth passes", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "authenticated", nil
			},
		}
		state := newTestState(runner)
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv(EnvJulesAPIKey, "dummy-key")

		// Create a clasp config inside the isolated home so the clasp check passes.
		_ = os.WriteFile(filepath.Join(home, ".clasprc.json"), []byte("{}"), 0o600)

		res, passed := checker.Check(context.Background(), state, false)
		// We expect it to pass because all lookups / runs succeed
		if !passed {
			for _, r := range res {
				if r.Status == statusFail {
					t.Errorf("Failed check: %s - %s", r.Name, r.Details)
				}
			}
			t.Error("Expected Auth check to pass")
		}
	})

	t.Run("missing JULES key skips without data race", func(t *testing.T) {
		runner := &FakeRunner{
			LookPathFunc: func(name string) (string, error) {
				return "/bin/" + name, nil
			},
			RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
				return "authenticated", nil
			},
		}
		state := newTestState(runner)
		t.Setenv("HOME", t.TempDir())
		t.Setenv(EnvJulesAPIKey, "")

		// Exercises the JULES-unset append branch (previously raced under -race). A missing
		// key must not fail AuthChecker — EnvVarsChecker owns that failure — so jules is
		// surfaced here as a skip instead.
		res, passed := checker.Check(context.Background(), state, false)
		if !passed {
			t.Error("Expected Auth check to pass when only JULES_API_KEY is missing (env var checker owns that failure)")
		}

		var foundJulesSkip bool
		for _, r := range res {
			if r.Name == "jules" && r.Status == statusSkip {
				foundJulesSkip = true
			}
		}
		if !foundJulesSkip {
			t.Errorf("Expected a skipped jules result, got %+v", res)
		}
	})
}

func TestRunAllChecks(t *testing.T) {
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			return "ok", nil
		},
	}
	state := newTestState(runner)
	t.Setenv(EnvJulesAPIKey, "dummy")
	t.Setenv(EnvStitchAccessToken, "dummy")

	results := RunAllChecks(context.Background(), state, false)
	if len(results.Tools) == 0 {
		t.Error("Expected Tools check results to be populated")
	}
	if len(results.Docker) == 0 {
		t.Error("Expected Docker check results to be populated")
	}
}

func TestVerifyCommand(t *testing.T) {
	runner := &FakeRunner{
		LookPathFunc: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
			return "ok", nil
		},
	}
	state := newTestState(runner)
	var buf strings.Builder
	state.Stdout = &buf

	t.Setenv(EnvJulesAPIKey, "dummy")
	t.Setenv(EnvStitchAccessToken, "dummy")

	app := &cli.Command{
		Commands: []*cli.Command{
			NewVerifyCmd(state),
		},
	}

	err := app.Run(context.Background(), []string{"dot", "verify", "--json"})
	if err != nil {
		t.Fatalf("Expected no error running dot verify --json, got %v", err)
	}

	output := buf.String()

	// Since secrets/env vars might not all pass in testing if config is complex,
	// let's just make sure it prints JSON.
	if !strings.Contains(output, "{") {
		t.Errorf("Expected JSON output, got %q", output)
	}
}

func TestCheckerNames(t *testing.T) {
	checkers := []Checker{
		&EnvVarsChecker{},
		&AuthChecker{},
		&SecretsChecker{},
		&DockerChecker{},
		&ToolsChecker{},
	}
	for _, c := range checkers {
		if c.Name() == "" {
			t.Error("Expected checker name to be non-empty")
		}
	}
}

func TestPrintHumanResults(t *testing.T) {
	results := &VerifyResults{
		Passed: true,
		EnvVars: []CheckResult{
			{Name: "VAR1", Status: statusPass, Details: "set"},
		},
		Auth: []CheckResult{
			{Name: "auth1", Status: statusFail, Details: "failed"},
		},
		Secrets: []CheckResult{
			{Name: "secret1", Status: statusWarn, Details: "warn"},
		},
		Docker: []CheckResult{
			{Name: "docker", Status: statusSkip, Details: "skipped"},
		},
		Tools: []CheckResult{
			{Name: "tool1", Status: statusPass, Details: "ok"},
		},
	}

	var buf strings.Builder
	PrintHumanResults(&buf, results)
	output := buf.String()

	if !strings.Contains(output, "VAR1") || !strings.Contains(output, "auth1") {
		t.Errorf("Expected printed results to contain names, got %q", output)
	}
}

func TestVerifyCmd_Errors(t *testing.T) {
	ctx := context.Background()

	origExiter := cli.OsExiter
	cli.OsExiter = func(code int) {}
	defer func() { cli.OsExiter = origExiter }()

	t.Run("JSON verification fails", func(t *testing.T) {
		// Mock a failing checker config
		state := newTestState(&FakeRunner{})
		// A required env var is missing to force failure
		state.Config.Verify.EnvVars.Required = []string{"NON_EXISTENT_VAR_12345"}
		t.Setenv("NON_EXISTENT_VAR_12345", "")

		app := &cli.Command{
			Commands: []*cli.Command{
				NewVerifyCmd(state),
			},
		}

		err := app.Run(ctx, []string{"dot", "verify", "--json"})
		if err == nil {
			t.Fatal("Expected CLI exit error, got nil")
		}
		if !strings.Contains(err.Error(), "Verification failed") {
			t.Errorf("Expected 'Verification failed' error message, got %v", err)
		}
	})

	t.Run("non-JSON verification fails", func(t *testing.T) {
		state := newTestState(&FakeRunner{})
		state.Config.Verify.EnvVars.Required = []string{"NON_EXISTENT_VAR_12345"}
		t.Setenv("NON_EXISTENT_VAR_12345", "")

		app := &cli.Command{
			Commands: []*cli.Command{
				NewVerifyCmd(state),
			},
		}

		err := app.Run(ctx, []string{"dot", "verify"})
		if err == nil {
			t.Fatal("Expected CLI exit error, got nil")
		}
		// Without JSON, it exits with empty message, but exits with code 1
		// cli.Exit("", 1) has message ""
		if err.Error() != "" {
			t.Errorf("Expected empty error message, got %v", err)
		}
	})
}
