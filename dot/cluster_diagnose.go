package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/urfave/cli/v3"
)

const ClusterDiagnosticSchemaVersion = "dot.cluster.diagnostics/v1"

const (
	defaultDiagnosticTimeout    = 10 * time.Second
	defaultDiagnosticWindow     = 30 * time.Minute
	defaultDiagnosticTailLines  = 100
	defaultDiagnosticMaxLines   = 500
	defaultDiagnosticMaxBytes   = 64 * 1024
	defaultDiagnosticMaxLogPods = 5
)

type ClusterDiagnosticConfig struct {
	RedactPatterns []string `yaml:"redact_patterns"`
}

type ClusterDiagnosticLimits struct {
	Timeout    time.Duration `json:"timeout"`
	TimeWindow time.Duration `json:"time_window"`
	TailLines  int           `json:"tail_lines"`
	MaxLines   int           `json:"max_lines"`
	MaxBytes   int           `json:"max_bytes"`
	MaxLogPods int           `json:"max_log_pods"`
}

type ClusterDiagnosticProbe struct {
	ID         string        `json:"id"`
	Category   string        `json:"category"`
	Tool       string        `json:"tool"`
	TimeWindow string        `json:"time_window"`
	Args       []string      `json:"args"`
	Timeout    time.Duration `json:"timeout"`
	MaxLines   int           `json:"max_lines"`
	MaxBytes   int           `json:"max_bytes"`
}

type ClusterDiagnosticTarget struct {
	Cluster     string `json:"cluster"`
	Context     string `json:"context"`
	Namespace   string `json:"namespace"`
	Fingerprint string `json:"fingerprint"`
}

type ClusterDiagnosticResult struct {
	ProbeID   string `json:"probe_id"`
	Status    string `json:"status"`
	Command   string `json:"command"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	Truncated bool   `json:"truncated"`
}

type ClusterDiagnosticManifest struct {
	CollectedAt   time.Time                 `json:"collected_at"`
	Target        ClusterDiagnosticTarget   `json:"target"`
	SchemaVersion string                    `json:"schema_version"`
	Results       []ClusterDiagnosticResult `json:"results"`
	Limits        ClusterDiagnosticLimits   `json:"limits"`
}

type clusterDiagnosticOptions struct {
	Now            func() time.Time
	Target         ClusterTargetOptions
	Namespace      string
	OutputPath     string
	RedactPatterns []string
	Limits         ClusterDiagnosticLimits
}

var (
	privateKeyPattern = regexp.MustCompile(`(?s)-----BEGIN [^-\n]*PRIVATE KEY-----.*?-----END [^-\n]*PRIVATE KEY-----`)
	jwtPattern        = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)
	knownTokenPattern = regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|AIza[0-9A-Za-z_-]{20,}|AKIA[0-9A-Z]{16})\b`)
	bearerPattern     = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/-]+=*`)
	sensitiveKV       = regexp.MustCompile(`(?im)((?:\b(?:authorization|cookie)\b|\b[A-Za-z0-9_]*(?:credential|password|passwd|private[_-]?key|secret|token|api[_-]?key)[A-Za-z0-9_]*\b)\s*[:=]\s*)(?:"[^"]*"|'[^']*'|[^\s,}\]]+)`)
	sensitiveLabel    = regexp.MustCompile(`(?im)((?:customer|tenant|internal)[A-Za-z0-9_.-]*\s*[:=]\s*)(?:"[^"]*"|'[^']*'|[^\s,}\]]+)`)
	secretKindPattern = regexp.MustCompile(`(?im)^\s*kind\s*:\s*Secret\s*$|"kind"\s*:\s*"Secret"`)
)

func defaultClusterDiagnosticLimits() ClusterDiagnosticLimits {
	return ClusterDiagnosticLimits{
		Timeout:    defaultDiagnosticTimeout,
		TimeWindow: defaultDiagnosticWindow,
		TailLines:  defaultDiagnosticTailLines,
		MaxLines:   defaultDiagnosticMaxLines,
		MaxBytes:   defaultDiagnosticMaxBytes,
		MaxLogPods: defaultDiagnosticMaxLogPods,
	}
}

