---
name: install-vertex-ai-mcp
description: Install the GCP Vertex AI MCP server in the current project's .gemini/settings.json so Gemini can call Vertex AI tools without going through the vertex-ai subagent.
---

# Install Vertex AI MCP

Drops the GCP Vertex AI MCP server into `.gemini/settings.json` for the current project. Use this when Vertex AI training, deployment, or inference work happens in nearly every session of the project — otherwise prefer the `vertex-ai` subagent (`~/.gemini/agents/vertex-ai.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project trains or tunes models on Vertex pipelines, manages endpoints, or runs predictions with foundation models.
- The user wants Vertex AI tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"vertex-ai"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "vertex-ai": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/models",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

The `aiplatform.googleapis.com` MCP gateway exposes a separate endpoint per toolset — `/mcp/models`, `/mcp/predict`, `/mcp/generate`, `/mcp/notebook`, `/mcp/endpoints`, `/mcp/tuning`, `/mcp/retrieval`, `/mcp/evaluation`, `/mcp/prompts`. Pick the suffix matching the workflow (e.g. `/mcp/tuning` for fine-tuning, `/mcp/predict` for online inference). For data-residency, swap to a regional host such as `https://europe-west4-aiplatform.googleapis.com/mcp/models`. See [supported products](https://docs.cloud.google.com/mcp/supported-products) for the full toolset matrix.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc vertex-ai
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `vertex-ai` subagent (`~/.gemini/agents/vertex-ai.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Vertex AI — MCP reference](https://docs.cloud.google.com/vertex-ai/docs/reference/mcp)
- [Vertex AI](https://docs.cloud.google.com/vertex-ai/docs)
- [Vertex AI Gen AI SDK](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/sdks/overview)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
