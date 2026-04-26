# Project Agent Context

## Objectives

- **No-Sudo**: Avoid sudo whenever possible, stay in user-space.
- **Modern CLI**: Prioritize powerful, fast, and modern CLI tools.
- **AI-First**: Every tool must be runnable from the CLI by an agent.

## Principles

- **AI-Driven**: Add tools and configs that maximize safe agent autonomy.
- **Consistent**: Default to `catppuccin-mocha`, vim mode, ASCII icons.
- **Idempotent**: Ensure non-interactive and reproducible setups via lockfiles.
- **Portable**: Support Linux, macOS (Apple Silicon), and Cloud Shell configs.
- **No-Icons**: Avoid Nerd Font icons whenever possible to enhance compatibility.

## Collaboration

- **Active Dialogue**: Challenge the user if requests are ambiguous or underspecified.
- **Commit Strategy**: Do not commit changes unless specifically requested by the user.
- **Concise Rules**: Keep all `AGENTS.md` rules under 88 characters for readability.
- **Context First**: Review existing configs before adding new tools or settings.
- **Verify Syntax**: Validate tool usage against the latest online documentation.
