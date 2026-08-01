---
name: dot-cli
description: Use the dot CLI (fmind/dotfiles) to verify the environment, pull repos, manage the local k3d cluster, log agent sessions, and prune caches. Use when running or scripting any dot command.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dot-cli
  created: 2026-07-31
  updated: 2026-08-01
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
| `dot cluster`       | `k`   | Manage and diagnose the verified local k3d cluster                                 |
| `dot prune`         | `x`   | Reclaim disk from agent session logs and caches (see below)                        |
| `dot agent doctor`  | —     | Read-only cross-agent discovery, hook, source, and lineage health                  |
| `dot agent hook`    | —     | Observable session and notification hook boundary with bounded failure spooling    |
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

Agent hook templates invoke `dot agent hook session ...` and `dot agent hook notify ...`. The wrapper preserves the original non-zero exit and writes bounded failure metadata to the owner-only `~/.agents/hook-failures/v1/` spool, capped at 100 records; it stores no transcript body or raw session ID.

Use `dot agent doctor` for a read-only cross-agent check of persona and skill discovery, hook commands, local capability probes, source-store presence, latest complete ingestion, latest hook failure, partial state, and archive lag. It never reads transcript bodies or contacts vendor services. Repair is explicit: `dot agent doctor --fix --dry-run` previews the targeted forced chezmoi apply, and `dot agent doctor --fix` performs it.

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
- Every target has a config section: `level` is the depth a bare flag (and `--all`) selects, `paths` are the directories it removes, and `agents.sessions` carries the type and retention of each session store. `--days N` overrides every store for one run; `--days 0` makes every age eligible but never bypasses normalized-successor verification.

```yaml
prune:
  agents:
    sessions:
      - path: ~/.claude/projects
        source: claude
        keep_days: 7
      - path: ~/.local/share/opencode/opencode.db
        source: opencode
        keep_days: 7 # retained until safe row-level compaction exists
      - path: ~/.agents/sessions
        source: archive
        keep_days: 30
    keep: [memory, memory.jsonl, MEMORY.md] # never pruned, however old
  docker:
    level: build # set to system on a machine with no local k3d cluster
  node:
    level: cache
    paths: [~/.npm/_npx]
```

- `dot prune --agents` is the only command that deletes session logs; `dot agent session` and migration only gather or copy them. Aged Claude, Codex, and Antigravity sources require one exact complete normalized successor; unnormalized, stale, partial, unreadable, interrupted, and ambiguous sources are retained and reported.
- OpenCode and Copilot mix lineages in shared SQLite files, so pruning inventories but retains those databases until a source-specific row compactor can prove each deletion independently.
- `--dry-run` prints every raw-source decision with its type, hashed lineage, age, size, reason, and successor evidence.
- Agent long-term memory (`memory/`, `memory.jsonl`, `MEMORY.md`) is never pruned, however old.
- Missing tools and a stopped Docker daemon are reported as skipped, not as failures; a failing target never stops the others.

## Tasks

Inside the dotfiles repo, prefer the `mise run` wrappers: they depend on `build`, so they always run the current source rather than a stale `~/.local/bin/dot`.

```bash
mise run prune               # dot prune --all=deep
mise run prune --agents      # session logs and caches, keeping k3d and the Go cache
mise run verify              # dot verify
mise run release -- -y       # dot release, non-interactive
```

Project-management tasks are aliased with an `m` prefix (`mp`, `mx`, `mr`); the bare letters belong to the common vocabulary (`f` format, `c` check, `t` test, `b` build). `mr` is _not_ shorthand for `mise run` in a script — that is an interactive-only fish abbreviation.

## Config

All commands read `~/.config/dot.yaml`; `dot config init` scaffolds it with the built-in defaults and `dot config validate` checks it (unknown keys are rejected). A malformed file is fatal for every command except the `config` group, so the file stays repairable.

## Cluster diagnostics

`dot cluster diagnose` verifies the isolated context and namespace, then writes an owner-only JSON manifest under `~/.local/state/dot/diagnostics/`. It never collects Secret objects, kubeconfigs, full environment dumps, or unlimited logs, and never uploads a bundle.

```bash
dot cluster diagnose --namespace default
dot cluster diagnose --namespace app --output ./app-diagnostics.json --since 20m --tail 80
dot cluster diagnose --redact-pattern 'customer-[0-9]+'
```

Every allowlisted probe has a timeout, snapshot or time-window declaration, retained line limit, and retained byte limit. Logs are additionally bounded by pod count and per-container tail. Missing metrics or another individual probe failure is recorded as a sanitized partial error rather than discarding the successful evidence. Configure persistent project-specific RE2 patterns under `cluster.diagnostics.redact_patterns`; repeat `--redact-pattern` for one collection.

The reusable contract for presentation layers is exported by the `dot` package as `ClusterDiagnosticPlan`, `ClusterDiagnosticProbe`, `ClusterDiagnosticManifest`, and `ClusterDiagnosticSchemaVersion`. Consumers must render the generated manifest; they must not recreate or widen the collection commands.

## Gotchas

- Rebuild after changing `dot/`: `mise run build` (or `mise run apply` to build and apply).
- `dot cluster` derives an owner-only kubeconfig at `~/.kube/dot/<cluster-name>.yaml` unless `cluster.kubeconfig_path` or `--kubeconfig` overrides it. It never merges, overwrites, or switches the default kubeconfig; every kubectl call uses explicit target flags.
- Cluster mutations print and immediately re-verify the managed name, selected context, namespace, and non-secret ownership fingerprint. A renamed context requires `dot cluster --context <name> <command>` and must still resolve to the managed cluster's authoritative API server.
- Diagnostic manifests contain only the verified target identity and non-secret fingerprint, bounded sanitized probe results, and partial errors. Review the local file before sharing it; collection is not proof that external disclosure is authorized.
- The local k3d cluster must stay off by default — run `dot cluster stop` as soon as the task is done.
- `dot commit` and `dot pull-request` call an AI CLI, so they need it installed and authenticated.

## See Also

- [dot-release](../dot-release/SKILL.md) — Cutting a dotfiles release.
- [chezmoi](../chezmoi/SKILL.md) — Source naming, templates, and the edit-then-apply loop.
- [k8s-local](../k8s-local/SKILL.md) — Working inside the shared local cluster.
