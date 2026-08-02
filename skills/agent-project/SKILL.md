---
name: agent-project
description: Bootstrap shared project instructions and skills for Antigravity, Codex, OpenCode, Claude, and Copilot, while keeping each tool's MCP, settings, and custom-agent files in native locations.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/agent-project
  created: 2026-06-23
  updated: 2026-08-02
---

# Set Up Agents on a Project

Create the portable instruction and skill layer once, then add only the tool-specific configuration the project needs.

## Workflow

1. **Create the Shared Layer**:
   ```bash
   mkdir -p .agents/skills .agents/prompts
   touch AGENTS.md
   ```
   Start `AGENTS.md` from the minimal [project template](templates/AGENTS.md), then put project-wide rules there, store reusable project skills in `.agents/skills/<name>/SKILL.md`, and keep prompt files in `.agents/prompts/` (globally gitignored by default).
1. **Add Tool-Specific Files Only When Needed**:
   - **Antigravity**: use `.agents/settings.json` for workspace settings and `.agents/mcp_config.json` for MCP. It respects `.gitignore`; do not invent an `.antigravityignore`.
   - **Codex**: use `.codex/config.toml` for trusted project overrides and MCP. Codex reads `AGENTS.md` and `.agents/skills` natively.
   - **OpenCode**: use `opencode.json` for settings and MCP. OpenCode reads `AGENTS.md` and `.agents/skills` natively.
   - **Claude**: create `CLAUDE.md` containing `@AGENTS.md`, link `.claude/skills` to `../.agents/skills`, and use `.mcp.json` for project MCP servers.
   - **Copilot**: Copilot CLI reads `AGENTS.md` and `.agents/skills` natively. Use `.github/copilot-instructions.md` only for additional repository-wide Copilot instructions and `.github/mcp.json` or `.mcp.json` for MCP.
1. **Inspect Before Creating Links**: Never replace an existing `CLAUDE.md` or `.claude/skills` path without reviewing and preserving its content.
1. **Keep Secrets Out**: Add local credentials, generated agent state, and secret-bearing overrides to `.gitignore`. Commit only portable configuration.
1. **Validate Each Installed CLI**: Start each supported CLI from the repository root and confirm that instructions, skills, and explicitly configured MCP servers load.

## Shared Layout

```text
<repository-root>/
├── AGENTS.md
└── .agents/
    ├── prompts/
    └── skills/
        └── <name>/
            └── SKILL.md
```

Optional tool-specific files:

```text
<repository-root>/
├── .agents/
│   ├── mcp_config.json
│   └── settings.json
├── .claude/
│   ├── agents/
│   └── skills -> ../.agents/skills
├── .codex/
│   └── config.toml
├── .github/
│   ├── agents/
│   ├── copilot-instructions.md
│   └── mcp.json
├── .opencode/
│   └── agents/
├── .mcp.json
├── CLAUDE.md
└── opencode.json
```

## Custom Agents

Custom-agent definitions are not portable across all five CLIs. Keep them in each tool's native location rather than treating `.agents/agents` as a universal format.

| Tool        | Project location                          |
| ----------- | ----------------------------------------- |
| Antigravity | `.agents/agents/<name>/agent.md`          |
| Codex       | `[agents.<name>]` in `.codex/config.toml` |
| OpenCode    | `.opencode/agents/<name>.md`              |
| Claude      | `.claude/agents/<name>.md`                |
| Copilot     | `.github/agents/<name>.agent.md`          |

Give parallel agents independent, bounded tasks and non-overlapping file ownership. The parent agent integrates the work and runs project-wide validation.

## Gotchas

1. **Instruction Consolidation**: Keep shared rules in `AGENTS.md`; use `CLAUDE.md` only as Claude's `@AGENTS.md` bridge and avoid duplicating the rule body.
1. **Scope Priority**: Project configuration overrides global defaults; add the smallest override needed.
1. **Strict Formats**: Keep JSON valid without comments unless the documented file explicitly supports JSONC, and validate TOML before launching an agent.
1. **Untrusted Configuration**: Review repository hooks, MCP servers, skills, plugins, and custom-agent definitions before enabling them.
