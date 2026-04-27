---
name: install-cloud-storage-mcp
description: Install the GCP Cloud Storage MCP server in the current project's .gemini/settings.json so Gemini can list, read, and manage GCS buckets and objects without going through a subagent.
---

# Install Cloud Storage MCP

Drops the GCP Cloud Storage MCP server into `.gemini/settings.json` for the current project. Use this when bucket/object listing, IAM, lifecycle, or signed-URL work happens in nearly every session.

## When to Trigger

- The repo deploys to GCS, hosts artefacts/static assets there, or imports `google-cloud-storage` / `@google-cloud/storage`.
- The user wants ad-hoc bucket/object tools in the main session.
- Verify first: `grep -q '"cloud-storage"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "cloud-storage": {
      "httpUrl": "https://storage.googleapis.com/storage/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc cloud-storage
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The Storage API must be enabled:

```bash
gcloud services enable storage.googleapis.com --project=<PROJECT_ID>
```

## Documentation

- [Cloud Storage — MCP reference](https://docs.cloud.google.com/storage/docs/reference/mcp)
- [Cloud Storage overview](https://docs.cloud.google.com/storage/docs)
- [`gcloud storage` CLI](https://docs.cloud.google.com/sdk/gcloud/reference/storage)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
