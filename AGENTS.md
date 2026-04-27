# Project Agent Rules

This is `fmind/dotfiles` — a chezmoi + mise dotfiles repo for Linux, macOS, and Cloud Shell.

**Gemini CLI is the priority AI coding system in this repo.** When agent surfaces need a home, target Gemini first.

## House rules

- **No-Sudo**: stay user-space; install via `mise`, `aqua`, or `pipx`. The `apt install` line in the README prereqs is the only allowed exception.
- **No-Icons**: ASCII only — these configs run over SSH and in Cloud Shell.
- **Catppuccin Mocha**: default theme everywhere it's supported.
- **Vim mode**: enable in every TUI that supports it.
- **Lockfiles**: bump via `mr u`; never hand-edit `mise.lock`.
- **Verify upstream**: check the tool's current docs before adding flags or keys.
- **Don't commit**: only when I explicitly ask. Run `mr n` first.

## Source layout

- `mise.toml` — repo-scoped tools and `mr <task>` workflows.
- `install.sh` — one-shot bootstrap (mise → chezmoi → apply).
- `dot_config/` — everything that lands in `~/.config/`.
- `dot_config/mise/config.toml.tmpl` — global toolchain (every CLI installed).
- `dot_gemini/` — Gemini CLI configs (primary agent surface; `GEMINI.md` is the persona). Subagent frontmatter must use `mcp_servers:` (snake_case) — Gemini CLI silently ignores the camelCase `mcpServers:` form.
- `dot_claude/settings.json` — Claude Code settings (secondary tool, no shared persona).
- `dot_copilot/config.json` — GitHub Copilot CLI settings.
- `dot_<file>` — top-level dotfiles (`~/.editrc`, `~/.gitconfig`, ...).
- `AGENTS.md` (this file) — repo rules.

## Chezmoi conventions

- `dot_foo` → `~/.foo`. Never write a literal leading dot in source paths.
- `<name>.tmpl` → Go-template; branch on `.chezmoi.os` / `.chezmoi.arch`.
- `symlink_<name>.tmpl` → symlink target written verbatim into the link.
- `private_*` → mode 0600. `executable_*` → mode 0755. `*.age` → encrypted.
- `.chezmoiignore` blocks repo-only files (`README.md`, `mise.toml`, ...) from `apply`.

## Editing workflow

1. Edit the source file under `dot_*` (NOT the deployed copy in `~`).
2. `mr c` to preview the diff (`chezmoi diff` works too).
3. `mr a` to apply. For tool changes, follow with `mr t` then `mr l`.
4. `mr n` runs pre-commit on all files (gitleaks included); fix issues before suggesting a commit.

## Adding a new tool

1. Add the binary to `dot_config/mise/config.toml.tmpl` (alphabetical order).
2. Drop its config under `dot_config/<tool>/`, templated where needed.
3. If the tool exposes an agent surface, add it under `dot_gemini/` (skill, command, or subagent).
4. `mr t` to install, `mr a` to deploy, `mr l` to lock, `mr d` to verify.

## Agent Skills

Gemini CLI is the primary skill consumer; skills load from `~/.gemini/skills/`.

A skill is a directory whose `SKILL.md` has YAML frontmatter (`name` matching the dir, `description` for when to activate) — spec at <https://geminicli.com/docs/cli/skills/>.

Tooling: the `skills` CLI from [`vercel-labs/skills`](https://github.com/vercel-labs/skills) is installed via mise — call it directly (`skills add ...`, `skills find ...`); fall back to `npx skills` only on a fresh checkout.

Two install scopes — pick per skill, ask if unsure:

- **Project** → `.agents/skills/<slug>/`. Run `skills add <slug>` from the repo root. Commits with the codebase, pinned per-project.
- **Global** → `~/.gemini/skills/<slug>/`. Run `skills add --global <slug>`, then track it in this repo with `chezmoi add ~/.gemini/skills/<slug>` (imports into `dot_gemini/skills/<slug>/`).

For wrappers around official bundles, hand-author a `dot_gemini/skills/install-*-skills/SKILL.md` documenting the exact `skills add ...` line — see existing examples.
