---
name: dot-release
description: Cut a versioned release for fmind/dotfiles using the Go dot CLI release command (alias r). Use when shipping a new tagged version of dotfiles.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dot-release
  created: 2026-07-08
  updated: 2026-07-08
---

# Dotfiles Release Command

Execute the unified Go CLI release command `dot release` (alias `dot r`) to turn Conventional Commits since the last tag into a versioned release (updates `CHANGELOG.md`, bumps `dot/version.go`, commits, tags, pushes, and publishes a GitHub release).

## Preconditions

- Clean working tree on the default branch (`main`), synced with `origin`.
- GitHub CLI (`gh`) is authenticated.
- Toolchain has `git-cliff` and `mise` installed.
- History follows Conventional Commits.

## Workflow

1. **Verify workspace state**: The command verifies that git is clean, `gh` is authenticated, and `git-cliff` is available.
1. **Compute next version**: It runs `git-cliff` to determine the next version (e.g., `v1.0.1`) based on the commits since the last tag.
1. **Check for changes**: If the computed version is identical to the current tag, it prints a message and stops (no release needed).
1. **Confirm with user**: Prompts `Proceed with releasing vX.Y.Z? [y/N]:` unless the `--yes` / `-y` flag is provided.
1. **Update version files**:
   - Reads [version.go](file:///home/fmind/.local/share/chezmoi/dot/version.go) and replaces `var Version = "..."` with the new version number.
1. **Generate Changelog**:
   - Updates [CHANGELOG.md](file:///home/fmind/.local/share/chezmoi/CHANGELOG.md) using `git-cliff`.
1. **Format and Verify**:
   - Runs `mise run format` and `mise run check` to ensure all linters, formatting, and unit tests pass before committing.
1. **Commit and Tag**:
   - Stages and commits `CHANGELOG.md` and `dot/version.go` with message `chore(release): vX.Y.Z`.
   - Creates an annotated tag `vX.Y.Z`.
1. **Push**:
   - Pushes the branch and the new tag to `origin`.
1. **Publish GitHub Release**:
   - Generates release notes from the latest changelog section and calls `gh release create` to publish the release on GitHub.

## Usage

Run via `mise`:

```bash
# Run the release workflow
mr r

# Run the release workflow and skip prompts (useful in scripts or agents)
mr r -- -y
```

Or run via the compiled `dot` binary:

```bash
# Run the release command
dot release

# Or using the alias
dot r -y
```

## Gotchas

- **Clean Working Tree**: The release process will fail if you have uncommitted or staged changes. Use `git stash` if needed.
- **Git Hooks**: The command runs `mise run check` before committing, so any lint or test failures will abort the release automatically.

## Documentation

- [release](../release/SKILL.md) — The parent release process template.
- [conventional-commit](../conventional-commit/SKILL.md) — The commit grammar required for changelog bumping.
