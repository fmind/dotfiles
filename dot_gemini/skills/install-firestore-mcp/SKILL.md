---
name: install-firestore-mcp
description: Install the GCP Firestore MCP server in the current project's .gemini/settings.json so Gemini can call Firestore tools without going through the firestore subagent.
---

# Install Firestore MCP

Drops the GCP Firestore MCP server into `.gemini/settings.json` for the current project. Use this when Firestore reads/writes happen in nearly every session of the project — otherwise prefer the `firestore` subagent (`~/.gemini/agents/firestore.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo contains `firestore.rules`, `firestore.indexes.json`, or imports `@google-cloud/firestore` / `firebase-admin/firestore`.
- The user wants Firestore tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"firestore"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "firestore": {
      "httpUrl": "https://firestore.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc firestore
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `firestore` subagent (`~/.gemini/agents/firestore.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Firestore — MCP reference](https://docs.cloud.google.com/firestore/docs/reference/mcp)
- [Firestore](https://docs.cloud.google.com/firestore/docs)
- [Firestore queries](https://docs.cloud.google.com/firestore/docs/query-data/queries)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
