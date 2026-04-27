---
name: install-cloud-trace-mcp
description: Install the GCP Cloud Trace MCP server in the current project's .gemini/settings.json so Gemini can call Cloud Trace tools without going through the cloud-trace subagent.
---

# Install Cloud Trace MCP

Drops the GCP Cloud Trace MCP server into `.gemini/settings.json` for the current project. Use this when distributed tracing and latency analysis happen in nearly every session of the project — otherwise prefer the `cloud-trace` subagent (`~/.gemini/agents/cloud-trace.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project investigates request latency, slow spans, or performance regressions across GCP microservices.
- The user wants trace tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"cloud-trace"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "cloud-trace": {
      "httpUrl": "https://cloudtrace.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc cloud-trace
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `cloud-trace` subagent (`~/.gemini/agents/cloud-trace.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Cloud Trace — MCP reference](https://docs.cloud.google.com/trace/docs/reference/mcp/mcp)
- [Cloud Trace](https://docs.cloud.google.com/trace/docs)
- [Trace concepts](https://docs.cloud.google.com/trace/docs/overview)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
