package dot

import (
	"cmp"
	"context"
	"errors"
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
	allDiff, err := GetCachedDiffUnfiltered(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to get unfiltered git diff: %w", err)
	}

	autoStaged := false
	rollback := func(cause error) error {
		if !autoStaged {
			return cause
		}
		if _, rollbackErr := state.Runner.Run(context.WithoutCancel(ctx), "", nil, "git", "reset", "--mixed"); rollbackErr != nil {
			return errors.Join(cause, fmt.Errorf("failed to restore initially clean index: %w", rollbackErr))
		}
		return cause
	}

	if strings.TrimSpace(allDiff) == "" {
		status, statusErr := state.Runner.Run(ctx, "", nil, "git", "status", "--porcelain")
		if statusErr != nil {
			return fmt.Errorf("failed to inspect working tree: %w", statusErr)
		}
		if strings.TrimSpace(status) == "" {
			_, _ = fmt.Fprintln(state.Stdout, "No changes to commit.")
			return nil
		}

		autoStaged = true
		if _, addErr := state.Runner.Run(ctx, "", nil, "git", "add", "-A"); addErr != nil {
			return rollback(fmt.Errorf("failed to stage working tree changes: %w", addErr))
		}
		allDiff, err = GetCachedDiffUnfiltered(ctx, state)
		if err != nil {
			return rollback(fmt.Errorf("failed to get staged git diff: %w", err))
		}
		if strings.TrimSpace(allDiff) == "" {
			return rollback(errors.New("git add -A completed without producing a staged diff"))
		}
		_, _ = fmt.Fprintln(state.Stdout, "No staged changes found. Staged all working tree changes for commit message generation...")
	}

	diff, err := GetCachedDiff(ctx, state)
	if err != nil {
		return rollback(fmt.Errorf("failed to get git diff: %w", err))
	}

	if strings.TrimSpace(diff) == "" {
		return rollback(errors.New("git changes exist, but every changed path is excluded from AI diff generation"))
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

	aiDiff := limitAIInput(diff, state.Config.Commit.MaxDiffSize)
	if scanErr := ScanDiffForSecrets(ctx, state, aiDiff); scanErr != nil {
		return rollback(scanErr)
	}

	msg, err := GenerateText(ctx, state, prompt, aiDiff, state.Config.Commit.MaxDiffSize)
	if err != nil {
		return rollback(err)
	}

	err = state.Runner.RunInteractive(ctx, "", "git", "commit", "-e", "-m", msg)
	if err != nil {
		return rollback(fmt.Errorf("git commit failed: %w", err))
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
