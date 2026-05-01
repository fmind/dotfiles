---
name: use-google-agents-cli
description: Guide for using the Google Agents CLI (agents-cli) to scaffold, evaluate, deploy, and publish ADK agents on the Gemini Enterprise Agent Platform.
---

# Use Google Agents CLI (agents-cli)

This skill documents how to use the Google Agents CLI (`agents-cli`), a toolkit for building, evaluating, and deploying agents on the Gemini Enterprise Agent Platform using the [Agent Development Kit (ADK)](https://adk.dev/).

## How it Works

`agents-cli` is both a **CLI** and a **bundle of skills**. The CLI scaffolds projects, runs evaluations, deploys to Google Cloud, and registers with Gemini Enterprise. The skills (`google-agents-cli-workflow`, `-adk-code`, `-scaffold`, `-eval`, `-deploy`, `-publish`, `-observability`) teach the coding agent the ADK development lifecycle.

The CLI itself is installed as a global `pipx` tool via mise. Run `setup` once per machine to install the bundled skills into your active coding agents (Claude Code, Gemini CLI, Codex, ...).

## One-time Setup

```bash
# Install the bundled skills into every detected coding agent.
agents-cli setup

# Authenticate against Google Cloud or AI Studio.
agents-cli login
agents-cli login --status

# Force a reinstall of the skills (e.g. after upgrading the CLI).
agents-cli update
```

After `setup`, the seven `google-agents-cli-*` skills become available alongside this wrapper skill.

## Usage Pattern

```bash
agents-cli <command> [subcommand] [flags]
```

The recommended lifecycle is **scaffold → run → eval → deploy → publish → observe**, with each phase covered by a dedicated skill.

## Scaffolding a Project

```bash
# Create a new agent project from the default template.
agents-cli create my-agent

# Add deployment, CI/CD, or RAG to an existing project.
agents-cli scaffold enhance

# Upgrade a project to the latest agents-cli template version.
agents-cli scaffold upgrade
```

`agents-cli create <name>` is the top-level shortcut; `agents-cli scaffold create <name>` is the equivalent under the `scaffold` namespace. The `scaffold` command always requires a subcommand (`create`, `enhance`, `upgrade`).

## Developing Locally

```bash
# Install project dependencies (uv-based).
agents-cli install

# Run the agent against a single prompt.
agents-cli run "Summarise the latest release notes"

# Run code-quality checks (Ruff).
agents-cli lint

# Inspect project config and CLI version.
agents-cli cmd-info
```

## Evaluating

```bash
# Run all evalsets defined in the project.
agents-cli eval run

# Compare two eval result files (regression checks).
agents-cli eval compare baseline.json candidate.json
```

## Deploying & Publishing

```bash
# Deploy to the default target (Agent Runtime, Cloud Run, or GKE).
agents-cli deploy

# Provision a single-project Google Cloud setup.
agents-cli infra single-project

# Provision a CI/CD pipeline plus staging/prod infrastructure.
agents-cli infra setup-cicd

# Register the deployed agent with Gemini Enterprise.
agents-cli publish gemini-enterprise
```

## Data & RAG

```bash
# Provision datastore infrastructure (Vector Search, Agent Search, ...).
agents-cli infra datastore

# Run the data-ingestion pipeline.
agents-cli data-ingestion
```

## Bundled Skills

Once `agents-cli setup` has run, these 7 skills are available:

- **google-agents-cli-workflow** — Always-active. Full lifecycle, code-preservation rules, model selection.
- **google-agents-cli-adk-code** — ADK Python API: agents, tools, orchestration, callbacks, state.
- **google-agents-cli-scaffold** — `scaffold` commands, template options, prototype-first workflow.
- **google-agents-cli-eval** — Evalset schema, metrics, LLM-as-judge, tool trajectory scoring.
- **google-agents-cli-deploy** — Agent Runtime, Cloud Run, GKE, CI/CD, secrets.
- **google-agents-cli-publish** — Gemini Enterprise registration.
- **google-agents-cli-observability** — Cloud Trace, Cloud Logging, BigQuery analytics, AgentOps, Phoenix.

## Important Notes

1. **Prerequisites:** Python 3.11+, `uv`, and Node.js — all already provided by mise.
2. **Not a coding agent:** `agents-cli` is a tool *for* coding agents, not a replacement for Claude Code or Gemini CLI.
3. **Standalone usage:** Every CLI command works without a coding agent attached; the skills only make agent-driven workflows smoother.
4. **Pin versions for production.** Expect breaking changes between minor versions; run `agents-cli cmd-info` to confirm what's installed and check the docs for current lifecycle status before depending on it.

## Documentation

- [agents-cli docs](https://google.github.io/agents-cli/)
- [CLI reference](https://google.github.io/agents-cli/cli/)
- [Bundled skills reference](https://google.github.io/agents-cli/reference/skills/)
- [GitHub: google/agents-cli](https://github.com/google/agents-cli)
- [PyPI: google-agents-cli](https://pypi.org/project/google-agents-cli/)
- [Agent Development Kit (ADK)](https://adk.dev/)
- [Gemini Enterprise Agent Platform](https://docs.cloud.google.com/gemini-enterprise-agent-platform/overview)
