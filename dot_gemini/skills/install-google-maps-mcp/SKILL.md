---
name: install-google-maps-mcp
description: Install the Google Maps Code Assist MCP server in the current project's .gemini/settings.json so Gemini can ground answers in Maps Platform docs and samples without going through a subagent.
---

# Install Google Maps MCP

Drops the Google Maps Code Assist MCP server into `.gemini/settings.json` for the current project. The server grounds answers in Maps Platform documentation, code samples, and reference material — useful when wiring up Maps JS, Places, Routes, Geocoding, or Address Validation.

## When to Trigger

- The repo uses `@googlemaps/*` libraries, embeds Maps JS, or calls Places / Routes / Geocoding / Address Validation APIs.
- The user wants doc-grounded Maps answers in the main session.
- Verify first: `grep -q '"google-maps"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "google-maps": {
      "httpUrl": "https://mapscodeassist.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc google-maps
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The Maps Code Assist API must be enabled on the project:

```bash
gcloud services enable mapscodeassist.googleapis.com --project=<PROJECT_ID>
```

## Documentation

- [Google Maps Code Assist MCP](https://developers.google.com/maps/ai/mcp)
- [Maps Platform documentation](https://developers.google.com/maps/documentation)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
