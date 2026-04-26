# Claude Code Subagents

Drop one Markdown file per subagent in this directory. Each file follows the
schema documented in `dot_claude/skills/create-claude-subagent/SKILL.md`:

```markdown
---
name: <agent-name>
description: <when the orchestrator should delegate to this agent>
tools: Read, Grep, Glob, Bash
model: sonnet
---

# <Agent Name>

You are the specialized <agent-name> agent. ...
```

After deploying with `mr a`, agents become invokable via the `Task` tool or
`/agents` in any session.

A bulk port of the 31 Gemini integration subagents under `dot_gemini/agents/`
is **deferred** — they rely on Gemini's `mcp_servers` frontmatter, and the
Claude Code MCP model lives in `.mcp.json` instead. Migrating them properly
is a redesign, not a 1-for-1 translation.
