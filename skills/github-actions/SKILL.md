---
name: github-actions
description: Canonical GitHub Actions CI/CD that runs the same `mise run` tasks (format, check, test) as the local git hooks, plus CD deploy templates. Use when setting up or editing repository workflows.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/github-actions
  created: 2026-07-04
  updated: 2026-07-09
---

# GitHub Actions CI/CD Standard

Canonical CI/CD workflows for GitHub repositories. The CI workflow runs the canonical [mise](../mise/SKILL.md) tasks (`format`, `check`, `test`) directly — the same tasks the local [lefthook](../lefthook/SKILL.md) hooks delegate to — so "passes locally" guarantees "passes in CI" without duplicating linter and check lists. The CD workflow provides templates for building and deploying applications based on the project's language stack.

## Principles

- **Single task vocabulary**: CI runs the same `mise run format/check/test` tasks that the local pre-commit/pre-push hooks delegate to. Driving both from one mise task set eliminates drift between local checks and CI.
- **Tools from `mise.toml`**: `jdx/mise-action` installs and caches the project toolchain, ensuring that the CI runner runs the identical tool versions pinned locally.
- **Least privilege**: Default to `permissions: contents: read`; widen permissions (like `packages: write` or `id-token: write`) only where needed in deployment jobs.
- **OIDC & Trusted Publishing**: Prefer OpenID Connect (OIDC) for keyless container signing (via `cosign`) and package publishing (via PyPI Trusted Publishing), eliminating long-lived credentials.
- **Downcased registry paths**: Dynamically sanitize and downcase repository references to prevent push failures to case-sensitive container registries.
- **Fail fast, cancel stale**: Concurrency settings cancel superseded runs on pull requests and feature branches automatically while preserving runs on the main branch.
- **Clean-state verification**: CI asserts an empty porcelain status (`test -z "$(git status --porcelain)"`) after formatting and generation so both tracked modifications and untracked artifacts fail the build.
- **Latest Actions**: Keep GitHub Actions dependencies up-to-date (e.g., `actions/checkout@v7`, `jdx/mise-action@v4`).

## Setup

1. Copy [ci.yml](references/ci.yml) to `.github/workflows/ci.yml`.
1. Copy [cd.yml](references/cd.yml) to `.github/workflows/cd.yml` and enable/customize the template corresponding to your project's language and deployment target.

## Templates

- **CI**: See [ci.yml](references/ci.yml) which runs `mise run format`, `mise run check` (static checks incl. `check:leaks`), and `mise run test` across the whole tree, then `git diff --exit-code` to fail if formatting or generation left changes. CI stays minimal; the `check:leaks` task covers commit-scope secret scanning.
- **CD**: See [cd.yml](references/cd.yml) which provides commented templates for Go containers (using `ko`), Python packages (using `uv`), and general Docker builds.

## Gotchas

- **Optional security job**: CI defaults to the `format`/`check`/`test` tasks only. For full-history secret and dependency scanning, add a separate `security` job with `fetch-depth: 0` running `gitleaks git` + `trivy fs` — see the [security-scan](../security-scan/SKILL.md) skill (or run those scans on demand).
- **Stable caches**: `jdx/mise-action` caches using `mise.toml`/`mise.lock` — commit `mise.lock` for reproducible caching.
- **Runtime warning mitigation**: Use current major versions of actions (e.g., `actions/checkout@v7` and `jdx/mise-action@v4`) to stay compliant with GitHub's latest runner runtime deprecations (Node 20+).

## Documentation

- [GitHub Actions Documentation](https://docs.github.com/actions)
- [jdx/mise-action](https://github.com/jdx/mise-action)
- Companion skills:
  - [mise](../mise/SKILL.md) / [lefthook](../lefthook/SKILL.md) — the execution and hook runner.
  - [containerize](../containerize/SKILL.md) — packaging standards referenced in CD templates.
  - [security-scan](../security-scan/SKILL.md) — security scans run by hooks and CD steps.
