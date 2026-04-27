---
name: install-adk-skills
description: Install Google's Agent Development Kit (ADK) skills via the google-agents-cli setup so the agent gains expert knowledge for scaffolding, developing, evaluating, observing, and deploying ADK agents.
---

# Install ADK Skills

Google's [Agent Development Kit (ADK)](https://adk.dev) skills are bundled inside the [`google-agents-cli`](https://google.github.io/agents-cli/) package — there is no standalone skills repo. A single setup command installs both the CLI and all 6 ADK skills.

## When to Trigger

- The repo contains ADK Python imports (`google.adk.*`), Java / Go / TS ADK packages, agent definitions in `agent.py` / `agent.go` / `agent.ts`, or `adk.yaml`-style config.
- The user mentions Agent Development Kit, ADK, multi-agent systems, agent evaluation suites, Agent Runtime deployment, or sequential / loop / parallel workflow agents.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -i adk`. If installed, skip.

## Install

```bash
# One command installs the CLI and all 6 ADK skills.
uvx google-agents-cli setup
```

The CLI auto-detects active coding agents (Claude Code, Gemini CLI, Cursor) and writes skills to `.agents/skills/` for cross-tool portability.

## What Gets Installed

7 skills at the time of writing:

- **google-agents-cli-workflow** — Always-active. Full lifecycle (scaffold → build → eval → deploy → publish → observe), code-preservation rules, model selection.
- **google-agents-cli-adk-code** — ADK Python API: agent types, tool definitions, orchestration patterns, callbacks, state management.
- **google-agents-cli-scaffold** — `scaffold` commands, template options, prototype-first workflow.
- **google-agents-cli-eval** — Evaluation metrics, evalset schema, LLM-as-judge, tool trajectory scoring.
- **google-agents-cli-deploy** — Deployment workflows (Agent Runtime, Cloud Run, GKE), service accounts, rollback.
- **google-agents-cli-publish** — Gemini Enterprise registration modes, programmatic usage, deployment metadata.
- **google-agents-cli-observability** — Cloud Trace, Cloud Logging, BigQuery analytics, AgentOps / Phoenix.

## Supported Languages

- **Python** — Primary, full feature parity.
- **Java** — Stable.
- **Go** — Stable.
- **TypeScript / JavaScript** — Supported.

## Related: ADK docs MCP

Pair the skills with the ADK docs MCP server for real-time docs lookups:

```bash
# Gemini CLI registers via the agent definition; for Claude Code:
claude mcp add adk-docs --transport stdio -- \
  uvx --from mcpdoc mcpdoc \
  --urls AgentDevelopmentKit:https://adk.dev/llms.txt \
  --transport stdio
```

## After Install

1. Restart the agent so the new skill descriptions are picked up by progressive disclosure.
2. `-scaffold` is the entry point for new agents; `-adk-code` covers ongoing iteration; `-eval` and `-observability` plug in once the agent is non-trivial; `-deploy` ships it; `-publish` registers with Gemini Enterprise. `-workflow` stays loaded throughout.
3. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs (`~/.gemini/skills/`) are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate ADK docs. Defer ADK-specific guidance to the installed skills and the docs MCP.
2. Skills come from `google-agents-cli`, not from `npx skills` — don't try to install via `npx skills add`.
3. Don't confuse `google-agents-cli` (ADK toolchain) with [Jules](https://jules.google) — different products, different skill bundles. See `install-jules-skills` for Jules.
4. `uvx` is provided by `uv` (already on the system via mise).

## Documentation

- [Agent Development Kit](https://adk.dev)
- [Coding with AI tutorial](https://adk.dev/tutorials/coding-with-ai/)
- [`google-agents-cli` docs](https://google.github.io/agents-cli/)
- [Bundled skills reference](https://google.github.io/agents-cli/reference/skills/)
- [GitHub: google/agents-cli](https://github.com/google/agents-cli)
- [LLMs index (MCP source)](https://adk.dev/llms.txt)
- [Full LLMs corpus](https://adk.dev/llms-full.txt)
