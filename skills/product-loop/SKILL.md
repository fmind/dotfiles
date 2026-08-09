---
name: product-loop
description: Discovery, specification, launch, and learning for a product in one loop. Build-or-stop calls, MVP wedges, demand tests, customer interviews, PRDs, requirements, user journeys, acceptance, positioning, onboarding, rollout, pricing, pivots.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/product-loop
  created: 2026-08-09
  updated: 2026-08-09
---

# Product Loop

One product decision cycle in four phases: **Discover** what deserves building, **Specify** what must be true, **Launch** to a bounded audience, then **Learn** whether the bet paid. Enter at the phase the evidence supports and stop at the next decision, rather than running all four by default.

## Phase Selection

| Situation                                                     | Phase    |
| ------------------------------------------------------------- | -------- |
| The problem, demand, or wedge is still unproven               | Discover |
| Discovery is validated and behavior must be pinned down       | Specify  |
| The change is built and needs a staged audience               | Launch   |
| The experiment, launch, or sales attempt has produced results | Learn    |

Entering the wrong phase is the common failure: writing requirements for an unvalidated problem, or claiming lessons from a launch that never shipped. Name the phase and its evidence before starting.

## Ground Rules

These apply to every phase.

- Separate observations, supplied evidence, inferences, and assumptions. Never convert enthusiasm into proof.
- Treat market size, demand, willingness to pay, and competitor claims as current facts that require primary evidence.
- Do not invent quotes, logos, metrics, demand, research, legal requirements, availability, or support capacity.
- Do not contact customers, create accounts, mutate a CRM or analytics, publish pages, buy ads, or spend money without explicit authorization.
- Planning an interview is read-only. Contacting or recording a person, retaining identifiable notes, or publishing a quote requires explicit authorization, applicable consent, and a stated data-retention boundary.
- Scale the artifact to the change. A compact, low-risk decision should not acquire empty sections merely to satisfy a template.
- Scale the dialogue to uncertainty. Proceed when the decision is already clear; ask one high-leverage question when an answer would materially change direction.

## Long or Ambiguous Inputs

Use a requirements echo only when the source is genuinely long, unstructured, contradictory, or likely to change the course of work. Summarize the mission, locked decisions, reversals, parked items, open questions, and assumptions added by the agent; label user statements, evidence, inference, and proposals separately. Ask for agreement before an irreversible or materially divergent step, but do not stall a clear request merely because it is long. Do not persist raw notes or personal context into repository instructions without explicit scope and privacy authority.

## Phase 1 — Discover

Turn a promising idea into an evidence-labeled decision. Optimize for learning and a narrow wedge, not for producing a large plan. Challenge the premise before refining the solution, and include doing nothing, a manual service, or a smaller change among the alternatives.

1. **Recover context:** Read supplied research, product artifacts, repository instructions, and implementation constraints. Use [technical-research](../technical-research/SKILL.md) when current external facts could change the decision.
1. **State the thesis:** Express the target user, painful job, proposed change, expected outcome, and why now in one sentence. Mark unsupported parts as assumptions.
1. **Interrogate the problem:** Determine how users solve it today, how often it occurs, what it costs them, who chooses or pays, and what evidence shows urgency.
1. **Find the wedge:** Identify the smallest end-to-end result that creates standalone value. Reject bundles of independent products and defer scale architecture until demand justifies it.
1. **Test founder logic:** Examine unique insight, distribution path, switching friction, business-model hypothesis, defensibility, operational ownership, and the founder's unfair access to the problem.
1. **Generate alternatives:** Present two or three materially different paths with trade-offs. Lead with the simplest recommended path and include a credible no-build path.
1. **Rank assumptions:** Score the load-bearing assumptions across value, usability, viability, feasibility, distribution, and trust by impact and uncertainty. Keep the list short.
1. **Design the cheapest decisive test:** Define the behavior to observe, target segment, sample or exposure, time box, success threshold, guardrail, and kill criterion. Prefer commitments and past behavior over compliments or stated intent.
1. **Make the call:** Recommend `BUILD`, `TEST FIRST`, `PARK`, or `STOP`, and name the evidence that would change the decision.

### Customer Interview Protocol

Use interviews to recover behavior, constraints, and commitments — not to pitch the solution or collect compliments.

