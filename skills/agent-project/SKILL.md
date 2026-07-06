---
name: agent-project
description: Bootstrap workspace agent configuration — AGENTS.md, the .agents/ layout, subagents, and skills — for Antigravity, OpenCode, and Claude. Use when initializing or onboarding a repository for agents.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/agent-project
  created: 2026-06-23
  updated: 2026-07-06
---

# Set Up Agents on a Project

Set up workspace agent configuration, subagents, and skills on a repository for Antigravity, OpenCode, and Claude.

## Workflow

1. **Bootstrap Folder Layout**:
   - **Antigravity**: Create the standard configuration folder and files:
     ```bash
     mkdir -p .agents/{agents,skills}
     echo '{}' > .agents/settings.json
     echo '{"mcpServers": {}}' > .agents/mcp_config.json
     touch .antigravityignore AGENTS.md
     ```
   - **OpenCode**: Create the workspace config file and skill folder:
     ```bash
     mkdir -p .agents/skills
     echo '{}' > opencode.json
     touch AGENTS.md
     ```
     _Note: For OpenCode CLI, you can alternatively run the `/init` terminal command to configure the repository automatically._
   - **Claude**: Create the workspace configuration, link it to the master rules, and symlink the skills directory:
     ```bash
     mkdir -p .agents/skills .claude
     ln -s ../.agents/skills .claude/skills
     echo '{"mcpServers": {}}' > .mcp.json
     echo '@AGENTS.md' > CLAUDE.md
     touch AGENTS.md
     ```
1. **Create AGENTS.md**: Add a master workspace rules file `AGENTS.md` at the repository root using the template at [AGENTS.md](templates/AGENTS.md). Claude is linked to this via `CLAUDE.md`.
1. **Define Exclusions**: Add build artifacts and secrets to `.antigravityignore` (Antigravity; same glob syntax as `.gitignore`). OpenCode and Claude honor `.gitignore` directly.

## Recommended Layout

### Antigravity

```text
<repository-root>/
├── AGENTS.md              # Master workspace instruction rules
└── .agents/               # Standard configuration folder
    ├── settings.json      # Antigravity settings overrides
    ├── mcp_config.json    # Antigravity MCP server definitions
    ├── agents/            # Custom subagents (*.md)
    └── skills/            # Workspace-scope skills (SKILL.md folders)
```

### OpenCode

```text
<repository-root>/
├── AGENTS.md              # Master workspace instruction rules
├── opencode.json          # OpenCode settings & MCP configuration
└── .agents/               # Standard configuration folder
    └── skills/            # Workspace-scope skills (SKILL.md folders)
```

### Claude

```text
<repository-root>/
├── AGENTS.md              # Master workspace instruction rules
├── CLAUDE.md              # Reference to AGENTS.md (contains "@AGENTS.md")
├── .mcp.json              # Claude MCP configuration
├── .claude/
│   └── skills/            # Symlink to ../.agents/skills
└── .agents/               # Standard configuration folder
    └── skills/            # Workspace-scope skills (SKILL.md folders)
```

## Gotchas

1. **Rule Consolidation**: All tools automatically parse project rules. Antigravity and OpenCode read `AGENTS.md` directly, while Claude is directed to it via `@AGENTS.md` inside `CLAUDE.md`.
1. **Scope Priority**: Workspace settings in `.agents/` (Antigravity), `.mcp.json` (Claude), or `opencode.json` (OpenCode) override global settings.
1. **Strict JSON**: Config files like `settings.json`, `mcp_config.json`, `.mcp.json`, and `opencode.json` must be valid JSON (no trailing commas or comments).
1. **Claude Skills Symlink**: Claude Code only discovers workspace skills in `.claude/skills/`. To load skills from the standard `.agents/skills/` directory, a symlink `.claude/skills` pointing to `../.agents/skills` must be created.
