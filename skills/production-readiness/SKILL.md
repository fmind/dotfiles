---
name: production-readiness
description: "Audit production operability and operational fitness: go/no-go, rollback, migrations, observability, recovery, support and halt thresholds; separate local, exact-head CI, runtime, deployed, and public release proof."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/production-readiness
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Production Readiness

Decide whether the exact candidate can be operated safely; this skill audits and recommends, it does not deploy, migrate, publish, or mutate production unless the user separately authorizes that action, while [cloud-run](../cloud-run/SKILL.md) deploys and [release](../release/SKILL.md) publishes.

## Workflow

1. **Resolve the candidate**: Record repository state, revision, artifact or image digest, configuration set, environment, dependency and schema versions, and the proposed rollout window; preserve dirty worktrees and never infer exact-head CI from another revision.
1. **Define working**: State critical user journeys, availability and correctness expectations, latency or capacity thresholds, data-loss tolerance, compliance constraints, and explicit stop conditions.
1. **Inspect the release delta**: Review code, dependency, configuration, infrastructure, identity, data, and operational changes; identify one-way doors, coupled releases, hidden manual steps, and compatibility windows.
1. **Audit security and identity**: Check least privilege, authentication and authorization, tenant isolation, secret lifecycle, supply chain, abuse controls, and safe defaults; use [threat-model](../threat-model/SKILL.md) for design risk and [secure](../secure/SKILL.md) for repository evidence.
1. **Audit data and migrations**: Verify forward and backward compatibility, rehearsal evidence, lock and duration risk, backup and restore, rollback semantics, data validation, and ownership of irreversible transitions.
1. **Audit observability**: Map each critical journey and failure mode to logs, metrics, traces, dashboards, and actionable alerts, each with an owner, a justified threshold, and a tested runbook.
1. **Audit capacity and cost**: Compare measured demand and headroom with explicit thresholds across saturation, rate limits, concurrency, timeouts, retries, quotas, degraded modes, and cost guardrails, without inventing traffic evidence.
1. **Audit rollout and recovery**: Prefer the smallest reversible exposure; define preflight checks, canary or staged progression, health windows, stop signals, the rollback owner and mechanism (see [cloud-run](../cloud-run/SKILL.md) for revision rollback), and post-rollback verification.
1. **Audit operations**: Confirm service ownership, support and escalation paths, dependency contacts, access, runbooks, maintenance burden, disaster recovery, and the first-hours monitoring plan.
1. **Gate the candidate**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).
1. **Verify proportionally**: Run only the authorized runtime or staging checks. Failed, stale, unavailable, or differently-scoped evidence remains a gap.
1. **Place the candidate on the proof ladder**: Never collapse these states; a candidate may be ready at one rung and blocked at the next, and every claim records artifact identity, environment, command or observation, timestamp, and source.
   - `source-ready`: the intended source and configuration are reviewable.
   - `local-green`: the repository-owned local checks pass for the candidate.
   - `exact-head-CI`: required CI passes for the exact immutable revision.
   - `runtime-proven`: the built artifact works across the authorized runtime boundary.
   - `deployed`: that artifact is running in the target environment.
   - `release-published`: the intended audience can obtain the released result (see [release](../release/SKILL.md)).
1. **Record the gates**: For each material gate report its status (pass, fail, blocked, or not run), exact evidence, owner, and smallest next step; lead with blockers, then gaps, verified gates, rollback plan, highest proven rung, and recommendation.
1. **Decide**: Return `GO`, `GO WITH CHANGES`, or `NO-GO`; authority, schedule pressure, and sunk cost cannot turn a failed hard gate into `GO`.

## Gotchas

- **Implicit authority**: Name any action that needs credentials, production access, external coordination, spend, or human approval instead of performing it.
- **Borrowed evidence**: A green run, probe, or deployment for a different revision or environment proves nothing about this candidate.

## Documentation

- Adapted from [ECC production audit](https://github.com/affaan-m/ECC/blob/59a99d669f5466d99d5be8b6fce8c5f2677766d0/skills/production-audit/SKILL.md), [agent-skills shipping and launch](https://github.com/addyosmani/agent-skills/blob/d2478bf0c73a6357df39a3ed6aff16acaa218843/skills/shipping-and-launch/SKILL.md).
- Companion skills: [quality-assurance](../quality-assurance/SKILL.md) (test campaign), [release](../release/SKILL.md) (publishing), [cloud-run](../cloud-run/SKILL.md) (deploy and rollback), [incident-response](../incident-response/SKILL.md) (active outage), [product-loop](../product-loop/SKILL.md) (audience, positioning, public rollout), [threat-model](../threat-model/SKILL.md) (design risk), [secure](../secure/SKILL.md) (repository evidence).
