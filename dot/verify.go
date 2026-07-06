package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
)

type VerifyResults struct {
	EnvVars []CheckResult `json:"env_vars"`
	Auth    []CheckResult `json:"auth"`
	Secrets []CheckResult `json:"secrets"`
	Docker  []CheckResult `json:"docker"`
	Tools   []CheckResult `json:"tools"`
	Passed  bool          `json:"passed"`
}

type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details,omitzero"`
}

// NewVerifyCmd constructs the top-level verify command.
func NewVerifyCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "verify",
		Aliases: []string{"v"},
		Usage:   "Run sanity checks on environment, CLI tool installations, and secrets",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output results in structured JSON format",
			},
			&cli.BoolFlag{
				Name:    "fix",
				Aliases: []string{"f"},
				Usage:   "Attempt to fix fixable errors (e.g. key permissions)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			isJSON := cmd.Bool("json")
			shouldFix := cmd.Bool("fix")

			_, err := RunVerify(ctx, state, isJSON, shouldFix)
			if err != nil {
				if isJSON {
					return cli.Exit("Verification failed", 1)
				}
				return cli.Exit("", 1)
			}
			return nil
		},
	}
}

// RunVerify runs all checks, handles formatting, and returns an error if any required check fails.
func RunVerify(ctx context.Context, state *GlobalState, isJSON, shouldFix bool) (*VerifyResults, error) {
	state.Logger.Debug("Starting system verification checks", "shouldFix", shouldFix)
	results := RunAllChecks(ctx, state, shouldFix)

	if isJSON {
		bz, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return results, err
		}
		_, _ = fmt.Fprintln(state.Stdout, string(bz))
		if !results.Passed {
			return results, errors.New("verification failed")
		}
		return results, nil
	}

	PrintHumanResults(state.Stdout, results)

	if !results.Passed {
		_, _ = fmt.Fprintln(state.Stdout, "\n"+red("✗ Verification failed. Run with --fix to resolve auto-fixable issues."))
		return results, errors.New("verification failed")
	}
	_, _ = fmt.Fprintln(state.Stdout, "\n"+green("✓ Verification passed."))
	return results, nil
}

// Checker defines the interface for running a category of sanity checks.
type Checker interface {
	Name() string
	Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool)
}

// RunAllChecks runs all sanity check suites concurrently and returns their aggregated results.
func RunAllChecks(ctx context.Context, state *GlobalState, shouldFix bool) *VerifyResults {
	results := &VerifyResults{Passed: true}
	var mu sync.Mutex
	var wg sync.WaitGroup

	checkers := []struct {
		checker Checker
		assign  func(res []CheckResult)
	}{
		{
			checker: &EnvVarsChecker{},
			assign:  func(res []CheckResult) { results.EnvVars = res },
		},
		{
			checker: &AuthChecker{},
			assign:  func(res []CheckResult) { results.Auth = res },
		},
		{
			checker: &SecretsChecker{},
			assign:  func(res []CheckResult) { results.Secrets = res },
		},
		{
			checker: &DockerChecker{},
			assign:  func(res []CheckResult) { results.Docker = res },
		},
		{
			checker: &ToolsChecker{},
			assign:  func(res []CheckResult) { results.Tools = res },
		},
	}

	for _, item := range checkers {
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					state.Logger.Error("Sanity checker panicked", "checker", item.checker.Name(), "panic", r)
					mu.Lock()
					item.assign([]CheckResult{
						{
							Name:    item.checker.Name(),
							Status:  statusFail,
							Details: fmt.Sprintf("PANIC: %v", r),
						},
					})
					results.Passed = false
					mu.Unlock()
				}
			}()

			state.Logger.Debug("Running sanity checker", "checker", item.checker.Name())
			childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			res, passed := item.checker.Check(childCtx, state, shouldFix)
			mu.Lock()
			item.assign(res)
			if !passed {
				results.Passed = false
			}
			mu.Unlock()
		})
	}

	wg.Wait()
	return results
}

// EnvVarsChecker verifies required and optional environment variables.
type EnvVarsChecker struct{}

func (c *EnvVarsChecker) Name() string { return "Environment Variables" }

func (c *EnvVarsChecker) Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool) {
	var results []CheckResult
	passed := true

	for _, name := range state.Config.Verify.EnvVars.Required {
		val := os.Getenv(name)
		if val != "" {
			results = append(results, CheckResult{Name: name, Status: statusPass, Details: "set"})
		} else {
			results = append(results, CheckResult{Name: name, Status: statusFail, Details: "MISSING (required)"})
			passed = false
		}
	}

	for _, name := range state.Config.Verify.EnvVars.Optional {
		val := os.Getenv(name)
		if val != "" {
			results = append(results, CheckResult{Name: name, Status: statusPass, Details: "set"})
		} else {
			results = append(results, CheckResult{Name: name, Status: statusWarn, Details: "unset (optional)"})
		}
	}

	return results, passed
}

