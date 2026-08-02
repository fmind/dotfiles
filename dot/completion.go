package dot

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

const (
	defaultCompletionTimeout     = 60 * time.Second
	defaultCompletionConcurrency = 4
)

// shellIntegration is a tool whose fish init script is cached wholesale rather than
// generated as a completion file.
type shellIntegration struct {
	tool string
	file string
	args []string
}

var shellIntegrations = []shellIntegration{
	{tool: "atuin", file: "atuin-init.fish", args: []string{"init", "fish"}},
	{tool: "carapace", file: "carapace-init.fish", args: []string{"_carapace", "fish"}},
}

// fishCacheRoot resolves the XDG cache root, falling back to ~/.cache.
func fishCacheRoot() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(home, ".cache")
}

// writeShellIntegration caches one tool's fish init script. A tool that is not
// installed is skipped silently: these integrations are optional by design.
func writeShellIntegration(ctx context.Context, state *GlobalState, integration shellIntegration, cacheDir string, mu *sync.Mutex, genErrors *[]error) {
	if _, err := state.Runner.LookPath(integration.tool); err != nil {
		return
	}
	out, err := state.Runner.Run(ctx, cacheDir, nil, integration.tool, integration.args...)

	mu.Lock()
	defer mu.Unlock()
	if err != nil {
		_, _ = fmt.Fprintf(state.Stdout, "  %s Failed to generate %s: %v\n", failIcon, integration.file, err)
		*genErrors = append(*genErrors, fmt.Errorf("failed to generate %s: %w", integration.file, err))
		return
	}
	if err := os.WriteFile(filepath.Join(cacheDir, integration.file), []byte(out), 0o600); err != nil {
		_, _ = fmt.Fprintf(state.Stdout, "  %s Failed to write %s: %v\n", failIcon, integration.file, err)
		*genErrors = append(*genErrors, fmt.Errorf("failed to write %s: %w", integration.file, err))
		return
	}
	_, _ = fmt.Fprintf(state.Stdout, "  %s Generated %s\n", passIcon, integration.file)
}

// NewCompletionCmd constructs the top-level completion command.
func NewCompletionCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "completion",
		Aliases: []string{"g"},
		Usage:   "Generate fish autocompletions for dot itself and external CLI tools",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunCompletionGenerate(ctx, state)
		},
	}
}

// dotCompletionTmpl is the native fish completion for the dot CLI. It mirrors
// urfave/cli's upstream fish integration: gather the typed words, re-invoke dot
// with the hidden --generate-shell-completion flag (enabled in NewApp), and split
// each "name:usage" line into a fish value/description pair. This keeps the
// completions in lockstep with the live command tree instead of a static list.
const dotCompletionTmpl = `
function __dot_perform_completion
	set -l args (commandline -opc)
	set -l lastArg (commandline -ct)
	if string match -q -- "-*" $lastArg
		set results ($args[1] $args[2..-1] $lastArg --generate-shell-completion 2>/dev/null)
	else
		set results ($args[1] $args[2..-1] --generate-shell-completion 2>/dev/null)
	end
	for line in $results
		if not string match -q -- "dot*" $line
			set -l parts (string split -m 1 ":" -- "$line")
			if test (count $parts) -eq 2
				printf "%s\t%s\n" "$parts[1]" "$parts[2]"
			else
				printf "%s\n" "$line"
			end
		end
	end
end
complete -c dot -e
complete -c dot -f -a '(__dot_perform_completion)'
`

