---
name: use-genkit-cli
description: Guide for using the genkit CLI — Developer UI, init, flow execution, evaluators, MCP server, and project commands across JS/TS, Go, Python, Dart.
---

# Use Genkit CLI

`genkit` (from the `genkit-cli` npm package) is the dev tool for [Genkit](https://genkit.dev). It launches a local Developer UI for visual debugging, runs flows, manages evaluators, and (since 2026) ships an MCP server.

The skill bundle `genkit-ai/skills` (or `firebase/agent-skills`) covers per-language flow authoring; this skill focuses on the **CLI workflow**.

## Install / Bootstrap

```bash
# Global install (used here for most commands).
npm install -g genkit-cli

# Or invoke ad-hoc with npx.
npx genkit <command>
```

## Developer UI

```bash
# Start the Genkit Developer UI against your app entrypoint.
# JS/TS:
genkit start -- npx tsx src/index.ts

# Python:
genkit start -- python -m my_app

# Go (uses the runner harness):
genkit start -- go run ./cmd/server

# UI defaults to http://localhost:4000.
genkit start --port 4001 --noui    # API-only, no browser
```

The UI surfaces flows, prompts, traces, evals, and lets you invoke flows with arbitrary inputs.

## Flow Execution

```bash
# Run a flow by name with JSON input.
genkit flow:run myFlow '{"prompt": "summarize this"}'

# Run from a file.
genkit flow:run myFlow @input.json

# Batch-execute many inputs against a flow.
genkit flow:batchRun myFlow inputs.json
```

There is no `genkit flow:list` subcommand — use the Developer UI (`genkit start`) to see the registered flows.

## Prompts (Dotprompt files)

The genkit CLI does not have `prompt:*` subcommands; prompts are inspected and re-rendered through the Developer UI (`genkit start`) or via the language SDK in your application code.

## Evaluators

```bash
# Inference-based: run a flow over a dataset and score the outputs.
genkit eval:flow myFlow --input evalsets/qa.json

# Raw evaluation on a pre-extracted dataset (no inference).
genkit eval:run factsEvalDataset.json

# Extract a dataset from prior traces.
genkit eval:extractData --label nightly-2026-04 --output factsEvalDataset.json
```

`eval:flow` and `eval:run` accept `--evaluators <list>` and `--batchSize <N>`; `eval:flow` also accepts `--input <dataset|file>` and `--context <auth-context>`. There is no `eval:list` subcommand or `--evalset` / `--flow` flag.

## MCP Server (since 2026)

```bash
# Start the MCP server bundled with genkit-cli.
genkit mcp

# Or via npx (no global install needed).
npx genkit mcp
```

> Distinct from `@genkit-ai/mcp` and `genkitx-mcp`, which are MCP **client/host** plugins for Genkit apps. To install the MCP server in a project's `.gemini/settings.json`, use `install-genkit-mcp`.

## Telemetry / Logs

Genkit traces are inspected in the Developer UI (`genkit start`) under the **Traces** tab. There is no `trace:tail` / `trace:list` CLI subcommand — for production traces, ship them to Cloud Trace via the `@genkit-ai/google-cloud` plugin and inspect them in the GCP console.

## Project Layout (typical)

```
my-genkit-app/
├── package.json
├── src/
│   ├── index.ts          # `import { genkit } from 'genkit'; ai = genkit({...})`
│   ├── flows/            # flows authored as `defineFlow`
│   └── prompts/
│       └── summarize.prompt
├── tools/                # custom tools (Zod-typed)
└── evalsets/
    └── qa.json
```

## Common Workflows

**Bootstrap a JS Genkit project.**
```bash
mkdir my-app && cd my-app
npm init -y
npm install genkit @genkit-ai/google-genai
# Scaffold IDE / coding-agent helpers (Cursor, Claude Code, Gemini CLI rules).
npx genkit init-ai-tools
genkit start -- npx tsx src/index.ts
```

**Iterate on a flow.**
1. Edit `src/flows/foo.ts`.
2. `genkit start` (auto-reloads).
3. Drive the flow from the UI or `genkit flow:run`.

**Run an eval set in CI.**
```bash
genkit eval:flow myFlow --input evalsets/qa.json --evaluators genkit/faithfulness,genkit/answer_relevance
```

## Important Notes

1. **`genkit-cli` is the dev tool, not a runtime dependency** — keep it out of production deps.
2. **The Developer UI hits your local app via the runner**; if the app crashes, the UI shows traces but flows fail. Read app logs alongside.
3. **MCP server invocation is `genkit mcp`** (not `genkit-cli mcp`) — the binary name is `genkit`.
4. **Python and Dart support is in Preview**; expect API surface changes. Check `genkit version` and the docs for the active language.

## Documentation

- [Genkit](https://genkit.dev)
- [Get started](https://genkit.dev/docs/get-started/)
- [Genkit MCP server](https://genkit.dev/docs/mcp-server/)
- [Developer UI guide](https://genkit.dev/docs/devtools/)
- [Evaluators](https://genkit.dev/docs/evaluation/)
- [Firebase Genkit docs](https://firebase.google.com/docs/genkit)
- [`genkit-cli` on npm](https://www.npmjs.com/package/genkit-cli)
