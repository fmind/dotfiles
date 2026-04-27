---
name: install-compute-engine-mcp
description: Install the GCP Compute Engine MCP server in the current project's .gemini/settings.json so Gemini can manage VM instances, disks, networks, and snapshots without going through a subagent.
---

# Install Compute Engine MCP

Drops the GCP Compute Engine MCP server into `.gemini/settings.json` for the current project. Use this when VM lifecycle work (instances, MIGs, disks, networks, snapshots) happens in nearly every session.

## When to Trigger

- The repo provisions VMs (Terraform `google_compute_instance`, gcloud scripts) or interacts with the Compute Engine API.
- The user wants ad-hoc VM tools available in the main session.
- Verify first: `grep -q '"compute-engine"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "compute-engine": {
      "httpUrl": "https://compute.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc compute-engine
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated. The Compute Engine API must be enabled:

```bash
gcloud services enable compute.googleapis.com --project=<PROJECT_ID>
```

## Documentation

- [Compute Engine — MCP reference](https://docs.cloud.google.com/compute/docs/reference/mcp)
- [Compute Engine overview](https://docs.cloud.google.com/compute/docs)
- [`gcloud compute` CLI](https://docs.cloud.google.com/sdk/gcloud/reference/compute)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
