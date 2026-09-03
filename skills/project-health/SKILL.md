---
name: project-health
description: "Refresh an existing repository so everything is current, consistent, simple, and working: upgrade tools, cut debt, sync docs, green gate. Use for a full project review."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/project-health
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Project Health

The recurring pass that makes an existing repository current, consistent, simple, and working; it fixes in place and proves the result with the gate. [repository-review](../repository-review/SKILL.md) owns the read-only audit with ranked findings; deployments, registries, and external services stay out of scope.

## Workflow

1. **Baseline**: record `git status --short`, then `mise run check` and `mise run test` as they stand; fix a red baseline first so later regressions are attributable.
1. **Toolchain and dependencies**: bump one ecosystem at a time and validate between each per [upgrade-tools](../upgrade-tools/SKILL.md).
1. **Stack conformance**: layout, tooling, and tasks match the stack skill ([go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md), [typescript-stack](../typescript-stack/SKILL.md)).
1. **Tasks and hooks**: `mise.toml` exposes the canonical task vocabulary per [mise](../mise/SKILL.md); hooks and CI call those tasks per [lefthook](../lefthook/SKILL.md) and [github-actions](../github-actions/SKILL.md).
1. **Complexity**: remove dead code, duplicated logic, stale config, and unused dependencies per [reduce-complexity](../reduce-complexity/SKILL.md).
1. **Security**: run the offline scans wired into `mise run check` per [secure](../secure/SKILL.md): `check:leaks`, `check:scan`, `check:vuln`, `check:sast`, `check:actions`.
1. **Docs**: sync `README.md` and `AGENTS.md` per [readme-agents](../readme-agents/SKILL.md); trim wider docs per [improve-docs](../improve-docs/SKILL.md).
1. **Agent files**: promote repeated instructions into `.agents/skills/` per [skillify](../skillify/SKILL.md).
1. **Final gate**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).
1. **Report**: what changed per area (the tree holds only intended changes), what was left alone and why, and the highest proven rung of the [proof ladder](../production-readiness/SKILL.md).

## Documentation

- [mise tasks](https://mise.jdx.dev/tasks/) — the gate vocabulary this pass proves.
- Companion skills: [repository-review](../repository-review/SKILL.md) (audit only), [new-project](../new-project/SKILL.md) (bootstrap layer), [git-add-commit-push](../git-add-commit-push/SKILL.md) (commit on request).
