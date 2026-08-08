---
name: code-review
description: Inspect changed code in an exact diff, patch, branch, PR, or candidate for correctness defects. Use for pre-merge or self-review, spec compliance, test gaps, and regression risk before broader QA, not whole-repo audit.
license: MIT
---

# Code Review

Find defects that would justify changing the candidate, with enough evidence for the author to reproduce and fix them.

## Boundary

- Review only by default. Do not edit code, resolve threads, commit, push, or approve a pull request unless the user separately requests that action.
- Resolve the exact target and record base, head, working-tree state, and whether the evidence covers a dirty candidate, local commit, or remote pull-request head.
- Preserve staged, unstaged, and untracked work. Use an isolated temporary worktree when broad validation would mutate or misrepresent the candidate.
- Review against repository instructions, the issue or spec contract, and current source behavior. A green suite does not prove the intended behavior was implemented.
- Report only actionable defects, material test gaps, or explicit requested nits. Do not manufacture findings to make the review look useful.

## Workflow

1. **Resolve intent and target:** Read the request, issue, spec, plan, and change description. Identify the exact comparison and proof already supplied.
1. **Inventory the delta:** Inspect changed files, generated artifacts, dependency or schema changes, and nearby code needed to understand behavior. Do not review the diff in isolation when invariants live elsewhere.
1. **Read tests first:** Determine what behavior the candidate claims, whether tests can fail for the defect class, and which requirements remain unproved.
1. **Trace intended versus implemented:** Map permissions, user journeys, data rules, failure semantics, and operational promises to concrete code paths and tests.
1. **Review by risk:** Examine correctness, data integrity, authorization, input boundaries, concurrency, resource lifecycle, error propagation, compatibility, migration, performance, observability, and rollback in proportion to the change. Challenge pass-through abstractions, hidden dependency construction, tests that bypass the public seam, and hypothetical flexibility with no second concrete use.
1. **Verify candidates:** Reproduce each suspected defect with code tracing, a focused test, or a safe isolated temporary experiment. Quote the specific file and line that makes the finding real.
1. **Run proportional checks:** Start with focused tests and static analysis; run the repository-owned full gate when cost and target coherence permit. Record exactly which candidate each result covers.
1. **Calibrate:** Discard preferences and speculation. Rank remaining findings by user impact, exploitability, data loss, regression likelihood, and confidence.

## Scope Discipline

Compare every changed dependency, configuration file, public API, generated artifact, and apparently unrelated hunk with the stated contract and its actual call or build path. Classify it as **keep** when it is necessary and connected, **split** when it is independently valuable or unrelated, or **justify** when the coupling is real but non-obvious. Path names alone do not prove scope creep. Report the classification with evidence; never stage, revert, discard, or rewrite the candidate merely because a detector labels a path unrelated.

## Finding Format

Use severities:

- **P0:** Immediate security breach, irreversible data loss, or broad outage risk.
- **P1:** Likely correctness, security, or availability defect that should block merge.
- **P2:** Material edge-case, maintainability, performance, or test defect worth fixing before or soon after merge.
- **P3:** Minor issue only when the user requested exhaustive review.

Write each finding as:

```text
[P1] Short imperative title — path/to/file.ext:line
Evidence: the exact behavior or code path.
Impact: who or what fails, under which condition.
Reproduction: the smallest command, scenario, or trace.
Correction: the minimum direction, without implementing it.
```

Lead with findings ordered by severity. If there are none, say so directly, then list testing and proof gaps. End with the target identity, checks run, and residual risks.

Use [repository-review](../repository-review/SKILL.md) for cross-cutting repository readiness. When the task is specifically to address existing GitHub review threads, use the available review-comment workflow after resolving the exact pull request.

## Sources

Adapted independently from [agent-skills code-review-and-quality at `d2478bf`](https://github.com/addyosmani/agent-skills/blob/d2478bf0c73a6357df39a3ed6aff16acaa218843/skills/code-review-and-quality/SKILL.md), [gstack review at `960c3a8`](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/review/SKILL.md), [pm-skills intended-vs-implemented at `18468a9`](https://github.com/phuryn/pm-skills/blob/18468a95b427e70e258b51389796367c6f684e7d/pm-ai-shipping/skills/intended-vs-implemented/SKILL.md), and [Matt Pocock's codebase design at `84fdeff`](https://github.com/mattpocock/skills/blob/84fdeffd12f2ee307994d1eb6feb48173b6e0502/skills/engineering/codebase-design/SKILL.md).
