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
      "oauth": {
        "enabled": true,
        "clientId": "$GOOGLE_OAUTH_CLIENT_ID",
        "clientSecret": "$GOOGLE_OAUTH_CLIENT_SECRET",
        "scopes": [
          "https://www.googleapis.com/auth/gmail.readonly",
          "https://www.googleapis.com/auth/gmail.compose"
        ]
      },
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

Workspace MCPs require a per-user OAuth 2.0 flow — **not** Application Default Credentials. One-time setup:

1. In GCP Console, create an OAuth 2.0 client of type **Desktop app** and enable the Gmail API on the project.
2. Export `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` in your shell (a single OAuth client can be reused across all Workspace MCPs).
3. Run `/mcp auth gmail` in Gemini CLI to trigger the browser consent flow. Tokens are cached in `~/.gemini/mcp-oauth-tokens.json`.

Default scopes (`gmail.readonly`, `gmail.compose`) cover read + draft. Add `gmail.send` if you actually want to send (the `gmail` agent intentionally avoids sending).

## Companion Agent

The `gmail` subagent (`~/.gemini/agents/gmail.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Gmail MCP server](https://developers.google.com/workspace/gmail/api/guides/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Gmail API](https://developers.google.com/gmail/api)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
