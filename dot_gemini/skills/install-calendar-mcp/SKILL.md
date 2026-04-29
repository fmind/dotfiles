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
      "oauth": {
        "enabled": true,
        "clientId": "$GOOGLE_OAUTH_CLIENT_ID",
        "clientSecret": "$GOOGLE_OAUTH_CLIENT_SECRET",
        "scopes": [
          "https://www.googleapis.com/auth/calendar.calendarlist.readonly",
          "https://www.googleapis.com/auth/calendar.events.freebusy",
          "https://www.googleapis.com/auth/calendar.events.readonly"
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
/mcp desc calendar
```

## Authentication

Workspace MCPs require a per-user OAuth 2.0 flow — **not** Application Default Credentials. One-time setup:

1. In GCP Console, create an OAuth 2.0 client of type **Desktop app** and enable the Calendar API on the project.
2. Export `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` in your shell (a single OAuth client can be reused across all Workspace MCPs).
3. Run `/mcp auth calendar` in Gemini CLI to trigger the browser consent flow. Tokens are cached in `~/.gemini/mcp-oauth-tokens.json`.

Default scopes are read-only (calendar list, events, free/busy). To create, update, or cancel events, swap `calendar.events.readonly` for the wider `calendar.events` scope.

## Companion Agent

The `calendar` subagent (`~/.gemini/agents/calendar.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Calendar MCP server](https://developers.google.com/workspace/calendar/api/guides/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Google Calendar API](https://developers.google.com/calendar/api)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
