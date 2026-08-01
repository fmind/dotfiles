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
	doctorScanLimit = 4096
	doctorStaleLag  = 24 * time.Hour
)

type AgentDoctorOptions struct {
	Fix    bool
	DryRun bool
}

type agentIntegration struct {
	Agent         string
	PersonaPath   string
	SkillsPath    string
	HookPath      string
	SourcePath    string
	HookCommands  []string
	ProbeNames    []string
	PersonaInFile bool
	HookJSON      bool
	Notifications bool
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
			&cli.BoolFlag{Name: "fix", Usage: "Apply the managed agent integration targets with chezmoi"},
			&cli.BoolFlag{Name: "dry-run", Usage: "Preview --fix without changing deployed files"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunAgentDoctor(ctx, state, AgentDoctorOptions{Fix: cmd.Bool("fix"), DryRun: cmd.Bool("dry-run")})
		},
	}
}

func agentIntegrations(home string) []agentIntegration {
	sharedPersona := filepath.Join(home, ".agents", "AGENTS.md")
	sharedSkills := filepath.Join(home, ".agents", "skills")
	return []agentIntegration{
		{Agent: sessionStoreAgy, PersonaPath: filepath.Join(home, ".gemini", "GEMINI.md"), SkillsPath: filepath.Join(home, ".gemini", "config", "skills"), HookPath: filepath.Join(home, ".gemini", "config", "hooks.json"), SourcePath: filepath.Join(home, ".gemini", "antigravity-cli", "brain"), HookCommands: []string{"dot agent hook session agy", "dot agent hook notify agy stop"}, ProbeNames: []string{"dot", "agy"}, HookJSON: true, Notifications: true},
		{Agent: sessionStoreClaude, PersonaPath: filepath.Join(home, ".claude", "CLAUDE.md"), SkillsPath: filepath.Join(home, ".claude", "skills"), HookPath: filepath.Join(home, ".claude", "settings.json"), SourcePath: filepath.Join(home, ".claude", "projects"), HookCommands: []string{"dot agent hook session claude", "dot agent hook notify claude"}, ProbeNames: []string{"dot", "claude"}, HookJSON: true, Notifications: true},
		{Agent: sessionStoreCodex, PersonaPath: filepath.Join(home, ".codex", "AGENTS.md"), SkillsPath: sharedSkills, HookPath: filepath.Join(home, ".codex", "config.toml"), SourcePath: filepath.Join(home, ".codex", "sessions"), HookCommands: []string{"dot agent hook session codex", "dot agent hook notify codex stop"}, ProbeNames: []string{"dot", "codex"}, Notifications: true},
		{Agent: sessionStoreOpenCode, PersonaPath: filepath.Join(home, ".config", "opencode", "opencode.json"), SkillsPath: sharedSkills, HookPath: filepath.Join(home, ".config", "opencode", "plugins", "session-log.ts"), SourcePath: filepath.Join(home, ".local", "share", "opencode", "opencode.db"), HookCommands: []string{"dot agent hook session opencode"}, ProbeNames: []string{"dot", "opencode", "sqlite3"}, PersonaInFile: true},
		{Agent: sessionStoreCopilot, PersonaPath: filepath.Join(home, ".copilot", "copilot-instructions.md"), SkillsPath: sharedSkills, HookPath: filepath.Join(home, ".copilot", "hooks", "session-log.json"), SourcePath: filepath.Join(home, ".copilot", "session-store.db"), HookCommands: []string{"dot agent hook copilot-session-end"}, ProbeNames: []string{"dot", "copilot", "sqlite3"}, HookJSON: true},
		// These canonical targets are checked through every resolved path above.
		{Agent: "canonical", PersonaPath: sharedPersona, SkillsPath: sharedSkills},
	}
}

func doctorRepairTargets(home string) []string {
	return []string{
		filepath.Join(home, ".agents", "AGENTS.md"), filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "CLAUDE.md"), filepath.Join(home, ".claude", "settings.json"), filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".codex", "AGENTS.md"), filepath.Join(home, ".codex", "config.toml"),
		filepath.Join(home, ".gemini", "GEMINI.md"), filepath.Join(home, ".gemini", "config", "hooks.json"), filepath.Join(home, ".gemini", "config", "skills"),
		filepath.Join(home, ".config", "opencode", "opencode.json"), filepath.Join(home, ".config", "opencode", "plugins", "session-log.ts"),
		filepath.Join(home, ".copilot", "copilot-instructions.md"), filepath.Join(home, ".copilot", "hooks", "session-log.json"),
	}
}

