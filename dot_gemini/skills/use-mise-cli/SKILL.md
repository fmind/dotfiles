---
name: use-mise-cli
description: Guide for using mise (the polyglot tool/runtime/version manager) — installing tools, project-pinning versions, env vars, tasks, and run-on-shell-init triggers.
---

# Use mise CLI

[mise](https://mise.jdx.dev) is a polyglot version & tool manager (asdf successor) plus an `.envrc`-style env loader and a task runner. One file (`.mise.toml`) pins runtimes, npm/pip/cargo packages, env vars, and project tasks.

## One-time Setup

```bash
# Install (Linux/macOS).
curl https://mise.run | sh

# Activate in your shell (chezmoi templates already do this).
echo 'eval "$(mise activate bash)"' >> ~/.bashrc
echo 'eval "$(mise activate zsh)"'  >> ~/.zshrc

# Sanity check.
mise --version
mise doctor          # detects setup issues
```

## Tools (Runtimes & Binaries)

```bash
# Install a tool globally (writes to ~/.config/mise/config.toml).
mise use -g node@22
mise use -g python@3.13 uv@latest

# Install for the current project (writes to ./.mise.toml).
mise use node@22 python@3.13

# Pin a tool you already have installed.
mise use --pin go@1.23

# Install everything declared in .mise.toml (idempotent — does nothing if up to date).
mise install

# What's active here?
mise current
mise ls
```

## `.mise.toml` (project file)

```toml
[tools]
node = "22"
python = "3.13"
"npm:@google/jules" = "latest"
"pipx:black" = "latest"
"cargo:ripgrep" = "14"

[env]
NODE_ENV = "development"
DATABASE_URL = "postgres://localhost/dev"

[env._.python.venv]
path = ".venv"
create = true       # auto-create + activate a venv when entering the dir

[tasks.build]
run = "npm run build && python build.py"

[tasks.test]
depends = ["build"]
run = "pytest"
```

Discover backends (`npm:`, `pipx:`, `cargo:`, `go:`, `aqua:`, `ubi:`, `asdf:`) with `mise registry` or `mise plugins ls-remote`.

## Tasks (lightweight task runner)

```bash
mise tasks                      # list all defined tasks
mise run build                  # run a task
mise run                        # interactive picker
mise watch test                 # re-run on file change

# Define a one-off shell task inline.
mise run -- 'echo "hello $USER"'
```

## Env Vars (`.env`-replacement)

```bash
# .mise.toml [env] is loaded automatically when you `cd` in.
# Variables can reference other variables and shell.
[env]
PROJECT_ROOT = "{{config_root}}"
PATH_add = "{{config_root}}/bin"

mise env                        # show what mise would load
mise env --shell zsh            # eval-able output
```

## Tracking Tools Across Machines

```bash
# After editing .mise.toml, sync the local install.
mise install

# Lock a precise version (creates mise.lock).
mise lock

# Update to latest (within constraints).
mise outdated
mise upgrade
```

## Common Workflows

**Onboard a new repo.**

```bash
git clone <repo> && cd <repo>
mise trust              # accept .mise.toml as trusted
mise install            # install all declared tools
mise current            # confirm
```

**Add a new tool to a project.**

```bash
mise use python@3.13            # writes to .mise.toml
mise use "npm:typescript@5"     # adds an npm-backed binary
mise lock                       # capture exact versions
```

**Switch shells / debug PATH.**

```bash
mise activate bash --shims      # if `eval $(mise activate)` isn't possible
mise which node                 # locate active binary
mise reshim                     # rebuild shim links
```

## Companion: chezmoi

This dotfile setup uses chezmoi to template `~/.config/mise/config.toml`. After `chezmoi apply`, run `mise install` to materialize tool versions on the new machine.

## Important Notes

1. **`mise trust`** is required once per `.mise.toml` per machine — it prevents arbitrary code execution from cloned repos.
2. **`asdf` plugins work via the `asdf:` backend**, but **prefer first-class backends (`npm:`, `pipx:`, `cargo:`, `aqua:`, `ubi:`)** — they're faster and don't need build tooling.
3. **`mise.lock`** is the lockfile — commit it so teammates and CI install identical versions.
4. **Env-var ordering matters** — variables loaded from `.mise.toml` override the shell unless prefixed with a default like `_.path` / `PATH_add`.
5. **Do not nest mise inside another tool manager** (volta, nvm, pyenv, asdf) — pick one.

## Documentation

- [mise home](https://mise.jdx.dev)
- [Configuration reference (`.mise.toml`)](https://mise.jdx.dev/configuration.html)
- [Tasks](https://mise.jdx.dev/tasks/)
- [Tool backends](https://mise.jdx.dev/dev-tools/backends/)
- [Env vars](https://mise.jdx.dev/configuration/settings.html#env)
- [GitHub: jdx/mise](https://github.com/jdx/mise)
