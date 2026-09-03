---
name: handover
description: Write a self-contained follow-up prompt carrying this session's decisions, constraints, and verified findings into a fresh agent session. Use when handing off work.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/handover
  created: "2026-08-11"
  updated: "2026-09-03"
---

# Handover

Turn everything this session established into one prompt a fresh agent can execute without asking a question. In the same harness, prefer `--continue`, `--resume`, or `/compact`; this skill is for a fresh or different harness. Writing a handover is a documentation act: produce the prompt file and stop, because the next session owns execution.

## Workflow

1. **Re-read the session**: recover the original request, every subsequent user correction, and the decisions each one settled. A later correction overrides an earlier instruction; carry the resolved position, not the debate.
1. **Separate instruction from inference**: mark what the user required, what you proposed and they approved, and what you assumed without confirmation. The next agent must know which lines it may revise.
1. **Verify before you write**: re-check every path, symbol, command, and version the prompt will name. Stale coordinates send the next session confidently to the wrong place; drop or re-derive anything you cannot confirm now.
1. **Capture the position**: completed work with its proof, in-progress work with its exact stopping point, and untouched work. Name the branch and whether the tree is clean.
1. **Record the trade-offs**: for each real decision, the choice, the alternatives, and why; for each dead end, what failed and how, so the next session does not retry it.
1. **State the proof contract**: the commands that must pass before the work is done (normally `mise run check` and `mise run test`) and any check that is currently red.
1. **Draft as an instruction, not a report**: address the receiving agent in the imperative; it needs orders and context, not a narrative of what you did.
1. **Test for self-containment**: read the draft as someone who never saw this session. Every pronoun must resolve, every "as discussed" becomes the substance, every file reference is a path.
1. **Write the file**: `.agents/prompts/<YYYY-MM-DD>-<slug>.md` under `git rev-parse --show-toplevel` (or `~/.agents/prompts/` outside a repository) using the [prompt template](references/prompt-template.md); create the directory if missing and report the path.

## Gotchas

- **Prompts are drafts, not history**: `.agents/prompts/` is a working inbox; do not commit handovers unless the user wants them tracked, and add `.agents/prompts/` to `.gitignore` when the repository does not already ignore `.agents/`.
- **Length follows substance**: a one-slice continuation is a short prompt; padding a thin handover with restated repository structure buries the parts that matter.
- **Point, do not transcribe**: cite `path/to/file.go:42` rather than pasting the code; the next agent can read the file but not your reasoning.
- **Carry the failures**: an approach that failed is as valuable as the one that worked, and only this session knows it.

## Documentation

- Companion skills: [implementation-plan](../implementation-plan/SKILL.md) (when the follow-up needs ordered slices rather than a prompt), [plan-execution](../plan-execution/SKILL.md) (what the receiving session runs), [prompt-design](../prompt-design/SKILL.md) (production prompt stacks, not session handovers), [agent-project](../agent-project/SKILL.md) (the `.agents/prompts/` layout).
