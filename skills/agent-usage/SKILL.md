---
name: agent-usage
description: Query, analyze, and track LLM token usage across AI agent harnesses in ~/.agents/usages using the dot CLI, DuckDB, or jq. Use when auditing token consumption, comparing harness efficiency, or inspecting agent costs.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/agent-usage
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Agent Usage

Token usage for all AI harnesses (`agy`, `claude`, `codex`, `copilot`, `grok`, `opencode`) is recorded to `~/.agents/usages/<harness>/<session_id>.json` on session completion and turn boundaries.

## Directory Layout & Schema

Every session record is stored atomically with permissions `0o600`:

```text
~/.agents/usages/
├── agy/
│   └── <session_id>.json
├── claude/
│   └── <session_id>.json
├── codex/
│   └── <session_id>.json
├── copilot/
│   └── <session_id>.json
├── grok/
│   └── <session_id>.json
└── opencode/
    └── <session_id>.json
```

Each record contains:

```json
{
  "timestamp": "2026-09-03T18:00:00Z",
  "harness": "claude",
  "agent": "claude",
  "session_id": "abc-123",
  "model": "claude-opus-5",
  "input_tokens": 12500,
  "output_tokens": 3400,
  "cached_tokens": 82000,
  "cache_write_tokens": 1200,
  "reasoning_tokens": 0,
  "total_tokens": 99100,
  "cost_usd": 0.1425,
  "turn_count": 8,
  "cwd": "~/project"
}
```

## Commands

```bash
dot agent usage stats                                              # summary table of token usage per harness
dot agent usage stats --by-model                                   # break down token usage by harness and model
dot agent usage stats --harness claude                             # filter stats to a specific harness
dot agent usage stats --since 24h --json                           # emit json array for scripting
dot agent usage list -n 20                                         # list recent session records
dot agent usage show claude <session_id>                           # inspect a specific session record
dot agent usage sync                                               # scan raw stores and backfill missing records
duckdb -c "SELECT harness, count(*), sum(total_tokens) FROM read_json_auto('~/.agents/usages/*/*.json', union_by_name=true) GROUP BY harness" # ad-hoc SQL
```

## Workflow

1. **Check aggregate usage**: run `dot agent usage stats` to inspect overall token consumption and costs across harnesses.
1. **Break down by model**: run `dot agent usage stats --by-model` when comparing prompt efficiency or model tiers.
1. **Filter by time window**: pass `--since 24h` or `--since 7d` to isolate a recent sprint or experiment.
1. **Deep dive with DuckDB**: query `~/.agents/usages/*/*.json` directly with `read_json_auto` in `duckdb` for custom aggregations, percentiles, or CSV exports.
1. **Synchronize historical sessions**: run `dot agent usage sync` after adding an integration to extract metrics from historical transcripts.

## Gotchas

- **Only Claude and OpenCode report cost**: `cost_usd` comes from Claude's `cost-state` transcript lines and OpenCode's session database; `agy`, `codex`, `copilot`, and `grok` expose no price, so they always total `$0.00`. Compare those four on tokens, never on cost.
- **Grok counts context, not consumption**: Grok records no cumulative token totals, so its `input_tokens` carries the final context-window occupancy and its output tokens are unobservable — a documented undercount.
- **Atomic rewrites prevent duplicates**: each session uses a single `<session_id>.json` file overwritten on turn updates, preventing double-counting across `Stop` and `SessionEnd` hooks.
- **Both harness and agent fields exist**: queries can group by either `harness` or `agent` interchangeably.
- **`sync` fails loud, hooks fail soft**: `dot agent usage sync` aborts on an unreadable store rather than reporting `Synced 0`, and it rewrites every record it can re-derive — so it is the way to backfill after an extractor changes.
- **Background hooks fail soft**: hooks spool errors to `~/.agents/hook-failures` so a failure in usage tracking never aborts the agent CLI.

## Documentation

- [DuckDB JSON Functions](https://duckdb.org/docs/data/json/overview)
- Companion skills: [dot-cli](../dot-cli/SKILL.md) (every `dot` command), [duckdb](../duckdb/SKILL.md) (file-based SQL analysis).
