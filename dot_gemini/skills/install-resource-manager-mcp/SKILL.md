---
name: install-resource-manager-mcp
description: Install the GCP Resource Manager MCP server in the current project's .gemini/settings.json so Gemini can call Resource Manager tools without going through the resource-manager subagent.
---

# Install Resource Manager MCP

Drops the GCP Resource Manager MCP server into `.gemini/settings.json` for the current project. Use this when project/folder/IAM management happens in nearly every session of the project — otherwise prefer the `resource-manager` subagent (`~/.gemini/agents/resource-manager.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project manages GCP project lifecycle, folder hierarchy, or IAM bindings across an organization.
- The user wants Resource Manager tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"resource-manager"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "resource-manager": {
      "httpUrl": "https://cloudresourcemanager.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc resource-manager
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `resource-manager` subagent (`~/.gemini/agents/resource-manager.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Resource Manager — MCP reference](https://docs.cloud.google.com/resource-manager/reference/mcp)
- [Resource Manager](https://docs.cloud.google.com/resource-manager/docs)
- [Resource hierarchy](https://docs.cloud.google.com/resource-manager/docs/cloud-platform-resource-hierarchy)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
