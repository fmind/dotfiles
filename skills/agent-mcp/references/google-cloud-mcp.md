# Google Cloud Managed MCP

Google Cloud exposes managed MCP endpoints for some products. The `gcloud beta services mcp` command below is carried over from the previous revision of [agent-mcp](SKILL.md) and was not re-verified against a current `gcloud`; confirm it with `gcloud beta services mcp --help` before relying on it.

1. **Enable** the API and its MCP endpoint, then authenticate:
   ```bash
   gcloud services enable <service>.googleapis.com
   gcloud beta services mcp enable <service>.googleapis.com   # unverified
   gcloud auth application-default login
   ```
1. **Configure** the documented product endpoint and the `x-goog-user-project` quota-project header with the host's native add command.
1. **Grant minimum IAM**: `roles/mcp.toolUser` plus only the product role the requested operations need; pin account and project per [gcloud](../gcloud/SKILL.md).

Use the current [supported products](https://docs.cloud.google.com/mcp/supported-products), [MCP overview](https://docs.cloud.google.com/mcp), and [MCP registry](https://registry.modelcontextprotocol.io) instead of a local server catalog.
