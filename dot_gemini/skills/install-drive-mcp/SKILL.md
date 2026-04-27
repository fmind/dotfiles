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
      "authProviderType": "google_credentials",
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

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `drive` subagent (`~/.gemini/agents/drive.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Drive MCP server](https://developers.google.com/workspace/drive/api/guides/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Google Drive API](https://developers.google.com/drive)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
