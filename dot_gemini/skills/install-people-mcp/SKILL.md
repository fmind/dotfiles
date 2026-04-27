---
name: install-people-mcp
description: Install the Google People MCP server in the current project's .gemini/settings.json so Gemini can call People API tools without going through the people subagent.
---

# Install People MCP

Drops the Google People MCP server into `.gemini/settings.json` for the current project. Use this when contact and directory lookups happen in nearly every session of the project — otherwise prefer the `people` subagent (`~/.gemini/agents/people.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project resolves Workspace directory entries, looks up contacts, or maintains contact metadata.
- The user wants People API tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"people"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "people": {
      "httpUrl": "https://people.googleapis.com/mcp/v1",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc people
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `people` subagent (`~/.gemini/agents/people.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Configure the People MCP server](https://developers.google.com/people/v1/configure-mcp-server)
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [People API](https://developers.google.com/people)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