func repairAgentIntegrations(ctx context.Context, state *GlobalState, home string, dryRun bool) error {
	args := []string{"apply"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--force")
	args = append(args, doctorRepairTargets(home)...)
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
		if err := repairAgentIntegrations(ctx, state, home, options.DryRun); err != nil {
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
	integrations := agentIntegrations(home)
	canonical := integrations[len(integrations)-1]
	results := make([]agentDoctorResult, 0, len(integrations)-1)
	registry := CapabilityProbeRegistry()
	for _, integration := range integrations[:len(integrations)-1] {
		discovery, discoveryOK := checkAgentDiscovery(integration, canonical)
		hooks, hooksOK := checkAgentHooks(integration)
		probeNames := append([]string(nil), integration.ProbeNames...)
		if integration.Notifications {
			if notifier := doctorNotifierProbe(state.Runner); notifier != "" {
				probeNames = append(probeNames, notifier)
			} else {
				hooks, hooksOK = "notification-unavailable", false
			}
		}
		probeResults, probesOK := runToolProbes(ctx, state.Runner, probeNames, registry)
		tools := summarizeDoctorProbes(probeResults)
		source, sourceTime, sourcePresent, sourceOK, sourceOmitted := inspectAgentSource(integration)
		lineage := inspectAgentLineage(home, integration.Agent)
		failure, failureOK := inspectLastHookFailure(home, integration.Agent)
		ingestion, lag, lineageOK := summarizeAgentLineage(now, sourceTime, sourcePresent, lineage)
		results = append(results, agentDoctorResult{
			Agent: integration.Agent, Discovery: discovery, Hooks: hooks, Tools: tools, Source: source,
			LastIngestion: ingestion, LastFailure: failure, ArchiveLag: lag, Omitted: sourceOmitted + lineage.Omitted,
			Healthy: discoveryOK && hooksOK && probesOK && sourceOK && lineageOK && failureOK,
		})
	}
	return results
}

func sameResolvedPath(path, target string) bool {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	want, err := filepath.EvalSymlinks(target)
	return err == nil && resolved == want
}

func checkAgentDiscovery(integration, canonical agentIntegration) (string, bool) {
	if integration.PersonaInFile {
		content, err := os.ReadFile(integration.PersonaPath) //nolint:gosec // fixed local integration path
		if err != nil || !json.Valid(content) || !strings.Contains(string(content), "~/.agents/AGENTS.md") {
			return "persona-broken", false
		}
	} else if !sameResolvedPath(integration.PersonaPath, canonical.PersonaPath) {
		return "persona-broken", false
	}
	if integration.SkillsPath == canonical.SkillsPath {
		if info, err := os.Stat(canonical.SkillsPath); err != nil || !info.IsDir() {
			return "skills-missing", false
		}
	} else if !sameResolvedPath(integration.SkillsPath, canonical.SkillsPath) {
		return "skills-broken", false
	}
	return "healthy", true
}

func checkAgentHooks(integration agentIntegration) (string, bool) {
	if integration.HookPath == "" {
		return "sync-only", true
	}
	content, err := os.ReadFile(integration.HookPath) //nolint:gosec // fixed local integration path
	if err != nil {
		return "missing", false
	}
	if integration.HookJSON && !json.Valid(content) {
		return "malformed", false
	}
	if integration.Agent == sessionStoreCopilot {
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
	for _, command := range integration.HookCommands {
		if !strings.Contains(string(content), command) {
			return "command-mismatch", false
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

func inspectAgentSource(integration agentIntegration) (string, time.Time, bool, bool, int) {
	info, err := os.Lstat(integration.SourcePath)
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
	walkErr := filepath.WalkDir(integration.SourcePath, func(path string, entry fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if seen >= doctorScanLimit {
			omitted = 1
			return fs.SkipAll
		}
		seen++
		if _, recognized := rawSessionIdentity(integration.SourcePath, path, integration.Agent); !recognized {
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

func inspectAgentLineage(home, agent string) agentLineageSummary {
	root := filepath.Join(home, ".agents", "sessions", sessionStoreVersion, agent)
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
		if seen >= doctorScanLimit {
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

func inspectLastHookFailure(home, agent string) (string, bool) {
	root := filepath.Join(home, ".agents", "hook-failures", hookFailureStoreVersion)
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return "none", true
	}
	if err != nil {
		return "unreadable", false
	}
	unreadable := false
	first := max(len(entries)-hookFailureLimit, 0)
	for index := len(entries) - 1; index >= first; index-- {
		if !entries[index].Type().IsRegular() || !strings.HasSuffix(entries[index].Name(), ".json") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(root, entries[index].Name())) //nolint:gosec // owner-only bounded spool
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

func summarizeAgentLineage(now, sourceTime time.Time, sourcePresent bool, summary agentLineageSummary) (string, string, bool) {
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
	return ingestion, lag.String(), lag <= doctorStaleLag && !sourceTime.After(now.Add(time.Minute))
}
