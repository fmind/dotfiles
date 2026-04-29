---
name: install-drive-mcp
description: Install the Google Drive MCP server in the current project's .gemini/settings.json so Gemini can call Drive tools without going through the drive subagent.
---

# Install Drive MCP

Drops the Google Drive MCP server into `.gemini/settings.json` for the current project. Use this when Drive search, reads, or sharing happen in nearly every session of the project — otherwise prefer the `drive` subagent (`~/.gemini/agents/drive.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project searches, summarizes, organizes, or shares Google Drive files (Docs, Sheets, Slides, PDFs).
- The user wants Drive tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"drive"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "drive": {
      "httpUrl": "https://drivemcp.googleapis.com/mcp/v1",
      "oauth": {
        "enabled": true,
        "clientId": "$GOOGLE_OAUTH_CLIENT_ID",
        "clientSecret": "$GOOGLE_OAUTH_CLIENT_SECRET",
        "scopes": [
          "https://www.googleapis.com/auth/drive.readonly",
          "https://www.googleapis.com/auth/drive.file"
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
/mcp desc drive
```

## Authentication

Workspace MCPs require a per-user OAuth 2.0 flow — **not** Application Default Credentials. One-time setup:

1. In GCP Console, create an OAuth 2.0 client of type **Desktop app** and enable the Drive API on the project.
2. Export `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` in your shell (a single OAuth client can be reused across all Workspace MCPs).
3. Run `/mcp auth drive` in Gemini CLI to trigger the browser consent flow. Tokens are cached in `~/.gemini/mcp-oauth-tokens.json`.

Default scopes (`drive.readonly`, `drive.file`) cover reads and app-created/picked files only. To manage or share arbitrary files, swap to the wider `drive` scope (verification-gated).

## Companion Agent

The `drive` subagent (`~/.gemini/agents/drive.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Drive MCP server](https://developers.google.com/workspace/drive/api/guides/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Google Drive API](https://developers.google.com/drive)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
