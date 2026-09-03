---
name: full-review
description: "Run this repository's read-only health pass: mise gate, dot CLI smoke test, tool currency, README drift. Use when asked to review or check the dot repository."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/.agents/skills/full-review
  created: "2026-08-22"
  updated: "2026-09-03"
---

# Full Review

Confirm the dot repository is healthy end to end: the gate is green, `dot` works, pinned tools are current, and `AGENTS.md`, `README.md`, and the skills describe the live repository. Every step is read-only or idempotent (no `chezmoi apply`, `mise upgrade`, `mise run release`, or `mise run prune`); [project-health](../../../skills/project-health/SKILL.md) is the fixing pass for any repository.

## Workflow

1. **Validate the gate**: `mise run check`, `mise run test`, and `mise run build`.
1. **Confirm the dot CLI**:

   ```bash
   dot verify        # environment, tools, secrets, install freshness
   dot --help        # every command carries an alias, matching skills/dot-cli/SKILL.md
   dot status        # git, docker, k3d smoke test
   mise run doctor   # chezmoi doctor + mise doctor
   mise run diff     # pending chezmoi diff, read-only
   ```

   A `dot verify` failure is a repository bug only when local state (an unauthenticated CLI, a missing optional secret) does not explain it.
1. **Check tool and dependency currency**, report only:

   ```bash
   mise outdated              # repository-scoped tools
   mise -C "$HOME" outdated   # global tool list
   mise --version             # warns when mise itself is behind
   go -C dot list -m -u all   # Go module updates
   ```

   `mise outdated` sees only mise pins; also scan action versions in `.github/workflows/*.yml`, plugin versions in `dprint.json`, and exact-version pins in `dot_config/mise/config.toml.tmpl`. Hand anything stale to [upgrade-tools](../../../skills/upgrade-tools/SKILL.md).
1. **Check README drift**: `README.md` is outside the docs contract; cross-check its skill count, prerequisites, install steps, and environment variables, and hand drift to [readme-agents](../../../skills/readme-agents/SKILL.md).
1. **Report**: every failure, staleness, or drift grouped by the step that surfaced it; fix small root-cause findings directly and route upgrades or doc rewrites to the skills above.

## Gotchas

- **Docs contract is in the gate**: `check:docs` and `check:skills` prove `AGENTS.md`, every `SKILL.md`, and their `mise run` and `dot` references against the live CLI; do not re-audit that layer by hand.
- **Complete means green**: report the review complete only when steps 1 and 2 pass.

## Documentation

- [mise tasks](https://mise.jdx.dev/tasks/) — the task vocabulary this review runs.
- Companion skills: [project-health](../../../skills/project-health/SKILL.md) (fix in place), [dot-cli](../../../skills/dot-cli/SKILL.md) (command reference).
- Hand-offs: [upgrade-tools](../../../skills/upgrade-tools/SKILL.md) (stale pins), [readme-agents](../../../skills/readme-agents/SKILL.md) (doc drift), [repository-review](../../../skills/repository-review/SKILL.md) (deeper audit).
