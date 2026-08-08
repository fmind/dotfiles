---
name: founder-discovery
description: Challenge product ideas before specification or implementation. Use to decide whether to build, evaluate a concept, choose an MVP wedge, test demand, plan or synthesize customer interviews, or question product direction.
license: MIT
---

# Founder Discovery

Turn a promising idea into an evidence-labeled decision. Optimize for learning and a narrow wedge, not for producing a large plan.

## Ground Rules

- Separate observations, supplied evidence, inferences, and assumptions. Never convert enthusiasm into proof.
- Challenge the premise before refining the solution. Include doing nothing, a manual service, or a smaller change among the alternatives.
- Prefer a painful, frequent job for a specific user over a broad persona or an impressive feature list.
- Treat market size, customer demand, willingness to pay, and competitor claims as current facts that require primary evidence.
- Do not contact customers, create accounts, publish pages, buy ads, or spend money without explicit authorization.
- Planning an interview is read-only. Contacting or recording a person, retaining identifiable notes, or publishing a quote requires explicit authorization, applicable consent, and a stated data-retention boundary.
- Scale the dialogue to uncertainty. Proceed directly when the evidence and decision are already clear; ask one high-leverage question when an answer would materially change the direction.

## Workflow

1. **Recover context:** Read supplied research, product artifacts, repository instructions, and relevant implementation constraints. Use [technical-research](../technical-research/SKILL.md) when current external facts could change the decision.
1. **State the thesis:** Express the target user, painful job, proposed change, expected outcome, and why now in one sentence. Mark unsupported parts as assumptions.
1. **Interrogate the problem:** Determine how users solve it today, how often it occurs, what it costs them, who chooses or pays, and what evidence shows urgency.
1. **Find the wedge:** Identify the smallest end-to-end result that creates standalone value. Reject bundles of independent products and defer scale architecture until demand justifies it.
1. **Test founder logic:** Examine unique insight, distribution path, switching friction, business-model hypothesis, defensibility, operational ownership, and the founder's unfair access to the problem.
1. **Generate alternatives:** Present two or three materially different paths with trade-offs. Lead with the simplest recommended path and include a credible no-build path.
1. **Rank assumptions:** Score the load-bearing assumptions across value, usability, viability, feasibility, distribution, and trust by impact and uncertainty. Keep the list short.
1. **Design the cheapest decisive test:** Define the behavior to observe, target segment, sample or exposure, time box, success threshold, guardrail, and kill criterion. Prefer commitments and past behavior over compliments or stated intent.
1. **Make the call:** Recommend `BUILD`, `TEST FIRST`, `PARK`, or `STOP`, and name the evidence that would change the decision.

## Customer Interview Protocol

Use interviews to recover behavior, constraints, and commitments—not to pitch the solution or collect compliments.

1. **Set the recruiting hypothesis:** Name the segment, role, triggering situation, selection method, sample bias, and the assumption each conversation could change.
1. **Establish consent:** Explain the purpose, note-taking or recording method, retention, and any intended quote use. A conversation is not permission to publish an identity or quotation.
1. **Ask about the last concrete occurrence:** Recover the trigger, current workaround, frequency, time or money cost, participants, decision-maker, buying process, prior spend, failed alternatives, and consequence of doing nothing.
1. **Avoid leading intent:** Replace “Would you use this?” with past-behavior and trade-off questions. Do not reveal the proposed solution until the problem evidence is complete; stated willingness to pay is weaker than a paid or time-bound commitment.
1. **Seek a next commitment:** Ask for the smallest behavior that advances evidence, such as another stakeholder introduction, relevant artifact, scheduled follow-up, design-partner time, trial, or payment signal. Never manufacture urgency or imply authority to transact.
1. **Synthesize without flattening:** Keep exact quotes separate from interpretation. Cluster by frequency and intensity, record contradictions and disconfirming cases, note segment and selection bias, and stop claiming saturation when new evidence still changes the model.

## Founder Brief

Return a compact artifact with:

- **Decision:** The recommendation and one-sentence rationale.
- **Target user and job:** A narrow user, situation, and painful outcome.
- **Evidence ledger:** Confirmed observations, exact customer evidence, contradictions, inferences, and assumptions in separate groups, with consent and sample-quality limits.
- **Wedge:** The smallest valuable outcome and explicit non-goals.
- **Alternatives:** The considered paths and why the recommendation wins.
- **Business path:** Distribution, switching, monetization, and operational ownership hypotheses.
- **Risks:** At most five load-bearing assumptions, ordered by decision impact.
- **Next test:** The cheapest decisive experiment with success, guardrail, and stop thresholds.
- **Open decision:** The single unresolved choice that most affects the next step.

If the direction survives discovery, use [product-spec](../product-spec/SKILL.md). Use [product-learning](../product-learning/SKILL.md) to evaluate the resulting experiment or launch.

## Sources

Adapted independently for this workflow from [gstack office-hours at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/office-hours/SKILL.md), [pm-skills assumption discovery at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/identify-assumptions-new/SKILL.md), [pm-skills interview script at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/interview-script/SKILL.md), and [marketingskills customer research at `7868cb9`](https://github.com/coreyhaines31/marketingskills/blob/7868cb9251fad80a73d26e488a5ad5f6c4a9f335/skills/customer-research/SKILL.md).
