# Dotfiles

Modern, high-performance dotfiles tuned for **AI-driven, CLI-first** development on macOS, Linux, and Cloud Shell.

Managed by [chezmoi](https://www.chezmoi.io/) and [mise](https://mise.jdx.dev/).

## Highlights

- **Shell:** `fish` + `starship` + `atuin` + `carapace` + `zoxide` + `fzf`.
- **Terminal:** `ghostty`, with `zellij` as multiplexer.
- **Editor:** `neovim` (LazyVim) with `catppuccin-mocha`, vim mode everywhere.
- **AI:** `gemini-cli` + `jules` (async) + GitHub Copilot, with curated subagents, skills, and slash-commands under [`dot_gemini/`](dot_gemini/).
- **Languages out of the box:** Python (`uv`), TypeScript / Angular (`pnpm`, `biome`), Go, Terraform, Docker, Kubernetes.
- **Theme:** `catppuccin-mocha` everywhere consistent (nvim, fzf, ghostty, starship, lazygit, gemini, ptpython).
- **Icons:** none — ASCII-only by default for portability over SSH and Cloud Shell.

## Prerequisites

```bash
# Linux only
sudo apt install -y git curl libatomic1 build-essential
# macOS only — install Homebrew + Ghostty if you want a GPU terminal
```

## Installation

```bash
# 1. Bootstrap chezmoi + dotfiles
curl -fsSL https://raw.githubusercontent.com/fmind/dotfiles/main/install.sh | bash

# 2. Full first-time setup (trust → install tools → apply → plugins → hooks)
~/.local/bin/mise -C ~/.local/share/chezmoi run bootstrap
```

## Tasks

Manage your environment with `mr <task>` (alias for `mise run`):

```bash
mr a   apply       Apply configurations with chezmoi
mr c   check       Preview chezmoi changes (dry-run)
mr d   diff        Show pending chezmoi diff
mr D   doctor      Run chezmoi/mise/nvim health checks
mr f   fish        Fish syntax check on every config fragment
mr h   hooks       Install repository pre-commit hooks
mr i   init        Initialize chezmoi config from template
mr L   lint        Run pre-commit on all files
mr l   lock        Refresh repository and home mise lockfiles
mr p   prune       Remove unused mise packages and clear cache
mr s   scan        Scan for leaked secrets with gitleaks
mr t   tools       Install global tools from ~/.config/mise/config.toml
mr r   trust       Trust this repository and home mise configurations
mr u   upgrade     Upgrade mise tools to latest versions and bump locks
mr v   vim         Install/sync Neovim plugins headlessly with LazyVim
mr b   bootstrap   Full first-time setup
```

## Repository layout

```text
.
├── install.sh                         # one-shot installer (mise → chezmoi → apply)
├── mise.toml                          # repo-scoped tools + tasks
├── dot_config/
│   ├── fish/   ...                    # shell, aliases, fzf, plugins
│   ├── nvim/   ...                    # LazyVim config, keymaps, plugins
│   ├── mise/config.toml               # global toolchain (every CLI you use)
│   ├── zellij/ ghostty/ yazi/ ...     # TUI clients
│   └── starship.toml lazygit/ gh/ ... # prompt + git tooling
├── dot_gemini/                        # Gemini CLI configuration (settings, agents, skills, commands)
├── dot_copilot/config.json            # GitHub Copilot CLI configuration
├── dot_gitconfig.tmpl                 # templated by chezmoi
└── vscode-settings.json               # snapshot of legacy VS Code settings (reference only)
```

## Required API keys & credentials

These are referenced by agents in `dot_gemini/agents/`.

```fish
# ~/.private.fish — never committed
set -gx GITHUB_MCP_PAT      "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
set -gx JULES_API_KEY       "jules_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
set -gx STITCH_ACCESS_TOKEN "stitch_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

OAuth-based tools instead need an interactive login the first time:

| Tool                      | One-time command                                |
| ------------------------- | ----------------------------------------------- |
| `gh`                      | `gh auth login`                                 |
| `gcloud`                  | `gcloud auth login && gcloud auth application-default login` |
| `firebase`                | `firebase login`                                |
| `clasp`                   | `clasp login`                                   |
| `gws`                     | `gws auth login`                                |
| `gemini` (default)        | run `gemini` once and pick OAuth-personal       |
| `jules` (if not API key)  | `jules auth login`                              |

`gcloud auth application-default login` is required by every Google Cloud / Workspace MCP that uses `authProviderType: google_credentials`.

## Update flow

```bash
mr u   # mise upgrade --bump  (locks new versions)
mr a   # chezmoi apply        (deploys updated configs)
mr v   # nvim plugin          (sync LazyVim)
```