func NewClusterDiagnoseCmd(state *GlobalState) *cli.Command {
	defaults := defaultClusterDiagnosticLimits()
	return &cli.Command{
		Name:    "diagnose",
		Aliases: []string{"g"},
		Usage:   "Collect a bounded, sanitized, read-only cluster evidence bundle",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "namespace", Aliases: []string{"n"}, Usage: "Namespace to verify and inspect (defaults to the target context namespace)"},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Owner-only JSON manifest path"},
			&cli.DurationFlag{Name: "timeout", Value: defaults.Timeout, Usage: "Timeout for every probe"},
			&cli.DurationFlag{Name: "since", Value: defaults.TimeWindow, Usage: "Event and log time window"},
			&cli.IntFlag{Name: "tail", Value: defaults.TailLines, Usage: "Maximum lines requested from each container log"},
			&cli.IntFlag{Name: "max-lines", Value: defaults.MaxLines, Usage: "Maximum retained lines per probe"},
			&cli.IntFlag{Name: "max-bytes", Value: defaults.MaxBytes, Usage: "Maximum retained bytes per probe"},
			&cli.IntFlag{Name: "max-log-pods", Value: defaults.MaxLogPods, Usage: "Maximum pods whose logs are collected"},
			&cli.StringSliceFlag{Name: "redact-pattern", Usage: "Additional RE2 pattern to redact (repeatable)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			patterns := append([]string{}, state.Config.Cluster.Diagnostics.RedactPatterns...)
			patterns = append(patterns, cmd.StringSlice("redact-pattern")...)
			return RunClusterDiagnose(ctx, state, clusterDiagnosticOptions{
				Target:         clusterTargetOptions(cmd),
				Namespace:      cmd.String("namespace"),
				OutputPath:     cmd.String("output"),
				RedactPatterns: patterns,
				Limits: ClusterDiagnosticLimits{
					Timeout:    cmd.Duration("timeout"),
					TimeWindow: cmd.Duration("since"),
					TailLines:  cmd.Int("tail"),
					MaxLines:   cmd.Int("max-lines"),
					MaxBytes:   cmd.Int("max-bytes"),
					MaxLogPods: cmd.Int("max-log-pods"),
				},
			})
		},
	}
}

func validateClusterDiagnosticLimits(limits ClusterDiagnosticLimits) error {
	if limits.Timeout <= 0 || limits.TimeWindow <= 0 || limits.TailLines <= 0 || limits.MaxLines <= 0 || limits.MaxBytes <= 0 || limits.MaxLogPods <= 0 {
		return errors.New("diagnostic timeout, time window, tail, line, byte, and log-pod bounds must all be positive")
	}
	return nil
}

func namespacedArgs(namespace string, args ...string) []string {
	if namespace == "" {
		return args
	}
	return append([]string{"--namespace", namespace}, args...)
}

