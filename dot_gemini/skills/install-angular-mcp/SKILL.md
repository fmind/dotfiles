---
name: install-angular-mcp
description: Install the Angular CLI MCP server in the current project's .gemini/settings.json so Gemini can call Angular tools without going through the angular subagent.
---

# Install Angular MCP

Drops the Angular CLI MCP server into `.gemini/settings.json` for the current project. Use this when Angular work happens in nearly every session of the project — otherwise prefer the `angular` subagent (`~/.gemini/agents/angular.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo contains `angular.json`, `nx.json` with Angular plugins, or imports `@angular/core`.
- The user wants Angular scaffolding, doc search, and `ai_tutor` tools in the main session without invoking the subagent.
- Verify first: `grep -q '"angular"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "angular": {
      "command": "npx",
      "args": ["-y", "@angular/cli", "mcp"],
      "includeTools": []
    }
  }
}
```

Optionally pass flags via `args` to scope the server: `--read-only`, `--local-only`, or `-E <name>` to enable experimental tools (`devserver`, `modernize`, `build`, `test`, `e2e`).

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc angular
```

Default tools: `ai_tutor`, `find_examples`, `get_best_practices`, `list_projects`, `onpush_zoneless_migration`, `search_documentation`.

## Authentication

No authentication required.

## Companion Agent

The `angular` subagent (`~/.gemini/agents/angular.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Angular AI hub](https://angular.dev/ai)
- [MCP server setup](https://angular.dev/ai/mcp)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
