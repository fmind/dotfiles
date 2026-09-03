# Handover Prompt Template

Sections of a handover prompt written by [handover](SKILL.md), in this order. Omit a section only when it is genuinely empty; never pad it.

- **Objective**: the outcome in one or two sentences, stated as the goal, not the history.
- **Context**: the repository, branch, relevant paths, and the shape of the surrounding code the work touches.
- **Constraints**: user-stated requirements marked as non-negotiable, project conventions that apply, and explicitly rejected approaches.
- **Established Facts**: findings verified this session, each with how it was verified (file read, command run, documentation consulted).
- **Current State**: done (with proof), in progress (with the exact stopping point), not started.
- **Tasks**: ordered, each with its acceptance criterion; prefer the smallest slice that produces working, verifiable behavior.
- **Verification**: the commands that must pass, and any check known to be red and why.
- **Open Questions**: decisions still owned by the user, with the options and your recommendation.
