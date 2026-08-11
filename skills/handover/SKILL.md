---
name: handover
description: Write a self-contained follow-up prompt carrying this session's decisions, constraints, and verified findings into a fresh agent session. Use when handing unfinished work to a new context window.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/handover
  created: 2026-08-11
  updated: 2026-08-11
---

# Handover

Turn everything this session established into one prompt a fresh agent can execute without asking a question.

## Authority Boundary

Writing a handover is a documentation act. Produce the prompt file and stop. Do not implement the work it describes, commit it, or push it unless the user separately asks. The next session owns execution.

## What a Handover Carries

The receiving agent starts with an empty context window. It sees the repository and the prompt — nothing else. Four classes of knowledge die with this session unless the prompt captures them:

- **User-stated constraints**: preferences, rejections, and scope limits the user gave in conversation. These are authoritative and unrecoverable from the code.
- **Verified findings**: facts confirmed this session by reading dependency source, running a command, or fetching documentation — with the evidence that settled them.
- **Rejected alternatives**: options considered and dropped, with the reason. Without these the next session re-litigates settled decisions or walks into a known dead end.
- **Current position**: what is done, what is half-done, and what has not started.

Repository state that the next agent can read for itself is not knowledge — it is noise. Point at files instead of transcribing them.

## Workflow

1. **Re-read the session:** Recover the original request, every subsequent user correction, and the decisions each one settled. A later correction overrides an earlier instruction; carry the resolved position, not the debate.
1. **Separate instruction from inference:** Mark what the user required, what you proposed and they approved, and what you assumed without confirmation. The next agent must know which lines it may revise.
1. **Verify before you write:** Re-check every path, symbol, command, and version the prompt will name. Stale coordinates are worse than absent ones — they send the next session confidently to the wrong place. Drop or re-derive anything you cannot confirm now.
1. **Capture the position:** Record completed work with its proof, in-progress work with its exact stopping point, and untouched work. Name the branch and whether the tree is clean.
1. **Record the trade-offs:** For each real decision, state the choice, the alternatives, and why. For each dead end, state what failed and how it failed, so the next session does not retry it.
1. **State the proof contract:** Name the commands that must pass before the work is done — normally `mise run check` and `mise run test` — and any check that is currently red.
1. **Draft as an instruction, not a report:** Address the receiving agent in the imperative. It needs orders and context, not a narrative of what you did.
1. **Test for self-containment:** Read the draft as someone who never saw this session. Every pronoun must resolve, every "as discussed" must be replaced by the substance, every file reference must be a path. Rewrite anything that only makes sense to you.
1. **Write the file:** Save to `.agents/prompts/<YYYY-MM-DD>-<slug>.md` and report the path so the user can paste or pipe it into the next session.

## Output Contract

Write to `.agents/prompts/` under the repository root, resolved with `git rev-parse --show-toplevel`. Outside a repository, fall back to `~/.agents/prompts/`. Create the directory if missing, and use a `<YYYY-MM-DD>-<slug>` filename so prompts sort chronologically and never collide.

The prompt file contains these sections, in this order. Omit a section only when it is genuinely empty — never pad it:

- **Objective** — the outcome in one or two sentences, stated as the goal, not the history.
- **Context** — the repository, branch, relevant paths, and the shape of the surrounding code the work touches.
- **Constraints** — user-stated requirements, marked as non-negotiable; project conventions that apply; explicitly rejected approaches.
- **Established Facts** — findings verified this session, each with how it was verified (file read, command run, documentation consulted).
- **Current State** — done (with proof), in progress (with the exact stopping point), not started.
- **Tasks** — ordered, each with its acceptance criterion. Prefer the smallest slice that produces working, verifiable behavior.
- **Verification** — the commands that must pass, and any check known to be red and why.
- **Open Questions** — decisions still owned by the user, with the options and your recommendation.

## Gotchas & Guidelines

- **No secrets**: never copy tokens, keys, credentials, or private URLs into the prompt. Name the variable or secret store instead.
- **Prompts are drafts, not history**: `.agents/prompts/` is a working inbox. Do not commit handovers unless the user wants them tracked; add `.agents/prompts/` to `.gitignore` when the repository does not already ignore `.agents/`.
- **Absolute paths break portability**: write repository-relative paths, or `~`-relative for user-level files.
- **Do not compress into jargon**: the receiving agent has no shared vocabulary with this session. Spell out the reasoning at each non-obvious decision.
- **Length follows substance**: a one-slice continuation is a short prompt. Padding a thin handover with restated repository structure buries the parts that matter.
- **Point, do not transcribe**: cite `path/to/file.go:42` rather than pasting the code. The next agent can read the file; it cannot read your reasoning.
- **Carry the failures**: an approach that failed is as valuable as the one that worked, and only this session knows it.

## Companion Skills

- [implementation-plan](../implementation-plan/SKILL.md) — when the follow-up needs ordered slices and proof boundaries rather than a prompt.
- [plan-execution](../plan-execution/SKILL.md) — what the receiving session runs once the handover is accepted.
- [prompt-design](../prompt-design/SKILL.md) — for production prompt stacks, not one-off session handovers.
