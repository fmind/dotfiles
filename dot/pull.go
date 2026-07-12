package dot

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

// RepoResult stores the details and status of a single git repository pull operation.
type RepoResult struct {
	Err        error
	PushErr    error
	RepoName   string
	ParentDir  string
	Branch     string
	Commits    int
	Ahead      int
	Dirty      bool
	NoUpstream bool
	Pushed     bool
}

// NewPullCmd constructs the top-level pull command.
func NewPullCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "pull",
		Aliases: []string{"p"},
		Usage:   "Concurrently pull all Git repositories listed in ~/.config/dot.yaml",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "push",
				Aliases: []string{"P"},
				Usage:   "Also push unpushed commits for clean repositories with an upstream",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunPull(ctx, state, cmd.Bool("push"))
		},
	}
}

// RunPull scans and concurrently updates all git repositories in configured pull directories.
// When push is true, clean repositories ahead of their upstream are also pushed.
func RunPull(ctx context.Context, state *GlobalState, push bool) error {
	if err := checkGit(state); err != nil {
		return err
	}

	reposToPull := findGitRepos(state)
	if len(reposToPull) == 0 {
		_, _ = fmt.Fprintln(state.Stdout, "No git repositories found in configured pull directories.")
		return nil
	}

	_, _ = fmt.Fprintf(state.Stdout, "%s\n\n", bold(fmt.Sprintf("Scanning and pulling %d repositories...", len(reposToPull))))

	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	results := make([]RepoResult, len(reposToPull))
	for i, path := range reposToPull {
		g.Go(func() error {
			results[i] = pullRepo(groupCtx, state, path, push)
			return nil
		})
	}

	_ = g.Wait()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	slices.SortFunc(results, func(a, b RepoResult) int {
		return strings.Compare(
			filepath.Join(a.ParentDir, a.RepoName),
			filepath.Join(b.ParentDir, b.RepoName),
		)
	})

	failedCount := 0
	for _, res := range results {
		dirtyFlag := ""
		if res.Dirty {
			dirtyFlag = " [" + yellow("dirty") + "]"
		}
		if res.NoUpstream {
			dirtyFlag += " [" + dim("no upstream") + "]"
		}

		parentBase := filepath.Base(res.ParentDir)

		_, _ = fmt.Fprintf(state.Stdout, "▶ %s [%s]%s\n", bold(parentBase+"/"+res.RepoName), green(res.Branch), dirtyFlag)

		switch {
		case res.Err != nil:
			_, _ = fmt.Fprintf(state.Stdout, "  %s %s\n\n", red("✗ Pull failed:"), strings.ReplaceAll(res.Err.Error(), "\n", "\n  "))
			failedCount++
		case res.NoUpstream:
			_, _ = fmt.Fprintf(state.Stdout, "  %s\n\n", dim("∅ skipped (no upstream)"))
		default:
			commitStr := "up to date"
			if res.Commits > 0 {
				commitStr = fmt.Sprintf("pulled %d commit(s)", res.Commits)
			}
			_, _ = fmt.Fprintf(state.Stdout, "  %s (%s)\n", green("✓ Pull successful"), commitStr)
			if renderPush(state, res) {
				failedCount++
			}
			_, _ = fmt.Fprintln(state.Stdout)
		}
	}

	if failedCount > 0 {
		return fmt.Errorf("failed to pull %d repositories", failedCount)
	}

	_, _ = fmt.Fprintln(state.Stdout, bold(green("✓ All repositories pulled successfully.")))
	return nil
}

// renderPush prints the push outcome for a repository and reports whether it failed.
func renderPush(state *GlobalState, res RepoResult) bool {
	switch {
	case res.PushErr != nil:
		_, _ = fmt.Fprintf(state.Stdout, "  %s %s\n", red("✗ Push failed:"), strings.ReplaceAll(res.PushErr.Error(), "\n", "\n  "))
		return true
	case res.Pushed:
		_, _ = fmt.Fprintf(state.Stdout, "  %s\n", green(fmt.Sprintf("↑ pushed %d commit(s)", res.Ahead)))
	case res.Ahead > 0:
		_, _ = fmt.Fprintf(state.Stdout, "  %s\n", yellow(fmt.Sprintf("↑ %d unpushed", res.Ahead)))
	}
	return false
}

