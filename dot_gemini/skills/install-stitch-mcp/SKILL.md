---
name: install-stitch-mcp
description: Install the Google Stitch MCP server in the current project's .gemini/settings.json so Gemini can call Stitch UI design tools without going through the stitch subagent.
---

# Install Stitch MCP

Drops the Google Stitch MCP server into `.gemini/settings.json` for the current project. Use this when AI-driven UI design generation happens in nearly every session of the project — otherwise prefer the `stitch` subagent (`~/.gemini/agents/stitch.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project iterates on app screens, multi-screen flows, or design systems and exports to Figma / React / HTML.
- The user wants Stitch tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"stitch"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing). Pick **one** of the two auth modes below.

### Option A — API key (recommended for individuals; non-expiring)

```json
{
  "mcpServers": {
    "stitch": {
      "httpUrl": "https://stitch.googleapis.com/mcp",
      "headers": {
        "X-Goog-Api-Key": "$STITCH_API_KEY"
      },
      "timeout": 300000,
      "includeTools": []
    }
  }
}
```

Generate the key in Stitch settings → API keys.

### Option B — Application Default Credentials (OAuth, refreshes hourly)

```json
{
  "mcpServers": {
    "stitch": {
      "httpUrl": "https://stitch.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "oauth": {
        "scopes": ["https://www.googleapis.com/auth/cloud_platform"]
      },
      "headers": {
        "X-Goog-User-Project": "$GOOGLE_CLOUD_PROJECT"
      },
      "timeout": 300000,
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc stitch
```

## Authentication

For Option B, run `gcloud auth application-default login` if not already authenticated, then enable the API:

```bash
gcloud services enable stitch.googleapis.com --project=<PROJECT_ID>
```

## Companion Agent

The `stitch` subagent (`~/.gemini/agents/stitch.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Google Stitch](https://stitch.withgoogle.com)
- [Stitch MCP setup](https://stitch.withgoogle.com/docs/mcp/setup/)
- [Stitch MCP guide](https://stitch.withgoogle.com/docs/mcp/guide/)
- [Stitch MCP reference](https://stitch.withgoogle.com/docs/mcp/reference/)
- [Gemini CLI extension (config templates)](https://github.com/gemini-cli-extensions/stitch)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
