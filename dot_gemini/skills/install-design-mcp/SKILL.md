---
name: install-design-mcp
description: Install Google Design Center MCP into .gemini/settings.json. Use when Design Center work is central to the project.
---

# Install Design Center MCP

Drops the Google Design Center MCP server into `.gemini/settings.json` for the current project. Use this when design-token or Material 3 component work happens in nearly every session of the project — otherwise prefer the `design-mcp` subagent (`~/.gemini/agents/design-mcp.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo consumes Material design tokens, generates UI scaffolds, or reconciles design diffs against a token set.
- The user wants design tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"design-mcp"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "design-mcp": {
      "httpUrl": "https://design.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc design-mcp
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `design-mcp` subagent (`~/.gemini/agents/design-mcp.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Design MCP reference](https://developers.google.com/design-mcp/reference/mcp)
- [Material Design 3](https://m3.material.io)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
