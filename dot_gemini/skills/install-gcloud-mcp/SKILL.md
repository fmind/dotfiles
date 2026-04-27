---
name: install-gcloud-mcp
description: Install the gcloud MCP server in the current project's .gemini/settings.json so Gemini can drive the local gcloud CLI as a typed tool.
---

# Install gcloud MCP

Drops the [`googleapis/gcloud-mcp`](https://github.com/googleapis/gcloud-mcp) server into `.gemini/settings.json` for the current project. The server wraps the locally-installed `gcloud` CLI and exposes its commands as MCP tools — useful when you want Gemini to operate on Google Cloud through the same paths you would type at the shell.

## When to Trigger

- The user routinely runs `gcloud` (deploys, IAM, services, secrets, projects).
- The user wants to delegate `gcloud` commands to the agent and have structured output back.
- Verify first: `grep -q '"gcloud"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "gcloud": {
      "command": "npx",
      "args": ["-y", "@google-cloud/gcloud-mcp"],
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc gcloud
```

## Authentication

Inherits the local `gcloud` configuration — there is no separate auth step for the MCP server itself. The agent operates as the currently-authenticated user/principal:

```bash
gcloud auth login
gcloud auth application-default login
gcloud config set project <PROJECT_ID>
gcloud config list
```

## Notes

- The server invokes the local `gcloud` binary; commands run with whatever permissions your account has.
- For destructive ops (delete/disable/IAM changes), prefer scoping `includeTools` to safe read-only tools, or use a dedicated service-account-bound shell.

## Documentation

- [`googleapis/gcloud-mcp`](https://github.com/googleapis/gcloud-mcp)
- [gcloud CLI reference](https://docs.cloud.google.com/sdk/gcloud/reference)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
