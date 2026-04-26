# Dotfiles

Modern, high-performance dotfiles tuned for **AI-driven, CLI-first** development on macOS, Linux, and Cloud Shell.

Managed by [chezmoi](https://www.chezmoi.io/) and [mise](https://mise.jdx.dev/).

## Highlights

- **Shell**: `fish` + `starship` + `atuin` + `carapace` + `zoxide` + `fzf`.
- **Terminal**: `ghostty`, with `zellij` as multiplexer and `yazi` as file manager.
- **Editor**: `neovim` (LazyVim) with `catppuccin-mocha`, vim mode everywhere.
- **AI agents**: `claude-code` + `gemini-cli` + `jules` (async) + GitHub Copilot, all sharing one persona via `~/.AGENTS.md`.
- **Languages out of the box**: Python (`uv`), TypeScript / Angular (`pnpm`, `biome`), Go, Terraform, Docker, Kubernetes.
- **Theme**: `catppuccin-mocha` everywhere consistent (nvim, fzf, ghostty, starship, lazygit, gemini, claude, ptpython).
- **Icons**: none — ASCII-only by default for portability over SSH and Cloud Shell.

## Prerequisites

```bash
# Linux only
sudo apt install -y git curl libatomic1 build-essential
# macOS only — install Homebrew + Ghostty if you want a GPU terminal
```

## Installation

The repo is private, so bootstrap from a local clone (requires an SSH key registered with GitHub):

```bash
# 1. Clone into the chezmoi source directory
git clone git@github.com:fmind/dotfiles.git ~/.local/share/chezmoi

# 2. Run the installer (mise → chezmoi → apply)
bash ~/.local/share/chezmoi/install.sh

# 3. Full first-time setup (trust → apply → hooks → tools → vim plugins)
~/.local/bin/mise -C ~/.local/share/chezmoi run bootstrap
```

On a fresh machine (e.g. Cloud Shell) without an SSH key, run `gh auth login` first, then `gh repo clone fmind/dotfiles ~/.local/share/chezmoi` before step 2.

## Tasks

Manage your environment with `mr <task>` (alias for `mise run`):

```text
mr a   apply       Apply configurations with chezmoi
mr b   bootstrap   Full first-time setup
mr c   check       Preview chezmoi changes (dry-run)
mr d   diff        Show pending chezmoi diff
mr h   hooks       Install repository pre-commit hooks
mr i   init        Initialize/regenerate chezmoi config
mr l   lock        Refresh repository and home mise lockfiles
mr n   lint        Run pre-commit on all files
mr o   doctor      Run chezmoi/mise health checks
mr p   prune       Remove unused mise packages and clear cache
mr r   trust       Trust this repository and home mise configurations
mr s   scan        Scan for leaked secrets with gitleaks
mr t   tools       Install global tools from ~/.config/mise/config.toml
mr u   upgrade     Upgrade mise tools to latest versions and bump locks
mr v   vim         Install/sync Neovim plugins headlessly with LazyVim
```

## Repository layout

```text
.
├── install.sh                        # one-shot installer (mise → chezmoi → apply)
├── mise.toml                         # repo-scoped tools + tasks
├── AGENTS.md                         # rules for agents editing THIS repo
├── dot_AGENTS.md                     # global AI persona (~/.AGENTS.md)
├── dot_claude/                       # Claude Code (settings, agents, commands, skills)
├── dot_gemini/                       # Gemini CLI (settings, agents, commands, skills)
├── dot_copilot/config.json           # GitHub Copilot CLI configuration
├── dot_config/
│   ├── mise/config.toml.tmpl         # global toolchain (every CLI you use)
│   ├── fish/                         # shell, aliases, fzf, plugins
│   ├── nvim/                         # LazyVim config, keymaps, plugins
│   ├── starship.toml                 # prompt
│   ├── ghostty/ zellij/ yazi/        # terminal, multiplexer, file manager
│   ├── atuin/ bat/ bottom/ ripgrep/  # CLI tooling
│   ├── lazygit/ lazydocker/ gh/      # git, docker, github
│   └── ptpython/ pnpm/ uv/ fastfetch/ # language + system tooling
├── dot_gitconfig.tmpl                # templated by chezmoi
├── dot_inputrc dot_editrc dot_pypirc # readline / editrc / pypi
└── dot_local/                        # ~/.local
```

## Required API keys & credentials

These are referenced by agents under `dot_claude/agents/` and `dot_gemini/agents/`.

If you use age encryption with chezmoi to manage secrets (like `secrets.fish.age`), provision your private key before applying:

```bash
mkdir -p ~/.config/chezmoi
# Place your age private key in ~/.config/chezmoi/key.txt BEFORE running chezmoi apply
```

```fish
# ~/.config/fish/conf.d/secrets.fish — never committed
set -gx GITHUB_MCP_PAT      "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
set -gx JULES_API_KEY       "jules_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
set -gx STITCH_ACCESS_TOKEN "stitch_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

OAuth-based tools instead need an interactive login the first time:

| Tool                      | One-time command                                             |
| ------------------------- | ------------------------------------------------------------ |
| `gh`                      | `gh auth login`                                              |
| `gcloud`                  | `gcloud auth login && gcloud auth application-default login` |
| `firebase`                | `firebase login`                                             |
| `clasp`                   | `clasp login`                                                |
| `gws`                     | `gws auth login`                                             |
| `claude` (default)        | run `claude` once and pick OAuth                             |
| `gemini` (default)        | run `gemini` once and pick OAuth-personal                    |
| `jules` (if no API key)   | `jules auth login`                                           |

`gcloud auth application-default login` is required by every Google Cloud / Workspace MCP that uses `authProviderType: google_credentials`.

## Update flow

```bash
mr u   # mise upgrade --bump  (locks new versions)
mr a   # chezmoi apply        (deploys updated configs)
mr v   # nvim plugin sync     (LazyVim restore)
```
