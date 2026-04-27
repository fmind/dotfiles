---
name: install-cx-agent-studio-mcp
description: Install the Conversational Agents (CX) Agent Studio MCP server in the current project's .gemini/settings.json so Gemini can author, deploy, and inspect Dialogflow CX agents without going through a subagent.
---

# Install CX Agent Studio MCP

Drops the **Customer Engagement AI / Conversational Agents (Dialogflow CX) Agent Studio** MCP server into `.gemini/settings.json` for the current project. Use this when building or maintaining CX virtual agents (flows, pages, intents, routes, webhooks, eval sets) is a regular part of the work.

## When to Trigger

- The repo defines Dialogflow CX agents (export ZIPs, flow JSON, agent.yaml, webhook handlers).
- The user wants Agent Studio tools (flow-design, eval, deploy) in the main session.
- Verify first: `grep -q '"cx-agent-studio"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "cx-agent-studio": {
      "httpUrl": "https://ces.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc cx-agent-studio
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The Customer Experience Studio (CES) and Dialogflow CX APIs must be enabled on the target project:

```bash
gcloud services enable dialogflow.googleapis.com ces.googleapis.com --project=<PROJECT_ID>
```

The caller needs `roles/mcp.toolUser` on the project (plus the Dialogflow CX role required by each tool — typically `roles/dialogflow.admin` for design work).

## Documentation

- [CX Agent Studio MCP server](https://docs.cloud.google.com/customer-engagement-ai/conversational-agents/ps/mcp-server)
- [Conversational Agents (Dialogflow CX) overview](https://docs.cloud.google.com/dialogflow/cx/docs)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
