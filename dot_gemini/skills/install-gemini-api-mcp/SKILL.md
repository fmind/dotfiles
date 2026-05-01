---
name: install-gemini-api-mcp
description: Install Gemini API docs MCP (live doc-search) into .gemini/settings.json. Use when Gemini API doc lookup is needed in nearly every session.
---

# Install Gemini API MCP

Drops the Gemini API docs MCP server into `.gemini/settings.json` for the current project. Use this when live Gemini API doc lookups happen in nearly every session of the project — otherwise prefer the `gemini-dev` subagent (`~/.gemini/agents/gemini-dev.md`), which loads the MCP only when invoked and keeps the parent context clean.

> **Provenance:** `gemini-api-docs-mcp.dev` is a **community-maintained** MCP server by Philipp Schmid ([philschmid/gemini-api-docs-mcp](https://github.com/philschmid/gemini-api-docs-mcp)) that Google **recommends** in the Gemini API "Coding agents" docs. It is not operated by Google. If you need a strictly-official source, replace this entry with an `mcpdoc` server pointed at `https://ai.google.dev/llms.txt` (see "Alternative" below).

## When to Trigger

- The repo uses `google-genai` (Python) or `@google/genai` (TS/JS) and the user frequently checks current API patterns.
- The user wants `search_docs` available in the main session without invoking the subagent.
- Verify first: `grep -q '"gemini-mcp"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "gemini-mcp": {
      "httpUrl": "https://gemini-api-docs-mcp.dev/",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc gemini-mcp
```

The server primarily exposes `search_docs(query, detail?)` plus resource files (`llms.txt`, `coding-agents.md.txt`, `deprecations.md.txt`, `migrate.md.txt`).

## Alternative — Strictly-Official Docs MCP

If you'd rather not rely on a third-party endpoint, swap in `mcpdoc` pointed at the official LLMs index Google ships at `ai.google.dev/llms.txt`:

```json
{
  "mcpServers": {
    "gemini-docs": {
      "command": "uvx",
      "args": [
        "--from", "mcpdoc", "mcpdoc",
        "--urls", "GeminiAPI:https://ai.google.dev/llms.txt",
        "--transport", "stdio"
      ],
      "includeTools": []
    }
  }
}
```

## Authentication

No authentication required.

## Companion Agent

The `gemini-dev` subagent (`~/.gemini/agents/gemini-dev.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Gemini API docs](https://ai.google.dev/gemini-api/docs)
- [Gemini API reference](https://ai.google.dev/api)
- [Coding agents setup (MCP + Skills) — recommends this server](https://ai.google.dev/gemini-api/docs/coding-agents)
- [philschmid/gemini-api-docs-mcp (community source)](https://github.com/philschmid/gemini-api-docs-mcp)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
