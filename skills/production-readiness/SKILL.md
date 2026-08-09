---
name: production-readiness
description: Audit operational fitness, safety, and operability of an exact service/agent/build/infrastructure candidate. Check go/no-go, rollback, migrations, observability, recovery, and deployed proof. Not for diff defects or repo-wide audits.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/production-readiness
  created: 2026-08-08
  updated: 2026-08-09
---

# Production Readiness

Decide whether the exact candidate can be operated safely. This skill audits and recommends; it does not deploy, migrate, publish, or mutate production unless the user separately authorizes that action.

## Proof Ladder

Never collapse these states:

1. `source-ready`: The intended source and configuration are reviewable.
1. `local-green`: The repository-owned local checks pass for the candidate.
1. `exact-head-CI`: Required CI passes for the exact immutable revision.
1. `runtime-proven`: The built artifact works across the authorized runtime boundary.
1. `deployed`: That artifact is running in the target environment.
1. `release-published`: The intended audience can obtain or use the released result.

A candidate may be ready at one level and blocked at the next. Record artifact identity, environment, command or observation, timestamp, and source for every proof claim.

## Workflow

1. **Resolve the candidate:** Record repository state, revision, build artifact or image digest, configuration set, environment, dependency and schema versions, and the proposed rollout window. Preserve dirty worktrees and do not infer exact-head CI from a different revision.
1. **Define working:** State the critical user journeys, availability and correctness expectations, latency or capacity thresholds, data-loss tolerance, compliance constraints, and explicit stop conditions.
1. **Inspect the release delta:** Review code, dependency, configuration, infrastructure, identity, data, and operational changes. Identify one-way doors, coupled releases, hidden manual steps, and compatibility windows.
1. **Audit security and identity:** Check least privilege, authentication and authorization, tenant isolation, secret lifecycle, supply chain, abuse controls, and safe defaults. Use [threat-model](../threat-model/SKILL.md) for design risk and [security-scan](../security-scan/SKILL.md) for repository evidence.
1. **Audit data and migrations:** Verify forward and backward compatibility, rehearsal evidence, lock and duration risk, backup and restore, rollback semantics, data validation, and ownership of irreversible transitions.
1. **Audit observability:** Map each critical journey and failure mode to structured logs, metrics, traces, dashboards, actionable alerts, and a tested runbook. Every alert needs an owner and a justified threshold.
1. **Audit capacity and cost:** Compare measured demand and headroom with explicit thresholds. Cover saturation, rate limits, concurrency, timeouts, retries, quotas, degraded modes, and cost guardrails without inventing traffic evidence.
1. **Audit rollout and recovery:** Prefer the smallest reversible exposure. Define preflight checks, canary or staged progression, health windows, automatic and human stop signals, rollback owner, rollback command or mechanism, and post-rollback verification.
1. **Audit operations:** Confirm service ownership, support and escalation paths, dependency contacts, access, runbooks, maintenance burden, disaster recovery, and the first-hours monitoring plan.
1. **Protect unrelated work:** Before any full gate, inspect the full gate's task definition and working-tree state. If it runs whole-tree write-formatters and unrelated or user changes are present, validate the exact candidate in an isolated temporary worktree or run equivalent non-mutating checks; never reformat unrelated work.
1. **Verify proportionally:** Run the repository-owned local gate, normally `mise run all`, and only the authorized runtime or staging checks. Failed, stale, unavailable, or differently-scoped evidence remains a gap.
1. **Make the gate decision:** Return `GO`, `CONDITIONAL GO`, or `NO-GO`. Authority pressure, schedule pressure, and sunk cost cannot turn a failed hard gate into `GO`.

## Gate Record

For every material gate report:

| Gate                            | Status                       | Evidence     | Owner                | Required action    |
| ------------------------------- | ---------------------------- | ------------ | -------------------- | ------------------ |
| Candidate and artifact identity | `READY`, `BLOCKER`, or `GAP` | Exact source | Named role or person | Smallest next step |

Lead with blockers, then conditional gaps, verified gates, rollback plan, proof ceiling, and recommendation. Name any action that requires credentials, production access, external coordination, spend, or human approval rather than performing it implicitly.

Use [quality-assurance](../quality-assurance/SKILL.md) for the test campaign, [incident-response](../incident-response/SKILL.md) during an active outage, and [product-loop](../product-loop/SKILL.md) for audience, positioning, channels, and public rollout.

## Sources

Adapted independently from [Everything Claude Code production audit at `59a99d6`](https://github.com/affaan-m/ECC/blob/59a99d669f5466d99d5be8b6fce8c5f2677766d0/skills/production-audit/SKILL.md) and [agent-skills shipping and launch at `d2478bf`](https://github.com/addyosmani/agent-skills/blob/d2478bf0c73a6357df39a3ed6aff16acaa218843/skills/shipping-and-launch/SKILL.md).
