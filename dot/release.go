package dot

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
)

var (
	semverTagPattern  = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
	fullCommitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

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
		if path := record[3:]; path != "CHANGELOG.md" && path != "dot/version.go" {
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
		Usage:   "Prepare and dispatch an exact-head GitHub release",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "Automatic yes to the release preparation prompt"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunRelease(ctx, state, cmd.Bool("yes"))
		},
		Commands: []*cli.Command{
			{
				Name:   "publish",
				Usage:  "Gate and publish a prepared commit from GitHub Actions",
				Hidden: true,
				Flags:  []cli.Flag{&cli.StringFlag{Name: "commit", Required: true}},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return RunReleasePublish(ctx, state, cmd.String("commit"), realReleaseWaiter{})
				},
			},
		},
	}
}

// RunRelease prepares and pushes one release commit, then dispatches the workflow.
// Tags and GitHub releases are deliberately absent from this local mutation path.
func RunRelease(ctx context.Context, state *GlobalState, force bool) error {
	if err := IsInsideWorkTree(ctx, state); err != nil {
		return err
	}
	if err := requireCleanReleaseTree(ctx, state); err != nil {
		return err
	}
	if _, err := state.Runner.Run(ctx, "", nil, "gh", "auth", "status"); err != nil {
		return errors.New("github CLI is not authenticated; run 'gh auth login' or set GH_TOKEN")
	}
	if _, err := state.Runner.LookPath("git-cliff"); err != nil {
		return errors.New("git-cliff is not installed; run 'mise run tools' or install it via mise")
	}
	if _, err := state.Runner.LookPath("mise"); err != nil {
		return errors.New("mise is not installed; release validation cannot run")
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
	preparedTag, prepared := preparedReleaseTag(ctx, state)
	if head != upstream {
		if !prepared {
			return fmt.Errorf("release branch diverged: HEAD %s does not equal %s/%s %s", head, config.Remote, config.DefaultBranch, upstream)
		}
		parent, parentErr := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "HEAD^")
		if parentErr != nil || strings.TrimSpace(parent) != upstream {
			return fmt.Errorf("prepared release commit is not directly ahead of %s/%s", config.Remote, config.DefaultBranch)
		}
	}

	if prepared {
		if head != upstream {
			if pushErr := pushPreparedCommit(ctx, state, config, head); pushErr != nil {
				return pushErr
			}
		}
		return dispatchPreparedRelease(ctx, state, config, head, preparedTag)
	}

	bumped, current, err := calculateReleaseVersion(ctx, state)
	if err != nil {
		return err
	}
	if bumped == current {
		_, _ = fmt.Fprintf(state.Stdout, "No new conventional commits since %s. Nothing to release.\n", current)
		return nil
	}
	_, _ = fmt.Fprintf(state.Stdout, "Current version: %s\nNext version:    %s\n", yellow(current), green(bumped))
	if !force && !confirmRelease(state.Stdin, state.Stdout, fmt.Sprintf("Prepare %s for CI-owned publication? [y/N]: ", bumped)) {
		_, _ = fmt.Fprintln(state.Stdout, "Release canceled.")
		return nil
	}
	snapshot, err := snapshotReleaseFiles()
	if err != nil {
		return err
	}
	if writeErr := writeReleaseVersion(bumped); writeErr != nil {
		return snapshot.restore(writeErr)
	}
	if _, cliffErr := state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--bump", "-o", "CHANGELOG.md"); cliffErr != nil {
		return snapshot.restore(fmt.Errorf("failed to generate CHANGELOG.md: %w", cliffErr))
	}
	if validationErr := validatePreparedRelease(ctx, state); validationErr != nil {
		return snapshot.restore(validationErr)
	}
	if _, stageErr := state.Runner.Run(ctx, "", nil, "git", "add", "CHANGELOG.md", "dot/version.go"); stageErr != nil {
		return rollbackStagedRelease(ctx, state, snapshot, fmt.Errorf("failed to stage release files: %w", stageErr))
	}
	if _, commitErr := state.Runner.Run(ctx, "", nil, "git", "commit", "-m", "chore(release): "+bumped); commitErr != nil {
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
	return dispatchPreparedRelease(ctx, state, config, head, bumped)
}

