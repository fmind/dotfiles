package dot

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

// Level names, ordered shallow to deep within each target. They are target-specific on
// purpose: `--go=module` and `--docker=system` say what is actually removed, which a
// generic `--deep` never could.
const (
	levelSessions = "sessions"
	levelBuild    = "build"
	levelSystem   = "system"
	levelModule   = "module"
	levelCache    = "cache"
	levelAll      = "all"
	levelConfigs  = "configs"
	levelShallow  = "shallow"
	levelDeep     = "deep"

	sessionStoreArchive  = "archive"
	sessionStoreAgy      = "agy"
	sessionStoreClaude   = "claude"
	sessionStoreCodex    = "codex"
	sessionStoreGrok     = "grok"
	sessionStoreOpenCode = "opencode"
	sessionStoreCopilot  = "copilot"
)

// pruneAllFlag selects every target at once; it is not a target itself.
const pruneAllFlag = "all"

// pruneLevelBare is the value urfave's parser assigns to a flag given without a value
// (`--docker`). It resolves to the target's shallowest level.
const pruneLevelBare = "true"

// PruneSessionStore is one agent session store and how long its files are kept.
// Retention is per store because they are not equivalent: the raw per-agent stores are
// disposable once `dot agent session` has produced an exact verified successor, while
// ~/.agents/sessions is the archive that survives them. keep_days 0 bypasses age only;
// raw-source proof is still mandatory.
type PruneSessionStore struct {
	Path     string `yaml:"path"`
	Source   string `yaml:"source,omitempty"`
	KeepDays int    `yaml:"keep_days"`
}

// PruneAgentsConfig is the policy for the agents target.
type PruneAgentsConfig struct {
	Sessions []PruneSessionStore `yaml:"sessions"`
	Keep     []string            `yaml:"keep"`
}

// PruneTargetConfig is the policy shared by the tool targets: the depth a bare flag
// selects, and the directories a file-based target removes. Targets that only call an
// external tool (docker, go, python) leave paths empty.
type PruneTargetConfig struct {
	Level string `yaml:"level"`
	// omitempty so `dot config show` does not advertise an empty list for the targets
	// that only call an external tool and own no directory.
	Paths []string `yaml:"paths,omitempty"`
}

// PruneConfig configures `dot prune`: one section per target, so every default the
// command applies is visible in `dot config show` and can be changed without a rebuild.
type PruneConfig struct {
	Agents PruneAgentsConfig `yaml:"agents"`
	Docker PruneTargetConfig `yaml:"docker"`
	Go     PruneTargetConfig `yaml:"go"`
	Python PruneTargetConfig `yaml:"python"`
	Node   PruneTargetConfig `yaml:"node"`
	Mise   PruneTargetConfig `yaml:"mise"`
	Tools  PruneTargetConfig `yaml:"tools"`
}

func defaultPruneConfig() PruneConfig {
	return PruneConfig{
		Agents: PruneAgentsConfig{
			// Derived from the canonical agent table so retention can never target a
			// directory ingestion no longer reads. Raw per-agent stores are disposable
			// after exact successor verification and are by far the biggest consumers of
			// disk; the SQLite-backed stores mix many lineages in one file, so they are
			// inventoried and retained until a source-specific compactor can delete rows
			// safely. The normalized archive is kept far longer than any raw source.
			Sessions: defaultRawSessionStores(),
			// Agent long-term memory lives inside the session stores (e.g.
			// ~/.claude/projects/<project>/memory/*.md). It is hand-curated state, not a
			// disposable session log, so any path segment named here survives pruning no
			// matter how old it is.
			Keep: []string{"memory", "memory.jsonl", "MEMORY.md"},
		},
		// Level is the depth a bare flag (and `--all`) selects; set one to its deeper
		// value to make that the norm on a machine, e.g. docker.level: system where no
		// local k3d cluster ever runs.
		Docker: PruneTargetConfig{Level: levelBuild},
		Go:     PruneTargetConfig{Level: levelBuild},
		Python: PruneTargetConfig{Level: levelCache},
		Node:   PruneTargetConfig{Level: levelCache, Paths: []string{"~/.npm/_npx"}},
		Mise: PruneTargetConfig{Level: levelCache, Paths: []string{
			"~/.local/share/mise/downloads",
			"~/.local/share/mise/http-tarballs",
		}},
		Tools: PruneTargetConfig{Level: levelCache, Paths: []string{
			cachePath("trivy"),
			cachePath("helm"),
		}},
	}
}

