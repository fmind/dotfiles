---
name: create-mcp-server
description: Build a custom MCP server. Defers to Anthropic's official `mcp-builder` skill for the full workflow; provides a FastMCP (Python) quickstart and structural conventions.
---

# Create MCP Server

Anthropic ships an **official** [`mcp-builder`](https://mcpservers.org/agent-skills/anthropic/mcp-builder) skill that teaches a four-phase workflow for authoring MCP servers in Python ([FastMCP](https://github.com/jlowin/fastmcp)) or TypeScript ([`@modelcontextprotocol/sdk`](https://www.npmjs.com/package/@modelcontextprotocol/sdk)). When the user wants to *build* an MCP server, prefer that skill — it covers planning, implementation, testing, and review in depth.

This skill bundles the dotfile-specific quickstart so you can spin up a server in 5 minutes without leaving the repo.

## When to Trigger

- The user wants to expose internal tools (custom APIs, scripts, data sources) to coding agents via MCP.
- The user mentions "build an MCP server", "wrap our API as MCP tools", FastMCP, or `@modelcontextprotocol/sdk`.

For deep authoring guidance, install the official skill:

```bash
# Anthropic's mcp-builder skill — the canonical authoring guide.
npx skills add anthropics/skills --skill mcp-builder
```

## FastMCP Quickstart (Python)

Compose with the existing `create-python-script` skill: drop a single-file MCP server with PEP 723 metadata.

```python
#!/usr/bin/env -S uv run --quiet --script
# /// script
# requires-python = ">=3.13"
# dependencies = ["fastmcp>=2.10", "httpx"]
# ///

from fastmcp import FastMCP
import httpx

mcp = FastMCP("acme-tools")

@mcp.tool()
def add(a: int, b: int) -> int:
    """Add two integers."""
    return a + b

@mcp.tool()
async def fetch_url(url: str) -> str:
    """Fetch a URL and return the response body as text."""
    async with httpx.AsyncClient() as client:
        r = await client.get(url, timeout=10.0)
        r.raise_for_status()
        return r.text

@mcp.resource("config://current")
def current_config() -> dict:
    """Read-only resource: current server config."""
    return {"name": "acme-tools", "version": "0.1.0"}

if __name__ == "__main__":
    mcp.run()                       # stdio by default
    # mcp.run(transport="http", port=8080)   # HTTP transport
```

Test locally:

```bash
chmod +x server.py
./server.py                         # spawns stdio MCP server
```

Register in `.gemini/settings.json`:

```json
{
  "mcpServers": {
    "acme-tools": {
      "command": "./server.py",
      "includeTools": ["add", "fetch_url"]
    }
  }
}
```

## TypeScript Quickstart

```typescript
// server.ts
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";

const server = new McpServer({ name: "acme-tools", version: "0.1.0" });

server.registerTool(
  "add",
  {
    description: "Add two integers",
    inputSchema: z.object({ a: z.number().int(), b: z.number().int() }),
  },
  async ({ a, b }) => ({ content: [{ type: "text", text: String(a + b) }] }),
);

const transport = new StdioServerTransport();
await server.connect(transport);
```

Run via `tsx server.ts` and register the same way as the Python version.

## Project Layout (multi-tool server)

```text
acme-mcp/
├── pyproject.toml           # or package.json
├── src/
│   └── acme_mcp/
│       ├── __init__.py
│       ├── server.py        # FastMCP() + tool registrations
│       ├── tools/
│       │   ├── search.py
│       │   └── write.py
│       └── resources/
│           └── config.py
└── tests/
    └── test_tools.py
```

Tools should be:

- **Pure functions where possible** — easier to unit-test.
- **Strictly typed** — Pydantic / Zod schemas double as tool metadata for the agent.
- **Idempotent** when they mutate — agents may retry on transient failures.

## Transport Choice

| Transport | Use when |
|-----------|----------|
| stdio | Local subprocess; simplest for `.gemini/settings.json` `command:` form |
| HTTP / SSE | Multi-client; remote hosting; pairs with `httpUrl:` in settings |
| Streamable HTTP | Modern HTTP transport with bidirectional streaming |

FastMCP: `mcp.run(transport="stdio" | "http" | "sse")`. SDK: pick the matching `*ServerTransport` class.

## Testing

```bash
# Quick interactive client (Anthropic).
npx -y @modelcontextprotocol/inspector ./server.py

# Or programmatically (Python).
from fastmcp.testing import Client
async with Client(mcp) as client:
    result = await client.call_tool("add", {"a": 1, "b": 2})
    assert result.content[0].text == "3"
```

## Authentication

For HTTP transport with auth:

- **Bearer tokens** — pass via `Authorization` header in MCP client config.
- **Google credentials** — set `authProviderType: "google_credentials"` on the client side; the server validates `Authorization: Bearer <id-token>` and checks the audience.
- **API keys** — pass via custom header (e.g. `X-Api-Key`) in MCP client config.

## Publishing

For internal use, push the server to GitHub and reference via `npx -y github:org/repo` (TS) or pin a Python wheel built with `uv build`. For public, register on [mcpservers.org](https://mcpservers.org) and add a [`server.json`](https://github.com/modelcontextprotocol/registry) for the MCP registry.

## Important Notes

1. **Prefer Anthropic's `mcp-builder` skill for the full authoring workflow** — this skill is a quickstart, not a substitute.
2. **Tool descriptions are loaded eagerly** — keep them tight. Verbose descriptions × many tools = context bloat.
3. **Validate inputs server-side** even if the client schema does — agents send odd payloads on retry.
4. **Don't expose write tools without explicit confirmation** — design for least-privilege and explicit user consent.
5. **stdio = local subprocess; HTTP = remote service** — pick once, don't try to support both transports unless required.

## Documentation

- [Anthropic `mcp-builder` skill (canonical)](https://mcpservers.org/agent-skills/anthropic/mcp-builder)
- [Model Context Protocol spec](https://modelcontextprotocol.io)
- [FastMCP (Python)](https://github.com/jlowin/fastmcp)
- [FastMCP docs](https://gofastmcp.com)
- [`@modelcontextprotocol/sdk` (TypeScript)](https://github.com/modelcontextprotocol/typescript-sdk)
- [MCP Inspector (testing)](https://github.com/modelcontextprotocol/inspector)
- [MCP Registry](https://github.com/modelcontextprotocol/registry)
- Companion skills: `create-python-script` (PEP 723 form), `configure-gemini-cli` (registering MCP servers).
