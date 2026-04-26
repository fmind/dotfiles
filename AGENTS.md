# Project Agent Rules

This is `fmind/dotfiles` — a chezmoi + mise dotfiles repo for Linux, macOS, and Cloud Shell.

See `README.md` for what it ships; this file is for agents editing the repo.

## Source layout

- `mise.toml` — repo-scoped tools and `mr <task>` workflows.
- `install.sh` — one-shot bootstrap (mise → chezmoi → apply).
- `dot_AGENTS.md` — global agent rules (deployed as `~/.AGENTS.md`).
- `dot_config/` — everything that lands in `~/.config/`.
- `dot_config/mise/config.toml.tmpl` — global toolchain (every CLI installed).
- `dot_claude/`, `dot_gemini/`, `dot_copilot/` — AI agent configs.
- `dot_<file>` — top-level dotfiles (`~/.editrc`, `~/.gitconfig`, ...).
- `AGENTS.md` (this file) — repo rules.

## Chezmoi conventions

- `dot_foo` → `~/.foo`. Never write a literal leading dot in source paths.
- `<name>.tmpl` → Go-template; branch on `.chezmoi.os` / `.chezmoi.arch`.
- `symlink_<name>.tmpl` → symlink; e.g. `dot_claude/symlink_CLAUDE.md.tmpl` and `dot_gemini/symlink_GEMINI.md.tmpl` both point at `~/.AGENTS.md`.
- `private_*` → mode 0600. `executable_*` → mode 0755. `*.age` → encrypted.
- `.chezmoiignore` blocks repo-only files (`README.md`, `mise.toml`, ...) from `apply`.

## Editing workflow

1. Edit the source file under `dot_*` (NOT the deployed copy in `~`).
2. `mr c` to preview the diff (`chezmoi diff` works too).
3. `mr a` to apply. For tool changes, follow with `mr t` then `mr l`.
4. `mr n` runs pre-commit on all files; fix issues before suggesting a commit.
5. `mr s` runs gitleaks; never commit secrets, even as redacted examples.

## House rules

- **No-Sudo**: stay user-space; install via `mise`, `aqua`, or `pipx`. The `apt install` line in the README prereqs is the only allowed exception.
- **No-Icons**: ASCII only — these configs run over SSH and in Cloud Shell.
- **Catppuccin Mocha**: default theme everywhere it's supported.
- **Vim mode**: enable in every TUI that supports it.
- **Lockfiles**: bump via `mr u`; never hand-edit `mise.lock`.
- **Verify upstream**: check the tool's current docs before adding flags or keys.
- **Don't commit**: only when I explicitly ask. Run `mr n && mr s` first.

## Adding a new tool

1. Add the binary to `dot_config/mise/config.toml.tmpl` (alphabetical order).
2. Drop its config under `dot_config/<tool>/`, templated where needed.
3. If the tool exposes an agent surface, mirror it under `dot_claude/` and `dot_gemini/` (skill, command, or subagent) to keep both ecosystems in parity.
4. `mr t` to install, `mr a` to deploy, `mr l` to lock, `mr d` to verify.
