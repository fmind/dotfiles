package dot

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Shared discovery targets. Every agent CLI is pointed at the same persona and
// skill tree so one edit reaches all of them.
const (
	sharedPersonaPath = "~/.agents/AGENTS.md"
	sharedSkillsPath  = "~/.agents/skills"
)

const (
	// defaultRawSessionKeepDays bounds a raw agent transcript store. Raw stores are
	// disposable once ingestion has produced a verified successor.
	defaultRawSessionKeepDays = 7
	// defaultArchiveKeepDays bounds the normalized archive, which is the evidence the
	// workspace brain reads long after every raw source is gone.
	defaultArchiveKeepDays = 30
)

// agentDefinition is the single description of one agent CLI integration: where it
// keeps raw transcripts, how it is wired into the shared persona and skills, and
// which hook commands dot expects to find in its configuration. Session ingestion,
// retention, and `dot agent doctor` all read this table, so the path dot ingests
// from can never drift from the path dot prunes or health-checks.
type agentDefinition struct {
	// Agent is the canonical short name used in stores, flags, and hook arguments.
	Agent string
	// Alias is the 1-letter shortcut for subcommand invocation under agent session.
	Alias string
	// Source is the ~-relative root of the agent's raw transcript store: a directory
	// for the file-based agents, a database file for the SQLite-backed ones.
	Source string
	// PersonaPath is where the agent CLI looks for its instructions file.
	PersonaPath string
	// SkillsPath is where the agent CLI looks for skills; empty means it reads the
	// shared tree directly rather than through its own link.
	SkillsPath string
	// HookPath is the agent's own configuration file that must reference HookCommands;
	// empty means the agent has no hook surface and is reachable only through sync.
	HookPath string
	// HookCommands are the exact `dot ...` invocations expected inside HookPath.
	HookCommands []string
	// ProbeNames are the executables the integration needs at runtime.
	ProbeNames []string
	// PersonaInFile marks an agent that references the shared persona from inside a
	// structured config file instead of linking to it.
	PersonaInFile bool
	// HookJSON marks a hook file that must parse as JSON.
	HookJSON bool
	// Notifications marks an integration that also drives desktop notifications.
	Notifications bool
}

// agentDefinitions returns the canonical integration table in ingestion order.
func agentDefinitions() []agentDefinition {
	return []agentDefinition{
		{
			Agent:         sessionStoreAgy,
			Alias:         "a",
			Source:        "~/.gemini/antigravity-cli/brain",
			PersonaPath:   "~/.gemini/GEMINI.md",
			SkillsPath:    "~/.gemini/config/skills",
			HookPath:      "~/.gemini/config/hooks.json",
			HookCommands:  []string{"dot agent hook session agy", "dot agent hook notify agy stop"},
			ProbeNames:    []string{"dot", "agy"},
			HookJSON:      true,
			Notifications: true,
		},
		{
			Agent:         sessionStoreClaude,
			Alias:         "c",
			Source:        "~/.claude/projects",
			PersonaPath:   "~/.claude/CLAUDE.md",
			SkillsPath:    "~/.claude/skills",
			HookPath:      "~/.claude/settings.json",
			HookCommands:  []string{"dot agent hook session claude", "dot agent hook notify claude"},
			ProbeNames:    []string{"dot", "claude"},
			HookJSON:      true,
			Notifications: true,
		},
		{
			Agent:         sessionStoreCodex,
			Alias:         "x",
			Source:        "~/.codex/sessions",
			PersonaPath:   "~/.codex/AGENTS.md",
			HookPath:      "~/.codex/config.toml",
			HookCommands:  []string{"dot agent hook session codex", "dot agent hook notify codex stop"},
			ProbeNames:    []string{"dot", "codex"},
			Notifications: true,
		},
		{
			Agent:         sessionStoreOpenCode,
			Alias:         "o",
			Source:        "~/.local/share/opencode/opencode.db",
			PersonaPath:   "~/.config/opencode/opencode.json",
			HookPath:      "~/.config/opencode/plugins/session-log.ts",
			HookCommands:  []string{"dot agent hook session opencode"},
			ProbeNames:    []string{"dot", "opencode", "sqlite3"},
			PersonaInFile: true,
		},
		{
			Agent:        sessionStoreCopilot,
			Alias:        "p",
			Source:       "~/.copilot/session-store.db",
			PersonaPath:  "~/.copilot/copilot-instructions.md",
			HookPath:     "~/.copilot/hooks/session-log.json",
			HookCommands: []string{"dot agent hook copilot-session-end"},
			ProbeNames:   []string{"dot", "copilot", "sqlite3"},
			HookJSON:     true,
		},
	}
}

// resolvedSkillsPath returns the agent's own skills link, or the shared tree when
// the agent reads that tree directly.
func (d agentDefinition) resolvedSkillsPath() string {
	if d.SkillsPath == "" {
		return sharedSkillsPath
	}
	return d.SkillsPath
}