func ClusterDiagnosticPlan(namespace string, limits ClusterDiagnosticLimits) []ClusterDiagnosticProbe {
	snapshot := "snapshot"
	window := limits.TimeWindow.String()
	probe := func(id, category, tool, timeWindow string, args ...string) ClusterDiagnosticProbe {
		return ClusterDiagnosticProbe{ID: id, Category: category, Tool: tool, Args: args, Timeout: limits.Timeout, TimeWindow: timeWindow, MaxLines: limits.MaxLines, MaxBytes: limits.MaxBytes}
	}
	return []ClusterDiagnosticProbe{
		probe("k3d-version", "versions", "k3d", snapshot, "version"),
		probe("kubernetes-version", "versions", "kubectl", snapshot, "version", "--output=json"),
		probe("nodes", "nodes", "kubectl", snapshot, "get", "nodes", "-o", "wide"),
		probe("workloads", "workloads", "kubectl", snapshot, namespacedArgs(namespace, "get", "pods,deployments,statefulsets,daemonsets,jobs", "-o", "wide")...),
		probe("controller-conditions", "conditions", "kubectl", snapshot, namespacedArgs(namespace, "get", "deployments,statefulsets,daemonsets,jobs", "-o", `jsonpath={range .items[*]}{.kind}{"\t"}{.metadata.namespace}{"\t"}{.metadata.name}{"\t"}{range .status.conditions[*]}{.type}{"="}{.status}{":"}{.reason}{";"}{end}{"\n"}{end}`)...),
		probe("recent-events", "events", "kubectl", window, namespacedArgs(namespace, "get", "events", "-o", "json")...),
		probe("node-usage", "resources", "kubectl", snapshot, "top", "nodes"),
		probe("pod-usage", "resources", "kubectl", snapshot, namespacedArgs(namespace, "top", "pods", "--containers")...),
		probe("storage", "storage", "kubectl", snapshot, namespacedArgs(namespace, "get", "persistentvolumes,persistentvolumeclaims,storageclasses", "-o", "wide")...),
		probe("log-targets", "logs", "kubectl", window, namespacedArgs(namespace, "get", "pods", "-o", `jsonpath={range .items[*]}{.metadata.namespace}{"\t"}{.metadata.name}{"\n"}{end}`)...),
	}
}

type diagnosticSanitizer struct {
	projectPatterns []*regexp.Regexp
}

