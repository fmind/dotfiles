package dot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	contextSchemaVersion = "1.0"
	defaultContextBytes  = 50000
	bytesPerToken        = 4
)

var defaultContextCollectors = []string{"instructions", "skills", "git", "tasks", "dependencies", "failures"}

type ContextConfig struct {
	Collectors            []string `yaml:"collectors"`
	InstructionFiles      []string `yaml:"instruction_files"`
	DependencyFiles       []string `yaml:"dependency_files"`
	FailureFiles          []string `yaml:"failure_files"`
	SensitivePathPatterns []string `yaml:"sensitive_path_patterns"`
	SensitiveEnvPatterns  []string `yaml:"sensitive_env_patterns"`
	MaxBytes              int      `yaml:"max_bytes"`
}

func defaultContextConfig() ContextConfig {
	return ContextConfig{
		Collectors:       append([]string(nil), defaultContextCollectors...),
		InstructionFiles: []string{"AGENTS.md"},
		DependencyFiles:  []string{"go.mod", "pyproject.toml", "package.json", "mise.toml"},
		MaxBytes:         defaultContextBytes,
	}
}

type ContextOptions struct {
	GeneratedAt time.Time
	Format      string
	Bytes       int
	Tokens      int
}

type contextSection struct {
	ID          string   `json:"id"`
	Fingerprint string   `json:"fingerprint"`
	ObservedAt  string   `json:"observed_at"`
	Error       string   `json:"error,omitempty"`
	Content     string   `json:"-"`
	Sources     []string `json:"sources"`
	Priority    int      `json:"priority"`
}

type contextUnit struct {
	section int
	line    int
}

type contextEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	GeneratedAt   string               `json:"generated_at"`
	Repository    string               `json:"repository"`
	SourceHead    string               `json:"source_head"`
	Sections      []contextJSONSection `json:"sections"`
	Adapters      []string             `json:"optional_adapters"`
	Budget        contextBudget        `json:"budget"`
}

type contextBudget struct {
	RequestedBytes  int `json:"requested_bytes,omitempty"`
	RequestedTokens int `json:"requested_tokens,omitempty"`
	EffectiveBytes  int `json:"effective_bytes"`
}

type contextJSONSection struct {
	ID           string   `json:"id"`
	Fingerprint  string   `json:"fingerprint"`
	ObservedAt   string   `json:"observed_at"`
	Status       string   `json:"status"`
	Error        string   `json:"error,omitempty"`
	Content      string   `json:"content,omitempty"`
	OmittedLines string   `json:"omitted_lines"`
	Sources      []string `json:"sources"`
	Priority     int      `json:"priority"`
	TotalBytes   int      `json:"total_bytes"`
	OmittedBytes int      `json:"omitted_bytes"`
}

func NewContextCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "context",
		Aliases: []string{"t"},
		Usage:   "Emit a bounded, redacted project context pack",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "bytes", Usage: "Maximum output bytes"},
			&cli.IntFlag{Name: "tokens", Usage: "Approximate token budget (four bytes per token)"},
			&cli.StringFlag{Name: "format", Value: "markdown", Usage: "Output format: markdown or json"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunContext(ctx, state, ContextOptions{Bytes: cmd.Int("bytes"), Tokens: cmd.Int("tokens"), Format: cmd.String("format"), GeneratedAt: time.Now().UTC()})
		},
	}
}

func RunContext(ctx context.Context, state *GlobalState, options ContextOptions) error {
	payload, err := BuildContext(ctx, state, options)
	if err != nil {
		return err
	}
	if scanErr := ScanContextPayload(ctx, state, payload); scanErr != nil {
		return scanErr
	}
	_, err = fmt.Fprint(state.Stdout, payload)
	return err
}

func BuildContext(ctx context.Context, state *GlobalState, options ContextOptions) (string, error) {
	budget, err := resolveContextBudget(state.Config.Context, options)
	if err != nil {
		return "", err
	}
	format := strings.ToLower(options.Format)
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" && format != "json" {
		return "", fmt.Errorf("unsupported context format %q", options.Format)
	}
	if options.GeneratedAt.IsZero() {
		return "", errors.New("context generation timestamp is required")
	}

	root, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to resolve project root: %w", err)
	}
	root = strings.TrimSpace(root)
	head, err := state.Runner.Run(ctx, root, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to fingerprint project HEAD: %w", err)
	}
	head = strings.TrimSpace(head)

	sections := collectContextSections(ctx, state, root, options.GeneratedAt)
	units, lines := contextUnits(sections)
	payload, err := packWithinBudget(budget.EffectiveBytes, len(units), func(selected []bool) []byte {
		chosen := make([]map[int]bool, len(sections))
		for i := range chosen {
			chosen[i] = make(map[int]bool)
		}
		for i, include := range selected {
			if include {
				unit := units[i]
				chosen[unit.section][unit.line] = true
			}
		}
		if format == "json" {
			return renderContextJSON(filepath.Base(root), head, options.GeneratedAt, budget, sections, lines, chosen)
		}
		return renderContextMarkdown(filepath.Base(root), head, options.GeneratedAt, budget, sections, lines, chosen)
	})
	if err != nil {
		return "", fmt.Errorf("failed to pack project context: %w", err)
	}
	return string(payload), nil
}

