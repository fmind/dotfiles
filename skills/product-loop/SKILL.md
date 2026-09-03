---
name: product-loop
description: Run the product loop — discover, specify, launch, learn — with build-or-stop calls, MVP, demand tests, PRDs, journeys, acceptance, positioning, onboarding, rollout, pricing, pivots. Use when deciding, launching, or reviewing a product bet.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/product-loop
  created: "2026-08-09"
  updated: "2026-09-03"
---

# Product Loop

One product decision cycle in four phases: **Discover** what deserves building, **Specify** what must be true, **Launch** to a bounded audience, then **Learn** whether the bet paid. Enter at the phase the evidence supports and stop at the next decision; repository planning belongs to [implementation-plan](../implementation-plan/SKILL.md) and interface critique to [product-design-review](../product-design-review/SKILL.md).

| Situation                                                     | Phase    |
| ------------------------------------------------------------- | -------- |
| The problem, demand, or wedge is still unproven               | Discover |
| Discovery is validated and behavior must be pinned down       | Specify  |
| The change is built and needs a staged audience               | Launch   |
| The experiment, launch, or sales attempt has produced results | Learn    |

## Ground Rules

- Separate observations, supplied evidence, inferences, and assumptions. Never convert enthusiasm into proof.
- Do not invent quotes, logos, metrics, demand, research, legal requirements, availability, or support capacity.
- Do not contact customers, create accounts, mutate a CRM or analytics, publish pages, buy ads, or spend money without explicit authorization.
- Use a requirements echo only when the input is long, contradictory, or course-changing; label user statements, evidence, inference, and proposals separately.

## Workflow

### Discover

Challenge the premise before refining the solution, and keep doing nothing, a manual service, or a smaller change among the alternatives. Close with the discovery brief from [briefs](references/briefs.md).

1. **Recover context**: read supplied research, product artifacts, and repository constraints; use [technical-research](../technical-research/SKILL.md) when external facts could change the decision.
1. **State the thesis**: target user, painful job, proposed change, expected outcome, and why now in one sentence; mark unsupported parts as assumptions.
1. **Interrogate the problem**: how users solve it today, how often, what it costs them, who chooses or pays, and what evidence shows urgency.
1. **Find the wedge**: the smallest end-to-end result with standalone value; reject bundles of independent products and defer scale architecture until demand justifies it.
1. **Test founder logic**: unique insight, distribution path, switching friction, business model, defensibility, operational ownership, and unfair access to the problem.
1. **Generate alternatives**: two or three materially different paths with trade-offs, led by the simplest and including a credible no-build path.
1. **Rank assumptions**: score value, usability, viability, feasibility, distribution, and trust assumptions by impact and uncertainty; keep the list short.
1. **Design the cheapest decisive test**: behavior to observe, segment, exposure, time box, success threshold, guardrail, and kill criterion; prefer commitments over compliments (see [demand-tests](references/demand-tests.md)).
1. **Make the call**: `BUILD`, `TEST FIRST`, `PARK`, or `STOP`, naming the evidence that would change it.

### Specify

Default to a reviewable draft. Do not edit files, create issues, implement code, or publish unless the user requested that mutation. Do not invent analytics, customer research, legal requirements, or operational evidence; keep requirements technology-agnostic where possible.

1. **Restate intent**: target user, current problem, desired outcome, business reason, and the evidence supporting the change.
1. **Set the boundary**: goals, non-goals, affected users, systems, data, platforms, and compatibility requirements.
1. **Prioritize journeys**: the primary journey as P1 and later journeys as P2/P3; each delivers value and is testable independently.
1. **Specify behavior**: functional requirements with stable identifiers and observable outcomes, covering permissions, validation, errors, loading, empty states, retries, cancellation, and recovery.
1. **Specify qualities**: measurable security, privacy, accessibility, performance, reliability, operability, portability, and cost constraints that matter here.
1. **Model information**: entities, ownership, lifecycle, retention, classification, and external boundaries without choosing storage.
1. **Define acceptance**: Given/When/Then scenarios with unhappy paths and boundary conditions; tie every must-have requirement to one scenario.
1. **Define learning and rollout**: success and guardrail metrics, instrumentation, rollout stages, rollback conditions, support implications, and the decision the evidence informs.
1. **Review for executability**: remove contradictions, vague adjectives, hidden scope, and unverifiable requirements; surface blocking questions instead of guessing.

Write the specification contract from [briefs](references/briefs.md). For a compact, low-risk change, combine or omit immaterial sections. Always retain the decision summary, goals and non-goals, identified requirements, acceptance and edge cases, relevant data and trust boundaries, and assumptions or open decisions; never omit a section because it exposes unresolved risk. After approval, hand off to implementation-plan, or to [project-backlog](../project-backlog/SKILL.md) only when the user asks for issue drafts.

### Launch

A code release is not a product launch, and a launch plan does not authorize publication, customer contact, account changes, advertising, or spend. Require a current [production-readiness](../production-readiness/SKILL.md) result for any customer-facing rollout and carry its blockers forward. Close with the launch brief from [briefs](references/briefs.md).

