---
name: install-google-workspace-mcp
description: Install one or more Google Workspace MCP servers (Calendar, Chat, Drive, Gmail, People) into .gemini/settings.json. Use when Workspace work is central to the project.
---

# Install Google Workspace MCP

Drops one or more **Google Workspace MCP servers** into `.gemini/settings.json` for the current project. Every Workspace server shares the same install shape — an `httpUrl` plus an `oauth` block — and authenticates via a per-user OAuth 2.0 flow (**not** Application Default Credentials).

Use this skill when one or more Workspace services are central to the project. Otherwise prefer the per-product subagent at `~/.gemini/agents/<service>.md`, which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

The repo, configuration, or user request indicates regular use of a Workspace service. Always:

1. Pick the **minimum** set actually needed in the main session.
2. Skip entries already installed: `grep -q '"<server-name>"' .gemini/settings.json 2>/dev/null`.
3. For a current canonical list (incl. new Workspace servers), see <https://docs.cloud.google.com/mcp/supported-products>.

## Catalog

All scope strings below are prefixed with `https://www.googleapis.com/auth/`.

| Server name (JSON key) | `httpUrl` | Default (read-only) scopes | Promote to write by replacing with |
|---|---|---|---|
| `calendar` | `https://calendarmcp.googleapis.com/mcp/v1` | `calendar.calendarlist.readonly`, `calendar.events.freebusy`, `calendar.events.readonly` | `calendar.events` to create / update / cancel |
| `chat` | `https://chatmcp.googleapis.com/mcp/v1` | `chat.spaces.readonly`, `chat.memberships.readonly`, `chat.messages.readonly`, `chat.users.readstate.readonly` | add `chat.messages`, `chat.memberships` to send / manage |
| `drive` | `https://drivemcp.googleapis.com/mcp/v1` | `drive.readonly`, `drive.file` | `drive` for arbitrary-file mgmt (verification-gated) |
| `gmail` | `https://gmailmcp.googleapis.com/mcp/v1` | `gmail.readonly`, `gmail.compose` | add `gmail.send` to actually send |
| `people` | `https://people.googleapis.com/mcp/v1` | `directory.readonly`, `userinfo.profile`, `contacts.readonly` | `contacts` to create / update / merge |

Defaults are intentionally **read-only**. Promote to write scopes only when the project actually needs to modify Workspace state.

## Install

For each chosen server, merge an entry into `.gemini/settings.json` under a single `mcpServers` object. The same OAuth client can be reused across all Workspace MCPs — keep `clientId` / `clientSecret` consistent across entries.

```json
{
  "mcpServers": {
    "<server-name>": {
      "httpUrl": "<endpoint from the catalog>",
      "oauth": {
        "enabled": true,
        "clientId": "$GOOGLE_OAUTH_CLIENT_ID",
        "clientSecret": "$GOOGLE_OAUTH_CLIENT_SECRET",
        "scopes": [
          "https://www.googleapis.com/auth/<scope-1>",
          "https://www.googleapis.com/auth/<scope-2>"
        ]
      },
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc <server-name>
```

## Authentication

Workspace MCPs use a per-user OAuth 2.0 flow. One-time setup:

1. In GCP Console, create an OAuth 2.0 client of type **Desktop app** and enable each per-product API on the project (`calendar.googleapis.com`, `chat.googleapis.com`, `drive.googleapis.com`, `gmail.googleapis.com`, `people.googleapis.com`).
2. Export the OAuth secrets in your shell — one client serves all Workspace MCPs:

   ```bash
   export GOOGLE_OAUTH_CLIENT_ID=...
   export GOOGLE_OAUTH_CLIENT_SECRET=...
   ```

3. Run `/mcp auth <server-name>` in Gemini CLI for each installed server to trigger the browser consent flow. Tokens cache at `~/.gemini/mcp-oauth-tokens.json`.

## Companion Agents

Each Workspace MCP has a sibling subagent at `~/.gemini/agents/<service>.md` (e.g. `gmail`, `drive`). The subagent wraps the same MCP and loads it lazily — keep the subagent installed for cross-project ad-hoc work even after pinning the MCP at the project level.

## Documentation

- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products) — canonical Workspace MCP list.
- [Configure Google Workspace MCP servers (overview)](https://developers.google.com/workspace/guides/configure-mcp-servers)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
- Per-product configure-MCP pages:
  - [Calendar](https://developers.google.com/workspace/calendar/api/guides/configure-mcp-server)
  - [Chat](https://developers.google.com/workspace/chat/api/guides/configure-mcp-server)
  - [Drive](https://developers.google.com/workspace/drive/api/guides/configure-mcp-server)
  - [Gmail](https://developers.google.com/workspace/gmail/api/guides/configure-mcp-server)
  - [People](https://developers.google.com/people/v1/configure-mcp-server)
