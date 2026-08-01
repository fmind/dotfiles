package dot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	ProbeHealthy         = "healthy"
	ProbeMissing         = "missing"
	ProbeBroken          = "broken"
	ProbeUnauthenticated = "unauthenticated"
	ProbeSkipped         = "skipped"

	defaultProbeTimeout = 3 * time.Second
	defaultOutputLimit  = 4 * 1024
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

// CapabilityProbeRegistry returns a defensive copy of the local, non-network probe registry.
func CapabilityProbeRegistry() map[string]CapabilityProbe {
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
			Timeout: defaultProbeTimeout, OutputLimit: defaultOutputLimit,
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

func runToolProbes(ctx context.Context, runner Runner, names []string, registry map[string]CapabilityProbe) ([]CheckResult, bool) {
	results := make([]CheckResult, len(names))
	passed := true
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, name := range names {
		wg.Go(func() {
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
				return
			}

			path, err := runner.LookPath(probe.Command)
			if err != nil {
				results[i] = CheckResult{Name: name, Status: statusFail, Condition: ProbeMissing, Details: "command not found"}
				mu.Lock()
				passed = false
				mu.Unlock()
				return
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
				return
			}
			results[i] = CheckResult{Name: name, Status: statusPass, Condition: ProbeHealthy, Path: path, Details: "capability probe passed"}
		})
	}
	wg.Wait()
	return results, passed
}
