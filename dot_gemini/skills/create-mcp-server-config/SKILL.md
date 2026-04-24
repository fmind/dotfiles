---
name: create-mcp-server-config
description: Add a Model Context Protocol (MCP) server to a Gemini agent or to global settings.
---

# Create MCP Server Config

This skill documents the canonical patterns for wiring a Model Context Protocol (MCP) server into Gemini CLI, either inside an agent file or globally.

## Where to declare an MCP server

| Scope                     | File                                              | When to use                                              |
| ------------------------- | ------------------------------------------------- | -------------------------------------------------------- |
| Agent-only                | `~/.gemini/agents/<agent>.md` (`mcp_servers:` block) | The MCP is specific to one persona.                      |
| Global, all agents        | `~/.gemini/settings.json` (`mcpServers` key)      | The MCP is generally useful (filesystem, git, fetch, …). |
| Workspace, all agents     | `.gemini/settings.json` (same key)                | Project-specific MCP (e.g. local DB).                    |

## Patterns

### 1. Local stdio MCP via npx

```yaml
mcp_servers:
  filesystem:
    command: npx
    args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
```

### 2. Local stdio MCP via uvx (Python)

```yaml
mcp_servers:
  git:
    command: uvx
    args: ["mcp-server-git"]
```

### 3. Local stdio MCP in a Docker container

```yaml
mcp_servers:
  github_local:
    command: docker
    args:
      - run
      - "-i"
      - "--rm"
      - "-e"
      - GITHUB_PERSONAL_ACCESS_TOKEN
      - "ghcr.io/github/github-mcp-server"
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: "$GITHUB_MCP_PAT"
```

### 4. Remote HTTP MCP with Google credentials (GCP / Workspace)

```yaml
mcp_servers:
  firestore:
    httpUrl: "https://firestore.googleapis.com/mcp"
    authProviderType: "google_credentials"
```

### 5. Remote HTTP MCP with bearer token

```yaml
mcp_servers:
  github_http:
    httpUrl: "https://api.githubcopilot.com/mcp/"
    headers:
      Authorization: "Bearer $GITHUB_MCP_PAT"
```

## Step-by-Step

1. **Identify the protocol.** Stdio (local binary) or HTTP (remote endpoint)?
1. **Identify the auth.** None, env var, bearer header, or
   `google_credentials` (uses the active `gcloud` ADC).
1. **Pick a scope.** Single agent vs global vs workspace.
1. **Drop in the snippet** from the matching pattern above.
1. **Restart Gemini CLI** so MCP servers re-handshake.
1. **Verify** with `/mcp list` (or whatever Gemini exposes) and a smoke prompt.

## Authoritative MCP catalogues

- Google Cloud / Workspace remote MCPs:
  <https://docs.cloud.google.com/mcp/supported-products>
- Reference local MCP servers:
  <https://github.com/modelcontextprotocol/servers>
- Google-maintained local MCPs: <https://github.com/Google/mcp>

## Guidelines

- Never inline secrets. Use `$ENV_VAR` or bracket templating
  (`{{ .ENV_VAR }}`) so the file is safe to commit.
- Keep agent-local MCPs minimal — duplication makes startup slower.
- For Google APIs, always prefer `authProviderType: google_credentials` over
  hand-rolled OAuth flows.