// cachePath renders a cache subdirectory as a ~-relative path when it sits under $HOME,
// so the default config reads the same on Linux (~/.cache) and macOS (~/Library/Caches).
func cachePath(name string) string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join("~", ".cache", name)
	}
	path := filepath.Join(dir, name)
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	relative, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return path
	}
	return filepath.Join("~", relative)
}

// PruneOptions describes one prune invocation: which targets run, at which level. Days
// overrides the retention of every configured session store when it is zero or more;
// pruneDaysFromConfig keeps each store's own keep_days.
type PruneOptions struct {
	Targets map[string]string
	Days    int
	DryRun  bool
}

// pruneDaysFromConfig marks an unset --days, so 0 stays meaningful (delete everything).
const pruneDaysFromConfig = -1

// pruneLevelValue backs the target flags. It reports IsBoolFlag so urfave's parser
// accepts both the bare form (`--docker`, which arrives as "true") and an explicit depth
// (`--docker=system`); a plain StringFlag would reject the bare form with "flag needs an
// argument", and a BoolFlag could not carry the depth.
type pruneLevelValue struct{ destination *string }

// Create implements cli.ValueCreator.
func (pruneLevelValue) Create(value string, destination *string, _ cli.NoConfig) cli.Value {
	*destination = value
	return &pruneLevelValue{destination: destination}
}

// ToString implements cli.ValueCreator.
func (pruneLevelValue) ToString(value string) string { return value }

func (p *pruneLevelValue) Set(value string) error {
	*p.destination = value
	return nil
}

func (p *pruneLevelValue) Get() any { return p.String() }

func (p *pruneLevelValue) String() string {
	if p.destination == nil {
		return ""
	}
	return *p.destination
}

// IsBoolFlag makes the value optional on the command line.
func (p *pruneLevelValue) IsBoolFlag() bool { return true }

// pruneLevelFlag is a string flag whose value may be omitted.
type pruneLevelFlag struct {
	cli.FlagBase[string, cli.NoConfig, pruneLevelValue]
	levels []string
}

// String renders the help line. urfave would otherwise print the generic `--docker
// string` placeholder of a string flag, which wrongly implies the value is mandatory;
// spelling the depths out is what makes the optional value discoverable.
func (f *pruneLevelFlag) String() string {
	return fmt.Sprintf("--%s [=%s], -%s\t%s", f.Name, strings.Join(f.levels, "|"), f.Aliases[0], f.Usage)
}

// pruneTarget is one prune domain. levels are ordered shallow to deep: a bare flag
// selects the configured level (the shallowest unless prune.<target>.level says
// otherwise), `--all=deep` selects the last, and anything else is named explicitly.
type pruneTarget struct {
	prune    func(ctx context.Context, run *pruneRun, level string) error
	settings func(cfg *PruneConfig) *PruneTargetConfig
	name     string
	alias    string
	usage    string
	summary  string
	levels   []string
}

func (t pruneTarget) shallowest() string { return t.levels[0] }

func (t pruneTarget) deepest() string { return t.levels[len(t.levels)-1] }

// configured returns the depth a bare flag selects for this target, rejecting a
// misconfigured level rather than silently falling back to the built-in default.
func (t pruneTarget) configured(cfg *PruneConfig) (string, error) {
	if t.settings == nil {
		return t.shallowest(), nil
	}
	level := t.settings(cfg).Level
	if level == "" {
		return t.shallowest(), nil
	}
	if !slices.Contains(t.levels, level) {
		return "", fmt.Errorf("invalid prune.%s.level %q (expected one of: %s)", t.name, level, strings.Join(t.levels, ", "))
	}
	return level, nil
}

