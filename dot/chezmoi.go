package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

// NewChezmoiCmd constructs the top-level chezmoi command group.
func NewChezmoiCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "chezmoi",
		Aliases: []string{"m"},
		Usage:   "Manage chezmoi configuration and clean up orphaned files",
		Commands: []*cli.Command{
			NewChezmoiCleanCmd(state),
		},
	}
}

// NewChezmoiCleanCmd constructs the chezmoi clean subcommand.
func NewChezmoiCleanCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "clean",
		Aliases: []string{"c"},
		Usage:   "Scan for previously managed chezmoi files and clean up unmanaged orphans",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Automatically answer yes to prompts and delete all orphaned files",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Prompt to delete all orphaned files",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			autoApprove := cmd.Bool("yes")
			interactive := cmd.Bool("interactive")
			return RunChezmoiClean(ctx, state, autoApprove, interactive)
		},
	}
}

// RunChezmoiClean implements the logic to identify and clean up orphaned dotfiles.
func RunChezmoiClean(ctx context.Context, state *GlobalState, autoApprove, interactive bool) error {
	// 1. Verify git and chezmoi are installed
	if err := checkGit(state); err != nil {
		return err
	}
	if _, err := state.Runner.LookPath("chezmoi"); err != nil {
		return ErrChezmoiNotInstalled
	}

	// 2. Determine chezmoi source path
	sourcePathBytes, err := state.Runner.Run(ctx, "", nil, "chezmoi", "source-path")
	if err != nil {
		return fmt.Errorf("failed to get chezmoi source path: %w", err)
	}
	chezmoiSourceDir := strings.TrimSpace(sourcePathBytes)
	if chezmoiSourceDir == "" {
		return errors.New("chezmoi source path is empty")
	}

	_, _ = fmt.Fprintf(state.Stdout, "Found chezmoi source directory: %s\n", chezmoiSourceDir)

	// 3. Get currently managed targets
	_, _ = fmt.Fprintln(state.Stdout, "Fetching currently managed chezmoi files...")
	managedBytes, err := state.Runner.Run(ctx, "", nil, "chezmoi", "managed")
	if err != nil {
		return fmt.Errorf("failed to run chezmoi managed: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	managedTargets := make(map[string]bool)
	for line := range strings.SplitSeq(managedBytes, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			managedTargets[resolveHome(homeDir, trimmed)] = true
		}
	}

	// 4. Retrieve deleted files from git history of chezmoi source dir
	_, _ = fmt.Fprintln(state.Stdout, "Scanning git history for deleted/moved files...")

	// --no-renames so a moved source file decomposes into delete(old)+add(new): git's
	// default rename detection would classify it as R (not D) and the old target left in
	// $HOME would be missed. Every candidate still passes the source-exists/managed/exists
	// guards below before backup, so broadening detection stays safe.

	// Historically deleted files
	logOut, err := state.Runner.Run(ctx, chezmoiSourceDir, nil, "git", "log", "--no-renames", "--diff-filter=D", "--name-only", "--pretty=format:")
	if err != nil {
		return fmt.Errorf("failed to run git log: %w", err)
	}

	// Unstaged deleted files
	unstagedOut, err := state.Runner.Run(ctx, chezmoiSourceDir, nil, "git", "diff", "--no-renames", "--name-only", "--diff-filter=D")
	if err != nil {
		return fmt.Errorf("failed to run git diff: %w", err)
	}

	// Staged deleted files
	stagedOut, err := state.Runner.Run(ctx, chezmoiSourceDir, nil, "git", "diff", "--cached", "--no-renames", "--name-only", "--diff-filter=D")
	if err != nil {
		return fmt.Errorf("failed to run git diff --cached: %w", err)
	}

	deletedSet := make(map[string]bool)
	collectDeleted := func(output string) {
		for line := range strings.SplitSeq(output, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				deletedSet[trimmed] = true
			}
		}
	}
	collectDeleted(logOut)
	collectDeleted(unstagedOut)
	collectDeleted(stagedOut)

	// Filter and find orphans
	var orphans []string

	for relPath := range deletedSet {
		// Ignore check
		if shouldIgnore(state.Config.ChezmoiClean, relPath) {
			continue
		}

		// Safety check: if file currently exists in source repo, it was re-added. Skip it.
		sourceAbsPath := filepath.Join(chezmoiSourceDir, relPath)
		if _, err := os.Stat(sourceAbsPath); err == nil {
			continue
		}

		// Map to target path using the temporary touch trick
		targetPath, err := getChezmoiTargetPath(ctx, state, chezmoiSourceDir, relPath)
		if err != nil || targetPath == "" {
			continue
		}

		// Resolve absolute target path
		targetAbs := resolveHome(homeDir, targetPath)

		// Check if it exists in the home directory
		if _, err := os.Stat(targetAbs); os.IsNotExist(err) {
			continue
		}

		// Check if it's currently managed by chezmoi
		if managedTargets[targetAbs] {
			continue
		}

		// It exists in home and is NOT managed anymore! It is an orphan.
		orphans = append(orphans, targetAbs)
	}

	if len(orphans) == 0 {
		_, _ = fmt.Fprintln(state.Stdout, green("✓ No orphaned files found in your home directory."))
		return nil
	}

	// 5. Present orphans and ask user
	_, _ = fmt.Fprintf(state.Stdout, "\n%s\n", bold(yellow(fmt.Sprintf("Detected %d orphaned file(s) in home directory:", len(orphans)))))
	for _, o := range orphans {
		_, _ = fmt.Fprintf(state.Stdout, "  ▶ %s\n", o)
	}
	_, _ = fmt.Fprintln(state.Stdout)

	if autoApprove {
		return backupOrphans(state, orphans)
	}

	if interactive {
		if confirm(state.Stdin, state.Stdout, "Do you want to move all orphaned files to a backup directory? [y/N]: ") {
			return backupOrphans(state, orphans)
		}
		_, _ = fmt.Fprintln(state.Stdout, "Clean up canceled. No files were modified.")
		return nil
	}

	// Preview mode: neither --yes nor --interactive was given. Guide the next step.
	_, _ = fmt.Fprintln(state.Stdout, "Re-run with --yes to back up and remove all, or --interactive to confirm.")
	return nil
}

