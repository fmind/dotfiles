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

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

// NewCompletionCmd constructs the top-level completion command.
func NewCompletionCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "completion",
		Aliases: []string{"g", "completions"},
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

	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	var mu sync.Mutex // Protect concurrent writes to state.Stdout and genErrors
	var genErrors []error

	for _, t := range tools {
		g.Go(func() error {
			writeToolCompletion(groupCtx, state, t, compDir, &mu, &genErrors)
			return nil
		})
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
func generateToolCompletion(ctx context.Context, state *GlobalState, tool string) (string, error) {
	_, err := state.Runner.LookPath(tool)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrToolNotInstalled, tool)
	}

	binary, args := GetCompletionCommand(state, tool)

	out, err := state.Runner.Run(ctx, "", nil, binary, args...)
	if err != nil {
		// Fallback to standard "t completion fish" if the attempted command was different.
		isStandardFallback := binary == tool && len(args) == 2 && args[0] == "completion" && args[1] == "fish"
		if !isStandardFallback {
			out, err = state.Runner.Run(ctx, "", nil, tool, "completion", "fish")
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
}

func defaultCompletionConfig() CompletionConfig {
	return CompletionConfig{
		Tools: []string{
			"ast-grep", "atlas", "atuin", "bat", "carapace", "chezmoi", "codex", "cosign", "delta", "dive",
			"dlv", "doggo", "dprint", "dyff", "flux", "gh", "git-lfs", "gitleaks", "golangci-lint", "goreleaser",
			"helm", "helmfile", "jules", "just", "k3d", "k9s", "kind", "ko", "kube-linter", "kubecolor",
			"kubectl", "kustomize", "lazygit", "lefthook", "mirrord", "mise", "opencode", "pluto",
			"rg", "ruff", "sg", "skaffold", "sqlc", "starship", "step", "stern",
			"terraform-docs", "trivy", "ty", "uv", "watchexec", "xh", "yq", "zellij",
		},
		Path: DefaultCompletionsPath,
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
			"sg":        {Args: []string{"completions", "fish"}},
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
