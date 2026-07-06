package dot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func checkGit(state *GlobalState) error {
	if _, err := state.Runner.LookPath("git"); err != nil {
		return ErrGitNotInstalled
	}
	return nil
}

// IsInsideWorkTree checks if the current directory is inside a Git working tree.
func IsInsideWorkTree(ctx context.Context, state *GlobalState) error {
	if err := checkGit(state); err != nil {
		return err
	}
	_, err := state.Runner.Run(ctx, "", nil, "git", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return ErrNotGitRepository
	}
	return nil
}

// buildExcludePathspecs returns pathspecs anchored at the repository root (":/" and the
// "top" exclude magic) so the diff covers the whole repo even when dot is invoked from a
// subdirectory — matching the scope of `git commit`/`gh pr create`, which act on the whole
// index/branch regardless of cwd. A cwd-relative "." would silently scope the diff to the
// current subtree and desync the generated commit/PR summary from what is actually committed.
func buildExcludePathspecs(excludes []string) []string {
	pathspecs := make([]string, 0, 1+len(excludes))
	pathspecs = append(pathspecs, ":/")
	for _, pat := range excludes {
		pathspecs = append(pathspecs, ":(exclude,top)"+pat)
	}
	return pathspecs
}

// GetCachedDiff returns the staged git diff, or empty string if no staged changes.
func GetCachedDiff(ctx context.Context, state *GlobalState) (string, error) {
	if err := checkGit(state); err != nil {
		return "", err
	}
	pathspecs := buildExcludePathspecs(state.Config.Commit.ExcludeDiff)
	diffArgs := append([]string{"diff", "--cached", "--"}, pathspecs...)
	return state.Runner.Run(ctx, "", nil, "git", diffArgs...)
}

// GetUnstagedDiff returns the unstaged git diff.
func GetUnstagedDiff(ctx context.Context, state *GlobalState) (string, error) {
	if err := checkGit(state); err != nil {
		return "", err
	}
	pathspecs := buildExcludePathspecs(state.Config.Commit.ExcludeDiff)
	diffArgs := append([]string{"diff", "--"}, pathspecs...)
	return state.Runner.Run(ctx, "", nil, "git", diffArgs...)
}

// GetBaseDiff returns the git diff against a base branch. It first attempts to diff with baseBranch... (merge base)
// and falls back to diffing directly with baseBranch if that fails.
func GetBaseDiff(ctx context.Context, state *GlobalState, baseBranch string) (string, error) {
	if err := checkGit(state); err != nil {
		return "", err
	}
	pathspecs := buildExcludePathspecs(state.Config.Commit.ExcludeDiff)
	// The "--" separator disambiguates revisions from pathspecs, matching the other
	// diff helpers. Without it, a path named like the base branch makes git abort with
	// "ambiguous argument ... both a revision and a filename" on the fallback form.
	diffArgs := append([]string{"diff", baseBranch + "...", "--"}, pathspecs...)
	diff, err := state.Runner.Run(ctx, "", nil, "git", diffArgs...)
	if err != nil {
		diffArgs = append([]string{"diff", baseBranch, "--"}, pathspecs...)
		diff, err = state.Runner.Run(ctx, "", nil, "git", diffArgs...)
		if err != nil {
			return "", fmt.Errorf("failed to get git diff against %s: %w", baseBranch, err)
		}
	}
	return diff, nil
}

// findGitRepos returns the paths of all git repositories located directly under the
// configured pull directories. Unreadable directories are logged and skipped.
func findGitRepos(state *GlobalState) []string {
	var repos []string
	for _, dir := range state.Config.Pull.Directories {
		absDir := ExpandPath(dir)
		entries, err := os.ReadDir(absDir)
		if err != nil {
			state.Logger.Warn("Skipping directory (could not read)", "path", absDir, "error", err)
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			repoPath := filepath.Join(absDir, entry.Name())
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
				repos = append(repos, repoPath)
			}
		}
	}
	return repos
}

// repoBranch resolves the current branch of a repository, falling back to the short
// commit hash when HEAD is detached.
func repoBranch(ctx context.Context, state *GlobalState, path string) (string, error) {
	branch, err := state.Runner.Run(ctx, path, nil, "git", "branch", "--show-current")
	if err != nil {
		return "", err
	}
	if branch = strings.TrimSpace(branch); branch != "" {
		return branch, nil
	}
	shortHead, err := state.Runner.Run(ctx, path, nil, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown", nil
	}
	return strings.TrimSpace(shortHead), nil
}
