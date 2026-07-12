---
name: upgrade-tools
description: Upgrade every pinned tool and dependency (mise, go.mod, pyproject.toml, GitHub Actions, dprint, ...) to latest stable, one ecosystem at a time with validation. Use when bumping a repo's toolchain or dependencies.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/upgrade-tools
  created: 2026-07-05
  updated: 2026-07-09
---

# Upgrade Tools & Dependencies

Bump every pinned tool and dependency in a repository to its **latest stable** version, one ecosystem at a time, validating after each so a bad bump is caught immediately. Covers the manifests this repo uses: [mise](../mise/SKILL.md) tool pins, Go modules, Python (`pyproject.toml`), OpenTofu/Terraform config, container images, [GitHub Actions](../github-actions/SKILL.md), and [dprint](../dprint/SKILL.md) plugins.

## Principles

- **Latest stable only**: no RCs/betas/pre-releases (except tools intentionally range-pinned pre-1.0, e.g. `ty>=0.0.51,<0.1`).
- **One ecosystem at a time**: upgrade → `mise run check` + `mise run test` → commit. Never bump everything then debug a wall of failures.
- **Lockfiles are the record**: commit `mise.lock`, `go.sum`, `uv.lock`, `.terraform.lock.hcl`. The manifest says "latest"; the lockfile says "which latest".
- **Validate, don't trust**: an upgrade isn't done until `mise run check` and `mise run test` pass. A green pre-existing baseline makes regressions obvious.
- **Respect semver majors**: `go get -u` / `uv lock --upgrade` stay within declared majors; a major bump is a deliberate, separately-reviewed change.

## Per-Manifest Playbook

### mise — tool versions (`mise.toml`, `mise.lock`)

Pins usually read `latest`; the lockfile pins the resolved version. Bump both:

```sh
mise upgrade --bump   # rewrites pinned versions to the newest resolved
mise lock             # refresh the lockfile to match installed versions
```

This repo orchestrates home + repo configs in one task — see the top-level `mise run upgrade` (bumps `~/.config/mise` and the repo, re-locks, applies, reinstalls). Commit the updated `dot_config/mise/mise.lock`.

### Go — modules & tools (`go.mod`, `go.sum`)

```sh
go get -u ./...                 # direct + indirect dependencies (stays within majors)
go get -u tool                  # bump every `tool` directive to latest (the `tool` meta-pattern; Go 1.24+)
go mod tidy                     # prune and reconcile go.sum
```

Bump the `go` directive when a newer stable toolchain ships (`go 1.NN.P`). Validate with `mise run check` (golangci-lint + govulncheck) and `mise run test`. See [go-stack](../go-stack/SKILL.md).

### Python — dependencies (`pyproject.toml`, `uv.lock`)

```sh
uv lock --upgrade               # bump every dependency in the lockfile
uv sync                         # install the upgraded set
```

Raise `requires-python` and dependency floors only when you rely on a newer feature; keep pre-1.0 tools range-pinned. To bump constraints inside `pyproject.toml`, run `uv add <package>@latest` or update the dependency array manually. Validate with `mise run check` + `mise run test`. See [python-stack](../python-stack/SKILL.md).

### OpenTofu / Terraform — providers & modules (`.terraform.lock.hcl`)

```sh
tofu init -upgrade              # bump provider and module versions within constraints
tofu providers lock -platform=linux_amd64 -platform=darwin_arm64  # refresh platform hashes for CI
```

Validate with `tofu validate` and `tflint`. Scan configuration with `trivy config` (see [security-scan](../security-scan/SKILL.md)).

### Container Images — base image pins (`Dockerfile`)

Locate the latest stable digest for your pinned base images (e.g., from [Chainguard Images](https://images.chainguard.dev) or Docker Hub) and update the tag/digest references:

```dockerfile
FROM python:3.14-slim
```

Validate by rebuilding the image (`mise run build:image`) and scanning (`mise run check:image` or `trivy image`). See [containerize](../containerize/SKILL.md).

### GitHub Actions — workflow pins (`.github/workflows/*.yml`)

Pin third-party actions by full commit SHA (supply-chain safety) and bump the SHA on release; keep first-party actions (`actions/checkout`, `jdx/mise-action`) at the latest major. Automate with `pinact run` or Dependabot (`.github/dependabot.yml`, `package-ecosystem: github-actions`). Validate with `actionlint`. See [github-actions](../github-actions/SKILL.md).

### dprint — formatter plugins (`dprint.json` / `dprint.jsonc`)

```sh
dprint config update            # rewrite plugin URLs to the latest wasm versions
```

Run for each config (root and nested `extends`). Validate with `dprint check`. See [dprint](../dprint/SKILL.md).

### Other ecosystems

Same shape — bump, then re-lock, then validate:

- **Node**: use `npx npm-check-updates -u` or `pnpm update --latest` to bump `package.json` constraints, then re-lock (`npm install` / `pnpm install` / `pnpm update`).
- **Rust**: `cargo update` → commit `Cargo.lock`. Use `cargo upgrade` (from `cargo-edit`) to bump constraints in `Cargo.toml`.
- **Agent skills**: `skills update -g -y` (this repo: `mise run skills`).

## Validate & Commit

1. After each ecosystem, run `mise run check` and `mise run test` (or the repo's equivalents); for config/markup, `dprint check`.
1. Run the full hook suite once at the end: `lefthook run pre-commit --all-files` then `lefthook run pre-push --all-files`.
1. Commit lockfiles alongside manifests, one Conventional Commit per ecosystem — `chore(deps): upgrade <ecosystem> to latest` (see [conventional-commit](../conventional-commit/SKILL.md)).
1. CI re-runs the same `mise run` tasks, so a green PR means the upgrade is reproducible.

## Documentation

- [mise: upgrade & lock](https://mise.jdx.dev/cli/upgrade.html)
- [Go: managing dependencies](https://go.dev/doc/modules/managing-dependencies) · [tool directives](https://go.dev/doc/modules/managing-dependencies#tools)
- [uv: locking & upgrading](https://docs.astral.sh/uv/concepts/projects/sync/#upgrading-locked-package-versions)
- [OpenTofu: Provider Dependency Lock File](https://opentofu.org/docs/cli/config/dependency-lock-files/)
- [Docker: Pinning Base Images](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#use-multi-stage-builds)
- [dprint: config update](https://dprint.dev/cli/#update)
- [GitHub Actions: security hardening (pin by SHA)](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions)
