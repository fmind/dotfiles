---
name: install-developer-knowledge-mcp
description: Install the GCP Developer Knowledge MCP server in the current project's .gemini/settings.json so Gemini can call doc retrieval tools without going through the developer-knowledge subagent.
---

# Install Developer Knowledge MCP

Drops the GCP Developer Knowledge MCP server into `.gemini/settings.json` for the current project. Use this when grounded GCP doc lookups happen in nearly every session of the project — otherwise prefer the `developer-knowledge` subagent (`~/.gemini/agents/developer-knowledge.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project regularly grounds answers in current Google Cloud documentation, samples, or reference architectures.
- The user wants doc-retrieval tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"developer-knowledge"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "developer-knowledge": {
      "httpUrl": "https://developerknowledge.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc developer-knowledge
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `developer-knowledge` subagent (`~/.gemini/agents/developer-knowledge.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Developer Knowledge MCP](https://developers.google.com/knowledge/mcp)
- [Google Cloud documentation](https://docs.cloud.google.com/docs)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
