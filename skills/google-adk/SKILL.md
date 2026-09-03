---
name: google-adk
description: Build, run, evaluate, and deploy Google ADK agents in Go or Python with agents-cli, Agent Runtime, or Cloud Run and the official ADK skills. Use for any ADK work.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/google-adk
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Google ADK Agents

The Agent Development Kit is the default agent framework: code-first, model-agnostic, Gemini-optimized, with a launcher serving console, web UI, REST, and A2A from one binary. Go is the default; Python when `agents-cli` scaffolding, evaluation, or Agent Runtime matter more than a single static binary. Language APIs live in [go-stack](../go-stack/SKILL.md) §6 and [python-stack](../python-stack/SKILL.md) §5.

## 1. Choose the Path

- **Go service, CLI, or single binary**: `google.golang.org/adk/v2` with `full.NewLauncher()` per [go-stack](../go-stack/SKILL.md) §6.
- **Python with generated CI, eval, and deploy scaffold**: `agents-cli`, run as `uvx google-agents-cli` (`uvx` ships with `uv`; nothing to install) per §2.
- **Managed runtime (sessions, memory, registry)**: Agent Runtime (Gemini Enterprise Agent Platform, formerly Vertex AI) via `agents-cli deploy --deployment-target agent_runtime`.
- **Self-managed container**: Cloud Run via `agents-cli deploy --deployment-target cloud_run`, or the Go launcher's `web` mode per [cloud-run](../cloud-run/SKILL.md).

## 2. Python Workflow (agents-cli)

```bash
uvx google-agents-cli create <name> --agent adk --deployment-target cloud_run --region europe-west1 --agent-guidance-filename AGENTS.md   # name <= 26 chars, lowercase
cd <name> && uvx google-agents-cli install --locked   # dependencies
uvx google-agents-cli playground                      # local web UI with live reload
uvx google-agents-cli run "hello"                     # one prompt, then shutdown
uvx google-agents-cli eval run                        # dataset in tests/, LLM-as-judge metrics (approval: model cost)
uvx google-agents-cli deploy --dry-run                # review; real deploys need explicit approval
```

- Agent Runtime targets omit `--session-type`; other targets accept `--session-type agent_platform_sessions`; `--cicd-runner github_actions` generates the workflow.
- Keep the generated `app/` layout (`agent.py`, `fast_api_app.py`) and `agents-cli-manifest.yaml`; add the canonical `mise.toml` tasks around it and never hand-write A2A plumbing.
- Test wiring and scaffold normalization are in [python-stack](../python-stack/SKILL.md) §5.

## 3. Rules for Every Agent

- **Auth**: the platform with Application Default Credentials (`gcloud auth application-default login` locally, the runtime service account in production); an API key only for AI Studio prototypes, never committed.
- **Environment**: `GOOGLE_GENAI_USE_VERTEXAI=true`, `GOOGLE_CLOUD_PROJECT=<project_id>`, `GOOGLE_CLOUD_LOCATION=<region>` (e.g. `europe-west1` or `global`) in `.env`.
- **Model pins**: name the current Flash or Pro generation (the scaffold's default is a good anchor) and record the choice in `AGENTS.md`; pin a dated snapshot only for strict reproducibility.
- **Never `-latest`**: its platform resolution is version-ambiguous and hot-swaps model quality under a stable name.
- **Tools are typed**: schema-documented functions with narrow permissions; destructive tools require confirmation.
- **Evaluate before deploy**: a dataset with expected trajectories and responses per [agent-evaluation](../agent-evaluation/SKILL.md), candidate against baseline.
- **Observability**: OpenTelemetry traces (`OTEL_EXPORTER_OTLP_ENDPOINT` or `--otel_to_cloud`) and structured logs without prompt bodies that may carry personal data.
- **Secrets and safety**: Secret Manager references; Model Armor or equivalent input and output filters for public agents per [threat-model](../threat-model/SKILL.md).

## Gotchas

- **Upstream code skill is Python-only**: for Go, `go doc` and the examples in `~/go/pkg/mod/google.golang.org/adk/v2` are the reference.
- **Runtime skills are not coding skills**: the [ADK skills page](https://adk.dev/skills/) describes `SkillToolset`, skills an agent loads for itself at run time.
- **Cloud skills live elsewhere**: Cloud Run and CLI guardrail skills come from [google-cloud](../google-cloud/SKILL.md).

## Official Skills

Upstream: `google/agents-cli`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add google/agents-cli --list
skills add google/agents-cli --skill <name> -y
```

## Documentation

- [ADK](https://google.github.io/adk-docs/) · [google/agents-cli](https://github.com/google/agents-cli) · [Agent Runtime](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale)
- Companion skills: [go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md), [cloud-run](../cloud-run/SKILL.md), [google-cloud](../google-cloud/SKILL.md), [agent-mcp](../agent-mcp/SKILL.md) (tool servers), [prompt-design](../prompt-design/SKILL.md).