// resolve maps a raw flag value to a concrete level.
func (t pruneTarget) resolve(value, configured string) string {
	if value == pruneLevelBare {
		return configured
	}
	return value
}

// pruneTargets is the single source of truth for the command: flags, help text, and
// execution order are all derived from it, so a new domain is one entry and one function.
var pruneTargets = []pruneTarget{
	{
		name:    "agents",
		alias:   "a",
		usage:   "Delete expired session logs from every store in prune.agents.sessions",
		levels:  []string{levelSessions},
		prune:   pruneAgentSessions,
		summary: "expired agent session logs",
	},
	{
		name:     "docker",
		alias:    "d",
		usage:    "Prune the Docker build cache; =system also prunes stopped containers, networks, and dangling images",
		levels:   []string{levelBuild, levelSystem},
		prune:    pruneDocker,
		settings: func(cfg *PruneConfig) *PruneTargetConfig { return &cfg.Docker },
		summary:  "Docker build cache",
	},
	{
		name:     "go",
		alias:    "g",
		usage:    "Clean the Go build and test caches; =module also clears the module cache",
		levels:   []string{levelBuild, levelModule},
		prune:    pruneGo,
		settings: func(cfg *PruneConfig) *PruneTargetConfig { return &cfg.Go },
		summary:  "Go build and test caches",
	},
	{
		name:     "python",
		alias:    "p",
		usage:    "Prune unused uv cache entries; =all wipes the uv cache and purges pip",
		levels:   []string{levelCache, levelAll},
		prune:    prunePython,
		settings: func(cfg *PruneConfig) *PruneTargetConfig { return &cfg.Python },
		summary:  "uv and pip caches",
	},
	{
		name:     "node",
		alias:    "n",
		usage:    "Remove the npx cache; =all also cleans the npm cache",
		levels:   []string{levelCache, levelAll},
		prune:    pruneNode,
		settings: func(cfg *PruneConfig) *PruneTargetConfig { return &cfg.Node },
		summary:  "npx and npm caches",
	},
	{
		name:     "mise",
		alias:    "m",
		usage:    "Prune unused tool versions, the cache, and downloads; =configs also prunes untracked config links",
		levels:   []string{levelCache, levelConfigs},
		prune:    pruneMise,
		settings: func(cfg *PruneConfig) *PruneTargetConfig { return &cfg.Mise },
		summary:  "mise versions, cache, and downloads",
	},
	{
		name:     "tools",
		alias:    "t",
		usage:    "Clear the Trivy, Helm, dprint, and golangci-lint caches",
		levels:   []string{levelCache},
		prune:    pruneTools,
		settings: func(cfg *PruneConfig) *PruneTargetConfig { return &cfg.Tools },
		summary:  "Trivy, Helm, dprint, and golangci-lint caches",
	},
}

// NewPruneCmd constructs the top-level prune command.
func NewPruneCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "prune",
		Aliases: []string{"x"},
		Usage:   "Reclaim disk space from agent session logs and development caches",
		Description: "Targets compose freely: `dot prune --agents --go` runs both, and each target " +
			"accepts an optional depth (`--docker=system`). Nothing runs unless a target is selected.",
		Flags: pruneFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := resolvePruneTargets(cmd, &state.Config.Prune)
			if err != nil {
				return err
			}
			return RunPrune(ctx, state, PruneOptions{
				Targets: targets,
				Days:    int(cmd.Int("days")),
				DryRun:  cmd.Bool("dry-run"),
			})
		},
	}
}

