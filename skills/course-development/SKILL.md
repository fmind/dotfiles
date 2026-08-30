---
name: course-development
description: Design executable technical curricula, lessons, and labs with prerequisites, guided practice, accessibility, platform checks, and course release acceptance.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/course-development
  created: 2026-08-30
  updated: 2026-08-30
---

# Develop a Technical Course

Turn a technical subject into a course that learners can understand, execute, and finish. Own the learning contract; use [hugo](../hugo/SKILL.md) for the site, [fmind-visuals](../fmind-visuals/SKILL.md) for slides, and [quality-assurance](../quality-assurance/SKILL.md) for the test campaign.

## Workflow

1. **Define the learner:** State prerequisites, target capability, available time, delivery platform, and accessibility constraints. Remove content that does not advance the target capability.
1. **Write observable outcomes:** Express what the learner will build, diagnose, explain, or decide. Give each module one primary outcome and a completion signal.
1. **Sequence the journey:** Move from a minimal working example through guided practice to an independent lab. Introduce each concept immediately before it is used.
1. **Make examples executable:** Pin dependencies, include expected commands and outputs, keep fixtures small, and run every code path in a clean environment. Never publish placeholder code or unverified APIs.
1. **Design practice and feedback:** Give labs a concrete starting state, success criteria, likely failure modes, and recovery hints. Keep solutions separate enough that learners can attempt the work first.
1. **Check the human surface:** Verify navigation, reading order, keyboard use, contrast, alt text, captions or transcripts, mobile layout, copy-paste behavior, and platform-specific constraints.
1. **Validate progressively:** Test the changed example first, then its lesson, module, navigation, links, and the repository's full gate. Exercise both the instructor and fresh-learner paths.
1. **Prepare release acceptance:** Record the exact candidate, supported platform, test evidence, known limitations, and rollback or correction path. Publishing remains a separate authorization.

## Quality Bar

- Prefer one coherent learning path over a catalog of disconnected features.
- Teach the reason and trade-off before adding abstraction.
- Keep sample output deterministic and redact credentials, personal data, and paid-service responses.
- Distinguish local rendering, platform preview, and publicly released evidence.

## Handoff

Report the learner outcome, changed modules, exercised labs, platform and accessibility evidence, and anything that still requires a human instructor or hosted-platform check.
