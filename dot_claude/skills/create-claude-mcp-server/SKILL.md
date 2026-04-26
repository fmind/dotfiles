---
name: create-claude-mcp-server
description: Guide for adding MCP servers to Claude Code
---

# Add an MCP Server to Claude Code

The Model Context Protocol (MCP) lets Claude Code reach external tools, APIs, and data sources via a small server process. For more details, refer to the [official Claude Code MCP documentation](https://docs.claude.com/en/docs/claude-code/mcp).

## Scopes

Claude Code resolves MCP servers from three scopes, listed from broadest to narrowest:

| Scope | File | Tracked in git? | Use for |
|---|---|---|---|
| User | `~/.claude.json` | No | Personal servers available everywhere. |
| Project | `.mcp.json` (repo root) | Yes | Servers the whole team needs. |
| Local | `.claude/settings.local.json` | No | Per-clone overrides and secrets. |

Project servers are the default in this dotfiles repo. Servers from `.mcp.json` are gated behind the user's `enableAllProjectMcpServers` setting (off by default).

## File Structure

`.mcp.json`:

```jsonc
{
  "mcp_servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "filesystem": {
      "command": "uvx",
      "args": ["mcp-server-filesystem", "/home/fmind/Documents"]
    }
  }
}
```

### Server transport types

- **Stdio (default):** the server runs as a subprocess; Claude Code talks to it over stdin/stdout.
- **HTTP:** add `"transport": "http"` and a `"url"`. Use for remote, multi-tenant servers.
- **SSE (deprecated for new servers):** legacy server-sent events transport.

## Adding a server via the CLI

Instead of editing JSON by hand, use the `claude mcp add` subcommands:

```bash
# Add a stdio server to user scope
claude mcp add github npx -y @modelcontextprotocol/server-github

# Add to project scope (writes .mcp.json)
claude mcp add --scope project filesystem uvx mcp-server-filesystem /tmp

# List configured servers
claude mcp list

# Remove
claude mcp remove github
```

## Authentication patterns

- **Env vars (recommended):** reference `${VAR}` in `env`. Source them from `~/.private.fish` or chezmoi-managed encrypted secrets.
- **OAuth / device flow:** the server handles auth on first run and caches tokens locally. Inspect the server's docs.
- **Static API keys:** if you must inline them, put the server in `~/.claude.json` (user scope) — never in `.mcp.json` (committed).

## Step-by-Step Creation

1. **Pick a server.** Check the [MCP server registry](https://github.com/modelcontextprotocol/servers) before writing your own.
2. **Decide the scope.** Personal → user scope. Team-shared → project scope. Anything with secrets → local.
3. **Add it** with `claude mcp add` so the JSON stays valid.
4. **Wire credentials** through env vars. Never commit secrets.
5. **Verify** with `claude mcp list` and by invoking a tool from the server in a session.

## Guidelines

- **Trust matters:** an MCP server can read everything Claude Code reads. Treat the binary like any other dependency.
- **Prefer stdio servers** unless you have a real reason to host one over HTTP.
- **Pin versions** for npm/uvx-based servers in `.mcp.json` so behavior is reproducible.
- **Keep `enableAllProjectMcpServers` off** globally; opt in per repo via `.claude/settings.local.json`.
