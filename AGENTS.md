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

## Mise Toolchain

- **Mise Config**: Declare globally installed tools in `dot_config/mise/config.toml`.
- **Mise Tasks**: Use `mise` tasks (e.g. `mr a`, `mr c`, `mr u`) for routine ops.
- **Bootstrap**: Run `mr b` for a full first-time setup on a clean machine.

## Gemini CLI

- **Agents** live in `dot_gemini/agents/<name>.md` (one per integration).
- **Skills** live in `dot_gemini/skills/<name>/SKILL.md` (workflows).
- **Commands** live in `dot_gemini/commands/<name>.toml` (slash-prompts).
- **Settings** live in `dot_gemini/settings.json` (theme, vim, MCP defaults).
- **Reflection**: Use `/reflect` to trigger the `reflect-on-history` skill.
- **Improvement**: Regularly reflect on history to evolve agents and skills.

## Secrets

- Never commit API keys. Place them either in:
  - `~/.private.fish` for secrets unique to this machine only.
  - `~/.config/fish/conf.d/secrets.fish` for API keys shared across machines.
- Manage `secrets.fish` with [chezmoi age encryption](https://www.chezmoi.io/user-guide/encryption/age/).
  - The source file stays encrypted in the repo.
  - `chezmoi apply` decrypts it to the target path.
