---
name: product-design-review
description: Critique customer interface flows using interaction, visual, responsive, keyboard and screen-reader accessibility, content, and runtime evidence. Use for UX audits, design reviews, UI polish, onboarding, empty states, or redesign.
license: MIT
---

# Product Design Review

Judge whether a real user can understand, trust, and complete the surface's primary job, then improve only what the requested scope authorizes.

## Design Posture

- Identify the audience, context, and single job of the surface before discussing aesthetics.
- The brief and established product identity win. In Fmind work, use [fmind-visuals](../fmind-visuals/SKILL.md) for brand truth; do not replace Tokyo Night or existing tokens by habit.
- Classify the task as **preserve**, **refine**, or **redesign**. A critique does not authorize code edits, factual copy changes, or a new visual identity.
- Prefer one justified signature over decorative variety. Remove elements that do not encode meaning or help the task.
- Review real content, states, and runtime behavior. A static happy-path screenshot is insufficient.
- Bound iteration: one batched desktop/mobile review, one coherent fix pass when authorized, and one confirmation pass.

## Workflow

1. **Recover product truth:** Read the brief, existing product or design artifacts, tokens, components, user research, and representative content. Name any missing evidence.
1. **Map the journey:** Identify entry points, primary action, decisions, exits, failure recovery, and time to first value.
1. **Inspect the live surface:** Use [quality-assurance](../quality-assurance/SKILL.md) to capture desktop and mobile states, DOM semantics, console/network errors, keyboard behavior, focus, reduced motion, and screenshots when a runnable app exists.
1. **Review comprehension:** Check information architecture, hierarchy, labels, vocabulary, affordances, progressive disclosure, cognitive load, and whether the next action is obvious.
1. **Review every state:** Cover first run, loading, empty, partial, success, validation, permission, error, offline, destructive confirmation, and recovery states that apply.
1. **Review craft:** Evaluate typography, spacing, alignment, color, contrast, density, imagery, motion, consistency, and relationship to the subject. Flag generic defaults only when they weaken the brief.
1. **Review inclusion:** Check semantic structure, keyboard access, focus visibility and restoration, touch targets, zoom/reflow, screen-reader names, contrast, motion preferences, localization, and plain-language copy.
1. **Review constraints:** Identify performance, browser, device, content-length, data-density, privacy, and implementation constraints that change the design recommendation.
1. **Prioritize:** Rank findings by blocked task, trust or accessibility harm, frequency, and effort. Recommend the smallest coherent improvement before aesthetic extras.
1. **Verify authorized changes:** Re-run the same representative states at desktop and mobile sizes and record evidence without entering an unbounded polish loop.

## Output

Return:

- **Mode and user job**
- **Evidence reviewed**
- **Critical journey map**
- **Findings by severity** with screenshot, state, selector, or source evidence
- **Preserve / change / remove** decisions
- **Recommended design direction** with explicit tokens or interaction rules when useful
- **Accessibility and responsive gaps**
- **Runtime and proof gaps**
- **Next smallest improvement**

## Sources

Adapted independently from [Anthropic frontend-design at `f17010c`](https://github.com/anthropics/skills/blob/f17010c9bb483898c1d9c9f42dde2b3a98889434/skills/frontend-design/SKILL.md), [Impeccable at `aee6ce9`](https://github.com/pbakaus/impeccable/blob/aee6ce9352b842217b3f57c78296a7a4fa35a7f3/.agents/skills/impeccable/SKILL.md), and [gstack design-review at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/design-review/SKILL.md).
