---
name: full-review
description: Run this dotfiles repo's non-destructive health check across mise, dot, dependencies, docs, and skills.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/.agents/skills/full-review
  created: 2026-08-22
  updated: 2026-08-30
---

# Full Repository Review

Confirm this dotfiles repo is healthy end to end: the local gate is green, `dot` works, every pinned tool/dependency is current, and `AGENTS.md`/`README.md`/`skills/` still describe the live repo. Every step below is read-only or idempotent — no `chezmoi apply`, no `mise upgrade`, no `mise run release`, no `mise run prune`.

## 1. Validate the gate

```bash
mise run check   # check:actions, check:chezmoi, check:dprint, check:docs, check:go, check:leaks, check:lua, check:python, check:scan, check:shell, check:skills
mise run test     # Go (dot/) + Python (skills/python-stack) suites
mise run build    # compiles the dot binary
```

`check:docs` and `check:skills` already assert consistency between the documentation and the live CLI (mise tasks, `dot` commands, the `AGENTS.md` layout inventory, every `SKILL.md` under `skills/` and `.agents/skills/`) — a green `check:docs` proves that layer; don't hand-re-audit it.

## 2. Confirm the dot CLI

```bash
dot verify        # sanity checks: environment, tools, secrets, install freshness
dot --help        # every subcommand needs an alias, matching skills/dot-cli/SKILL.md
dot status        # git/docker/k3d smoke test
mise run doctor    # chezmoi doctor + mise doctor
mise run diff      # pending chezmoi diff, read-only
```

Only treat a `dot verify` failure as a repo bug when it is not explained by local environment state (an unauthenticated CLI, a missing optional secret) — those are sandbox facts, not defects.

## 3. Check tool and dependency currency (report, don't bump)

```bash
mise outdated                 # repo-scoped tools
mise -C "$HOME" outdated      # global tool list
mise --version                # warns here when mise itself is behind: `mise self-update`
go -C dot list -m -u all      # Go module updates
```

`mise outdated` only covers mise-managed pins. Also scan what it can't see: action versions in `.github/workflows/*.yml`, plugin versions in `dprint.json`, and any exact-version (not `"latest"`) tool pin in `dot_config/mise/config.toml.tmpl`. If something is genuinely stale, don't bump it here — hand off to [upgrade-tools](../../../skills/upgrade-tools/SKILL.md), one ecosystem at a time, validating after each.

## 4. Check documentation and skills for drift

Step 1's `check:docs` already proves `AGENTS.md` and every skill's `mise run`/`dot` references against the live CLI. The one gap it leaves: `README.md` isn't part of that contract. Read it and cross-check its skill count, prerequisites, install steps, and documented environment variables against the actual repository state. If either file has drifted, hand off to [readme-agents](../../../skills/readme-agents/SKILL.md).

## 5. Report

List every failure, staleness, or drift found, grouped by the step that surfaced it. Fix small, root-cause findings directly; route larger fix efforts (a real upgrade, a doc rewrite) to the dedicated skill named above instead of doing that work inline. Only report the review complete once steps 1 and 2 are green.

## See Also

- [mise](../../../skills/mise/SKILL.md) — the task vocabulary this review runs.
- [dot-cli](../../../skills/dot-cli/SKILL.md) — `dot` subcommand reference.
- [upgrade-tools](../../../skills/upgrade-tools/SKILL.md) — bump anything found stale in step 3.
- [readme-agents](../../../skills/readme-agents/SKILL.md) — fix any `AGENTS.md`/`README.md` drift found in step 4.
- [repository-review](../../../skills/repository-review/SKILL.md) — a deeper cross-cutting audit when this quick pass isn't enough.
