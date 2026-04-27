---
name: install-agent-search-mcp
description: Install the GCP Agent Search MCP server in the current project's .gemini/settings.json so Gemini can call Discovery Engine tools without going through the agent-search subagent.
---

# Install Agent Search MCP

Drops the GCP Agent Search (Discovery Engine) MCP server into `.gemini/settings.json` for the current project. Use this when grounded enterprise search happens in nearly every session of the project — otherwise prefer the `agent-search` subagent (`~/.gemini/agents/agent-search.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo queries Discovery Engine data stores or builds grounded RAG over Cloud Storage / BigQuery corpora.
- The user wants enterprise search tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"agent-search"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "agent-search": {
      "httpUrl": "https://discoveryengine.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc agent-search
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `agent-search` subagent (`~/.gemini/agents/agent-search.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Agent Search MCP reference (Discovery Engine)](https://docs.cloud.google.com/generative-ai-app-builder/docs/reference/mcp)
- [Agent Search MCP — search resource](https://docs.cloud.google.com/generative-ai-app-builder/docs/reference/mcp/search)
- [Agent Search overview (formerly Vertex AI Search)](https://docs.cloud.google.com/generative-ai-app-builder/docs)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