func resolveContextBudget(config ContextConfig, options ContextOptions) (contextBudget, error) {
	maxInt := int(^uint(0) >> 1)
	if options.Bytes < 0 || options.Tokens < 0 || options.Bytes > 0 && options.Tokens > 0 || options.Tokens > maxInt/bytesPerToken {
		return contextBudget{}, errors.New("--bytes and --tokens must be positive and mutually exclusive")
	}
	budget := contextBudget{RequestedBytes: options.Bytes, RequestedTokens: options.Tokens}
	switch {
	case options.Bytes > 0:
		budget.EffectiveBytes = options.Bytes
	case options.Tokens > 0:
		budget.EffectiveBytes = options.Tokens * bytesPerToken
	case config.MaxBytes > 0:
		budget.EffectiveBytes = config.MaxBytes
	default:
		budget.EffectiveBytes = defaultContextBytes
	}
	return budget, nil
}

func collectContextSections(ctx context.Context, state *GlobalState, root string, observed time.Time) []contextSection {
	collectors := state.Config.Context.Collectors
	if len(collectors) == 0 {
		collectors = defaultContextCollectors
	}
	allowed := map[string]func() contextSection{
		"instructions": func() contextSection {
			return collectProjectFiles(root, "instructions", 1, state.Config.Context.InstructionFiles, observed)
		},
		"skills": func() contextSection { return collectSkillMetadata(root, observed) },
		"git":    func() contextSection { return collectGitContext(ctx, state, root, observed) },
		"tasks": func() contextSection {
			return collectTaskContext(ctx, state, root, observed)
		},
		"dependencies": func() contextSection { return collectDependencies(ctx, state, root, observed) },
		"failures": func() contextSection {
			return collectProjectFiles(root, "failures", 6, state.Config.Context.FailureFiles, observed)
		},
	}
	sections := make([]contextSection, 0, len(collectors))
	seen := make(map[string]bool)
	for _, name := range collectors {
		if seen[name] {
			continue
		}
		collector, ok := allowed[name]
		if !ok {
			sections = append(sections, finalizeContextSection(contextSection{ID: name, Priority: 99, ObservedAt: observed.Format(time.RFC3339), Error: "collector is not allowlisted"}))
			continue
		}
		seen[name] = true
		sections = append(sections, collector())
	}
	sort.SliceStable(sections, func(i, j int) bool { return sections[i].Priority < sections[j].Priority })
	return sections
}

func collectProjectFiles(root, id string, priority int, paths []string, observed time.Time) contextSection {
	section := contextSection{ID: id, Priority: priority, ObservedAt: observed.Format(time.RFC3339)}
	if len(paths) == 0 {
		section.Content = "No sources configured.\n"
		return finalizeContextSection(section)
	}
	for _, path := range paths {
		content, err := readProjectFile(root, path)
		if err != nil {
			section.Error = errors.Join(errorFromString(section.Error), fmt.Errorf("%s: %w", path, err)).Error()
			continue
		}
		section.Sources = append(section.Sources, filepath.ToSlash(path))
		section.Content += "## " + filepath.ToSlash(path) + "\n" + content + "\n"
	}
	return finalizeContextSection(section)
}

func collectSkillMetadata(root string, observed time.Time) contextSection {
	section := contextSection{ID: "skills", Priority: 2, ObservedAt: observed.Format(time.RFC3339)}
	skillsDir := filepath.Join(root, "skills")
	info, err := os.Lstat(skillsDir)
	if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		section.Error = "skills directory unavailable"
		return finalizeContextSection(section)
	}
	if boundaryErr := ensureResolvedWithinRoot(root, skillsDir); boundaryErr != nil {
		section.Error = "skills directory escapes project root"
		return finalizeContextSection(section)
	}
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		section.Error = "skills directory unavailable"
		return finalizeContextSection(section)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join("skills", entry.Name(), "SKILL.md")
		content, readErr := readProjectFile(root, path)
		if readErr != nil {
			continue
		}
		metadata := skillFrontmatter(content)
		section.Sources = append(section.Sources, filepath.ToSlash(path))
		section.Content += "- " + strings.ReplaceAll(metadata, "\n", " ") + "\n"
	}
	if section.Content == "" {
		section.Content = "No project-local skill metadata found.\n"
	}
	return finalizeContextSection(section)
}

