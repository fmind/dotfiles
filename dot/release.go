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

func validateReleaseStatus(status string) error {
	unexpected := make([]string, 0)
	for _, record := range strings.Split(status, "\x00") {
		if record == "" {
			continue
		}
		if len(record) < 4 || record[2] != ' ' {
			return fmt.Errorf("malformed git status record %q", record)
		}
		if strings.ContainsAny(record[:2], "RC") {
			return fmt.Errorf("release validation does not allow renamed or copied paths: %q", record[3:])
		}
		path := record[3:]
		if path != "CHANGELOG.md" && path != "dot/version.go" {
			unexpected = append(unexpected, path)
		}
	}
	if len(unexpected) > 0 {
		return fmt.Errorf("release validation changed unrelated paths: %s", strings.Join(unexpected, ", "))
	}
	return nil
}

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
	if _, err = state.Runner.LookPath("mise"); err != nil {
		return errors.New("mise is not installed; release validation cannot run")
	}

	// 3. Compute bumped version
	bumped, err := state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--bumped-version")
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}
	bumped = strings.TrimSpace(bumped)
	semverTag := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
	if !semverTag.MatchString(bumped) {
		return fmt.Errorf("git-cliff returned invalid semantic version tag %q", bumped)
	}

	current, err := state.Runner.Run(ctx, "", nil, "git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		current = "v0.0.0"
	}
	current = strings.TrimSpace(current)

	if bumped == current {
		_, _ = fmt.Fprintf(state.Stdout, "No new conventional commits since %s. Nothing to release.\n", current)
		return nil
	}

	currentBranch, err := state.Runner.Run(ctx, "", nil, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch = strings.TrimSpace(currentBranch)
	if currentBranch == "" {
		return errors.New("cannot release from a detached HEAD")
	}
	upstreamRemote, err := state.Runner.Run(ctx, "", nil, "git", "config", "--get", "branch."+currentBranch+".remote")
	if err != nil {
		return fmt.Errorf("failed to resolve upstream remote for branch %q: %w", currentBranch, err)
	}
	upstreamRemote = strings.TrimSpace(upstreamRemote)
	if upstreamRemote == "" || upstreamRemote == "." || strings.HasPrefix(upstreamRemote, "-") {
		return fmt.Errorf("current branch %q has no upstream", currentBranch)
	}
	upstreamRef, err := state.Runner.Run(ctx, "", nil, "git", "config", "--get", "branch."+currentBranch+".merge")
	if err != nil {
		return fmt.Errorf("failed to resolve upstream ref for branch %q: %w", currentBranch, err)
	}
	upstreamRef = strings.TrimSpace(upstreamRef)
	if !strings.HasPrefix(upstreamRef, "refs/heads/") || strings.TrimPrefix(upstreamRef, "refs/heads/") == "" {
		return fmt.Errorf("branch %q has invalid upstream ref %q", currentBranch, upstreamRef)
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
	re := regexp.MustCompile(`(?m)^var Version = "[^"\r\n]*"$`)
	if matches := re.FindAllIndex(content, -1); len(matches) != 1 {
		return fmt.Errorf("dot/version.go must contain exactly one expected version assignment; found %d", len(matches))
	}
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

	// 6. Run the same mandatory validation stages as the repository gate.
	_, _ = fmt.Fprintln(state.Stdout, "Formatting files...")
	if err = state.Runner.RunInteractive(ctx, "", "mise", "run", "format"); err != nil {
		return fmt.Errorf("project formatting failed: %w", err)
	}
	_, _ = fmt.Fprintln(state.Stdout, "Running checks...")
	if err = state.Runner.RunInteractive(ctx, "", "mise", "run", "check"); err != nil {
		return fmt.Errorf("project checks failed: %w", err)
	}
	_, _ = fmt.Fprintln(state.Stdout, "Running tests...")
	if err = state.Runner.RunInteractive(ctx, "", "mise", "run", "test"); err != nil {
		return fmt.Errorf("project tests failed: %w", err)
	}
	status, err = state.Runner.Run(ctx, "", nil, "git", "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return fmt.Errorf("failed to inspect release changes after validation: %w", err)
	}
	if err = validateReleaseStatus(status); err != nil {
		return err
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
	pushRefspec := "HEAD:" + upstreamRef
	err = state.Runner.RunInteractive(ctx, "", "git", "push", "--atomic", upstreamRemote, pushRefspec, bumped)
	if err != nil {
		return fmt.Errorf("atomic push of %s and tag %s to %s failed: %w", pushRefspec, bumped, upstreamRemote, err)
	}

	// 9. Publish GitHub Release
	if err = publishGitHubRelease(ctx, state, bumped); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(state.Stdout, "%s Release %s successfully published!\n", passIcon, bumped)
	return nil
}

func publishGitHubRelease(ctx context.Context, state *GlobalState, bumped string) error {
	tempFile, err := os.CreateTemp("", "release-notes-*.md")
	if err != nil {
		return fmt.Errorf("failed to create release notes file: %w", err)
	}
	tempPath := tempFile.Name()
	if closeErr := tempFile.Close(); closeErr != nil {
		return errors.Join(
			fmt.Errorf("failed to close release notes file: %w", closeErr),
			removeTemporaryFile(tempPath, "release notes"),
		)
	}

	notes, err := state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--latest", "--strip", "all")
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to generate latest changelog: %w", err),
			removeTemporaryFile(tempPath, "release notes"),
		)
	}

	err = os.WriteFile(tempPath, []byte(notes), 0o644) //nolint:gosec // G304, G703
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to write release notes: %w", err),
			removeTemporaryFile(tempPath, "release notes"),
		)
	}

	err = state.Runner.RunInteractive(ctx, "", "gh", "release", "create", bumped, "--title", bumped, "--notes-file", tempPath)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to create github release: %w", err),
			removeTemporaryFile(tempPath, "release notes"),
		)
	}
	return removeTemporaryFile(tempPath, "release notes")
}

func confirmRelease(stdin io.Reader, stdout io.Writer, msg string) bool {
	_, _ = fmt.Fprint(stdout, msg)
	reader := bufio.NewReader(stdin)
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))
	return ans == "y" || ans == "yes"
}
