---
name: configure-mise-toml
description: Guide for authoring .mise.toml — pinning runtimes & tools, env vars, project tasks, hooks, and lockfile management.
---

# Configure mise (`.mise.toml`)

`.mise.toml` is mise's per-project config file — it pins runtime versions, tool installs (npm/pip/cargo/go/aqua/ubi backends), env vars, and lightweight tasks. One file replaces what used to take `.tool-versions`, `.envrc`, `Makefile`, and a hand-rolled `bin/` script.

## File Locations

| File | Scope |
|------|-------|
| `.mise.toml` | Project (commits with the repo) |
| `.mise.local.toml` | Project, **gitignored** — personal overrides |
| `~/.config/mise/config.toml` | User-global |
| `mise.lock` | Lockfile — commit it for reproducible installs |

Project files always win over user-global settings.

## Minimal `.mise.toml`

```toml
[tools]
node = "22"
python = "3.13"

[env]
NODE_ENV = "development"
```

After editing: `mise install` to materialize, `mise current` to verify.

## Tools (Backends)

```toml
[tools]
# First-class runtimes.
node    = "22"
python  = "3.13"
go      = "1.23"
rust    = "1.81"
java    = "openjdk-21"
ruby    = "3.3"
deno    = "latest"
bun     = "latest"
elixir  = "1.17"

# Backend-prefixed tools.
"npm:typescript"          = "5"
"npm:@google/jules"       = "latest"
"pipx:black"              = "latest"
"pipx:ruff"               = "latest"
"cargo:ripgrep"           = "14"
"cargo:fd-find"           = "10"
"go:github.com/golangci/golangci-lint/cmd/golangci-lint" = "latest"
"aqua:hashicorp/terraform" = "1.10"
"ubi:cli/cli"             = "latest"   # GitHub releases via ubi
```

Backends:

| Prefix | Source | Use when |
|--------|--------|----------|
| (none) | mise core registry | First-class runtimes |
| `npm:` | npm registry | JS/TS CLIs |
| `pipx:` | PyPI (isolated) | Python CLIs |
| `cargo:` | crates.io | Rust binaries |
| `go:` | Go modules | Go tools (`go install` style) |
| `aqua:` | aqua-registry | Pre-built binaries with sig verification |
| `ubi:` | GitHub releases | Anything else with `<owner>/<repo>` releases |
| `asdf:` | asdf plugin | Last resort — slower, build-from-source |

## Env Vars

```toml
[env]
PROJECT_ROOT = "{{config_root}}"
PATH_add = "{{config_root}}/bin"     # prepends to PATH
NODE_ENV = "development"

# Per-runtime extras.
[env._.python]
venv = { path = ".venv", create = true }   # auto-create + activate venv on cd

[env._.node]
package_manager = "pnpm"
```

Templates (`{{ ... }}`) use Tera syntax. Common variables: `{{config_root}}`, `{{env.HOME}}`, `{{cwd}}`.

## Tasks

```toml
[tasks.build]
description = "Build TypeScript and run unit tests"
run = """
npm run build
pytest -q
"""

[tasks.test]
depends = ["build"]
run = "vitest run"

[tasks.deploy]
description = "Deploy to staging"
run = "gcloud run deploy svc --source . --project=$PROJECT --region=$REGION"
env = { PROJECT = "my-project", REGION = "us-central1" }

[tasks.fmt]
run = ["ruff format .", "prettier -w ."]

[tasks.watch-test]
run = "vitest"
```

```bash
mise tasks                # list
mise run build            # invoke
mise run                  # interactive picker
mise watch test           # re-run on change
```

## Hooks (run on enter / leave / file change)

```toml
[hooks]
enter = "echo 'entering {{config_root}}'"
leave = "echo 'leaving'"

[[watch_files]]
patterns = ["package.json", "pnpm-lock.yaml"]
run = "pnpm install"
```

## Settings (mise behavior)

```toml
[settings]
experimental = true
trusted_config_paths = ["{{config_root}}"]
auto_install = true              # silently install missing tools
not_found_auto_install = true
status = { show_tools = true, show_env = true }
```

User-global settings go in `~/.config/mise/config.toml`; project-level overrides merge in.

## Lockfile (`mise.lock`)

```bash
mise lock                 # write lockfile reflecting installed versions
mise install              # honors lockfile when present
mise outdated             # show updates available within constraints
mise upgrade              # apply updates and re-lock
```

Commit `mise.lock` so teammates and CI converge to identical versions.

## Trust

Per-machine, mise will refuse to load `.mise.toml` it hasn't seen before:

```bash
mise trust                # accept the current project's .mise.toml
mise trust --untrust      # revoke
```

## Common Workflows

**Onboard a new repo.**

```bash
git clone <repo> && cd <repo>
mise trust && mise install
mise current
```

**Add a runtime + tool.**

```bash
mise use node@22                   # writes [tools] node = "22"
mise use "npm:typescript@5"        # adds an npm-backed binary
mise lock                          # capture exact resolved versions
```

**Run CI-equivalent locally.**

```bash
mise install --frozen              # error if mise.lock is out of date
mise run test
```

## Important Notes

1. **`.mise.toml` commits with the repo; `.mise.local.toml` is for personal overrides** (gitignore it).
2. **Lockfiles matter** — without `mise.lock`, two developers can end up on subtly different versions of the same `node@22.x.y`.
3. **`PATH_add` prepends; `PATH` overwrites** — almost always you want `PATH_add`.
4. **`auto_install = true`** silently installs missing tools on `cd` — convenient for local dev, undesirable in CI (use `--frozen`).
5. **First-class registry tools** are faster than `asdf:` plugins; check `mise registry` for what's first-class before reaching for `asdf:`.

## Documentation

- [mise home](https://mise.jdx.dev)
- [Configuration reference](https://mise.jdx.dev/configuration.html)
- [Backends](https://mise.jdx.dev/dev-tools/backends/)
- [Tasks](https://mise.jdx.dev/tasks/)
- [Templates](https://mise.jdx.dev/templates.html)
- [Settings](https://mise.jdx.dev/configuration/settings.html)
- [GitHub: jdx/mise](https://github.com/jdx/mise)
