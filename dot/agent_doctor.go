package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	defaultDoctorScanLimit = 4096
	defaultDoctorStaleLag  = 24 * time.Hour
)

type AgentDoctorOptions struct {
	Fix    bool
	DryRun bool
}

type agentDoctorResult struct {
	Agent         string
	Discovery     string
	Hooks         string
	Tools         string
	Source        string
	LastIngestion string
	LastFailure   string
	ArchiveLag    string
	Omitted       int
	Healthy       bool
}

type agentLineageSummary struct {
	LastComplete time.Time
	Latest       time.Time
	Partial      bool
	Unreadable   bool
	Omitted      int
}

func NewAgentDoctorCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Check cross-agent discovery, hooks, sources, and ingestion health",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fix", Aliases: []string{"f"}, Usage: "Apply the managed agent integration targets with chezmoi"},
			&cli.BoolFlag{Name: "dry-run", Aliases: []string{"N"}, Usage: "Preview --fix without changing deployed files"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunAgentDoctor(ctx, state, AgentDoctorOptions{Fix: cmd.Bool("fix"), DryRun: cmd.Bool("dry-run")})
		},
	}
}

// doctorRepairTargets derives the chezmoi apply targets from the canonical agent
// table, so a new integration becomes repairable by adding one table entry.
func doctorRepairTargets(cfg AgentConfig) []string {
	targets := []string{ExpandPath(sharedPersonaPath), ExpandPath(sharedSkillsPath)}
	for _, definition := range cfg.definitions() {
		targets = append(targets, ExpandPath(definition.PersonaPath))
		if definition.SkillsPath != "" {
			targets = append(targets, ExpandPath(definition.SkillsPath))
		}
		if definition.HookPath != "" {
			targets = append(targets, ExpandPath(definition.HookPath))
		}
	}
	slices.Sort(targets)
	return slices.Compact(targets)
}