// pruneFlags derives one optional-value flag per target from the table, plus the
// aggregate and modifier flags.
func pruneFlags() []cli.Flag {
	flags := make([]cli.Flag, 0, len(pruneTargets)+3)
	for _, target := range pruneTargets {
		flags = append(flags, newPruneLevelFlag(target.name, target.alias, target.usage, target.levels))
	}
	return append(flags,
		newPruneLevelFlag(pruneAllFlag, "A", "Select every target at its configured depth; =deep selects the deepest depth of each",
			[]string{levelShallow, levelDeep}),
		&cli.IntFlag{
			Name:        "days",
			Aliases:     []string{"D"},
			Value:       pruneDaysFromConfig,
			Usage:       "Override the retention of every agent session store (0 empties them)",
			DefaultText: "each store's keep_days from the config file",
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"N"},
			Usage:   "Report what would be removed without deleting anything",
		},
	)
}

func newPruneLevelFlag(name, alias, usage string, levels []string) *pruneLevelFlag {
	return &pruneLevelFlag{
		levels: levels,
		FlagBase: cli.FlagBase[string, cli.NoConfig, pruneLevelValue]{
			Name:    name,
			Aliases: []string{alias},
			Usage:   usage,
			// An unselected target has no meaningful default to advertise.
			HideDefault: true,
			Validator: func(value string) error {
				if value == "" || value == pruneLevelBare || slices.Contains(levels, value) {
					return nil
				}
				return fmt.Errorf("invalid level %q for --%s (expected one of: %s)", value, name, strings.Join(levels, ", "))
			},
		},
	}
}

// resolvePruneTargets maps the parsed flags to the level each target runs at, using
// prune.<target>.level wherever the command line did not name a depth. An explicit target
// flag always wins over --all, so `--all --go=module` is a legal way to deepen one target.
func resolvePruneTargets(cmd *cli.Command, cfg *PruneConfig) (map[string]string, error) {
	deepAll := cmd.String(pruneAllFlag) == levelDeep
	selectAll := cmd.String(pruneAllFlag) != ""

	targets := make(map[string]string, len(pruneTargets))
	for _, target := range pruneTargets {
		value := cmd.String(target.name)
		if value == "" && !selectAll {
			continue
		}

		configured, err := target.configured(cfg)
		if err != nil {
			return nil, err
		}
		switch {
		case value != "":
			targets[target.name] = target.resolve(value, configured)
		case deepAll:
			targets[target.name] = target.deepest()
		default:
			targets[target.name] = configured
		}
	}
	return targets, nil
}

// RunPrune executes every selected target in table order. A failing target never stops
// the others: everything reclaimable is reclaimed, and the failures are reported together
// at the end rather than hidden behind the first error.
func RunPrune(ctx context.Context, state *GlobalState, opts PruneOptions) error {
	if len(opts.Targets) == 0 {
		printPruneTargets(state)
		return nil
	}
	if opts.Days < pruneDaysFromConfig {
		return errors.New("retention days cannot be negative")
	}

	run := &pruneRun{state: state, days: opts.Days, dryRun: opts.DryRun}
	title := "Prune"
	if opts.DryRun {
		title = "Prune (dry run)"
	}
	section(state.Stdout, title)

	var errs []error
	for _, target := range pruneTargets {
		level, selected := opts.Targets[target.name]
		if !selected {
			continue
		}
		state.Logger.Debug("Pruning target", "target", target.name, "level", level, "dry_run", opts.DryRun)
		if err := target.prune(ctx, run, level); err != nil {
			_, _ = fmt.Fprintf(state.Stdout, "%s %s: %v\n", failIcon, target.name, err)
			errs = append(errs, err)
		}
	}

	// Only bytes dot deleted itself can be measured; docker, go, uv, npm, and mise report
	// their own reclaimed space on their own terms.
	verb := "Reclaimed"
	if opts.DryRun {
		verb = "Would reclaim"
	}
	_, _ = fmt.Fprintf(state.Stdout, "%s %s %s from files dot removed directly.\n", passIcon, verb, humanBytes(run.freed))
	return errors.Join(errs...)
}

