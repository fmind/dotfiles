---
name: dot-cli
description: Use the dot CLI (fmind/dot) to verify the environment, pull repos, manage the local k3d cluster, log agent sessions, and prune caches. Use for any dot command.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/dot-cli
  created: "2026-07-31"
  updated: "2026-09-03"
---

# Dot CLI

`dot` is the unified CLI of `fmind/dot`, built to `~/.local/bin/dot`. Every command has a one-letter alias (`pull-request` also answers to `pr`); `dot <command> --help` gives the exact flags.

## Commands

| Command            | Alias     | Purpose                                                                                                                                                                  |
| ------------------ | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `dot verify`       | `v`       | Sanity checks on environment, tools, secrets, and install freshness (`--json`, `--fix`)                                                                                  |
| `dot status`       | `s`       | Unified git, docker, and k3d status summary (`--json`)                                                                                                                   |
| `dot pull`         | `p`       | Concurrently pull the repositories listed in `~/.config/dot.yaml` (`--push` also pushes clean repos)                                                                     |
| `dot commit`       | `c`       | AI Conventional Commit from the staged diff; runs `git add -A` first when nothing is staged (`--type`, `--scope`)                                                        |
| `dot pull-request` | `pr`, `b` | AI PR description then `gh pr create` (`--base`, `--title`, `--draft`, `--label`, `--reviewer`, `--assignee`)                                                            |
| `dot release`      | `r`       | Prepare, tag, and push a dot release (`--yes`); see `.agents/skills/dot-release` inside the dot repository                                                               |
| `dot cluster`      | `k`       | Local k3d cluster: `start` (`s`), `stop` (`x`), `status` (`t`), `diagnose` (`g`), `delete --yes` (`d`), `namespace` (`n`)                                                |
| `dot prune`        | `x`       | Reclaim disk space from agent session logs and caches; flags in [references/prune-flags.md](references/prune-flags.md), flow in [reclaim-disk](../reclaim-disk/SKILL.md) |
| `dot agent`        | `a`       | Agent integrations: `doctor` (`d`), `hook` (`h`), `session` (`s`), `usage` (`u`)                                                                                         |
| `dot notify`       | `n`       | Desktop notification: `dot notify <agent> <event>` for hooks, `dot notify <summary> [headline] [details...]` for alerts                                                  |
| `dot chezmoi`      | `m`       | `clean` (`c`) finds `$HOME` orphans once managed by chezmoi and moves them to timestamped recoverable backups (`--yes`, `--interactive`)                                 |
| `dot config`       | `f`       | `~/.config/dot.yaml`: `show` (`s`), `path` (`p`), `init` (`i`), `edit` (`e`), `validate` (`v`)                                                                           |
| `dot login`        | `l`       | OAuth login wrappers: `github` (`g`), `workspace` (`w`), `gcp` (`c`), `clasp` (`a`)                                                                                      |
| `dot setup`        | `u`       | `workspace` (`w`) enables GCP APIs on a project and links it to `gws`: `dot setup workspace [PROJECT_ID]`                                                                |
| `dot completion`   | `g`       | Generate fish completions for `dot` and external CLIs                                                                                                                    |
| `dot context`      | `t`       | Bounded, redacted project context pack (`--bytes`, `--tokens`, `--format json`)                                                                                          |
| `dot version`      | `i`       | Print the version and embedded build metadata                                                                                                                            |

Global flags: `--config/-c <path>` (or `DOT_CONFIG_PATH`) and `--verbose` (or `DOT_VERBOSE`).

## Workflow

1. **Health**: `dot verify`, then `dot agent doctor` for persona, hooks, skills, and session-store health (`--fix` reapplies the managed integration targets with chezmoi, `--dry-run` previews).
1. **Sessions**: `dot agent session sync` ingests new transcripts from every agent into `~/.agents/sessions/v1/`; `list`, `show`, and `export` read the store.
1. **Legacy sessions**: `dot agent session migrate` dry-runs the selection of the most complete transcript per lineage; `--apply` writes it.
1. **Cluster**: `dot cluster start`, then `dot cluster diagnose --namespace <ns> --output <file> --since 20m` for a sanitized evidence bundle, and `dot cluster stop` when done; see [k8s-local](../k8s-local/SKILL.md).
1. **Disk**: `dot prune --dry-run --all=deep` to preview, then prune per [reclaim-disk](../reclaim-disk/SKILL.md); every target and depth is in [references/prune-flags.md](references/prune-flags.md), and long-term agent memory (`memory/`, `MEMORY.md`) is never pruned.
1. **Rebuild after source edits**: inside the dot repository, `mise run deploy` rebuilds the binary and reapplies it to `~/.local/bin/dot`.

## Gotchas

- **Stale binary**: when `dot agent doctor` reports `command-unavailable`, run `mise run deploy` inside the dot repository to sync the binary with the deployed hooks.
- **`dot commit` stages everything**: with nothing staged it runs `git add -A` before generating the message; stage selectively first when the tree holds unrelated changes.
- **Cluster resting state**: keep the k3d cluster stopped when inactive (`dot cluster stop`); `dot cluster delete --yes` removes it without a prompt.

## Documentation

- [fmind/dot](https://github.com/fmind/dot) — source, README, and the repository-scoped skills (`.agents/skills/dot-release`, `.agents/skills/chezmoi`).
- Companion skills: [agent-usage](../agent-usage/SKILL.md) (token accounting behind `dot agent usage`), [reclaim-disk](../reclaim-disk/SKILL.md) (prune flow), [k8s-local](../k8s-local/SKILL.md) (cluster inside a project), [mise](../mise/SKILL.md) (`mise run deploy`).
