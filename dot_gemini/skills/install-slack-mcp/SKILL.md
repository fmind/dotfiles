---
name: install-slack-mcp
description: Install the reference Slack MCP server in the current project's .gemini/settings.json so Gemini can read channels, post messages, and manage threads without going through a subagent.
---

# Install Slack MCP

Drops the reference [`@modelcontextprotocol/server-slack`](https://www.npmjs.com/package/@modelcontextprotocol/server-slack) into `.gemini/settings.json` for the current project. Use this when the agent regularly posts to Slack, summarizes channels, or triages threads.

> **Heads-up:** The reference server was [archived in May 2025](https://github.com/modelcontextprotocol/servers-archived/tree/main/src/slack) and no longer receives updates from the MCP team. It's still installable on npm and works for typical bot-token use, but new projects should consider the actively maintained fork [`korotovsky/slack-mcp-server`](https://github.com/korotovsky/slack-mcp-server) (see *Alternative* below) — especially if you need new Slack API surface or stdio-only operation.

## When to Trigger

- The user automates Slack notifications, summaries, or incident updates.
- The user wants Slack tools in the main session.
- Verify first: `grep -q '"slack"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "slack": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-slack"],
      "env": {
        "SLACK_BOT_TOKEN": "$SLACK_BOT_TOKEN",
        "SLACK_TEAM_ID": "$SLACK_TEAM_ID"
      },
      "includeTools": []
    }
  }
}
```

Optional: pin `SLACK_CHANNEL_IDS` (comma-separated) to scope reads.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc slack
```

## Authentication

1. Create a Slack app at <https://api.slack.com/apps> (or reuse an existing one).
2. Add the bot scopes you need (typical minimum: `channels:history`, `channels:read`, `chat:write`, `users:read`).
3. Install the app to the workspace, then copy the **Bot User OAuth Token** (`xoxb-…`) and the workspace **Team ID** (`T…`):

```bash
export SLACK_BOT_TOKEN=xoxb-...
export SLACK_TEAM_ID=T...
```

## Alternative

For a permission-light, stdio-only fork that avoids re-installing the bot, try [`korotovsky/slack-mcp-server`](https://github.com/korotovsky/slack-mcp-server).

## Documentation

- [`@modelcontextprotocol/server-slack` on npm](https://www.npmjs.com/package/@modelcontextprotocol/server-slack)
- [Slack MCP server guide (slack.com)](https://slack.com/help/articles/48855576908307-Guide-to-the-Slack-MCP-server)
- [Slack OAuth scopes reference](https://api.slack.com/scopes)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
