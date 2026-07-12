package dot

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
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
func findPRTemplate(paths []string) (string, bool, error) {
	for _, p := range paths {
		info, err := os.Stat(p)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", false, fmt.Errorf("failed to inspect PR template %s: %w", p, err)
		}
		// GitHub also supports a `.github/PULL_REQUEST_TEMPLATE/` directory holding
		// multiple templates; skip directories rather than failing on the EISDIR read.
		if info.IsDir() {
			continue
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return "", false, fmt.Errorf("failed to read PR template %s: %w", p, err)
		}
		return string(content), true, nil
	}
	return "", false, nil
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

	allDiff, err := GetBaseDiffUnfiltered(ctx, state, baseBranch)
	if err != nil {
		return err
	}

	if strings.TrimSpace(allDiff) == "" {
		_, _ = fmt.Fprintf(state.Stdout, "No changes detected against base branch '%s'.\n", baseBranch)
		return nil
	}

	diff, err := GetBaseDiff(ctx, state, baseBranch)
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		return fmt.Errorf("changes exist against base branch %q, but every changed path is excluded from AI diff generation", baseBranch)
	}

	prompt := cmp.Or(state.Config.PR.Prompt, DefaultPRPrompt)

	templates := state.Config.PR.Templates
	if len(templates) == 0 {
		templates = defaultPRTemplates
	}

	template, found, err := findPRTemplate(templates)
	if err != nil {
		return err
	}
	if found {
		prompt = fmt.Sprintf("%s\n\nFollow the structure, guidelines, and sections of the repository's pull request template provided below:\n\n%s", prompt, template)
	}

	aiBinary := GetAIBinary(state)
	_, _ = fmt.Fprintf(state.Stdout, "Generating PR description using %s...\n", aiBinary)

	aiDiff := limitAIInput(diff, state.Config.Commit.MaxDiffSize)
	if scanErr := ScanDiffForSecrets(ctx, state, aiDiff); scanErr != nil {
		return scanErr
	}

	description, err := GenerateText(ctx, state, prompt, aiDiff, state.Config.Commit.MaxDiffSize)
	if err != nil {
		return err
	}

	return withPRDescriptionFile(description, func(path string) error {
		_, _ = fmt.Fprintln(state.Stdout, "Launching 'gh pr create'...")
		args := []string{"pr", "create", "-B", baseBranch, "-F", path}
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
		if runErr := state.Runner.RunInteractive(ctx, "", ghPath, args...); runErr != nil {
			return fmt.Errorf("gh pr create failed: %w", runErr)
		}
		return nil
	})
}

func withPRDescriptionFile(description string, run func(path string) error) error {
	tempFile, err := os.CreateTemp("", "pr-desc-*.md")
	if err != nil {
		return fmt.Errorf("failed to create PR description file: %w", err)
	}
	path := tempFile.Name()
	if writeErr := writePRDescription(tempFile, description); writeErr != nil {
		return errors.Join(writeErr, removeTemporaryFile(path, "PR description"))
	}
	return errors.Join(run(path), removeTemporaryFile(path, "PR description"))
}

func writePRDescription(file io.WriteCloser, description string) error {
	if _, err := io.WriteString(file, description); err != nil {
		writeErr := fmt.Errorf("failed to write PR description: %w", err)
		if closeErr := file.Close(); closeErr != nil {
			return errors.Join(writeErr, fmt.Errorf("failed to close PR description file: %w", closeErr))
		}
		return writeErr
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close PR description file: %w", err)
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