// printPruneTargets explains the available targets when none was selected, so an
// argument-less `dot prune` is a no-op that teaches instead of a destructive surprise.
func printPruneTargets(state *GlobalState) {
	section(state.Stdout, "Prune")
	_, _ = fmt.Fprintf(state.Stdout, "%s No target selected; nothing was pruned. Available targets:\n", skipIcon)
	for _, target := range pruneTargets {
		levels := target.levels[0]
		if len(target.levels) > 1 {
			levels = strings.Join(target.levels, "|")
		}
		_, _ = fmt.Fprintf(state.Stdout, "  -%s, --%-8s [=%s] %s\n", target.alias, target.name, levels, target.summary)
	}
	_, _ = fmt.Fprintf(state.Stdout, "  -A, --%-8s [=%s] every target above\n", pruneAllFlag, levelShallow+"|"+levelDeep)
	_, _ = fmt.Fprintln(state.Stdout, dim("Run `dot prune --help` for the full flag reference."))
}

// pruneRun carries the state shared by every target function within one invocation.
type pruneRun struct {
	state *GlobalState
	// days is the --days override, or pruneDaysFromConfig when each session store
	// keeps its own retention.
	days   int
	dryRun bool
	freed  int64
}

// retention resolves how many days of history a session store keeps.
func (r *pruneRun) retention(dir PruneSessionStore) (int, error) {
	days := dir.KeepDays
	if r.days != pruneDaysFromConfig {
		days = r.days
	}
	if days < 0 {
		return 0, fmt.Errorf("invalid retention for %s: keep_days cannot be negative", dir.Path)
	}
	return days, nil
}

func (r *pruneRun) passf(target, format string, args ...any) {
	_, _ = fmt.Fprintf(r.state.Stdout, "%s %s: %s\n", passIcon, target, fmt.Sprintf(format, args...))
}

func (r *pruneRun) skipf(target, format string, args ...any) {
	_, _ = fmt.Fprintf(r.state.Stdout, "%s %s: %s\n", skipIcon, target, fmt.Sprintf(format, args...))
}

// exec runs an external tool for a target. A tool that is not installed is reported as
// skipped rather than failing the prune: the shell scripts this replaces guarded every
// call with `command -v`, and no workstation has all of them.
func (r *pruneRun) exec(ctx context.Context, target, summary, name string, args ...string) error {
	if _, err := r.state.Runner.LookPath(name); err != nil {
		r.skipf(target, "%s is not installed", name)
		return nil
	}
	if r.dryRun {
		r.skipf(target, "would run %s %s", name, strings.Join(args, " "))
		return nil
	}
	if _, err := r.state.Runner.Run(ctx, "", nil, name, args...); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	r.passf(target, "%s", summary)
	return nil
}

// removeTree deletes path and everything under it.
func (r *pruneRun) removeTree(target, path string) error {
	return r.removePaths(target, path, []string{path})
}

// removeContents empties path but keeps the directory itself, for tools that expect
// their cache directory to exist and only recreate the entries inside it.
func (r *pruneRun) removeContents(target, path string) error {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, filepath.Join(path, entry.Name()))
	}
	return r.removePaths(target, path, paths)
}

// removePaths measures, then deletes, the given paths, reporting them under one label.
func (r *pruneRun) removePaths(target, label string, paths []string) error {
	var total int64
	present := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Lstat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("failed to inspect %s: %w", path, err)
		}
		size, err := dirBytes(path)
		if err != nil {
			return fmt.Errorf("failed to measure %s: %w", path, err)
		}
		total += size
		present = append(present, path)
	}

	if len(present) == 0 {
		r.skipf(target, "nothing to remove in %s", label)
		return nil
	}
	r.freed += total
	if r.dryRun {
		r.skipf(target, "would remove %s (%s)", label, humanBytes(total))
		return nil
	}
	for _, path := range present {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}
	r.passf(target, "removed %s (%s)", label, humanBytes(total))
	return nil
}

