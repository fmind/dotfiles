# AGENTS.md (Project)

This is `fmind/dotfiles` — a chezmoi + mise dotfiles repo for **AI-CLI-first** development on Linux and Mac OS with Go and Python as the core programming languages.

User-facing install and usage docs live in `README.md`; this file is for agents working inside the repo.

## House rules

- **Chezmoi Force**: always use `chezmoi apply --force` when applying changes in automation/scripts to prevent getting blocked by interactive prompts when target files have changed.
- **Edit-source**: change files in this chezmoi source (`~/.local/share/chezmoi/...`), never their deployed copies under `~/.gemini` or `~/.config`.
- **GitHub Access**: use the `gh` CLI for all repository, issue, and PR operations.
- **Git Push to Main**: it is allowed to commit and push directly to the `main` branch (no need to create a feature branch first).
- **Lint-before-done**: `mr ma` must pass before reporting a task complete.
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
- `.chezmoiignore` keeps repo-only files (the `/dot` Go CLI, `/skills`, `/AGENTS.md`, `README.md`, `LICENSE`, `install.sh`, `mise.toml`/`mise.lock`, `lefthook.yml`, `dprint.json`, `ruff.toml`, and `go.work`/`go.work.sum`) out of `apply`, plus the Ghostty `.desktop` file on non-Linux hosts and `secrets.fish` when the age key is absent.
- `.chezmoi.toml.tmpl` seeds the per-machine chezmoi config (git identity, age recipient, editor/cd/diff/merge commands) on `chezmoi init`.

## Workflows

- Tasks run via `mr <alias>` (= `mise run`):
  - **First-time setup**: `mr i` (trust → tools → hooks → vim → krew).
  - **Routine update**: `mr f` (fast standard routine synchronization).
  - **Iterate**: edit source → `mr a` to apply (`mr d` to preview the diff) → `mr mc` for quick static checks (or `mr ma` for the full pre-commit + pre-push gate) → `mr x` to verify dotfiles sanity.
  - **Add a tool**: append to `dot_config/mise/config.toml.tmpl` (alphabetical) — use `mise registry` to find tools → `mr t` to deploy and install → `mr k` to refresh and stage the lockfile.
  - **Upgrade tools**: `mr u` bumps versions, re-locks, re-applies.
  - **Release**: `mr r` bumps the version in `dot/version.go`, updates `CHANGELOG.md`, tags, pushes, and publishes a GitHub release using `git-cliff` and `gh`.
  - **Manage skills**: author first-party skills directly under `skills/`, review every external skill and bundled script before installation, and validate the collection with `gh skill publish --dry-run`.
  - **Custom AI Utilities**: Deployed via `dot_local/bin/` to `~/.local/bin/` (e.g. `dot` CLI) and added to PATH.

- The unified `dot` CLI command-line utility (source in `dot/`) is compiled to `~/.local/bin/dot` and provides the following subcommands:
  - `dot verify` (alias `v`) — Runs sanity checks on system environments, CLI tool installations, and secret configurations.
  - `dot pull` (alias `p`) — Concurrently pulls all active development Git repositories defined in `~/.config/dot.yaml`; `--push`/`-P` also pushes clean repositories that are ahead of their upstream.
  - `dot commit` (alias `c`) — Automatically generates and applies a Conventional Commit message from current git diffs via `agy`.
  - `dot cluster` (alias `k`) — Creates, starts, stops, or inspects the shared local k3d Kubernetes cluster.
  - `dot login` (alias `l`) — Interactive OAuth login wrapper command targeting `github` (via `gh`), `workspace` (via `gws`), `gcp` (via `gcloud` user and Application Default Credentials), or `clasp` (via `clasp login`).
  - `dot setup` (alias `u`) — Custom setup wrapper to enable APIs on the active GCP Google Workspace project.
  - `dot completion` (aliases `g`, `completions`) — Automatically generates fish autocompletions for dot itself and external CLI tools.
  - `dot pr` (alias `pr`) — Generates a structured pull request description via AI and triggers `gh pr create`.
  - `dot release` (alias `r`) — Bumps the version in `dot/version.go`, updates `CHANGELOG.md`, tags, pushes, and publishes a GitHub release.
  - `dot status` (alias `s`) — Provides a unified summary status of local development Git repositories, active docker containers, and local k3d Kubernetes configurations; supports `--json`/`-j` for scripting.
  - `dot agent` (alias `a`) — Normalizes agent session transcripts into `~/.agents/sessions/`. `agy`, `claude`, and `codex` are wired to each tool's `Stop` hook; `opencode` fires from its `session.idle` plugin; `copilot` has no live hook API, so its `~/.copilot/session-store.db` is captured by `dot agent session sync`. `sync` also backfills every source's untracked sessions and `clean` prunes logs past a retention window.
  - `dot chezmoi clean` (group alias `m`, subcommand aliases `c`, `cc`) — Scans for previously managed chezmoi files and cleans up unmanaged orphans in home directory.
  - `dot config` (alias `f`) — Inspects, scaffolds, edits, and validates the `~/.config/dot.yaml` configuration file (`show`, `path`, `init`, `edit`, `validate`).
  - `dot version` (alias `n`) — Prints the version enriched with the embedded VCS revision so an installed binary can be matched against the current sources.

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
- `.chezmoiignore` — Chezmoi exclude patterns to ignore repository files from deployment.
- `.chezmoi.toml.tmpl` — Template config initialized as the host-specific chezmoi configuration.
- `.claude/` — Workspace-scoped session records and state for the Claude Code CLI.
- `.gemini/` — Workspace configurations and metadata for the Antigravity CLI.
- `.github/` — GitHub Actions CI and Dependabot dependency-update configuration.
- `.gitignore` — Git pattern definitions to exclude files from version control.
- `.gitleaks.toml` — Security configuration and secrets scanner allowlist for GitLeaks.
- `AGENTS.md` (this file) — Repository guide, conventions, and instruction guidelines for AI agents.
- `dot/` — Go CLI source package containing the unified `dot` command-line utility.
- `dot_agents/` — Source folder containing unified global instructions (`AGENTS.md`) and the canonical skills symlink template.
- `dot_claude/` — Claude Code CLI configuration template and symlinks.
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
- `mise.toml` — Project-scoped task definitions and mise configuration for task runs.
- `README.md` — Human-centric documentation detailing requirements, installation steps, and secrets.
- `run_once_after_install-antigravity-cli.sh.tmpl` — Post-install hook script to automate Antigravity CLI installation.
- `skills/` — Storage directory holding global agent skills symlinked into active agent directories.
