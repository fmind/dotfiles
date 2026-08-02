package dot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	defaultReleaseRemote   = "origin"
	defaultReleaseBranch   = "main"
	defaultReleaseWorkflow = "cd.yml"
	releaseCIWorkflow      = "ci.yml"
	releasePollInterval    = 15 * time.Second
	releaseGateTimeout     = 20 * time.Minute
	// Repository-root-relative paths, matching the root-relative records that
	// `git status --porcelain` emits; file operations join them onto the root
	// resolved by releaseRoot so the command works from any subdirectory.
	releaseVersionFile   = "dot/version.go"
	releaseChangelogFile = "CHANGELOG.md"
	releaseCliffConfig   = "dot_config/git-cliff/cliff.toml"
)

var semverTagPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type ReleaseConfig struct {
	Remote        string `yaml:"remote"`
	DefaultBranch string `yaml:"default_branch"`
	Workflow      string `yaml:"workflow"`
}

func defaultReleaseConfig() ReleaseConfig {
	return ReleaseConfig{Remote: defaultReleaseRemote, DefaultBranch: defaultReleaseBranch, Workflow: defaultReleaseWorkflow}
}

func validateReleaseStatus(status string) error {
	var unexpected []string
	for record := range strings.SplitSeq(status, "\x00") {
		if record == "" {
			continue
		}
		if len(record) < 4 || record[2] != ' ' {
			return fmt.Errorf("malformed git status record %q", record)
		}
		if strings.ContainsAny(record[:2], "RC") {
			return fmt.Errorf("release validation does not allow renamed or copied paths: %q", record[3:])
		}
		if path := record[3:]; path != releaseChangelogFile && path != releaseVersionFile {
			unexpected = append(unexpected, path)
		}
	}
	if len(unexpected) > 0 {
		return fmt.Errorf("release validation changed unrelated paths: %s", strings.Join(unexpected, ", "))
	}
	return nil
}

func NewReleaseCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "release",
		Aliases: []string{"r"},
		Usage:   "Prepare, tag, and push a release commit to trigger CD publication",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "Automatic yes to the release preparation prompt"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunRelease(ctx, state, cmd.Bool("yes"))
		},
	}
}