func collectGitContext(ctx context.Context, state *GlobalState, root string, observed time.Time) contextSection {
	section := contextSection{ID: "git", Priority: 3, Sources: []string{"git status --short --branch", "git log -5"}, ObservedAt: observed.Format(time.RFC3339)}
	status, statusErr := state.Runner.Run(ctx, root, nil, "git", "status", "--short", "--branch")
	log, logErr := state.Runner.Run(ctx, root, nil, "git", "log", "-5", "--pretty=format:%h %s")
	section.Content = "## Status\n" + status + "\n## Recent commits\n" + log + "\n"
	if joined := errors.Join(statusErr, logErr); joined != nil {
		section.Error = redactProjectRoot(joined.Error(), root)
	}
	return finalizeContextSection(section)
}

func collectTaskContext(ctx context.Context, state *GlobalState, root string, observed time.Time) contextSection {
	section := contextSection{ID: "tasks", Priority: 4, Sources: []string{"mise tasks --json"}, ObservedAt: observed.Format(time.RFC3339)}
	if _, err := state.Runner.LookPath("mise"); err != nil {
		section.Error = "mise unavailable"
		return finalizeContextSection(section)
	}
	raw, err := state.Runner.Run(ctx, root, nil, "mise", "tasks", "--json")
	if err != nil {
		section.Error = redactProjectRoot(err.Error(), root)
		return finalizeContextSection(section)
	}
	var tasks []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Aliases     []string `json:"aliases"`
	}
	if decodeErr := json.Unmarshal([]byte(raw), &tasks); decodeErr != nil {
		section.Error = "mise returned invalid task metadata"
		return finalizeContextSection(section)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	for i := range tasks {
		sort.Strings(tasks[i].Aliases)
	}
	content, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		section.Error = "failed to encode task metadata"
		return finalizeContextSection(section)
	}
	section.Content = string(content) + "\n"
	return finalizeContextSection(section)
}

func collectDependencies(ctx context.Context, state *GlobalState, root string, observed time.Time) contextSection {
	tracked, err := state.Runner.Run(ctx, root, nil, "git", "ls-files", "-z")
	if err != nil {
		return finalizeContextSection(contextSection{ID: "dependencies", Priority: 5, ObservedAt: observed.Format(time.RFC3339), Error: err.Error()})
	}
	allowed := make(map[string]bool)
	for _, name := range state.Config.Context.DependencyFiles {
		allowed[name] = true
	}
	if len(allowed) == 0 {
		for _, name := range defaultContextConfig().DependencyFiles {
			allowed[name] = true
		}
	}
	var paths []string
	for _, path := range strings.Split(tracked, "\x00") {
		if allowed[filepath.Base(path)] || allowed[filepath.ToSlash(path)] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return collectProjectFiles(root, "dependencies", 5, paths, observed)
}

func readProjectFile(root, relative string) (string, error) {
	if filepath.IsAbs(relative) || relative == "" {
		return "", errors.New("path must be project-relative")
	}
	path := filepath.Clean(filepath.Join(root, relative))
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes project root")
	}
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New("source does not exist")
		}
		return "", errors.New("source metadata is unavailable")
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", errors.New("source must be a regular non-symlink file")
	}
	if boundaryErr := ensureResolvedWithinRoot(root, path); boundaryErr != nil {
		return "", boundaryErr
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("source is unreadable")
	}
	return string(content), nil
}

func ensureResolvedWithinRoot(root, path string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return errors.New("project root is unavailable")
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return errors.New("source resolution failed")
	}
	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("source resolves outside project root")
	}
	return nil
}

func redactProjectRoot(value, root string) string {
	value = strings.ReplaceAll(value, root, ".")
	return strings.ReplaceAll(value, filepath.ToSlash(root), ".")
}

func skillFrontmatter(content string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) == 3 {
		var lines []string
		for _, line := range strings.Split(parts[1], "\n") {
			if strings.HasPrefix(line, "name:") || strings.HasPrefix(line, "description:") {
				lines = append(lines, strings.TrimSpace(line))
			}
		}
		return strings.Join(lines, " | ")
	}
	return "metadata unavailable"
}

func finalizeContextSection(section contextSection) contextSection {
	hash := sha256.Sum256([]byte(section.Content + "\n" + section.Error + "\n" + strings.Join(section.Sources, "\n")))
	section.Fingerprint = "sha256:" + hex.EncodeToString(hash[:])
	return section
}

func errorFromString(message string) error {
	if message == "" {
		return nil
	}
	return errors.New(message)
}

