---
name: install-cloud-logging-mcp
description: Install the GCP Cloud Logging MCP server in the current project's .gemini/settings.json so Gemini can call Logging tools without going through the cloud-logging subagent.
---

# Install Cloud Logging MCP

Drops the GCP Cloud Logging MCP server into `.gemini/settings.json` for the current project. Use this when log queries happen in nearly every session of the project — otherwise prefer the `cloud-logging` subagent (`~/.gemini/agents/cloud-logging.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project troubleshoots GCP services, queries log entries, or manages log sinks/buckets.
- The user wants log tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"cloud-logging"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "cloud-logging": {
      "httpUrl": "https://logging.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc cloud-logging
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `cloud-logging` subagent (`~/.gemini/agents/cloud-logging.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Cloud Logging — MCP reference](https://docs.cloud.google.com/logging/docs/reference/v2_mcp/mcp)
- [Cloud Logging](https://docs.cloud.google.com/logging/docs)
- [Logging Query Language](https://docs.cloud.google.com/logging/docs/view/logging-query-language)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
