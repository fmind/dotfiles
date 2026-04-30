---
name: install-clasp-mcp
description: Install Clasp MCP (Google Apps Script tools) into .gemini/settings.json. Use when Apps Script work is central to the project.
---

# Install Clasp MCP

Drops the Clasp (Apps Script) MCP server into `.gemini/settings.json` for the current project. Use this when Apps Script work happens in nearly every session of the project — otherwise prefer the `clasp` subagent (`~/.gemini/agents/clasp.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo contains `.clasp.json`, `appsscript.json`, or pushes/pulls to Google Apps Script.
- The user wants Clasp tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"clasp"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "clasp": {
      "command": "clasp",
      "args": ["mcp"],
      "includeTools": []
    }
  }
}
```

> Clasp's MCP mode is marked **EXPERIMENTAL** in the upstream README.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc clasp
```

## Authentication

Requires the Clasp CLI to be installed and logged in:

```bash
npm install -g @google/clasp
clasp login
```

## Companion Agent

The `clasp` subagent (`~/.gemini/agents/clasp.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Clasp on GitHub](https://github.com/google/clasp)
- [Clasp MCP mode (experimental)](https://github.com/google/clasp#mcp-experimental)
- [Apps Script reference](https://developers.google.com/apps-script/reference)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