1. **Choose the launch job**: learn, onboard design partners, validate willingness to pay, expand a proven segment, announce general availability, or re-engage users; one launch cannot optimize all of them.
1. **Sharpen positioning**: audience, painful alternative, promised outcome, differentiator, proof, and honest limitation; one defensible promise over a feature inventory.
1. **Design the offer**: eligibility, pricing or trial, onboarding path, time-to-value, documentation, trust material, support path, feedback channel, and capacity limit.
1. **Stage exposure**: internal, named design partners, limited alpha, controlled beta, then general availability, each with entry, observation window, graduation, pause, and rollback thresholds.
1. **Select channels**: where target users already seek solutions; tie each to a message, owner, asset, expected behavior, cost ceiling, and measurement plan.
1. **Prepare the experience**: walk first impression, authentication, setup, first value, billing, failure recovery, help, cancellation, and follow-up; name unresolved friction instead of hiding it in copy.
1. **Prepare operations**: release readiness, support coverage, incident escalation, feedback triage, known-issue messaging, abuse handling, and one launch-day decision channel.
1. **Set the scorecard**: one primary behavior, leading indicators, guardrails, segment breakdowns, success threshold, kill threshold, and decision date; vanity traffic is not success.
1. **Run a preflight**: rehearse the checklist and critical journey against the exact candidate; mark anything untested.
1. **Make the call**: `LAUNCH`, `LIMIT EXPOSURE`, or `HOLD`, with evidence, blockers, reversible next step, and owner. Validate a segment or willingness to pay with the bounded sales and pricing experiment in [demand-tests](references/demand-tests.md).

### Learn

Recover the original hypothesis, segment, baseline, thresholds, kill criterion, and decision date before interpreting the result. Never fabricate analytics, customer feedback, causality, sample quality, or confidence; effort, code volume, issue count, merges, and stars are not customer-value evidence. Segment before averaging, and close with the learning review from [briefs](references/briefs.md).

1. **Reconstruct the bet**: user, job, intervention, expected behavior, mechanism, time horizon, thresholds, and assumptions as they stood before the outcome.
1. **Validate the measurement**: event semantics, denominators, exposure, identity, missing data, contamination, novelty, selection bias, seasonality, and whether the measured candidate matches the change.
1. **Build the evidence ledger**: quantitative results, cohorts, customer observations, sales commitments and losses, payment signals, and support outcomes, each with source, consent boundary, window, segment, and confidence.
1. **Compare expected with observed**: baseline, target, actual, guardrails, and uncertainty side by side; mark each threshold met, missed, inconclusive, or unmeasured.
1. **Explain carefully**: a few competing explanations, each with supporting evidence, disconfirming evidence, and the cheapest discriminating observation.
1. **Check second-order effects**: trust, quality, retention, support load, reliability, cost, abuse, accessibility, and adjacent segments; a local metric win can still harm the product.
1. **Decide**: `CONTINUE`, `ITERATE`, `PIVOT`, `STOP`, or `EXTEND TEST`, tied to the predeclared thresholds with any deviation explained.
1. **Choose the next learning step**: the smallest test that resolves the most decision-critical uncertainty, with owner, segment, exposure, duration, primary behavior, guardrails, and thresholds.
1. **Capture durable learning**: what changed in the product model, which assumptions stay open, what not to repeat, and which artifacts to update if authorized; return to Discover when the evidence invalidates the wedge.

## Documentation

- References: [briefs](references/briefs.md) (discovery brief, specification contract, launch brief, learning review) and [demand-tests](references/demand-tests.md) (customer interview protocol, founder-led sales and pricing experiment).
- Companion skills: [technical-research](../technical-research/SKILL.md) (external facts), [production-readiness](../production-readiness/SKILL.md) (rollout gate), [implementation-plan](../implementation-plan/SKILL.md) (repository plan), [project-backlog](../project-backlog/SKILL.md) (issue drafts), [product-design-review](../product-design-review/SKILL.md) (interface critique).
- Adapted from [gstack office-hours](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/office-hours/SKILL.md), [gstack spec](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/spec/SKILL.md), [gstack ship](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/ship/SKILL.md), [gstack retro](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/retro/SKILL.md), [Spec Kit spec template](https://github.com/github/spec-kit/blob/684b3d8e05263a7c1948d3d0699ab1cb4f77c3d5/templates/spec-template.md), [pm-skills assumptions](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/identify-assumptions-new/SKILL.md), [pm-skills interview script](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/interview-script/SKILL.md), [pm-skills experiments](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/brainstorm-experiments-new/SKILL.md), [pm-skills pricing](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-strategy/skills/pricing-strategy/SKILL.md), [marketingskills launch](https://github.com/coreyhaines31/marketingskills/blob/7868cb9251fad80a73d26e488a5ad5f6c4a9f335/skills/launch/SKILL.md), [marketingskills customer research](https://github.com/coreyhaines31/marketingskills/blob/7868cb9251fad80a73d26e488a5ad5f6c4a9f335/skills/customer-research/SKILL.md).
