---
name: plan-review
description: Red-team a product, architecture, implementation, migration, or launch plan before execution. Use for pre-mortems, strategy or scope challenge, and independent testing of assumptions, dependencies, and failure modes.
license: MIT
---

# Plan Review

Improve a plan by attacking its load-bearing assumptions while course correction is still cheap.

## Review Boundary

- Review only unless the user explicitly requests revisions or implementation.
- Read the plan, its source requirements, repository reality, and available evidence. Do not review a summary when the full artifact is available.
- Steelman the intended outcome before criticizing the approach.
- Prefer five decision-changing findings over a long generic risk list.
- Do not inflate scope in the name of ambition. The strongest review may recommend a smaller plan, a cheaper test, or no build.
- Separate verified conflicts, evidence-backed risks, assumptions, and questions.

## Lenses

Apply the lenses relevant to the plan and name which were used:

- **Founder:** Target user, painful job, wedge, why now, distribution, switching, monetization, defensibility, and kill assumptions.
- **Product:** Journey completeness, independent value, success metrics, guardrails, accessibility, trust, support, and learning loop.
- **Engineering:** Architecture, data flow, interfaces, invariants, failure handling, security, privacy, performance, compatibility, and maintainability.
- **Delivery:** Dependency order, vertical slices, test seams, migrations, observability, rollout, rollback, operational ownership, and authority boundaries.

## Workflow

1. **Reconstruct intent:** State the desired outcome, non-goals, constraints, evidence, and proof required for completion.
1. **Inspect current reality:** Verify relevant source paths, interfaces, dependencies, runtime assumptions, and existing mechanisms the plan proposes to replace or duplicate.
1. **Map claims:** Extract the decisions and assumptions on which the plan depends. Identify any requirement with no task, task with no requirement, or success claim with no proof.
1. **Challenge the premise:** Ask whether the problem is real, the chosen scope is the smallest useful wedge, and a no-build or manual alternative could learn more cheaply.
1. **Attack failure modes:** Imagine the plan failed through lack of value, integration breakage, data loss, abuse, operational burden, migration, adoption, or rollback failure. Trace concrete chains rather than naming categories.
1. **Test intended versus planned:** Compare documented permissions, user journeys, data rules, and operational promises with the actual tasks and verification steps.
1. **Rank findings:** Score each issue by impact, likelihood, confidence, and cheapness to test. Promote only issues that could change the decision or execution order.
1. **Offer remedies:** Give the smallest corrective change, cheapest decisive test, and kill or rollback criterion. Present numbered alternatives when a real trade-off remains.
1. **Issue a verdict:** Return `ACCEPT`, `ACCEPT WITH CHANGES`, `REVISE`, or `STOP AND TEST`, with the minimum conditions for the next state.

## Output

Lead with the verdict, then provide:

- **Load-bearing assumptions**
- **Findings by severity** with evidence and impact
- **Missing proof**
- **Cheapest tests and kill criteria**
- **Recommended scope changes**
- **Revised critical path** only when revisions were requested
- **Residual risks and owner decisions**

For a full repository audit, use [repository-review](../repository-review/SKILL.md). For diff-level implementation findings, use [code-review](../code-review/SKILL.md).

## Sources

Adapted independently from [gstack founder plan review at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/plan-ceo-review/SKILL.md), [gstack engineering plan review at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/plan-eng-review/SKILL.md), and [pm-skills strategy red-team at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-execution/skills/strategy-red-team/SKILL.md).
