---
name: agent-mcp
description: Configure MCP servers for Antigravity, OpenCode, and Claude — stdio or remote transport, workspace or user scope. Use when adding, connecting, or troubleshooting an MCP server.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/agent-mcp
  created: 2026-06-23
  updated: 2026-07-06
---

# Configure Agent MCP Servers

Configure Model Context Protocol (MCP) servers for Antigravity, OpenCode, and Claude — stdio runners, remote HTTP endpoints, and Google Cloud managed MCP.

## Workflow

1. **Locate Config**:
   - **Antigravity**: Workspace `.agents/mcp_config.json` (default) or global `~/.gemini/antigravity-cli/mcp_config.json`. Add servers under the `"mcpServers"` object.
   - **OpenCode**: Workspace `opencode.json` (default) or global `~/.config/opencode/opencode.json`. Add servers under the `"mcp"` object.
   - **Claude**: Workspace `.mcp.json` (default) or global `~/.claude.json`. Add servers under the `"mcpServers"` object.
1. **Pick the Server**: Find the server, transport, and auth from an authoritative catalog (see [Catalogs](#catalogs)) rather than a local snapshot — upstream docs stay current.
1. **Write Configuration**: Choose the transport (stdio for local runners, remote HTTP for hosted endpoints) and use the matching schema below.
1. **Verify**: Run `/mcp` inside the Antigravity or Claude session, `opencode mcp list`, or `claude mcp list` to check server status.

## Schemas

### Stdio (local runner)

Antigravity:

```json
"server-name": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-everything"],
  "env": { "ENV_VAR": "value" }
}
```

OpenCode:

```json
"server-name": {
  "type": "local",
  "command": ["npx", "-y", "@modelcontextprotocol/server-everything"],
  "environment": { "ENV_VAR": "value" },
  "enabled": true
}
```

Claude:

```json
"server-name": {
  "type": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-everything"],
  "env": { "ENV_VAR": "value" }
}
```

### Remote (HTTP endpoint)

Antigravity (use `serverUrl`; `httpUrl` is the legacy Gemini-CLI alias):

```json
"server-name": {
  "serverUrl": "https://run.googleapis.com/mcp",
  "authProviderType": "google_credentials",
  "oauth": { "scopes": ["https://www.googleapis.com/auth/cloud-platform"] },
  "headers": { "x-goog-user-project": "PROJECT_ID" }
}
```

OpenCode:

```json
"server-name": {
  "type": "remote",
  "url": "https://run.googleapis.com/mcp",
  "headers": { "x-goog-user-project": "PROJECT_ID" },
  "enabled": true
}
```

Claude:

```json
"server-name": {
  "type": "http",
  "url": "https://run.googleapis.com/mcp",
  "headers": { "x-goog-user-project": "PROJECT_ID" }
}
```

## Google Cloud Managed MCP

Most Google Cloud products expose a managed MCP server behind a common pattern:

1. **Enable** the API and its MCP endpoint:
   ```bash
   gcloud services enable <service>.googleapis.com
   gcloud beta services mcp enable <service>.googleapis.com
   gcloud auth application-default login
   ```
1. **Configure** the remote server (Antigravity example above), pointing `serverUrl` at `https://<service>.googleapis.com/mcp` and setting the `x-goog-user-project` billing/quota header.
1. **Grant IAM**: the principal needs `roles/mcp.toolUser` **plus** a product role (e.g. `roles/bigquery.user` for BigQuery reads, `roles/run.developer` for Cloud Run).

The current list of managed products, endpoints, and required roles is maintained upstream — see [Catalogs](#catalogs). Prefer it over hardcoding a per-service copy here.

## Catalogs

- **Google Cloud managed MCP**: [Supported products](https://docs.cloud.google.com/mcp/supported-products) · [MCP overview](https://docs.cloud.google.com/mcp)
- **Official & community servers**: [MCP registry](https://registry.modelcontextprotocol.io)
- **Vendor servers** (GitHub, Atlassian, Firebase, Terraform, Databricks, Airtable, …): consult each vendor's current MCP docs for the endpoint, transport, and auth.

## Gotchas

1. **Authentication Errors (401/403)**: Ensure Google Application Default Credentials (ADC) are configured (`gcloud auth application-default login`) or OAuth tokens are authorized.
1. **Stdio Executables**: Verify the required runner/binary (`npx`, `uv`, `docker`, etc.) is on your system `PATH`.
1. **Remote Key**: Antigravity expects `serverUrl` for remote HTTP servers; Claude expects `type: "http"` and `url`; `httpUrl` still works as a legacy alias in Antigravity but is deprecated.