// RunCompletionGenerate generates and writes fish completions for all configured CLI tools to the configured completions path.
func RunCompletionGenerate(ctx context.Context, state *GlobalState) error {
	path := cmp.Or(state.Config.Completions.Path, DefaultCompletionsPath)
	compDir := ExpandPath(path)
	err := os.MkdirAll(compDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create completions directory: %w", err)
	}

	tools := state.Config.Completions.Tools
	_, _ = fmt.Fprintf(state.Stdout, "=> Generating Fish Autocompletions for %d tools in %s...\n\n", len(tools), compDir)

	var mu sync.Mutex // Protect concurrent writes to state.Stdout and genErrors
	var genErrors []error

	// Prepare the shell-integration cache directory before any worker starts. Every
	// append to genErrors must happen either here, on the single goroutine that owns
	// it before the pool exists, or under mu — appending after the workers launch
	// would race their guarded appends.
	fishCacheDir := filepath.Join(fishCacheRoot(), "fish")
	cacheErr := os.MkdirAll(fishCacheDir, 0o700)
	if cacheErr != nil {
		_, _ = fmt.Fprintf(state.Stdout, "  %s Failed to create fish cache directory: %v\n", failIcon, cacheErr)
		genErrors = append(genErrors, fmt.Errorf("failed to create fish cache directory: %w", cacheErr))
	}

	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(positiveOr(state.Config.Completions.Concurrency, defaultCompletionConcurrency))

	for _, t := range tools {
		g.Go(func() error {
			writeToolCompletion(groupCtx, state, t, compDir, &mu, &genErrors)
			return nil
		})
	}
	if cacheErr == nil {
		for _, integration := range shellIntegrations {
			g.Go(func() error {
				writeShellIntegration(groupCtx, state, integration, fishCacheDir, &mu, &genErrors)
				return nil
			})
		}
	}

	_ = g.Wait()

	// An interrupted run (Ctrl-C) leaves only partial output: skipped tools are not
	// errors, so surface the cancellation directly instead of printing "✓ updated".
	// Mirrors the post-Wait context check in pull.go and status.go.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	dotPath := filepath.Join(compDir, "dot.fish")
	if err := os.WriteFile(dotPath, []byte(strings.TrimSpace(dotCompletionTmpl)+"\n"), 0o644); err != nil {
		_, _ = fmt.Fprintf(state.Stdout, "  %s Failed to write dot.fish: %v\n", failIcon, err)
		genErrors = append(genErrors, fmt.Errorf("failed to write completions for dot: %w", err))
	} else {
		_, _ = fmt.Fprintf(state.Stdout, "  %s Generated completions for dot\n", passIcon)
	}

	if len(genErrors) > 0 {
		_, _ = fmt.Fprintf(state.Stdout, "\n%s\n", red("✗ Failed to generate some completions in "+compDir))
		return errors.Join(genErrors...)
	}

	_, _ = fmt.Fprintf(state.Stdout, "\n%s\n", green("✓ Completions updated in "+compDir))
	return nil
}

// writeToolCompletion generates fish completions for a tool and writes them to the completions directory.
func writeToolCompletion(ctx context.Context, state *GlobalState, tool, compDir string, mu *sync.Mutex, genErrors *[]error) {
	if ctx.Err() != nil {
		return
	}

	out, err := generateToolCompletion(ctx, state, tool)
	if err != nil {
		mu.Lock()
		if errors.Is(err, ErrToolNotInstalled) {
			_, _ = fmt.Fprintf(state.Stdout, "  %s %s is not installed, skipping\n", skipIcon, tool)
		} else {
			_, _ = fmt.Fprintf(state.Stdout, "  %s Failed to generate completions for %s\n", skipIcon, tool)
			*genErrors = append(*genErrors, fmt.Errorf("failed to generate completions for %s: %w", tool, err))
		}
		mu.Unlock()
		return
	}

	filePath := filepath.Join(compDir, tool+".fish")
	err = os.WriteFile(filePath, []byte(out), 0o644)
	if err != nil {
		mu.Lock()
		_, _ = fmt.Fprintf(state.Stdout, "  %s Failed to write %s.fish: %v\n", failIcon, tool, err)
		*genErrors = append(*genErrors, fmt.Errorf("failed to write completions for %s: %w", tool, err))
		mu.Unlock()
		return
	}

	mu.Lock()
	_, _ = fmt.Fprintf(state.Stdout, "  %s Generated completions for %s\n", passIcon, tool)
	mu.Unlock()
}

// generateToolCompletion attempts to generate fish completion output for a single tool.
// It returns the output string, or an error if the generation fails.
// If the tool is not installed, it returns ErrToolNotInstalled.
func generateToolCompletion(ctx context.Context, state *GlobalState, tool string) (output string, resultErr error) {
	_, err := state.Runner.LookPath(tool)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrToolNotInstalled, tool)
	}

	// Completion commands should be read-only, but some CLIs interpret an unknown
	// subcommand as a database or output filename. Isolate those side effects from
	// the caller's repository and discard them after each command finishes.
	workDir, err := os.MkdirTemp("", "dot-completion-")
	if err != nil {
		return "", fmt.Errorf("failed to create completion workspace: %w", err)
	}
	defer func() {
		if cleanupErr := removeTemporaryDirectory(workDir, "completion workspace"); cleanupErr != nil {
			resultErr = errors.Join(resultErr, cleanupErr)
		}
	}()

	binary, args := GetCompletionCommand(state, tool)

	timeout := positiveOr(state.Config.Completions.Timeout.Duration(), defaultCompletionTimeout)
	run := func(name string, commandArgs ...string) (string, error) {
		commandCtx, cancel := context.WithTimeout(ctx, timeout)
		out, runErr := state.Runner.Run(commandCtx, workDir, nil, name, commandArgs...)
		cancel()
		if runErr == nil && strings.TrimSpace(out) == "" {
			return "", errors.New("completion command returned no output")
		}
		return out, runErr
	}

	out, err := run(binary, args...)
	if err != nil {
		// Fallback to standard "t completion fish" if the attempted command was different.
		isStandardFallback := binary == tool && len(args) == 2 && args[0] == "completion" && args[1] == "fish"
		if !isStandardFallback {
			out, err = run(tool, "completion", "fish")
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate completions: %w", err)
	}

	return out, nil
}

