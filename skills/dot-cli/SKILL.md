---
name: dot-cli
description: Use the dot CLI (fmind/dotfiles) to verify the environment, pull repos, manage the local k3d cluster, log agent sessions, and prune caches. Use when running or scripting any dot command.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dot-cli
  created: 2026-07-31
  updated: 2026-07-31
---

# Dot CLI

`dot` is the unified CLI of `fmind/dotfiles` (source in `dot/`, built to `~/.local/bin/dot`). Every command has a one-letter alias, and `dot <command> --help` is always the authoritative flag reference — this skill covers what help output cannot tell you.

## Commands

| Command             | Alias | Purpose                                                                            |
| ------------------- | ----- | ---------------------------------------------------------------------------------- |
| `dot verify`        | `v`   | Sanity-check environment, tools, secrets, install freshness (`--json`, `--fix`)    |
| `dot status`        | `s`   | Unified git + docker + k3d summary (`--json`)                                      |
| `dot pull`          | `p`   | Pull every repo in `~/.config/dot.yaml` concurrently (`--push`)                    |
| `dot commit`        | `c`   | Write a Conventional Commit from the staged diff via AI                            |
| `dot pull-request`  | `pr`  | Draft a PR description via AI, then `gh pr create`                                 |
| `dot release`       | `r`   | Bump version, changelog, tag, publish (see [dot-release](../dot-release/SKILL.md)) |
| `dot cluster`       | `k`   | `start` / `stop` / `status` / `delete` / `namespace` the local k3d cluster         |
| `dot prune`         | `x`   | Reclaim disk from agent session logs and caches (see below)                        |
| `dot agent session` | `a s` | Atomically log, sync, and migrate private agent-session lineage (gathers only)     |
| `dot notify`        | `n`   | Desktop notification for agent hooks or custom alerts                              |
| `dot chezmoi clean` | `m c` | Delete `$HOME` orphans left by files chezmoi no longer manages                     |
| `dot config`        | `f`   | `show` / `path` / `init` / `edit` / `validate` `~/.config/dot.yaml`                |
| `dot login`         | `l`   | OAuth wrappers: `github`, `workspace`, `gcp`, `clasp`                              |
| `dot setup`         | `u`   | Enable GCP APIs for the active Workspace project                                   |
| `dot completion`    | `g`   | Regenerate fish completions for `dot` and external tools                           |
| `dot version`       | `i`   | Version plus the embedded VCS revision                                             |

Global flags: `--config/-c <path>` (or `DOT_CONFIG_PATH`) and `--verbose` (or `DOT_VERBOSE`).

## Session ingestion

Live hooks and `dot agent session sync` write the same append-only store under `~/.agents/sessions/v1/`. A hashed `(agent, session_id)` directory contains immutable source generations; each generation is keyed by the source fingerprint and parser version, holds `manifest.json` plus `transcript.jsonl`, and appears only after both owner-only files validate and their temporary directory is atomically renamed. Outcomes report a truncated lineage hash, record counts, malformed/skipped counts, and completeness without printing the raw session ID.

Run `dot agent session migrate` before relying on the new store for historical evidence. It is a read-only dry run by default, deterministically selects the most complete legacy transcript for every lineage, reports duplicate/partial/skipped/malformed totals, and leaves every legacy file in place. `dot agent session migrate --apply` copies each selection into the versioned store without deleting the old archive.

## Prune

Targets are flags, so they compose; each accepts an optional depth, and nothing runs unless a target is named.

```bash
dot prune                      # lists the targets, deletes nothing
dot prune --agents --dry-run   # report what would go, delete nothing
dot prune -a -g -m             # session logs + Go caches + mise, at their configured depth
dot prune --docker=system      # deeper: also stopped containers, networks, dangling images
dot prune --all                # every target, default depth
dot prune --all=deep           # every target, deepest depth
```

| Target | Flag | Default depth                      | Deeper depth                           |
| ------ | ---- | ---------------------------------- | -------------------------------------- |
| agents | `-a` | expired session logs (per store)   | —                                      |
| docker | `-d` | build cache                        | `=system` containers, networks, images |
| go     | `-g` | build + test caches                | `=module` module cache                 |
| python | `-p` | `uv cache prune`                   | `=all` uv wipe + `pip cache purge`     |
| node   | `-n` | npx cache                          | `=all` npm cache                       |
| mise   | `-m` | versions, cache, downloads         | `=configs` untracked config links      |
| tools  | `-t` | Trivy, Helm, dprint, golangci-lint | —                                      |

- **`--docker=system` deletes stopped local k3d clusters** — stopped k3d containers _are_ the cluster. Use the default depth while a cluster matters.
- Every target has a config section: `level` is the depth a bare flag (and `--all`) selects, `paths` are the directories it removes, and `agents.sessions` carries the retention of each session store. `--days N` overrides every store for one run; `--days 0` empties them.

```yaml
prune:
  agents:
    sessions:
      - path: ~/.claude/projects
        keep_days: 7
      - path: ~/.agents/sessions
        keep_days: 30 # 0 empties the store, whatever the file age
    keep: [memory, memory.jsonl, MEMORY.md] # never pruned, however old
  docker:
    level: build # set to system on a machine with no local k3d cluster
  node:
    level: cache
    paths: [~/.npm/_npx]
```

- `dot prune --agents` is the only command that deletes session logs; `dot agent session` just gathers them.
- Agent long-term memory (`memory/`, `memory.jsonl`, `MEMORY.md`) is never pruned, however old.
- Missing tools and a stopped Docker daemon are reported as skipped, not as failures; a failing target never stops the others.

## Tasks

Inside the dotfiles repo, prefer the `mise run` wrappers: they depend on `build`, so they always run the current source rather than a stale `~/.local/bin/dot`.

```bash
mise run prune               # dot prune --all=deep
mise run prune:agents        # session logs and caches, keeping k3d and the Go cache
mise run verify              # dot verify
mise run release -- -y       # dot release, non-interactive
```

Project-management tasks are aliased with an `m` prefix (`mp`, `mpa`, `mx`, `mr`); the bare letters belong to the common vocabulary (`f` format, `c` check, `t` test, `b` build). `mr` is _not_ shorthand for `mise run` in a script — that is an interactive-only fish abbreviation.

## Config

All commands read `~/.config/dot.yaml`; `dot config init` scaffolds it with the built-in defaults and `dot config validate` checks it (unknown keys are rejected). A malformed file is fatal for every command except the `config` group, so the file stays repairable.

## Gotchas

- Rebuild after changing `dot/`: `mise run build` (or `mise run apply` to build and apply).
- The local k3d cluster must stay off by default — `dot cluster stop local` as soon as the task is done.
- `dot commit` and `dot pull-request` call an AI CLI, so they need it installed and authenticated.

## See Also

- [dot-release](../dot-release/SKILL.md) — Cutting a dotfiles release.
- [chezmoi](../chezmoi/SKILL.md) — Source naming, templates, and the edit-then-apply loop.
- [k8s-local](../k8s-local/SKILL.md) — Working inside the shared local cluster.
