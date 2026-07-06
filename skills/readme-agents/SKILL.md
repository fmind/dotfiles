---
name: readme-agents
description: Sync AGENTS.md (for agents) and README.md (for humans) with the codebase's current tools, layout, and usage. Use when either drifts from the project's actual state.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/readme-agents
  created: 2026-06-23
  updated: 2026-07-06
---

# Sync README and AGENTS Files

Refresh `AGENTS.md` and `README.md` to accurately reflect the repository's state.

## Scope

- **`AGENTS.md` (For AI)**: Tooling, commands, technical conventions, and repository layout.
- **`README.md` (For Humans)**: Project purpose, installation, setup, and usage.

## Workflow

1. Research: Scan the root directory and key config files (e.g., `pyproject.toml`, `go.mod`, `mise.toml`, `lefthook.yml`) to identify tools and structure.
1. Update AGENTS.md:
   - Refresh tool list and commands.
   - Update the **Layout** section (a bullet list of top-level files/dirs with a one-sentence purpose each).
1. Update README.md:
   - Ensure purpose, prerequisites, and usage examples are current.
   - Remove stale instructions.

## Gotchas

- **Separation of Concerns**: Do not mix user-facing installation instructions (for `README.md`) with technical conventions and rules (for `AGENTS.md`).
- **No-Hard-Wrap**: Keep each paragraph on a single line in both files.

## Documentation

- [AGENTS.md Open Standard](https://agents.md) — the README-for-humans vs AGENTS.md-for-agents split.
