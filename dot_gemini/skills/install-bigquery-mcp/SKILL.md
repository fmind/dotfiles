---
name: install-bigquery-mcp
description: Install the GCP BigQuery MCP server in the current project's .gemini/settings.json so Gemini can run SQL queries, manage datasets/tables, and inspect jobs without going through a subagent.
---

# Install BigQuery MCP

Drops the GCP BigQuery MCP server into `.gemini/settings.json` for the current project. Use this when SQL queries, dataset/table mgmt, or job inspection happen in nearly every session.

## When to Trigger

- The repo contains `*.sql`, `bq` CLI scripts, BigQuery schemas (`schema.json`), or imports `google-cloud-bigquery` / `@google-cloud/bigquery`.
- The user wants ad-hoc query / table tools available in the main session.
- Verify first: `grep -q '"bigquery"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "bigquery": {
      "httpUrl": "https://bigquery.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc bigquery
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The BigQuery API must be enabled:

```bash
gcloud services enable bigquery.googleapis.com --project=<PROJECT_ID>
```

Set a default billing project for queries:

```bash
gcloud config set project <PROJECT_ID>
```

## Documentation

- [BigQuery — MCP reference](https://docs.cloud.google.com/bigquery/docs/reference/mcp)
- [BigQuery overview](https://docs.cloud.google.com/bigquery/docs)
- [BigQuery `bq` CLI](https://docs.cloud.google.com/bigquery/docs/bq-command-line-tool)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