// RunRelease prepares, tags, and pushes a release commit and tag to trigger CD publication.
func RunRelease(ctx context.Context, state *GlobalState, force bool) error {
	if err := IsInsideWorkTree(ctx, state); err != nil {
		return err
	}
	if err := requireCleanReleaseTree(ctx, state); err != nil {
		return err
	}
	if _, err := state.Runner.Run(ctx, "", nil, "gh", "auth", "status"); err != nil {
		// Keep the probe's own error: a network failure or a broken gh install
		// needs a different fix than a missing login.
		return fmt.Errorf("github CLI is not authenticated (run 'gh auth login' or set GH_TOKEN): %w", err)
	}
	if _, err := state.Runner.LookPath("git-cliff"); err != nil {
		return errors.New("git-cliff is not installed; run 'mise run tools' or install it via mise")
	}
	if _, err := state.Runner.LookPath("mise"); err != nil {
		return errors.New("mise is not installed; release validation cannot run")
	}

	root, err := releaseRoot(ctx, state)
	if err != nil {
		return err
	}

	config := state.Config.Release
	branch, err := releaseBranch(ctx, state)
	if err != nil {
		return err
	}
	if branch != config.DefaultBranch {
		return fmt.Errorf("release preparation requires branch %q, current branch is %q", config.DefaultBranch, branch)
	}
	if _, fetchErr := state.Runner.Run(ctx, "", nil, "git", "fetch", "--prune", "--tags", config.Remote); fetchErr != nil {
		return fmt.Errorf("failed to fetch release remote %q: %w", config.Remote, fetchErr)
	}

	head, upstream, err := releaseHeads(ctx, state, config)
	if err != nil {
		return err
	}
	preparedTag, prepared := preparedReleaseTag(ctx, state, root)
	if head != upstream {
		if !prepared {
			return fmt.Errorf("release branch diverged: HEAD %s does not equal %s/%s %s", head, config.Remote, config.DefaultBranch, upstream)
		}
		parent, parentErr := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "HEAD^")
		if parentErr != nil {
			return fmt.Errorf("failed to resolve the prepared release commit's parent: %w", parentErr)
		}
		if strings.TrimSpace(parent) != upstream {
			return fmt.Errorf("prepared release commit is not directly ahead of %s/%s", config.Remote, config.DefaultBranch)
		}
	}

	if prepared {
		if head != upstream {
			if pushErr := pushPreparedCommit(ctx, state, config, head); pushErr != nil {
				return pushErr
			}
		}
		if pushTagErr := pushReleaseTag(ctx, state, config, preparedTag, head); pushTagErr != nil {
			return pushTagErr
		}
		reapplyDotBinary(ctx, state, root)
		_, _ = fmt.Fprintf(state.Stdout, "%s Prepared and tagged %s at %s.\n", passIcon, preparedTag, head)
		return nil
	}

	bumped, current, err := calculateReleaseVersion(ctx, state, root)
	if err != nil {
		return err
	}
	if bumped == current {
		_, _ = fmt.Fprintf(state.Stdout, "No new conventional commits since %s. Nothing to release.\n", current)
		return nil
	}
	_, _ = fmt.Fprintf(state.Stdout, "Current version: %s\nNext version:    %s\n", yellow(current), green(bumped))
	if !force && !confirmRelease(state.Stdin, state.Stdout, fmt.Sprintf("Prepare and tag %s for publication? [y/N]: ", bumped)) {
		_, _ = fmt.Fprintln(state.Stdout, "Release canceled.")
		return nil
	}
	snapshot, err := snapshotReleaseFiles(root)
	if err != nil {
		return err
	}
	if writeErr := writeReleaseVersion(root, bumped); writeErr != nil {
		return snapshot.restore(writeErr)
	}
	if _, cliffErr := state.Runner.Run(ctx, root, nil, "git-cliff", "--config", releaseCliffConfig, "--bump", "-o", releaseChangelogFile); cliffErr != nil {
		return snapshot.restore(fmt.Errorf("failed to generate %s: %w", releaseChangelogFile, cliffErr))
	}
	if validationErr := validatePreparedRelease(ctx, state, root); validationErr != nil {
		return snapshot.restore(validationErr)
	}
	if _, stageErr := state.Runner.Run(ctx, root, nil, "git", "add", releaseChangelogFile, releaseVersionFile); stageErr != nil {
		return rollbackStagedRelease(ctx, state, snapshot, fmt.Errorf("failed to stage release files: %w", stageErr))
	}
	if _, commitErr := state.Runner.Run(ctx, root, nil, "git", "commit", "-m", "chore(release): "+bumped); commitErr != nil {
		return rollbackStagedRelease(ctx, state, snapshot, fmt.Errorf("git commit failed: %w", commitErr))
	}
	head, err = state.Runner.Run(ctx, "", nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("failed to resolve prepared release commit: %w", err)
	}
	head = strings.TrimSpace(head)
	if err := pushPreparedCommit(ctx, state, config, head); err != nil {
		return err
	}
	if err := pushReleaseTag(ctx, state, config, bumped, head); err != nil {
		return err
	}
	reapplyDotBinary(ctx, state, root)
	_, _ = fmt.Fprintf(state.Stdout, "%s Released and tagged %s at %s.\n", passIcon, bumped, head)
	return nil
}

type releaseFileSnapshot struct {
	root          string
	version       []byte
	changelog     []byte
	versionMode   os.FileMode
	changelogMode os.FileMode
}

func snapshotReleaseFiles(root string) (releaseFileSnapshot, error) {
	version, versionMode, err := snapshotReleaseFile(filepath.Join(root, releaseVersionFile))
	if err != nil {
		return releaseFileSnapshot{}, err
	}
	changelog, changelogMode, err := snapshotReleaseFile(filepath.Join(root, releaseChangelogFile))
	if err != nil {
		return releaseFileSnapshot{}, err
	}
	return releaseFileSnapshot{root: root, version: version, changelog: changelog, versionMode: versionMode, changelogMode: changelogMode}, nil
}

