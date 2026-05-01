---
name: install-gemini-enterprise-mcp
description: Install Gemini Enterprise Agent Platform MCP into .gemini/settings.json. Use when the Agent Platform is central to the project.
---

# Install Gemini Enterprise MCP

Drops the Gemini Enterprise Agent Platform MCP server into `.gemini/settings.json` for the current project. Use this when enterprise agent design/deploy work happens in nearly every session of the project — otherwise prefer the `gemini-enterprise` subagent (`~/.gemini/agents/gemini-enterprise.md`), which loads the MCP only when invoked and keeps the parent context clean.

> **Note:** Vertex AI was rebranded to **Gemini Enterprise Agent Platform**. Both products share the `aiplatform.googleapis.com` MCP host but use different toolset paths (`/mcp/<toolset>`). This skill targets the Agent Platform tools (retrieval, evaluation, prompts); for Vertex AI Gen AI / training / endpoints, see the **Vertex AI** section of `install-gcp-mcp`.

## When to Trigger

- The project designs, deploys, or governs agents on the Gemini Enterprise Agent Platform.
- The user wants Enterprise toolset access in the main session without invoking the subagent.
- Verify first: `grep -q '"gemini-enterprise-retrieval"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

The `aiplatform.googleapis.com` MCP gateway has **no bare `/mcp` endpoint** — each toolset is a separate MCP server. Register one entry per Agent Platform toolset; the Vertex AI generation/training toolsets (`models`, `predict`, `generate`, `notebook`, `endpoints`, `tuning`) live in `install-gcp-mcp` instead.

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "gemini-enterprise-retrieval": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/retrieval",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "gemini-enterprise-evaluation": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/evaluation",
      "authProviderType": "google_credentials",
      "includeTools": []
    },
    "gemini-enterprise-prompts": {
      "httpUrl": "https://aiplatform.googleapis.com/mcp/prompts",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

Drop the entries you don't need. For data-residency, swap to a regional host such as `https://europe-west4-aiplatform.googleapis.com/mcp/retrieval`. See [supported products](https://docs.cloud.google.com/mcp/supported-products) for the full toolset matrix.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc gemini-enterprise-retrieval
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `gemini-enterprise` subagent (`~/.gemini/agents/gemini-enterprise.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Use the Agent Platform remote MCP server](https://docs.cloud.google.com/gemini-enterprise-agent-platform/reference/use-agent-platform-mcp)
- [Gemini Enterprise Agent Platform — MCP reference](https://docs.cloud.google.com/gemini-enterprise-agent-platform/reference/mcp)
- [Gemini Enterprise overview](https://docs.cloud.google.com/gemini/enterprise/docs)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
