# AGENTS.md (Project)

This is `fmind/dotfiles` — a chezmoi + mise dotfiles repo for **AI-CLI-first** development on Linux and Mac OS with Go and Python as the core programming languages.

User-facing install and usage docs live in `README.md`; this file is for agents working inside the repo.

## House rules

- **Chezmoi Force**: always use `chezmoi apply --force` when applying changes in automation/scripts to prevent getting blocked by interactive prompts when target files have changed.
- **Edit-source**: change files in this chezmoi source (`~/.local/share/chezmoi/...`), never their deployed copies under `~/.gemini` or `~/.config`.
- **GitHub Access**: use the `gh` CLI for all repository, issue, and PR operations.
- **Git Push to Main**: it is allowed to commit and push directly to the `main` branch (no need to create a feature branch first).
- **Lint-before-done**: `mise run all` (format + check + test, the same gate CI runs) must pass before reporting a task complete.
- **Markdown Lists**: only use `1.` for all numbered list items in markdown files (e.g. `1. first`, `1. second`) to ensure correct dynamic rendering.
- **No-Hard-Wrap**: every `*.md` keeps each paragraph on a single line.
- **No-Sudo**: stay user-space; install via `mise`.
- **README Scope**: only keep setup-related instructions (installation, ssh, API, credentials, auth) in `README.md` and avoid describing repository tasks, shortcuts, aliases, or workflows.
- **Secrets**: `*.age` files are encrypted; never modify or commit decrypted versions.
- **Theme**: **Tokyo Night (Moon)** is the default across every tool that supports theming (small tools that only inherit terminal ANSI colors are left to the terminal).
- **Vim mode**: enable in every TUI that supports it.

## Conventions

- `dot_foo` → `~/.foo`. Never write a literal leading dot in source paths.
- `<name>.tmpl` → Go-template; branch on `.chezmoi.os` / `.chezmoi.arch`.
- `modify_*` containing `chezmoi:modify-template` → partial file ownership with the existing target in `.chezmoi.stdin`; these files must not use a `.tmpl` suffix.
- `symlink_<name>.tmpl` → symlink target written verbatim into the link.
- `private_*` → mode 0600. `executable_*` → mode 0755. `*.age` → encrypted.
- `run_once_after_*.sh` → executed by `chezmoi apply` once per unique content hash; use it for one-shot install/bootstrap steps.
- `run_onchange_after_*.sh` → executed by `chezmoi apply` after files are written, only when the script's content changes.
- `.chezmoiignore` keeps repo-only files (the `/dot` Go CLI, `/skills`, `/AGENTS.md`, `README.md`, `LICENSE`, `CHANGELOG.md`, `install.sh`, `mise.toml`/`mise.lock`, `lefthook.yml`, `dprint.json`, `ruff.toml`, `trivy.yaml`, and `go.work`/`go.work.sum`) out of `apply`, plus the Ghostty `.desktop` file on non-Linux hosts and `secrets.fish` when the age key is absent.
- `.chezmoi.toml.tmpl` seeds the per-machine chezmoi config (git identity, age recipient, editor/cd/diff/merge commands) on `chezmoi init`.

## Workflows

Tasks run via `mise run <task>`. **Do not use `mr`** — it is a fish abbreviation defined under `status is-interactive`, so it expands only at a human prompt and is an unknown command in every script, hook, and agent shell.

Aliases split into two namespaces so a mistyped letter can never fire the wrong kind of task:

- **Common tasks** take the canonical one-letter alias from the mise skill: `a` all, `b` build, `c` check, `f` format, `i` install, `t` test, `w` watch (plus `c*`/`f*` for subtasks, e.g. `cg` check:go, `fd` format:dprint).
- **Project management** tasks take an `m`-prefixed alias: `ma` apply, `md` diff, `mf` full, `mg` completions, `mh` hooks, `mk` lock, `mo` doctor, `mp` prune, `mpa` prune:agents, `mpr` preview, `mr` release, `msk` skills, `mt` tools, `mtr` trust, `mu` upgrade, `mv` vim, `mw` krew, `mx` verify.

