# Agent Context

## Objectives

- **No-Sudo First**: Avoid sudo whenever possible, staying in user-space.
- **Out-of-the-Box**: Everything should work immediately after bootstrapping.
- **Minimalism**: Eliminate redundant scripts, config bloat, and custom logic.
- **Modern 2026 CLI**: Use powerful, modern, fast CLI for boosting productivity.

## Principles

- **AI-Driven**: Add tools and configurations that maximize safe agent autonomy.
- **Declarative Setup**: Prefer `chezmoi` and `mise` over imperative bash scripts.
- **High Performance**: Favor modern, high-performance solutions over legacy ones.
- **Lean Toolchain**: Avoid overlapping tools and unnecessary feature bloat.
- **Maintained**: Use only actively supported tools with clear documentation.
- **Portability**: Support Linux, macOS (Apple Silicon), and Cloud Shell in configs.
- **Sane Defaults**: Configure tools to work out-of-the-box with minimal config.
- **Robust Idempotency**: Ensure non-interactive, reproducible setups via lockfiles.
- **Catppuccin Mocha**: Use "catppuccin-mocha" as the default theme everywhere.
- **Vim-Centric**: Prefer tools with native and intuitive Vim-style keybindings.

## Collaboration

- **Continuous Learning**: Always check for new tools and assess their relevance.
- **Context First**: Review existing configurations before adding new tools/logic.
- **Verify Online**: Validate tool usage against the latest online documentation.
- **Active Dialogue**: Challenge the user if requests are ambiguous or underspecified.
- **Commit Strategy**: Do not commit changes unless specifically requested by the user.
- **Rule Length**: Keep all `AGENTS.md` rules under 88 characters for readability.
