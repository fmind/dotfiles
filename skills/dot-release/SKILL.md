---
name: dot-release
description: Cut a versioned release for fmind/dotfiles using the Go dot CLI release command (alias r). Use when shipping a new tagged version of dotfiles.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dot-release
  created: 2026-07-08
  updated: 2026-07-09
---

# Dotfiles Release

Run `dot release` (alias `dot r`) to turn Conventional Commits since the last tag into a versioned release. The command handles everything: compute next semver via `git-cliff`, update [version.go](file:///home/fmind/.local/share/chezmoi/dot/version.go) and [CHANGELOG.md](file:///home/fmind/.local/share/chezmoi/CHANGELOG.md), format, lint, test, commit, tag, push, and publish a GitHub release.

## Preconditions

- Clean working tree on `main`, synced with `origin` (stash or commit first).
- `gh` authenticated, `git-cliff` and `mise` installed.
- Commit history follows Conventional Commits.

## Usage

```bash
# Agent / non-interactive (skips confirmation prompt)
mise run r -- -y
```

## Gotchas

- Lint or test failures during `mise run check` abort the release before any commit is made.

## See Also

- [release](../release/SKILL.md) — Generic release process template.
- [conventional-commit](../conventional-commit/SKILL.md) — Commit grammar for changelog bumping.