// AuthChecker verifies authentication status for external services and CLIs.
type AuthChecker struct{}

func (c *AuthChecker) Name() string { return "CLI Authentication" }

func (c *AuthChecker) Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool) {
	type authTask struct {
		label   string
		cmdName string
		args    []string
	}

	tasks := []authTask{
		{"gh", "gh", []string{"auth", "status"}},
		{"gcloud", "gcloud", []string{"auth", "print-access-token"}},
		{"gcloud-adc", "gcloud", []string{"auth", "application-default", "print-access-token"}},
		{"gws", "gws", []string{"auth", "status"}},
	}

	// jules authenticates via JULES_API_KEY, not an interactive login, so we only probe it
	// when the key is present. A missing key is reported once as a required env var by
	// EnvVarsChecker; probing (and failing) here too would double-report one root cause.
	julesConfigured := os.Getenv(EnvJulesAPIKey) != ""
	if julesConfigured {
		tasks = append(tasks, authTask{"jules", "jules", []string{"remote", "list", "--repo"}})
	}

	results := make([]CheckResult, len(tasks)+1) // +1 for clasp
	passed := true
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Go(func() {
			_, err := state.Runner.LookPath(task.cmdName)
			if err != nil {
				results[i] = CheckResult{Name: task.label, Status: statusSkip, Details: task.cmdName + " not installed"}
				return
			}

			_, err = state.Runner.Run(ctx, "", nil, task.cmdName, task.args...)
			if err == nil {
				results[i] = CheckResult{Name: task.label, Status: statusPass, Details: "authenticated"}
			} else {
				results[i] = CheckResult{Name: task.label, Status: statusFail, Details: "NOT authenticated"}
				mu.Lock()
				passed = false
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	// Run the local checks after the workers join so that all mutation of the shared
	// results slice happens on a single goroutine (the pre-Wait append previously
	// reallocated the slice while workers were still writing it: a data race).
	claspIdx := len(tasks)
	home, err := os.UserHomeDir()
	if err == nil {
		claspJSON := filepath.Join(home, ".clasprc.json")
		if _, err := os.Stat(claspJSON); err == nil {
			results[claspIdx] = CheckResult{Name: "clasp", Status: statusPass, Details: "authenticated"}
		} else {
			results[claspIdx] = CheckResult{Name: "clasp", Status: statusSkip, Details: ".clasprc.json not found at " + claspJSON}
		}
	} else {
		results[claspIdx] = CheckResult{Name: "clasp", Status: statusSkip, Details: "could not resolve home directory"}
	}

	if !julesConfigured {
		// Surface jules in the auth section as a skip (not a second failure): the missing
		// key is already the authoritative FAIL in the Environment Variables section.
		results = append(results, CheckResult{Name: "jules", Status: statusSkip, Details: EnvJulesAPIKey + " not set (see Environment Variables)"})
	}

	return results, passed
}

// SecretsChecker verifies existence and file permissions of private keys/secrets.
type SecretsChecker struct{}

func (c *SecretsChecker) Name() string { return "Secrets & Encryption" }

func (c *SecretsChecker) Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool) {
	var results []CheckResult
	passed := true

	for _, sec := range state.Config.Verify.Secrets {
		absPath := ExpandPath(sec.Path)
		info, err := os.Stat(absPath)
		if os.IsNotExist(err) {
			results = append(results, CheckResult{Name: sec.Path, Status: statusWarn, Details: "MISSING"})
			continue
		} else if err != nil {
			results = append(results, CheckResult{Name: sec.Path, Status: statusFail, Details: fmt.Sprintf("Error checking: %s", err)})
			passed = false
			continue
		}

		perms := info.Mode().Perm()

		// A missing required_perms (0) means "presence-only": we cannot judge the
		// mode, so never flag it — and, crucially, never chmod it to 0000 under --fix.
		if sec.RequiredPerm == 0 {
			results = append(results, CheckResult{Name: sec.Path, Status: statusPass, Details: fmt.Sprintf("present (permissions: %04o)", perms)})
			continue
		}

		// Secure when the file carries no bits looser than the configured maximum
		// (e.g. 0600 accepts a stricter 0400 but flags 0644). An exact-match check
		// would wrongly report 0400 as insecure and --fix would then loosen it.
		expectedPerms := os.FileMode(sec.RequiredPerm)
		if perms&^expectedPerms == 0 {
			results = append(results, CheckResult{Name: sec.Path, Status: statusPass, Details: fmt.Sprintf("secure (permissions: %04o)", perms)})
		} else {
			if shouldFix {
				err := os.Chmod(absPath, expectedPerms)
				if err == nil {
					results = append(results, CheckResult{Name: sec.Path, Status: statusPass, Details: fmt.Sprintf("repaired (permissions: %04o)", expectedPerms)})
					continue
				}
			}
			results = append(results, CheckResult{Name: sec.Path, Status: statusFail, Details: fmt.Sprintf("INSECURE permissions: %04o (expected %04o)", perms, expectedPerms)})
			passed = false
		}
	}

	return results, passed
}

// DockerChecker checks if Docker CLI is installed and the daemon is reachable.
type DockerChecker struct{}

func (c *DockerChecker) Name() string { return "Docker Service" }

func (c *DockerChecker) Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool) {
	var results []CheckResult
	passed := true

	_, err := state.Runner.LookPath("docker")
	if err != nil {
		results = append(results, CheckResult{Name: "docker", Status: statusFail, Details: "Docker not installed"})
		passed = false
	} else {
		_, err := state.Runner.Run(ctx, "", nil, "docker", "info")
		if err == nil {
			results = append(results, CheckResult{Name: "docker", Status: statusPass, Details: "running"})
		} else {
			results = append(results, CheckResult{Name: "docker", Status: statusFail, Details: "NOT running"})
			passed = false
		}
	}

	return results, passed
}

// ToolsChecker checks if required CLI binaries exist in the system PATH.
type ToolsChecker struct{}

func (c *ToolsChecker) Name() string { return "CLI Tools" }

func (c *ToolsChecker) Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool) {
	var results []CheckResult
	passed := true

	for _, tool := range state.Config.Verify.Tools {
		path, err := state.Runner.LookPath(tool)
		if err == nil {
			results = append(results, CheckResult{Name: tool, Status: statusPass, Details: fmt.Sprintf("in PATH (%s)", path)})
		} else {
			results = append(results, CheckResult{Name: tool, Status: statusFail, Details: "MISSING"})
			passed = false
		}
	}

	return results, passed
}

