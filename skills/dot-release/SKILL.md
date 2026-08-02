---
name: dot-release
description: Prepare a versioned fmind/dotfiles release locally, then verify the CI-owned exact-head tag and GitHub publication. Use when shipping a new dotfiles version without creating tags or releases from a workstation.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dot-release
  created: 2026-07-08
  updated: 2026-07-31
---

# Dotfiles Release

Run `dot release` (alias `dot r`) to turn Conventional Commits since the last tag into a prepared release commit. The local command fetches `origin`, requires a clean `main` with `HEAD == origin/main`, computes the next semver via `git-cliff`, updates `dot/version.go` and `CHANGELOG.md`, runs the complete local gate, commits, pushes only that commit, and dispatches `.github/workflows/cd.yml`. GitHub Actions then waits for the exact release commit's CI run before creating an annotated immutable tag and publishing the release.

## Preconditions

- Clean working tree on `main`; the command fetches and proves equality with `origin/main` before mutation.
- `gh` authenticated, `git-cliff` and `mise` installed.
- Commit history follows Conventional Commits.

## Usage

```bash
# Agent / non-interactive (skips confirmation prompt)
mise run release -- -y
```

## Gotchas

- Lint or test failures during `mise run check` abort the release before any commit is made.
- A push or workflow-dispatch interruption is resumable: rerun the same command and it reuses the prepared release commit instead of bumping again.
- Never create or move the tag manually. Follow the reported workflow URL, then verify the exact commit, annotated tag, and public GitHub release after the workflow succeeds.

## See Also

- [release](../release/SKILL.md) — Generic release process template.
- [conventional-commit](../conventional-commit/SKILL.md) — Commit grammar for changelog bumping.
