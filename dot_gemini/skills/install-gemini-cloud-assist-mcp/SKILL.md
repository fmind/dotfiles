---
name: install-gemini-cloud-assist-mcp
description: Install the Gemini Cloud Assist MCP server in the current project's .gemini/settings.json so Gemini can call Cloud Assist tools (architecture design, troubleshooting, cost analysis) without going through the gemini-cloud-assist subagent.
---

# Install Gemini Cloud Assist MCP

Drops the **Gemini Cloud Assist** MCP server into `.gemini/settings.json` for the current project. Gemini Cloud Assist is the AI assistant for *Google Cloud infrastructure* — design, troubleshoot, and optimize GCP architectures. (It is a **different product** from Gemini Code Assist, which assists with code in IDEs and on GitHub.)

Use this skill when Cloud Assist work happens in nearly every session — otherwise prefer the `gemini-cloud-assist` subagent (`~/.gemini/agents/gemini-cloud-assist.md`), which loads the MCP only when invoked.

## When to Trigger

- The project regularly designs GCP architectures, troubleshoots Cloud incidents, or analyzes Google Cloud cost / performance.
- The user wants Cloud Assist tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"gemini-cloud-assist"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "gemini-cloud-assist": {
      "httpUrl": "https://geminicloudassist.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc gemini-cloud-assist
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The Gemini Cloud Assist API must be enabled on the target project:

```bash
gcloud services enable geminicloudassist.googleapis.com --project=<PROJECT_ID>
```

## Companion Agent

The `gemini-cloud-assist` subagent (`~/.gemini/agents/gemini-cloud-assist.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Use the Gemini Cloud Assist remote MCP server](https://docs.cloud.google.com/cloud-assist/use-gemini-cloud-assist-mcp)
- [Gemini Cloud Assist MCP reference](https://docs.cloud.google.com/gemini/docs/geminicloudassist/reference/mcp)
- [Gemini Cloud Assist overview](https://docs.cloud.google.com/gemini/docs/cloud-assist/overview)
- [GoogleCloudPlatform/gemini-cloud-assist-mcp (companion repo)](https://github.com/GoogleCloudPlatform/gemini-cloud-assist-mcp)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