// PrintHumanResults outputs the verification results in a user-friendly console format.
func PrintHumanResults(w io.Writer, res *VerifyResults) {
	section(w, "Environment Variables")
	for _, env := range res.EnvVars {
		printRow(w, env.Status, fmt.Sprintf("%-32s", env.Name), env.Details)
	}

	_, _ = fmt.Fprintln(w)
	section(w, "CLI Authentication")
	for _, au := range res.Auth {
		printRow(w, au.Status, fmt.Sprintf("%-12s", au.Name), au.Details)
	}

	_, _ = fmt.Fprintln(w)
	section(w, "Secrets & Encryption")
	for _, sec := range res.Secrets {
		printRow(w, sec.Status, sec.Name, sec.Details)
	}

	_, _ = fmt.Fprintln(w)
	section(w, "System Services")
	for _, dk := range res.Docker {
		printRow(w, dk.Status, dk.Name, dk.Details)
	}

	_, _ = fmt.Fprintln(w)
	section(w, "CLI Tools")
	for _, tl := range res.Tools {
		printRow(w, tl.Status, fmt.Sprintf("%-12s", tl.Name), tl.Details)
	}
}

func printRow(w io.Writer, status, name, details string) {
	var icon string
	switch status {
	case statusPass:
		icon = passIcon
	case statusFail:
		icon = failIcon
	case statusWarn:
		icon = warnIcon
	case statusSkip:
		icon = skipIcon
	}
	_, _ = fmt.Fprintf(w, "  %s %s %s\n", icon, name, details)
}

// VerifyConfig represents the configuration for system sanity checks.
type VerifyConfig struct {
	EnvVars EnvVarsConfig  `yaml:"env_vars"`
	Tools   []string       `yaml:"tools"`
	Secrets []SecretConfig `yaml:"secrets"`
}

// EnvVarsConfig represents the environment variables verification configuration.
type EnvVarsConfig struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// SecretConfig represents the configuration for checking a secret/key file.
type SecretConfig struct {
	Path         string `yaml:"path"`
	RequiredPerm int    `yaml:"required_perms"`
}

func defaultVerifyConfig() VerifyConfig {
	return VerifyConfig{
		EnvVars: EnvVarsConfig{
			Required: []string{EnvJulesAPIKey, EnvStitchAccessToken},
			Optional: []string{EnvStudioAPIKey, EnvKaggleAPIToken, EnvGWSProject, EnvAntigravityCloudProject, EnvAntigravityCloudLocation},
		},
		Tools: []string{
			"age", "agy", "chezmoi", "clasp", "docker", "dprint", "gcloud", "gh", "git", "go", "gws",
			"helm", "helmfile", "jules", "k3d", "k9s", "kubectl", "lefthook", "mise", "nvim",
			"opencode", "python", "skaffold", "uv",
		},
		Secrets: []SecretConfig{
			{
				Path:         "~/.config/chezmoi/key.txt",
				RequiredPerm: 0o600,
			},
		},
	}
}