func newDiagnosticSanitizer(patterns []string) (diagnosticSanitizer, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		candidate, err := regexp.Compile(pattern)
		if err != nil {
			return diagnosticSanitizer{}, fmt.Errorf("invalid diagnostic redaction pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, candidate)
	}
	return diagnosticSanitizer{projectPatterns: compiled}, nil
}

func (sanitizer diagnosticSanitizer) sanitize(value string) string {
	if secretKindPattern.MatchString(value) {
		return "[REDACTED KUBERNETES SECRET]"
	}
	value = privateKeyPattern.ReplaceAllString(value, "[REDACTED PRIVATE KEY]")
	value = jwtPattern.ReplaceAllString(value, "[REDACTED TOKEN]")
	value = knownTokenPattern.ReplaceAllString(value, "[REDACTED TOKEN]")
	value = bearerPattern.ReplaceAllString(value, "Bearer [REDACTED]")
	value = sensitiveKV.ReplaceAllString(value, "${1}[REDACTED]")
	value = sensitiveLabel.ReplaceAllString(value, "${1}[REDACTED]")
	for _, pattern := range sanitizer.projectPatterns {
		value = pattern.ReplaceAllString(value, "[REDACTED PROJECT VALUE]")
	}
	return value
}

func boundDiagnosticText(value string, maxLines, maxBytes int) (string, bool) {
	const marker = "[TRUNCATED]"
	lines := strings.Split(value, "\n")
	truncated := false
	if len(lines) > maxLines {
		if maxLines == 1 {
			lines = []string{marker}
		} else {
			lines = append(lines[:maxLines-1], marker)
		}
		truncated = true
	}
	value = strings.Join(lines, "\n")
	if len(value) > maxBytes {
		if maxBytes <= len(marker) {
			value = marker[:maxBytes]
		} else {
			limit := maxBytes - len(marker) - 1
			value = value[:limit]
			for !utf8.ValidString(value) {
				value = value[:len(value)-1]
			}
			value = strings.TrimRight(value, "\n") + "\n" + marker
		}
		truncated = true
	}
	return value, truncated
}

func diagnosticCommand(probe ClusterDiagnosticProbe) string {
	return strings.Join(append([]string{probe.Tool}, probe.Args...), " ")
}

type diagnosticEventList struct {
	Items []struct {
		EventTime     time.Time `json:"eventTime"`
		LastTimestamp time.Time `json:"lastTimestamp"`
		Metadata      struct {
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Namespace         string    `json:"namespace"`
		} `json:"metadata"`
		InvolvedObject struct {
			Kind string `json:"kind"`
			Name string `json:"name"`
		} `json:"involvedObject"`
		Type    string `json:"type"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"items"`
}

func filterRecentDiagnosticEvents(output string, now time.Time, window time.Duration) (string, error) {
	var events diagnosticEventList
	if err := json.Unmarshal([]byte(output), &events); err != nil {
		return "", fmt.Errorf("failed to decode Kubernetes events: %w", err)
	}
	cutoff := now.Add(-window)
	lines := make([]string, 0, len(events.Items))
	for _, event := range events.Items {
		timestamp := event.EventTime
		if timestamp.IsZero() {
			timestamp = event.LastTimestamp
		}
		if timestamp.IsZero() {
			timestamp = event.Metadata.CreationTimestamp
		}
		if timestamp.Before(cutoff) {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s/%s\t%s\t%s", timestamp.UTC().Format(time.RFC3339), event.Type, event.Metadata.Namespace, event.InvolvedObject.Kind, event.InvolvedObject.Name, event.Reason, event.Message))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n"), nil
}

func collectDiagnosticProbe(ctx context.Context, state *GlobalState, target clusterTarget, probe ClusterDiagnosticProbe, sanitizer diagnosticSanitizer, now time.Time) (ClusterDiagnosticResult, string) {
	probeCtx, cancel := context.WithTimeout(ctx, probe.Timeout)
	defer cancel()
	args := probe.Args
	if probe.Tool == "kubectl" {
		args = target.kubectlArgs(append([]string{"--request-timeout=" + probe.Timeout.String()}, args...)...)
	}
	output, err := state.Runner.Run(probeCtx, "", nil, probe.Tool, args...)
	rawOutput := output
	if err == nil && probe.ID == "recent-events" {
		output, err = filterRecentDiagnosticEvents(output, now, targetDiagnosticWindow(probe))
	}
	result := ClusterDiagnosticResult{ProbeID: probe.ID, Status: "ok", Command: sanitizer.sanitize(diagnosticCommand(probe))}
	result.Output, result.Truncated = boundDiagnosticText(sanitizer.sanitize(output), probe.MaxLines, probe.MaxBytes)
	if err != nil {
		result.Status = "error"
		var errorTruncated bool
		result.Error, errorTruncated = boundDiagnosticText(sanitizer.sanitize(err.Error()), probe.MaxLines, probe.MaxBytes)
		result.Truncated = result.Truncated || errorTruncated
	}
	return result, rawOutput
}

func targetDiagnosticWindow(probe ClusterDiagnosticProbe) time.Duration {
	window, err := time.ParseDuration(probe.TimeWindow)
	if err != nil {
		return 0
	}
	return window
}

func parseDiagnosticLogTargets(output, fallbackNamespace string, maximum int) [][2]string {
	seen := make(map[string]struct{})
	targets := make([][2]string, 0, maximum)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 1 && fallbackNamespace != "" {
			fields = []string{fallbackNamespace, fields[0]}
		}
		if len(fields) != 2 {
			continue
		}
		key := fields[0] + "\x00" + fields[1]
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, [2]string{fields[0], fields[1]})
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i][0] == targets[j][0] {
			return targets[i][1] < targets[j][1]
		}
		return targets[i][0] < targets[j][0]
	})
	if len(targets) > maximum {
		targets = targets[:maximum]
	}
	return targets
}

func diagnosticLogProbe(index int, namespace, pod string, limits ClusterDiagnosticLimits) ClusterDiagnosticProbe {
	args := namespacedArgs(namespace, "logs", pod, "--all-containers=true", "--prefix=true", "--since="+limits.TimeWindow.String(), "--tail="+strconv.Itoa(limits.TailLines))
	return ClusterDiagnosticProbe{ID: fmt.Sprintf("logs/%03d", index+1), Category: "logs", Tool: "kubectl", Args: args, Timeout: limits.Timeout, TimeWindow: limits.TimeWindow.String(), MaxLines: limits.MaxLines, MaxBytes: limits.MaxBytes}
}

