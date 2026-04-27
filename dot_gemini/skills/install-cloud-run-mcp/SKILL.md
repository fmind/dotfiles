---
name: install-cloud-run-mcp
description: Install the GCP Cloud Run MCP server in the current project's .gemini/settings.json so Gemini can call Cloud Run tools without going through the cloud-run subagent.
---

# Install Cloud Run MCP

Drops the GCP Cloud Run MCP server into `.gemini/settings.json` for the current project. Use this when Cloud Run deploys, traffic splits, or revision management happen in nearly every session of the project — otherwise prefer the `cloud-run` subagent (`~/.gemini/agents/cloud-run.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo deploys services or jobs to Cloud Run (e.g. `service.yaml`, `Dockerfile` targeting Cloud Run, `cloudbuild.yaml`).
- The user wants Cloud Run tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"cloud-run"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "cloud-run": {
      "httpUrl": "https://run.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc cloud-run
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `cloud-run` subagent (`~/.gemini/agents/cloud-run.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Cloud Run — MCP reference](https://docs.cloud.google.com/run/docs/reference/mcp)
- [Cloud Run](https://docs.cloud.google.com/run/docs)
- [Deploy from source](https://docs.cloud.google.com/run/docs/deploying-source-code)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
