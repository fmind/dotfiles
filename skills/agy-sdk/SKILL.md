---
name: agy-sdk
description: "Orchestrate multi-agent executions with the google-antigravity Python SDK: subagents, policies, token budgets, hooks, and MCP over Gemini. Use when building an Antigravity SDK agent."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/agy-sdk
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Antigravity SDK

`google-antigravity` embeds the Antigravity harness as a Python library: it owns the agent loop, tool execution, and subagent delegation, so orchestration is expressed as configuration rather than a hand-rolled loop. The interactive CLI (`agy`) is a separate product on separate credentials — see §1. Project conventions come from [python-stack](../python-stack/SKILL.md); [google-adk](../google-adk/SKILL.md) stays the default agent framework, and this SDK is the choice when the harness itself (sandboxed tools, policies, budgets) is the value.

## 1. Install and Authenticate

```bash
uv add google-antigravity                              # ships a ~126 MB harness binary; a git clone alone will not run
export GEMINI_API_KEY="<aistudio-key>"                 # Gemini Developer API, pay-as-you-go with a free tier
gcloud auth application-default login                  # Vertex path instead: LocalAgentConfig(vertex=True, project=..., location=...)
```

**Billing is the Gemini API, never the Antigravity subscription.** The SDK reads only `GEMINI_API_KEY` (or `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION` with Vertex ADC); it never touches the OAuth login the `agy` CLI and IDE write to `~/.gemini`, so a Google AI Pro/Ultra plan grants it nothing. Without a key it fails closed at connect time with `AntigravityValidationError: A Gemini API key is required.` Keep the key in the environment or a secret manager per [sops-secrets](../sops-secrets/SKILL.md); never inline it in `LocalAgentConfig(api_key=...)` in committed code.

## 2. Orchestrate

[`references/orchestrator.py`](references/orchestrator.py) is a runnable parent-plus-two-subagent fan-out; the pieces that matter:

- **Static subagents**: `types.SubagentConfig(name, description, system_instructions, tools)` in `LocalAgentConfig(subagents=[...])` gives each worker its own context window and instructions. Prefer these — a named role is reviewable, whereas dynamic self-cloning is not.
- **Dynamic subagents**: `types.CapabilitiesConfig(enable_subagents=True)` alone lets the parent clone itself on demand, inheriting its toolset. Use it only for open-ended decomposition.
- **Register tools twice**: any callable a subagent uses must appear in the parent's `tools=[...]` as well, or the subagent starts without it.
- **Bound the fan-out**: `max_subagent_depth` caps nesting and `allowed_subagents` pins the roster; both belong in `CapabilitiesConfig`.
- **Typed results**: `response_schema=<pydantic model>` plus `await response.structured_output()` returns a validated `dict`, which is what makes a subagent's output safe to route programmatically.
- **Resume**: `conversation_id` with `session_continuation_mode` (`CREATE_ONLY`, `CREATE_OR_RESUME`, `RESUME`) and `save_dir` persists a long orchestration across processes.

## 3. Bound the Run

Two independent limits, and confusing them is how an unattended run burns a quota:

- **Policies decide _which_ tools run**: `policy.allow`, `deny`, `ask_user`, `workspace_only`, `allow_all`, `deny_all`. Custom Python functions and read-only built-ins are permitted by default; `run_command` and the write tools are not.
- **Budgets decide _how much_ runs**: `types.BudgetConfig(max_model_calls, max_tool_calls, max_input_tokens, max_output_tokens, max_total_tokens)`. This is the only hard stop on a delegating parent, so always set one.
- **Observe the cost**: `response.usage_metadata.total_token_count` per turn, and `hooks.pre_turn` / `post_turn` / `pre_tool_call_decide` / `post_tool_call` / `on_tool_error` / `on_session_end` for auditing per [observability](../observability/SKILL.md).

## 4. Extend

- **Tools**: plain functions with type hints and a docstring, passed to `tools=[...]`; filter built-ins with `CapabilitiesConfig(enabled_tools=...)` or `disabled_tools=...`.
- **Skills**: `skills_paths=["~/.agents/skills"]` loads this repository's `SKILL.md` catalog straight into an SDK agent, accepting either one skill directory or a parent of many.
- **MCP**: `mcp_servers=[types.McpStdioServer(name=..., command=..., args=[...], env={...})]` or `McpStreamableHttpServer`; server selection and host wiring live in [agent-mcp](../agent-mcp/SKILL.md).
- **Triggers**: `triggers.every(seconds, callback)` and `triggers.on_file_change(...)` drive background work without an external scheduler.

## Gotchas

- **`BuiltinTools.read_only()` omits `START_SUBAGENT`**: passing it verbatim as `enabled_tools` silently disables delegation, and the config then rejects `max_subagent_depth` at validation. Append `types.BuiltinTools.START_SUBAGENT` explicitly.
- **`policy.safe_defaults(handler)` takes a required handler** and routes every write to a human, so it blocks forever in an unattended run; compose explicit `allow`/`deny`/`workspace_only` policies for automation and reserve it for interactive tools.
- **Policies do not gate custom tools**: the engine only sees built-ins, so a custom function is executed as written — keep destructive work out of it.
- **Preview surface**: the SDK is pre-1.0 (`0.1.x`) and the published docs already lag the shipped API; verify signatures with `python -c "import inspect, google.antigravity"` before coding against a doc snippet.
- **Two products, one name**: this SDK runs the harness locally, while the `antigravity-preview-*` agent on the Gemini Interactions API runs in a Google-hosted sandbox and is billed and configured separately.

## Official Skills

Upstream: `Google-Antigravity/antigravity-sdk-python`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add Google-Antigravity/antigravity-sdk-python --list
skills add Google-Antigravity/antigravity-sdk-python --skill <name> -y
```

## Documentation

- [SDK overview](https://antigravity.google/docs/sdk/overview/) · [Subagents](https://antigravity.google/docs/sdk/subagents/) · [Policies](https://antigravity.google/docs/sdk/policies/) · [Lifecycle](https://antigravity.google/docs/sdk/lifecycle/)
- [antigravity-sdk-python](https://github.com/Google-Antigravity/antigravity-sdk-python) · [antigravity-sdk-python skills](https://github.com/google-antigravity/antigravity-sdk-python/tree/main/skills) · [Gemini API pricing](https://ai.google.dev/gemini-api/docs/pricing)
- Companion skills: [python-stack](../python-stack/SKILL.md), [google-adk](../google-adk/SKILL.md), [agent-mcp](../agent-mcp/SKILL.md), [prompt-design](../prompt-design/SKILL.md), [agent-evaluation](../agent-evaluation/SKILL.md), [observability](../observability/SKILL.md).
