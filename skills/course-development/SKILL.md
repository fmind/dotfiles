---
name: course-development
description: "Build a technical course: lessons, executable labs, prerequisites, guided practice, accessibility, release acceptance. Use when writing or revising a course, chapter, lab, or tutorial."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/course-development
  created: "2026-08-30"
  updated: "2026-09-03"
---

# Develop a Technical Course

Turn a technical subject into a course that learners can understand, execute, and finish. Own the learning contract: [hugo](../hugo/SKILL.md) owns the site stack, [quality-assurance](../quality-assurance/SKILL.md) the test campaign, and [playwright](../playwright/SKILL.md) browser checks. A course repository's own `AGENTS.md` overrides this skill wherever they differ.

## Workflow

1. **Define the learner**: state prerequisites, target capability, available time, delivery platform, and accessibility constraints; cut content that does not advance the capability.
1. **Write observable outcomes**: give each page one primary outcome and a completion signal; closing bullets state what the learner can now do, name, or predict, never attendance.
1. **Frame every page**: follow [page](references/page.md): front matter with an explicit lowercase `slug` (plus `aliases` for any previously served URL), an "In one glance" abstract (You will / You need / Time), H2s that name their teaching purpose, `## Your turn` with the seven exercise fields, and `## What you can do now` ending in a `Continue to` link.
1. **Mirror the source**: quote code with `{{< include path="<file>" region="<name>" lang="go" >}}` over named regions of the shipped reference implementation; paste command output verbatim in `text` blocks, and derive every quoted count from a generated manifest (`data/captures.yaml` via `mise run docs:captures` in the reference course).
1. **Make labs executable**: state a prediction before the exercise, then declare the Mode (`inspect`, `temporary experiment` with target-specific preflight and cleanup, or `keep`), a preflight command, ordered steps, the gate command that proves completion, and the final state; label offline, live-model, container, Kubernetes, cloud, destructive, and paid commands.
1. **Explain diagrams**: follow every [mermaid](../mermaid/SKILL.md) diagram with `**Diagram in words:**` prose; define terms at first use and put the reason beside each command.
1. **Check the human surface**: verify navigation, reading order, keyboard use, contrast, alt text, mobile layout, and copy-paste on rendered pages with playwright, then run the course's accessibility gate (`mise run check:accessibility` in the reference course).
1. **Validate progressively**: run the docs and link gates on the changed page first (`mise run check:docs` and `mise run check:links`), then the learner gate from a clean clone (`mise run install`, `mise run doctor`, `mise run check:core`, `mise run test`), then the definition of done in the course's `AGENTS.md`.
1. **Prepare release acceptance**: record the exact candidate, supported platforms, test evidence, known limitations, and correction path, and report the highest proven rung of the [proof ladder](../production-readiness/SKILL.md); publishing remains a separate authorization.

## Gotchas

- **Pages grow**: a rewrite that adds a definition pays for it by cutting tease, restatement, and asides.
- **Prerequisite creep**: `You need` declares machine state as the command that produces it (`mise run install` done), never "Chapter N finished".
- **Frozen routes**: a published slug never changes; a route change records the old address in the course's released-URL ledger or fails the build.
- **Scene headings**: an H2 names its technical subject and purpose, never a persona, a clock time, or a riddle.
- **Optional exercises**: keep them inline with a bold `**Optional exercise:**` label so they stay out of the sidebar.

## Documentation

- [Hextra](https://imfing.github.io/hextra/docs/) · [Hugo](https://gohugo.io/documentation/)
- Reference course: `~/mlops-courses/agentops-open-course` (its `AGENTS.md` owns the page frame, gates, and authoring rules).
- Companion skills: [hugo](../hugo/SKILL.md) (site stack), [mermaid](../mermaid/SKILL.md) (diagrams), [playwright](../playwright/SKILL.md) (browser checks), [quality-assurance](../quality-assurance/SKILL.md) (test campaign), [production-readiness](../production-readiness/SKILL.md) (proof ladder).
