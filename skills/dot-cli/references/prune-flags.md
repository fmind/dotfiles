# `dot prune` Flags

Every target and depth accepted by `dot prune`; the safety-ordered workflow lives in [reclaim-disk](../reclaim-disk/SKILL.md). Nothing runs unless a target is selected, and targets compose freely (`dot prune --agents --go`). The binary's `dot prune --help` stays the live reference.

## Targets

| Flag        | Short | Depths            | Removes                                                                                     |
| ----------- | ----- | ----------------- | ------------------------------------------------------------------------------------------- |
| `--agents`  | `-a`  | `sessions`        | Expired session logs from every store in `prune.agents.sessions`                             |
| `--docker`  | `-d`  | `build`, `system` | Docker build cache; `=system` also stopped containers, networks, and dangling images         |
| `--go`      | `-g`  | `build`, `module` | Go build and test caches; `=module` also the module cache                                    |
| `--python`  | `-p`  | `cache`, `all`    | Unused uv cache entries; `=all` wipes the uv cache and purges pip                            |
| `--node`    | `-n`  | `cache`, `all`    | The npx cache; `=all` also the npm cache                                                     |
| `--mise`    | `-m`  | `cache`, `configs` | Unused tool versions, cache, and downloads; `=configs` also untracked config links           |
| `--tools`   | `-t`  | `cache`           | Trivy, Helm, dprint, and golangci-lint caches                                                |
| `--all`     | `-A`  | `shallow`, `deep` | Every target at its configured depth; `=deep` selects the deepest depth of each              |

## Modifiers

| Flag        | Short | Effect                                                                                                              |
| ----------- | ----- | ------------------------------------------------------------------------------------------------------------------- |
| `--days N`  | `-D`  | Override age retention for every agent session store; `0` makes every age eligible, and source safety checks still apply. Defaults to each store's `keep_days`. |
| `--dry-run` | `-N`  | Report what would be removed without deleting anything.                                                             |

## Gotchas

- **Preview first**: `dot prune --dry-run --all=deep` is the safe way to see the deepest possible sweep before committing to it.
- **Memory is never pruned**: long-term agent memory (`memory/`, `MEMORY.md`) is out of scope for every target.
- **`--go=module` is expensive to undo**: the module cache re-downloads on the next build, so prefer the default `build` depth unless the disk is genuinely full.
