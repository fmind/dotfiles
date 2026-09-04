# AGENTS.md (Project)

This is `fmind/dot` — a chezmoi + mise dotfiles repo for **AI-CLI-first** development on Linux and Mac OS with Go and Python as the core programming languages.

User-facing install and usage docs live in `README.md`; this file is for agents working inside the repo.

## House rules

- **Chezmoi**: Edit the source tree (this repository), never deployed copies under `$HOME`; automation always runs `chezmoi apply --force` so a prompt can never block. Naming, templates, and secrets: [chezmoi skill](.agents/skills/chezmoi/SKILL.md).
- **GitHub Access**: Use `gh` CLI for repository, issue, and PR operations.
- **Git Push to Main**: Direct commit/push to `main` branch is permitted (no feature branch required).
- **Lint-before-done**: `mise run all` (format + check + test + build, the same gate CI runs) must pass before reporting a task complete.
- **Markdown Lists**: Only use `1.` for all numbered list items in markdown files to ensure correct dynamic rendering.
- **No-Hard-Wrap**: Every `*.md` keeps each paragraph on a single line.
- **No-Sudo**: Stay user-space; install via `mise`.
- **README Scope**: Keep setup/auth instructions in `README.md`; exclude repository tasks, aliases, and workflows.
- **Secrets**: `*.age` files are encrypted; never modify or commit decrypted versions.
- **Theme**: **Tokyo Night (Moon)** is default across every tool that supports theming.
- **Vim mode**: Enable in every TUI that supports it.

## Workflows

Tasks run via `mise run <task>`. **Do not use `mr`** — it is an interactive-only fish abbreviation.

Aliases split into two namespaces so a mistyped letter can never fire the wrong kind of task:

- **Common tasks** take the canonical one-letter alias from the mise skill: `a` all, `b` build, `c` check, `f` format, `i` install, `t` test, `w` watch (plus `c*`/`f*` for subtasks, e.g. `cg` check:go, `fd` format:dprint).
- **Project management** tasks take an `m`-prefixed alias: `ma` apply, `md` diff, `me` deploy, `mf` full, `mg` completions, `mh` hooks, `mk` lock, `mo` doctor, `mp` prune, `mr` release, `mt` tools, `mtr` trust, `mu` upgrade, `mv` vim, `mx` verify.

Key routines:

- **First-time setup**: `mise run install` (trust → tools → hooks → vim).
- **Routine update**: `mise run full` (synchronize environment).
- **Iterate**: Edit source → `mise run apply` (`mise run diff` to preview) → `mise run check` (or `mise run all`) → `mise run verify`.
- **Add tool**: Append to `dot_config/mise/config.toml.tmpl` → `mise run tools` → `mise run lock`.
- **Upgrade tools**: `mise run upgrade` (upgrades tool pins and lockfiles).
- **Reclaim disk**: `mise run prune` (reclaim development caches and agent transcripts).
- **Release**: `mise run release` (runs validation, tags, pushes `main` and tag).
- **Manage skills**: Author global skills directly under `skills/` (repo-scoped ones under `.agents/skills/`) with the `skillify` skill and the limits in `dot_agents/AGENTS.md`; register each in `skills/contracts.json`, add a routing probe in `dot/testdata/skills/routing-boundaries.json`, and validate with `mise run check:skills` plus `mise run test`.
- **Create visuals**: Use `fmind-visuals` skill (Slidev for decks, Mermaid for diagrams).

> Note: If `mise` fails with `command not found` in an agent shell, call `~/.local/bin/mise` directly.

The unified `dot` CLI (source in `dot/`) is compiled to `~/.local/bin/dot`; every command and alias is documented once in [`skills/dot-cli/SKILL.md`](skills/dot-cli/SKILL.md), with the `dot prune` flag matrix in [`references/prune-flags.md`](skills/dot-cli/references/prune-flags.md); `dot <command> --help` remains authoritative for the complete flag list.

## Agents

Two assets are authored once and consumed by all agent CLIs:

- **Persona** — `dot_agents/AGENTS.md` deploys to `~/.agents/AGENTS.md`, symlinked in by Antigravity, Claude, Codex, Copilot, and Grok; OpenCode alone reads it through the `instructions` array in `opencode.json.tmpl` because it has no per-host instruction filename.
- **Skills** — `skills/` is symlinked to `~/.agents/skills/` (consumed by all agent CLIs).