// pullTimeout bounds each repository's fetch/pull so one wedged remote can't hang the whole run.
const pullTimeout = 2 * time.Minute

// pullRepo checks branch/dirty state, then fetches and pulls a single repository.
// Repositories without an upstream are skipped rather than reported as failures.
// When push is true, a clean repository ahead of its upstream is also pushed.
func pullRepo(ctx context.Context, state *GlobalState, path string, push bool) RepoResult {
	ctx, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()

	res := RepoResult{
		RepoName:  filepath.Base(path),
		ParentDir: filepath.Dir(path),
	}

	branch, err := repoBranch(ctx, state, path)
	if err != nil {
		res.Err = err
		return res
	}
	res.Branch = branch

	status, err := state.Runner.Run(ctx, path, nil, "git", "status", "--porcelain")
	if err != nil {
		res.Err = fmt.Errorf("failed to check worktree status: %w", err)
		return res
	}
	res.Dirty = strings.TrimSpace(status) != ""

	if _, fetchErr := state.Runner.Run(ctx, path, nil, "git", "fetch", "--prune"); fetchErr != nil {
		if ctx.Err() != nil {
			res.Err = fmt.Errorf("fetch timed out: %w", ctx.Err())
			return res
		}
		// A repository without an upstream often has no default remote either, so
		// `git fetch` fails before the later rev-list check can classify it as a skip.
		if _, upstreamErr := state.Runner.Run(ctx, path, nil, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); upstreamErr != nil {
			res.NoUpstream = true
			return res
		}
		res.Err = fmt.Errorf("failed to fetch repository: %w", fetchErr)
		return res
	}

	// A missing upstream makes rev-list (and git pull) fail; treat it as a skip, not a
	// failure. A canceled context, however, means a real timeout/interrupt, not a missing
	// upstream, so surface it as an error instead of silently reporting "no upstream".
	behind, err := state.Runner.Run(ctx, path, nil, "git", "rev-list", "--count", "HEAD..@{u}")
	if err != nil {
		if ctx.Err() != nil {
			res.Err = fmt.Errorf("timed out: %w", ctx.Err())
		} else {
			res.NoUpstream = true
		}
		return res
	}
	cnt, parseErr := strconv.Atoi(strings.TrimSpace(behind))
	if parseErr != nil {
		res.Err = fmt.Errorf("failed to parse behind count %q: %w", strings.TrimSpace(behind), parseErr)
		return res
	}
	res.Commits = cnt

	// --ff-only so an unattended bulk pull never creates a surprise merge commit or leaves
	// a repo half-merged: a diverged branch fails cleanly and is reported as a pull failure.
	if _, perr := state.Runner.Run(ctx, path, nil, "git", "pull", "--ff-only"); perr != nil {
		res.Err = perr
		return res
	}

	ahead, err := state.Runner.Run(ctx, path, nil, "git", "rev-list", "--count", "@{u}..HEAD")
	if err != nil {
		res.Err = fmt.Errorf("failed to determine ahead count: %w", err)
		return res
	}
	aheadCount, parseErr := strconv.Atoi(strings.TrimSpace(ahead))
	if parseErr != nil {
		res.Err = fmt.Errorf("failed to parse ahead count %q: %w", strings.TrimSpace(ahead), parseErr)
		return res
	}
	res.Ahead = aheadCount

	if push {
		pushRepo(ctx, state, path, &res)
	}

	return res
}

// pushRepo pushes a repository when it is clean and holds unpushed commits. Dirty
// working trees are left untouched so a push never races an in-progress edit.
func pushRepo(ctx context.Context, state *GlobalState, path string, res *RepoResult) {
	if res.Ahead == 0 || res.Dirty {
		return
	}
	if _, err := state.Runner.Run(ctx, path, nil, "git", "push"); err != nil {
		res.PushErr = err
		return
	}
	res.Pushed = true
}

// PullConfig represents the configuration for pulling Git repositories.
type PullConfig struct {
	Directories []string `yaml:"directories"`
}

func defaultPullConfig() PullConfig {
	return PullConfig{
		Directories: []string{
			"~/internals",
			"~/externals",
			"~/workspaces",
		},
	}
}
