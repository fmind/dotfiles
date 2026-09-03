---
name: skillify
description: Turn this conversation, a repeated workflow, or an oversized AGENTS.md section into a global or local SKILL.md. Use when asked to skillify or capture a workflow.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/skillify
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Skillify

Capture what this session learned as a skill the next session can run without the conversation. The [agent-skills skill](../agent-skills/SKILL.md) owns the package format and validation; this skill owns the extraction and the writing.

## Workflow

1. **Extract from the session**: the goal, the user's trigger phrases, the exact commands that worked (with flags), the decisions and why, the dead ends, and the tools required; drop session-specific paths, one-off values, and secrets.
1. **Check the catalog**: `skills list` and `skills list -g`, then read any neighbor with an overlapping description; extend it when the workflow is the same, write a new skill only for a distinct trigger, and link neighbors instead of copying them.
1. **Choose the scope**:
   - **Global** (reusable, tool-generic): `~/.agents/skills/<name>/`, the `skills/` directory of the dot repository; add its CLI names to `skills/contracts.json`, then run `mise run check:skills` and `mise run test` there.
   - **Local** (repository-specific commands, data, or conventions): `.agents/skills/<name>/` in the project; add `.claude/skills -> ../.agents/skills` if missing per [agent-project](../agent-project/SKILL.md).
1. **Write from the template**: copy [skill.md](templates/skill.md) and apply the Skill Authoring Limits of the global `~/.agents/AGENTS.md` for name, description, size, shape, and placement; long configs and examples go to `references/`.
1. **Validate**: frontmatter `name` equals the directory; every link resolves; every required tool is named in the body; `gh skill publish --dry-run <parent-dir>` passes for a global skill.
1. **Test once**: run the skill's steps in a scratch directory or on the current repository and fix what does not reproduce; a skill that was never executed is a draft.
1. **Report**: the path, the description, the scope, and whether the routing probes in `dot/testdata/skills/` need a new prompt for the skill.

## Extracting from AGENTS.md

When `AGENTS.md` grows past rules and layout, move each multi-step section into its own local skill, leave a one-line pointer ("see `.agents/skills/<name>`"), and re-run [readme-agents](../readme-agents/SKILL.md).

## Gotchas

- **Descriptions route, bodies instruct**: the description decides when the skill loads; the body decides what happens. Do not summarize the workflow in the description.
- **Dates**: set `created` and `updated` to today; bump `updated` on every later edit.
- **Third-party content**: when the workflow came from an external skill, install it per [agent-skills](../agent-skills/SKILL.md) instead of retyping it.

## Documentation

- [Agent Skills specification](https://agentskills.io/specification)
- Companion skills: [agent-skills](../agent-skills/SKILL.md) (package format, validation, publishing), [agent-project](../agent-project/SKILL.md) (local skill layout), [readme-agents](../readme-agents/SKILL.md) (trim `AGENTS.md` after extraction).