func verifyDiagnosticNamespace(ctx context.Context, state *GlobalState, target clusterTarget, namespace string, timeout time.Duration) error {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := target.kubectlArgs("--request-timeout="+timeout.String(), "get", "namespace", namespace, "--ignore-not-found", "-o", "name")
	output, err := state.Runner.Run(probeCtx, "", nil, "kubectl", args...)
	if err != nil {
		return fmt.Errorf("failed to verify diagnostic namespace %q: %w", namespace, err)
	}
	if strings.TrimSpace(output) != "namespace/"+namespace {
		return fmt.Errorf("diagnostic namespace %q does not exist on verified context %q", namespace, target.Context)
	}
	return nil
}

func defaultDiagnosticOutputPath(now time.Time, cluster string) string {
	name := fmt.Sprintf("%s-%s.json", cluster, now.UTC().Format("20060102T150405Z"))
	return ExpandPath(filepath.Join("~", ".local", "state", "dot", "diagnostics", name))
}

func RunClusterDiagnose(ctx context.Context, state *GlobalState, options clusterDiagnosticOptions) error {
	if err := validateClusterDiagnosticLimits(options.Limits); err != nil {
		return err
	}
	sanitizer, sanitizerErr := newDiagnosticSanitizer(options.RedactPatterns)
	if sanitizerErr != nil {
		return sanitizerErr
	}
	if toolErr := requireTools(state, "k3d", "kubectl"); toolErr != nil {
		return toolErr
	}
	if kubeconfigErr := ensureClusterKubeconfig(ctx, state, options.Target); kubeconfigErr != nil {
		return kubeconfigErr
	}
	target, err := resolveClusterTarget(ctx, state, options.Target, true)
	if err != nil {
		return err
	}
	namespace := options.Namespace
	if namespace == "" {
		namespace = target.Namespace
	}
	if namespaceErr := verifyDiagnosticNamespace(ctx, state, target, namespace, options.Limits.Timeout); namespaceErr != nil {
		return namespaceErr
	}
	target.Namespace = namespace
	reportClusterTarget(state, target)

	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	manifest := ClusterDiagnosticManifest{
		SchemaVersion: ClusterDiagnosticSchemaVersion,
		CollectedAt:   now().UTC(),
		Target: ClusterDiagnosticTarget{
			Cluster: sanitizer.sanitize(target.Name), Context: sanitizer.sanitize(target.Context), Namespace: sanitizer.sanitize(target.Namespace), Fingerprint: target.Fingerprint,
		},
		Limits:  options.Limits,
		Results: make([]ClusterDiagnosticResult, 0),
	}
	for _, probe := range ClusterDiagnosticPlan(namespace, options.Limits) {
		result, rawOutput := collectDiagnosticProbe(ctx, state, target, probe, sanitizer, manifest.CollectedAt)
		manifest.Results = append(manifest.Results, result)
		if probe.ID != "log-targets" || result.Status != "ok" {
			continue
		}
		for index, logTarget := range parseDiagnosticLogTargets(rawOutput, namespace, options.Limits.MaxLogPods) {
			logResult, _ := collectDiagnosticProbe(ctx, state, target, diagnosticLogProbe(index, logTarget[0], logTarget[1], options.Limits), sanitizer, manifest.CollectedAt)
			manifest.Results = append(manifest.Results, logResult)
		}
	}

	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode diagnostic manifest: %w", err)
	}
	payload = append(payload, '\n')
	outputPath := options.OutputPath
	if outputPath == "" {
		outputPath = defaultDiagnosticOutputPath(manifest.CollectedAt, target.Name)
	} else {
		outputPath = ExpandPath(outputPath)
	}
	if err := writeOwnerOnlyFile(outputPath, payload, "diagnostic manifest"); err != nil {
		return err
	}

	failures := 0
	truncated := 0
	for _, result := range manifest.Results {
		if result.Status == "error" {
			failures++
		}
		if result.Truncated {
			truncated++
		}
	}
	_, _ = fmt.Fprintf(state.Stdout, "Diagnostic bundle: %s\nSchema: %s\nResults: %d succeeded, %d partial errors, %d truncated\nUpload: disabled\n", outputPath, manifest.SchemaVersion, len(manifest.Results)-failures, failures, truncated)
	return nil
}
