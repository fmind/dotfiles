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

Bootstrap first, then install the managed toolchain from the applied home
config. This keeps the initial bootstrap small while keeping the full tool
install explicit.

### Quick Install

Bootstrap your environment with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/fmind/dotfiles/main/install.sh | bash
```

Then install the managed tools:

```bash
~/.local/bin/mise -C "$HOME/.local/share/chezmoi" run tools
```

The `tools` task reliably installs all tools via a single `mise install` command.
Mise natively handles dependency graphs (such as installing `node` before any `npm:`
packages) because all tool dependencies are explicitly declared in the
`~/.config/mise/config.toml` file.

### Manual Installation

If you prefer a manual setup:

```bash
# 1. Install Chezmoi and Mise
curl -fsSL https://get.chezmoi.io | bash -s -- -b ~/.local/bin
curl -fsSL https://mise.run | bash

# 2. Initialize and apply dotfiles
~/.local/bin/chezmoi init --apply fmind

# 3. Trust the mise configurations
~/.local/bin/mise trust -y ~/.local/share/chezmoi/mise.toml
~/.local/bin/mise trust -y ~/.config/mise/config.toml

# 4. Install the managed toolchain
~/.local/bin/mise -C ~/.local/share/chezmoi run tools
```

The bootstrap step installs base dependencies, `chezmoi`, and `mise`, then
applies the dotfiles. The full managed toolchain remains a separate step.

## Setup Tasks

Manage your environment using built-in `mise` tasks:

```bash
mise run apply      # Apply with chezmoi
mise run check      # Preview chezmoi changes
mise run docker     # Build and run the container
mise run hooks      # Install repository pre-commit hooks
mise run init       # Initialize chezmoi config from template
mise run lock       # Refresh repository and home mise lockfiles
mise run tools      # Install tools from ~/.config/mise/config.toml
mise run trust      # Trust this repository and home mise configurations
mise run upgrade    # Upgrade mise tools to latest versions and lock packages
```

`apply` and `check` do not install tools. Tool installation stays isolated in
`mise run tools`.

From a fresh system, you only need `install.sh` and then `mise run tools`. If
iterating locally, run `chezmoi cd` to edit the template files directly.

## Docker

The Docker image bootstraps `chezmoi`, `mise`, and `fish`, but it does not
install the full home toolchain during the image build.

After starting the container, finalize it with:

```bash
mise -C "$HOME/.local/share/chezmoi" run tools
```

The image keeps `ca-certificates` for HTTPS bootstrap downloads and `libatomic1`
for Node-based tools from the home `mise` config.

## AI Support

This repository includes an `AGENTS.md` file that provides foundational
mandates and technical context for AI agents to follow inside this repo.
