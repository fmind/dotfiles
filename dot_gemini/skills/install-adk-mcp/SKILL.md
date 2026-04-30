---
name: install-adk-mcp
description: Install ADK docs MCP into .gemini/settings.json (Agent Development Kit documentation tools). Use when ADK work is central to the project.
---

# Install ADK MCP

Drops the Agent Development Kit docs MCP server into `.gemini/settings.json` for the current project. Use this when ADK work happens in nearly every session of the project — otherwise prefer the `adk` subagent (`~/.gemini/agents/adk.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo scaffolds Google ADK agents (Python, Java, Go, or TypeScript) or imports `google.adk` / `@google/adk`.
- The user wants ADK docs lookup available in the main session without invoking the subagent.
- Verify first: `grep -q '"adk-docs"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "adk-docs": {
      "command": "uvx",
      "args": [
        "--from",
        "mcpdoc",
        "mcpdoc",
        "--urls",
        "AgentDevelopmentKit:https://adk.dev/llms.txt",
        "--transport",
        "stdio"
      ],
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc adk-docs
```

## Authentication

No authentication required. `uvx` fetches `mcpdoc` on first run; ensure `uv` is installed (`pipx install uv` or via mise).

## Companion Agent

The `adk` subagent (`~/.gemini/agents/adk.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [ADK](https://adk.dev)
- [LLMs index (MCP source)](https://adk.dev/llms.txt)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
