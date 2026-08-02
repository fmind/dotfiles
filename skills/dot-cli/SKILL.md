---
name: dot-cli
description: Use the dot CLI (fmind/dotfiles) to verify the environment, pull repos, manage the local k3d cluster, log agent sessions, and prune caches. Use when running or scripting any dot command.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dot-cli
  created: 2026-07-31
  updated: 2026-08-02
---

# Dot CLI

`dot` is the unified CLI of `fmind/dotfiles` (source in `dot/`, built to `~/.local/bin/dot`). All subcommands have single-letter aliases. Use `dot <command> --help` for precise flag usage.

## Commands

| Command             | Alias | Purpose                                                                                                   |
| ------------------- | ----- | --------------------------------------------------------------------------------------------------------- |
| `dot verify`        | `v`   | Sanity-check environment, tools, secrets, and install freshness (`--json`, `--fix`)                       |
| `dot status`        | `s`   | Unified git, docker, and k3d status summary (`--json`)                                                    |
| `dot pull`          | `p`   | Pull repos declared in `~/.config/dot.yaml` (`--push` to push)                                            |
| `dot commit`        | `c`   | Generate Conventional Commit via AI (`agy`) from staged changes                                           |
| `dot pull-request`  | `pr`  | Draft PR description via AI and invoke `gh pr create`                                                     |
| `dot release`       | `r`   | Prepare, tag, and publish dotfiles release (see [dot-release](../../.agents/skills/dot-release/SKILL.md)) |
| `dot cluster`       | `k`   | Manage local k3d cluster and gather diagnostic evidence                                                   |
| `dot prune`         | `x`   | Reclaim disk space from agent session logs and dev/tool caches                                            |
| `dot agent`         | `a`   | Agent session management (`doctor`, `hook`, `session`)                                                    |
| `dot notify`        | `n`   | Trigger desktop notifications for agent hooks or alerts                                                   |
| `dot chezmoi`       | `m`   | chezmoi configuration management (`clean` subcommand)                                                     |
| `dot chezmoi clean` | `m c` | Find and delete `$HOME` orphan files unmanaged by chezmoi                                                 |
| `dot config`        | `f`   | Scaffold (`init`), edit, show, and validate `~/.config/dot.yaml`                                          |
| `dot login`         | `l`   | OAuth login wrapper (`github`, `workspace`, `gcp`, `clasp`)                                               |
| `dot setup`         | `u`   | Enable GCP APIs on active Workspace project                                                               |
| `dot completion`    | `g`   | Generate fish autocompletions for `dot` and tools                                                         |
| `dot context`       | `t`   | Export bounded, secret-scanned project context pack                                                       |
| `dot version`       | `i`   | Print binary version and VCS revision                                                                     |

Global flags: `--config/-c <path>` (or `DOT_CONFIG_PATH`) and `--verbose` (or `DOT_VERBOSE`).

## Core Workflows

### Disk Pruning (`dot prune` / `x`)

Runs target-based cleanup using rules defined in `~/.config/dot.yaml`. Nothing is deleted unless targets are specified or `--all` is used.

- Basic usage: `dot prune -a -g -m` (prune agent session logs, Go cache, and mise cache).
- Deep pruning: `dot prune --all=deep` or `dot prune --docker=system` (includes containers, networks, dangling images).
- Safety: Agent long-term memory (`memory/`, `MEMORY.md`) is never pruned. Use `--dry-run` to preview deletions.

### Agent Management (`dot agent` / `a`)

Ingests and verifies cross-agent transcripts in `~/.agents/sessions/v1/`.

- `dot agent doctor` (`a doctor`): Check persona, hooks, skills, and store health. Use `--fix` (`-f`) to trigger a targeted chezmoi repair on hook mismatches.
- `dot agent session sync`: Ingest and validate agent transcripts into the append-only store.
- `dot agent session migrate`: Migrate legacy agent sessions into the versioned store (`--apply` to execute).

### Cluster Diagnostics (`dot cluster` / `k`)

- `dot cluster diagnose`: Export sanitized, owner-only diagnostic manifest to `~/.local/state/dot/diagnostics/`. Redacts secrets automatically.
- Example: `dot cluster diagnose --namespace app --output ./app-diagnostics.json --since 20m`.

## Gotchas

- **Source edits**: Always rebuild with `mise run build` (or `mise run apply`) after modifying files under `dot/`.
- **K3d cluster**: `--docker=system` during `dot prune` deletes stopped k3d clusters. Keep the k3d cluster stopped when inactive (`dot cluster stop`).
- **Hook stale install**: If `dot agent doctor` reports `command-unavailable`, run `mise run apply` to sync the compiled binary with newly deployed chezmoi hooks.

## See Also

- [dot-release](../../.agents/skills/dot-release/SKILL.md) — Cutting a dotfiles release.
- [chezmoi](../../.agents/skills/chezmoi/SKILL.md) — Source naming, templates, and the edit-then-apply loop.
- [k8s-local](../k8s-local/SKILL.md) — Working inside the shared local cluster.
