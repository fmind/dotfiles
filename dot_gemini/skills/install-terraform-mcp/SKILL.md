---
name: install-terraform-mcp
description: Install HashiCorp Terraform MCP into .gemini/settings.json. Use when Terraform work is central to the project.
---

# Install Terraform MCP

Drops the HashiCorp Terraform MCP server into `.gemini/settings.json` for the current project. Use this when IaC plan/apply/lookup work happens in nearly every session of the project — otherwise prefer the `terraform` subagent (`~/.gemini/agents/terraform.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The repo contains `*.tf` / `*.tfvars`, `terraform.tfstate`, or a `.terraform/` directory.
- The user wants Terraform tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"terraform"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

The HashiCorp Terraform MCP server is a Go binary, distributed as a Docker image (`hashicorp/terraform-mcp-server`) or installable via `go install`. It is **not** an npm package.

### Option A — Docker (recommended)

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "terraform": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "hashicorp/terraform-mcp-server"],
      "includeTools": []
    }
  }
}
```

### Option B — Local Go binary

Install once (`go install github.com/hashicorp/terraform-mcp-server/cmd/terraform-mcp-server@latest`), then point Gemini at the resulting binary:

```json
{
  "mcpServers": {
    "terraform": {
      "command": "terraform-mcp-server",
      "args": ["stdio"],
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc terraform
```

## Authentication

No authentication required for the MCP server itself. Cloud-provider operations still rely on the local Terraform binary and your provider credentials (e.g. `gcloud auth application-default login` for the Google provider).

## Companion Agent

The `terraform` subagent (`~/.gemini/agents/terraform.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Terraform MCP server](https://github.com/hashicorp/terraform-mcp-server)
- [Terraform Google provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
