---
name: install-firebase-mcp
description: Install the Firebase MCP server in the current project's .gemini/settings.json so Gemini can call Firebase tools without going through the firebase subagent.
---

# Install Firebase MCP

Drops the Firebase MCP server into `.gemini/settings.json` for the current project. Use this when Firebase work happens in nearly every session of the project — otherwise prefer the `firebase` subagent (`~/.gemini/agents/firebase.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo contains `firebase.json`, `firestore.rules`, `apphosting.yaml`, or imports `firebase/*` / `firebase-admin`.
- The user wants Firebase tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"firebase"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "firebase": {
      "command": "npx",
      "args": ["-y", "firebase-tools@latest", "mcp"],
      "includeTools": []
    }
  }
}
```

Optionally pass `--only` to scope the server to specific feature groups:

```json
"args": ["-y", "firebase-tools@latest", "mcp", "--only", "auth,firestore,storage"]
```

Feature groups: `core`, `firestore`, `auth`, `dataconnect`, `storage`, `messaging`, `functions`, `remoteconfig`, `crashlytics`, `apphosting`, `realtimedatabase`.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc firebase
```

## Authentication

Requires the Firebase CLI to be logged in:

```bash
firebase login
firebase projects:list
firebase use <project-id>
```

## Companion Agent

The `firebase` subagent (`~/.gemini/agents/firebase.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Firebase MCP server (CLI docs)](https://firebase.google.com/docs/cli/mcp-server)
- [Firebase MCP server (AI assistance docs)](https://firebase.google.com/docs/ai-assistance/mcp-server)
- [Firebase agent skills (companion bundle)](https://firebase.google.com/docs/ai-assistance/agent-skills)
- [Firebase CLI reference](https://firebase.google.com/docs/cli)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
