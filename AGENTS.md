# AGENTS.md (Project)

This is `fmind/dotfiles` — a chezmoi + mise dotfiles repo for **AI-CLI-first** development on Linux and Mac OS with Go and Python as the core programming languages.

User-facing install and usage docs live in `README.md`; this file is for agents working inside the repo.

## House rules

- **Chezmoi Force**: Always use `chezmoi apply --force` in scripts/automation to prevent interactive prompt blocks.
- **Edit-source**: Change files in chezmoi source (`~/.local/share/chezmoi/...`), never deployed copies under `~/.gemini` or `~/.config`.
- **GitHub Access**: Use `gh` CLI for repository, issue, and PR operations.
- **Git Push to Main**: Direct commit/push to `main` branch is permitted (no feature branch required).
- **Lint-before-done**: `mise run all` (format + check + test, the same gate CI runs) must pass before reporting a task complete.
- **Markdown Lists**: Only use `1.` for all numbered list items in markdown files to ensure correct dynamic rendering.
- **No-Hard-Wrap**: Every `*.md` keeps each paragraph on a single line.
- **No-Sudo**: Stay user-space; install via `mise`.
- **README Scope**: Keep setup/auth instructions in `README.md`; exclude repository tasks, aliases, and workflows.
- **Secrets**: `*.age` files are encrypted; never modify or commit decrypted versions.
- **Theme**: **Tokyo Night (Moon)** is default across every tool that supports theming.
- **Vim mode**: Enable in every TUI that supports it.

## Conventions

- `dot_foo` → `~/.foo`. Never write a literal leading dot in source paths.
- `<name>.tmpl` → Go-template; branch on `.chezmoi.os` / `.chezmoi.arch`.
- `modify_*` containing `chezmoi:modify-template` → partial file ownership via `.chezmoi.stdin`; omit `.tmpl` suffix.
- `symlink_<name>.tmpl` → symlink target written verbatim into the link.
- `private_*` (mode 0600), `executable_*` (mode 0755), `*.age` (encrypted).
- `run_once_after_*.sh` → executed by `chezmoi apply` once per content hash.
- `run_onchange_after_*.sh` → executed by `chezmoi apply` when content changes.
- `.chezmoiignore` keeps non-deployed repository files out of `apply`.
- `.chezmoi.toml.tmpl` seeds per-machine chezmoi config (git identity, age recipient, editor).

## Workflows

Tasks run via `mise run <task>`. **Do not use `mr`** — it is an interactive-only fish abbreviation.

Aliases split into two namespaces so a mistyped letter can never fire the wrong kind of task:

- **Common tasks** take the canonical one-letter alias from the mise skill: `a` all, `b` build, `c` check, `f` format, `i` install, `t` test, `w` watch (plus `c*`/`f*` for subtasks, e.g. `cg` check:go, `fd` format:dprint).
- **Project management** tasks take an `m`-prefixed alias: `ma` apply, `md` diff, `me` deploy, `mf` full, `mg` completions, `mh` hooks, `mk` lock, `mo` doctor, `mp` prune, `mr` release, `mt` tools, `mtr` trust, `mu` upgrade, `mv` vim, `mw` krew, `mx` verify.

Key routines:

- **First-time setup**: `mise run install` (trust → tools → hooks → vim → krew).
- **Routine update**: `mise run full` (synchronize environment).
- **Iterate**: Edit source → `mise run apply` (`mise run diff` to preview) → `mise run check` (or `mise run all`) → `mise run verify`.
- **Add tool**: Append to `dot_config/mise/config.toml.tmpl` → `mise run tools` → `mise run lock`.
- **Upgrade tools**: `mise run upgrade` (upgrades tool pins and lockfiles).
- **Reclaim disk**: `mise run prune` (reclaim development caches and agent transcripts).
- **Release**: `mise run release` (runs validation, tags, pushes `main` and tag).
- **Manage skills**: Author directly under `skills/`; validate with `gh skill publish --dry-run`.
- **Create visuals**: Use `fmind-visuals` skill (Slidev for decks, Mermaid for diagrams).

> Note: If `mise` fails with `command not found` in an agent shell, call `~/.local/bin/mise` directly.

The unified `dot` CLI command-line utility (source in `dot/`) is compiled to `~/.local/bin/dot`. Detailed CLI configuration and subcommand flags are documented in [`skills/dot-cli/SKILL.md`](skills/dot-cli/SKILL.md).