// backupOrphans moves each orphaned file into a timestamped backup directory so the
// operation stays recoverable instead of permanently deleting user data.
func backupOrphans(state *GlobalState, orphans []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}
	backupDir := filepath.Join(home, ".cache", "dot", "chezmoi-clean", time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	failed := 0
	for _, o := range orphans {
		rel, relErr := filepath.Rel(home, o)
		if relErr != nil || strings.HasPrefix(rel, "..") {
			// Outside home: fall back to the full path (separators sanitized) so two
			// orphans sharing a basename cannot collide and overwrite each other's backup.
			rel = strings.TrimPrefix(filepath.ToSlash(o), "/")
		}
		dest := uniqueBackupPath(filepath.Join(backupDir, rel))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			_, _ = fmt.Fprintf(state.Stderr, "Error preparing backup for %s: %v\n", o, err)
			failed++
			continue
		}
		if err := os.Rename(o, dest); err != nil {
			_, _ = fmt.Fprintf(state.Stderr, "Error backing up %s: %v\n", o, err)
			failed++
			continue
		}
		_, _ = fmt.Fprintf(state.Stdout, "Backed up: %s\n", o)
	}
	// Report failure loudly rather than printing a false "complete" while orphans remain
	// in place (which would also mask the problem from scripts keying off the exit code).
	if failed > 0 {
		return fmt.Errorf("failed to back up %d of %d orphaned file(s); see errors above", failed, len(orphans))
	}
	_, _ = fmt.Fprintf(state.Stdout, "%s\n", green("✓ Clean up complete. Backups saved to "+backupDir))
	return nil
}

// uniqueBackupPath returns dest, or dest with a numeric suffix if a file already
// exists there, so a backup never silently overwrites an earlier one.
func uniqueBackupPath(dest string) string {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest
	}
	ext := filepath.Ext(dest)
	base := strings.TrimSuffix(dest, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s.%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// resolveHome resolves p against homeDir when it is relative, returning a cleaned
// absolute path. Absolute inputs are only cleaned.
func resolveHome(homeDir, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(homeDir, p))
}

