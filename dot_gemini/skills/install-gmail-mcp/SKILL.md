---
name: install-gmail-mcp
description: Install the Gmail MCP server in the current project's .gemini/settings.json so Gemini can call Gmail tools without going through the gmail subagent.
---

# Install Gmail MCP

Drops the Gmail MCP server into `.gemini/settings.json` for the current project. Use this when inbox triage, search, or drafting happens in nearly every session of the project — otherwise prefer the `gmail` subagent (`~/.gemini/agents/gmail.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project triages, searches, summarizes, or drafts Gmail messages.
- The user wants Gmail tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"gmail"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "gmail": {
      "httpUrl": "https://gmailmcp.googleapis.com/mcp/v1",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc gmail
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `gmail` subagent (`~/.gemini/agents/gmail.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Gmail MCP server](https://developers.google.com/workspace/gmail/api/guides/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Gmail API](https://developers.google.com/gmail/api)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
