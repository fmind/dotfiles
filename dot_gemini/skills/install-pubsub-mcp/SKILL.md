---
name: install-pubsub-mcp
description: Install the GCP Pub/Sub MCP server in the current project's .gemini/settings.json so Gemini can manage topics, subscriptions, and publish/pull messages without going through a subagent.
---

# Install Pub/Sub MCP

Drops the GCP Pub/Sub MCP server into `.gemini/settings.json` for the current project. Use this when eventing/messaging work (topics, subscriptions, dead-letter queues, message replay) happens in nearly every session.

## When to Trigger

- The repo configures Pub/Sub topics/subscriptions (Terraform, gcloud scripts, IaC) or imports `google-cloud-pubsub` / `@google-cloud/pubsub`.
- The user wants to publish, pull, or inspect messages from the agent.
- Verify first: `grep -q '"pubsub"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "pubsub": {
      "httpUrl": "https://pubsub.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc pubsub
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The Pub/Sub API must be enabled:

```bash
gcloud services enable pubsub.googleapis.com --project=<PROJECT_ID>
```

## Documentation

- [Pub/Sub — MCP reference](https://docs.cloud.google.com/pubsub/docs/reference/mcp)
- [Pub/Sub overview](https://docs.cloud.google.com/pubsub/docs)
- [`gcloud pubsub` CLI](https://docs.cloud.google.com/sdk/gcloud/reference/pubsub)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
