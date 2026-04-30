---
name: install-genkit-mcp
description: Install Genkit MCP into .gemini/settings.json. Use when Genkit work is central to the project.
---

# Install Genkit MCP

Drops the Genkit CLI MCP server into `.gemini/settings.json` for the current project. Use this when Genkit flow/prompt work happens in nearly every session of the project — otherwise prefer the `genkit` subagent (`~/.gemini/agents/genkit.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo imports `genkit`, `@genkit-ai/*`, or contains `*.prompt` files / Genkit flows.
- The user wants Genkit tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"genkit"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "genkit": {
      "command": "npx",
      "args": ["genkit", "mcp"],
      "includeTools": []
    }
  }
}
```

The MCP server ships inside the `genkit-cli` npm package and is invoked via the `genkit` binary (per the official docs: `npx genkit mcp`). If `genkit-cli` is installed globally (`npm install -g genkit-cli`), drop `npx` and use `"command": "genkit"`. Do not confuse with `@genkit-ai/mcp` or `genkitx-mcp`, which are MCP **client/host** plugins for Genkit apps.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc genkit
```

## Authentication

No authentication required.

## Companion Agent

The `genkit` subagent (`~/.gemini/agents/genkit.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Genkit](https://genkit.dev)
- [MCP server](https://genkit.dev/docs/mcp-server/)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
