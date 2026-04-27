---
name: install-vertex-ai-search-mcp
description: Install the GCP Vertex AI Search MCP server in the current project's .gemini/settings.json so Gemini can call Discovery Engine tools without going through the vertex-ai-search subagent.
---

# Install Vertex AI Search MCP

Drops the GCP Vertex AI Search (Discovery Engine) MCP server into `.gemini/settings.json` for the current project. Use this when enterprise search and grounded answer generation happen in nearly every session of the project — otherwise prefer the `vertex-ai-search` subagent (`~/.gemini/agents/vertex-ai-search.md`), which loads the MCP only when invoked and keeps the parent context clean.

> **Note:** Vertex AI Search was rebranded to **Agent Search** at Cloud Next 2026 (April 2026). The endpoint is unchanged; doc URLs are gradually migrating to the new name. This skill overlaps with `install-agent-search-mcp` — they target the same hostname; choose one or scope `includeTools` to disambiguate.

## When to Trigger

- The project configures data stores, ingests corpora, or queries Vertex AI Search for grounded answers.
- The user wants Vertex AI Search tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"vertex-ai-search"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "vertex-ai-search": {
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
/mcp desc vertex-ai-search
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `vertex-ai-search` subagent (`~/.gemini/agents/vertex-ai-search.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Discovery Engine — MCP reference](https://docs.cloud.google.com/generative-ai-app-builder/docs/reference/mcp)
- [Agent Search MCP — search resource](https://docs.cloud.google.com/generative-ai-app-builder/docs/reference/mcp/search)
- [Vertex AI Search / Agent Search overview](https://docs.cloud.google.com/generative-ai-app-builder/docs)
- [Discovery Engine API](https://docs.cloud.google.com/generative-ai-app-builder/docs/reference/rest)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