// GetCompletionCommand returns the binary name and command-line arguments needed to output fish completions for a tool.
func GetCompletionCommand(state *GlobalState, tool string) (string, []string) {
	if cfg, ok := state.Config.Completions.CustomCommands[tool]; ok {
		return cmp.Or(cfg.Binary, tool), cfg.Args
	}
	return tool, []string{"completion", "fish"}
}

// DefaultCompletionsPath is the default directory path where fish shell completions are generated.
const DefaultCompletionsPath = "~/.config/fish/completions"

// ToolConfig represents the custom command configuration to generate completions for a tool.
type ToolConfig struct {
	Binary string   `yaml:"binary"`
	Args   []string `yaml:"args"`
}

// CompletionConfig represents the configuration for CLI autocompletions generation.
type CompletionConfig struct {
	Path           string                `yaml:"path"`
	CustomCommands map[string]ToolConfig `yaml:"custom_commands"`
	Tools          []string              `yaml:"tools"`
	// Timeout bounds one tool's completion command; Concurrency caps how many run at
	// once. Both fall back to the built-in default when absent or non-positive.
	Timeout     Duration `yaml:"timeout"`
	Concurrency int      `yaml:"concurrency"`
}

func defaultCompletionConfig() CompletionConfig {
	return CompletionConfig{
		Tools: []string{
			"ast-grep", "atlas", "atuin", "bat", "carapace", "chezmoi", "codex", "cosign", "delta", "dive",
			"dlv", "doggo", "dprint", "dyff", "flux", "gh", "git-lfs", "gitleaks", "golangci-lint", "goreleaser",
			"helm", "helmfile", "jules", "just", "k3d", "k9s", "kind", "ko", "kube-linter", "kubecolor",
			"kubectl", "kustomize", "lazygit", "lefthook", "mirrord", "mise", "opencode", "pluto",
			"rg", "ruff", "skaffold", "sqlc", "starship", "step", "stern",
			"terraform-docs", "trivy", "ty", "uv", "watchexec", "xh", "yq", "zellij",
		},
		Path:        DefaultCompletionsPath,
		Timeout:     Duration(defaultCompletionTimeout),
		Concurrency: defaultCompletionConcurrency,
		CustomCommands: map[string]ToolConfig{
			"ast-grep":  {Args: []string{"completions", "fish"}},
			"atlas":     {Args: []string{"completion", "fish"}},
			"atuin":     {Args: []string{"gen-completions", "--shell", "fish"}},
			"bat":       {Args: []string{"--completion", "fish"}},
			"carapace":  {Args: []string{"_carapace", "fish"}},
			"codex":     {Args: []string{"completion", "fish"}},
			"delta":     {Args: []string{"--generate-completion", "fish"}},
			"dive":      {Args: []string{"completion", "fish"}},
			"dlv":       {Args: []string{"completion", "fish"}},
			"doggo":     {Args: []string{"completions", "fish"}},
			"dprint":    {Args: []string{"completions", "fish"}},
			"gh":        {Args: []string{"completion", "-s", "fish"}},
			"git-lfs":   {Binary: "git", Args: []string{"lfs", "completion", "fish"}},
			"just":      {Args: []string{"--completions", "fish"}},
			"lazygit":   {Args: []string{"completion", "fish"}},
			"mirrord":   {Args: []string{"completions", "fish"}},
			"rg":        {Args: []string{"--generate", "complete-fish"}},
			"ruff":      {Args: []string{"generate-shell-completion", "fish"}},
			"starship":  {Args: []string{"completions", "fish"}},
			"stern":     {Args: []string{"--completion", "fish"}},
			"ty":        {Args: []string{"generate-shell-completion", "fish"}},
			"uv":        {Args: []string{"generate-shell-completion", "fish"}},
			"watchexec": {Args: []string{"--completions", "fish"}},
			"xh":        {Args: []string{"--generate", "complete-fish"}},
			"yq":        {Args: []string{"shell-completion", "fish"}},
			"zellij":    {Args: []string{"setup", "--generate-completion", "fish"}},
		},
	}
}
