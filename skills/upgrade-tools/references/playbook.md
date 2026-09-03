# Upgrade Playbook

Per-manifest commands for the [upgrade-tools](SKILL.md) workflow: bump, re-lock, validate with `mise run check` and `mise run test`, then move to the next ecosystem.

## mise (`mise.toml`, `mise.lock`)

Bump the pins and refresh the lockfile per the [mise skill](../mise/SKILL.md) (`mise upgrade --bump`, then `mise lock`). One bump covers every backend, so a `npm:` or `go:` tool pinned in `mise.toml` is a mise pin and needs no separate ecosystem step.

## Go (`go.mod`, `go.sum`)

```sh
go get -u ./...   # direct and indirect dependencies, within majors
go get -u tool    # every `tool` directive (Go 1.24+)
go mod tidy       # prune and reconcile go.sum
```

Bump the `go` directive when a newer stable toolchain ships. Validate with `mise run check` (golangci-lint, govulncheck) and `mise run test`. See [go-stack](../go-stack/SKILL.md).

## Python (`pyproject.toml`, `uv.lock`)

```sh
uv lock --upgrade                 # bump every locked dependency within its constraint
uv lock --upgrade-package <pkg>   # bump one package
uv sync                           # install the upgraded set
```

Raise `requires-python` and dependency floors in `pyproject.toml` by hand, only when a newer feature is needed; keep pre-1.0 tools range-pinned. See [python-stack](../python-stack/SKILL.md).

## Node (`package.json`, `pnpm-lock.yaml`)

```sh
pnpm update --latest   # bump constraints and the lockfile
```

See [typescript-stack](../typescript-stack/SKILL.md).

## Rust (`Cargo.toml`, `Cargo.lock`)

```sh
cargo update    # re-lock within constraints
cargo upgrade   # bump constraints (cargo-edit)
```

## Hugo (`go.mod`, `go.sum`)

Hugo Modules ride on Go modules, so a theme bump is a module bump; bump the `hugo-extended` pin in `mise.toml` in the same pass because a new theme often needs a newer Hugo:

```sh
hugo mod get -u ./...   # every imported module (theme, mounts)
hugo mod tidy           # prune modules the site no longer imports
```

Validate with `mise run build` and check the rendered links with `lychee`. See [hugo](../hugo/SKILL.md).

## OpenTofu (`.terraform.lock.hcl`)

```sh
tofu init -upgrade                                                 # providers and modules within constraints
tofu providers lock -platform=linux_amd64 -platform=darwin_arm64   # platform hashes for CI
```

Validate with `tofu validate`, `tflint`, and `trivy config`. See [terraform-stack](../terraform-stack/SKILL.md).

## Container images (`Dockerfile`)

Update the tag or digest of every `FROM` line to the latest stable from the image's registry (Chainguard, Docker Hub), rebuild with `mise run build`, and scan with `trivy image`. See [containerize](../containerize/SKILL.md).

## GitHub Actions (`.github/workflows/*.yml`)

Pin every action to a major tag (`owner/action@vN`) so patches flow within the major; let [dependabot](../dependabot/SKILL.md) open the major bumps and validate with `actionlint`. See [github-actions](../github-actions/SKILL.md).

## dprint (`dprint.json`)

```sh
dprint config update   # rewrite plugin URLs to the latest versions
```

Run it for each config (root and nested `extends`); validate with `dprint check`. See [dprint](../dprint/SKILL.md).

## Agent skills

`skills update -y` where external skills are installed; skip it in repositories that only author first-party skills.
