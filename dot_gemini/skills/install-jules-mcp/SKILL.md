---
name: install-jules-mcp
description: NOT AVAILABLE — Jules has no CLI-based MCP server. MCP for Jules is configured through the Jules web UI Settings instead. This skill documents what to do.
---

# Install Jules MCP

> **Heads-up:** As of 2026, the Jules CLI (`@google/jules`, "Jules Tools") does **not** ship an `mcp` subcommand. Earlier guides that referenced `jules mcp` are inaccurate.
>
> Jules' MCP integration is wired up through the Jules **web Settings UI** (Linear, Stitch, Neon, Tinybird, Context7, Supabase connectors are exposed there). There is currently no `.gemini/settings.json` configuration that brings Jules into a Gemini CLI session as an MCP tool.

## What to Use Instead

Pick one of:

- **Run Jules sessions from the CLI**, then pull their results into your local working tree:

  ```bash
  jules remote new --prompt "Add tests for the parser module"
  jules remote list
  jules remote pull <id>
  ```

- **Use the Jules extension for Gemini CLI** (`gemini extensions install jules`) for an in-CLI dispatch UX. This is the closest substitute for an MCP integration.

- **Install the Jules skills bundle** (`install-jules-skills`) so the local coding agent learns to dispatch Jules sessions effectively.

If Google publishes a Jules MCP server in the future, this skill should be rewritten to point at it. Track the [Jules CLI release notes](https://jules.google/docs/cli/reference/) for updates.

## Companion Agent

The `jules` subagent (`~/.gemini/agents/jules.md`) drives the Jules CLI. Keep it for ad-hoc cross-project work.

## Documentation

- [Jules home](https://jules.google)
- [Jules Tools CLI reference](https://jules.google/docs/cli/reference/)
- [Jules extension for Gemini CLI](https://developers.googleblog.com/en/introducing-the-jules-extension-for-gemini-cli/)
- [Jules MCP integration via web UI (changelog)](https://jules.google/docs/changelog/2026-02-02/)
- [`google-labs-code/jules-skills`](https://github.com/google-labs-code/jules-skills)
