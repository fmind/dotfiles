---
name: install-chat-mcp
description: Install the Google Chat MCP server in the current project's .gemini/settings.json so Gemini can call Chat tools without going through the chat subagent.
---

# Install Chat MCP

Drops the Google Chat MCP server into `.gemini/settings.json` for the current project. Use this when Chat reads/posts happen in nearly every session of the project — otherwise prefer the `chat` subagent (`~/.gemini/agents/chat.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project automates messaging, summarizes threads, or posts notifications to Google Chat spaces.
- The user wants Chat tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"chat"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "chat": {
      "httpUrl": "https://chatmcp.googleapis.com/mcp/v1",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc chat
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `chat` subagent (`~/.gemini/agents/chat.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the Chat MCP server](https://developers.google.com/workspace/chat/api/guides/configure-mcp-server)
- [Chat MCP reference](https://developers.google.com/workspace/chat/api/reference/mcp)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Google Chat API](https://developers.google.com/workspace/chat)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
