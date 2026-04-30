# Project Agent Rules

This is `fmind/dotfiles` — a chezmoi + mise dotfiles repo for Linux, macOS, and Cloud Shell.

**Gemini CLI is the priority AI coding system in this repo.** When agent surfaces need a home, target Gemini first.

## House rules

- **Catppuccin Mocha**: default theme everywhere it's supported.
- **Don't commit**: only when I explicitly ask. Run `mr n` first.
- **Lockfiles**: bump via `mr u`; never hand-edit `mise.lock`.
- **No-Icons**: ASCII only — these configs run over SSH and in Cloud Shell.
- **No-Sudo**: stay user-space; install via `mise`, `aqua`, or `pipx`. The `apt install` line in the README prereqs is the only allowed exception.
- **Verify upstream**: check the tool's current docs before adding flags or keys.
- **Vim mode**: enable in every TUI that supports it.

## Chezmoi conventions

- `dot_foo` → `~/.foo`. Never write a literal leading dot in source paths.
- `<name>.tmpl` → Go-template; branch on `.chezmoi.os` / `.chezmoi.arch`.
- `symlink_<name>.tmpl` → symlink target written verbatim into the link.
- `private_*` → mode 0600. `executable_*` → mode 0755. `*.age` → encrypted.
- `run_onchange_after_*.sh` → executed by `chezmoi apply` after files are written, only when the script's content changes (used here to bootstrap Gemini CLI extensions).
- `.chezmoiignore` blocks repo-only files (`README.md`, `mise.toml`, ...) from `apply`.

## Adding a new tool

1. Add the binary to `dot_config/mise/config.toml.tmpl` (alphabetical order).
2. Drop its config under `dot_config/<tool>/`, templated where needed.
3. If the tool exposes an agent surface, add it under `dot_gemini/` (skill, command, or subagent).
4. `mr t` to install, `mr a` to deploy, `mr l` to lock, `mr d` to verify.

## Agent Skills

Gemini CLI is the primary skill consumer; skills load from `~/.gemini/skills/`.

A skill is a directory whose `SKILL.md` has YAML frontmatter (`name` matching the dir, `description` for when to activate) — spec at <https://geminicli.com/docs/cli/skills/>.

Tooling: the `skills` CLI from [`vercel-labs/skills`](https://github.com/vercel-labs/skills) is installed via mise — call it directly (`skills add ...`, `skills find ...`).

Two install scopes — pick per skill, ask if unsure:

- **Project** → `.agents/skills/<slug>/`. Run `skills add <slug>` from the repo root. Commits with the codebase, pinned per-project.
- **Global** → `~/.gemini/skills/<slug>/`. Run `skills add --global <slug>`, then track it in this repo with `chezmoi add ~/.gemini/skills/<slug>` (imports into `dot_gemini/skills/<slug>/`).

For wrappers around official bundles, hand-author a `dot_gemini/skills/install-*-skills/SKILL.md` documenting the exact `skills add ...` line — see existing examples.

## Source layout

- `mise.toml` — repo-scoped tools and `mr <task>` workflows.
- `install.sh` — one-shot bootstrap (mise → chezmoi → apply).
- `dot_config/` — everything that lands in `~/.config/`.
- `dot_config/mise/config.toml.tmpl` — global toolchain (every CLI installed).
- `dot_gemini/` — Gemini CLI configs (primary agent surface; `GEMINI.md` is the persona). Subagent frontmatter must use `mcp_servers:` (snake_case) — Gemini CLI silently ignores the camelCase `mcpServers:` form.
- `dot_claude/` — Claude Code settings, plus symlinks `CLAUDE.md → ~/.gemini/GEMINI.md` and `skills → ~/.gemini/skills` so persona and skills are shared with Gemini.
- `dot_copilot/config.json` — GitHub Copilot CLI settings.
- `dot_local/bin/` — custom user-space executables (`deep-prompt`, `deep-research`, `dotfiles-verify`, `gcp-dotfiles-setup`).
- `dot_<file>` — top-level dotfiles (`~/.editrc`, `~/.gitconfig`, ...).
- `run_onchange_after_install-gemini-extensions.sh` — chezmoi auto-run script; installs/updates Gemini CLI extensions (`fgate`, `googleworkspace/cli`, `chrome-devtools-mcp`) on every `mr a` when its content changes.
- `.pre-commit-config.yaml` / `.markdownlint.json` / `.yamllint` — lint and secret-scan hygiene: gitleaks, markdownlint-cli2 (auto-fix), taplo (TOML format), yamllint, shellcheck, shfmt, and the standard `pre-commit-hooks` set. `.tmpl` files are excluded from style linters since chezmoi Go-template syntax breaks parsers.
- `.github/workflows/ci.yml` — CI runs `pre-commit` on push/PR to `main`; keep `mr n` clean locally so CI passes.
- `AGENTS.md` (this file) — repo rules.
