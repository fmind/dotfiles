---
name: plan-review
description: Red-team a product, architecture, implementation, migration, or launch plan before execution. Use for pre-mortems, scope challenge, and testing assumptions, dependencies, and failure modes, not code.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/plan-review
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Plan Review

Attack a plan's load-bearing assumptions while course correction is still cheap; [implementation-plan](../implementation-plan/SKILL.md) writes plans and [diff-review](../diff-review/SKILL.md) reviews code.

## Workflow

1. **Reconstruct intent**: Read the full plan, its source requirements, and repository reality, never a summary; state the desired outcome, non-goals, constraints, evidence, and proof required for completion.
1. **Steelman first**: State the strongest case for the intended outcome before criticizing the approach.
1. **Inspect current reality**: Verify the source paths, interfaces, dependencies, runtime assumptions, and existing mechanisms the plan replaces or duplicates.
1. **Map claims**: Extract the decisions and assumptions the plan depends on; flag any requirement without a task, task without a requirement, or success claim without proof.
1. **Choose lenses**: Name which lenses were applied.
   - **Engineering**: architecture, data flow, interfaces, invariants, failure handling, security, privacy, performance, compatibility, and maintainability.
   - **Delivery**: dependency order, vertical slices, test seams, migrations, observability, rollout, rollback, operational ownership, and authority boundaries.
   - **Founder and product**: use the discovery and specification lenses of [product-loop](../product-loop/SKILL.md).
1. **Challenge the premise**: Ask whether the problem is real, the scope is the smallest useful wedge, and a no-build or manual alternative could learn more cheaply.
1. **Attack failure modes**: Imagine the plan failed through missing value, integration breakage, data loss, abuse, operational burden, migration, adoption, or rollback; trace concrete chains, not categories.
1. **Test intended versus planned**: Compare documented permissions, journeys, data rules, and operational promises with the actual tasks and verification steps.
1. **Rank findings**: Score each issue by impact, likelihood, confidence, and cheapness to test; promote only issues that could change the decision or execution order, and prefer five decision-changing findings over a long generic list.
1. **Offer remedies**: Give the smallest corrective change, cheapest decisive test, and kill or rollback criterion; present numbered alternatives when a real trade-off remains.
1. **Issue the verdict**: Return `GO`, `GO WITH CHANGES`, or `NO-GO` with the minimum conditions for the next state, then report:
   - Load-bearing assumptions, and findings ranked `P0`–`P3` (scale in [diff-review](../diff-review/SKILL.md)) with evidence and impact.
   - Missing proof, the cheapest tests and kill criteria, and recommended scope changes.
   - Residual risks and owner decisions, plus a revised critical path only when revisions were requested.

## Gotchas

- **Review only unless the user explicitly requests revisions or implementation**: Do not rewrite the plan or edit code while reviewing.
- **Do not inflate scope**: The strongest review may recommend a smaller plan, a cheaper test, or no build.
- **Separate verified conflicts, evidence-backed risks, assumptions, and questions**: never present a hunch as a finding.

## Documentation

- Adapted from [gstack founder plan review](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/plan-ceo-review/SKILL.md), [gstack engineering plan review](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/plan-eng-review/SKILL.md), [pm-skills strategy red-team](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-execution/skills/strategy-red-team/SKILL.md).
- Companion skills: [implementation-plan](../implementation-plan/SKILL.md) (the plan under review), [product-loop](../product-loop/SKILL.md) (founder and product lenses), [repository-review](../repository-review/SKILL.md) (full repository audit), [diff-review](../diff-review/SKILL.md) (code and the `P0`–`P3` scale).
