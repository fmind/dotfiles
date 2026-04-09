# Agent Context

## Objectives

- **Modern 2026 CLI**: Modern development on CLI as of 2026.
- **Robust & Simple**: Fast, efficient, and robust deployment with minimal complexity.
- **No-Sudo First**: Avoid sudo whenever possible, staying in user-space.
- **Out-of-the-Box**: Everything should work immediately after bootstrapping.

## Core Principles

- **AI-Driven**: Add tools and configurations that maximize safe agent autonomy.
- **High Performance**: Favor modern, high-performance solutions over legacy ones.
- **Lean Toolchain**: Avoid overlapping tools and unnecessary feature bloat.
- **Maintained**: Use only actively supported tools with clear documentation.
- **Portability**: Support Linux, macOS (Apple Silicon), and Cloud Shell in all configs.
- **Sane Defaults**: Configure tools to work out-of-the-box with minimal config.
- **Idempotency**: Ensure installations are non-interactive and use mise lockfiles.
- **Catppuccin Mocha**: Use "catppuccin-mocha" as the default theme everywhere.
- **Vim-Centric**: Prefer tools with native and intuitive Vim-style keybindings.

## AI Collaboration

- **Continuous Learning**: Always check for new tools and assess their relevance.
- **Verify Online**: Validate tool usage against the latest online documentation.
- **Active Dialogue**: Challenge the user if requests are ambiguous or underspecified.
- **Commit Strategy**: Do not commit changes unless specifically requested by the user.
- **Rule Length**: Keep all `AGENTS.md` rules under 88 characters for readability.

## Tooling & Ecosystem

- **Fish Shell**: Focus on a `fish`-first shell experience for all CLI interactions.
- **Chezmoi**: Manage all dotfiles through `chezmoi` templates and state.
- **Mise Managed**: Use `mise` for all tool versioning and project task management.
- **Split Bootstrap**: Keep `install.sh` minimal; use `mise run tools` for tools.
- **Explicit Tool Deps**: Keep helpers like `pipx` in `dot_config/mise/config.toml`.
- **Docker Ready**: Maintain a lean, automated Dockerfile for repeatable builds.
- **Modern CLI**: Prioritize tools like `eza`, `bat`, `yazi`, `zellij`, and `btop`.
- **Neovim**: Maintain a modular and performant `nvim` setup for coding.
- **Starship**: Use `starship` for a consistent and informative shell prompt.