- **First-time setup**: `mise run install` (trust → tools → hooks → vim → krew).
- **Routine update**: `mise run full` (fast standard routine synchronization).
- **Iterate**: edit source → `mise run apply` (`mise run diff` to preview) → `mise run check` for quick static checks (or `mise run all` for the full CI gate) → `mise run verify` for dotfiles sanity.
- **Add a tool**: append to `dot_config/mise/config.toml.tmpl` (alphabetical) — use `mise registry` to find tools → `mise run tools` to deploy and install → `mise run lock` to refresh and stage the lockfile.
- **Upgrade tools**: `mise run upgrade` bumps versions, updates Neovim plugins, re-locks (`mise.lock` + `lazy-lock.json`), re-applies.
- **Reclaim disk**: `mise run prune:agents` while local k3d clusters or a warm Go cache still matter; `mise run prune` (`--all=deep`) otherwise.
- **Release**: `mise run release` bumps the version in `dot/version.go`, updates `CHANGELOG.md`, tags, pushes, and publishes a GitHub release using `git-cliff` and `gh`.
- **Manage skills**: author first-party skills directly under `skills/` and validate with `gh skill publish --dry-run`. No external skill is vendored here; install reviewed upstream ones on demand with `skills add <repo> --all -y` (candidates are listed in the `agent-skills` skill).
- **Create visuals**: use `fmind-visuals` for the brand contract and routing; Slidev is the only default for new decks, Mermaid is the default for diagrams, LikeC4 remains the architecture-model option, and D2 remains the bespoke composition option.
- **Custom AI Utilities**: Deployed via `dot_local/bin/` to `~/.local/bin/` (e.g. `dot` CLI) and added to PATH.

> If `mise` itself fails with `command not found` in an agent shell, the harness captured mise's shell function without `__MISE_EXE`; call `~/.local/bin/mise` directly.

