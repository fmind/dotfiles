---
name: install-agent-registry-mcp
description: Install the GCP Agent Registry MCP server in the current project's .gemini/settings.json so Gemini can call Agent Registry tools without going through the agent-registry subagent.
---

# Install Agent Registry MCP

Drops the GCP Agent Registry MCP server into `.gemini/settings.json` for the current project. Use this when Agent Registry work happens in nearly every session of the project — otherwise prefer the `agent-registry` subagent (`~/.gemini/agents/agent-registry.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo publishes or consumes agent definitions in the GCP Agent Registry.
- The user wants Agent Registry tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"agent-registry"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "agent-registry": {
      "httpUrl": "https://agentregistry.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc agent-registry
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `agent-registry` subagent (`~/.gemini/agents/agent-registry.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Use the Agent Registry MCP server](https://docs.cloud.google.com/agent-registry/use-agentregistry-mcp)
- [Agent Registry — MCP reference](https://docs.cloud.google.com/agent-registry/reference/mcp)
- [Agent Registry overview](https://docs.cloud.google.com/agent-registry/overview)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
