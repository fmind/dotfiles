# Dotfiles

Modern, high-performance dotfiles optimized for AI-driven development and a `fish`-first shell experience.

Managed via [Chezmoi](https://www.chezmoi.io/) and [Mise](https://mise.jdx.dev/).

## Core Principles

- **High Performance**: Modern, lightweight, and fast CLI tools.
- **Vim-Centric**: Native Vim-style keybindings across the entire toolchain.
- **Portable**: Robust support for Linux, macOS (Apple Silicon), and Cloud Shell.
- **AI-Optimized**: Configurations and tools designed for seamless AI agent collaboration.

## Key Tools

- **Shell**: `fish` with `starship` prompt.
- **Editor**: `neovim` (modular and performant).
- **Multiplexer**: `zellij` (modern terminal workspace).
- **File Manager**: `yazi` (terminal file manager with previews).
- **Git**: `lazygit` and `gh` CLI.
- **Utilities**: `bat` (cat), `btop` (monitor), `ripgrep` (search).
- **Runtime**: `mise` for tool versioning and task management.

## Installation

Bootstrap first, then install the managed toolchain from the applied home config. This keeps the initial install deterministic while still making the full tool install a single follow-up command.

### Quick Install

Bootstrap your environment with a single command:

```bash
curl -sL https://raw.githubusercontent.com/fmind/dotfiles/main/install.sh | bash
```

Open a new shell, then finish the toolchain installation:

```bash
mise -C "$HOME" install node python && mise -C "$HOME" install
```

If you already have `gh`, run `gh auth login` first to reduce GitHub API rate limits. If you use a token instead, export `GITHUB_TOKEN` before running the install.

The staged `node python` install is intentional: the repo uses `npm:` and `pipx:` managed tools, and those runtimes need to exist before the full toolchain install can resolve them cleanly on a fresh machine.

The bootstrap script now also trusts the applied `~/.config/mise/config.toml`, so you do not need a separate `mise trust` step after `install.sh`.

### Manual Installation

If you prefer a manual setup:

```bash
# 1. Install Chezmoi and Mise
curl -sS https://get.chezmoi.io | sh -s -- -b ~/.local/bin
curl https://mise.run | sh

# 2. Initialize and apply dotfiles
~/.local/bin/chezmoi init --apply fmind

# 3. Trust the applied global mise config and install the managed toolchain
mise trust ~/.config/mise/config.toml
mise -C "$HOME" install node python && mise -C "$HOME" install
```

The bootstrap step installs only the base dependencies needed to apply the dotfiles. The full managed toolchain is installed explicitly from the applied home config in the final step, so you do not need a second `chezmoi apply` first.

## Setup Tasks

Manage your environment using built-in `mise` tasks:

```bash
mise run bootstrap  # Bootstrap Chezmoi, Mise, and apply dotfiles from a local clone
mise run toolchain  # Install tools from ~/.config/mise/config.toml
mise run apply      # Apply latest configurations
mise run update     # Update from remote and apply
mise run check      # Dry-run changes
mise run docker     # Build and run the bootstrap development container
```

The `bootstrap` task now uses the current checkout as the Chezmoi source instead of cloning the remote repository again, and `mise run toolchain` trusts the applied global config before installing tools.

From a fresh local clone, trust the repo manifest once with `mise trust ./mise.toml` before using repo-local `mise run ...` tasks.

## Docker

The Docker image installs `fish` during the image build so the container can start in the managed shell directly. The optional host CA secret wiring stays in place because local validation still needed it for TLS trust during the bootstrap download step.

After starting the container, finalize it with:

```bash
mise -C "$HOME" install node python && mise -C "$HOME" install
```

If you already have `gh`, run `gh auth login` in the container first to reduce GitHub API rate limits. If you already have `GITHUB_TOKEN` in the host environment, `mise run docker` passes it through to the container automatically.

If your machine exports `SSL_CERT_FILE`, `mise run docker` also forwards it as an optional BuildKit secret so the container can trust the same CA bundle during the build.

## AI Support

This repository includes an `AGENTS.md` file that provides foundational mandates and technical context for AI agents (like Gemini or Copilot) to help them better understand and assist within this environment.
