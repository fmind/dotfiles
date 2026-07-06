package dot

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

// NewCommitCmd constructs the top-level commit command.
func NewCommitCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "commit",
		Aliases: []string{"c"},
		Usage:   "Generate and apply a Conventional Commit message from git diff using AI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage:   "Pre-specify commit type (e.g. feat, fix)",
			},
			&cli.StringFlag{
				Name:    "scope",
				Aliases: []string{"s"},
				Usage:   "Pre-specify commit scope",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			commitType := cmd.String("type")
			commitScope := cmd.String("scope")
			return RunCommit(ctx, state, commitType, commitScope)
		},
	}
}

// RunCommit generates a conventional commit message from git diff using the AI agent, and invokes git commit.
func RunCommit(ctx context.Context, state *GlobalState, commitType, commitScope string) error {
	if err := IsInsideWorkTree(ctx, state); err != nil {
		return err
	}

	allowedTypes := strings.Join(state.Config.Commit.AllowedTypes, ", ")
	diff, err := GetCachedDiff(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	useAll := false
	if strings.TrimSpace(diff) == "" {
		diff, err = GetUnstagedDiff(ctx, state)
		if err != nil {
			return fmt.Errorf("failed to get git diff: %w", err)
		}

		if strings.TrimSpace(diff) == "" {
			_, _ = fmt.Fprintln(state.Stdout, "No changes to commit.")
			return nil
		}
		_, _ = fmt.Fprintln(state.Stdout, "No staged changes found. Generating commit message from unstaged tracked files...")
		useAll = true
	}

	// cmp.Or falls back to the built-in prompt if a config sets commit.prompt to "",
	// mirroring pr.go so an empty override never sends a blank system prompt to the AI.
	prompt := cmp.Or(state.Config.Commit.Prompt, DefaultCommitPrompt)
	if strings.Contains(prompt, "%s") {
		prompt = fmt.Sprintf(prompt, allowedTypes)
	}
	if commitType != "" || commitScope != "" {
		if commitType != "" && commitScope != "" {
			prompt = fmt.Sprintf("%s Use type '%s' and scope '%s'.", prompt, commitType, commitScope)
		} else if commitType != "" {
			prompt = fmt.Sprintf("%s Suggest a scope and use '%s' as the type.", prompt, commitType)
		} else {
			prompt = fmt.Sprintf("%s Use scope '%s' and suggest an appropriate type.", prompt, commitScope)
		}
	}

	msg, err := GenerateText(ctx, state, prompt, diff, state.Config.Commit.MaxDiffSize)
	if err != nil {
		return err
	}

	commitArgs := []string{"commit"}
	if useAll {
		commitArgs = append(commitArgs, "-a")
	}
	commitArgs = append(commitArgs, "-e", "-m", msg)

	err = state.Runner.RunInteractive(ctx, "", "git", commitArgs...)
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

// CommitConfig represents the configuration for Conventional Commits.
type CommitConfig struct {
	Prompt       string   `yaml:"prompt"`
	AllowedTypes []string `yaml:"allowed_types"`
	ExcludeDiff  []string `yaml:"exclude_diff"`
	MaxDiffSize  int      `yaml:"max_diff_size"`
}

func defaultCommitConfig() CommitConfig {
	return CommitConfig{
		AllowedTypes: []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
		MaxDiffSize:  DefaultMaxDiffSize,
		ExcludeDiff: []string{
			"*-lock.json", "*lock.yaml", "package-lock.json", "pnpm-lock.yaml", "uv.lock",
			"poetry.lock", "bun.lockb", "composer.lock", "go.sum", "cargo.lock", "pdm.lock",
			"mix.lock", "pixi.lock",
		},
		Prompt: DefaultCommitPrompt,
	}
}
