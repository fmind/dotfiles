---
name: install-gemini-skills
description: Install Gemini API skills bundle (Gemini API, Live API, Interactions API, Vertex AI Gen AI SDK). Use for Gemini API projects.
---

# Install Gemini Skills

Google publishes the official [`google-gemini/gemini-skills`](https://github.com/google-gemini/gemini-skills) bundle. It's the canonical source for Gemini API and Vertex AI Gen AI SDK guidance; this skill explains when and how to install it.

## When to Trigger

- The repo imports `google-genai` (Python) or `@google/genai` (TS/JS), or calls Vertex AI Gen AI endpoints.
- The user mentions Gemini API, Live API, Interactions API, function calling, structured output, multimodal prompting, caching, batch prediction, grounding with Google Search, or Vertex AI Gen AI.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E 'gemini-api|gemini-live|gemini-interactions|vertex-ai-api'`. If installed, skip.

## Install

```bash
# List available skills.
npx skills add google-gemini/gemini-skills --list

# Install one (repeat per skill — project scope by default).
npx skills add google-gemini/gemini-skills --skill gemini-api-dev --project
npx skills add google-gemini/gemini-skills --skill gemini-live-api-dev --project
npx skills add google-gemini/gemini-skills --skill gemini-interactions-api --project
npx skills add google-gemini/gemini-skills --skill vertex-ai-api-dev --project
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## What Gets Installed

4 skills at the time of writing:

- **gemini-api-dev** — Prompt routing, multimodal prompting, function calling, structured outputs.
- **gemini-live-api-dev** — WebSocket streaming, voice activity detection, real-time audio/video/text.
- **gemini-interactions-api** — Unified Interactions API: stateful chat, Deep Research agents, server-side tools (Search, Maps, code exec).
- **vertex-ai-api-dev** — Vertex AI Gen AI SDK: tools, caching, batch prediction on Google Cloud.

## SDK Discipline

The skills explicitly target the **Google GenAI SDK** family:

- Python: `google-genai` (NOT the legacy `google-generativeai`).
- TypeScript/JS: `@google/genai` (NOT `@google/generative-ai`).
- Go / Java / Kotlin: GenAI SDKs.

Avoid `generationConfig`, `GenerativeModel`, and other deprecated naming. The bundled skills steer towards current idioms automatically.

## After Install

1. Restart the agent so the new skill descriptions are picked up by progressive disclosure.
2. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.
3. Pair with the Gemini docs MCP server (`https://gemini-api-docs-mcp.dev/`) for real-time docs lookups via `search_docs`.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate Gemini API docs. Defer to the installed skills and the docs MCP.
2. The Live API skill assumes WebSocket support in the runtime; don't apply it to plain HTTP-only environments.
3. The Vertex AI skill assumes a configured GCP project and `gcloud auth application-default login`.
4. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [Gemini API docs](https://ai.google.dev/gemini-api/docs)
- [Coding agents setup (MCP + Skills)](https://ai.google.dev/gemini-api/docs/coding-agents)
- [`google-gemini/gemini-skills` repo](https://github.com/google-gemini/gemini-skills)
- [Live API](https://ai.google.dev/gemini-api/docs/live-api)
- [Interactions API](https://ai.google.dev/gemini-api/docs/interactions)
- [Function calling](https://ai.google.dev/gemini-api/docs/function-calling)
- [Structured output](https://ai.google.dev/gemini-api/docs/structured-output)