1. **Set the recruiting hypothesis:** Name the segment, role, triggering situation, selection method, sample bias, and the assumption each conversation could change.
1. **Establish consent:** Explain the purpose, note-taking or recording method, retention, and any intended quote use. A conversation is not permission to publish an identity or quotation.
1. **Ask about the last concrete occurrence:** Recover the trigger, current workaround, frequency, time or money cost, participants, decision-maker, buying process, prior spend, failed alternatives, and consequence of doing nothing.
1. **Avoid leading intent:** Replace "Would you use this?" with past-behavior and trade-off questions. Do not reveal the proposed solution until the problem evidence is complete; stated willingness to pay is weaker than a paid or time-bound commitment.
1. **Seek a next commitment:** Ask for the smallest behavior that advances evidence, such as a stakeholder introduction, relevant artifact, scheduled follow-up, design-partner time, trial, or payment signal. Never manufacture urgency or imply authority to transact.
1. **Synthesize without flattening:** Keep exact quotes separate from interpretation. Cluster by frequency and intensity, record contradictions and disconfirming cases, note segment and selection bias, and stop claiming saturation when new evidence still changes the model.

### Discovery Brief

- **Decision** with a one-sentence rationale
- **Target user and job** — a narrow user, situation, and painful outcome
- **Evidence ledger** — confirmed observations, exact customer evidence, contradictions, inferences, and assumptions in separate groups, with consent and sample-quality limits
- **Wedge** and explicit non-goals
- **Alternatives** and why the recommendation wins
- **Business path** — distribution, switching, monetization, and operational ownership hypotheses
- **Risks** — at most five load-bearing assumptions, ordered by decision impact
- **Next test** with success, guardrail, and stop thresholds
- **Open decision** that most affects the next step

## Phase 2 — Specify

Define what must be true for a product change to create value without prescribing unnecessary implementation. Requires either validated discovery or a clearly stated user outcome; if the premise is still uncertain, return to Discover.

**Default to a reviewable draft.** Do not edit files, create issues, implement code, or publish anything unless the user requested that mutation. Keep business and user requirements technology-agnostic where possible, and record technical constraints only when they are genuine boundaries. Do not invent analytics, customer research, legal requirements, or operational evidence.

1. **Restate intent:** Define the target user, current problem, desired outcome, business reason, and evidence supporting the change.
1. **Set the boundary:** List goals, non-goals, affected users, systems, data, platforms, and compatibility requirements.
1. **Prioritize journeys:** Write the primary journey as P1 and optional later journeys as P2/P3. Each journey must deliver value and be testable independently.
1. **Specify behavior:** Give functional requirements stable identifiers and precise observable outcomes. Cover permissions, validation, errors, loading, empty states, retries, cancellation, and recovery where relevant.
1. **Specify qualities:** Define measurable security, privacy, accessibility, performance, reliability, operability, portability, and cost constraints that matter to this change.
1. **Model information:** Identify key entities, ownership, lifecycle, retention, classification, and external system boundaries without prematurely choosing storage details.
1. **Define acceptance:** Express representative scenarios as Given/When/Then, including unhappy paths and boundary conditions. Tie every must-have requirement to at least one scenario.
1. **Define learning and rollout:** State success and guardrail metrics, instrumentation needs, rollout stages, rollback conditions, support implications, and the decision the evidence will inform.
1. **Review for executability:** Remove contradictions, duplicate requirements, vague adjectives, hidden scope, and requirements that cannot be verified. Surface blocking questions instead of guessing. Cut scope until the first priority is independently valuable, testable, deployable, and demonstrable.

### Specification Contract

For a substantial or cross-cutting change, produce: decision summary; problem and evidence; target users and jobs; goals and non-goals; prioritized user journeys; functional requirements; quality requirements; data and trust boundaries; acceptance scenarios and edge cases; success metrics and guardrails; rollout, rollback, and support; assumptions, dependencies, and open decisions.

For a compact, low-risk change, combine or omit immaterial sections. Always retain the decision summary, goals and non-goals, identified requirements, acceptance and edge cases, relevant data and trust boundaries, and assumptions or open decisions. Never omit a section merely because its content exposes unresolved risk.

After approval, use [implementation-plan](../implementation-plan/SKILL.md). Use [project-backlog](../project-backlog/SKILL.md) only when the user asks to turn accepted work into issue drafts or authorized GitHub issues.