// pruneSessions deletes files under root last modified before cutoff, then removes the
// directories emptied along the way. It reports the number of files and bytes affected so
// the caller can report each session store separately.
func (r *pruneRun) pruneSessions(root string, cutoff time.Time, keep []string) (int, int64, error) {
	var (
		emptied []string
		files   int
		total   int64
	)

	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == root {
				return nil
			}
			if slices.Contains(keep, entry.Name()) {
				return filepath.SkipDir
			}
			emptied = append(emptied, path)
			return nil
		}
		if slices.Contains(keep, entry.Name()) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			// The owning agent may rotate a session log mid-walk; a file that
			// vanished needs no pruning.
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if !info.ModTime().Before(cutoff) {
			return nil
		}

		files++
		total += info.Size()
		if r.dryRun {
			return nil
		}
		//nolint:gosec // G122: the walked root is the owner-only configured session store, not an attacker-writable path
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove session log %s: %w", path, err)
		}
		return nil
	})
	if errors.Is(walkErr, os.ErrNotExist) {
		return 0, 0, nil
	}
	if walkErr != nil {
		return 0, 0, fmt.Errorf("failed to scan %s: %w", root, walkErr)
	}

	r.freed += total
	if !r.dryRun {
		if err := removeEmptyDirs(emptied); err != nil {
			return files, total, err
		}
	}
	return files, total, nil
}

// pruneAgentSessions expires every configured session store. Raw per-agent sources are
// removed only after an exact complete normalized successor is verified; the normalized
// archive and backwards-compatible custom stores retain their age-based policy.
func pruneAgentSessions(_ context.Context, run *pruneRun, _ string) error {
	cfg := run.state.Config.Prune.Agents
	now := time.Now()

	var errs []error
	pruned := 0
	for _, dir := range cfg.Sessions {
		source := sessionStoreSource(dir)
		if dir.Source != "" && !isKnownSessionStoreSource(source) {
			errs = append(errs, fmt.Errorf("invalid session source %q for %s", dir.Source, dir.Path))
			continue
		}
		path := ExpandPath(dir.Path)
		if _, err := os.Stat(path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				errs = append(errs, fmt.Errorf("failed to inspect %s: %w", dir.Path, err))
			}
			continue
		}

		days, err := run.retention(dir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cutoff := now.AddDate(0, 0, -days)
		var files int
		var total int64
		if source == sessionStoreAgy || source == sessionStoreClaude || source == sessionStoreCodex || source == sessionStoreGrok || source == sessionStoreOpenCode || source == sessionStoreCopilot {
			files, total, err = run.pruneRawSessions(path, source, cutoff, now, cfg.Keep)
		} else {
			files, total, err = run.pruneSessions(path, cutoff, cfg.Keep)
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}

		pruned++
		run.state.Logger.Debug("Pruned agent sessions", "dir", path, "source", source, "days", days, "files", files, "bytes", total)
		verb, report := "deleted", run.passf
		if run.dryRun {
			verb, report = "would delete", run.skipf
		}
		report("agents", "%s %d file(s) older than %d days in %s (%s)", verb, files, days, dir.Path, humanBytes(total))
	}

	if pruned == 0 && len(errs) == 0 {
		run.skipf("agents", "no session stores found")
	}
	return errors.Join(errs...)
}

// pruneDocker prunes the build cache and, at the system level, the rest of the daemon's
// reclaimable resources.
func pruneDocker(ctx context.Context, run *pruneRun, level string) error {
	if _, err := run.state.Runner.LookPath("docker"); err != nil {
		run.skipf("docker", "docker is not installed")
		return nil
	}
	if _, err := run.state.Runner.Run(ctx, "", nil, "docker", "info"); err != nil {
		run.skipf("docker", "docker daemon is not running")
		return nil
	}

	if err := run.exec(ctx, "docker", "pruned the build cache", "docker", "builder", "prune", "-af"); err != nil {
		return err
	}
	if level != levelSystem {
		return nil
	}
	// `docker system prune` removes stopped containers, and a stopped local k3d cluster
	// *is* a set of stopped containers — deleting them destroys the cluster. That is why
	// this only runs when the deeper level is asked for by name.
	return run.exec(ctx, "docker", "pruned stopped containers, networks, and dangling images",
		"docker", "system", "prune", "-f")
}

