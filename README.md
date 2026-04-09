# Dotfiles

Modern, high-performance dotfiles optimized for AI-driven development.

Managed via [Chezmoi](https://www.chezmoi.io/) and [Mise](https://mise.jdx.dev/).

## Principles

- **High Performance**: Modern, lightweight, and fast CLI tools.
- **Vim-Centric**: Native Vim-style keybindings across the entire toolchain.
- **Portable**: Robust support for Linux, macOS (Apple Silicon), and Cloud Shell.
- **AI-Optimized**: Configurations and tools designed for seamless AI agent collaboration.

## Key Tools

- **Shell**: `fish` with `starship` prompt.
- **Editor**: `neovim` (modular and performant).
- **Multiplexer**: `zellij` (modern terminal workspace).
- **File Manager**: `yazi` (terminal file manager with previews).
- **Git and Docker**: `lazygit` and `gh`, `lazydocker` for containers.
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

## Docker

The Docker image bootstraps `chezmoi`, `mise`, and `fish`, but it does not
install the full home toolchain during the image build.

After starting the container, finalize it with:

```bash
mise -C "$HOME/.local/share/chezmoi" run tools
```
