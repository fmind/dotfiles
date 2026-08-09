package dot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	ProbeHealthy         = "healthy"
	ProbeMissing         = "missing"
	ProbeBroken          = "broken"
	ProbeUnauthenticated = "unauthenticated"
	ProbeSkipped         = "skipped"

	// Sized for the slowest CLIs in the registry rather than the fastest, and measured
	// under concurrency rather than idle: the interpreted wrappers (clasp on Node,
	// gcloud on Python) load large bundles from disk and degrade superlinearly when
	// probes overlap — 5s idle, 20-35s alongside seven peers. This bound exists to
	// catch a *hung* tool, not to benchmark start-up, and a healthy probe returns the
	// moment it finishes, so a generous bound is only ever paid on a genuine hang.
	// Override per machine with `verify.probe_timeout`.
	defaultProbeTimeout = 45 * time.Second
	defaultOutputLimit  = 4 * 1024

	// defaultProbeConcurrency caps simultaneous probes. Unbounded fan-out is what
	// makes the per-probe timeout above lie: the registry holds ~30 tools, several of
	// them interpreted CLIs that each start a Node or Python runtime, and starting
	// them all at once multiplies a 1-7s cold start several-fold through CPU and page-
	// cache contention — reporting healthy tools as broken. Matches pull's default.
	defaultProbeConcurrency = 8
)

// CapabilityProbe is the stable contract shared by verify and agent-specific health checks.
type CapabilityProbe struct {
	Name         string
	Command      string
	Args         []string
	Timeout      time.Duration
	OutputLimit  int
	RequiresAuth bool
}

// CapabilityProbeRegistry returns a defensive copy of the local, non-network probe
// registry bounded by the built-in default timeout.
func CapabilityProbeRegistry() map[string]CapabilityProbe {
	return CapabilityProbeRegistryWithTimeout(defaultProbeTimeout)
}

// CapabilityProbeRegistryWithTimeout is CapabilityProbeRegistry with a caller-chosen
// per-probe bound. Some probes are slow to start on a cold cache (interpreted CLIs
// especially), and a bound tuned for a warm machine reports them broken when they
// are merely slow — so the bound is a setting rather than a constant.
func CapabilityProbeRegistryWithTimeout(timeout time.Duration) map[string]CapabilityProbe {
	timeout = positiveOr(timeout, defaultProbeTimeout)
	args := map[string][]string{
		"age": {"--version"}, "agy": {"--help"}, "chezmoi": {"--version"}, "clasp": {"--version"},
		"claude": {"--version"}, "codex": {"--version"}, "copilot": {"--version"}, "docker": {"--version"},
		"dprint": {"--version"}, "gcloud": {"--version"}, "gh": {"--version"}, "git": {"--version"},
		"git-cliff": {"--version"}, "gitleaks": {"version"}, "go": {"version"}, "gws": {"--version"},
		"gdbus": {"help"}, "helm": {"version", "--short"}, "helmfile": {"--version"}, "jules": {"--version"}, "k3d": {"version"},
		"k9s": {"version", "--short"}, "kubectl": {"version", "--client"}, "lefthook": {"version"}, "mise": {"--version"},
		"notify-send": {"--help"}, "nvim": {"--version"}, "opencode": {"--version"}, "osascript": {"-e", "return \"ok\""}, "python": {"--version"}, "skaffold": {"version"},
		"sqlite3": {"--version"}, "trivy": {"--version"}, "uv": {"--version"},
	}
	args["dot"] = []string{"version"}

	registry := make(map[string]CapabilityProbe, len(args))
	for name, probeArgs := range args {
		registry[name] = CapabilityProbe{
			Name: name, Command: name, Args: append([]string(nil), probeArgs...),
			Timeout: timeout, OutputLimit: defaultOutputLimit,
		}
	}
	for _, name := range []string{"clasp", "gcloud", "gh", "gws", "jules"} {
		probe := registry[name]
		probe.RequiresAuth = true
		registry[name] = probe
	}
	return registry
}

type boundedRunner interface {
	RunBounded(ctx context.Context, dir, name string, limit int, args ...string) (string, error)
}

func runCapabilityProbe(ctx context.Context, runner Runner, probe CapabilityProbe) (string, error) {
	if bounded, ok := runner.(boundedRunner); ok {
		return bounded.RunBounded(ctx, "", probe.Command, probe.OutputLimit, probe.Args...)
	}
	output, err := runner.Run(ctx, "", nil, probe.Command, probe.Args...)
	return truncateProbeOutput(output, probe.OutputLimit), err
}

func truncateProbeOutput(output string, limit int) string {
	if limit <= 0 || len(output) <= limit {
		return output
	}
	return output[:limit] + "…"
}

func probeFailureDetails(err error, limit int) string {
	details := strings.Join(strings.Fields(err.Error()), " ")
	return truncateProbeOutput("probe failed: "+details, limit)
}

// runToolProbes probes every named tool, running at most `limit` probes at once.
// A non-positive limit falls back to defaultProbeConcurrency; the bound is what keeps
// each probe's timeout meaningful, so there is deliberately no unbounded mode.
func runToolProbes(ctx context.Context, runner Runner, names []string, registry map[string]CapabilityProbe, limit int) ([]CheckResult, bool) {
	results := make([]CheckResult, len(names))
	passed := true
	var mu sync.Mutex

	// Plain errgroup rather than WithContext: a broken probe is a reported result, not
	// a group error, so one slow tool must never cancel the probes still queued behind it.
	var group errgroup.Group
	group.SetLimit(positiveOr(limit, defaultProbeConcurrency))

	for i, name := range names {
		group.Go(func() error {
			defer func() {
				if recovered := recover(); recovered != nil {
					results[i] = CheckResult{Name: name, Status: statusFail, Condition: ProbeBroken, Details: fmt.Sprintf("probe panicked: %v", recovered)}
					mu.Lock()
					passed = false
					mu.Unlock()
				}
			}()
			probe, ok := registry[name]
			if !ok {
				results[i] = CheckResult{Name: name, Status: statusSkip, Condition: ProbeSkipped, Details: "no capability probe registered"}
				return nil
			}

			path, err := runner.LookPath(probe.Command)
			if err != nil {
				results[i] = CheckResult{Name: name, Status: statusFail, Condition: ProbeMissing, Details: "command not found"}
				mu.Lock()
				passed = false
				mu.Unlock()
				return nil
			}

			probe.Command = path
			probeCtx, cancel := context.WithTimeout(ctx, probe.Timeout)
			defer cancel()
			_, err = runCapabilityProbe(probeCtx, runner, probe)
			if err != nil {
				results[i] = CheckResult{Name: name, Status: statusFail, Condition: ProbeBroken, Path: path, Details: probeFailureDetails(err, probe.OutputLimit)}
				mu.Lock()
				passed = false
				mu.Unlock()
				return nil
			}
			results[i] = CheckResult{Name: name, Status: statusPass, Condition: ProbeHealthy, Path: path, Details: "capability probe passed"}
			return nil
		})
	}
	// The group func never returns an error; Wait is only the barrier.
	_ = group.Wait()
	return results, passed
}
