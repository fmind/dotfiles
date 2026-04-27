---
name: install-error-reporting-mcp
description: Install the GCP Error Reporting MCP server in the current project's .gemini/settings.json so Gemini can call Error Reporting tools without going through the error-reporting subagent.
---

# Install Error Reporting MCP

Drops the GCP Error Reporting MCP server into `.gemini/settings.json` for the current project. Use this when error triage happens in nearly every session of the project — otherwise prefer the `error-reporting` subagent (`~/.gemini/agents/error-reporting.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project triages crashes/exceptions, monitors regressions, or assigns error groups across services and versions.
- The user wants error-triage tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"error-reporting"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "error-reporting": {
      "httpUrl": "https://clouderrorreporting.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc error-reporting
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `error-reporting` subagent (`~/.gemini/agents/error-reporting.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Error Reporting — MCP reference](https://docs.cloud.google.com/error-reporting/reference_mcp/mcp)
- [Error Reporting](https://docs.cloud.google.com/error-reporting/docs)
- [Reporting errors API](https://docs.cloud.google.com/error-reporting/reference/rest)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
