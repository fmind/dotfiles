---
name: upgrade-tools
description: Upgrade pinned tools and dependencies to latest stable one ecosystem at a time, validating mise, language, action, and formatter changes.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/upgrade-tools
  created: 2026-07-05
  updated: 2026-08-30
---

# Upgrade Tools & Dependencies

Bump every pinned tool and dependency in a repository to its **latest stable** version, one ecosystem at a time, validating after each so a bad bump is caught immediately. Covers the manifests this repo uses: [mise](../mise/SKILL.md) tool pins, Go modules, Python (`pyproject.toml`), OpenTofu/Terraform config, container images, [GitHub Actions](../github-actions/SKILL.md), and [dprint](../dprint/SKILL.md) plugins.

## Principles

- **Latest stable only**: no RCs/betas/pre-releases (except tools intentionally range-pinned pre-1.0, e.g. `ty>=0.0.51,<0.1`).
- **One ecosystem at a time**: upgrade → `mise run check` + `mise run test` → commit. Never bump everything then debug a wall of failures.
- **Lockfiles are the record**: commit `mise.lock`, `go.sum`, `uv.lock`, `.terraform.lock.hcl`. The manifest says "latest"; the lockfile says "which latest".
- **Validate, don't trust**: an upgrade isn't done until `mise run check` and `mise run test` pass. A green pre-existing baseline makes regressions obvious.
- **Respect semver majors**: `go get -u` / `uv lock --upgrade` stay within declared majors; a major bump is a deliberate, separately-reviewed change.

## Order of Operations

Ecosystems are not independent, so sequence matters. Run only the sections a repository actually has, in this order, validating between each:

1. **mise first.** It provisions the toolchain every later step runs — bumping `go`, `uv`, `tofu`, or `dprint` after using the old binaries means re-validating everything.
1. **Language dependencies next** (Go, Python, Rust, Node), because they determine whether the new toolchain compiles the project at all.
1. **Infrastructure and images after that** (OpenTofu, Dockerfile bases), which consume the language artifacts.
1. **CI and formatter config last** (GitHub Actions, dprint), the outermost layer and the least likely to cascade.

Stop at the first ecosystem that fails and fix it before continuing. Bumping the remaining ecosystems on top of a broken one is what turns a ten-minute upgrade into an afternoon of bisecting.

## Per-Manifest Playbook

### mise — tool versions (`mise.toml`, `mise.lock`)

Pins usually read `latest`; the lockfile pins the resolved version. Bump both:

```sh
mise upgrade --bump   # rewrites pinned versions to the newest resolved
mise lock             # refresh the lockfile to match installed versions
```

This repo orchestrates home + repo configs in one task — see the top-level `mise run upgrade` (bumps `~/.config/mise` and the repo, re-locks, applies, reinstalls). Commit the updated `dot_config/mise/mise.lock`.

`mise upgrade --bump` covers every backend in one pass — registry short names plus `aqua:`, `github:`, `go:`, `npm:`, and `pipx:` entries. Tools pinned there need no separate ecosystem step; a `npm:` or `go:` tool in `mise.toml` is a mise pin, not a project dependency. Re-pin anything deliberately held back (a comment above the pin should say why, as with a parser ABI or a broken upstream asset) instead of letting `--bump` carry it forward silently.

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

### Hugo — theme & site modules (`go.mod`, `go.sum`)

Hugo Modules ride on the Go module system, so a Hextra or theme bump is a Go module bump:

```sh
hugo mod get -u ./...           # bump every imported module (theme, mounts) to latest
hugo mod tidy                   # prune modules the site no longer imports
```

Bump the `hugo-extended` pin in `mise.toml` in the same pass — a new theme release often requires a newer Hugo. Validate with `hugo --gc --minify` (or `mise run build`) and check rendered links with `lychee`. See [hugo](../hugo/SKILL.md).

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

Validate by rebuilding the image (`mise run build:image`) and scanning (`mise run check:image` when the project defines it, else `trivy image`). See [containerize](../containerize/SKILL.md).

### GitHub Actions — workflow pins (`.github/workflows/*.yml`)

Pin every action to a major-version tag (`actions/checkout@v7`, `jdx/mise-action@v4`) and let the tag track security patches within the major. Do not pin SHAs — they turn every upstream patch into review noise. Automate the major bumps with Dependabot (`.github/dependabot.yml`, `package-ecosystem: github-actions`) — see [dependabot](../dependabot/SKILL.md). Validate with `actionlint`. See [github-actions](../github-actions/SKILL.md).

### dprint — formatter plugins (`dprint.json` / `dprint.jsonc`)

```sh
dprint config update            # rewrite plugin URLs to the latest wasm versions
```

Run for each config (root and nested `extends`). Validate with `dprint check`. See [dprint](../dprint/SKILL.md).

### Other ecosystems

Same shape — bump, then re-lock, then validate:

- **Node**: use `npx npm-check-updates -u` or `pnpm update --latest` to bump `package.json` constraints, then re-lock (`npm install` / `pnpm install` / `pnpm update`).
- **Rust**: `cargo update` → commit `Cargo.lock`. Use `cargo upgrade` (from `cargo-edit`) to bump constraints in `Cargo.toml`.
- **Agent skills**: `skills update -y` where external skills are installed; skip it in repositories that only author first-party skills.

## Validate & Commit

1. After each ecosystem, run `mise run check` and `mise run test` (or the repo's equivalents); for config/markup, `dprint check`.
1. Run the full hook suite once at the end: `lefthook run pre-commit --all-files` then `lefthook run pre-push --all-files`.
1. Commit lockfiles alongside manifests, one Conventional Commit per ecosystem — `chore(deps): upgrade <ecosystem> to latest` (see [conventional-commit](../conventional-commit/SKILL.md)).
1. CI re-runs the same `mise run` tasks, so a green PR means the upgrade is reproducible.

## Documentation

- [mise: upgrade & lock](https://mise.jdx.dev/cli/upgrade.html)
- [Go: managing dependencies](https://go.dev/doc/modules/managing-dependencies) · [tool directives](https://go.dev/doc/modules/managing-dependencies#tools)
- [uv: locking & upgrading](https://docs.astral.sh/uv/concepts/projects/sync/#upgrading-locked-package-versions)
- [OpenTofu: Dependency Lock File](https://opentofu.org/docs/language/files/dependency-lock/)
- [Docker: Pinning Base Images](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#use-multi-stage-builds)
- [dprint: config update](https://dprint.dev/cli/#update)
- [Hugo Modules](https://gohugo.io/hugo-modules/)
- [GitHub Actions: using third-party actions](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions)
