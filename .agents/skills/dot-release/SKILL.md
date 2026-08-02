---
name: dot-release
description: Prepare, tag, and push a versioned fmind/dotfiles release locally to trigger GitHub Actions CD publication.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/.agents/skills/dot-release
  created: 2026-07-08
  updated: 2026-08-02
---

# Dotfiles Release

Run `dot release` (alias `dot r`) to turn Conventional Commits since the last tag into a release commit and tag. The local command fetches `origin`, requires a clean `main` with `HEAD == origin/main`, computes the next semver via `git-cliff`, updates `dot/version.go` and `CHANGELOG.md`, runs the complete local gate (`format`, `check`, `test`), creates the commit and annotated tag, pushes `main` and `refs/tags/v*` to `origin`, and automatically reapplies the updated local `dot` binary. GitHub Actions (`cd.yml`) triggers on tag push to create the GitHub release.

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

- Lint or test failures during local gate abort the release before any commit or tag is created.
- Pushing a release tag automatically triggers the `.github/workflows/cd.yml` workflow on GitHub Actions to publish the release.

## See Also

- [release](../../../skills/release/SKILL.md) — Generic release process template.
- [conventional-commit](../../../skills/conventional-commit/SKILL.md) — Commit grammar for changelog bumping.
