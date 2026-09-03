---
name: readme-agents
description: Sync AGENTS.md (for agents) and README.md (for humans) with the codebase's current tools, layout, and usage. Use when either drifts from the project's state.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/readme-agents
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Sync README and AGENTS Files

Refresh `AGENTS.md` (tooling, commands, conventions, and layout for agents) and `README.md` (purpose, installation, and usage for humans) so both reflect the repository's current state. [improve-docs](../improve-docs/SKILL.md) owns the wider documentation set.

## Workflow

1. **Research**: scan the root directory and key config files (`pyproject.toml`, `go.mod`, `mise.toml`, `lefthook.yml`) to identify tools and structure.
1. **Update AGENTS.md**: refresh the tool list and commands, and update the **Layout** section (a bullet list of top-level files and directories with a one-sentence purpose each).
1. **Update README.md**: make purpose, prerequisites, and usage examples current and remove stale instructions.
1. **Validate**: run `lychee README.md AGENTS.md` so refreshed links are proven live, not assumed.

## Gotchas

- **Separation of concerns**: user-facing installation instructions belong in `README.md`; technical conventions and rules belong in `AGENTS.md`.
- **Workflows become skills**: a multi-step section that outgrows `AGENTS.md` moves to a skill per [skillify](../skillify/SKILL.md), leaving a one-line pointer.

## Documentation

- [AGENTS.md Open Standard](https://agents.md) · [lychee](https://lychee.cli.rs)
- Companion skills: [agent-project](../agent-project/SKILL.md) (the initial `AGENTS.md` and layout), [improve-docs](../improve-docs/SKILL.md) (docs beyond these two files).
