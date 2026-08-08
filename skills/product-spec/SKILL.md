---
name: product-spec
description: Document feature contract/PRD/specification after validated discovery, interviews, or a clear outcome. Define requirements and acceptance, or clarify ambiguous behavior before implementation planning.
license: MIT
---

# Product Spec

Define what must be true for a product change to create value without prescribing unnecessary implementation.

## Boundaries

- Require either validated discovery or a clearly stated user outcome. If the premise is still uncertain, use [founder-discovery](../founder-discovery/SKILL.md) first.
- Inspect current product behavior and repository constraints before describing a change to an existing system.
- Default to a reviewable draft. Do not edit files, create issues, implement code, or publish anything unless the user requested that mutation.
- Keep business and user requirements technology-agnostic where possible. Record technical constraints only when they are genuine boundaries.
- Mark assumptions and unresolved decisions explicitly. Do not invent analytics, customer research, legal requirements, or operational evidence.
- Cut scope until the first priority is independently valuable, testable, deployable, and demonstrable.
- Scale the document to the change. A compact, low-risk feature should not acquire empty sections or repeated prose merely to satisfy a template.

## Long or Ambiguous Inputs

Use a requirements echo only when the source is genuinely long, unstructured, contradictory, or likely to change the course of work. Summarize the mission, locked decisions, reversals, parked items, open questions, and assumptions added by the agent; label user statements, evidence, inference, and proposals separately. Ask for agreement before an irreversible or materially divergent step, but do not stall a clear request merely because it is long. Do not persist raw notes or personal context into repository instructions without explicit scope and privacy authority.

## Workflow

1. **Restate intent:** Define the target user, current problem, desired outcome, business reason, and evidence supporting the change.
1. **Set the boundary:** List goals, non-goals, affected users, systems, data, platforms, and compatibility requirements.
1. **Prioritize journeys:** Write the primary journey as P1 and optional later journeys as P2/P3. Each journey must deliver value and be testable independently.
1. **Specify behavior:** Give functional requirements stable identifiers and precise observable outcomes. Cover permissions, validation, errors, loading, empty states, retries, cancellation, and recovery where relevant.
1. **Specify qualities:** Define measurable security, privacy, accessibility, performance, reliability, operability, portability, and cost constraints that matter to this change.
1. **Model information:** Identify key entities, ownership, lifecycle, retention, classification, and external system boundaries without prematurely choosing storage details.
1. **Define acceptance:** Express representative scenarios as Given/When/Then, including unhappy paths and boundary conditions. Tie every must-have requirement to at least one scenario.
1. **Define learning and rollout:** State success and guardrail metrics, instrumentation needs, rollout stages, rollback conditions, support implications, and the decision the evidence will inform.
1. **Review for executability:** Remove contradictions, duplicate requirements, vague adjectives, hidden scope, and requirements that cannot be verified. Surface blocking questions instead of guessing.

## Specification Contract

For a substantial or cross-cutting change, produce these sections:

1. **Decision summary**
1. **Problem and evidence**
1. **Target users and jobs**
1. **Goals and non-goals**
1. **Prioritized user journeys**
1. **Functional requirements**
1. **Quality requirements**
1. **Data and trust boundaries**
1. **Acceptance scenarios and edge cases**
1. **Success metrics and guardrails**
1. **Rollout, rollback, and support**
1. **Assumptions, dependencies, and open decisions**

For a compact, low-risk change, combine or omit immaterial sections. Always retain the decision summary, goals and non-goals, identified requirements, acceptance and edge cases, relevant data and trust boundaries, and assumptions or open decisions. Never omit a section merely because its content exposes unresolved risk.

After approval, use [implementation-plan](../implementation-plan/SKILL.md). Use [project-backlog](../project-backlog/SKILL.md) only when the user asks to turn accepted work into issue drafts or authorized GitHub issues.

## Sources

Adapted independently from [GitHub Spec Kit's specification template at `684b3d8`](https://github.com/github/spec-kit/blob/684b3d8e05263a7c1948d3d0699ab1cb4f77c3d5/templates/spec-template.md) and [gstack spec at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/spec/SKILL.md).