func snapshotReleaseFile(path string) ([]byte, os.FileMode, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to snapshot %s: %w", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to inspect %s: %w", path, err)
	}
	return content, info.Mode().Perm(), nil
}

func (snapshot releaseFileSnapshot) restore(cause error) error {
	var restoreErrors []error
	if err := os.WriteFile(filepath.Join(snapshot.root, releaseVersionFile), snapshot.version, snapshot.versionMode); err != nil {
		restoreErrors = append(restoreErrors, fmt.Errorf("failed to restore %s: %w", releaseVersionFile, err))
	}
	if err := os.WriteFile(filepath.Join(snapshot.root, releaseChangelogFile), snapshot.changelog, snapshot.changelogMode); err != nil {
		restoreErrors = append(restoreErrors, fmt.Errorf("failed to restore %s: %w", releaseChangelogFile, err))
	}
	return errors.Join(append([]error{cause}, restoreErrors...)...)
}

func rollbackStagedRelease(ctx context.Context, state *GlobalState, snapshot releaseFileSnapshot, cause error) error {
	_, resetErr := state.Runner.Run(context.WithoutCancel(ctx), "", nil, "git", "reset", "--mixed", "HEAD")
	if resetErr != nil {
		cause = errors.Join(cause, fmt.Errorf("failed to restore release index: %w", resetErr))
	}
	return snapshot.restore(cause)
}

func requireCleanReleaseTree(ctx context.Context, state *GlobalState) error {
	status, err := state.Runner.Run(ctx, "", nil, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return errors.New("working directory has uncommitted or staged changes; commit or stash them first")
	}
	return nil
}

// releaseRoot resolves the repository root once so every file operation and
// cwd-sensitive subprocess anchors to it, wherever the command was invoked.
func releaseRoot(ctx context.Context, state *GlobalState) (string, error) {
	root, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to resolve the repository root: %w", err)
	}
	if root = strings.TrimSpace(root); root == "" {
		return "", errors.New("git returned an empty repository root")
	}
	return root, nil
}

func releaseBranch(ctx context.Context, state *GlobalState) (string, error) {
	branch, err := state.Runner.Run(ctx, "", nil, "git", "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	if branch = strings.TrimSpace(branch); branch == "" {
		return "", errors.New("cannot prepare a release from a detached HEAD")
	}
	return branch, nil
}

func releaseHeads(ctx context.Context, state *GlobalState, config ReleaseConfig) (string, string, error) {
	head, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve HEAD: %w", err)
	}
	upstream, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", config.Remote+"/"+config.DefaultBranch)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve %s/%s: %w", config.Remote, config.DefaultBranch, err)
	}
	return strings.TrimSpace(head), strings.TrimSpace(upstream), nil
}

func preparedReleaseTag(ctx context.Context, state *GlobalState, root string) (string, bool) {
	subject, err := state.Runner.Run(ctx, "", nil, "git", "log", "-1", "--pretty=%s")
	if err != nil {
		return "", false
	}
	tag := strings.TrimPrefix(strings.TrimSpace(subject), "chore(release): ")
	if !semverTagPattern.MatchString(tag) {
		return "", false
	}
	version, err := readReleaseVersion(root)
	return tag, err == nil && tag == "v"+version
}

func calculateReleaseVersion(ctx context.Context, state *GlobalState, root string) (string, string, error) {
	bumped, err := state.Runner.Run(ctx, root, nil, "git-cliff", "--config", releaseCliffConfig, "--bumped-version")
	if err != nil {
		return "", "", fmt.Errorf("failed to calculate next version: %w", err)
	}
	bumped = strings.TrimSpace(bumped)
	if !semverTagPattern.MatchString(bumped) {
		return "", "", fmt.Errorf("git-cliff returned invalid semantic version tag %q", bumped)
	}
	current, err := state.Runner.Run(ctx, "", nil, "git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		current = "v0.0.0"
	}
	return bumped, strings.TrimSpace(current), nil
}

