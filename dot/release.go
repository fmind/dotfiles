package dot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"
)

// NewReleaseCmd constructs the top-level release command.
func NewReleaseCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "release",
		Aliases: []string{"r"},
		Usage:   "Bump version, update CHANGELOG, and publish a GitHub release",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Automatic yes to prompts; assume 'yes' as answer to all questions",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			force := cmd.Bool("yes")
			return RunRelease(ctx, state, force)
		},
	}
}

// RunRelease performs the full release sequence.
func RunRelease(ctx context.Context, state *GlobalState, force bool) error {
	// 1. Verify inside work tree
	if err := IsInsideWorkTree(ctx, state); err != nil {
		return err
	}

	// 2. Check preconditions
	// Check git clean
	status, err := state.Runner.Run(ctx, "", nil, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return errors.New("working directory has uncommitted or staged changes; commit or stash them first")
	}

	// Check gh auth
	_, err = state.Runner.Run(ctx, "", nil, "gh", "auth", "status")
	if err != nil {
		return errors.New("github CLI is not authenticated; run 'gh auth login' or set GH_TOKEN")
	}

	// Check git-cliff
	if _, err = state.Runner.LookPath("git-cliff"); err != nil {
		return errors.New("git-cliff is not installed; run 'mr t' or install it via mise")
	}

	// 3. Compute bumped version
	bumped, err := state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--bumped-version")
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}
	bumped = strings.TrimSpace(bumped)

	current, err := state.Runner.Run(ctx, "", nil, "git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		current = "v0.0.0"
	}
	current = strings.TrimSpace(current)

	if bumped == current {
		_, _ = fmt.Fprintf(state.Stdout, "No new conventional commits since %s. Nothing to release.\n", current)
		return nil
	}

	_, _ = fmt.Fprintf(state.Stdout, "Current version: %s\n", yellow(current))
	_, _ = fmt.Fprintf(state.Stdout, "Next version:    %s\n", green(bumped))

	// Confirm unless forced
	if !force {
		question := fmt.Sprintf("Proceed with releasing %s? [y/N]: ", bumped)
		if !confirmRelease(state.Stdin, state.Stdout, question) {
			_, _ = fmt.Fprintln(state.Stdout, "Release canceled.")
			return nil
		}
	}

	// 4. Update Version in dot/version.go
	content, err := os.ReadFile("dot/version.go") //nolint:gosec // G304, G703
	if err != nil {
		return fmt.Errorf("failed to read version.go: %w", err)
	}
	versionNum := strings.TrimPrefix(bumped, "v")
	re := regexp.MustCompile(`var Version = ".*"`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`var Version = "%s"`, versionNum))
	err = os.WriteFile("dot/version.go", []byte(newContent), 0o644) //nolint:gosec // G304, G703
	if err != nil {
		return fmt.Errorf("failed to write version.go: %w", err)
	}

	// 5. Generate CHANGELOG.md
	_, err = state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--bump", "-o", "CHANGELOG.md")
	if err != nil {
		return fmt.Errorf("failed to generate CHANGELOG.md: %w", err)
	}

	// 6. Format and Check project
	if _, err = state.Runner.LookPath("mise"); err == nil {
		_, _ = fmt.Fprintln(state.Stdout, "Formatting files...")
		_ = state.Runner.RunInteractive(ctx, "", "mise", "run", "format")
		_, _ = fmt.Fprintln(state.Stdout, "Running checks...")
		if err = state.Runner.RunInteractive(ctx, "", "mise", "run", "check"); err != nil {
			return fmt.Errorf("project checks failed: %w", err)
		}
	}

	// 7. Commit & Tag
	_, err = state.Runner.Run(ctx, "", nil, "git", "add", "CHANGELOG.md", "dot/version.go")
	if err != nil {
		return fmt.Errorf("failed to stage release files: %w", err)
	}
	_, err = state.Runner.Run(ctx, "", nil, "git", "commit", "-m", "chore(release): "+bumped)
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	_, err = state.Runner.Run(ctx, "", nil, "git", "tag", "-a", bumped, "-m", bumped)
	if err != nil {
		return fmt.Errorf("git tag failed: %w", err)
	}

	// 8. Push
	currentBranch, err := state.Runner.Run(ctx, "", nil, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch = strings.TrimSpace(currentBranch)

	err = state.Runner.RunInteractive(ctx, "", "git", "push", "origin", currentBranch)
	if err != nil {
		return fmt.Errorf("git push origin %s failed: %w", currentBranch, err)
	}
	err = state.Runner.RunInteractive(ctx, "", "git", "push", "origin", bumped)
	if err != nil {
		return fmt.Errorf("git push tag %s failed: %w", bumped, err)
	}

	// 9. Publish GitHub Release
	tempFile, err := os.CreateTemp("", "release-notes-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()
	_ = tempFile.Close()

	notes, err := state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--latest", "--strip", "all")
	if err != nil {
		return fmt.Errorf("failed to generate latest changelog: %w", err)
	}

	err = os.WriteFile(tempPath, []byte(notes), 0o644) //nolint:gosec // G304, G703
	if err != nil {
		return fmt.Errorf("failed to write release notes: %w", err)
	}

	err = state.Runner.RunInteractive(ctx, "", "gh", "release", "create", bumped, "--title", bumped, "--notes-file", tempPath)
	if err != nil {
		return fmt.Errorf("failed to create github release: %w", err)
	}

	_, _ = fmt.Fprintf(state.Stdout, "%s Release %s successfully published!\n", passIcon, bumped)
	return nil
}

func confirmRelease(stdin io.Reader, stdout io.Writer, msg string) bool {
	_, _ = fmt.Fprint(stdout, msg)
	reader := bufio.NewReader(stdin)
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))
	return ans == "y" || ans == "yes"
}
