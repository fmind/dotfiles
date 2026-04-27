---
name: install-calendar-mcp
description: Install the Google Calendar MCP server in the current project's .gemini/settings.json so Gemini can call Calendar tools without going through the calendar subagent.
---

# Install Calendar MCP

Drops the Google Calendar MCP server into `.gemini/settings.json` for the current project. Use this when calendar reads/writes happen in nearly every session of the project — otherwise prefer the `calendar` subagent (`~/.gemini/agents/calendar.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project orchestrates scheduling, availability checks, or event automation against Google Calendar.
- The user wants calendar tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"calendar"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "calendar": {
      "httpUrl": "https://calendarmcp.googleapis.com/mcp/v1",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc calendar
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `calendar` subagent (`~/.gemini/agents/calendar.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Calendar MCP server](https://developers.google.com/workspace/calendar/api/guides/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Google Calendar API](https://developers.google.com/calendar/api)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