## Phase 3 — Launch

Turn a technically credible product into a staged learning and distribution plan. A code release is not automatically a product launch, and **a launch plan does not authorize publication**, customer contact, account changes, advertising, or spend.

Require a current [production-readiness](../production-readiness/SKILL.md) result for any customer-facing rollout, and carry blockers forward rather than translating them into marketing caveats. Define who owns product, engineering, support, security, communications, and the final go or no-go decision, even when one founder fills every role.

1. **Choose the launch job:** Decide whether this launch should learn, onboard design partners, validate willingness to pay, expand a proven segment, announce general availability, or re-engage existing users. One launch cannot optimize all goals equally.
1. **Sharpen positioning:** State the specific audience, painful alternative, promised outcome, differentiator, proof, and honest limitation. Prefer one defensible promise over a feature inventory.
1. **Design the offer:** Specify eligibility, pricing or trial, onboarding path, time-to-value, documentation, trust material, support path, feedback channel, and capacity limit.
1. **Stage exposure:** Progress through the smallest appropriate sequence — internal, named design partners, limited alpha, controlled beta, then general availability. Define entry, observation window, graduation, pause, and rollback thresholds for each stage.
1. **Select channels:** Use channels where the target users already seek solutions. Tie each channel to a message, owner, asset, expected behavior, cost ceiling, and measurement plan; avoid simultaneous broad distribution that destroys attribution.
1. **Prepare the experience:** Walk the entire path from first impression through authentication, setup, first value, billing where applicable, failure recovery, help, cancellation, and follow-up. Name unresolved friction rather than hiding it in launch copy.
1. **Prepare operations:** Confirm release readiness, support coverage, incident escalation, feedback triage, known-issue messaging, moderation or abuse handling, and a single launch-day decision channel.
1. **Set the scorecard:** Choose one primary behavior, leading indicators, quality and trust guardrails, segment breakdowns, success threshold, kill threshold, and decision date. Vanity traffic without activation or retention is not success.
1. **Run a preflight:** Rehearse the launch checklist and critical journey against the exact candidate. Verify links, analytics semantics, capacity, rollback, communications approval, and support response; mark anything untested.
1. **Make the call:** Return `LAUNCH`, `LIMIT EXPOSURE`, or `HOLD`, with the evidence, blockers, reversible next step, and explicit owner.

### Founder-Led Sales and Pricing Experiment

Use a bounded commercial experiment when the launch job is to validate a segment, buying process, value metric, or willingness to pay.

1. **Define the account hypothesis:** Name the ideal customer profile, triggering job, buyer, user, approver, budget owner, account list source, exclusion criteria, and founder capacity. A broad lead list is not a segment.
1. **Design the progression:** Separate discovery, qualification, solution fit, commercial proposal, and next commitment. Track time to first value, objections, lost reasons, no-decisions, stakeholder changes, and the exact commitment that advances or ends an opportunity.
1. **Form the pricing hypothesis:** State the value metric, packaging, price or range, trial or paid-pilot structure, discount guardrail, contract assumptions, and why the customer captures more value than the price. Price-sensitivity surveys and competitive anchors are inputs, not proof of buying behavior.
1. **Predeclare evidence:** Define the account count or exposure, observation window, qualified-conversation rate, next-commitment or payment signal, activation and retention guardrails, support cost, success threshold, and kill threshold. Do not reinterpret free interest as willingness to pay.
1. **Protect authority:** Drafting targets, scripts, pricing, and a CRM schema is allowed. Outreach, message sending, CRM or account mutation, contracts, payment or production price changes, publication, advertising, discounts, and spend require explicit authorization for the exact action and scope.

### Launch Brief

- **Decision and objective**
- **Audience, job, positioning, and proof**
- **Offer and onboarding path**
- **Stage plan** with entry, graduation, pause, and rollback thresholds
- **Channel and asset matrix**
- **Operational owners and support plan**
- **Measurement scorecard and observation window**
- **Known risks, non-goals, and authority-gated actions**
- **Commercial experiment, pricing hypothesis, commitments, and discount guardrails** when applicable
- **Next decision date**

## Phase 4 — Learn

Convert outcomes into a decision without rewriting the original bet after seeing the result. Optimize for the next reduction in uncertainty, not for defending prior work.

