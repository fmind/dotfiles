---
name: agent-mcp
description: Configure MCP servers for Antigravity, Claude Code, Codex, Copilot, Grok, and OpenCode with each host's native add command at project or user scope. Use when an agent needs an MCP server.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/agent-mcp
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Configure Agent MCP Servers

Add Model Context Protocol servers through each host's native command so the config file shape stays correct; [agent-project](../agent-project/SKILL.md) owns the surrounding repository layout.

## Commands

One stdio example per host. For a remote server replace the trailing command with `--transport http <url>` (Claude Code, Copilot, Grok), `--url <url>` (Codex, OpenCode), or a plain URL argument (Antigravity):

```bash
agy mcp add --env KEY=value <name> -- npx -y <package>                   # Antigravity CLI; flags before <name>
claude mcp add --scope project -e KEY=value <name> -- npx -y <package>   # default scope is local, not project
codex mcp add <name> --env KEY=value -- npx -y <package>                 # no scope flag: writes ~/.codex/config.toml
copilot mcp add --env KEY=value <name> -- npx -y <package>               # user configuration
grok mcp add --scope project -e KEY=value <name> -- npx -y <package>     # --scope user is the default
opencode mcp add <name> --env KEY=value                                  # prompts for the remaining fields
```

## Config Files

| Host        | User scope                                           | Project scope                                        |
| ----------- | ---------------------------------------------------- | ---------------------------------------------------- |
| Antigravity | `~/.gemini/config/mcp_config.json` (per its docs)    | `.agents/mcp_config.json` (per its docs)             |
| Claude Code | `~/.claude.json` (`--scope user` or default `local`) | `.mcp.json` (`--scope project`)                      |
| Codex       | `~/.codex/config.toml` under `[mcp_servers.<name>]`  | `.codex/config.toml` (trusted projects, hand-edited) |
| Copilot     | `~/.copilot/mcp-config.json`                         | `.mcp.json` or `.github/mcp.json` (hand-edited)      |
| Grok        | `~/.grok/config.toml`                                | `./.grok/config.toml`                                |
| OpenCode    | `~/.config/opencode/opencode.json` under `"mcp"`     | `opencode.json` under `"mcp"`                        |

## Workflow

1. **Review the server**: verify publisher, executable or URL, requested credentials, and tools against upstream documentation before adding it.
1. **Choose transport**: stdio for a local executable, streamable HTTP for a hosted endpoint; legacy SSE only when the provider requires it.
1. **Keep secrets external**: pass `KEY=value` from the environment or use the host's OAuth flow (`codex mcp add --bearer-token-env-var`, `opencode mcp auth`); never write a token into a committed file.
1. **Add with the native command**: hand-written JSON drifts because the remote-URL key differs per host (`serverUrl` in Antigravity, `httpUrl` in Gemini CLI, `url` elsewhere), so let the CLI write it.
1. **Verify**: `agy mcp list`, `claude mcp list`, `codex mcp list`, `copilot mcp list`, `grok mcp list`, or `opencode mcp list`.

## Gotchas

- **Claude default scope is `local`**: the server lands in `~/.claude.json` for this path only; pass `--scope project` to share it through `.mcp.json`.
- **Repository trust**: review project MCP files before starting their servers; do not auto-approve every repository-provided server.
- **Runner resolution**: `npx`, `uvx`, and `docker` must resolve from the agent's environment, not only from your shell.
- **Tool scope**: enable only the tools a workflow needs (`copilot mcp add --tools`) and keep write-capable tools approval-gated.
- **Auth errors**: confirm OAuth, Application Default Credentials, scopes, and IAM before broadening permissions.

## Documentation

- [Model Context Protocol](https://modelcontextprotocol.io) · [MCP registry](https://registry.modelcontextprotocol.io) · [Google Cloud managed MCP](references/google-cloud-mcp.md)
- Companion skills: [agent-project](../agent-project/SKILL.md) (repository layout), [gcloud](../gcloud/SKILL.md) (project and IAM context for managed servers).
