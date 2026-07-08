package dot

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

// NewPrCmd constructs the top-level PR command.
func NewPrCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "pr",
		Aliases: []string{"pr"},
		Usage:   "Generate a GitHub Pull Request description from git diff using AI, and invoke gh pr create",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "base",
				Aliases: []string{"b"},
				Value:   "main",
				Usage:   "The base branch to diff against",
			},
			&cli.StringFlag{
				Name:    "title",
				Aliases: []string{"t"},
				Usage:   "The title of the pull request",
			},
			&cli.BoolFlag{
				Name:    "draft",
				Aliases: []string{"d"},
				Usage:   "Create the pull request as a draft",
			},
			&cli.StringSliceFlag{
				Name:    "label",
				Aliases: []string{"l"},
				Usage:   "Add labels to the pull request",
			},
			&cli.StringSliceFlag{
				Name:    "reviewer",
				Aliases: []string{"r"},
				Usage:   "Request review from users or teams",
			},
			&cli.StringSliceFlag{
				Name:    "assignee",
				Aliases: []string{"a"},
				Usage:   "Assign the pull request to users",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			baseBranch := cmd.String("base")
			if !cmd.IsSet("base") && state.Config.PR.BaseBranch != "" {
				baseBranch = state.Config.PR.BaseBranch
			}
			return RunPr(ctx, state, cmd, baseBranch)
		},
	}
}

var defaultPRTemplates = []string{
	".github/pull_request_template.md",
	".github/PULL_REQUEST_TEMPLATE.md",
	"pull_request_template.md",
	"PULL_REQUEST_TEMPLATE.md",
}

// findPRTemplate searches for a pull request template file in the specified paths.
func findPRTemplate(paths []string) (string, bool) {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			content, err := os.ReadFile(p)
			if err == nil {
				return string(content), true
			}
		}
	}
	return "", false
}

// RunPr generates a pull request description from git diff and opens `gh pr create`.
func RunPr(ctx context.Context, state *GlobalState, cmd *cli.Command, baseBranch string) error {
	if err := IsInsideWorkTree(ctx, state); err != nil {
		return err
	}

	ghPath, err := state.Runner.LookPath("gh")
	if err != nil {
		return ErrGhNotInstalled
	}

	diff, err := GetBaseDiff(ctx, state, baseBranch)
	if err != nil {
		return err
	}

	if strings.TrimSpace(diff) == "" {
		_, _ = fmt.Fprintf(state.Stdout, "No changes detected against base branch '%s'.\n", baseBranch)
		return nil
	}

	prompt := cmp.Or(state.Config.PR.Prompt, DefaultPRPrompt)

	templates := state.Config.PR.Templates
	if len(templates) == 0 {
		templates = defaultPRTemplates
	}

	if template, found := findPRTemplate(templates); found {
		prompt = fmt.Sprintf("%s\n\nFollow the structure, guidelines, and sections of the repository's pull request template provided below:\n\n%s", prompt, template)
	}

	aiBinary := GetAIBinary(state)
	_, _ = fmt.Fprintf(state.Stdout, "Generating PR description using %s...\n", aiBinary)

	description, err := GenerateText(ctx, state, prompt, diff, state.Config.Commit.MaxDiffSize)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "pr-desc-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if _, err = tempFile.WriteString(description); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to write PR description: %w", err)
	}
	_ = tempFile.Close()

	_, _ = fmt.Fprintln(state.Stdout, "Launching 'gh pr create'...")
	args := []string{"pr", "create", "-B", baseBranch, "-F", tempFile.Name()}
	if cmd != nil {
		if title := cmd.String("title"); title != "" {
			args = append(args, "-t", title)
		}
		if cmd.Bool("draft") {
			args = append(args, "-d")
		}
		for _, label := range cmd.StringSlice("label") {
			args = append(args, "-l", label)
		}
		for _, reviewer := range cmd.StringSlice("reviewer") {
			args = append(args, "-r", reviewer)
		}
		for _, assignee := range cmd.StringSlice("assignee") {
			args = append(args, "-a", assignee)
		}
	}
	err = state.Runner.RunInteractive(ctx, "", ghPath, args...)
	if err != nil {
		return fmt.Errorf("gh pr create failed: %w", err)
	}

	return nil
}

// PRConfig represents the configuration for GitHub Pull Request generation.
type PRConfig struct {
	BaseBranch string   `yaml:"base_branch"`
	Prompt     string   `yaml:"prompt"`
	Templates  []string `yaml:"templates"`
}

func defaultPRConfig() PRConfig {
	return PRConfig{
		BaseBranch: "main",
		Prompt:     DefaultPRPrompt,
		Templates:  defaultPRTemplates,
	}
}
