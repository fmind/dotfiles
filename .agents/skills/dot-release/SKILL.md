---
name: dot-release
description: Run dot release (or mise run release) to gate, bump, tag, and push a fmind/dot release that cd.yml publishes. Use when cutting a release of this repository.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/.agents/skills/dot-release
  created: "2026-07-08"
  updated: "2026-09-03"
---

# Dot Release

`dot release` (alias `dot r`, wrapped as `mise run release`) turns the Conventional Commits since the last tag into a release commit and tag of this repository; the generic process and the post-publication verification live in [release](../../../skills/release/SKILL.md).

## Commands

```bash
mise run release -- -y   # non-interactive: skips the confirmation prompt
dot release              # interactive: confirms before mutating
```

## Workflow

The command performs every step itself; the agent checks the preconditions and reads the result.

1. **Preconditions**: clean working tree on `main`, `gh` authenticated, `git-cliff` and `mise` installed, Conventional Commits history.
1. **Fetch and prove**: the command fetches `origin` and requires `HEAD == origin/main` before any mutation.
1. **Compute and write**: the next semver from `git-cliff`, then `dot/version.go` and `CHANGELOG.md`.
1. **Gate**: the local gate (`format`, `check`, `test`) runs; a failure aborts and resets the staged release files before any commit or tag exists.
1. **Commit, tag, push**: the release commit, the annotated tag, the push of `main` and `refs/tags/v*` to `origin`, then the updated local `dot` binary is reapplied.
1. **Publish**: the tag push triggers `.github/workflows/cd.yml`, which creates the GitHub release; verify it with the release skill's `Verify` steps.

## Gotchas

- **No `mr`**: the fish abbreviation is interactive-only; agents call `mise run release` or `dot release`.
- **Diverged `main`**: the command refuses a dirty tree or a `HEAD` that differs from `origin/main`; commit or stash, then sync before retrying.

## Documentation

- [git-cliff](https://git-cliff.org) · `.github/workflows/cd.yml`
- Companion skills: [release](../../../skills/release/SKILL.md) (generic process), [conventional-commit](../../../skills/conventional-commit/SKILL.md) (commit grammar), [dot-cli](../../../skills/dot-cli/SKILL.md) (`dot release`).