type releaseFileSnapshot struct {
	version       []byte
	changelog     []byte
	versionMode   os.FileMode
	changelogMode os.FileMode
}

func snapshotReleaseFiles() (releaseFileSnapshot, error) {
	version, versionMode, err := snapshotReleaseFile("dot/version.go")
	if err != nil {
		return releaseFileSnapshot{}, err
	}
	changelog, changelogMode, err := snapshotReleaseFile("CHANGELOG.md")
	if err != nil {
		return releaseFileSnapshot{}, err
	}
	return releaseFileSnapshot{version: version, changelog: changelog, versionMode: versionMode, changelogMode: changelogMode}, nil
}

func snapshotReleaseFile(path string) ([]byte, os.FileMode, error) {
	content, err := os.ReadFile(path) //nolint:gosec // repository-owned fixed paths
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
	if err := os.WriteFile("dot/version.go", snapshot.version, snapshot.versionMode); err != nil { //nolint:gosec // repository-owned fixed path
		restoreErrors = append(restoreErrors, fmt.Errorf("failed to restore dot/version.go: %w", err))
	}
	if err := os.WriteFile("CHANGELOG.md", snapshot.changelog, snapshot.changelogMode); err != nil { //nolint:gosec // repository-owned fixed path
		restoreErrors = append(restoreErrors, fmt.Errorf("failed to restore CHANGELOG.md: %w", err))
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

func preparedReleaseTag(ctx context.Context, state *GlobalState) (string, bool) {
	subject, err := state.Runner.Run(ctx, "", nil, "git", "log", "-1", "--pretty=%s")
	if err != nil {
		return "", false
	}
	tag := strings.TrimPrefix(strings.TrimSpace(subject), "chore(release): ")
	if !semverTagPattern.MatchString(tag) {
		return "", false
	}
	version, err := readReleaseVersion()
	return tag, err == nil && tag == "v"+version
}

func calculateReleaseVersion(ctx context.Context, state *GlobalState) (string, string, error) {
	bumped, err := state.Runner.Run(ctx, "", nil, "git-cliff", "--config", "dot_config/git-cliff/cliff.toml", "--bumped-version")
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

func readReleaseVersion() (string, error) {
	content, err := os.ReadFile("dot/version.go")
	if err != nil {
		return "", fmt.Errorf("failed to read version.go: %w", err)
	}
	matches := regexp.MustCompile(`(?m)^var Version = "([^"\r\n]*)"$`).FindAllSubmatch(content, -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("dot/version.go must contain exactly one expected version assignment; found %d", len(matches))
	}
	return string(matches[0][1]), nil
}

func writeReleaseVersion(tag string) error {
	version, err := readReleaseVersion()
	if err != nil {
		return err
	}
	content, err := os.ReadFile("dot/version.go")
	if err != nil {
		return fmt.Errorf("failed to read version.go: %w", err)
	}
	old := `var Version = "` + version + `"`
	updated := strings.Replace(string(content), old, `var Version = "`+strings.TrimPrefix(tag, "v")+`"`, 1)
	if err := os.WriteFile("dot/version.go", []byte(updated), 0o644); err != nil { //nolint:gosec // repository-owned fixed path
		return fmt.Errorf("failed to write version.go: %w", err)
	}
	return nil
}

func validatePreparedRelease(ctx context.Context, state *GlobalState) error {
	for _, task := range []string{"format", "check", "test"} {
		_, _ = fmt.Fprintf(state.Stdout, "Running %s...\n", task)
		if err := state.Runner.RunInteractive(ctx, "", "mise", "run", task); err != nil {
			return fmt.Errorf("project %s failed: %w", task, err)
		}
	}
	status, err := state.Runner.Run(ctx, "", nil, "git", "status", "--porcelain=v1", "-z", "--untracked-files=all")
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

func dispatchPreparedRelease(ctx context.Context, state *GlobalState, config ReleaseConfig, commit, tag string) error {
	output, err := state.Runner.Run(ctx, "", nil, "gh", "workflow", "run", config.Workflow, "--ref", config.DefaultBranch, "-f", "commit="+commit)
	if err != nil {
		return fmt.Errorf("prepared commit %s was pushed but release workflow dispatch failed: %w", commit, err)
	}
	url := strings.TrimSpace(output)
	if url == "" {
		repoURL, viewErr := state.Runner.Run(ctx, "", nil, "gh", "repo", "view", "--json", "url", "--jq", ".url")
		if viewErr == nil {
			url = strings.TrimSpace(repoURL) + "/actions/workflows/" + config.Workflow
		}
	}
	_, _ = fmt.Fprintf(state.Stdout, "%s Prepared %s at %s. CI-owned publication: %s\n", passIcon, tag, commit, url)
	return nil
}

type releaseWaiter interface {
	Wait(context.Context, time.Duration) error
	Now() time.Time
}

type realReleaseWaiter struct{}

func (realReleaseWaiter) Wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (realReleaseWaiter) Now() time.Time { return time.Now() }

type releaseWorkflowRun struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"headSha"`
	URL        string `json:"url"`
	DatabaseID int64  `json:"databaseId"`
}

func RunReleasePublish(ctx context.Context, state *GlobalState, commit string, waiter releaseWaiter) error {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return errors.New("release publication is restricted to GitHub Actions")
	}
	if !fullCommitPattern.MatchString(commit) {
		return fmt.Errorf("invalid release commit %q", commit)
	}
	head, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(head) != commit {
		return fmt.Errorf("checked-out HEAD does not equal prepared commit %s", commit)
	}
	ciURL, err := waitForExactHeadCI(ctx, state, commit, waiter)
	if err != nil {
		return err
	}
	version, err := readReleaseVersion()
	if err != nil {
		return err
	}
	tag := "v" + version
	if !semverTagPattern.MatchString(tag) {
		return fmt.Errorf("prepared version %q is not semantic", tag)
	}
	if tagErr := ensureImmutableTag(ctx, state, state.Config.Release.Remote, tag, commit); tagErr != nil {
		return tagErr
	}
	releaseURL, err := ensureGitHubRelease(ctx, state, tag)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(state.Stdout, "%s Published %s from exact-head CI %s: %s\n", passIcon, tag, ciURL, releaseURL)
	return nil
}

func waitForExactHeadCI(ctx context.Context, state *GlobalState, commit string, waiter releaseWaiter) (string, error) {
	deadline := waiter.Now().Add(releaseGateTimeout)
	for {
		raw, err := state.Runner.Run(ctx, "", nil, "gh", "run", "list", "--workflow", releaseCIWorkflow, "--commit", commit, "--event", "push", "--limit", "2", "--json", "databaseId,status,conclusion,url,headSha")
		if err != nil {
			return "", fmt.Errorf("failed to resolve exact-head CI: %w", err)
		}
		var runs []releaseWorkflowRun
		if err := json.Unmarshal([]byte(raw), &runs); err != nil {
			return "", fmt.Errorf("invalid exact-head CI response: %w", err)
		}
		if len(runs) > 1 {
			return "", fmt.Errorf("ambiguous exact-head CI: found %d runs for %s", len(runs), commit)
		}
		if len(runs) == 1 {
			run := runs[0]
			if run.HeadSHA != commit {
				return "", fmt.Errorf("exact-head CI returned mismatched commit %s", run.HeadSHA)
			}
			if run.Status == "completed" {
				if run.Conclusion != "success" {
					return "", fmt.Errorf("exact-head CI %d concluded %s", run.DatabaseID, run.Conclusion)
				}
				return run.URL, nil
			}
		}
		if !waiter.Now().Before(deadline) {
			return "", fmt.Errorf("timed out waiting for exact-head CI for %s", commit)
		}
		if err := waiter.Wait(ctx, releasePollInterval); err != nil {
			return "", fmt.Errorf("waiting for exact-head CI: %w", err)
		}
	}
}

func remoteAnnotatedTag(ctx context.Context, state *GlobalState, remote, tag string) (string, bool, error) {
	output, err := state.Runner.Run(ctx, "", nil, "git", "ls-remote", "--tags", remote, "refs/tags/"+tag, "refs/tags/"+tag+"^{}")
	if err != nil {
		return "", false, fmt.Errorf("failed to inspect remote tag %s: %w", tag, err)
	}
	if strings.TrimSpace(output) == "" {
		return "", false, nil
	}
	var direct, peeled string
	for line := range strings.Lines(output) {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return "", false, fmt.Errorf("malformed remote tag response for %s", tag)
		}
		if strings.HasSuffix(fields[1], "^{}") {
			peeled = fields[0]
		} else {
			direct = fields[0]
		}
	}
	if direct == "" || peeled == "" {
		return "", true, fmt.Errorf("remote tag %s exists but is not annotated", tag)
	}
	return peeled, true, nil
}

func ensureImmutableTag(ctx context.Context, state *GlobalState, remote, tag, commit string) error {
	peeled, exists, err := remoteAnnotatedTag(ctx, state, remote, tag)
	if err != nil {
		return err
	}
	if exists {
		if peeled != commit {
			return fmt.Errorf("immutable tag %s already targets %s, not %s", tag, peeled, commit)
		}
		return nil
	}
	if _, err := state.Runner.Run(ctx, "", nil, "git", "tag", "-a", tag, "-m", tag, commit); err != nil {
		return fmt.Errorf("failed to create annotated tag %s: %w", tag, err)
	}
	if err := state.Runner.RunInteractive(ctx, "", "git", "push", remote, "refs/tags/"+tag); err == nil {
		verifiedCommit, verifiedExists, inspectErr := remoteAnnotatedTag(ctx, state, remote, tag)
		if inspectErr != nil || !verifiedExists || verifiedCommit != commit {
			return fmt.Errorf("published tag %s failed immutable readback verification", tag)
		}
		return nil
	}
	peeled, exists, inspectErr := remoteAnnotatedTag(context.WithoutCancel(ctx), state, remote, tag)
	if inspectErr == nil && exists && peeled == commit {
		return nil
	}
	return fmt.Errorf("failed to publish immutable tag %s", tag)
}

func ensureGitHubRelease(ctx context.Context, state *GlobalState, tag string) (string, error) {
	view := func() (string, error) {
		return state.Runner.Run(ctx, "", nil, "gh", "release", "view", tag, "--json", "url", "--jq", ".url")
	}
	if url, err := view(); err == nil {
		return strings.TrimSpace(url), nil
	}
	if err := state.Runner.RunInteractive(ctx, "", "gh", "release", "create", tag, "--verify-tag", "--generate-notes", "--title", tag); err != nil {
		if url, viewErr := view(); viewErr == nil {
			return strings.TrimSpace(url), nil
		}
		return "", fmt.Errorf("failed to publish GitHub release %s: %w", tag, err)
	}
	url, err := view()
	if err != nil {
		return "", fmt.Errorf("release %s was created but its URL is unavailable: %w", tag, err)
	}
	return strings.TrimSpace(url), nil
}

func confirmRelease(stdin io.Reader, stdout io.Writer, msg string) bool {
	_, _ = fmt.Fprint(stdout, msg)
	reader := bufio.NewReader(stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
