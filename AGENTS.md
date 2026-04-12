# Agent Context

## Objectives

- **No-Sudo**: Avoid sudo whenever possible, stay in user-space.
- **Modern CLI**: Prioritize powerful, fast, and modern CLI tools.

## Principles

- **AI-Driven**: Add tools and configs that maximize safe agent autonomy.
- **Consistent**: Default to "catppuccin-mocha" theme (if built-in) and Vim.
- **Idempotent**: Ensure non-interactive and reproducible setups via lockfiles.
- **Portable**: Support Linux, macOS (Apple Silicon), and Cloud Shell configs.

## Collaboration

- **Active Dialogue**: Challenge the user if requests are ambiguous or underspecified.
- **Commit Strategy**: Do not commit changes unless specifically requested by the user.
- **Concise Rules**: Keep all `AGENTS.md` rules under 88 characters for readability.
- **Context First**: Review existing configs before adding new tools or settings.
- **Verify Syntax**: Validate tool usage against the latest online documentation.

## Mise Toolchain

- **Mise Config**: Declare globally installed tools in `dot_config/mise/config.toml`.
- **Mise Lock**: Run `mise lock` to record package versions after adding a tool.
- **Mise Tasks**: Add routine commands as aliases or tasks in `mise.toml`.
