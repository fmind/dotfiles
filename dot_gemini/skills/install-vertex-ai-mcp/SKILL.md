---
name: install-vertex-ai-mcp
description: Install the GCP Vertex AI MCP server in the current project's .gemini/settings.json so Gemini can call Vertex AI tools without going through the vertex-ai subagent.
---

# Install Vertex AI MCP

Drops the GCP Vertex AI MCP server into `.gemini/settings.json` for the current project. Use this when Vertex AI training, deployment, or inference work happens in nearly every session of the project — otherwise prefer the `vertex-ai` subagent (`~/.gemini/agents/vertex-ai.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project trains or tunes models on Vertex pipelines, manages endpoints, or runs predictions with foundation models.
- The user wants Vertex AI tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"vertex-ai-models"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

The `aiplatform.googleapis.com` MCP gateway has **no bare `/mcp` endpoint** — each toolset is a separate MCP server. Register one entry per toolset you need; agent-platform toolsets (`/mcp/retrieval`, `/mcp/evaluation`, `/mcp/prompts`) belong to `install-gemini-enterprise-mcp` instead.

```json
{
  "mcpServers": {
    "vertex-ai-models": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/models",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "vertex-ai-predict": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/predict",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "vertex-ai-generate": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/generate",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "vertex-ai-notebook": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/notebook",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "vertex-ai-endpoints": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/endpoints",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "vertex-ai-tuning": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/tuning",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

Pick the toolset matching the workflow — `models` (registry), `predict` (online inference), `generate` (Gemini/foundation models), `notebook` (managed notebooks), `endpoints` (deploy/undeploy), `tuning` (custom-model training). Drop the entries you don't need. For data-residency, swap to a regional host such as `https://europe-west4-aiplatform.googleapis.com/mcp/models`. See [supported products](https://docs.cloud.google.com/mcp/supported-products) for the full toolset matrix.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc vertex-ai-models
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
