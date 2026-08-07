---
name: github-actions
description: Canonical GitHub Actions CI/CD that runs the same `mise run` tasks (format, check, test) as the local git hooks, plus CD deploy templates. Use when setting up or editing repository workflows.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/github-actions
  created: 2026-07-04
  updated: 2026-08-07
---

# GitHub Actions CI/CD Standard

Canonical CI/CD workflows for GitHub repositories. The CI workflow runs the canonical [mise](../mise/SKILL.md) tasks (`format`, `check`, `test`) directly — the same tasks the local [lefthook](../lefthook/SKILL.md) hooks delegate to — so "passes locally" guarantees "passes in CI" without duplicating linter and check lists. The CD workflow provides templates for building and deploying applications based on the project's language stack.

## Principles

- **Single task vocabulary**: CI runs the same `mise run format/check/test` tasks that the local pre-commit/pre-push hooks delegate to. Driving both from one mise task set eliminates drift between local checks and CI; workflow syntax and first-party skill contracts belong inside `check`, not in parallel CI-only steps.
- **Workflows are linted too**: a `check:actions` task runs `actionlint` (correctness: workflow schema, expression types, `shellcheck` on `run:` scripts) and `zizmor` (security: template injection, credential persistence, cache poisoning, unpinned actions), so workflow regressions fail the same gate as code. zizmor runs offline by default, so the task needs no GitHub token; [zizmor.yml](references/zizmor.yml) relaxes its hash-pin default to the tag-pinning policy from [upgrade-tools](../upgrade-tools/SKILL.md).
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
1. Copy [security.yml](references/security.yml) to `.github/workflows/security.yml` for the scheduled full-history scan.
1. Copy [zizmor.yml](references/zizmor.yml) to `.github/zizmor.yml`, pin `actionlint` and `zizmor` in `mise.toml` `[tools]`, and expose a `check:actions` task running `actionlint` then `zizmor .github/workflows/`.

## Templates

- **CI**: See [ci.yml](references/ci.yml) which runs `mise run format`, `mise run check` (static checks incl. `check:leaks`), and `mise run test` across the whole tree, then asserts an empty porcelain status so formatting or generation drift fails the build. CI stays minimal; the `check:leaks` task covers commit-scope secret scanning.
- **Security**: See [security.yml](references/security.yml), a scheduled/manual companion that rescans the full history and checkout with the same pinned scanners at `fetch-depth: 0`. It reports nothing: a finding or scanner error simply fails the job, which is what the notification is for.
- **CD**: See [cd.yml](references/cd.yml) which provides commented templates for Go containers (using `ko`), Python packages (using `uv`), and general Docker builds.

## Gotchas

- **Separate security cadence**: Keep full-history scanning out of push CI — it needs `fetch-depth: 0` and minutes of runtime, which push CI should not pay on every commit. Use a scheduled/manual workflow with a job timeout instead, and let a non-zero scanner exit fail it.
- **Stable caches**: `jdx/mise-action` caches using `mise.toml`/`mise.lock` — commit `mise.lock` for reproducible caching.
- **Runtime warning mitigation**: Use current major versions of actions (e.g., `actions/checkout@v7` and `jdx/mise-action@v4`) to stay compliant with GitHub's latest runner runtime deprecations (Node 20+).
- **Templates expand everywhere in `run:`**: GitHub substitutes `${{ ... }}` before the shell sees the script — including inside shell comments — so an inline `github.ref_name` expansion is code injection via a crafted tag name. Read the runner's default environment variables (`$GITHUB_REF_NAME`) or pass the expression through the step's `env:` block instead.
- **Cache CI, not CD**: `mise-action` caching is safe on CI, but set `cache: false` in deploy jobs — a poisoned shared cache must not be able to reach signed release artifacts.

## Documentation

- [GitHub Actions Documentation](https://docs.github.com/actions)
- [jdx/mise-action](https://github.com/jdx/mise-action)
- Companion skills:
  - [mise](../mise/SKILL.md) / [lefthook](../lefthook/SKILL.md) — the execution and hook runner.
  - [containerize](../containerize/SKILL.md) — packaging standards referenced in CD templates.
  - [security-scan](../security-scan/SKILL.md) — security scans run by hooks and CD steps.
