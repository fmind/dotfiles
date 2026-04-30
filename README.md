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

Just run:

```bash
# 1. Clone into the chezmoi source directory
git clone https://github.com/fmind/dotfiles.git ~/.local/share/chezmoi

# 2. Run the installer (mise → chezmoi → apply)
bash ~/.local/share/chezmoi/install.sh

# 3. Full first-time setup (trust → apply → hooks → tools → vim plugins)
~/.local/bin/mise -C ~/.local/share/chezmoi run full
```

> To push changes back, swap the remote to SSH once your key is registered at <https://github.com/settings/keys>:
>
> ```bash
> git -C ~/.local/share/chezmoi remote set-url origin git@github.com:fmind/dotfiles.git
> ```

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
mr e   verify      Verify environment readiness (logins, secrets, sync, extensions)
mr v   vim         Install/sync Neovim plugins headlessly with LazyVim
```

## Required API keys & credentials

If you use age encryption with chezmoi to manage secrets (like `secrets.fish.age`), provision your private key before applying:

```bash
mkdir -p ~/.config/chezmoi
# Place your age private key in ~/.config/chezmoi/key.txt BEFORE running chezmoi apply
chmod 600 ~/.config/chezmoi/key.txt
```

> **Back up `~/.config/chezmoi/key.txt`.** It is **not** managed by chezmoi and never committed. Lose it and every `*.age` file in this repo becomes unrecoverable.
> Store a copy in a password manager (1Password / Bitwarden secure note) or an offline encrypted vault.

Static API keys go in a machine-local fish file — never committed. Two locations are sourced automatically:

- `~/.private.fish` — sourced from `~/.config/fish/config.fish` (recommended for machine-specific secrets).
- `~/.config/fish/conf.d/*.fish` — auto-loaded by fish (e.g. a `secrets.fish`).

```fish
set -gx JULES_API_KEY              "jules_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
set -gx STITCH_ACCESS_TOKEN        "stitch_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
set -gx GOOGLE_OAUTH_CLIENT_ID     "xxxxxxxxxx.apps.googleusercontent.com"
set -gx GOOGLE_OAUTH_CLIENT_SECRET "GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx"
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

### OAuth client for Workspace MCP

The Workspace MCP servers (Calendar, Chat, Drive, Gmail, People) auth via a per-user OAuth 2.0 flow — **not** ADC.

You need a "Desktop app" OAuth client.  You only do this **once per Google account**; the same client ID is reused across every project.

1. Pick or create a GCP project, then point gcloud at it: `gcloud config set project <PROJECT_ID>`.
2. Enable the Workspace APIs and check billing:

   ```bash
   gcp-dotfiles-setup            # idempotent; enables every Google API used by the MCPs
   ```

3. Configure the **OAuth consent screen** at <https://console.cloud.google.com/auth/overview>:
   - User type: **Internal** if you own a Workspace org (recommended — no verification), otherwise **External** + add yourself as a test user.
   - App name + support email — anything sensible; only you will see this.
   - Add the sensitive Workspace scopes you plan to use (Gmail, Calendar, Drive, Chat, People). The minimum is the read-only set listed in `dot_gemini/skills/install-workspace-mcp/SKILL.md`.
4. Create the **OAuth 2.0 Client ID** at <https://console.cloud.google.com/auth/clients>:
   - Application type: **Desktop app**.
   - Name: e.g. `gemini-cli-workspace-mcp`.
   - Click **Download JSON** — the file lands as `client_secret_<id>.json` in `~/Downloads`.
5. Copy the values into `~/.private.fish`:

   ```fish
   set -gx GOOGLE_OAUTH_CLIENT_ID     "<client_id from JSON>"
   set -gx GOOGLE_OAUTH_CLIENT_SECRET "<client_secret from JSON>"
   ```

6. In any project that pins a Workspace MCP server in `.gemini/settings.json`, run `/mcp auth <server-name>` inside Gemini CLI to trigger the browser consent flow.

Delete the downloaded `client_secret_*.json` once the values are in `~/.private.fish`.
