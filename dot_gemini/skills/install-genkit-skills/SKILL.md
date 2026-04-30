---
name: install-genkit-skills
description: Install Genkit skills bundle (AI flows, prompts, tools, RAG, agentic patterns) for JS/TS, Go, Python, Dart. Use for Genkit projects.
---

# Install Genkit Skills

The Genkit team publishes the official [`genkit-ai/skills`](https://github.com/genkit-ai/skills) bundle. It's the canonical source for Genkit-specific guidance across supported languages; this skill explains when and how to install it.

## When to Trigger

- The repo contains `genkit.config.ts`, `genkit.config.js`, imports from `@genkit-ai/*`, `genkit/*`, `genkit-go`, the Genkit Python SDK, or Genkit Dart packages.
- The user mentions Genkit, Dotprompt (`.prompt`) files, Genkit flows / tools / evaluators, or the Genkit Developer UI.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -i genkit`. If installed, skip.

## Install

```bash
# List available skills.
npx skills add genkit-ai/skills --list

# Install one (repeat per language — project scope by default).
npx skills add genkit-ai/skills --skill developing-genkit-js
npx skills add genkit-ai/skills --skill developing-genkit-go
npx skills add genkit-ai/skills --skill developing-genkit-dart
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## What Gets Installed

3 skills at the time of writing — one per supported language:

- **developing-genkit-js** — Node.js / TypeScript (Stable).
- **developing-genkit-go** — Go (Stable).
- **developing-genkit-dart** — Dart (Preview).

Each skill covers flows, prompts, tools (Zod-typed structured I/O), RAG primitives (indexers, embedders, retrievers, vector stores), agentic patterns (tool calling, interrupts, sessions), evaluators, and OTel-style observability. Genkit Python (Preview) is documented at [genkit.dev/docs/python/](https://genkit.dev/docs/python/) but does not yet ship a dedicated skill in this bundle.

## Related: `genkit` CLI and MCP

The skills assume the Genkit CLI is reachable (installed via `npm install -g genkit-cli` or invoked as `npx -y genkit-cli`):

```bash
# Developer UI for visual debugging.
genkit start -- <your-app-entrypoint>

# MCP server (ships inside genkit-cli).
genkit mcp
```

> Don't confuse `genkit-cli` (the dev tool / MCP server) with `@genkit-ai/mcp` or `genkitx-mcp` — the latter are MCP **client/host** plugins for use *inside* Genkit apps.

## After Install

1. Restart the agent so the new skill descriptions are picked up by progressive disclosure.
2. Install only the language(s) the project actually uses — no need to pull all four.
3. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate Genkit docs. Defer to the installed skills.
2. Dart support is in **Preview** — APIs may still shift; cross-check with the docs.
3. Firebase's `firebase/agent-skills` bundle ships `developing-genkit-js`, `-go`, `-python`, and `-dart` (the same JS/Go/Dart slugs plus a Python one). If you're already installing Firebase skills, these overlap — pick one source.
4. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [Genkit](https://genkit.dev)
- [`genkit-ai/skills` repo](https://github.com/genkit-ai/skills)
- [Genkit MCP server](https://genkit.dev/docs/mcp-server/)
- [Get started](https://genkit.dev/docs/get-started/)
- [RAG (JS)](https://genkit.dev/docs/js/rag/)
- [Agentic patterns (JS)](https://genkit.dev/docs/js/agentic-patterns/)
- [AI-assisted development (JS)](https://genkit.dev/docs/js/develop-with-ai/)
- [Firebase Genkit docs](https://firebase.google.com/docs/genkit)