func repairAgentIntegrations(ctx context.Context, state *GlobalState, dryRun bool) error {
	args := []string{"apply"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--force")
	args = append(args, doctorRepairTargets(state.Config.Agent)...)
	if _, err := state.Runner.Run(ctx, "", nil, "chezmoi", args...); err != nil {
		return fmt.Errorf("failed to repair agent integrations: %w", err)
	}
	return nil
}

func RunAgentDoctor(ctx context.Context, state *GlobalState, options AgentDoctorOptions) error {
	if options.DryRun && !options.Fix {
		return errors.New("--dry-run requires --fix")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if options.Fix {
		if err := repairAgentIntegrations(ctx, state, options.DryRun); err != nil {
			return err
		}
	}
	results := gatherAgentDoctor(ctx, state, home, time.Now())
	healthy := true
	_, _ = fmt.Fprintln(state.Stdout, "Agent doctor")
	for _, result := range results {
		if !result.Healthy {
			healthy = false
		}
		_, _ = fmt.Fprintf(state.Stdout, "%s %s: discovery=%s hooks=%s tools=%s source=%s ingestion=%s failure=%s lag=%s omitted=%d\n", map[bool]string{true: passIcon, false: failIcon}[result.Healthy], result.Agent, result.Discovery, result.Hooks, result.Tools, result.Source, result.LastIngestion, result.LastFailure, result.ArchiveLag, result.Omitted)
	}
	if !healthy {
		return errors.New("agent doctor found unhealthy integrations")
	}
	return nil
}

func gatherAgentDoctor(ctx context.Context, state *GlobalState, home string, now time.Time) []agentDoctorResult {
	cfg := state.Config.Agent
	definitions := cfg.definitions()
	results := make([]agentDoctorResult, 0, len(definitions))
	registry := CapabilityProbeRegistryWithTimeout(state.Config.Verify.ProbeTimeout.Duration())
	// Resolve the hook-command surface once: every agent invokes the same dot binary,
	// so probing per agent would only repeat identical work.
	runnable := newDotCommandProber(ctx, state.Runner)
	for _, definition := range definitions {
		discovery, discoveryOK := checkAgentDiscovery(definition)
		hooks, hooksOK := checkAgentHooks(definition, runnable)
		probeNames := append([]string(nil), definition.ProbeNames...)
		if definition.Notifications {
			if notifier := doctorNotifierProbe(state.Runner); notifier != "" {
				probeNames = append(probeNames, notifier)
			} else {
				hooks, hooksOK = "notification-unavailable", false
			}
		}
		probeResults, probesOK := runToolProbes(ctx, state.Runner, probeNames, registry, state.Config.Verify.ProbeConcurrency)
		tools := summarizeDoctorProbes(probeResults)
		source, sourceTime, sourcePresent, sourceOK, sourceOmitted := inspectAgentSource(cfg, definition)
		lineage := inspectAgentLineage(cfg, home, definition.Agent)
		failure, failureOK := inspectLastHookFailure(cfg, home, definition.Agent)
		ingestion, lag, lineageOK := summarizeAgentLineage(cfg, now, sourceTime, sourcePresent, lineage)
		results = append(results, agentDoctorResult{
			Agent: definition.Agent, Discovery: discovery, Hooks: hooks, Tools: tools, Source: source,
			LastIngestion: ingestion, LastFailure: failure, ArchiveLag: lag, Omitted: sourceOmitted + lineage.Omitted,
			Healthy: discoveryOK && hooksOK && probesOK && sourceOK && lineageOK && failureOK,
		})
	}
	return results
}

// dotCommandProber reports whether a `dot ...` subcommand path exists in the dot
// binary on PATH — the one agent hooks actually invoke, which is not necessarily the
// binary running this check. A hook wired to a command a stale installed binary does
// not implement fails silently at runtime; only probing the deployed binary sees it.
type dotCommandProber func(args []string) bool

func newDotCommandProber(ctx context.Context, runner Runner) dotCommandProber {
	binary, err := runner.LookPath("dot")
	if err != nil {
		// Without a dot on PATH no hook can run at all. Report every command missing
		// rather than quietly passing the check that exists to catch exactly this.
		return func([]string) bool { return false }
	}
	cache := make(map[string]bool)
	return func(args []string) bool {
		if len(args) == 0 {
			return true
		}
		key := strings.Join(args, " ")
		if known, cached := cache[key]; cached {
			return known
		}
		// `--help` on an existing command exits 0; urfave/cli exits non-zero with
		// "No help topic for ..." when any segment of the path is unknown.
		_, runErr := runner.Run(ctx, "", nil, binary, append(slices.Clone(args), "--help")...)
		cache[key] = runErr == nil
		return cache[key]
	}
}

// hookCommandArguments extracts the subcommand path of a `dot ...` hook command. A
// hook command ends in positional operands (the agent name, then an event name);
// those are arguments, not subcommands, so probing them would fail against a
// perfectly healthy binary. Truncate at the first operand.
func hookCommandArguments(command string, agents map[string]bool) []string {
	fields := strings.Fields(command)
	if len(fields) < 2 || fields[0] != "dot" {
		return nil
	}
	args := fields[1:]
	for index, field := range args {
		if strings.HasPrefix(field, "-") || agents[field] {
			return args[:index]
		}
	}
	return args
}

func sameResolvedPath(path, target string) bool {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	want, err := filepath.EvalSymlinks(target)
	return err == nil && resolved == want
}

func checkAgentDiscovery(definition agentDefinition) (string, bool) {
	personaPath := ExpandPath(definition.PersonaPath)
	canonicalPersona := ExpandPath(sharedPersonaPath)
	canonicalSkills := ExpandPath(sharedSkillsPath)
	if definition.PersonaInFile {
		content, err := os.ReadFile(personaPath)
		if err != nil || !json.Valid(content) || !strings.Contains(string(content), sharedPersonaPath) {
			return "persona-broken", false
		}
	} else if !sameResolvedPath(personaPath, canonicalPersona) {
		return "persona-broken", false
	}
	skillsPath := ExpandPath(definition.resolvedSkillsPath())
	if skillsPath == canonicalSkills {
		if info, err := os.Stat(canonicalSkills); err != nil || !info.IsDir() {
			return "skills-missing", false
		}
	} else if !sameResolvedPath(skillsPath, canonicalSkills) {
		return "skills-broken", false
	}
	return "healthy", true
}

func checkAgentHooks(definition agentDefinition, runnable dotCommandProber) (string, bool) {
	if definition.HookPath == "" {
		return "sync-only", true
	}
	content, err := os.ReadFile(ExpandPath(definition.HookPath))
	if err != nil {
		return "missing", false
	}
	if definition.HookJSON && !json.Valid(content) {
		return "malformed", false
	}
	if definition.Agent == sessionStoreCopilot {
		var config struct {
			Version int `json:"version"`
		}
		if err := json.Unmarshal(content, &config); err != nil {
			return "malformed", false
		}
		if config.Version != 1 {
			return "unsupported-version", false
		}
	}
	agents := agentNameSet()
	for _, command := range definition.HookCommands {
		if !strings.Contains(string(content), command) {
			return "command-mismatch", false
		}
		// Wiring is necessary but not sufficient: chezmoi deploys hook files
		// independently of the dot binary, so a config can reference a subcommand the
		// installed binary predates. That hook then fails silently on every event,
		// which is precisely the drift this check exists to surface.
		if !runnable(hookCommandArguments(command, agents)) {
			return "command-unavailable", false
		}
	}
	return "healthy", true
}

func doctorNotifierProbe(runner Runner) string {
	switch runtime.GOOS {
	case "darwin":
		return "osascript"
	case "linux":
		for _, name := range []string{"notify-send", "gdbus"} {
			if _, err := runner.LookPath(name); err == nil {
				return name
			}
		}
	}
	return ""
}

func summarizeDoctorProbes(results []CheckResult) string {
	failed := make([]string, 0)
	for _, result := range results {
		if result.Status != statusPass {
			failed = append(failed, result.Name+":"+result.Condition)
		}
	}
	if len(failed) == 0 {
		return "healthy"
	}
	slices.Sort(failed)
	return strings.Join(failed, ",")
}

func inspectAgentSource(cfg AgentConfig, definition agentDefinition) (string, time.Time, bool, bool, int) {
	sourcePath := ExpandPath(definition.Source)
	scanLimit := cfg.scanLimit()
	info, err := os.Lstat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return "missing", time.Time{}, false, true, 0
	}
	if err != nil {
		return "unreadable", time.Time{}, false, false, 0
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "linked", time.Time{}, false, false, 0
	}
	if !info.IsDir() {
		return "present", info.ModTime(), true, true, 0
	}
	latest := time.Time{}
	seen := 0
	omitted := 0
	walkErr := filepath.WalkDir(sourcePath, func(path string, entry fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if seen >= scanLimit {
			omitted = 1
			return fs.SkipAll
		}
		seen++
		if _, recognized := rawSessionIdentity(sourcePath, path, definition.Agent); !recognized {
			return nil
		}
		entryInfo, infoErr := entry.Info()
		if infoErr == nil && entryInfo.ModTime().After(latest) {
			latest = entryInfo.ModTime()
		}
		return nil
	})
	if walkErr != nil {
		return "unreadable", latest, true, false, omitted
	}
	return "present", latest, true, true, omitted
}

func inspectAgentLineage(cfg AgentConfig, home, agent string) agentLineageSummary {
	root := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, agent)
	scanLimit := cfg.scanLimit()
	summary := agentLineageSummary{}
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return summary
	} else if err != nil {
		summary.Unreadable = true
		return summary
	}
	seen := 0
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if entry.IsDir() || entry.Name() != "manifest.json" {
			return nil
		}
		if seen >= scanLimit {
			summary.Omitted = 1
			return fs.SkipAll
		}
		seen++
		manifest, err := readSessionManifest(filepath.Dir(path))
		if err != nil || manifest.Agent != agent {
			summary.Unreadable = true
			return nil
		}
		ingested, err := time.Parse(time.RFC3339Nano, manifest.IngestedAt)
		if err != nil {
			summary.Unreadable = true
			return nil
		}
		if ingested.After(summary.Latest) {
			summary.Latest = ingested
			summary.Partial = manifest.Completeness != sessionComplete
		}
		if manifest.Completeness == sessionComplete && ingested.After(summary.LastComplete) {
			summary.LastComplete = ingested
		}
		return nil
	})
	if walkErr != nil {
		summary.Unreadable = true
	}
	return summary
}

