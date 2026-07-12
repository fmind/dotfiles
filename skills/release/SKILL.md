---
name: release
description: Cut a versioned release — bump semver, generate the changelog with git-cliff, tag, and publish a GitHub release. Use when shipping a new tagged version.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/release
  created: 2026-07-04
  updated: 2026-07-09
---

# Release Process

Turn the Conventional Commits since the last tag into a versioned release: a bumped `CHANGELOG.md`, package manifest updates, an annotated git tag, and a GitHub release. Uses **git-cliff** (config: `dot_config/git-cliff/cliff.toml`, deployed to `~/.config/git-cliff/cliff.toml`).

## Preconditions

- Clean working tree on the default branch (`main`), synced with `origin`.
- History follows [conventional-commit](../conventional-commit/SKILL.md) — git-cliff groups commits by `type` and skips `chore(release)`/`chore(deps)`.

## Workflow

1. **Compute the next version** from the commit types since the last tag (`feat` → minor, `fix`/others → patch, `!`/`BREAKING CHANGE` → major):

   ```bash
   # Resolves the local/global cliff.toml configuration
   git-cliff --bumped-version

   # Or explicitly reference the global config if a local one is not present:
   git-cliff --config ~/.config/git-cliff/cliff.toml --bumped-version
   ```

1. **Update package manifests** (if the project doesn't use dynamic/VCS-based versioning) to match the computed version (e.g., `vX.Y.Z` or `X.Y.Z`):
   - **Python**: Bump `version` in `pyproject.toml` (unless using dynamic versioning via `hatch-vcs` or similar).
   - **Node.js**: Run `npm version --no-git-tag-version X.Y.Z` or update `package.json`.
   - **Go / OpenTofu**: Versioned via git tags (no file changes needed).
1. **Generate the changelog** for that version (pass `--config` to be explicit):
   ```bash
   git-cliff --config ~/.config/git-cliff/cliff.toml --bump -o CHANGELOG.md
   ```
1. **Commit** the changelog and manifest changes with a release commit (excluded from the changelog by design):
   ```bash
   git add CHANGELOG.md
   # plus the manifest you bumped in step 2, if any (Python: pyproject.toml · Node: package.json · Go/OpenTofu: none)
   git commit -m "chore(release): vX.Y.Z"
   ```
1. **Tag** annotated and push commit + tag:
   ```bash
   git tag -a vX.Y.Z -m "vX.Y.Z"
   git push --follow-tags
   ```
1. **Publish** the GitHub release using only the latest section as notes (write to a temp file to stay shell-agnostic):
   ```bash
   mkdir -p .agents/tmp
   git-cliff --config ~/.config/git-cliff/cliff.toml --latest --strip all > .agents/tmp/release-notes.md
   gh release create vX.Y.Z --title "vX.Y.Z" --notes-file .agents/tmp/release-notes.md
   ```
1. Print the release URL and the resolved version.

## Gotchas

- **Semver source of truth**: let `git-cliff --bumped-version` decide from commits; only override for a deliberate bump (e.g. first stable `v1.0.0`).
- **Tag prefix**: keep the `v` prefix consistent — git-cliff, `gh`, and Go module tags all expect `vX.Y.Z`.
- **Pre-1.0**: breaking changes bump the minor, not the major, until `v1.0.0`. Features/fixes bump the patch. This is controlled by `features_always_bump_minor = false` and `breaking_always_bump_major = false` in `cliff.toml`.
- **First Release (No Tags)**: if the repository has no tags, `git-cliff` defaults to the `initial_tag` (configured as `v0.1.0` in `cliff.toml` to match the `tag_pattern`).
- **Idempotency**: if the tag already exists, stop — never move a published tag.
- **Config Resolution**: if `git-cliff` is run without `--config`, it looks for a local `cliff.toml` or `git-cliff/cliff.toml` at the repository root. Fall back to `--config ~/.config/git-cliff/cliff.toml` when running in a repository without a custom setup.

## Documentation

- [git-cliff Documentation](https://git-cliff.org)
- Companion skills:
  - [conventional-commit](../conventional-commit/SKILL.md) — the commit grammar git-cliff parses.
  - [github-pull-request](../github-pull-request/SKILL.md) — merge work before releasing.
