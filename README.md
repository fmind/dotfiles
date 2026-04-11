# Dotfiles

Modern, high-performance dotfiles optimized for AI-driven development.

Managed via [Chezmoi](https://www.chezmoi.io/) and [Mise](https://mise.jdx.dev/).

## Key Tools

- **Shell**: `fish` + `starship`
- **Editor**: `neovim`
- **Multiplexer**: `zellij`
- **File Manager**: `yazi`
- **Git & Docker**: `lazygit`, `gh`, `lazydocker`
- **Utilities**: `bat`, `bottom`, `ripgrep`
- **Runtime**: `mise`

## Prerequisites

Ensure these system packages are installed on your host system:

```bash
sudo apt install -y git curl sudo libatomic1 build-essential
```

## Installation

```bash
# 1. Bootstrap config
curl -fsSL https://raw.githubusercontent.com/fmind/dotfiles/main/install.sh | bash

# 2. Install toolchain
~/.local/bin/mise -C ~/.local/share/chezmoi run tools

# 3. Install vim plugins
~/.local/bin/mise -C ~/.local/share/chezmoi run vim
```

## Tasks

Manage your environment with `mise run <task>`:

```bash
apply      # Apply configurations with chezmoi
check      # Preview changes before applying
docker     # Build and run the container
hooks      # Install repository pre-commit hooks
init       # Initialize chezmoi config from template
lock       # Refresh repository and home mise lockfiles
tools      # Install tools from ~/.config/mise/config.toml
trust      # Trust this repository and home mise configurations
upgrade    # Upgrade mise tools to latest versions and lock packages
vim        # Install Neovim plugins headlessly with LazyVim
```

## Docker

The Docker image bootstraps the core setup but omits the full home toolchain.

Finalize the installation inside the container with:

```bash
mise -C ~/.local/share/chezmoi run tools
mise -C ~/.local/share/chezmoi run vim
```
