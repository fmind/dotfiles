---
name: github-actions
description: Configure GitHub Actions CI/CD so workflows run the same mise format, check, test, build gate as local hooks, lint with actionlint and zizmor, and deploy keyless. Use when adding or fixing workflows.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/github-actions
  created: "2026-07-04"
  updated: "2026-09-03"
---

# GitHub Actions

CI runs the canonical [mise](../mise/SKILL.md) `all` task so it never drifts from the local [lefthook](../lefthook/SKILL.md) hooks; CD builds, signs, and publishes from a version tag with OIDC instead of long-lived credentials.

## Workflow

1. **CI**: copy [ci.yml](references/ci.yml) to `.github/workflows/ci.yml`; it runs `mise run all`, asserts an empty porcelain status so drift fails the build, and fetches 100 commits to match the `check:leaks` bound.
1. **Security**: copy [security.yml](references/security.yml) to `.github/workflows/security.yml`: a scheduled full-history [gitleaks](../gitleaks/SKILL.md) and [trivy](../trivy/SKILL.md) rescan where any finding fails the job.
1. **CD**: copy [cd.yml](references/cd.yml) to `.github/workflows/cd.yml` and enable one job: a `ko` container signed keyless with [cosign](../cosign/SKILL.md), a PyPI package via Trusted Publishing, or a Dockerfile image.
1. **Lint the workflows**: copy [zizmor.yml](references/zizmor.yml) to `.github/zizmor.yml`, pin `actionlint`, `shellcheck`, and `zizmor` in `mise.toml` `[tools]`, and expose `check:actions`:

   ```toml
   [tasks."check:actions"]
   description = "Lint and audit GitHub Actions workflows (actionlint + zizmor)"
   run = ["actionlint", "zizmor --offline .github/workflows/"]
   ```

1. **Verify locally**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).

## Principles

- **One gate**: CI runs `mise run all`, the same tasks the hooks call plus the production build; no CI-only steps.
- **Tools from `mise.toml`**: `jdx/mise-action` installs and caches the pinned toolchain so CI runs the versions pinned locally; commit `mise.lock` for stable caches.
- **Least privilege**: top-level `permissions: contents: read`; widen (`packages: write`, `id-token: write`) only in the job that needs it.
- **Keyless**: OIDC for cosign signing and PyPI Trusted Publishing; no long-lived registry or package credentials.
- **Fail fast, cancel stale**: `concurrency` cancels superseded runs on pull requests and branches while `main` keeps every run.
- **Current majors**: keep actions on the current major tag (`actions/checkout@v7`, `jdx/mise-action@v4`) so runner runtime deprecations never bite.

## Gotchas

- **Injection and cache poisoning**: `${{ ... }}` expands in `run:` before the shell runs, even inside comments, and a shared cache can reach signed artifacts; the fixes are the `template-injection` and `cache-poisoning` rows of [zizmor](../zizmor/SKILL.md).
- **Separate security cadence**: full-history scans need `fetch-depth: 0` and minutes of runtime; keep them in the scheduled workflow with a job timeout, not in push CI.
- **Downcase registry paths**: GHCR rejects uppercase owners; derive `IMAGE_REPOSITORY` with `tr '[:upper:]' '[:lower:]'` as the templates do.
- **Tag pins by policy**: [zizmor.yml](references/zizmor.yml) relaxes `unpinned-uses` to the major-tag policy of [upgrade-tools](../upgrade-tools/SKILL.md); hash pins are the Scorecard default, not ours.

## Documentation

- [GitHub Actions](https://docs.github.com/actions) · [jdx/mise-action](https://github.com/jdx/mise-action)
- Companion skills: [zizmor](../zizmor/SKILL.md) (findings and fixes), [trivy](../trivy/SKILL.md) (`security.yml`), [containerize](../containerize/SKILL.md) (`build:image`), [cosign](../cosign/SKILL.md) (signing), [secure](../secure/SKILL.md) (checklist).