// shouldIgnore returns true if the path is repository metadata or internal tooling configs.
func shouldIgnore(cfg ChezmoiCleanConfig, path string) bool {
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) == 0 {
		return true
	}

	for _, prefix := range cfg.IgnoredPrefixes {
		if parts[0] == prefix {
			return true
		}
	}

	base := filepath.Base(path)
	for _, f := range cfg.IgnoredFiles {
		if base == f {
			return true
		}
	}

	// Chezmoi internal scripts
	if strings.HasPrefix(base, "run_once_") || strings.HasPrefix(base, "run_onchange_") || strings.HasPrefix(base, "run_before_") {
		return true
	}

	return false
}

// ChezmoiCleanConfig represents the configuration for the chezmoi-clean command.
type ChezmoiCleanConfig struct {
	IgnoredPrefixes []string `yaml:"ignored_prefixes"`
	IgnoredFiles    []string `yaml:"ignored_files"`
}

func defaultChezmoiCleanConfig() ChezmoiCleanConfig {
	return ChezmoiCleanConfig{
		IgnoredPrefixes: []string{
			".git",
			".github",
			".agents",
			".copilot",
			".claude",
			".gemini",
			"skills",
			"dot",
			"dot_agents",
			"dot_claude",
			"dot_gemini",
			"dot_copilot",
		},
		IgnoredFiles: []string{
			"README.md",
			"LICENSE",
			"AGENTS.md",
			"go.work",
			"go.work.sum",
			"go.mod",
			"go.sum",
			"dprint.json",
			"lefthook.yml",
			"mise.toml",
			"install.sh",
			"skill-lock.json",
		},
	}
}

func cleanupChezmoiProbe(sourcePath string, createdDirs []string) error {
	var cleanupErrors []error
	if sourcePath != "" {
		if err := os.Remove(sourcePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove chezmoi probe %s: %w", sourcePath, err))
		}
	}
	for _, dir := range createdDirs {
		if err := os.Remove(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove chezmoi probe directory %s: %w", dir, err))
		}
	}
	return errors.Join(cleanupErrors...)
}

// getChezmoiTargetPath maps a relative source path to its target home path using a temporary touch.
func getChezmoiTargetPath(ctx context.Context, state *GlobalState, chezmoiSourceDir, relPath string) (targetPath string, resultErr error) {
	sourceAbsPath := filepath.Join(chezmoiSourceDir, relPath)
	parentDir := filepath.Dir(sourceAbsPath)

	// Determine which parent directories need to be created
	var createdDirs []string
	currDir := parentDir
	for currDir != chezmoiSourceDir && currDir != "." && currDir != "/" {
		if _, err := os.Stat(currDir); os.IsNotExist(err) {
			createdDirs = append(createdDirs, currDir)
			currDir = filepath.Dir(currDir)
		} else {
			break
		}
	}

	// Create directories if they do not exist
	if len(createdDirs) > 0 {
		err := os.MkdirAll(parentDir, 0o755)
		if err != nil {
			return "", err
		}
	}

	// Touch the probe file with O_EXCL so the helper can never truncate/clobber a real
	// source file that unexpectedly exists at this path; the deferred cleanup below only
	// removes what we create here, keeping the function safe regardless of caller ordering.
	f, err := os.OpenFile(sourceAbsPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", errors.Join(err, cleanupChezmoiProbe("", createdDirs))
	}
	if closeErr := f.Close(); closeErr != nil {
		return "", errors.Join(
			fmt.Errorf("failed to close chezmoi probe %s: %w", sourceAbsPath, closeErr),
			cleanupChezmoiProbe(sourceAbsPath, createdDirs),
		)
	}

	// Defer cleanup of touched files and folders
	defer func() {
		if cleanupErr := cleanupChezmoiProbe(sourceAbsPath, createdDirs); cleanupErr != nil {
			resultErr = errors.Join(resultErr, cleanupErr)
		}
	}()

	// Query target path
	targetPathBytes, err := state.Runner.Run(ctx, "", nil, "chezmoi", "target-path", sourceAbsPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(targetPathBytes), nil
}
