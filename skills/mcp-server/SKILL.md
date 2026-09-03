---
name: mcp-server
description: Author an MCP server in Go, Python, or TypeScript with the official SDKs over stdio or Streamable HTTP and test it with the MCP Inspector. Use when building an MCP server.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/mcp-server
  created: "2026-09-03"
  updated: "2026-09-03"
---

# MCP Server

Author a Model Context Protocol server that exposes typed tools, resources, and prompts to agents. Go with the official `github.com/modelcontextprotocol/go-sdk` is the default; Python and TypeScript use their official SDKs. Registering a finished server in a host belongs to [agent-mcp](../agent-mcp/SKILL.md).

## Workflow

1. **Pick the stack**: Go for a binary that ships by `go install` or `ko`; Python or TypeScript when the project already lives there.
   ```bash
   go get github.com/modelcontextprotocol/go-sdk@latest                      # package mcp; skip -pre tags
   uv add "mcp[cli]"                                                        # from mcp.server import MCPServer (formerly FastMCP)
   pnpm add @modelcontextprotocol/server @modelcontextprotocol/node         # McpServer + stdio; node adds Streamable HTTP
   ```
1. **Define typed tools**: one struct (or type-hinted function) per tool; the SDK derives the JSON Schema from struct tags or type hints, so the schema never drifts from the code. Go: `mcp.NewServer(&mcp.Implementation{Name: "<slug>", Version: "v0.1.0"}, nil)` then `mcp.AddTool(server, &mcp.Tool{Name: "...", Description: "..."}, handler)`.
1. **Choose the transport**: stdio for a local tool, Streamable HTTP for a hosted one, mounted as a plain `http.Handler`.
   ```go
   err := server.Run(ctx, &mcp.StdioTransport{}) // local
   handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true})
   err = http.ListenAndServe(":"+os.Getenv("PORT"), handler) // Cloud Run injects PORT
   ```
   Python: `uv run mcp run server.py --transport streamable-http`; TypeScript: `StdioServerTransport` from `@modelcontextprotocol/server/stdio` or the HTTP transport from `@modelcontextprotocol/node`.
1. **Test locally** with the Inspector before wiring any host; unit-test through the in-memory transport (`mcp.NewInMemoryTransports()` in Go, `Client(mcp)` in Python) so the suite needs no subprocess.
   ```bash
   npx @modelcontextprotocol/inspector ./bin/<slug>                            # stdio, web UI (Python: uv run mcp dev server.py)
   npx @modelcontextprotocol/inspector --cli ./bin/<slug> --method tools/list  # scripted
   npx @modelcontextprotocol/inspector --transport http --server-url http://localhost:8080/mcp
   ```
1. **Ship**: local servers install as a binary (`go install`, `uv tool install`); hosted servers containerize per [containerize](../containerize/SKILL.md) and deploy per [cloud-run](../cloud-run/SKILL.md), private by default with an identity token or OAuth in front.
1. **Register** the server in each host with [agent-mcp](../agent-mcp/SKILL.md) and verify one real tool call end to end.

## Gotchas

- **Stdout is the protocol on stdio**: log to stderr (`slog.New(slog.NewTextHandler(os.Stderr, nil))`); one stray `fmt.Println` corrupts the stream.
- **Stateless on Cloud Run**: instances scale to zero and requests are not pinned, so keep `Stateless: true` (or the SDK equivalent) and no in-memory sessions.
- **Tools are the attack surface**: validate every argument, bound file and network access to what the tool advertises, and treat tool inputs as untrusted.
- **Package names moved**: the TypeScript SDK split into `@modelcontextprotocol/server`, `client`, and `node`; `@modelcontextprotocol/sdk` is the legacy single package.

## Official Skills

Upstream: `anthropics/skills`. List the current release, then install the MCP builder skill at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add anthropics/skills --list
skills add anthropics/skills --skill <name> -y
```

## Documentation

- [MCP specification](https://modelcontextprotocol.io/specification) · [Go SDK](https://github.com/modelcontextprotocol/go-sdk) · [Python SDK](https://github.com/modelcontextprotocol/python-sdk) · [TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk) · [Inspector](https://github.com/modelcontextprotocol/inspector)
- Companion skills: [agent-mcp](../agent-mcp/SKILL.md) (host registration), [go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md), [cloud-run](../cloud-run/SKILL.md) (hosted deploy).
