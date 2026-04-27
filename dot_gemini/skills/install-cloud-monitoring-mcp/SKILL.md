---
name: install-cloud-monitoring-mcp
description: Install the GCP Cloud Monitoring MCP server in the current project's .gemini/settings.json so Gemini can call Monitoring tools without going through the cloud-monitoring subagent.
---

# Install Cloud Monitoring MCP

Drops the GCP Cloud Monitoring MCP server into `.gemini/settings.json` for the current project. Use this when metric queries and alert work happen in nearly every session of the project — otherwise prefer the `cloud-monitoring` subagent (`~/.gemini/agents/cloud-monitoring.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project queries metrics with MQL/PromQL, manages alert policies, or maintains SLOs.
- The user wants metric tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"cloud-monitoring"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "cloud-monitoring": {
      "httpUrl": "https://monitoring.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc cloud-monitoring
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `cloud-monitoring` subagent (`~/.gemini/agents/cloud-monitoring.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Cloud Monitoring — MCP reference](https://docs.cloud.google.com/monitoring/api/ref_v3_mcp/mcp)
- [Cloud Monitoring](https://docs.cloud.google.com/monitoring/docs)
- [Monitoring Query Language (MQL)](https://docs.cloud.google.com/monitoring/mql)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
