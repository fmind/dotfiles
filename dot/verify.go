package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
)

// defaultVerifyTimeout bounds one checker suite so a hung external CLI cannot stall
// the whole verification run.
const defaultVerifyTimeout = 30 * time.Second

type VerifyResults struct {
	EnvVars []CheckResult `json:"env_vars"`
	Auth    []CheckResult `json:"auth"`
	Secrets []CheckResult `json:"secrets"`
	Docker  []CheckResult `json:"docker"`
	Tools   []CheckResult `json:"tools"`
	Install []CheckResult `json:"install"`
	Passed  bool          `json:"passed"`
}

type CheckResult struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Condition string `json:"condition,omitzero"`
	Path      string `json:"path,omitzero"`
	Details   string `json:"details,omitzero"`
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
	timeout := positiveOr(state.Config.Verify.Timeout.Duration(), defaultVerifyTimeout)
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
		{
			checker: &InstallChecker{},
			assign:  func(res []CheckResult) { results.Install = res },
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
			childCtx, cancel := context.WithTimeout(ctx, timeout)
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
			path, err := state.Runner.LookPath(task.cmdName)
			if err != nil {
				results[i] = CheckResult{Name: task.label, Status: statusSkip, Condition: ProbeSkipped, Details: task.cmdName + " not installed"}
				return
			}

			_, err = state.Runner.Run(ctx, "", nil, task.cmdName, task.args...)
			if err == nil {
				results[i] = CheckResult{Name: task.label, Status: statusPass, Condition: ProbeHealthy, Path: path, Details: "authenticated"}
			} else {
				results[i] = CheckResult{Name: task.label, Status: statusFail, Condition: ProbeUnauthenticated, Path: path, Details: "NOT authenticated"}
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

// InstallChecker compares the deployed `dot` binary against the source checkout.
//
// The binary is built from this repo and deployed by chezmoi, so nothing forces
// the two to stay in step: a release can be tagged and pushed while the copy on
// PATH is still whatever was last applied — possibly built from a dirty tree.
// Every other check here runs *through* that stale binary, so the skew is
// invisible precisely when it matters most.
type InstallChecker struct{}

func (c *InstallChecker) Name() string { return "Install Freshness" }

// dotBuildInputs are the repository paths the binary is actually compiled from.
// Test files and the module's task config are deliberately excluded: they change
// constantly and cannot alter the shipped binary, so counting them would report a
// perfectly current install as stale after a test-only or docs-only commit.
var dotBuildInputs = []string{"dot/*.go", "dot/go.mod", "dot/go.sum", ":!dot/*_test.go"}

// shortCommit abbreviates a commit for display, matching the 12-character form
// the binary embeds so both sides of a staleness message line up. An empty value
// means no commit ever touched the build inputs.
func shortCommit(commit string) string {
	if commit == "" {
		return "none"
	}
	return commit[:min(len(commit), 12)]
}

// lastBuildInputCommit returns the newest commit reachable from ref that touched
// the binary's build inputs.
//
// Comparing this value between HEAD and the revision baked into the installed
// binary answers the question that actually matters — "was this binary built from
// sources as new as the checkout?" — rather than "has any commit landed since?".
// Both sides run the same query, so an equality test is enough and no exit-code
// interpretation is needed.
func lastBuildInputCommit(ctx context.Context, state *GlobalState, source, ref string) (string, error) {
	args := append([]string{"rev-list", "-1", ref, "--"}, dotBuildInputs...)
	out, err := state.Runner.Run(ctx, source, nil, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// installedRevision extracts the short VCS revision and dirty flag from the
// output of `dot version` (e.g. "dot 1.7.0 (56fc23c35a3e, dirty)").
func installedRevision(version string) (revision string, dirty, ok bool) {
	open := strings.LastIndex(version, "(")
	closing := strings.LastIndex(version, ")")
	if open == -1 || closing == -1 || closing < open {
		return "", false, false
	}
	body := version[open+1 : closing]
	revision, flag, hasFlag := strings.Cut(body, ",")
	revision = strings.TrimSpace(revision)
	if revision == "" {
		return "", false, false
	}
	return revision, hasFlag && strings.TrimSpace(flag) == "dirty", true
}

func (c *InstallChecker) Check(ctx context.Context, state *GlobalState, _ bool) ([]CheckResult, bool) {
	const name = "dot binary"

	source, err := state.Runner.Run(ctx, "", nil, "chezmoi", "source-path")
	if err != nil {
		return []CheckResult{{Name: name, Status: statusSkip, Details: "chezmoi source path unavailable"}}, true
	}
	source = strings.TrimSpace(source)

	wanted, err := lastBuildInputCommit(ctx, state, source, "HEAD")
	if err != nil {
		return []CheckResult{{Name: name, Status: statusSkip, Details: "source checkout is not a git repository"}}, true
	}

	path, err := state.Runner.LookPath("dot")
	if err != nil {
		return []CheckResult{{Name: name, Status: statusSkip, Details: "dot not installed on PATH"}}, true
	}

	version, err := state.Runner.Run(ctx, "", nil, path, "version")
	if err != nil {
		return []CheckResult{{Name: name, Status: statusFail, Details: "installed binary failed to report its version"}}, false
	}

	revision, dirty, ok := installedRevision(version)
	if !ok {
		// A binary built outside a Git checkout carries no revision at all; that
		// is a packaging choice, not a staleness signal.
		return []CheckResult{{Name: name, Status: statusSkip, Details: "installed binary carries no VCS revision"}}, true
	}

	// A revision the checkout cannot resolve is unknowable, not stale: the binary
	// may predate a rebase, or come from another clone entirely. Say so instead of
	// asserting a staleness verdict the evidence does not support.
	built, err := lastBuildInputCommit(ctx, state, source, revision)
	if err != nil {
		return []CheckResult{{
			Name:    name,
			Status:  statusWarn,
			Details: fmt.Sprintf("installed binary revision %s is not present in this checkout", revision),
		}}, true
	}

	switch {
	case built != wanted:
		return []CheckResult{{
			Name:   name,
			Status: statusFail,
			// Both sides are named because the binary can also be *ahead* of the
			// checkout (an older commit is checked out), which "outdated" would misstate.
			Details: fmt.Sprintf("STALE: binary carries dot sources from %s, checkout has %s (run `mise run apply`)", shortCommit(built), shortCommit(wanted)),
		}}, false
	case dirty:
		return []CheckResult{{
			Name:    name,
			Status:  statusWarn,
			Details: fmt.Sprintf("built from %s with uncommitted changes", revision),
		}}, true
	default:
		return []CheckResult{{Name: name, Status: statusPass, Details: "built from current dot sources (" + revision + ")"}}, true
	}
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
type ToolsChecker struct {
	Registry map[string]CapabilityProbe
}

func (c *ToolsChecker) Name() string { return "CLI Tools" }

func (c *ToolsChecker) Check(ctx context.Context, state *GlobalState, shouldFix bool) ([]CheckResult, bool) {
	registry := c.Registry
	if registry == nil {
		registry = CapabilityProbeRegistryWithTimeout(state.Config.Verify.ProbeTimeout.Duration())
	}
	return runToolProbes(ctx, state.Runner, state.Config.Verify.Tools, registry)
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
	section(w, "Install Freshness")
	for _, in := range res.Install {
		printRow(w, in.Status, in.Name, in.Details)
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
	// Timeout bounds one checker suite and ProbeTimeout bounds one capability probe
	// inside it, so a single slow CLI is tuned without loosening the whole suite.
	// Both fall back to the built-in default when absent or non-positive.
	Timeout      Duration `yaml:"timeout"`
	ProbeTimeout Duration `yaml:"probe_timeout"`
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
			Optional: []string{EnvStudioAPIKey, EnvKaggleAPIToken, EnvHuggingfaceAPIToken, EnvGWSProject, EnvAntigravityCloudProject, EnvAntigravityCloudLocation},
		},
		// Keep in sync with the toolchain `mise run check` needs: gitleaks and
		// trivy gate every commit, so a missing one fails the hook, not the scan.
		Tools: []string{
			"age", "agy", "chezmoi", "clasp", "claude", "codex", "copilot", "docker", "dprint", "gcloud", "gh", "git", "git-cliff", "gitleaks", "go", "gws",
			"helm", "helmfile", "jules", "k3d", "k9s", "kubectl", "lefthook", "mise", "nvim",
			"opencode", "python", "skaffold", "sqlite3", "trivy", "uv",
		},
		Secrets: []SecretConfig{
			{
				Path:         "~/.config/chezmoi/key.txt",
				RequiredPerm: 0o600,
			},
		},
		Timeout:      Duration(defaultVerifyTimeout),
		ProbeTimeout: Duration(defaultProbeTimeout),
	}
}