// AgentDoctorConfig bounds the read-only health scan so one oversized store cannot
// turn `dot agent doctor` into an unbounded walk.
type AgentDoctorConfig struct {
	// StaleLag is how far the raw source may run ahead of the last complete ingestion
	// before the integration is reported unhealthy.
	StaleLag Duration `yaml:"stale_lag"`
	// ScanLimit caps the files inspected per store; the overflow is reported as
	// `omitted` rather than silently dropped.
	ScanLimit int `yaml:"scan_limit"`
}

// HookFailureConfig bounds the owner-only spool that records why a hook failed. The
// spool stores metadata only: never a transcript body, never a raw session ID.
type HookFailureConfig struct {
	// Limit is how many failure records are retained before the oldest are trimmed.
	Limit int `yaml:"limit"`
	// DetailLimit caps the stored error text, in bytes.
	DetailLimit int `yaml:"detail_limit"`
}

// AgentConfig configures agent session ingestion and health checks. Source roots are
// settings because agent CLIs relocate their stores between releases: a move should
// be a config edit, not a rebuild.
type AgentConfig struct {
	// Sources overrides the raw transcript root of individual agents, keyed by agent
	// name. Unlisted agents keep their built-in default.
	Sources      map[string]string `yaml:"sources"`
	Doctor       AgentDoctorConfig `yaml:"doctor"`
	HookFailures HookFailureConfig `yaml:"hook_failures"`
}

func defaultAgentConfig() AgentConfig {
	definitions := agentDefinitions()
	sources := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		sources[definition.Agent] = definition.Source
	}
	return AgentConfig{
		Sources: sources,
		Doctor: AgentDoctorConfig{
			StaleLag:  Duration(defaultDoctorStaleLag),
			ScanLimit: defaultDoctorScanLimit,
		},
		HookFailures: HookFailureConfig{
			Limit:       defaultHookFailureLimit,
			DetailLimit: defaultHookFailureDetailLimit,
		},
	}
}

// scanLimit and staleLag resolve the bounded-scan knobs, falling back to the
// built-in defaults so an absent or zeroed key never disables the bound.
func (c AgentConfig) scanLimit() int {
	return positiveOr(c.Doctor.ScanLimit, defaultDoctorScanLimit)
}

func (c AgentConfig) staleLag() time.Duration {
	return positiveOr(c.Doctor.StaleLag.Duration(), defaultDoctorStaleLag)
}

func (c AgentConfig) hookFailureLimit() int {
	return positiveOr(c.HookFailures.Limit, defaultHookFailureLimit)
}

func (c AgentConfig) hookFailureDetailLimit() int {
	return positiveOr(c.HookFailures.DetailLimit, defaultHookFailureDetailLimit)
}

// SourceRoot returns the expanded raw transcript root for an agent, preferring the
// configured override and falling back to the built-in default.
func (c AgentConfig) SourceRoot(agent string) (string, error) {
	if path, ok := c.Sources[agent]; ok && strings.TrimSpace(path) != "" {
		return ExpandPath(path), nil
	}
	for _, definition := range agentDefinitions() {
		if definition.Agent == agent {
			return ExpandPath(definition.Source), nil
		}
	}
	return "", fmt.Errorf("unknown agent %q", agent)
}

// definitions returns the canonical table with every source root replaced by its
// configured value, so callers read one already-resolved view.
func (c AgentConfig) definitions() []agentDefinition {
	definitions := agentDefinitions()
	for index := range definitions {
		if path, ok := c.Sources[definitions[index].Agent]; ok && strings.TrimSpace(path) != "" {
			definitions[index].Source = path
		}
	}
	return definitions
}

// agentNameSet returns the canonical agent names as a lookup set.
func agentNameSet() map[string]bool {
	definitions := agentDefinitions()
	names := make(map[string]bool, len(definitions))
	for _, definition := range definitions {
		names[definition.Agent] = true
	}
	return names
}

// defaultRawSessionStores derives the retention entries from the canonical table so
// `dot prune` can never expire a directory that ingestion no longer reads.
func defaultRawSessionStores() []PruneSessionStore {
	definitions := agentDefinitions()
	stores := make([]PruneSessionStore, 0, len(definitions)+1)
	for _, definition := range definitions {
		stores = append(stores, PruneSessionStore{
			Path:     definition.Source,
			Source:   definition.Agent,
			KeepDays: defaultRawSessionKeepDays,
		})
	}
	// Stable order keeps `dot config show` output from churning between runs.
	slices.SortFunc(stores, func(left, right PruneSessionStore) int {
		return strings.Compare(left.Path, right.Path)
	})
	return append(stores, PruneSessionStore{
		Path:     sessionArchiveRoot,
		Source:   sessionStoreArchive,
		KeepDays: defaultArchiveKeepDays,
	})
}