**Rule: every global skill lives in `skills/`.**

## Layout

- `.agents/` — Workspace-scoped state, session records, project skills, and scratch scripts for AI agents.
- `.antigravitycli/` — Workspace-scoped session records and state for Antigravity CLI.
- `.chezmoi.toml.tmpl` — Host-specific chezmoi configuration template.
- `.chezmoiignore` — Chezmoi exclude rules for non-deployed repository files.
- `.chezmoitemplates/` — Shared template partials pulled in by `modify_` scripts via `includeTemplate`.
- `.claude/` — Claude Code workspace state (ignored) plus the tracked `skills` link to `.agents/skills/` so Claude discovers the repo-local skills.
- `.gemini/` — Workspace configurations and metadata for Antigravity CLI.
- `.github/` — GitHub Actions workflows (`ci.yml`, `cd.yml`, `security.yml`), the `zizmor.yml` audit policy, and Dependabot config.
- `.gitignore` — Git exclusion rules.
- `.gitleaks.toml` — Security configuration and secrets scanner allowlist for GitLeaks.
- `.stylua.toml` — StyleLua formatting policy for managed Neovim Lua sources.
- `AGENTS.md` (this file) — Repository guide, conventions, and layout for AI agents.
- `CHANGELOG.md` — Versioned release history generated from Conventional Commits.
- `CLAUDE.md` — Symlink to `AGENTS.md` so Claude Code loads the same project instructions.
- `dot/` — Go CLI source package containing the unified `dot` command-line utility.
- `dot_agents/` — Source folder containing unified global instructions (`AGENTS.md`) and skills symlink template.
- `dot_claude/` — Claude Code CLI settings template and persona/skills symlinks.
- `dot_codex/` — OpenAI Codex CLI partial configuration modifier and persona symlink.
- `dot_config/` — Custom configuration templates deployed to `~/.config/`.
- `dot_copilot/` — GitHub Copilot CLI integration configurations and symlink templates.
- `dot_duckdbrc` — DuckDB CLI settings deployed to `~/.duckdbrc`.
- `dot_gemini/` — Antigravity CLI config settings and symlinks deployed to `~/.gemini/`.
- `dot_gitconfig.tmpl` — Global Git configuration template deployed to `~/.gitconfig`.
- `dot_grok/` — Grok Build CLI partial configuration modifier, LSP servers, hooks, and persona/skills symlinks.
- `dot_inputrc` — Readline prompt styling configurations deployed to `~/.inputrc`.
- `dot_kube/` — kubectl settings deployed to `~/.kube/`.
- `dot_local/` — Executables and application configurations deployed to `~/.local/`.
- `dot_npmrc` — npm configuration deployed to `~/.npmrc`.
- `dot_skaffold/` — Skaffold settings deployed to `~/.skaffold/`.
- `dot_sqliterc` — SQLite interactive shell settings deployed to `~/.sqliterc`.
- `dot_terraform.d/` — Terraform provider plugin cache deployed to `~/.terraform.d/`.
- `dot_terraformrc` — Terraform CLI configuration deployed to `~/.terraformrc`.
- `dprint.json` — Layout settings and format plugins for dprint code formatter.
- `go.work` — Go workspace file targeting the `dot` CLI package.
- `install.sh` — Bootstrapping shell script installing mise and chezmoi.
- `lefthook.yml` — Lefthook Git hooks manager settings (pre-commit, pre-push, post-commit).
- `LICENSE` — MIT License file.
- `mise.lock` — Cross-platform lockfile for repository-scoped mise toolchain.
- `mise.toml` — Project-scoped task definitions and toolchain configuration.
- `modify_dot_bashrc` — Partial chezmoi ownership for Bash interactive shell mise activation.
- `modify_dot_profile` — Partial chezmoi ownership for login shell environment and mise activation.
- `README.md` — Human-centric documentation detailing requirements, installation, and setup.
- `ruff.toml` — Python linter and formatter configuration for Ruff.
- `run_once_after_install-antigravity-cli.sh.tmpl` — Post-install hook for Antigravity CLI.
- `run_once_after_install-grok.sh.tmpl` — Post-install hook for Grok Build CLI.
- `skills/` — Storage directory holding global agent skills symlinked to `~/.agents/skills/`.
- `trivy.yaml` — Security scanner policy configuration for Trivy.
- `verify-lazy-lock.sh` — Fail-closed validation for Lazy plugin checkouts and commits.
