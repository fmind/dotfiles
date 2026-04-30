# Dotfiles

Modern, high-performance dotfiles tuned for **AI-driven, CLI-first** development on macOS, Linux, and Cloud Shell.

Managed by [chezmoi](https://www.chezmoi.io/) and [mise](https://mise.jdx.dev/).

## Highlights

- **Shell**: `fish` + `starship` + `atuin` + `carapace` + `zoxide` + `fzf`.
- **Terminal**: `ghostty`, with `zellij` as multiplexer and `yazi` as file manager.
- **Editor**: `neovim` (LazyVim) with `catppuccin-mocha`, vim mode everywhere.
- **AI agents**: `gemini-cli` (with auto-installed extensions: `fgate`, `googleworkspace/cli`, `chrome-devtools-mcp`) + `jules` (async) + GitHub Copilot + Claude Code.
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

The repo is private and SSH-only. Register an SSH key with GitHub before installing.

On a fresh machine run `ssh-keygen` and add the public key at <https://github.com/settings/keys>. Then:

```bash
# 1. Clone into the chezmoi source directory
git clone git@github.com:fmind/dotfiles.git ~/.local/share/chezmoi

# 2. Run the installer (mise → chezmoi → apply)
bash ~/.local/share/chezmoi/install.sh

# 3. Full first-time setup (trust → apply → hooks → tools → vim plugins)
~/.local/bin/mise -C ~/.local/share/chezmoi run full
```

## Tasks

Manage your environment with `mr <task>` (alias for `mise run`):

```text
mr a   apply       Apply configurations with chezmoi
mr c   check       Preview chezmoi changes (dry-run)
mr d   diff        Show pending chezmoi diff
mr f   full        Full first-time setup
mr h   hooks       Install repository pre-commit hooks
mr i   init        Initialize/regenerate chezmoi config
mr l   lock        Refresh repository and home mise lockfiles
mr n   lint        Run pre-commit on all files
mr o   doctor      Run chezmoi/mise health checks
mr p   prune       Remove unused mise packages and clear cache
mr r   trust       Trust this repository and home mise configurations
mr t   tools       Install global tools from ~/.config/mise/config.toml
mr u   upgrade     Upgrade mise tools to latest versions
mr v   vim         Install/sync Neovim plugins headlessly with LazyVim
```

## Required API keys & credentials

If you use age encryption with chezmoi to manage secrets (like `secrets.fish.age`), provision your private key before applying:

```bash
mkdir -p ~/.config/chezmoi
# Place your age private key in ~/.config/chezmoi/key.txt BEFORE running chezmoi apply
chmod 600 ~/.config/chezmoi/key.txt
```

> **Back up `~/.config/chezmoi/key.txt`.** It is **not** managed by chezmoi and never committed. Lose it and every `*.age` file in this repo becomes unrecoverable. Store a copy in a password manager (1Password / Bitwarden secure note) or an offline encrypted vault. The matching public recipient lives in `.chezmoi.toml.tmpl` under `[age].recipient` — that one is fine to share.

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
| `gemini` (default)        | run `gemini` once and pick OAuth-personal                    |
| `jules`                   | `jules auth login`                                           |

`gcloud auth application-default login` is required by every Google Cloud / Workspace MCP that uses `authProviderType: google_credentials`.
