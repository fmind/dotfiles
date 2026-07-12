# AGENTS.md (Project)

This file provides the context and rules for AI agents.

## Project Identity

- **Name**: <Project Name>
- **Description**: <1-2 sentences on what this project does.>

## Tech Stack & Tools

- **Language**: Python (uv, see pyproject.toml) / Go (go.mod)
- **Task Runner**: `mise` (see `mise.toml`)
- **Formatting**: `dprint` (config/markup) + `ruff` / `gofumpt`
- **Linting**: `ruff` / `golangci-lint`
- **Git Hooks**: `lefthook` (see `lefthook.yml`)
- **Testing**: `pytest` / `gotestsum`

## Rules & Standards

- **Elegant & Refined**: Write minimalist, correct, and self-documenting code. Avoid complex solutions.
- **Fail Fast**: Use standard error handling and assertions; do not swallow exceptions.
- **Git Commits**: Strictly use Conventional Commits (e.g., `feat:`, `fix:`, `chore:`).

## Workspace Layout

- `AGENTS.md` — Shared project instructions for Antigravity, Codex, OpenCode, and Copilot; Claude imports it from `CLAUDE.md`.
- `.agents/skills/` — Portable project skills discovered by Antigravity, Codex, OpenCode, and Copilot; Claude links `.claude/skills` here.