// pruneGo cleans the Go caches.
func pruneGo(ctx context.Context, run *pruneRun, level string) error {
	if err := run.exec(ctx, "go", "cleaned the build and test caches", "go", "clean", "-cache", "-testcache"); err != nil {
		return err
	}
	if level != levelModule {
		return nil
	}
	// Refilling the module cache re-downloads every dependency of every project, so it is
	// deliberately not part of the default level.
	return run.exec(ctx, "go", "cleaned the module cache", "go", "clean", "-modcache")
}

// prunePython prunes the uv cache and, at the deeper level, wipes it along with pip's.
func prunePython(ctx context.Context, run *pruneRun, level string) error {
	if level != levelAll {
		return run.exec(ctx, "python", "pruned unused uv cache entries", "uv", "cache", "prune")
	}
	if err := run.exec(ctx, "python", "cleaned the uv cache", "uv", "cache", "clean"); err != nil {
		return err
	}
	return run.exec(ctx, "python", "purged the pip cache", "pip", "cache", "purge")
}

// pruneNode removes the configured npx cache and, at the deeper level, the npm cache.
func pruneNode(ctx context.Context, run *pruneRun, level string) error {
	for _, path := range run.state.Config.Prune.Node.Paths {
		if err := run.removeTree("node", ExpandPath(path)); err != nil {
			return err
		}
	}
	if level != levelAll {
		return nil
	}
	return run.exec(ctx, "node", "cleaned the npm cache", "npm", "cache", "clean", "--force")
}

// pruneMise prunes unused tool versions, the cache, and the download staging areas.
func pruneMise(ctx context.Context, run *pruneRun, level string) error {
	if _, err := run.state.Runner.LookPath("mise"); err != nil {
		run.skipf("mise", "mise is not installed")
		return nil
	}
	if err := run.exec(ctx, "mise", "pruned unused tool versions", "mise", "prune", "-y"); err != nil {
		return err
	}
	if err := run.exec(ctx, "mise", "cleared the cache", "mise", "cache", "clear"); err != nil {
		return err
	}

	// Downloads and tarballs are staging areas for installs that already succeeded; mise
	// recreates both directories on demand, so only their contents go.
	for _, path := range run.state.Config.Prune.Mise.Paths {
		if err := run.removeContents("mise", ExpandPath(path)); err != nil {
			return err
		}
	}
	if level != levelConfigs {
		return nil
	}
	return run.exec(ctx, "mise", "pruned untracked config links", "mise", "prune", "--configs", "-y")
}

// pruneTools clears the caches of the linters and scanners.
func pruneTools(ctx context.Context, run *pruneRun, _ string) error {
	var errs []error
	// Trivy's database and Helm's chart cache are re-downloaded on demand and neither
	// tool offers a clean subcommand that works when the binary itself is absent, so the
	// configured directories are removed directly.
	for _, path := range run.state.Config.Prune.Tools.Paths {
		if err := run.removeTree("tools", ExpandPath(path)); err != nil {
			errs = append(errs, err)
		}
	}
	if err := run.exec(ctx, "tools", "cleared the dprint cache", "dprint", "clear-cache"); err != nil {
		errs = append(errs, err)
	}
	if err := run.exec(ctx, "tools", "cleared the golangci-lint cache", "golangci-lint", "cache", "clean"); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// dirBytes sums the apparent size of every regular file at or under path.
func dirBytes(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if info.Mode().IsRegular() {
			total += info.Size()
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	return total, err
}

// removeEmptyDirs deletes the given directories that are empty, deepest first, so a
// parent left empty by its children goes in the same pass. It expects dirs in the lexical
// pre-order that filepath.WalkDir produces.
func removeEmptyDirs(dirs []string) error {
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to inspect directory %s: %w", dirs[i], err)
		}
		if len(entries) > 0 {
			continue
		}
		if err := os.Remove(dirs[i]); err != nil {
			return fmt.Errorf("failed to remove empty directory %s: %w", dirs[i], err)
		}
	}
	return nil
}

// humanBytes renders a byte count with binary units, e.g. "42.1 MiB".
func humanBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
