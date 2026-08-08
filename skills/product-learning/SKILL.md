---
name: product-learning
description: Synthesize completed interviews, experiments, sales losses, prices, payment signals, launches, or incidents. Compare outcomes with the original bet and decide whether to continue, iterate, pivot, or stop against thresholds.
license: MIT
---

# Product Learning

Convert outcomes into a decision without rewriting the original bet after seeing the result. Optimize for the next reduction in uncertainty, not for defending prior work.

## Evidence Rules

- Recover the original hypothesis, target segment, baseline, success threshold, guardrail, kill criterion, and decision date before interpreting the result.
- Separate measured behavior, qualitative evidence, implementation or instrumentation facts, inference, and speculation.
- Never fabricate analytics, customer feedback, causality, sample quality, or confidence. Missing and biased data stay explicit gaps.
- Keep exact interview quotes, observed commitments, commercial offers, discounts, and buying behavior separate from synthesis. Record consent, segment, selection bias, contradictions, and provenance before generalizing.
- Segment before averaging. New and returning users, channels, cohorts, plans, environments, and time windows can hide opposite effects.
- Effort, agent-days, code volume, issue count, merges, and stars are not customer-value evidence.
- Do not contact users, mutate analytics, create issues, update persistent memory, publish a retrospective, or change a roadmap without explicit authorization.

## Workflow

1. **Reconstruct the bet:** State the user, job, intervention, expected behavior, mechanism, time horizon, thresholds, and assumptions as they existed before the outcome.
1. **Validate the measurement:** Check event semantics, denominators, exposure, identity and deduplication, missing data, contamination, novelty, selection bias, seasonality, and whether the measured candidate matches the intended change.
1. **Build the evidence ledger:** Collect quantitative results, cohort breakdowns, customer observations, interview contradictions, sales commitments and losses, pricing offers and payment signals, support evidence, operational outcomes, and implementation facts. Attach source, consent boundary, time window, segment, scope, and confidence to each claim.
1. **Compare expected with observed:** Show baseline, target, actual, guardrails, and uncertainty side by side. Mark each threshold as met, missed, inconclusive, or unmeasured.
1. **Explain carefully:** Generate a small set of competing explanations. For each, name supporting evidence, disconfirming evidence, and the cheapest observation that would distinguish it from alternatives.
1. **Check second-order effects:** Review trust, quality, retention, support load, reliability, cost, abuse, accessibility, and effects on adjacent segments. A local metric win can still harm the product.
1. **Decide:** Recommend `CONTINUE`, `ITERATE`, `PIVOT`, `STOP`, or `EXTEND TEST`. Tie the call to the predeclared thresholds; explain any justified deviation explicitly.
1. **Choose the next learning step:** Prefer the smallest test that resolves the most decision-critical uncertainty. Define the owner, segment, exposure, duration, primary behavior, guardrails, success threshold, and kill threshold.
1. **Capture durable learning:** Record what changed in the product model, which assumptions remain open, what should not be repeated, and which artifacts should be updated if the user authorizes it.

## Learning Review

Return:

- **Decision and confidence**
- **Original hypothesis and thresholds**
- **Evidence ledger with scope and quality**
- **Expected versus observed scorecard by segment**
- **Competing explanations and discriminating evidence**
- **Second-order effects and proof gaps**
- **Continue, iterate, pivot, stop, or extend rationale**
- **Next cheapest test with success and kill thresholds**
- **Authorized follow-ups versus proposed follow-ups**

Use [founder-discovery](../founder-discovery/SKILL.md) when the evidence invalidates the target problem or wedge, and [product-launch](../product-launch/SKILL.md) before expanding exposure.

## Sources

Adapted independently from [gstack retro at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/retro/SKILL.md) and [pm-skills experiment design at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/brainstorm-experiments-new/SKILL.md).