func inspectLastHookFailure(cfg AgentConfig, home, agent string) (string, bool) {
	root := filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion)
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return "none", true
	}
	if err != nil {
		return "unreadable", false
	}
	unreadable := false
	first := max(len(entries)-cfg.hookFailureLimit(), 0)
	for index := len(entries) - 1; index >= first; index-- {
		if !entries[index].Type().IsRegular() || !strings.HasSuffix(entries[index].Name(), ".json") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(root, entries[index].Name()))
		if readErr != nil {
			unreadable = true
			continue
		}
		var record hookFailureRecord
		if json.Unmarshal(content, &record) != nil {
			unreadable = true
			continue
		}
		if record.Agent == agent {
			return record.OccurredAt + ":" + record.Operation, true
		}
	}
	if unreadable {
		return "unreadable", false
	}
	return "none", true
}

func summarizeAgentLineage(cfg AgentConfig, now, sourceTime time.Time, sourcePresent bool, summary agentLineageSummary) (string, string, bool) {
	if summary.Unreadable {
		return "unreadable", "unknown", false
	}
	if summary.LastComplete.IsZero() {
		if summary.Partial {
			return "partial-only", "unknown", false
		}
		return "none", "unknown", true
	}
	ingestion := summary.LastComplete.UTC().Format(time.RFC3339)
	if summary.Partial && summary.Latest.After(summary.LastComplete) {
		return ingestion + ":newer-partial", "unknown", false
	}
	if !sourcePresent || sourceTime.IsZero() || !sourceTime.After(summary.LastComplete) {
		return ingestion, "0s", true
	}
	lag := sourceTime.Sub(summary.LastComplete).Truncate(time.Second)
	return ingestion, lag.String(), lag <= cfg.staleLag() && !sourceTime.After(now.Add(time.Minute))
}