- The unified `dot` CLI command-line utility (source in `dot/`) is compiled to `~/.local/bin/dot` and provides the following subcommands:
  - `dot verify` (alias `v`) — Runs sanity checks on system environments, CLI tool installations, secret configurations, and install freshness (the deployed binary against the source checkout's `HEAD`).
  - `dot pull` (alias `p`) — Concurrently pulls all active development Git repositories defined in `~/.config/dot.yaml`; `--push`/`-P` also pushes clean repositories that are ahead of their upstream.
  - `dot commit` (alias `c`) — Automatically generates and applies a Conventional Commit message from current git diffs via `agy`.
  - `dot cluster` (alias `k`) — Creates, starts, stops, or inspects the shared local k3d Kubernetes cluster.
  - `dot login` (alias `l`) — Interactive OAuth login wrapper command targeting `github` (via `gh`), `workspace` (via `gws`), `gcp` (via `gcloud` user and Application Default Credentials), or `clasp` (via `clasp login`).
  - `dot setup` (alias `u`) — Custom setup wrapper to enable APIs on the active GCP Google Workspace project.
  - `dot completion` (alias `g`) — Automatically generates fish autocompletions for dot itself and external CLI tools.
  - `dot pull-request` (alias `pr`) — Generates a structured pull request description via AI and triggers `gh pr create`.
  - `dot release` (alias `r`) — Bumps the version in `dot/version.go`, updates `CHANGELOG.md`, tags, pushes, and publishes a GitHub release.
  - `dot status` (alias `s`) — Provides a unified summary status of local development Git repositories, active docker containers, and local k3d Kubernetes configurations; supports `--json`/`-j` for scripting.
  - `dot agent` (alias `a`) — Normalizes agent session transcripts into `~/.agents/sessions/`. `agy`, `claude`, and `codex` are wired to each tool's `Stop` hook; `opencode` fires from its `session.idle` plugin; `copilot` has no live hook API, so its `~/.copilot/session-store.db` is captured by `dot agent session sync`. `sync` also backfills every source's untracked sessions. The command only gathers: deleting expired session logs belongs to `dot prune --agents`.
  - `dot notify` (alias `n`) — Sends an OS-independent desktop notification for agent hook events (`<agent> <stop|needs-input>`) or custom alerts (`<summary> [headline] [details...]`), naming the project and zellij pane to return to so background agents announce themselves instead of waiting to be checked; Claude, Codex, and Antigravity fire it from their `Stop` hooks when a turn is finished, and Claude also fires `needs-input` from `Notification` so a blocked agent surfaces instead of idling.
  - `dot prune` (alias `x`) — Reclaims disk space from agent session logs and development caches, and owns all session retention (both the raw per-agent stores and `~/.agents/sessions`, each with its own `keep_days` under `prune.agents.sessions`). Targets compose as flags (`--agents`, `--docker`, `--go`, `--python`, `--node`, `--mise`, `--tools`, or `--all`) and each accepts an optional depth (`--docker=system`, `--go=module`, `--all=deep`); every target has a `prune.<target>` config section carrying its default depth and cache paths, `--dry-run` reports without deleting, and `--days` overrides every configured retention (`--days 0` empties the stores).
  - `dot chezmoi clean` (group alias `m`, subcommand aliases `c`, `cc`) — Scans for previously managed chezmoi files and cleans up unmanaged orphans in home directory.
  - `dot config` (alias `f`) — Inspects, scaffolds, edits, and validates the `~/.config/dot.yaml` configuration file (`show`, `path`, `init`, `edit`, `validate`).
  - `dot context` — Emits a deterministic project-only context pack as Markdown or versioned JSON, within an explicit byte or approximate token budget; collectors and sensitive path/environment patterns are allowlisted through `context` in `~/.config/dot.yaml`, and the exact final payload is secret-scanned before output.
  - `dot version` (alias `i`) — Prints the version enriched with the embedded VCS revision so an installed binary can be matched against the current sources.

## Agents

Two assets are authored once and consumed by all agent CLIs through native discovery, symlinks, or deterministic synchronization:

- **Persona** — `dot_agents/AGENTS.md` deploys to `~/.agents/AGENTS.md`.
  - Codex consumes it via a symlink at `~/.codex/AGENTS.md` pointing to `~/.agents/AGENTS.md`.
  - Antigravity consumes it via a symlink at `~/.gemini/GEMINI.md` pointing to `~/.agents/AGENTS.md`.
  - OpenCode consumes it via the `instructions` option in `opencode.json` pointing to `~/.agents/AGENTS.md`.
  - Claude consumes it via a symlink at `~/.claude/CLAUDE.md` pointing to `~/.agents/AGENTS.md`.
  - Copilot consumes it via a symlink at `~/.copilot/copilot-instructions.md` pointing to `~/.agents/AGENTS.md`.
- **Skills** — `skills/` (at the root of this repository) is the canonical home and is symlinked to `~/.agents/skills/` via a chezmoi symlink template.
  - Codex, OpenCode, and GitHub Copilot CLI discover `~/.agents/skills/<name>/SKILL.md` natively.
  - Claude consumes the canonical directory through `dot_claude/symlink_skills.tmpl`, which deploys `~/.claude/skills` as a symlink to `~/.agents/skills`.
  - Antigravity products discover shared global skills from `~/.gemini/config/skills` via a symlink template.

**Rule: every global skill lives in `skills/`.**

## Layout

- `.agents/` — Workspace-scoped state, session records, and scratch scripts for AI agents.
- `.antigravitycli/` — Workspace-scoped session records, configuration settings, and state for Antigravity CLI.
- `.chezmoi.toml.tmpl` — Template config initialized as the host-specific chezmoi configuration.
- `.chezmoiignore` — Chezmoi exclude patterns to ignore repository files from deployment.
- `.chezmoiremove` — Paths chezmoi deletes from the destination, so retired dotfiles disappear on every machine.
- `.claude/` — Workspace-scoped session records and state for the Claude Code CLI.
- `.gemini/` — Workspace configurations and metadata for the Antigravity CLI.
- `.github/` — GitHub Actions CI and Dependabot dependency-update configuration.
- `.gitignore` — Git pattern definitions to exclude files from version control.
- `.gitleaks.toml` — Security configuration and secrets scanner allowlist for GitLeaks.
- `.stylua.toml` — Two-space StyleLua formatting policy for managed Neovim Lua sources (chezmoi ignores dot-prefixed source files automatically).
- `AGENTS.md` (this file) — Repository guide, conventions, and instruction guidelines for AI agents.
- `CHANGELOG.md` — Versioned release history generated from Conventional Commits.
- `dot/` — Go CLI source package containing the unified `dot` command-line utility.
- `dot_agents/` — Source folder containing unified global instructions (`AGENTS.md`) and the canonical skills symlink template.
- `dot_claude/` — Claude Code CLI settings template plus the persona and skills symlinks. `settings.json.tmpl` fully owns `~/.claude/settings.json`: it is a plain file, not a `modify_` template, so `chezmoi apply` resets whatever `/model` and the effort picker last wrote back to the repo defaults (Opus 5 1M, high effort). Change the default here, not at runtime.
- `dot_codex/` — OpenAI Codex CLI partial configuration modifier plus the shared persona symlink into `~/.codex/`; runtime model and trust state are preserved across applies.
- `dot_config/` — Custom configuration templates deployed to the user's `~/.config/` directory.
- `dot_copilot/` — GitHub Copilot CLI integration configurations and symlink templates.
- `dot_duckdbrc` — DuckDB CLI settings deployed to `~/.duckdbrc`.
- `dot_gemini/` — Antigravity CLI config settings and symlinks deployed to `~/.gemini/`.
- `dot_gitconfig.tmpl` — Global Git configuration template deployed to `~/.gitconfig`.
- `dot_inputrc` — GNU Readline configurations deployed to `~/.inputrc` for prompt styling.
- `dot_kube/` — kubectl settings deployed to `~/.kube/` (`kuberc` defaults and kubecolor `color.yaml`).
- `dot_local/` — Executables, application configurations, and helpers deployed to `~/.local/`.
- `dot_npmrc` — npm configuration deployed to `~/.npmrc`.
- `dot_skaffold/` — Skaffold settings deployed to `~/.skaffold/` for local Kubernetes development.
- `dot_sqliterc` — SQLite interactive shell settings deployed to `~/.sqliterc`.
- `dot_terraform.d/` — Terraform/OpenTofu CLI data deployed to `~/.terraform.d/` (provider plugin cache).
- `dot_terraformrc` — Terraform/OpenTofu CLI configuration deployed to `~/.terraformrc`.
- `dprint.json` — Layout settings and format plugins configured for the dprint code formatter.
- `go.work` — Go workspace file targeting the `dot` CLI package.
- `go.work.sum` — Go workspace dependency lock file.
- `install.sh` — Bootstrapping shell script to install mise and chezmoi, and apply dotfiles.
- `lefthook.yml` — Lefthook Git hooks manager settings for automated formatting, linting, and testing.
- `LICENSE` — MIT License file governing use of the dotfiles repository.
- `mise.lock` — Cross-platform lockfile for the repository-scoped mise toolchain.
- `mise.toml` — Project-scoped task definitions and mise configuration for task runs.
- `modify_dot_bashrc` — Partial chezmoi ownership of Bash interactive-shell mise activation.
- `modify_dot_profile` — Partial chezmoi ownership of login-shell mise paths and activation.
- `README.md` — Human-centric documentation detailing requirements, installation steps, and secrets.
- `ruff.toml` — Python linter and formatter configuration for Ruff.
- `run_once_after_install-antigravity-cli.sh.tmpl` — Post-install hook script to automate Antigravity CLI installation.
- `skills/` — Storage directory holding global agent skills symlinked into active agent directories.
- `trivy.yaml` — Security scanner policy configuration for Trivy vulnerabilities, misconfigurations, and secrets.
