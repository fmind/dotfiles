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

### Quick Install

Boostrap your environment with a single command:

```bash
curl -sL https://raw.githubusercontent.com/fmind/dotfiles/main/install.sh | bash
```

### Manual Installation

If you prefer a manual setup:

```bash
# 1. Install Chezmoi and Mise
curl -sS https://get.chezmoi.io | sh -s -- -b ~/.local/bin
curl https://mise.run | sh

# 2. Initialize and apply dotfiles
~/.local/bin/chezmoi init --apply fmind
```

## Setup Tasks

Manage your environment using built-in `mise` tasks:

```bash
mise run install  # Bootstrap dependencies
mise run apply    # Apply latest configurations
mise run update   # Update from remote and apply
mise run check    # Dry-run changes
mise run docker   # Build and run development container
```

## AI Support

This repository includes an `AGENTS.md` file that provides foundational mandates and technical context for AI agents (like Gemini or Copilot) to help them better understand and assist within this environment.
