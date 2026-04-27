---
name: install-github-mcp
description: Install the GitHub MCP server (HTTP or local Docker) in the current project's .gemini/settings.json so Gemini can call GitHub tools without going through the github subagent.
---

# Install GitHub MCP

Drops the GitHub MCP server into `.gemini/settings.json` for the current project. Use this when GitHub work (PRs, issues, Actions, code search) happens in nearly every session of the project — otherwise prefer the `github` subagent (`~/.gemini/agents/github.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo lives on GitHub and the user routinely reviews PRs, manages issues, dispatches workflows, or searches code across the org.
- The user wants GitHub tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"github_http\|github_local"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Two transports — pick one (HTTP preferred, Docker fallback). Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "github_http": {
      "httpUrl": "https://api.githubcopilot.com/mcp/",
      "headers": {
        "Authorization": "Bearer $GITHUB_MCP_PAT"
      },
      "includeTools": []
    },
    "github_local": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN",
        "ghcr.io/github/github-mcp-server"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "$GITHUB_MCP_PAT"
      },
      "includeTools": []
    }
  }
}
```

Use `github_http` when network access to the hosted MCP is available; fall back to `github_local` (Docker) otherwise. Keep only one in `mcpServers` to avoid duplicate tool registrations.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc github_http
```

## Authentication

Both transports require the `GITHUB_MCP_PAT` environment variable to be exported with a fine-grained Personal Access Token (typical scopes: `repo`, `workflow`, `read:org`).

```bash
export GITHUB_MCP_PAT=ghp_xxx
```

## Companion Agent

The `github` subagent (`~/.gemini/agents/github.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [GitHub MCP server (repo)](https://github.com/github/github-mcp-server)
- [About Model Context Protocol (MCP)](https://docs.github.com/en/copilot/concepts/context/mcp)
- [Set up the GitHub MCP server](https://docs.github.com/en/copilot/how-tos/provide-context/use-mcp/set-up-the-github-mcp-server)
- [Fine-grained PATs](https://docs.github.com/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