func readReleaseVersion(root string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, releaseVersionFile))
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", releaseVersionFile, err)
	}
	matches := regexp.MustCompile(`(?m)^var Version = "([^"\r\n]*)"$`).FindAllSubmatch(content, -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("%s must contain exactly one expected version assignment; found %d", releaseVersionFile, len(matches))
	}
	return string(matches[0][1]), nil
}

func writeReleaseVersion(root, tag string) error {
	version, err := readReleaseVersion(root)
	if err != nil {
		return err
	}
	path := filepath.Join(root, releaseVersionFile)
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", releaseVersionFile, err)
	}
	old := `var Version = "` + version + `"`
	updated := strings.Replace(string(content), old, `var Version = "`+strings.TrimPrefix(tag, "v")+`"`, 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil { //nolint:gosec // G703: root-joined repository-owned path
		return fmt.Errorf("failed to write %s: %w", releaseVersionFile, err)
	}
	return nil
}

func validatePreparedRelease(ctx context.Context, state *GlobalState, root string) error {
	for _, task := range []string{"format", "check", "test"} {
		_, _ = fmt.Fprintf(state.Stdout, "Running %s...\n", task)
		if err := state.Runner.RunInteractive(ctx, root, "mise", "run", task); err != nil {
			return fmt.Errorf("project %s failed: %w", task, err)
		}
	}
	status, err := state.Runner.Run(ctx, root, nil, "git", "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return fmt.Errorf("failed to inspect release changes after validation: %w", err)
	}
	return validateReleaseStatus(status)
}

func pushPreparedCommit(ctx context.Context, state *GlobalState, config ReleaseConfig, commit string) error {
	refspec := "HEAD:refs/heads/" + config.DefaultBranch
	if err := state.Runner.RunInteractive(ctx, "", "git", "push", config.Remote, refspec); err == nil {
		return nil
	} else {
		if _, fetchErr := state.Runner.Run(context.WithoutCancel(ctx), "", nil, "git", "fetch", config.Remote, config.DefaultBranch); fetchErr == nil {
			remote, resolveErr := state.Runner.Run(context.WithoutCancel(ctx), "", nil, "git", "rev-parse", config.Remote+"/"+config.DefaultBranch)
			if resolveErr == nil && strings.TrimSpace(remote) == commit {
				return nil
			}
		}
		return fmt.Errorf("failed to push prepared commit %s to %s: %w", commit, refspec, err)
	}
}

func pushReleaseTag(ctx context.Context, state *GlobalState, config ReleaseConfig, tag, commit string) error {
	refspec := "refs/tags/" + tag
	if _, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", refspec); err != nil {
		if _, tagErr := state.Runner.Run(ctx, "", nil, "git", "tag", "-a", tag, "-m", tag, commit); tagErr != nil {
			return fmt.Errorf("failed to create tag %s: %w", tag, tagErr)
		}
	}
	if err := state.Runner.RunInteractive(ctx, "", "git", "push", config.Remote, refspec); err == nil {
		return nil
	} else {
		if _, fetchErr := state.Runner.Run(context.WithoutCancel(ctx), "", nil, "git", "fetch", "--tags", config.Remote); fetchErr == nil {
			remoteTag, resolveErr := state.Runner.Run(context.WithoutCancel(ctx), "", nil, "git", "rev-parse", refspec)
			if resolveErr == nil && strings.TrimSpace(remoteTag) == commit {
				return nil
			}
		}
		return fmt.Errorf("failed to push tag %s to %s: %w", tag, config.Remote, err)
	}
}

func reapplyDotBinary(ctx context.Context, state *GlobalState, root string) {
	if _, err := state.Runner.Run(ctx, root, nil, "mise", "run", "--force", "build"); err == nil {
		home, err := os.UserHomeDir()
		if err == nil {
			_, _ = state.Runner.Run(ctx, root, nil, "chezmoi", "apply", "--force", filepath.Join(home, ".local", "bin", "dot"))
		}
	}
}

func confirmRelease(stdin io.Reader, stdout io.Writer, msg string) bool {
	_, _ = fmt.Fprint(stdout, msg)
	reader := bufio.NewReader(stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