func contextUnits(sections []contextSection) ([]contextUnit, [][]string) {
	lines := make([][]string, len(sections))
	maxLines := 0
	for i, section := range sections {
		lines[i] = strings.SplitAfter(section.Content, "\n")
		if len(lines[i]) == 1 && lines[i][0] == "" {
			lines[i] = nil
		}
		maxLines = max(maxLines, len(lines[i]))
	}
	var units []contextUnit
	for line := range maxLines {
		for section := range sections {
			if line < len(lines[section]) {
				units = append(units, contextUnit{section: section, line: line})
			}
		}
	}
	return units, lines
}

func contextSectionView(section contextSection, lines []string, selected map[int]bool) contextJSONSection {
	var content strings.Builder
	var omitted []int
	includedBytes := 0
	for i, line := range lines {
		if selected[i] {
			content.WriteString(line)
			includedBytes += len(line)
		} else {
			omitted = append(omitted, i+1)
		}
	}
	status := "partial"
	if len(lines) == 0 || len(omitted) == 0 {
		status = "complete"
	} else if len(selected) == 0 {
		status = "omitted-section"
	}
	return contextJSONSection{ID: section.ID, Priority: section.Priority, Sources: section.Sources, Fingerprint: section.Fingerprint, ObservedAt: section.ObservedAt, Status: status, Error: section.Error, TotalBytes: len(section.Content), OmittedBytes: len(section.Content) - includedBytes, OmittedLines: formatIntRanges(omitted), Content: content.String()}
}

func renderContextJSON(root, head string, generated time.Time, budget contextBudget, sections []contextSection, lines [][]string, selected []map[int]bool) []byte {
	envelope := contextEnvelope{SchemaVersion: contextSchemaVersion, GeneratedAt: generated.Format(time.RFC3339), Repository: root, SourceHead: head, Budget: budget, Adapters: []string{"session-metadata (optional, unavailable in project-only v1)"}}
	for i, section := range sections {
		envelope.Sections = append(envelope.Sections, contextSectionView(section, lines[i], selected[i]))
	}
	payload, _ := json.MarshalIndent(envelope, "", "  ")
	return append(payload, '\n')
}

func renderContextMarkdown(root, head string, generated time.Time, budget contextBudget, sections []contextSection, lines [][]string, selected []map[int]bool) []byte {
	var out strings.Builder
	fmt.Fprintf(&out, "# Project context\n\n- Schema: %s\n- Generated: %s\n- Repository: %s\n- Source HEAD: %s\n- Effective budget: %d bytes\n- Requested bytes: %d\n- Requested tokens: %d\n- Optional adapter: session-metadata (unavailable in project-only v1)\n\n## Manifest\n", contextSchemaVersion, generated.Format(time.RFC3339), root, head, budget.EffectiveBytes, budget.RequestedBytes, budget.RequestedTokens)
	views := make([]contextJSONSection, len(sections))
	for i, section := range sections {
		views[i] = contextSectionView(section, lines[i], selected[i])
		fmt.Fprintf(&out, "- %s | priority=%d | status=%s | total_bytes=%d | omitted_bytes=%d | omitted_lines=%s | observed=%s | fingerprint=%s | sources=%s | error=%s\n", section.ID, section.Priority, views[i].Status, views[i].TotalBytes, views[i].OmittedBytes, views[i].OmittedLines, section.ObservedAt, section.Fingerprint, strings.Join(section.Sources, ","), cmpContextValue(section.Error))
	}
	for i, view := range views {
		if view.Content == "" {
			continue
		}
		fmt.Fprintf(&out, "\n## %s\n\n", sections[i].ID)
		for _, line := range strings.SplitAfter(view.Content, "\n") {
			if line == "" {
				continue
			}
			out.WriteString("    ")
			out.WriteString(line)
		}
	}
	return []byte(out.String())
}

func cmpContextValue(value string) string {
	if value == "" {
		return "none"
	}
	return strings.ReplaceAll(value, "\n", "; ")
}

func ScanContextPayload(ctx context.Context, state *GlobalState, payload string) error {
	if err := scanPayloadForSecrets(ctx, state, payload); err != nil {
		return fmt.Errorf("context payload secret scan failed: %w", err)
	}
	for _, pattern := range state.Config.Context.SensitivePathPatterns {
		if pattern != "" && strings.Contains(payload, pattern) {
			return errors.New("context payload contains a configured sensitive path pattern")
		}
	}
	for _, pattern := range state.Config.Context.SensitiveEnvPatterns {
		for _, entry := range os.Environ() {
			name, value, found := strings.Cut(entry, "=")
			if !found || value == "" {
				continue
			}
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				return fmt.Errorf("invalid sensitive environment pattern %q: %w", pattern, err)
			}
			if matched && strings.Contains(payload, value) {
				return fmt.Errorf("context payload contains the value of environment variable %s matching a sensitive pattern", name)
			}
		}
	}
	return nil
}