- `dot verify` (alias `v`) — Sanity checks on environment, tools, secrets, and install freshness (`HEAD`).
- `dot pull` (alias `p`) — Pull development repositories in `~/.config/dot.yaml` (`--push` to push).
- `dot commit` (alias `c`) — Generate Conventional Commit from git diff via AI (`agy`).
- `dot cluster` (alias `k`) — Manage local k3d Kubernetes cluster and output diagnostic evidence bundles.
- `dot login` (alias `l`) — OAuth login wrapper for `github`, `workspace`, `gcp`, or `clasp`.
- `dot setup` (alias `u`) — Enable GCP APIs on active Google Workspace project.
- `dot completion` (alias `g`) — Generate fish autocompletions.
- `dot pull-request` (alias `pr`) — Generate PR description via AI and trigger `gh pr create`.
- `dot release` (alias `r`) — Prepare, tag, and push release commit/tag (`main == origin/main`).
- `dot status` (alias `s`) — Unified summary of git repos, docker containers, and k3d cluster (`--json`).
- `dot agent` (alias `a`) — Normalize session transcripts into `~/.agents/sessions/v1/`, run `doctor`, `hook`, `session`.
- `dot notify` (alias `n`) — Desktop notification for agent hook events or custom alerts.
- `dot prune` (alias `x`) — Reclaim disk space from agent transcripts and build/tool caches.
- `dot chezmoi clean` (alias `m c`) — Scan and clean unmanaged home directory orphan files.
- `dot config` (alias `f`) — Inspect, scaffold, edit, and validate `~/.config/dot.yaml`.
- `dot context` (alias `t`) — Emit bounded, secret-scanned project context pack (Markdown or JSON).
- `dot version` (alias `i`) — Print binary version enriched with embedded VCS revision.

## Agents

Two assets are authored once and consumed by all agent CLIs:

- **Persona** — `dot_agents/AGENTS.md` deploys to `~/.agents/AGENTS.md` (symlinked by Codex, Antigravity, OpenCode, Claude, Copilot).
- **Skills** — `skills/` is symlinked to `~/.agents/skills/` (consumed by all agent CLIs).

**Rule: every global skill lives in `skills/`.**

## Layout

- `.agents/` — Workspace-scoped state, session records, and scratch scripts for AI agents.
- `.antigravitycli/` — Workspace-scoped session records and state for Antigravity CLI.
- `.chezmoi.toml.tmpl` — Host-specific chezmoi configuration template.
- `.chezmoiignore` — Chezmoi exclude rules for non-deployed repository files.
- `.claude/` — Workspace-scoped session records and state for Claude Code CLI.
- `.gemini/` — Workspace configurations and metadata for Antigravity CLI.
- `.github/` — GitHub Actions workflows (`ci.yml`, `cd.yml`, `security.yml`) and Dependabot config.
- `.gitignore` — Git exclusion rules.
- `.gitleaks.toml` — Security configuration and secrets scanner allowlist for GitLeaks.
- `.stylua.toml` — StyleLua formatting policy for managed Neovim Lua sources.
- `AGENTS.md` (this file) — Repository guide, conventions, and layout for AI agents.
- `CHANGELOG.md` — Versioned release history generated from Conventional Commits.
- `dot/` — Go CLI source package containing the unified `dot` command-line utility.
- `dot_agents/` — Source folder containing unified global instructions (`AGENTS.md`) and skills symlink template.
- `dot_claude/` — Claude Code CLI settings template and persona/skills symlinks.
- `dot_codex/` — OpenAI Codex CLI partial configuration modifier and persona symlink.
- `dot_config/` — Custom configuration templates deployed to `~/.config/`.
- `dot_copilot/` — GitHub Copilot CLI integration configurations and symlink templates.
- `dot_duckdbrc` — DuckDB CLI settings deployed to `~/.duckdbrc`.
- `dot_gemini/` — Antigravity CLI config settings and symlinks deployed to `~/.gemini/`.
- `dot_gitconfig.tmpl` — Global Git configuration template deployed to `~/.gitconfig`.
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
- `go.work.sum` — Go workspace dependency checksum file.
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
- `skills/` — Storage directory holding global agent skills symlinked to `~/.agents/skills/`.
- `trivy.yaml` — Security scanner policy configuration for Trivy.