- Recover the original hypothesis, target segment, baseline, success threshold, guardrail, kill criterion, and decision date **before** interpreting the result.
- Separate measured behavior, qualitative evidence, implementation or instrumentation facts, inference, and speculation.
- **Never fabricate analytics**, customer feedback, causality, sample quality, or confidence. Missing and biased data stay explicit gaps.
- Keep exact interview quotes, observed commitments, commercial offers, discounts, and buying behavior separate from synthesis. Record consent, segment, selection bias, contradictions, and provenance before generalizing.
- Segment before averaging. New and returning users, channels, cohorts, plans, environments, and time windows can hide opposite effects.
- Effort, agent-days, code volume, issue count, merges, and stars are not customer-value evidence.
- Do not mutate analytics, create issues, update persistent memory, publish a retrospective, or change a roadmap without explicit authorization.

1. **Reconstruct the bet:** State the user, job, intervention, expected behavior, mechanism, time horizon, thresholds, and assumptions as they existed before the outcome.
1. **Validate the measurement:** Check event semantics, denominators, exposure, identity and deduplication, missing data, contamination, novelty, selection bias, seasonality, and whether the measured candidate matches the intended change.
1. **Build the evidence ledger:** Collect quantitative results, cohort breakdowns, customer observations, interview contradictions, sales commitments and losses, pricing offers and payment signals, support evidence, operational outcomes, and implementation facts. Attach source, consent boundary, time window, segment, scope, and confidence to each claim.
1. **Compare expected with observed:** Show baseline, target, actual, guardrails, and uncertainty side by side. Mark each threshold as met, missed, inconclusive, or unmeasured.
1. **Explain carefully:** Generate a small set of competing explanations. For each, name supporting evidence, disconfirming evidence, and the cheapest observation that would distinguish it from alternatives.
1. **Check second-order effects:** Review trust, quality, retention, support load, reliability, cost, abuse, accessibility, and effects on adjacent segments. A local metric win can still harm the product.
1. **Decide:** Recommend `CONTINUE`, `ITERATE`, `PIVOT`, `STOP`, or `EXTEND TEST`. Tie the call to the predeclared thresholds; explain any justified deviation explicitly.
1. **Choose the next learning step:** Prefer the smallest test that resolves the most decision-critical uncertainty. Define the owner, segment, exposure, duration, primary behavior, guardrails, success threshold, and kill threshold.
1. **Capture durable learning:** Record what changed in the product model, which assumptions remain open, what should not be repeated, and which artifacts should be updated if the user authorizes it.

### Learning Review

- **Decision and confidence**
- **Original hypothesis and thresholds**
- **Evidence ledger with scope and quality**
- **Expected versus observed scorecard by segment**
- **Competing explanations and discriminating evidence**
- **Second-order effects and proof gaps**
- **Continue, iterate, pivot, stop, or extend rationale**
- **Next cheapest test with success and kill thresholds**
- **Authorized follow-ups versus proposed follow-ups**

Preserve exact objections, commitments, prices offered, discounts, segment, and provenance before expanding the audience or changing pricing. Return to Discover when the evidence invalidates the target problem or wedge, and re-enter Launch before expanding exposure.

## Sources

Adapted independently from [gstack office-hours at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/office-hours/SKILL.md), [gstack spec at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/spec/SKILL.md), [gstack ship at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/ship/SKILL.md), [gstack retro at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/retro/SKILL.md), [GitHub Spec Kit's specification template at `684b3d8`](https://github.com/github/spec-kit/blob/684b3d8e05263a7c1948d3d0699ab1cb4f77c3d5/templates/spec-template.md), [pm-skills assumption discovery at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/identify-assumptions-new/SKILL.md), [pm-skills interview script at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/interview-script/SKILL.md), [pm-skills experiment design at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-discovery/skills/brainstorm-experiments-new/SKILL.md), [pm-skills pricing strategy at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-product-strategy/skills/pricing-strategy/SKILL.md), [marketingskills launch strategy at `7868cb9`](https://github.com/coreyhaines31/marketingskills/blob/7868cb9251fad80a73d26e488a5ad5f6c4a9f335/skills/launch/SKILL.md), and [marketingskills customer research at `7868cb9`](https://github.com/coreyhaines31/marketingskills/blob/7868cb9251fad80a73d26e488a5ad5f6c4a9f335/skills/customer-research/SKILL.md).
